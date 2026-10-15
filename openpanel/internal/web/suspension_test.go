package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withCaddyDomainsDir(t *testing.T, files map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		os.WriteFile(filepath.Join(dir, name+".conf"), []byte(body), 0644)
	}
	orig := CaddyDomainsDir
	CaddyDomainsDir = dir
	t.Cleanup(func() { CaddyDomainsDir = orig })
}

func TestReadDomainSuspension(t *testing.T) {
	withCaddyDomainsDir(t, map[string]string{
		"a.com": "# comment: hi\n# suspended_by: admin\n# reason: DMCA\na.com {\n}\n",
		"b.com": "b.com {\n# suspended_by: admin\n}\n",
	})
	if got := ReadDomainSuspension("a.com"); got != (DomainSuspension{By: "admin", Reason: "DMCA", Comment: "hi"}) {
		t.Fatalf("got %+v", got)
	}
	if IsAdminSuspended("b.com") {
		t.Fatal("marker inside the caddy block must not count, only the header")
	}
	if IsAdminSuspended("../a.com") || IsAdminSuspended("missing.com") {
		t.Fatal("bad or missing domain should not be suspended")
	}
}

func TestBlockAdminSuspended(t *testing.T) {
	withCaddyDomainsDir(t, map[string]string{
		"locked.com": "# suspended_by: admin\nlocked.com {\n}\n",
		"mine.com":   "# suspended_by: user\nmine.com {\n}\n",
	})
	a := testBulkApp()
	reached := false
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, r *http.Request) { reached = true }
	mux.Handle("/domains/docroot", BlockAdminSuspended(a)(http.HandlerFunc(ok)))
	mux.Handle("/api/domains/{rest...}", BlockAdminSuspended(a)(http.HandlerFunc(ok)))
	mux.Handle("/domains/save-dns-zone/{domain}", BlockAdminSuspended(a)(http.HandlerFunc(ok)))

	cases := []struct {
		name, method, path, body string
		blocked                  bool
	}{
		{"form post on locked", "POST", "/domains/docroot", "domain_name=locked.com", true},
		{"form post on user suspended", "POST", "/domains/docroot", "domain_name=mine.com", false},
		{"get stays readable", "GET", "/domains/docroot?domain_name=locked.com", "", false},
		{"api rest path", "PUT", "/api/domains/locked.com/vhost", "", true},
		{"api other domain", "PUT", "/api/domains/mine.com/vhost", "", false},
		{"dns path value", "POST", "/domains/save-dns-zone/locked.com", "", true},
	}
	for _, c := range cases {
		reached = false
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		if c.body != "" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else if c.method != "GET" {
			req.Header.Set("Content-Type", "application/json")
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if reached == c.blocked {
			t.Errorf("%s: blocked=%v, want %v (status %d)", c.name, !reached, c.blocked, rec.Code)
		}
	}
}
