package dynamicdns

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := DynamicDNSPageData{
			LayoutData:  baseLayout(mgr, "/domains/dynamic-dns"),
			DomainNames: []string{"example.com"},
			BaseURL:     "https://203.0.113.10:2083/",
		}
		w := httptest.NewRecorder()
		if err := dynamicDNSPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No Dynamic DNS entries yet") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with entries", func(t *testing.T) {
		entry := DynDNSEntry{LineNumber: 5, Subdomain: "home", TTL: "300", Type: "A", Record: "1.2.3.4", Token: "abcdefghijklmnop", LastUpdated: "2026-01-01T00:00:00Z"}
		data := DynamicDNSPageData{
			LayoutData:  baseLayout(mgr, "/domains/dynamic-dns"),
			DomainNames: []string{"example.com"},
			Hosts:       []DynHost{newDynHost("example.com", entry)},
			BaseURL:     "https://203.0.113.10:2083/",
		}
		w := httptest.NewRecorder()
		if err := dynamicDNSPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"home", "1.2.3.4", "dynamic-dns/update?token=abcdefghijklmnop"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
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
