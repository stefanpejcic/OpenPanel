package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFrameGuardSetsHeaders(t *testing.T) {
	h := frameGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	for _, path := range []string{"/login", "/website?domain=example.com", "/api/wordpress"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
			t.Errorf("%s: X-Frame-Options = %q", path, got)
		}
		if got := rec.Header().Get("Content-Security-Policy"); got != "frame-ancestors 'self'" {
			t.Errorf("%s: Content-Security-Policy = %q", path, got)
		}
	}
}
