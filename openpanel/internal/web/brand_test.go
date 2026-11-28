package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func brandServer(dir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /brand/{file}", ServeBrandFile(dir))
	return mux
}

func TestServeBrandFile(t *testing.T) {
	dir := t.TempDir()
	writeOverride(t, dir, "logo.png", "\x89PNG\r\n\x1a\nfake")
	writeOverride(t, dir, "logo_dark.svg", `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	writeOverride(t, dir, "secret.key", "nope")
	mux := brandServer(dir)

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		return rec
	}

	rec := get("/brand/logo.png?v=abc")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("logo.png: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("versioned url should be cached for good, got %q", rec.Header().Get("Cache-Control"))
	}
	if get("/brand/logo.png").Header().Get("Cache-Control") != "no-cache" {
		t.Error("url without ?v= should not be cached")
	}

	rec = get("/brand/logo_dark.svg")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("logo_dark.svg: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if csp := rec.Header().Get("Content-Security-Policy"); csp == "" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("svg needs the sandbox csp and nosniff, got csp %q", csp)
	}

	for _, p := range []string{"/brand/secret.key", "/brand/logo.html", "/brand/favicon.png", "/brand/..%2fsecret.key", "/brand/logo_dark.png"} {
		if code := get(p).Code; code != 404 {
			t.Errorf("%s: want 404, got %d", p, code)
		}
	}
}

func TestServeBrandFileSkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need admin on windows")
	}
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "passwd")
	writeOverride(t, filepath.Dir(outside), "passwd", "root:x:0:0")
	if err := os.Symlink(outside, filepath.Join(dir, "logo.png")); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	brandServer(dir).ServeHTTP(rec, httptest.NewRequest("GET", "/brand/logo.png", nil))
	if rec.Code != 404 {
		t.Fatalf("symlinked logo: want 404, got %d", rec.Code)
	}
}
