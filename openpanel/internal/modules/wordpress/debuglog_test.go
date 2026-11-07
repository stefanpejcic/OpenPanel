package wordpress

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/gorilla/csrf"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func TestParseDebugLogSetting(t *testing.T) {
	cases := map[string]string{
		"true":   "",
		"false":  "",
		"1":      "",
		"":       "",
		`"true"`: "",
		`"1"`:    "",
		`"/var/www/html/example.com/logs/debug.log"`: "/var/www/html/example.com/logs/debug.log",
		"/var/www/html/example.com/x.log\n":          "/var/www/html/example.com/x.log",
	}
	for in, want := range cases {
		if got := parseDebugLogSetting(in); got != want {
			t.Errorf("parseDebugLogSetting(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDebugLogLocation(t *testing.T) {
	const docroot, host = "/var/www/html/example.com", "/home/u/docker-data/volumes/u_html_data/_data/example.com"
	cases := []struct {
		setting, wantHost, wantShown string
		custom, ok                   bool
	}{
		{"", host + "/wp-content/debug.log", "wp-content/debug.log", false, true},
		{docroot + "/logs/debug.log", host + "/logs/debug.log", "logs/debug.log", true, true},
		{docroot + "/wp-content/../private/php.log", host + "/private/php.log", "private/php.log", true, true},
		{docroot + "/../other.com/debug.log", "", "", true, false},
		{"/var/www/html/example.com.evil/debug.log", "", "", true, false},
		{"/etc/passwd", "", "", true, false},
		{docroot, "", "", true, false},
		{"logs/debug.log", "", "", true, false},
		{docroot + "/a\x00b", "", "", true, false},
	}
	for _, tc := range cases {
		gotHost, shown, custom, err := debugLogLocation(docroot, host, tc.setting)
		var pe *errDebugLogPath
		if tc.ok != (err == nil) || (!tc.ok && !errors.As(err, &pe)) {
			t.Errorf("%q: err = %v, want ok=%v", tc.setting, err, tc.ok)
			continue
		}
		if gotHost != tc.wantHost || shown != tc.wantShown || custom != tc.custom {
			t.Errorf("%q = %q, %q, %v", tc.setting, gotHost, shown, custom)
		}
	}
}

func writeLog(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "debug.log")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestTailLines(t *testing.T) {
	// last N lines, newest last, CRLF trimmed
	var b strings.Builder
	for i := 1; i <= 300; i++ {
		fmt.Fprintf(&b, "line %d\r\n", i)
	}
	lines, size, more, err := tailLines(writeLog(t, b.String()), 1<<20, 200)
	if err != nil || !more || len(lines) != 200 || lines[0] != "line 101" || lines[199] != "line 300" || size != int64(b.Len()) {
		t.Fatalf("last lines: %d lines, first %q, more=%v size=%d err=%v", len(lines), lines[0], more, size, err)
	}

	// byte cap: only the end is read and the cut first line is dropped
	lines, _, more, err = tailLines(writeLog(t, strings.Repeat("x", 100)+"\nabc\ndef\n"), 10, 200)
	if err != nil || !more || strings.Join(lines, "|") != "abc|def" {
		t.Fatalf("byte cap: %q more=%v err=%v", lines, more, err)
	}
	lines, _, _, _ = tailLines(writeLog(t, strings.Repeat("y", 50)), 10, 200)
	if len(lines) != 0 {
		t.Errorf("one cut line should leave nothing, got %q", lines)
	}

	// short file, no trailing newline
	lines, _, more, err = tailLines(writeLog(t, "a\nb"), 1<<20, 200)
	if err != nil || more || strings.Join(lines, "|") != "a|b" {
		t.Errorf("short: %q more=%v err=%v", lines, more, err)
	}

	// empty and missing
	if lines, _, _, err = tailLines(writeLog(t, ""), 1<<20, 200); err != nil || lines == nil || len(lines) != 0 {
		t.Errorf("empty: %q err=%v", lines, err)
	}
	if _, _, _, err = tailLines(filepath.Join(t.TempDir(), "nope.log"), 1<<20, 200); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing: err = %v", err)
	}
}

func TestTailLinesRefusesSpecialFiles(t *testing.T) {
	dir := t.TempDir()
	target := writeLog(t, "secret\n")
	link := filepath.Join(dir, "debug.log")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("no symlinks:", err)
	}
	if _, _, _, err := tailLines(link, 1<<20, 200); !errors.Is(err, errNotRegular) {
		t.Errorf("symlink: err = %v", err)
	}
	fifo := filepath.Join(dir, "fifo.log")
	if err := syscall.Mkfifo(fifo, 0o644); err == nil {
		if _, _, _, err := tailLines(fifo, 1<<20, 200); !errors.Is(err, errNotRegular) {
			t.Errorf("fifo: err = %v", err)
		}
		if err := clearDebugLog(fifo); !errors.Is(err, errNotRegular) {
			t.Errorf("clear fifo: err = %v", err)
		}
	}
}

func TestReadAndClearDebugLog(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "debug.log")
	res, err := readDebugLog(missing, "wp-content/debug.log", false)
	if err != nil || res.Exists || res.Lines == nil {
		t.Errorf("missing: %+v err=%v", res, err)
	}
	if err := clearDebugLog(missing); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("clear missing: err = %v", err)
	}

	p := writeLog(t, "[28-Oct-2026 09:14:02 UTC] PHP Notice: one\n[28-Oct-2026 09:20:41 UTC] PHP Warning: two\n")
	res, err = readDebugLog(p, "wp-content/debug.log", false)
	if err != nil || !res.Exists || len(res.Lines) != 2 || res.More {
		t.Fatalf("read: %+v err=%v", res, err)
	}
	if err := clearDebugLog(p); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(p); err != nil || info.Size() != 0 || info.Mode().Perm() != 0o644 {
		t.Errorf("after clear: %v %v", info, err)
	}
	res, _ = readDebugLog(p, "wp-content/debug.log", false)
	if !res.Exists || len(res.Lines) != 0 {
		t.Errorf("cleared log should exist and be empty: %+v", res)
	}
}

// noDB fails every query, so the ownership check says no
type noDB struct{}

func (noDB) Open(string) (driver.Conn, error) { return nil, errors.New("no database in tests") }

var registerNoDB sync.Once

func appWithoutDB(t *testing.T) *appctx.App {
	registerNoDB.Do(func() { sql.Register("wp-nodb", noDB{}) })
	db, err := sql.Open("wp-nodb", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &appctx.App{DB: db}
}

func TestDebugLogToolsGate(t *testing.T) {
	a := appWithoutDB(t)
	cases := []struct {
		method, action string
		want           int
	}{
		{http.MethodGet, "debug-log-clear", http.StatusMethodNotAllowed},
		{http.MethodPost, "debug-log", http.StatusMethodNotAllowed},
		{http.MethodGet, "debug-log", http.StatusForbidden},
		{http.MethodPost, "debug-log-clear", http.StatusForbidden},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(tc.method, "/wordpress/tools/"+tc.action+"?site=example.com", nil)
		r.SetPathValue("action", tc.action)
		w := httptest.NewRecorder()
		handleWPTools(a, w, r)
		if w.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.action, w.Code, tc.want)
		}
	}

	r := httptest.NewRequest(http.MethodGet, "/wordpress/tools/debug-log?site=bad%20domain", nil)
	r.SetPathValue("action", "debug-log")
	w := httptest.NewRecorder()
	handleWPTools(a, w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("bad site = %d, want 400", w.Code)
	}
}

// the clear POST goes through the panel's CSRF middleware like every other form
func TestDebugLogClearNeedsCSRF(t *testing.T) {
	a := appWithoutDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /token", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(csrf.Token(r))) })
	mux.HandleFunc("/wordpress/tools/{action}", func(w http.ResponseWriter, r *http.Request) { handleWPTools(a, w, r) })
	protect := csrf.Protect([]byte("0123456789abcdef0123456789abcdef"), csrf.Path("/"), csrf.Secure(false), csrf.FieldName("csrf_token"))(mux)
	serve := func(r *http.Request) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		protect.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
		return w
	}

	post := func() *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/wordpress/tools/debug-log-clear", strings.NewReader("site=example.com"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return r
	}
	if w := serve(post()); w.Code != http.StatusForbidden || strings.Contains(w.Body.String(), "own this domain") {
		t.Fatalf("no token: %d %q, want a CSRF rejection", w.Code, w.Body.String())
	}

	tw := serve(httptest.NewRequest(http.MethodGet, "/token", nil))
	r := post()
	for _, c := range tw.Result().Cookies() {
		r.AddCookie(c)
	}
	r.Header.Set("X-CSRF-Token", tw.Body.String())
	if w := serve(r); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "You do not own this domain.") {
		t.Fatalf("with token: %d %q, want the ownership check", w.Code, w.Body.String())
	}
}
