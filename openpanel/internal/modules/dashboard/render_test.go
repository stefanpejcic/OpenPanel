package dashboard

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// exercises the full base.html + partials + dashboard.html + includes template tree through html/template's real executor - the only way to catch a broken {{template}} wiring, a mismatched field name, or a nil-map access before a real request hits it. Covers a user with a broad, representative feature set so most conditional branches actually render.
func TestRenderDashboardPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	userAllowed := map[string]bool{
		"dashboard": true, "websites": true, "wordpress": true, "domains": true, "mysql": true,
		"emails": true, "ftp": true, "docker": true, "account": true,
		"twofa": true, "favorites": true, "services": true, "info": true,
		"filemanager": true, "phpmyadmin": true, "ai_assistant": true, "autoinstaller": true, "dns": true, "ssl": true,
	}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			BrandName:       "Test Panel",
			CSRFToken:       "test-csrf-token",
			PanelDir:        "ltr",
			PanelVersion:    "1.0.0",
			MenuStyle:       "modern",
			NavItems:        web.BuildSidebarNav(userAllowed, nil, "/dashboard"),
			NavTrail:        []web.NavLink{{Label: "Dashboard"}},
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
		Sections:           buildDashboardSections(mgr.Translator("en"), userAllowed, nil, "modern", nil),
		TourShow:           true,
		TwofaEnabled:       false,
		TwofaNag:           "yes",
		TwofaStatusMessage: twofaStatusMessage(mgr.Translator("en"), false),
		IPAddress:          "1.2.3.4",
		LastIP:             "5.6.7.8",
		LastLoginAgo:       "2 hours ago",
		LastLoginTime:      "2026-10-29 12:00:00",
		IPCountyFlag:       "yes",
		NS1:                "ns1.example.net",
		NS2:                "ns2.example.net",
		HowToGuides:        "yes",
		HowToTopics:        []HowToTopic{{Link: "https://example.com/a", Title: "Point a domain to this server"}, {Link: "https://example.com/b", Title: "Install WordPress"}, {Link: "https://example.com/c", Title: "Restore a backup"}},
		UserWebsitesCount:  3,
		Websites: []DashboardSite{
			{Name: "wp.example.com", Type: "WordPress", PHP: "8.3", Href: "/website?domain=wp.example.com"},
			{Name: "shop.example.com", Type: "PHP", PHP: "8.2", Href: "/website?domain=shop.example.com"},
			{Name: "api.example.com", Type: "Node.js", Href: "/website?domain=api.example.com", Suspended: true},
			{Name: "node.example.com", Type: "Node.js", Href: "/website?domain=node.example.com", Service: "nodeapp", Port: "3000", State: "running"},
			{Name: "n8n.example.com", Type: "n8n", Href: "/website?domain=n8n.example.com", Service: "n8n", Port: "5678", State: "exited"},
		},
		MainDomainsCount: 2,
		DBUsage:          1,
		EmailCount:       5,
		FTPCount:         1,
		WebsitesLimit:    10,
		DomainsLimit:     10,
		DBLimit:          5,
		EmailLimit:       50,
		FTPLimit:         5,
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
		`<span aria-current="page">Dashboard</span>`,
		`id="dashboard-menu"`,
		"PHP 8.3",
		"--color-hue-green",
		`href="/files/wp.example.com"`,
		`5.6.7.8</span> · <span title="2026-10-29 12:00:00">2 hours ago</span>`,
		`<span class="op-hint" title="2026-10-29 12:00:00">2 hours ago</span>`,
		`<em title="2026-10-29 12:00:00">2 hours ago</em>`,
		`class="op-content op-icons-line"`,
		`<h2>General information</h2>`,
		"nodeapp:3000",
		"n8n:5678",
		`<span class="op-status is-success">Running</span>`,
		`<span class="op-status is-muted">Stopped</span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered dashboard missing %q", want)
		}
	}
}

// covers a plan with an upsell target: features the plan doesn't grant but the upsell plan does should render greyed-out (sidebar link + dashboard icon) with an upgrade prompt, instead of being omitted like a plain non-upsell-eligible feature
func TestRenderDashboardPageUpsellEligible(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	userAllowed := map[string]bool{"dashboard": true, "websites": true}
	upsellAllowed := map[string]bool{"wordpress": true, "docker": true}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			CSRFToken:       "tok",
			PanelDir:        "ltr",
			NavGroups:       web.BuildClassicSidebarNav(userAllowed, upsellAllowed, "/dashboard"),
			UserAllowed:     userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed),
			UpsellAllowed:   upsellAllowed,
			UpsellPlanName:  "Business",
			UpsellURL:       "https://example.com/upgrade",
			CurrentUsername: "upselluser",
			RequestPath:     "/dashboard",
			AdminPort:       "2087",
			T:               mgr.Translator("en"),
		},
		Sections: buildDashboardSections(mgr.Translator("en"), userAllowed, upsellAllowed, "classic", nil),
	}

	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	for _, want := range []string{
		"sidebar-toggle-all", // classic sidebar keeps Expand all
		"websites-menu",
		"Upgrade to",
		"Business",
		`href="/dashboard/upgrade"`,
		"is-locked",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered dashboard missing %q for an upsell-eligible feature", want)
		}
	}

	// the sidebar's Upgrade link is Enterprise-only
	upgradeSection := `<div id="tour-upgrade">`
	if strings.Contains(body, upgradeSection) {
		t.Error("did not expect the sidebar Upgrade link without Enterprise")
	}
	data.IsEnterprise = true
	w = httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body = w.Body.String()
	i := strings.Index(body, upgradeSection)
	if i < 0 || !strings.Contains(body[i:], `href="/dashboard/upgrade"`) || !strings.Contains(body[i:], "Business</a>") {
		t.Error("expected the sidebar Upgrade link to Business pointing at /dashboard/upgrade on Enterprise")
	}
}

// covers the opposite end: a user with almost no features enabled, so most conditional sections render empty - the branch most likely to hit a nil map / missing key panic that a feature-rich test run wouldn't exercise
func TestRenderDashboardPageMinimalUser(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			CSRFToken:       "tok",
			PanelDir:        "ltr",
			NavGroups:       web.BuildClassicSidebarNav(userAllowed, nil, "/dashboard"),
			UserAllowed:     userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed),
			CurrentUsername: "minimal",
			RequestPath:     "/dashboard",
			AdminPort:       "2087",
			T:               mgr.Translator("en"),
		},
		Sections: buildDashboardSections(mgr.Translator("en"), userAllowed, nil, "classic", nil),
	}

	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "minimal") {
		t.Error("expected username to appear in rendered output")
	}
}

// exercises the flash message stack and impersonation banner, both absent from the other two tests
func TestRenderDashboardPageWithFlashAndImpersonation(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true}

	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title:           "Dashboard",
			CSRFToken:       "tok",
			PanelDir:        "ltr",
			NavGroups:       web.BuildClassicSidebarNav(userAllowed, nil, "/dashboard"),
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
		Sections: buildDashboardSections(mgr.Translator("en"), userAllowed, nil, "classic", nil),
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

// a new account sees the setup wizard over the dashboard
func TestRenderDashboardPageOnboarding(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{
		"dashboard": true, "domains": true, "websites": true, "wordpress": true, "mysql": true, "php": true,
		"change_ws": true, "change_db": true, "varnish": true, "backups": true, "waf": true, "ip_blocker": true,
		"twofa": true, "passkeys": true, "filemanager": true, "autoinstaller": true,
	}
	data := DashboardPageData{
		LayoutData: web.LayoutData{
			Title: "Dashboard", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
			MenuStyle: "modern", NavItems: web.BuildSidebarNav(userAllowed, nil, "/dashboard"),
			NavTrail: []web.NavLink{{Label: "Dashboard"}}, UserAllowed: userAllowed,
			UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "newuser",
			HostingPlanName: "Starter", AvatarType: "letter", RequestPath: "/dashboard", AdminPort: "2087",
			T: mgr.Translator("en"),
		},
		Sections:           buildDashboardSections(mgr.Translator("en"), userAllowed, nil, "modern", nil),
		OnboardingShow:     true,
		TwofaStatusMessage: twofaStatusMessage(mgr.Translator("en"), false),
		IPAddress:          "1.2.3.4",
	}
	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "onboardingWizard(") {
		t.Error("expected the onboarding wizard")
	}
}
