package wordpress

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
