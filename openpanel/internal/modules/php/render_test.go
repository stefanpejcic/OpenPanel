package php

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "php": true, "php_options": true, "php_extensions": true, "php_ini": true, "phpmyadmin": true, "docker": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderDefaultPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DefaultVersionPageData{
		LayoutData: baseLayout(mgr, "/php/default"), PHPDefaultVersion: "8.2",
		InstalledVersions: []string{"8.1", "8.2"}, Service: "php-fpm-8.2",
	}
	w := httptest.NewRecorder()
	if err := defaultVersionPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "8.2") {
		t.Error("expected current default version in body")
	}
}

func TestRenderPHPSettingsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no domains", func(t *testing.T) {
		data := PHPSettingsPageData{LayoutData: baseLayout(mgr, "/php/domains")}
		w := httptest.NewRecorder()
		if err := settingsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet.") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with rows", func(t *testing.T) {
		data := PHPSettingsPageData{
			LayoutData: baseLayout(mgr, "/php/domains"),
			Domains: []PHPDomainRow{
				{DomainID: 1, DomainURL: "example.com", PHPVersion: "8.2", Level: "good"},
				{DomainID: 2, DomainURL: "old.example.com", PHPVersion: "/", Level: "unset"},
			},
			VersionCounts:        []PHPVersionCount{{Version: "8.2", Count: 1, Label: "Latest", Level: "good"}},
			AvailablePHPVersions: []string{"8.1", "8.2"},
		}
		w := httptest.NewRecorder()
		if err := settingsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "example.com") {
			t.Error("expected domain row")
		}
		if !strings.Contains(body, "Latest") {
			t.Error("expected version counter label")
		}
	})
}

func TestRenderPHPInfoPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := PHPInfoPageData{
		LayoutData: web.LayoutData{Title: "PHP 8.2 Info", T: mgr.Translator("en")},
		Version:    "8.2", Service: "php-fpm-8.2", FileContent: "phpversion is 8.2.10",
	}
	w := httptest.NewRecorder()
	if err := infoPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "phpversion is 8.2.10") {
		t.Error("expected file content in body")
	}
	if !strings.Contains(body, "<html") {
		t.Error("expected a standalone HTML document, not the panel shell")
	}
}

func TestRenderPHPIniEditorPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := PHPIniEditorPageData{LayoutData: baseLayout(mgr, "/php/php_ini_editor"), InstalledVersions: []string{"8.1", "8.2"}}
		w := httptest.NewRecorder()
		if err := iniEditorPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "PHP.INI Editor") {
			t.Error("expected selector heading")
		}
	})

	t.Run("editor with issues", func(t *testing.T) {
		data := PHPIniEditorPageData{
			LayoutData: baseLayout(mgr, "/php/php8.2.ini/editor"), Version: "8.2", FileContent: "memory_limit = 256M",
			Issues: []HealthIssue{{ID: "php-ini-syntax:8.2", Severity: "error", Message: "boom"}},
		}
		w := httptest.NewRecorder()
		if err := iniEditorPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "memory_limit = 256M") {
			t.Error("expected file content in editor textarea")
		}
		if !strings.Contains(body, "reportHealthIssues") {
			t.Error("expected health issue script block")
		}
	})
}

func TestRenderPHPOptionsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := PHPOptionsPageData{LayoutData: baseLayout(mgr, "/php/options"), InstalledVersions: []string{"8.2"}}
		w := httptest.NewRecorder()
		if err := optionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Edit PHP options") {
			t.Error("expected selector heading")
		}
	})

	t.Run("fields", func(t *testing.T) {
		data := PHPOptionsPageData{
			LayoutData: baseLayout(mgr, "/php/php8.2/options"), Version: "8.2",
			Fields: []OptionField{
				{Key: "display_errors", Kind: "checkbox_binary", Value: "1", Checked: true},
				{Key: "post_max_size", Kind: "unit", Value: "256M", NumberPart: "256", UnitPart: "M"},
			},
		}
		w := httptest.NewRecorder()
		if err := optionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "display_errors") || !strings.Contains(body, "post_max_size") {
			t.Error("expected option keys in body")
		}
	})
}

func TestRenderPHPExtensionsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := PHPExtensionsPageData{LayoutData: baseLayout(mgr, "/php/extensions"), InstalledVersions: []string{"8.2"}}
		w := httptest.NewRecorder()
		if err := extensionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "PHP Extensions") {
			t.Error("expected selector heading")
		}
	})

	t.Run("with extensions", func(t *testing.T) {
		data := PHPExtensionsPageData{
			LayoutData: baseLayout(mgr, "/php/php8.2/extensions"), Version: "8.2",
			Extensions:      []ExtensionRow{{Name: "bcmath", State: "active"}, {Name: "xdebug", State: "disabled"}},
			RecentlyRemoved: []string{"redis"},
		}
		w := httptest.NewRecorder()
		if err := extensionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "bcmath") || !strings.Contains(body, "xdebug") {
			t.Error("expected extension rows in body")
		}
		if !strings.Contains(body, "redis") {
			t.Error("expected recently-removed extension chip")
		}
	})
}

func TestRenderPHPMyAdminUnavailablePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := PHPMyAdminUnavailablePageData{LayoutData: baseLayout(mgr, "/mysql/phpmyadmin"), ErrorMessage: "Please contact support."}
	w := httptest.NewRecorder()
	if err := phpMyAdminUnavailablePage.Render(w, 503, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if w.Code != 503 {
		t.Errorf("status = %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Please contact support.") {
		t.Error("expected error message in body")
	}
}
