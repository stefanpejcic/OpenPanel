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
