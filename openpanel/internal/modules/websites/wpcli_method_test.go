package websites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsLocalRedirect(t *testing.T) {
	for _, p := range []string{"/sites", "/website?domain=example.com"} {
		if !isLocalRedirect(p) {
			t.Errorf("isLocalRedirect(%q) should be true", p)
		}
	}
	for _, p := range []string{"", "https://evil.test", "//evil.test", `/\evil.test`, "evil.test", "/x?u=https://evil.test"} {
		if isLocalRedirect(p) {
			t.Errorf("isLocalRedirect(%q) should be false", p)
		}
	}
}

func TestWPDocrootInContainer(t *testing.T) {
	for in, want := range map[string]string{
		"/var/www/html/example.com":     "/var/www/html/example.com",
		"var/www/html/example.com/blog": "/var/www/html/example.com/blog",
		"/var/www/html/":                "/var/www/html",
	} {
		if got, ok := wpDocrootInContainer(in); !ok || got != want {
			t.Errorf("wpDocrootInContainer(%q) = %q, %v, want %q", in, got, ok, want)
		}
	}
	if got, ok := wpDocrootInContainer("/var/www/html/../../etc"); ok {
		t.Errorf("traversal should be refused, got %q", got)
	}
}

func TestWPDBSizeFromOutput(t *testing.T) {
	if got, ok := wpDBSizeFromOutput("wpsmoke\t852 KB"); !ok || got != "852 KB" {
		t.Errorf("got %q, %v, want 852 KB", got, ok)
	}
	if _, ok := wpDBSizeFromOutput("wpsmoke"); ok {
		t.Error("a row without a size should fail")
	}
}

func TestScanVerifyDBRefusesQuotes(t *testing.T) {
	// the name check has to fail before anything reaches mysql
	if scanVerifyDB(context.Background(), "nobody", "x' OR '1'='1") {
		t.Error("a quoted db name should never verify")
	}
}

func TestWordPressWPCLIWriteActionsNeedPost(t *testing.T) {
	for action := range wsWriteWPCLIActions {
		if !wsAllowedWPCLIActions[action] {
			t.Errorf("write action %q is missing from wsAllowedWPCLIActions", action)
		}
		req := httptest.NewRequest(http.MethodGet, "/wordpress/wp-cli/"+action+"?website=example.com&docroot=/var/www/html/example.com", nil)
		req.SetPathValue("action", action)
		rec := httptest.NewRecorder()
		// a nil app is fine, the method check has to happen before anything touches it
		handleWordPressWPCLI(nil, rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want 405", action, rec.Code)
		}
	}
}
