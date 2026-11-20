package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTheme points ThemesDir at a temp dir and makes a theme folder in it with readme.txt
func newTheme(t *testing.T, name string) string {
	t.Helper()
	orig := ThemesDir
	ThemesDir = t.TempDir()
	t.Cleanup(func() { ThemesDir = orig; SetTheme("") })
	dir := filepath.Join(ThemesDir, name)
	os.MkdirAll(filepath.Join(dir, "user"), 0o755)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("Theme Name: "+name+"\n"), 0o644)
	return dir
}

func TestSetTheme(t *testing.T) {
	dir := newTheme(t, "mytheme")
	os.MkdirAll(filepath.Join(ThemesDir, "noreadme"), 0o755)

	for _, bad := range []string{"", "missing", "noreadme", "..", "../mytheme", "/etc"} {
		SetTheme(bad)
		if customTemplateDir != "" {
			t.Fatalf("%q should keep the built-in templates", bad)
		}
	}

	os.WriteFile(filepath.Join(dir, "a.html"), []byte(`{{define "layout"}}custom{{end}}`), 0o644)
	SetTheme("mytheme")
	if customTemplateDir != dir {
		t.Fatalf("expected %s to be used, got %q", dir, customTemplateDir)
	}
	f, err := pageFS.Open("templates/a.html")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	// files the theme doesn't have come from the binary
	if _, err := pageFS.Open("templates/base.html"); err != nil {
		t.Fatalf("missing file should fall back to embedded: %v", err)
	}
}

func TestThemePartialsAndStatic(t *testing.T) {
	dir := newTheme(t, "kit")
	os.MkdirAll(filepath.Join(dir, "partials"), 0o755)
	os.WriteFile(filepath.Join(dir, "partials", "_theme_only.html"), []byte(`{{define "theme_only"}}x{{end}}`), 0o644)
	os.MkdirAll(filepath.Join(dir, "static", "css"), 0o755)
	os.WriteFile(filepath.Join(dir, "static", "css", "theme.css"), []byte("body{}"), 0o644)
	SetTheme("kit")

	got := pagePaths([]string{"templates/base.html"})
	if len(got) != 2 || got[1] != "templates/partials/_theme_only.html" {
		t.Fatalf("theme partial not added: %v", got)
	}
	if got := pagePaths([]string{"templates/user/email_template.html"}); len(got) != 1 {
		t.Fatalf("theme partials only go on base.html pages: %v", got)
	}
	if themeStaticPath("css/theme.css") == "" || themeStaticPath("../readme.txt") != "" {
		t.Fatal("theme static lookup wrong")
	}
}

func TestBrokenThemeTemplateFallsBack(t *testing.T) {
	dir := newTheme(t, "broken")
	os.WriteFile(filepath.Join(dir, "user", "email_template.html"), []byte(`{{define "reset_password_email"}}{{if}}`), 0o644)
	SetTheme("broken")

	p := &Page{paths: []string{"templates/user/email_template.html"}}
	if _, err := p.parse(defaultTheme); err != nil {
		t.Fatalf("broken theme template should fall back to the built-in one: %v", err)
	}

	os.WriteFile(filepath.Join(dir, "user", "email_template.html"), []byte(`{{define "reset_password_email"}}custom body{{end}}`), 0o644)
	frag := MustLoadFragment("user/email_template.html")
	w := httptest.NewRecorder()
	if err := frag.ExecuteTemplate(w, "reset_password_email", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.Body.String(), "custom body") {
		t.Fatalf("fragment didn't use the theme template: %q", w.Body.String())
	}
}

func TestThemePageRenderFallsBack(t *testing.T) {
	dir := newTheme(t, "needsdata")
	os.WriteFile(filepath.Join(dir, "user", "email_template.html"), []byte(`{{define "layout"}}themed {{.Missing.Field}}{{end}}`), 0o644)
	SetTheme("needsdata")

	p := &Page{paths: []string{"templates/user/email_template.html"}}
	w := httptest.NewRecorder()
	// built-in file has no "layout" so the fallback errors too, what matters is the half rendered themed output is gone
	p.Render(w, 200, struct{}{})
	if strings.Contains(w.Body.String(), "themed") {
		t.Fatalf("partial theme output leaked: %q", w.Body.String())
	}

	os.WriteFile(filepath.Join(dir, "user", "email_template.html"), []byte(`{{define "layout"}}themed ok{{end}}`), 0o644)
	p = &Page{paths: []string{"templates/user/email_template.html"}}
	w = httptest.NewRecorder()
	if err := p.Render(w, 200, nil); err != nil || w.Body.String() != "themed ok" {
		t.Fatalf("got %q %v", w.Body.String(), err)
	}
}
