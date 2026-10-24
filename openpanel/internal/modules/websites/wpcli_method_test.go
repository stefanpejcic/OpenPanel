package websites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestDebugAndUpdateFormsFromBody(t *testing.T) {
	post := httptest.NewRequest(http.MethodPost, "/wordpress/wp-cli/update_debug", strings.NewReader("csrf_token=x&website=example.com&WP_DEBUG=on&SAVEQUERIES=on"))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	got := debugOptionsFrom(post)
	want := map[string]string{"WP_DEBUG": "true", "WP_DEBUG_LOG": "false", "WP_DEBUG_DISPLAY": "false", "SCRIPT_DEBUG": "false", "SAVEQUERIES": "true"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("debug %s = %q, want %q", k, got[k], v)
		}
	}
	if post.FormValue("website") != "example.com" {
		t.Error("website should come from the POST body")
	}

	// the token-auth API still sends everything in the query string
	get := httptest.NewRequest(http.MethodGet, "/api/sites/wp-cli/update_update_preferences?WP_AUTO_UPDATE_CORE=minor&WP_AUTO_UPDATE_PLUGINS=true", nil)
	prefs := updatePrefsFrom(get)
	if prefs["WP_AUTO_UPDATE_CORE"] != "minor" || prefs["WP_AUTO_UPDATE_PLUGINS"] != "true" || prefs["WP_AUTO_UPDATE_THEMES"] != "false" {
		t.Errorf("prefs from query = %v", prefs)
	}
}

func TestDebugAndUpdateFormsStayOutOfTheURL(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "example.com/blog", Docroot: "/var/www/html/example.com/blog"}
	w := httptest.NewRecorder()
	if err := wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com/blog"}}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	// both forms hold the page's csrf_token, putting them in the URL leaks it into access logs
	for _, bad := range []string{`update_debug?" + formData`, `update_update_preferences?" + formData`} {
		if strings.Contains(body, bad) {
			t.Errorf("page still builds the URL from the form: %s", bad)
		}
	}
	for _, good := range []string{`fetch("/wordpress/wp-cli/update_debug", {`, `fetch("/wordpress/wp-cli/update_update_preferences", {`} {
		if !strings.Contains(body, good) {
			t.Errorf("page should POST the form body: %s", good)
		}
	}
}

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
