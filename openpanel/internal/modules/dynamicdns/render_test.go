package dynamicdns

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "dynamic_dns": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderDynamicDNSPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := DynamicDNSPageData{
			LayoutData: baseLayout(mgr, "/domains/dynamic-dns"),
			Domains:    []appctx.Domain{{DomainURL: "example.com"}},
			BaseURL:    "https://185.7.32.112:2083/",
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
		entries := []DynDNSEntry{
			{LineNumber: 5, Subdomain: "home", TTL: "300", Type: "A", Record: "1.2.3.4", Token: "abcdefghijklmnop", LastUpdated: "2026-01-01T00:00:00Z", Index: 0},
		}
		data := DynamicDNSPageData{
			LayoutData:    baseLayout(mgr, "/domains/dynamic-dns"),
			Domains:       []appctx.Domain{{DomainURL: "example.com"}},
			DomainEntries: []DomainEntries{{DomainName: "example.com", Entries: entries}},
			AllEntries:    entries,
			BaseURL:       "https://185.7.32.112:2083/",
		}
		w := httptest.NewRecorder()
		if err := dynamicDNSPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "home") || !strings.Contains(body, "1.2.3.4") {
			t.Error("expected entry fields in body")
		}
		if !strings.Contains(body, "abcdefgh…") {
			t.Error("expected truncated token preview")
		}
		if !strings.Contains(body, "2026-01-01 00:00:00 UTC") {
			t.Error("expected T/Z-replaced last-updated timestamp")
		}
	})
}
