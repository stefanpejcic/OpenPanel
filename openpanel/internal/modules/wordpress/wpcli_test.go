package wordpress

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	if _, ok := sanitizeName("example.com"); !ok {
		t.Error("valid domain should pass")
	}
	if _, ok := sanitizeName(""); ok {
		t.Error("empty name should fail")
	}
	if _, ok := sanitizeName("../etc/passwd"); ok {
		t.Error("path traversal / slash should fail sanitizeName")
	}
	if _, ok := sanitizeName("evil; rm -rf /"); ok {
		t.Error("shell metacharacters should fail")
	}
}

func TestSanitizeWPCLIPath(t *testing.T) {
	if _, ok := sanitizeWPCLIPath("var/www/html/example.com"); !ok {
		t.Error("valid path should pass")
	}
	if _, ok := sanitizeWPCLIPath("$(whoami)"); ok {
		t.Error("command substitution should fail")
	}
	if _, ok := sanitizeWPCLIPath(""); ok {
		t.Error("empty path should fail")
	}
}

func TestSanitizePHPVersion(t *testing.T) {
	if v, ok := sanitizePHPVersion(""); !ok || v != "" {
		t.Error("empty php_version should pass through as empty")
	}
	if _, ok := sanitizePHPVersion("8.2"); !ok {
		t.Error("valid version should pass")
	}
	if _, ok := sanitizePHPVersion("8.2; rm -rf /"); ok {
		t.Error("injection attempt should fail")
	}
}

func TestWPCLIRealPath(t *testing.T) {
	vol := "/home/alice/docker-data/volumes/alice_html_data/_data"
	ok := map[string]string{
		"/var/www/html/example.com":      vol + "/example.com",
		"/var/www/html/example.com/blog": vol + "/example.com/blog",
		"example.com":                    vol + "/example.com",
	}
	for in, want := range ok {
		got, valid := wpCLIRealPath(vol, in)
		if !valid || got != want {
			t.Errorf("wpCLIRealPath(%q) = %q, %v, want %q", in, got, valid, want)
		}
	}
	for _, in := range []string{
		"/var/www/html/../../../../../bob/docker-data/volumes/bob_html_data/_data/site",
		"../../../../../bob/docker-data/volumes/bob_html_data/_data",
		"/var/www/html/../../../../../../etc",
	} {
		if got, valid := wpCLIRealPath(vol, in); valid {
			t.Errorf("wpCLIRealPath(%q) = %q, should be refused", in, got)
		}
	}
}

func TestGetWPConfigDBInfoRejectsSQL(t *testing.T) {
	write := func(t *testing.T, dbName, prefix string) string {
		dir := t.TempDir()
		cfg := "<?php\ndefine( 'DB_NAME', '" + dbName + "' );\n$table_prefix = '" + prefix + "';\n"
		if err := os.WriteFile(filepath.Join(dir, "wp-config.php"), []byte(cfg), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	if db, prefix, err := getWPConfigDBInfo(write(t, "acct_wp1", "wp_")); err != nil || db != "acct_wp1" || prefix != "wp_" {
		t.Fatalf("normal config = %q, %q, %v", db, prefix, err)
	}
	if _, _, err := getWPConfigDBInfo(write(t, "my-site.db", "x7_")); err != nil {
		t.Errorf("dots and dashes in DB_NAME should still work: %v", err)
	}
	bad := [][2]string{
		{"acct_wp1", "wp_users` WHERE 1; DROP DATABASE other; -- "},
		{"x` ; DROP DATABASE other; -- ", "wp_"},
		{"-e", "wp_"},
		{"acct_wp1", "wp-"},
	}
	for _, c := range bad {
		if db, prefix, err := getWPConfigDBInfo(write(t, c[0], c[1])); err == nil {
			t.Errorf("DB_NAME %q / prefix %q should be refused, got %q, %q", c[0], c[1], db, prefix)
		}
	}
}

func TestDroppableDB(t *testing.T) {
	for _, n := range []string{"acct_wp1", "my-site.db"} {
		if !droppableDB(n) {
			t.Errorf("droppableDB(%q) should be true", n)
		}
	}
	for _, n := range []string{"mysql", "INFORMATION_SCHEMA", "performance_schema", "sys", "", "a`b", "-x"} {
		if droppableDB(n) {
			t.Errorf("droppableDB(%q) should be false", n)
		}
	}
}

func TestPluginsToReinstall(t *testing.T) {
	list := `[{"name":"akismet","status":"inactive","version":"5.5"},{"name":"hello","status":"active","version":"1.7.2"},
{"name":"mu-plugin","status":"must-use","version":""},{"name":"advanced-cache.php","status":"dropin","version":""},
{"name":"--require=x.php","status":"active","version":"1"},{"name":"odd","status":"active","version":"--exec=x"}]`
	got, err := pluginsToReinstall([]byte(list))
	if err != nil {
		t.Fatal(err)
	}
	want := []wpPlugin{{"akismet", "inactive", "5.5"}, {"hello", "active", "1.7.2"}, {"odd", "active", ""}}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("plugin %d = %v, want %v", i, got[i], want[i])
		}
	}
	if _, err := pluginsToReinstall([]byte("Warning: not json")); err == nil {
		t.Error("non-JSON output should fail")
	}
}

func TestLoginBaseURL(t *testing.T) {
	cases := []struct{ stored, site, want string }{
		{"https://example.com", "example.com", "https://example.com"},
		{"http://www.example.com/blog/", "example.com/blog", "http://www.example.com/blog"},
		{"https://example.com:8443/blog", "example.com/blog", "https://example.com:8443/blog"},
		{"", "example.com/blog", "https://example.com/blog"},
		// anything else would hand the one-time admin token to another host
		{"https://evil.test", "example.com", "https://example.com"},
		{"https://example.com@evil.test/", "example.com", "https://example.com"},
		{"https://example.com.evil.test", "example.com", "https://example.com"},
		{"javascript:alert(1)", "example.com", "https://example.com"},
	}
	for _, c := range cases {
		if got := loginBaseURL(c.stored, c.site); got != c.want {
			t.Errorf("loginBaseURL(%q, %q) = %q, want %q", c.stored, c.site, got, c.want)
		}
	}
}

func TestPHPSerializeAssoc(t *testing.T) {
	got := phpSerializeAssoc(map[string]any{"user_id": 5, "expires": 1234}, []string{"user_id", "expires"})
	want := `a:2:{s:7:"user_id";i:5;s:7:"expires";i:1234;}`
	if got != want {
		t.Errorf("phpSerializeAssoc = %q, want %q", got, want)
	}
}

func TestWPCLIParamsFromForm(t *testing.T) {
	form := url.Values{"domain": {"example.com"}, "type": {"core"}}
	req := httptest.NewRequest(http.MethodPost, "/wp-cli/download?extra=1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	params := wpCLIParams(req)
	if params["domain"] != "example.com" {
		t.Errorf("params[domain] = %q, want example.com", params["domain"])
	}
	if params["extra"] != "1" {
		t.Errorf("params[extra] (from query) = %q, want 1", params["extra"])
	}
}

func TestWPCLIParamsFromJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/wp-cli/download", strings.NewReader(`{"domain":"example.com","type":"core"}`))
	req.Header.Set("Content-Type", "application/json")
	params := wpCLIParams(req)
	if params["domain"] != "example.com" {
		t.Errorf("params[domain] = %q, want example.com", params["domain"])
	}
}

func TestInstalledWordPressVersion(t *testing.T) {
	dir := t.TempDir()
	if got := installedWordPressVersion(dir, "latest"); got != "latest" {
		t.Fatalf("missing version.php should fall back, got %q", got)
	}
	os.MkdirAll(filepath.Join(dir, "wp-includes"), 0o755)
	os.WriteFile(filepath.Join(dir, "wp-includes", "version.php"), []byte("<?php\n/**\n * @global string $wp_version\n */\n$wp_version = '6.8.1';\n$wp_db_version = 60421;\n"), 0o644)
	if got := installedWordPressVersion(dir, "latest"); got != "6.8.1" {
		t.Fatalf("got %q, want 6.8.1", got)
	}
}
