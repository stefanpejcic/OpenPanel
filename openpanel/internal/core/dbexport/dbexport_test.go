package dbexport

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
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
	Send(w, "db.sql", "application/sql", []byte("SELECT 1;"), true)
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
