package domains

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{
		"dashboard": true, "domains": true, "docroot": true, "redirects": true, "ssl": true,
		"edit_vhost": true, "domain_suspend": true, "capitalize_domains": true, "filemanager": true,
		"dns": true, "waf": true, "php": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestRenderDomainsPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := DomainsPageData{LayoutData: baseLayout(mgr, "/domains")}
		w := httptest.NewRecorder()
		if err := domainsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet.") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with rows", func(t *testing.T) {
		rows := []DomainRow{
			{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/example.com", PHPVersion: "8.3", SiteCount: 2, HTTPS: "Automatic", Status: "Not Suspended", DNS: true, WAF: "On", SSL: SSLCertInfo{Expires: "2027-01-14", Issuer: "Let's Encrypt"}},
			{DomainID: 2, DomainURL: "shop.example.com", Docroot: "/var/www/html/shop.example.com", PHPVersion: "8.2", IsSubdomain: true, HTTPS: "Automatic", Status: "Not Suspended", DNS: true, WAF: "DetectionOnly", SSL: SSLCertInfo{Problem: "expiring", Expires: "2026-11-03", Issuer: "Let's Encrypt"}},
			{DomainID: 3, DomainURL: "mystore.io", Docroot: "/var/www/html/mystore.io", PHPVersion: "8.3", HTTPS: "Automatic", Status: "Not Suspended", DNS: true, Cloudflare: true, SSL: SSLCertInfo{Problem: "missing"}},
			{DomainID: 4, DomainURL: "blog.example.com", Docroot: "/var/www/html/blog.example.com", PHPVersion: "7.4", IsSubdomain: true, HTTPS: "Custom", Status: "Suspended", SuspendComment: "unpaid invoice", SSL: SSLCertInfo{Problem: "expired", Expires: "2026-09-30"}},
		}
		for i := range rows {
			rows[i].Filter = map[string]string{"status": "active", "php": rows[i].PHPVersion, "ssl": "auto", "type": "domain", "redirect": "no"}
		}
		layout := baseLayout(mgr, "/domains")
		layout.BulkActions = domainsBulkActions(layout.T, []string{"7.4", "8.2", "8.3"})
		data := DomainsPageData{LayoutData: layout, Domains: rows, TotalDomains: 4, TotalPages: 1, CurrentPage: 1, Filters: buildDomainFilters(layout.T, layout.UserAllowed, rows)}
		w := httptest.NewRecorder()
		if err := domainsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "example.com") {
			t.Error("expected domain row")
		}
		if !strings.Contains(body, "unpaid invoice") {
			t.Error("expected suspend comment for suspended domain")
		}
		if !strings.Contains(body, `<div class="op-eyebrow">4 total</div>`) {
			t.Error("expected plain total without a plan limit")
		}
	})

	t.Run("plan limit eyebrow", func(t *testing.T) {
		data := DomainsPageData{LayoutData: baseLayout(mgr, "/domains"), TotalDomains: 8, MainDomains: 6, DomainsLimit: 10, TotalPages: 1, CurrentPage: 1}
		w := httptest.NewRecorder()
		if err := domainsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `<div class="op-eyebrow">6 of 10 on your plan</div>`) {
			t.Error("expected the plan usage eyebrow")
		}
	})
}

func TestRenderNewDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := NewDomainPageData{LayoutData: baseLayout(mgr, "/domains/new"), Sites: []appctx.Domain{{DomainURL: "example.com", Docroot: "/var/www/html/example.com"}}}
	w := httptest.NewRecorder()
	if err := newDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "addDomainForm") {
		t.Error("expected add-domain form")
	}
}

func TestRenderDeleteDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no blockers", func(t *testing.T) {
		data := DeleteDomainPageData{LayoutData: baseLayout(mgr, "/domains/delete"), DomainURL: "example.com", Docroot: "/var/www/html/example.com"}
		w := httptest.NewRecorder()
		if err := deleteDomainPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `>Delete Domain <svg`) {
			t.Error("expected delete button when no sites/subdomains block it")
		}
	})

	t.Run("blocked by sites", func(t *testing.T) {
		data := DeleteDomainPageData{
			LayoutData: baseLayout(mgr, "/domains/delete"), DomainURL: "example.com",
			SiteCount: 1, Sites: []SiteRow{{SiteName: "example.com"}},
		}
		w := httptest.NewRecorder()
		if err := deleteDomainPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if strings.Contains(w.Body.String(), `>Delete Domain <svg`) {
			t.Error("expected delete button hidden when sites exist")
		}
	})
}

func TestRenderSuspendUnsuspendPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("suspend selector", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/suspend"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := suspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain selector option")
		}
	})

	t.Run("suspend confirm", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/suspend"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := suspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "confirm-suspend") {
			t.Error("expected confirm button")
		}
	})

	t.Run("unsuspend confirm", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/unsuspend"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := unsuspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "confirm-unsuspend") {
			t.Error("expected confirm button")
		}
	})
}

func TestRenderDocrootPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DocrootPageData{LayoutData: baseLayout(mgr, "/domains/docroot"), DomainName: "example.com", Docroot: "/var/www/html/example.com"}
	w := httptest.NewRecorder()
	if err := docrootPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "/var/www/html/example.com") {
		t.Error("expected current docroot value")
	}
}

func TestRenderDomainLogsPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	files := []DomainLogFile{{Domain: "example.com", Size: 15874, Exists: true}, {Domain: "empty.com", Exists: true}, {Domain: "gone.com"}}
	render := func(t *testing.T, data DomainLogsPageData, status int) string {
		t.Helper()
		w := httptest.NewRecorder()
		if err := domainLogsPage.Render(w, status, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}
	mustHave := func(t *testing.T, body string, wants ...string) {
		t.Helper()
		for _, want := range wants {
			if !strings.Contains(body, want) {
				t.Errorf("rendered page missing %q", want)
			}
		}
	}

	t.Run("picker", func(t *testing.T) {
		body := render(t, DomainLogsPageData{LayoutData: baseLayout(mgr, "/domains/log"), Domains: files}, 200)
		mustHave(t, body, "Choose a domain", `data-href="/domains/log/example.com"`, `data-bytes="15874"`, ">Empty<", ">No log<", "3 domains")
	})

	t.Run("missing log", func(t *testing.T) {
		body := render(t, DomainLogsPageData{LayoutData: baseLayout(mgr, "/domains/log/gone.com"), Domains: files, Missing: "gone.com"}, 404)
		mustHave(t, body, `class="op-alert is-danger"`, "Log file not found for domain", "gone.com", "Pick one of your domains below.")
	})

	t.Run("empty log", func(t *testing.T) {
		data := buildDomainLogsData("empty.com", nil, url.Values{}, 1000, 10000)
		data.LayoutData, data.Domains = baseLayout(mgr, "/domains/log/empty.com"), files
		body := render(t, data, 200)
		mustHave(t, body, "No requests logged yet", `href="http://empty.com"`)
		if strings.Contains(body, "Download raw") {
			t.Error("empty log shouldn't offer a download")
		}
	})

	t.Run("with entries", func(t *testing.T) {
		data := buildDomainLogsData("example.com", readAccessLog(t), url.Values{"status": {"4"}}, 1000, 10000)
		data.LayoutData, data.Domains = baseLayout(mgr, "/domains/log/example.com"), files
		body := render(t, data, 200)
		mustHave(t, body,
			`x-data="opAccessLogs()"`, `href="/domains/log/example.com/raw"`, "Download raw", `<option value="example.com" class="punycode" selected>`,
			`class="op-filter is-set" href="/domains/log/example.com" data-keep-view title="Client error" aria-current="true"`, `4xx <span class="op-count">1</span>`,
			`href="/domains/log/example.com?status=5"`, `onclick="opSortToggle(event, this, sortTableColumn)"`, `data-sort-key="dur"`, `data-sort-col="status" data-sort-value="403"`,
			`data-sort-col="ts" data-sort-value="1793259377.651421"`, "op-badge op-mono al-code is-warning", "sqlmap-test", "js/table-sort.js",
			`id="al-detail-tpl"`, "Copy line", "Formatted", "Raw", "al-sortsel", "Showing 1 - 1 out of 1 rows (8 in the log)",
		)
		if !strings.Contains(body, `window.AL_ROWS = [{"raw":`) {
			t.Error("row details not embedded")
		}
	})

	t.Run("no match", func(t *testing.T) {
		data := buildDomainLogsData("example.com", readAccessLog(t), url.Values{"q": {"zzz"}}, 1000, 10000)
		data.LayoutData, data.Domains = baseLayout(mgr, "/domains/log/example.com"), files
		body := render(t, data, 200)
		mustHave(t, body, "No requests match", "Nothing contains “zzz”.", "Clear filters", `value="zzz"`)
	})

	t.Run("pages", func(t *testing.T) {
		data := buildDomainLogsData("example.com", readAccessLog(t), url.Values{"page": {"2"}}, 3, 10000)
		data.LayoutData, data.Domains = baseLayout(mgr, "/domains/log/example.com"), files
		body := render(t, data, 200)
		mustHave(t, body, `class="op-pager"`, `href="/domains/log/example.com?page=3"`, `aria-current="page">2<`, "Showing 4 - 6 out of 8 rows", "Show all 8 rows")
	})
}

func TestRenderCapitalizePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CapitalizePageData{LayoutData: baseLayout(mgr, "/domains/capitalize/example.com"), DomainURL: "example.com", CapitalizedDomain: "ExampleDomain.com"}
	w := httptest.NewRecorder()
	if err := capitalizePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "ExampleDomain.com") {
		t.Error("expected capitalized domain value")
	}
}

func TestRenderVhostPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := VhostPageData{LayoutData: baseLayout(mgr, "/domains/vhosts"), WebServerPreference: "nginx", Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := vhostPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Nginx") {
			t.Error("expected capitalized web server name")
		}
	})

	t.Run("editor", func(t *testing.T) {
		data := VhostPageData{LayoutData: baseLayout(mgr, "/domains/vhosts"), DomainName: "example.com", WebServerPreference: "nginx", VhostContent: "server { listen 80; }"}
		w := httptest.NewRecorder()
		if err := vhostPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "server { listen 80; }") {
			t.Error("expected vhost content in textarea")
		}
	})
}

func TestRenderRedirectPages(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(t *testing.T, domains []appctx.Domain, existing map[string]string, open string) string {
		t.Helper()
		rules := map[string]redirectRule{}
		for d, to := range existing {
			rules[d] = redirectRule{To: to}
		}
		data := buildRedirectPageData(baseLayout(mgr, "/domains/redirect"), domains, rules, open)
		w := httptest.NewRecorder()
		if err := redirectPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}
	domains := []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "old-shop.com"}, {DomainURL: "blog.example.com"}, {DomainURL: "abc.onion"}}

	t.Run("no domains", func(t *testing.T) {
		body := render(t, nil, nil, "")
		if !strings.Contains(body, "No domains yet") || strings.Contains(body, "New Redirect") {
			t.Error("expected the no-domain state without a New Redirect button")
		}
	})

	t.Run("empty", func(t *testing.T) {
		body := render(t, domains, nil, "")
		for _, want := range []string{"No redirects", "Add a redirect", "0 of 3 domains redirected", "How redirects work", `<option value="old-shop.com">`, `name="return_to" value="redirects"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in empty state", want)
			}
		}
		if strings.Contains(body, "abc.onion") {
			t.Error(".onion domains can't redirect and shouldn't be listed")
		}
	})

	t.Run("with redirects", func(t *testing.T) {
		body := render(t, domains, map[string]string{"old-shop.com": "https://example.com/shop", "blog.example.com": "http://medium.com/@me"}, "old-shop.com")
		for _, want := range []string{
			"2 of 3 domains redirected", "Search redirects", "and www.old-shop.com", `href="https://example.com/shop"`, "http://medium.com/@me",
			"302 Temporary", "2 redirects", "1 domain has no redirect", "Delete this redirect?", `action="/domains/redirect/delete"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
		if strings.Contains(body, `<option value="old-shop.com">`) {
			t.Error("a redirected domain shouldn't be offered for a new redirect")
		}
		if !strings.Contains(body, "&#34;old-shop.com&#34;, {") {
			t.Error("expected ?domain= to open that domain's form")
		}
	})
}

func TestRenderDomainsRedirectColumn(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	rows := []DomainRow{
		{DomainID: 1, DomainURL: "old-shop.com", Status: "Not Suspended", HTTPS: "Automatic", RedirectURL: "http://example.com/"},
		{DomainID: 2, DomainURL: "example.com", Status: "Not Suspended", HTTPS: "Automatic"},
	}
	data := DomainsPageData{LayoutData: baseLayout(mgr, "/domains"), Domains: rows, TotalDomains: 2, TotalPages: 1, CurrentPage: 1}
	w := httptest.NewRecorder()
	if err := domainsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if strings.Contains(body, `pattern="https://.*"`) || strings.Count(body, `pattern="https?://.*"`) != 2 {
		t.Error("both inline redirect editors should take http:// and https://")
	}
	if !strings.Contains(body, `href="/domains/redirect?domain=old-shop.com"`) || !strings.Contains(body, "Edit redirect") {
		t.Error("expected the row menu's redirect item")
	}
	if !strings.Contains(body, `onsubmit="return confirm(`) {
		t.Error("expected removing a redirect to ask first")
	}
}

func TestRenderSSLPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no certificate", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No Certificate!") {
			t.Error("expected no-certificate state")
		}
	})

	t.Run("autossl", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), DomainName: "example.com", CurrentSetting: "autossl", Keys: "cert info here"}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Auto SSL") {
			t.Error("expected Auto SSL status")
		}
	})

	t.Run("overview", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), Domains: []appctx.Domain{{DomainURL: "example.com"}}, Certs: []SSLRow{
			{Domain: "example.com", HTTPS: "Automatic", SSL: SSLCertInfo{Expires: "2027-01-14", Issuer: "Let's Encrypt", Days: 78}},
			{Domain: "shop.example.com", HTTPS: "Automatic", SSL: SSLCertInfo{Problem: "expiring", Expires: "2026-11-04", Issuer: "Let's Encrypt", Days: 6}},
			{Domain: "old-site.org", HTTPS: "Custom", SSL: SSLCertInfo{Problem: "expired", Expires: "2026-09-12", Issuer: "Sectigo", Days: -46}},
			{Domain: "mystore.io", HTTPS: "Automatic", SSL: SSLCertInfo{Problem: "missing"}},
			{Domain: "plain.example.org", HTTPS: "None"},
		}}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"78 days left", "Expiring soon", "Next expiry 2027-01-14", "Not issued yet", `name="action" value="generate"`, "No SSL",
			`x-data="opInstallCert({`, `open: false`, `aria-labelledby="cert-install-title"`, `name="return_to" value="list"`, `name="certificate"`, `name="private_key"`, `/domains/ssl/check`, `href="/domains/ssl?install=shop.example.com"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q on the overview", want)
			}
		}
	})

	t.Run("overview with the install sheet open", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), Install: "shop.example.com", Certs: []SSLRow{{Domain: "example.com", HTTPS: "Automatic"}, {Domain: "shop.example.com", HTTPS: "Automatic"}}}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `open: true`) || !strings.Contains(body, `domain: &#34;shop.example.com&#34;`) {
			t.Error("expected the sheet open for shop.example.com")
		}
	})
}
