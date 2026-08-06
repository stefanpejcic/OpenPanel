package dashboard

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// TestRenderDashboardPage exercises the full base.html + partials +
// dashboard.html + includes template tree through html/template's real
// executor - the only way to catch a broken {{template}} wiring, a
// mismatched field name, or a nil-map access before a real request hits
// it. Covers a user with a broad, representative feature set so most
// conditional branches (sidebar groups, icon sections, usage widgets,
// twofa nag, custom section) actually render.
func TestRenderDashboardPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	userAllowed := map[string]bool{
		"dashboard": true, "websites": true, "wordpress": true, "domains": true, "mysql": true,
		"emails": true, "ftp": true, "docker": true, "account": true,
		"twofa": true, "favorites": true, "services": true, "info": true,
		"filemanager": true, "phpmyadmin": true,
	}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			BrandName:       "Test Panel",
			CSRFToken:       "test-csrf-token",
			PanelDir:        "ltr",
			PanelVersion:    "1.0.0",
			NavGroups:       web.BuildSidebarNav(userAllowed, "/dashboard"),
			UserAllowed:     userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed),
			IsEnterprise:    true,
			CurrentUsername: "testuser",
			HostingPlanName: "Pro Plan",
			AvatarType:      "letter",
			RequestPath:     "/dashboard",
			AdminPort:       "2087",
			T:               mgr.Translator("en"),
		},
		Sections:           buildDashboardSections(userAllowed),
		TourShow:           true,
		TwofaEnabled:       false,
		TwofaNag:           "yes",
		TwofaStatusMessage: twofaStatusMessage(mgr.Translator("en"), false),
		IPAddress:          "1.2.3.4",
		LastIP:             "5.6.7.8",
		IPCountyFlag:       "yes",
		UserWebsitesCount:  3,
		MainDomainsCount:   2,
		DBUsage:            1,
		EmailCount:         5,
		FTPCount:           1,
		WebsitesLimit:      10,
		DomainsLimit:       10,
		DBLimit:            5,
		EmailLimit:         50,
		FTPLimit:           5,
	}

	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	for _, want := range []string{
		"Test Panel",
		"testuser",
		"Pro Plan",
		"test-csrf-token",
		`id="dashboard-sortable-area"`,
		`id="dashboard_usage"`,
		`id="dashboard_info"`,
		`id="dashboard_twofa"`,
		"websites-menu",
		"mysql-menu",
		"docker-menu",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered dashboard missing %q", want)
		}
	}
}

// TestRenderDashboardPageMinimalUser covers the opposite end: a user with
// almost no features enabled, so most conditional sections render empty -
// this is the branch most likely to hit a nil map / missing key panic that
// a feature-rich test run wouldn't exercise.
func TestRenderDashboardPageMinimalUser(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			CSRFToken:       "tok",
			PanelDir:        "ltr",
			NavGroups:       web.BuildSidebarNav(userAllowed, "/dashboard"),
			UserAllowed:     userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed),
			CurrentUsername: "minimal",
			RequestPath:     "/dashboard",
			AdminPort:       "2087",
			T:               mgr.Translator("en"),
		},
		Sections: buildDashboardSections(userAllowed),
	}

	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "minimal") {
		t.Error("expected username to appear in rendered output")
	}
}

// TestRenderDashboardPageWithFlashAndImpersonation exercises the flash
// message stack and impersonation banner, both of which are absent from
// the other two tests.
func TestRenderDashboardPageWithFlashAndImpersonation(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			CSRFToken:       "tok",
			PanelDir:        "ltr",
			NavGroups:       web.BuildSidebarNav(userAllowed, "/dashboard"),
			UserAllowed:     userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed),
			CurrentUsername: "impersonated-user",
			RequestPath:     "/dashboard",
			AdminPort:       "2087",
			Impersonating:   true,
			Flashes: []web.FlashDisplay{
				{Index: 1, ZClass: "z-10", Text: "Something happened"},
			},
			T: mgr.Translator("en"),
		},
		Sections: buildDashboardSections(userAllowed),
	}

	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "impersonate-banner") {
		t.Error("expected impersonation banner to render")
	}
	if !strings.Contains(body, "Something happened") {
		t.Error("expected flash message to render")
	}
}
