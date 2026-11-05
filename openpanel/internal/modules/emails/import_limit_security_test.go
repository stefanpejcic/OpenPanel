package emails

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// minimal db: only the domains table, user 1 owns mine.test
type v9Driver struct{}
type v9Conn struct{}
type v9Stmt struct{ q string }
type v9Rows struct {
	cols []string
	vals [][]driver.Value
	i    int
}

func (v9Driver) Open(string) (driver.Conn, error)          { return v9Conn{}, nil }
func (v9Conn) Prepare(q string) (driver.Stmt, error)       { return &v9Stmt{q}, nil }
func (v9Conn) Close() error                                { return nil }
func (v9Conn) Begin() (driver.Tx, error)                   { return nil, fmt.Errorf("no tx") }
func (*v9Stmt) Close() error                               { return nil }
func (*v9Stmt) NumInput() int                              { return -1 }
func (*v9Stmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (r *v9Rows) Columns() []string                        { return r.cols }
func (r *v9Rows) Close() error                             { return nil }
func (r *v9Rows) Next(d []driver.Value) error {
	if r.i >= len(r.vals) {
		return io.EOF
	}
	copy(d, r.vals[r.i])
	r.i++
	return nil
}
func (st *v9Stmt) Query(args []driver.Value) (driver.Rows, error) {
	switch {
	case strings.Contains(st.q, "SELECT user_id FROM domains WHERE domain_url"):
		if fmt.Sprint(args[0]) == "mine.test" {
			return &v9Rows{cols: []string{"u"}, vals: [][]driver.Value{{int64(1)}}}, nil
		}
		return &v9Rows{cols: []string{"u"}}, nil
	case strings.Contains(st.q, "FROM domains WHERE user_id"):
		return &v9Rows{cols: []string{"a", "b", "c", "d"}, vals: [][]driver.Value{{int64(1), "/var/www/html/mine.test", "mine.test", "8.3"}}}, nil
	}
	return &v9Rows{cols: []string{"x"}}, nil
}

type v9Redis struct {
	mu sync.Mutex
	kv map[string]string
}

func v9StartRedis(t *testing.T) string {
	dir, _ := os.MkdirTemp("", "v9r")
	sock := filepath.Join(dir, "r.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close(); os.RemoveAll(dir) })
	m := &v9Redis{kv: map[string]string{}}
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

func v9ReadRESP(br *bufio.Reader) ([]string, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	n, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		h, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		l, _ := strconv.Atoi(strings.TrimSpace(h[1:]))
		buf := make([]byte, l+2)
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, err
		}
		out = append(out, string(buf[:l]))
	}
	return out, nil
}

func (m *v9Redis) handle(c net.Conn) {
	defer c.Close()
	br := bufio.NewReader(c)
	for {
		args, err := v9ReadRESP(br)
		if err != nil || len(args) == 0 {
			return
		}
		m.mu.Lock()
		var resp string
		switch strings.ToUpper(args[0]) {
		case "HELLO":
			resp = "-ERR unknown\r\n"
		case "GET":
			if v, ok := m.kv[args[1]]; ok {
				resp = fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			} else {
				resp = "$-1\r\n"
			}
		case "SET":
			m.kv[args[1]] = args[2]
			resp = "+OK\r\n"
		case "DEL":
			n := 0
			for _, k := range args[1:] {
				if _, ok := m.kv[k]; ok {
					delete(m.kv, k)
					n++
				}
			}
			resp = fmt.Sprintf(":%d\r\n", n)
		default:
			resp = "+OK\r\n"
		}
		m.mu.Unlock()
		c.Write([]byte(resp))
	}
}

var v9Once sync.Once

func TestImportRespectsPlanMailboxLimit(t *testing.T) {
	v9Once.Do(func() { sql.Register("v9fake", v9Driver{}) })
	db, _ := sql.Open("v9fake", "x")
	c := cache.New(v9StartRedis(t))
	a := &appctx.App{Config: config.Config{}, SecretKey: []byte("k0123456789abcdef0123456789abcde"), Sessions: session.NewStore([]byte("k0123456789abcdef0123456789abcde")), Cache: c, DB: db}
	ctx := context.Background()
	seed := func(k string, v any) {
		b, _ := json.Marshal(v)
		if err := c.Raw().Set(ctx, cache.KeyPrefix+k, b, time.Hour).Err(); err != nil {
			t.Fatal(err)
		}
	}
	// dummy tenant: user 1 'v9user', plan 1 allows 1 mailbox with 1G max quota, and already has 1 mailbox
	seed("get_user_details_with_plan:1", appctx.UserDetails{Username: "v9user", Context: "v9ctx", PlanID: 1, PlanName: "p", Found: true})
	seed("load_user_features:v9user:v9ctx", []string{"emails", "email_import"})
	seed("app.get_openpanel_version", "test")
	seed("query_plan_email_mailbox_limits:1", map[string]string{"EmailLimit": "1", "MaxEmailQuota": "1G"})
	seed("query_plan_details_by_id:1", map[string]any{})
	seed(emailsCacheKey("v9user"), []string{"* old@mine.test ( 0 / 1G ) [0%]"})

	bin := t.TempDir()
	logf := filepath.Join(t.TempDir(), "argv.log")
	script := "#!/bin/sh\necho \"$*\" >> " + logf + "\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "opencli"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	readLog := func() string { b, _ := os.ReadFile(logf); return string(b) }
	asUser := func(r *http.Request) *http.Request { return auth.WithAIAssistant(r, 1) }

	// baseline: single create is refused at the plan limit
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/emails", strings.NewReader(`{"domain":"mine.test","username":"single","password":"Pw123456!"}`))
	apiEmailsCreate(a, rr, asUser(req))
	t.Logf("CREATE status=%d body=%s", rr.Code, strings.TrimSpace(rr.Body.String()))
	t.Logf("CREATE opencli argv log after create: %q", readLog())

	// preview with 3 rows
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "x.csv")
	fw.Write([]byte("email,password,quota\nimp1@mine.test,Pw123456!,\nimp2@mine.test,Pw123456!,0\nimp3@mine.test,Pw123456!,50G\n"))
	mw.Close()
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/emails/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	apiEmailImportPreview(a, rr, asUser(req))
	t.Logf("PREVIEW status=%d", rr.Code)
	var prev struct {
		Token string      `json:"import_token"`
		Valid []ImportRow `json:"valid_users"`
	}
	json.Unmarshal(rr.Body.Bytes(), &prev)
	t.Logf("PREVIEW valid rows=%d", len(prev.Valid))

	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/emails/import/confirm", strings.NewReader(`{"import_token":"`+prev.Token+`"}`))
	apiEmailImportConfirm(a, rr, asUser(req))
	t.Logf("CONFIRM status=%d body=%s", rr.Code, strings.TrimSpace(rr.Body.String()))
	t.Logf("opencli argv log (all calls):\n%s", readLog())
	adds := strings.Count(readLog(), "email-setup email add imp")
	qs := strings.Count(readLog(), "email-setup quota set imp")
	t.Logf("RESULT import email-add calls=%d import quota-set calls=%d", adds, qs)
	if adds != 0 {
		t.Fatalf("import created %d mailboxes past the plan limit", adds)
	}
	if !strings.Contains(rr.Body.String(), "limit reached") {
		t.Fatalf("confirm response doesn't report the limit: %s", rr.Body.String())
	}
}
