package dynamicdns

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// fakeToken repeats a word to token length, so fixtures never look like real secrets
func fakeToken(word string) string { return strings.Repeat(word, 16)[:16] }

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "domains": true, "dns": true, "dynamic_dns": true, "ssl": true, "redirects": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestRenderDynamicDNSPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	render := func(t *testing.T, data DynamicDNSPageData, selectHost string) string {
		t.Helper()
		fillHostSummary(&data, selectHost, now)
		w := httptest.NewRecorder()
		if err := dynamicDNSPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}
	domains := []string{"example.com", "shop.example.com"}

	t.Run("no domains", func(t *testing.T) {
		body := render(t, DynamicDNSPageData{LayoutData: baseLayout(mgr, "/domains/dynamic-dns"), BaseURL: "https://203.0.113.10:2083/"}, "")
		if !strings.Contains(body, "No domains yet") || strings.Contains(body, "New Host") {
			t.Error("expected the no-domain state without a New Host button")
		}
	})

	t.Run("empty", func(t *testing.T) {
		body := render(t, DynamicDNSPageData{LayoutData: baseLayout(mgr, "/domains/dynamic-dns"), DomainNames: domains, BaseURL: "https://203.0.113.10:2083/"}, "")
		for _, want := range []string{"No Dynamic DNS hosts yet", "home.example.com", "Add your first host", "How it works", "New Dynamic DNS host", `name="action" value="create"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in empty state", want)
			}
		}
		if strings.Contains(body, "Connect a device") {
			t.Error("connect card should be hidden without hosts")
		}
	})

	t.Run("with hosts", func(t *testing.T) {
		hosts := []DynHost{
			newDynHost("example.com", DynDNSEntry{LineNumber: 12, Subdomain: "home", Record: "198.51.100.23", Token: fakeToken("home"), LastUpdated: "2026-10-29T08:12:05Z"}),
			newDynHost("example.com", DynDNSEntry{LineNumber: 13, Subdomain: "nas", Record: "0.0.0.0", Token: fakeToken("nas"), LastUpdated: "2026-10-29T10:40:11Z"}),
			newDynHost("shop.example.com", DynDNSEntry{LineNumber: 9, Subdomain: "office", Record: "10.0.2.2", Token: fakeToken("office"), LastUpdated: "2026-10-26T13:28:42Z"}),
		}
		body := render(t, DynamicDNSPageData{LayoutData: baseLayout(mgr, "/domains/dynamic-dns"), Hosts: hosts, DomainNames: domains, BaseURL: "https://203.0.113.10:2083/"}, "nas.example.com")
		for _, want := range []string{
			"home.example.com", "198.51.100.23", "3 hours ago", "2026-10-29 08:12:05 UTC",
			"Waiting for first call", "Not set yet", "2 days ago",
			"/dynamic-dns/update?token=home••••••••",
			"3 hosts across 2 domains",
			"Connect a device", "Router / NAS", "400 IPv4 address required",
			"Set up a device", "Delete this host?", `name="action" value="edit"`, `name="action" value="delete"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
		if !strings.Contains(body, ", 1, ") {
			t.Error("expected ?host=nas.example.com to preselect the second host")
		}
	})
}

func TestTimeAgo(t *testing.T) {
	tr := i18n.NewManager(t.TempDir(), nil).Translator("en")
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	cases := map[string]string{
		"2026-10-29T11:59:30Z": "just now",
		"2026-10-29T11:59:00Z": "1 minute ago",
		"2026-10-29T11:15:00Z": "45 minutes ago",
		"2026-10-29T09:00:00Z": "3 hours ago",
		"2026-10-28T12:00:00Z": "1 day ago",
		"2026-08-29T12:00:00Z": "2 months ago",
		"2024-10-29T12:00:00Z": "2 years ago",
		"garbage":              "",
	}
	for in, want := range cases {
		if got := timeAgo(tr, in, now); got != want {
			t.Errorf("timeAgo(%q) = %q, want %q", in, got, want)
		}
	}
}
