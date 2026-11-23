package apirisk

import (
	"net/http/httptest"
	"testing"
)

func TestForRequestUsesTheDocs(t *testing.T) {
	cases := []struct {
		method, pattern, path string
		want                  Risk
	}{
		{"GET", "GET /api/domains", "/api/domains", Read},
		{"POST", "POST /api/domains", "/api/domains", Change},
		{"DELETE", "DELETE /api/domains/{domain}", "/api/domains/example.com", Destructive},
		// documented as a password change, so destructive though it's a PATCH
		{"PATCH", "PATCH /api/emails/{email}", "/api/emails/a@example.com", Destructive},
		{"POST", "POST /api/domains/{domain}/suspend", "/api/domains/example.com/suspend", Destructive},
		// undocumented routes fall back to the path
		{"POST", "POST /api/made-up/{id}/restore", "/api/made-up/1/restore", Destructive},
		{"POST", "", "/api/made-up/thing", Change},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.path, nil)
		r.Pattern = c.pattern
		if got := ForRequest(r); got != c.want {
			t.Errorf("%s %s (pattern %q) = %s, want %s", c.method, c.path, c.pattern, got, c.want)
		}
	}
}

func TestDocKeyMatchesMuxPatterns(t *testing.T) {
	if docKey("GET", "/api/files/<path:path>") != docKey("GET", "/api/files/{path...}") {
		t.Error("doc and mux keys differ for a wildcard path")
	}
	if docKey("GET", "/api/x/<int:id>") != docKey("GET", "/api/x/{id}") {
		t.Error("doc and mux keys differ for a typed param")
	}
}

func TestWPCLIGetActions(t *testing.T) {
	for action, want := range map[string]Risk{"update_now": Change, "update_debug": Change, "update_site_information": Change, "update_update_preferences": Change, "site_info": Read, "debug_info": Read, "update_info": Read} {
		r := httptest.NewRequest("GET", "/api/sites/wp-cli/"+action+"?website=example.com&docroot=/var/www/html/example.com", nil)
		r.Pattern = "GET /api/sites/wp-cli/{action}"
		if got := ForRequest(r); got != want {
			t.Errorf("GET wp-cli %s = %s, want %s", action, got, want)
		}
	}
	if Classify("GET", "/api/sites/wp-cli/<action>", "") != Change {
		t.Error("the documented route should count as a change, its action decides at request time")
	}
}

func TestExposesSecrets(t *testing.T) {
	for path, want := range map[string]bool{
		"/api/sites/example.com/wp-info":      true,
		"/api/sites/example.com/blog/wp-info": true,
		"/api/phpmyadmin":                     true,
		"/api/sites/example.com":              false,
		"/api/mysql/databases":                false,
	} {
		if got := ExposesSecrets(httptest.NewRequest("GET", path, nil)); got != want {
			t.Errorf("%s = %v, want %v", path, got, want)
		}
	}
	if ExposesSecrets(httptest.NewRequest("POST", "/api/phpmyadmin", nil)) {
		t.Error("only reads hand back secrets")
	}
}
