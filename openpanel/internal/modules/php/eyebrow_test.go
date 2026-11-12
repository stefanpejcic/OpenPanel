package php

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestIniLineCount(t *testing.T) {
	cases := map[string]int{"": 0, "a": 1, "a\n": 1, "a\nb\n": 2, "a\n\nb": 3}
	for in, want := range cases {
		if got := (PHPIniEditorPageData{FileContent: in}).LineCount(); got != want {
			t.Errorf("LineCount(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestWebServerLabel(t *testing.T) {
	cases := map[string]string{"apache": "Apache", "openlitespeed": "OpenLiteSpeed", "nginx": "Nginx", "": "", "caddy": "caddy"}
	for in, want := range cases {
		if got := webServerLabel(in); got != want {
			t.Errorf("webServerLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderPHPEyebrows(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	w := httptest.NewRecorder()
	settings := PHPSettingsPageData{LayoutData: baseLayout(mgr, "/php/domains"), AvailablePHPVersions: []string{"8.5", "8.4"}, DefaultVersion: "8.5"}
	if err := settingsPage.Render(w, 200, settings); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "Default 8.5 · 2 versions installed") {
		t.Error("expected default version in settings eyebrow")
	}

	w = httptest.NewRecorder()
	ini := PHPIniEditorPageData{
		LayoutData: baseLayout(mgr, "/php/php8.5.ini/editor"), Version: "8.5", WebServer: "Apache",
		FilePath: "/etc/php/8.5/fpm/php.ini", SavedPath: "/home/testuser/php.ini/8.5.ini", FileContent: "a=1\nb=2\n",
	}
	if err := iniEditorPage.Render(w, 200, ini); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"php-fpm-8.5 · Apache", "Mounted as", "/home/testuser/php.ini/8.5.ini", `<span id="ini-line-count">2</span> lines`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in ini editor", want)
		}
	}

	w = httptest.NewRecorder()
	opts := PHPOptionsPageData{LayoutData: baseLayout(mgr, "/php/php8.5/options"), Version: "8.5", WebServer: "Apache"}
	if err := optionsPage.Render(w, 200, opts); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "php-fpm-8.5 · Apache") {
		t.Error("expected web server in options eyebrow")
	}
}
