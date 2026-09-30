package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/static
var testStaticEmbedFS embed.FS

func testStaticFS(t *testing.T) fs.FS {
	t.Helper()
	sub, err := fs.Sub(testStaticEmbedFS, "testdata/static")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	return sub
}

func writeOverride(t *testing.T, dir, rel, contents string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNewStaticAssetsNoOverrides(t *testing.T) {
	sa, err := NewStaticAssets(testStaticFS(t), "", true)
	if err != nil {
		t.Fatalf("NewStaticAssets: %v", err)
	}
	if sa.CustomCSS || sa.CustomJS {
		t.Errorf("expected no overrides detected, got CustomCSS=%v CustomJS=%v", sa.CustomCSS, sa.CustomJS)
	}

	req := httptest.NewRequest(http.MethodGet, "/css/custom.css", nil)
	w := httptest.NewRecorder()
	sa.Handler().ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "embedded-default") {
		t.Errorf("expected embedded default content, got %q", w.Body.String())
	}
}

func TestNewStaticAssetsWithOverride(t *testing.T) {
	dir := t.TempDir()
	writeOverride(t, dir, "custom.css", "body{color:red} /* real admin override */")

	sa, err := NewStaticAssets(testStaticFS(t), dir, true)
	if err != nil {
		t.Fatalf("NewStaticAssets: %v", err)
	}
	if !sa.CustomCSS {
		t.Error("expected CustomCSS=true when a real override file exists")
	}
	if sa.CustomJS {
		t.Error("expected CustomJS=false, no override written")
	}

	req := httptest.NewRequest(http.MethodGet, "/css/custom.css", nil)
	w := httptest.NewRecorder()
	sa.Handler().ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "real admin override") {
		t.Errorf("expected disk override content, got %q", w.Body.String())
	}
}

func TestNewStaticAssetsIgnoresEmptyOverride(t *testing.T) {
	dir := t.TempDir()
	writeOverride(t, dir, "custom.css", "") // 0 bytes should not count as an override

	sa, err := NewStaticAssets(testStaticFS(t), dir, true)
	if err != nil {
		t.Fatalf("NewStaticAssets: %v", err)
	}
	if sa.CustomCSS {
		t.Error("expected an empty override file to be ignored (0 bytes should not count as an override)")
	}
}

func TestServeRootFile(t *testing.T) {
	sa, err := NewStaticAssets(testStaticFS(t), "", true)
	if err != nil {
		t.Fatalf("NewStaticAssets: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	w := httptest.NewRecorder()
	sa.ServeRootFile("robots.txt")(w, req)

	if !strings.Contains(w.Body.String(), "User-agent") {
		t.Errorf("expected robots.txt content, got %q", w.Body.String())
	}
}

func TestNewStaticAssetsSkipsCustomCodeWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeOverride(t, dir, "custom.css", "body{color:red} /* real admin override */")

	sa, err := NewStaticAssets(testStaticFS(t), dir, false)
	if err != nil {
		t.Fatalf("NewStaticAssets: %v", err)
	}
	if sa.CustomCSS {
		t.Error("expected custom.css to be ignored when custom code is disabled")
	}
}

func TestLoadCustomCode(t *testing.T) {
	dir := t.TempDir()
	writeOverride(t, dir, "in_header.html", "<meta name=\"x\" content=\"head\">\n")
	writeOverride(t, dir, "in_footer.html", "   \n")

	LoadCustomCode(dir)
	t.Cleanup(func() { LoadCustomCode("") })
	if got := string(funcMap["customHeader"].(func() template.HTML)()); !strings.Contains(got, `content="head"`) {
		t.Errorf("expected header html, got %q", got)
	}
	if got := funcMap["customFooter"].(func() template.HTML)(); got != "" {
		t.Errorf("expected whitespace-only footer to be dropped, got %q", got)
	}
}
