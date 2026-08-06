package dns

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "dns": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderDNSListPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no domains", func(t *testing.T) {
		data := DNSPageData{LayoutData: baseLayout(mgr, "/domains/edit-dns-zone")}
		w := httptest.NewRecorder()
		if err := dnsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with zone-having domains", func(t *testing.T) {
		data := DNSPageData{
			LayoutData: baseLayout(mgr, "/domains/edit-dns-zone"),
			Domains:    []appctx.Domain{{DomainURL: "example.com"}},
		}
		w := httptest.NewRecorder()
		if err := dnsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain option")
		}
	})
}

func TestRenderDNSTablePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DNSPageData{
		LayoutData: baseLayout(mgr, "/domains/edit-dns-zone/example.com"),
		Domain:     "example.com", ViewMode: "table", Serial: "2026010101",
		Rows: []ZoneRow{
			{LineNumber: 5, EndLineNumber: 5, Name: "www", TTL: "14400", Type: "A", RawLine: "www 14400 IN A 1.2.3.4", DisplayValue: "1.2.3.4"},
			{LineNumber: 6, EndLineNumber: 6, Name: "@", TTL: "14400", Type: "TXT", RawLine: `@ 14400 IN TXT "v=spf1 -all"`, DisplayValue: "v=spf1 -all", Comment: "spf record"},
		},
		TotalRecords: 2,
		Issues:       []HealthIssue{{ID: "dns-zone-syntax:example.com", Severity: "error", Message: "boom"}},
	}
	w := httptest.NewRecorder()
	if err := dnsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "1.2.3.4") || !strings.Contains(body, "v=spf1 -all") {
		t.Error("expected zone record values in body")
	}
	if !strings.Contains(body, "2026010101") {
		t.Error("expected serial number in body")
	}
	if !strings.Contains(body, "reportHealthIssues") {
		t.Error("expected health issue script block")
	}
}

func TestRenderDNSCodePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DNSPageData{
		LayoutData: baseLayout(mgr, "/domains/edit-dns-zone/example.com"),
		Domain:     "example.com", ViewMode: "code", ZoneContent: "$TTL 14400\n@ IN SOA ns1.example.com. admin.example.com. (\n  2026010101 )\n",
	}
	w := httptest.NewRecorder()
	if err := dnsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "2026010101") {
		t.Error("expected zone content in textarea")
	}
}
