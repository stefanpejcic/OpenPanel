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
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderDefaultPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	api := map[string]VersionInfo{"8.4": {StatusLabel: "Latest", IsLatestVersion: true}, "8.2": {StatusLabel: "Security fixes only"}}
	data := DefaultVersionPageData{
		LayoutData: baseLayout(mgr, "/php/default"), PHPDefaultVersion: "8.2", Service: "php-fpm-8.2",
		Choices: phpVersionChoices([]string{"8.2", "8.4"}, "8.2", false, api),
	}
	w := httptest.NewRecorder()
	if err := defaultVersionPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="new_php_version" value="8.2" checked disabled`) {
		t.Error("expected the current default locked")
	}
	if !strings.Contains(body, `name="new_php_version" value="8.4" x-model="choice"`) || !strings.Contains(body, "The latest version") {
		t.Error("expected a selectable 8.4 card with its description")
	}
	if strings.Index(body, `name="new_php_version" value="8.4"`) > strings.Index(body, `name="new_php_version" value="8.2"`) {
		t.Error("newest version should be listed first")
	}
}

func TestPHPVersionChoices(t *testing.T) {
	got := phpVersionChoices([]string{"7.4", "8.10", "8.2", "8.9"}, "8.2", false, nil)
	var order []string
	for _, c := range got {
		order = append(order, c.Version)
	}
	if strings.Join(order, " ") != "8.10 8.9 8.2 7.4" {
		t.Errorf("order = %v", order)
	}
	if got[2].Description != "" || !got[2].Current {
		t.Errorf("8.2 should be current with no description when the API is unknown, got %+v", got[2])
	}

	ls := phpVersionChoices([]string{"7.4", "8.3", "8.5"}, "8.5.1", true, nil)
	if len(ls) != 2 || ls[0].Version != "8.5" || !ls[0].Current {
		t.Errorf("litespeed should only list its versions and match 8.5.1 to 8.5, got %+v", ls)
	}

	// the labels the php-versions API really returns
	for label, want := range map[string]string{"Supported (Latest)": "The latest", "Security-Fixes Only": "Security fixes", "Supported": "Fully supported", "Unsupported": "No longer"} {
		if d := phpVersionDescription(VersionInfo{StatusLabel: label}, true); !strings.HasPrefix(d, want) {
			t.Errorf("%s: %q", label, d)
		}
	}
	if d := phpVersionDescription(VersionInfo{StatusLabel: "Supported", IsEOLVersion: true}, true); !strings.HasPrefix(d, "No longer") {
		t.Errorf("isEOLVersion should win over the label, got %q", d)
	}

	api := map[string]VersionInfo{"8.5": {StatusLabel: "Supported (Latest)", IsLatestVersion: true}, "7.2": {StatusLabel: "Unsupported", IsEOLVersion: true}}
	old := phpVersionChoices([]string{"8.5", "7.2", "7.1", "5.6"}, "8.5", false, api)
	for _, c := range old[1:] {
		if !strings.HasPrefix(c.Description, "No longer") {
			t.Errorf("%s should be end of life, got %q", c.Version, c.Description)
		}
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
				buildOptionField("date.timezone", "", []string{"Africa/Abidjan", "Europe/Belgrade"}),
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
		for _, want := range []string{"dbTuning('/php/php8.2/options/recommendations')", `id="db-tuning"`, "Optimize PHP", "Confirm Changes", `x-ref="form"`, `name="display_errors" value="0"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
		if !strings.Contains(body, `<option value="" selected>`) || strings.Contains(body, `value="Africa/Abidjan" selected`) {
			t.Error("an unset timezone should select the empty option, not the first zone")
		}
		if n := strings.Count(body, "review(&#34;"); n != 3 {
			t.Errorf("expected an optimize button per option, got %d", n)
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
