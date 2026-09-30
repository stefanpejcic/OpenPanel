package dbexport

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func TestHostDirStaysInWebRoot(t *testing.T) {
	ok := map[string]string{
		"/var/www/html":             "/home/bob/docker-data/volumes/bob_html_data/_data",
		"/var/www/html/backups":     "/home/bob/docker-data/volumes/bob_html_data/_data/backups",
		"/var/www/html/a/../backup": "/home/bob/docker-data/volumes/bob_html_data/_data/backup",
	}
	for in, want := range ok {
		if got, err := hostDir("bob", in); err != nil || got != want {
			t.Errorf("hostDir(%q) = %q, %v", in, got, err)
		}
	}
	for _, in := range []string{"", "/etc", "/var/www/html/../../etc", "/var/www/htmlx", "backups"} {
		if _, err := hostDir("bob", in); err == nil {
			t.Errorf("hostDir(%q) should be rejected", in)
		}
	}
}

func TestSendGzip(t *testing.T) {
	w := httptest.NewRecorder()
	if err := Send(w, exec.Command("printf", "SELECT 1;"), "db.sql", "application/sql", true); err != nil {
		t.Fatal(err)
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="db.sql.gz"` {
		t.Errorf("disposition %q", got)
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := io.ReadAll(zr); string(b) != "SELECT 1;" {
		t.Errorf("got %q", b)
	}
}

func TestSendFailsBeforeHeaders(t *testing.T) {
	for _, cmd := range []*exec.Cmd{exec.Command("false"), exec.Command("true"), exec.Command("/nonexistent-dump"), exec.Command("sh", "-c", "echo header; exit 2")} {
		w := httptest.NewRecorder()
		if err := Send(w, cmd, "db.sql", "application/sql", false); err == nil {
			t.Errorf("%v: expected error", cmd.Args)
		}
		if w.Header().Get("Content-Disposition") != "" || w.Body.Len() != 0 {
			t.Errorf("%v: nothing should be sent on early failure", cmd.Args)
		}
	}
}

func TestSendAbortsOnMidStreamFailure(t *testing.T) {
	defer func() {
		if r := recover(); r != http.ErrAbortHandler {
			t.Errorf("expected ErrAbortHandler panic, got %v", r)
		}
	}()
	_ = Send(httptest.NewRecorder(), exec.Command("sh", "-c", "head -c 200000 /dev/zero; exit 1"), "db.sql", "application/sql", false)
}
