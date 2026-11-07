package domains

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestDocrootRefusesBackupsFolder(t *testing.T) {
	// a docroot here would publish every site's database dump and wp-config.php copy
	for _, in := range []string{"/var/www/html/backups", "/var/www/html/backups/", "/var/www/html/backups/example.com", "/var/www/html/./backups", "/var/www/html/x/../backups/a"} {
		if got, ok := resolveUnderVarWWWHTML(in); ok || !insideBackups(got) {
			t.Errorf("resolveUnderVarWWWHTML(%q) = %q, %v, want refused as the backups folder", in, got, ok)
		}
	}
	for _, in := range []string{"/var/www/html/backups2", "/var/www/html/example.com/backups", "/var/www/html/example.com"} {
		if _, ok := resolveUnderVarWWWHTML(in); !ok {
			t.Errorf("resolveUnderVarWWWHTML(%q) should be allowed", in)
		}
	}
}

func TestHTTPURLRERejectsEmbeddedWhitespace(t *testing.T) {
	valid := []string{"http://example.com", "https://example.com/path?q=1"}
	invalid := []string{"", "ftp://example.com", "http://example.com\nDirective evil", "http://exa mple.com", "http://example.com\r\nDirective evil"}
	for _, s := range valid {
		if !httpURLRE.MatchString(s) {
			t.Errorf("expected %q to match httpURLRE", s)
		}
	}
	for _, s := range invalid {
		if httpURLRE.MatchString(s) {
			t.Errorf("expected %q to be rejected by httpURLRE", s)
		}
	}
}

func TestBuildPageEntries(t *testing.T) {
	entries := buildPageEntries(5, 10)
	var numbers []int
	ellipses := 0
	for _, e := range entries {
		if e.IsEllipsis {
			ellipses++
		} else {
			numbers = append(numbers, e.Number)
		}
	}
	// 1, ellipsis, 3,4,5,6,7, ellipsis, 10
	want := []int{1, 3, 4, 5, 6, 7, 10}
	if len(numbers) != len(want) {
		t.Fatalf("expected numbers %v, got %v", want, numbers)
	}
	for i, n := range want {
		if numbers[i] != n {
			t.Errorf("position %d: expected %d, got %d", i, n, numbers[i])
		}
	}
	if ellipses != 2 {
		t.Errorf("expected 2 ellipses, got %d", ellipses)
	}
}

func TestBuildLogPageEntries(t *testing.T) {
	t.Run("small total, no ellipsis needed", func(t *testing.T) {
		entries := buildLogPageEntries(1, 3)
		for _, e := range entries {
			if e.IsEllipsis {
				t.Errorf("did not expect an ellipsis for a 3-page total, got %+v", entries)
			}
		}
	})

	t.Run("large total, both ellipses", func(t *testing.T) {
		entries := buildLogPageEntries(10, 20)
		if entries[0].Number != 1 {
			t.Errorf("expected first entry to be page 1, got %+v", entries[0])
		}
		if !entries[1].IsEllipsis {
			t.Errorf("expected an ellipsis after page 1, got %+v", entries[1])
		}
		last := entries[len(entries)-1]
		if last.Number != 20 {
			t.Errorf("expected last entry to be page 20, got %+v", last)
		}
	})
}

func TestResolveUnderVarWWWHTML(t *testing.T) {
	cases := []struct {
		in       string
		wantOK   bool
		wantPath string
	}{
		{"/var/www/html/", true, "/var/www/html"},
		{"/var/www/html/example.com", true, "/var/www/html/example.com"},
		{"/etc/passwd", false, "/etc/passwd"},
		// filepath.Clean resolves ".." lexically here, landing on /var/etc/passwd not /etc/passwd - still correctly rejected as outside /var/www/html/
		{"/var/www/html/../../etc/passwd", false, "/var/etc/passwd"},
	}
	for _, c := range cases {
		got, ok := resolveUnderVarWWWHTML(c.in)
		if ok != c.wantOK {
			t.Errorf("resolveUnderVarWWWHTML(%q) ok = %v, want %v", c.in, ok, c.wantOK)
		}
		if got != c.wantPath {
			t.Errorf("resolveUnderVarWWWHTML(%q) = %q, want %q", c.in, got, c.wantPath)
		}
	}
}

func TestInsertOrReplaceRedirect(t *testing.T) {
	t.Run("insert before log block", func(t *testing.T) {
		content := "example.com {\n    root * /srv\n    log {\n        output file /var/log/x.log\n    }\n}\n"
		got := insertOrReplaceRedirect(content, redirectRule{To: "https://target.example"}, "example.com")
		if !containsLine(got, "redir https://target.example") {
			t.Errorf("expected inserted redir line, got:\n%s", got)
		}
	})

	t.Run("replace existing redir before anchor", func(t *testing.T) {
		content := "example.com {\n    redir https://old.example\n    import domain_log\n}\n"
		got := insertOrReplaceRedirect(content, redirectRule{To: "https://new.example"}, "example.com")
		if containsLine(got, "redir https://old.example") {
			t.Errorf("expected old redir line replaced, got:\n%s", got)
		}
		if !containsLine(got, "redir https://new.example") {
			t.Errorf("expected new redir line present, got:\n%s", got)
		}
	})
}

func TestRemoveRedirect(t *testing.T) {
	orig := "http://a.example {\n  import domain_log a.example\n}\nhttps://a.example {\n  import domain_log a.example\n}\n"
	added := insertOrReplaceRedirect(orig, redirectRule{To: "https://example.com"}, "old.example")
	if got := removeRedirect(added, "https://example.com"); got != orig {
		t.Errorf("add then remove should give the original back, got:\n%q", got)
	}
	// only the exact target goes, a longer url sharing the prefix stays
	other := "x {\n  redir https://example.com/path\n  import domain_log x\n}"
	if got := removeRedirect(other, "https://example.com"); got != other {
		t.Errorf("unrelated redir line removed:\n%q", got)
	}
	if got := removeRedirect("x {\n\tredir https://t.example 301\n}", "https://t.example"); got != "x {\n}" {
		t.Errorf("redir with a status code not removed: %q", got)
	}
}

func containsLine(content, substr string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == substr {
			return true
		}
	}
	return false
}

func TestCapitalizedDomainsFile(t *testing.T) {
	got := capitalizedDomainsFile("alice")
	want := "/home/alice/capitalized_domains.json"
	if got != want {
		t.Errorf("capitalizedDomainsFile(%q) = %q, want %q", "alice", got, want)
	}
}

func TestVhostFilePath(t *testing.T) {
	got := vhostFilePath("alice", "example.com")
	want := "/home/alice/docker-data/volumes/alice_webserver_data/_data/example.com.conf"
	if got != want {
		t.Errorf("vhostFilePath = %q, want %q", got, want)
	}
}

func TestReadWriteTextFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.conf")
	if err := writeTextFile(path, "hello world"); err != nil {
		t.Fatalf("writeTextFile: %v", err)
	}
	got, err := readTextFile(path)
	if err != nil {
		t.Fatalf("readTextFile: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestGetRedirectURLFromRealFile(t *testing.T) {
	// getRedirectURL() hardcodes /etc/openpanel/caddy/domains/, not writable in a sandboxed test env, so this just confirms the not-found path returns "" rather than panicking
	if got := getRedirectURL("no-such-domain-ever.example"); got != "" {
		t.Errorf("expected empty string for a nonexistent domain conf, got %q", got)
	}
}

func TestRedirectTargetProblem(t *testing.T) {
	existing := map[string]string{"old-shop.com": "https://example.com/", "blog.example.com": "https://medium.com/@me", "alt.com": "https://example.com:8443/"}
	cases := []struct {
		domain, target, want, other string
	}{
		{"example.com", "https://openpanel.com/", "", ""},
		{"example.com", "http://openpanel.com/path?q=1", "", ""},
		{"example.com", "", "Enter the URL visitors should go to.", ""},
		{"example.com", "openpanel.com", "Start with http:// or https:// and leave out spaces.", ""},
		{"example.com", "ftp://openpanel.com", "Start with http:// or https:// and leave out spaces.", ""},
		{"example.com", "https://a.com/x y", "Start with http:// or https:// and leave out spaces.", ""},
		{"example.com", "https://", "Start with http:// or https:// and leave out spaces.", ""},
		{"example.com", "https:///path", "That doesn't look like a full URL.", ""},
		{"example.com", "https://example.com/other", "This points back at %(domain)s, so visitors would loop forever.", ""},
		{"example.com", "http://WWW.Example.com.", "This points back at %(domain)s, so visitors would loop forever.", ""},
		{"www.example.com", "https://example.com/", "This points back at %(domain)s, so visitors would loop forever.", ""},
		{"example.com", "https://example.com:443/", "This points back at %(domain)s, so visitors would loop forever.", ""},
		{"example.com", "https://example.com:8080/", "", ""},
		{"example.com", "https://old-shop.com/", "%(other)s already redirects to %(domain)s. Saving this would send visitors in a circle.", "old-shop.com"},
		{"example.com", "https://www.old-shop.com/x", "%(other)s already redirects to %(domain)s. Saving this would send visitors in a circle.", "old-shop.com"},
		{"example.com", "https://blog.example.com/", "", ""},
		{"example.com", "https://alt.com/", "", ""},
		{"old-shop.com", "https://example.net/", "", ""},
		{"example.com", "https://a.com/?x={env.SECRET}", "The URL can't contain { } \" ` or \\ characters.", ""},
		{"example.com", "https://a.com/\"x", "The URL can't contain { } \" ` or \\ characters.", ""},
	}
	for _, c := range cases {
		msg, other := redirectTargetProblem(c.domain, c.target, existing)
		if msg != c.want || other != c.other {
			t.Errorf("redirectTargetProblem(%q, %q) = %q, %q; want %q, %q", c.domain, c.target, msg, other, c.want, c.other)
		}
	}
}

func TestBuildRedirectPageData(t *testing.T) {
	domains := []appctx.Domain{{DomainURL: "a.com"}, {DomainURL: "b.com"}, {DomainURL: "x.onion"}, {DomainURL: "c.com"}}
	data := buildRedirectPageData(web.LayoutData{}, domains, map[string]redirectRule{"b.com": {To: "https://a.com/"}}, "c.com")
	if data.DomainCount != 3 || len(data.Rows) != 1 || data.Rows[0] != (RedirectRow{Domain: "b.com", To: "https://a.com/"}) {
		t.Errorf("unexpected rows: %+v (count %d)", data.Rows, data.DomainCount)
	}
	if strings.Join(data.Free, ",") != "a.com,c.com" || data.Open != "c.com" {
		t.Errorf("unexpected free list %v or open %q", data.Free, data.Open)
	}
}

func TestRedirectBack(t *testing.T) {
	r := httptest.NewRequest("POST", "/domains/redirect", strings.NewReader("return_to=redirects"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()
	if got := redirectBack(r, "a.com", true); got != "/domains/redirect?domain=a.com" {
		t.Errorf("failed save from the Redirects page went to %q", got)
	}
	if got := redirectBack(r, "a.com", false); got != "/domains/redirect" {
		t.Errorf("saved from the Redirects page went to %q", got)
	}
	plain := httptest.NewRequest("POST", "/domains/redirect", nil)
	_ = plain.ParseForm()
	if got := redirectBack(plain, "a.com", true); got != "/domains" {
		t.Errorf("domains list form went to %q", got)
	}
}
