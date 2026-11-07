package php

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// statuses as the test VM's api_versions.json has them
var wizardAPI = map[string]VersionInfo{
	"8.5": {StatusLabel: "Supported (Latest)", IsLatestVersion: true},
	"8.4": {StatusLabel: "Supported", IsSecureVersion: true},
	"8.3": {StatusLabel: "Security-Fixes Only", IsSecureVersion: true},
	"8.1": {StatusLabel: "Unsupported", IsEOLVersion: true},
	"7.4": {StatusLabel: "Unsupported", IsEOLVersion: true},
}

func TestWizardVersions(t *testing.T) {
	domains := []WizardDomain{{Domain: "old.example.com", Version: "7.4"}}
	got := wizardVersions([]string{"7.4", "8.5", "8.3"}, "8.5", domains, false, wizardAPI)
	want := []WizardVersion{
		{Version: "8.5", Label: "Supported (Latest)", Level: "good", Installed: true, InUse: true},
		{Version: "8.3", Label: "Security-Fixes Only", Level: "secure", Installed: true},
		{Version: "7.4", Label: "Unsupported", Level: "unsupported", Installed: true, InUse: true},
		{Version: "8.4", Label: "Supported", Level: "secure"},
		{Version: "8.1", Label: "Unsupported", Level: "unsupported"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("version %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// LiteSpeed only offers its own image tags, and no API means no labels
	ls := wizardVersions([]string{"8.3"}, "8.3", nil, true, nil)
	if len(ls) != 4 || ls[0].Version != "8.3" || !ls[0].Installed || ls[0].Label != "" || ls[1].Version != "8.5" || ls[1].Installed {
		t.Errorf("litespeed versions = %+v", ls)
	}
}

func TestWizardOpen(t *testing.T) {
	versions := wizardVersions([]string{"8.5", "8.3", "5.6"}, "8.3", nil, false, wizardAPI)
	cases := []struct{ requested, def, want string }{
		{"5.6", "8.3", "5.6"},
		{"", "8.3", "8.3"},
		{"8.4", "8.3", "8.3"}, // not installed
		{"../x", "8.3", "8.3"},
		{"", "8.3.12", "8.3"}, // LiteSpeed reports the full version
		{"", "9.9", "8.5"},
	}
	for _, c := range cases {
		if got := wizardOpen(c.requested, c.def, versions); got != c.want {
			t.Errorf("wizardOpen(%q, %q) = %q, want %q", c.requested, c.def, got, c.want)
		}
	}
	if got := wizardOpen("8.5", "8.5", nil); got != "" {
		t.Errorf("no versions = %q", got)
	}
}

func TestRenderWizardPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	domains := []WizardDomain{{Domain: "a.example.com", Version: "8.5", Level: "good"}, {Domain: "old.example.com", Version: "7.4", Level: "unsupported"}}
	data := PHPWizardPageData{
		LayoutData: baseLayout(mgr, "/php/wizard"), Default: "8.5", WebServer: "apache",
		Versions: wizardVersions([]string{"8.5", "7.4"}, "8.5", domains, false, wizardAPI),
		Domains:  domains, DomainCount: 3, Open: "8.5", Tab: "extensions",
	}
	w := httptest.NewRecorder()
	if err := wizardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"PHP Wizard", "apache · php-fpm per version", "phpWizard(", `<option value="8.5">PHP 8.5 (default)</option>`, `<option value="8.4" disabled>PHP 8.4 · ask your provider</option>`,
		"Ask your provider", "8.5, 7.4", `action="/php/domains"`, `name="new_php_version"`, "/php/wizard?v=",
		"/install-extensions", "/php/domains/bulk", "/php/default?output=json", "'.ini/editor'", "op-phpw-az",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("wizard page missing %q", want)
		}
	}
	if strings.Contains(body, "Install PHP") {
		t.Error("users can't install PHP versions, the page shouldn't offer it")
	}

	// LiteSpeed has no per-domain versions and no domains tab
	data.IsLitespeed, data.WebServer, data.Domains = true, "openlitespeed", nil
	w = httptest.NewRecorder()
	if err := wizardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body = w.Body.String()
	if !strings.Contains(body, "All on the default version") || strings.Contains(body, "tab = 'domains'") {
		t.Error("litespeed wizard should drop the domains tab")
	}

	// without extension and options features those tabs are gone
	data.IsLitespeed = false
	data.UserAllowed = map[string]bool{"php": true}
	w = httptest.NewRecorder()
	if err := wizardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if body = w.Body.String(); strings.Contains(body, "op-phpw-az") || strings.Contains(body, "presetKeys().length") || strings.Contains(body, ".ini/editor") {
		t.Error("tabs for features the user lacks are still rendered")
	}
}
