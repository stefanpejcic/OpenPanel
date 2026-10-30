package dns

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "dns": true, "domains": true, "ssl": true, "redirects": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestRenderDNSListPage(t *testing.T) {
	webtest.ModernOnly(t)
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

	t.Run("zone list", func(t *testing.T) {
		data := DNSPageData{
			LayoutData: baseLayout(mgr, "/domains/edit-dns-zone"),
			Domains:    []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "shop.example.com"}},
			Zones: []ZoneSummary{
				{Domain: "example.com", Records: 9, Types: []TypeCount{{"A", 3}, {"CNAME", 3}, {"MX", 1}, {"TXT", 2}}},
				{Domain: "shop.example.com", Records: 4, Types: []TypeCount{{"A", 2}, {"TXT", 2}}},
			},
			ZoneRecords: 13,
		}
		w := httptest.NewRecorder()
		if err := dnsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"2 zones", "13 records", `href="/domains/edit-dns-zone/shop.example.com"`, "CNAME 3", "/domains/export-dns-zone/example.com"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in zone list", want)
			}
		}
	})
}

func TestSummarizeZone(t *testing.T) {
	z := summarizeZone("example.com", `$TTL 14400
@ IN SOA ns1.example.com. admin.example.com. ( 1 7200 3600 1209600 86400 )
@ 14400 IN A 1.2.3.4
www 14400 IN A 1.2.3.4
@ 14400 IN TXT "v=spf1 -all"
`)
	if z.Records != 3 || len(z.Types) != 2 || z.Types[0] != (TypeCount{"A", 2}) || z.Types[1] != (TypeCount{"TXT", 1}) {
		t.Errorf("unexpected summary: %+v", z)
	}
}

func TestRenderDNSTablePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DNSPageData{
		LayoutData: baseLayout(mgr, "/domains/edit-dns-zone/example.com"),
		Domain:     "example.com", ViewMode: "table", Serial: "2026010101",
		Domains: []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "example.org"}},
		Rows: []ZoneRow{
			{LineNumber: 5, EndLineNumber: 5, Name: "www", TTL: "14400", Type: "A", RawLine: "www 14400 IN A 1.2.3.4", DisplayValue: "1.2.3.4"},
			{LineNumber: 6, EndLineNumber: 6, Name: "@", TTL: "14400", Type: "TXT", RawLine: `@ 14400 IN TXT "v=spf1 -all"`, DisplayValue: "v=spf1 -all", Comment: "spf record"},
			{LineNumber: 7, EndLineNumber: 7, Name: "@", TTL: "14400", Type: "MX", RawLine: "@ 14400 IN MX 10 mail.example.com.", DisplayValue: "10 mail.example.com."},
			{LineNumber: 8, EndLineNumber: 8, Name: "mail", TTL: "14400", Type: "CNAME", RawLine: "mail 14400 IN CNAME example.com.", DisplayValue: "example.com."},
			{LineNumber: 9, EndLineNumber: 9, Name: "@", TTL: "14400", Type: "AAAA", RawLine: "@ 14400 IN AAAA 2001:db8::24", DisplayValue: "2001:db8::24"},
			{LineNumber: 10, EndLineNumber: 10, Name: "default._domainkey", TTL: "14400", Type: "TXT", RawLine: `default._domainkey 14400 IN TXT "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1SU1LfVLPHCozMxH2Mo4lgOEePzNm0tRgeLezV6ffAt0gunVTLw7onLRnrq0/IzW7yWR7QkrmBL7jTKEn5u+qKhbwKfBstIs+bMY2Zkp18gnTxKLxoS2tFczGkPLPgizskuemMghRniWaoLcyehkd3qqGElvW/VDL5AaWTg0nLVkjRo9z+40RQzuVaE8AkAFmxZzow3x+VJYKdx/tVxYz3txGTHMwUKgRoLe3t3QCJyfnwHp8E8eWU"`, DisplayValue: "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1SU1LfVLPHCozMxH2Mo4lgOEePzNm0tRgeLezV6ffAt0gunVTLw7onLRnrq0/IzW7yWR7QkrmBL7jTKEn5u+qKhbwKfBstIs+bMY2Zkp18gnTxKLxoS2tFczGkPLPgizskuemMghRniWaoLcyehkd3qqGElvW/VDL5AaWTg0nLVkjRo9z+40RQzuVaE8AkAFmxZzow3x+VJYKdx/tVxYz3txGTHMwUKgRoLe3t3QCJyfnwHp8E8eWU"},
		},
		TotalRecords: 6,
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
