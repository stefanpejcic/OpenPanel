package account

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// fakeUser is one row of the users table as these tests see it
type fakeUser struct {
	username, email, context string
	twofa                    bool
}

type fakeToken struct {
	id, uid int
	hash    string
	ro      bool
}

// fakeStore is a tiny in-memory stand-in for the handful of panel tables the auth paths touch
type fakeStore struct {
	mu       sync.Mutex
	users    map[int]*fakeUser
	tokens   []fakeToken
	nextID   int
	passkeys map[int]int
	execs    []string
}

var (
	fakeStoresMu sync.Mutex
	fakeStores   = map[string]*fakeStore{}
	fakeDriverOn sync.Once
)

type fakeDriver struct{}
type fakeConn struct{ s *fakeStore }
type fakeStmt struct {
	s *fakeStore
	q string
}
type fakeRows struct {
	cols []string
	vals [][]driver.Value
	i    int
}

func (fakeDriver) Open(dsn string) (driver.Conn, error) {
	fakeStoresMu.Lock()
	defer fakeStoresMu.Unlock()
	return fakeConn{fakeStores[dsn]}, nil
}
func (c fakeConn) Prepare(q string) (driver.Stmt, error) { return &fakeStmt{c.s, q}, nil }
func (fakeConn) Close() error                            { return nil }
func (fakeConn) Begin() (driver.Tx, error)               { return nil, errors.New("no tx") }
func (*fakeStmt) Close() error                           { return nil }
func (*fakeStmt) NumInput() int                          { return -1 }
func (r *fakeRows) Columns() []string                    { return r.cols }
func (r *fakeRows) Close() error                         { return nil }
func (r *fakeRows) Next(d []driver.Value) error {
	if r.i >= len(r.vals) {
		return io.EOF
	}
	copy(d, r.vals[r.i])
	r.i++
	return nil
}

func argInt(v driver.Value) int {
	switch x := v.(type) {
	case int64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func argBool(v driver.Value) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	}
	return false
}

func (st *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	s := st.s
	s.mu.Lock()
	defer s.mu.Unlock()
	s.execs = append(s.execs, st.q)
	switch {
	case strings.HasPrefix(st.q, "INSERT INTO mcp_tokens"):
		s.tokens = append(s.tokens, fakeToken{id: s.nextID, uid: argInt(args[0]), hash: args[3].(string), ro: argBool(args[4])})
		s.nextID++
		return driver.RowsAffected(1), nil
	case strings.HasPrefix(st.q, "DELETE FROM mcp_tokens WHERE id = ? AND user_id = ?"):
		id, uid := argInt(args[0]), argInt(args[1])
		for i, t := range s.tokens {
			if t.id == id && t.uid == uid {
				s.tokens = append(s.tokens[:i], s.tokens[i+1:]...)
				return driver.RowsAffected(1), nil
			}
		}
		return driver.RowsAffected(0), nil
	case strings.HasPrefix(st.q, "DELETE FROM mcp_tokens WHERE user_id = ?"):
		uid, kept, n := argInt(args[0]), s.tokens[:0], 0
		for _, t := range s.tokens {
			if t.uid == uid {
				n++
				continue
			}
			kept = append(kept, t)
		}
		s.tokens = kept
		return driver.RowsAffected(int64(n)), nil
	case strings.HasPrefix(st.q, "DELETE FROM user_passkeys WHERE user_id = ?"):
		uid := argInt(args[0])
		n := s.passkeys[uid]
		delete(s.passkeys, uid)
		return driver.RowsAffected(int64(n)), nil
	}
	return driver.RowsAffected(1), nil
}

func (st *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	s := st.s
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case strings.Contains(st.q, "FROM mcp_tokens WHERE token_hash"):
		cols := []string{"id", "user_id", "read_only", "expires_at", "last_used_at"}
		for _, t := range s.tokens {
			if t.hash == args[0].(string) {
				return &fakeRows{cols: cols, vals: [][]driver.Value{{int64(t.id), int64(t.uid), t.ro, nil, nil}}}, nil
			}
		}
		return &fakeRows{cols: cols}, nil
	case strings.Contains(st.q, "SELECT users.username, users.server"):
		if u := s.users[argInt(args[0])]; u != nil {
			return &fakeRows{cols: []string{"u", "s", "e", "pid", "pn"}, vals: [][]driver.Value{{u.username, u.context, u.email, int64(1), "default"}}}, nil
		}
		return &fakeRows{cols: []string{"u", "s", "e", "pid", "pn"}}, nil
	case strings.Contains(st.q, "SELECT twofa_enabled, otp_secret FROM users WHERE id"):
		if u := s.users[argInt(args[0])]; u != nil {
			return &fakeRows{cols: []string{"t", "o"}, vals: [][]driver.Value{{u.twofa, "JBSWY3DPEHPK3PXP"}}}, nil
		}
		return &fakeRows{cols: []string{"t", "o"}}, nil
	case strings.Contains(st.q, "SELECT id, username, email FROM users WHERE email"):
		// case-insensitive like the shipped latin1_swedish_ci column
		for id, u := range s.users {
			if strings.EqualFold(u.email, args[0].(string)) {
				return &fakeRows{cols: []string{"id", "u", "e"}, vals: [][]driver.Value{{int64(id), u.username, u.email}}}, nil
			}
		}
		return &fakeRows{cols: []string{"id", "u", "e"}}, nil
	case strings.Contains(st.q, "SELECT username FROM users WHERE id"):
		if u := s.users[argInt(args[0])]; u != nil {
			return &fakeRows{cols: []string{"u"}, vals: [][]driver.Value{{u.username}}}, nil
		}
	}
	return &fakeRows{cols: []string{"x"}}, nil
}

func (s *fakeStore) addToken(uid int, raw string, ro bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	sum := sha256.Sum256([]byte(raw))
	id := s.nextID
	s.tokens = append(s.tokens, fakeToken{id: id, uid: uid, hash: hex.EncodeToString(sum[:]), ro: ro})
	s.nextID++
	return id
}

func (s *fakeStore) tokenCount(uid int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, t := range s.tokens {
		if t.uid == uid {
			n++
		}
	}
	return n
}

// featureContext writes a features.txt and returns a users.server value that LoadUserFeatures resolves to it
func featureContext(t *testing.T, features ...string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "features.txt"), []byte(strings.Join(features, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return "../" + strings.TrimPrefix(dir, "/")
}

// newFakeApp wires an App to an in-memory db and a throwaway redis on a unix socket
func newFakeApp(t *testing.T, cfg config.Config, users map[int]*fakeUser) (*appctx.App, *fakeStore) {
	t.Helper()
	fakeDriverOn.Do(func() { sql.Register("accountfake", fakeDriver{}) })
	store := &fakeStore{users: users, nextID: 1, passkeys: map[int]int{}}
	dsn := t.Name()
	fakeStoresMu.Lock()
	fakeStores[dsn] = store
	fakeStoresMu.Unlock()
	db, err := sql.Open("accountfake", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	c := cache.New(startMiniRedis(t))
	secret := []byte("test-secret-key-0123456789abcdef")
	if cfg == nil {
		cfg = config.Config{}
	}
	orig := weakpassScript
	weakpassScript = filepath.Join(t.TempDir(), "missing.sh")
	t.Cleanup(func() { weakpassScript = orig })
	return &appctx.App{
		Config:    cfg,
		SecretKey: secret,
		Sessions:  session.NewStore(secret),
		Cache:     c,
		I18n:      i18n.NewManager(t.TempDir(), c),
		DB:        db,
	}, store
}

// miniRedis speaks just enough RESP for go-redis: strings with ttl, NX, GETDEL, hashes and an empty SCAN
type miniRedis struct {
	mu   sync.Mutex
	kv   map[string]string
	exp  map[string]time.Time
	hash map[string]map[string]string
}

func startMiniRedis(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mr")
	if err != nil {
		t.Fatal(err)
	}
	sock := filepath.Join(dir, "r.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close(); os.RemoveAll(dir) })
	m := &miniRedis{kv: map[string]string{}, exp: map[string]time.Time{}, hash: map[string]map[string]string{}}
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go m.handle(c)
		}
	}()
	return sock
}

func readRESP(br *bufio.Reader) ([]string, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	n, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		l, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		sz, _ := strconv.Atoi(strings.TrimSpace(l[1:]))
		buf := make([]byte, sz+2)
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, err
		}
		out = append(out, string(buf[:sz]))
	}
	return out, nil
}

func bulk(s string) string { return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s) }

func (m *miniRedis) get(k string) (string, bool) {
	if e, ok := m.exp[k]; ok && time.Now().After(e) {
		delete(m.kv, k)
		delete(m.exp, k)
	}
	v, ok := m.kv[k]
	return v, ok
}

func (m *miniRedis) handle(c net.Conn) {
	defer c.Close()
	br := bufio.NewReader(c)
	for {
		cmd, err := readRESP(br)
		if err != nil || len(cmd) == 0 {
			return
		}
		m.mu.Lock()
		resp := "+OK\r\n"
		switch strings.ToUpper(cmd[0]) {
		case "HELLO":
			resp = "-ERR unknown command\r\n"
		case "PING":
			resp = "+PONG\r\n"
		case "GET":
			if v, ok := m.get(cmd[1]); ok {
				resp = bulk(v)
			} else {
				resp = "$-1\r\n"
			}
		case "GETDEL":
			if v, ok := m.get(cmd[1]); ok {
				delete(m.kv, cmd[1])
				resp = bulk(v)
			} else {
				resp = "$-1\r\n"
			}
		case "SET":
			nx, ttl := false, time.Duration(0)
			for i := 3; i < len(cmd); i++ {
				switch strings.ToUpper(cmd[i]) {
				case "NX":
					nx = true
				case "EX":
					n, _ := strconv.Atoi(cmd[i+1])
					ttl = time.Duration(n) * time.Second
					i++
				case "PX":
					n, _ := strconv.Atoi(cmd[i+1])
					ttl = time.Duration(n) * time.Millisecond
					i++
				}
			}
			if _, exists := m.get(cmd[1]); nx && exists {
				resp = "$-1\r\n"
				break
			}
			m.kv[cmd[1]] = cmd[2]
			if ttl > 0 {
				m.exp[cmd[1]] = time.Now().Add(ttl)
			}
		case "DEL":
			n := 0
			for _, k := range cmd[1:] {
				if _, ok := m.kv[k]; ok {
					n++
				}
				delete(m.kv, k)
				delete(m.hash, k)
			}
			resp = fmt.Sprintf(":%d\r\n", n)
		case "HSET":
			h := m.hash[cmd[1]]
			if h == nil {
				h = map[string]string{}
				m.hash[cmd[1]] = h
			}
			for i := 2; i+1 < len(cmd); i += 2 {
				h[cmd[i]] = cmd[i+1]
			}
			resp = ":1\r\n"
		case "HGETALL":
			h := m.hash[cmd[1]]
			resp = fmt.Sprintf("*%d\r\n", len(h)*2)
			for k, v := range h {
				resp += bulk(k) + bulk(v)
			}
		case "EXPIRE":
			resp = ":1\r\n"
		case "SCAN":
			resp = "*2\r\n$1\r\n0\r\n*0\r\n"
		}
		m.mu.Unlock()
		if _, err := c.Write([]byte(resp)); err != nil {
			return
		}
	}
}

// smtpSink accepts plain smtp on loopback and reports every RCPT TO address
func smtpSink(t *testing.T) (string, chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	rcpts := make(chan string, 50)
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				_, _ = io.WriteString(c, "220 sink\r\n")
				inData := false
				for {
					l, err := br.ReadString('\n')
					if err != nil {
						return
					}
					if inData {
						if l == ".\r\n" {
							inData = false
							_, _ = io.WriteString(c, "250 ok\r\n")
						}
						continue
					}
					u := strings.ToUpper(l)
					switch {
					case strings.HasPrefix(u, "RCPT TO:"):
						rcpts <- strings.Trim(strings.TrimSpace(l[8:]), "<>")
						_, _ = io.WriteString(c, "250 ok\r\n")
					case strings.HasPrefix(u, "DATA"):
						inData = true
						_, _ = io.WriteString(c, "354 go\r\n")
					case strings.HasPrefix(u, "QUIT"):
						_, _ = io.WriteString(c, "221 bye\r\n")
						return
					default:
						_, _ = io.WriteString(c, "250 ok\r\n")
					}
				}
			}(c)
		}
	}()
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	return port, rcpts
}
