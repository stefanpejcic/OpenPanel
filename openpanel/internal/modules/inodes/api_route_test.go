package inodes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// the documented /api/inodes used to redirect to /api/inodes/, which clients that don't follow redirects never get past
func TestAPIRouteWithoutTrailingSlash(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPI(mux, &appctx.App{Config: config.Config{}})
	for _, target := range []string{"/api/inodes", "/api/inodes/", "/api/inodes/site"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		// no token, so the API auth answers, the point is that it isn't a redirect
		if w.Code >= 300 && w.Code < 400 {
			t.Errorf("%s: HTTP %d redirect to %q", target, w.Code, w.Header().Get("Location"))
		}
	}
}
