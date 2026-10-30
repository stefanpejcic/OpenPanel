package emails

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sieveparser"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "emails": true, "webmail": true, "phpmyadmin": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestRenderAccountsPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AccountsPageData{
		LayoutData: baseLayout(mgr, "/emails"),
		Rows: []EmailListRow{
			parseEmailListRow("* info@demo.rs ( 0 / 2.0G ) [45%]"),
			parseEmailListRow("* sales@demo.rs ( 1.7G / 2.0G ) [86%]"),
			parseEmailListRow("* billing@demo.rs ( 1.9G / 2.0G ) [96%]"),
			parseEmailListRow("* old@demo.rs ( 12M / 1.0G ) [1%]"),
		},
		TotalCount: 4,
	}
	data.Rows[1].HasFilter = true
	data.Rows[3].Suspended = true
	data.UserAllowed["email_aliases"], data.UserAllowed["email_deliverability"] = true, true
	data.fillAccountsSummary(50, []AliasEntry{{Source: "hello@demo.rs", Targets: []string{"info@demo.rs"}}, {Source: "billing@demo.rs", Targets: []string{"info@demo.rs", "sales@demo.rs"}}})
	w := httptest.NewRecorder()
	if err := accountsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "info@demo.rs") {
		t.Error("expected address in body")
	}
	for _, want := range []string{"4 of 50", "1.7 GB / 2 GB", "12 MB / 1 GB", `data-href="/emails/edit/info@demo.rs"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
	if data.AliasCount != 2 || data.AliasInboxes != 2 || data.NearQuotaCount != 2 {
		t.Errorf("summary = %d aliases, %d inboxes, %d near quota", data.AliasCount, data.AliasInboxes, data.NearQuotaCount)
	}
}

func TestHumanQuota(t *testing.T) {
	for in, want := range map[string]string{"0": "0 MB", "12M": "12 MB", "1.7G": "1.7 GB", "2.0G": "2 GB", "512K": "512 KB", "~": "~"} {
		if got, _ := humanQuota(in); got != want {
			t.Errorf("humanQuota(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderAccountsPageEmpty(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AccountsPageData{LayoutData: baseLayout(mgr, "/emails")}
	w := httptest.NewRecorder()
	if err := accountsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "No emails yet.") {
		t.Error("expected empty state text")
	}
}

func newEmailData(mgr *i18n.Manager, maxGB float64, count, limit int) NewEmailPageData {
	data := NewEmailPageData{
		LayoutData:           baseLayout(mgr, "/emails/new"),
		Domains:              []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "shop.example.org"}},
		MaxEmailQuotaNumeric: maxGB, AllocatedUnit: "G", MaxQuotaBytes: maxGB * (1 << 30),
		EmailCount: count, EmailLimit: limit, Taken: []string{"Info@example.com"},
		MailHost: "panel.example.com", MailSecure: true, Wait: maxGB > 0,
	}
	data.HostingPlanName = "Developer plus"
	data.fill()
	return data
}

func TestRenderNewEmailPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(data NewEmailPageData) string {
		w := httptest.NewRecorder()
		if err := newEmailPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	t.Run("capped plan", func(t *testing.T) {
		data := newEmailData(mgr, 2, 3, 500)
		data.Domain = "shop.example.org"
		data.fill()
		body := render(data)
		for _, want := range []string{`value="shop.example.org" selected`, "2 GB", "Plan maximum", "256 MB", "3 of 500", "panel.example.com:993", "panel.example.com:465", "Developer plus", `name="gb"`, "info@example.com", "Mail apps will use"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
		if strings.Contains(body, "<fieldset disabled") || strings.Contains(body, "reached max number") {
			t.Error("form should be enabled under the limit")
		}
	})

	t.Run("plan limit reached", func(t *testing.T) {
		body := render(newEmailData(mgr, 2, 5, 5))
		if !strings.Contains(body, "<fieldset disabled") || !strings.Contains(body, "Error: reached max number of email accounts on the hosting plan.") {
			t.Error("expected the disabled form and the plan limit alert")
		}
	})

	t.Run("no quota cap", func(t *testing.T) {
		data := newEmailData(mgr, 0, 0, 0)
		data.MailSecure = false
		body := render(data)
		for _, want := range []string{"Unlimited", "no limit", "panel.example.com:143", "STARTTLS"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
	})

	t.Run("no domains", func(t *testing.T) {
		data := newEmailData(mgr, 2, 0, 500)
		data.Domains = nil
		if body := render(data); !strings.Contains(body, "No domains yet") || strings.Contains(body, `name="password"`) {
			t.Error("expected the no-domains empty state instead of the form")
		}
	})
}

func TestRenderAccountsPageAfterCreate(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	rows := []EmailListRow{parseEmailListRow("* info@demo.rs ( 0 / 2.0G ) [0%]"), parseEmailListRow("* hello@demo.rs ( 0 / 1.0G ) [0%]")}
	rows[1].IsNew = true
	data := AccountsPageData{LayoutData: baseLayout(mgr, "/emails"), Rows: rows, TotalCount: 2,
		Connect: AccountsConnect{Created: "hello@demo.rs", Welcome: true, Host: "panel.example.com", Secure: true}}
	data.fillAccountsSummary(2, nil)
	w := httptest.NewRecorder()
	if err := accountsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{`class="op-em-new"`, "op-em-tag-new", "Your mailbox is ready. Connect a mail app", `data-op-copy="panel.example.com"`, "/emails/configuration/", "You&#39;ve used all 2 email accounts on your plan", `aria-disabled="true"`, "Plan limit reached: 2 of 2 email accounts used."} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
	if strings.Contains(body, `href="/emails/new" id="showAddEmailFormBtn"`) {
		t.Error("New Email should not link anywhere at the plan limit")
	}
}

func TestRenderSingleAccountPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := SingleAccountPageData{
		LayoutData: baseLayout(mgr, "/emails/edit/info@demo.rs"),
		Quota:      parseSingleEmailQuota("* info@demo.rs ( 0 / 2.0G ) [45%]"),
		ServerIP:   "1.2.3.4", DedicatedIP: "1.2.3.4",
		SendRestriction: "ACCEPT", ReceiveRestriction: "ACCEPT",
	}
	w := httptest.NewRecorder()
	if err := singleAccountPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "info@demo.rs") {
		t.Error("expected address in body")
	}
}

func TestRenderDeletePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with address", func(t *testing.T) {
		data := DeletePageData{LayoutData: baseLayout(mgr, "/emails/delete/a@b.com"), Address: "a@b.com"}
		w := httptest.NewRecorder()
		if err := deletePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "a@b.com") {
			t.Error("expected address in body")
		}
	})

	t.Run("select address", func(t *testing.T) {
		data := DeletePageData{LayoutData: baseLayout(mgr, "/emails/delete"), Addresses: []string{"a@b.com", "c@d.com"}}
		w := httptest.NewRecorder()
		if err := deletePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "c@d.com") {
			t.Error("expected addresses list in body")
		}
	})
}

func TestRenderInfoPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := InfoPageData{LayoutData: baseLayout(mgr, "/emails/info/a@b.com"), Address: "a@b.com", Scheme: "https", Hostname: "mail.example.com"}
	w := httptest.NewRecorder()
	if err := infoPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "mail.example.com") {
		t.Error("expected hostname in body")
	}
}

func TestRenderAliasesPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AliasesPageData{
		LayoutData: baseLayout(mgr, "/emails/aliases"),
		AliasList: []AliasEntry{
			{Source: "alias@b.com", Targets: []string{"real@b.com"}},
			{Source: "team@b.com", Targets: []string{"ana@b.com", "marko@b.com", "real@b.com"}},
			{Source: "hello@b.com", Targets: []string{"info@demo.rs"}},
		},
	}
	w := httptest.NewRecorder()
	if err := aliasesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "alias@b.com") || !strings.Contains(w.Body.String(), "real@b.com") {
		t.Error("expected alias row in body")
	}
}

func TestRenderAliasesCreateSheet(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	domains := []appctx.Domain{{DomainURL: "b.com"}, {DomainURL: "c.org"}}

	t.Run("closed", func(t *testing.T) {
		data := AliasesPageData{LayoutData: baseLayout(mgr, "/emails/aliases"), Domains: domains, AliasList: []AliasEntry{{Source: "sales@b.com", Targets: []string{"x@y.com"}}}}
		w := httptest.NewRecorder()
		if err := aliasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{`action="/emails/aliases/new"`, `name="username"`, `name="domain"`, `name="target"`, `open: false`, `&#34;sales@b.com&#34;`, `@click.prevent="$dispatch(`, `<option class="punycode" value="c.org">`} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
	})

	t.Run("reopened after an error", func(t *testing.T) {
		data := AliasesPageData{LayoutData: baseLayout(mgr, "/emails/aliases"), Domains: domains, NewOpen: true, NewDomain: "c.org", NewUser: "info"}
		w := httptest.NewRecorder()
		if err := aliasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `open: true`) || !strings.Contains(body, `domain: &#34;c.org&#34;`) || !strings.Contains(body, `user: &#34;info&#34;`) {
			t.Error("expected the sheet open with domain and username")
		}
	})

	t.Run("no domains", func(t *testing.T) {
		data := AliasesPageData{LayoutData: baseLayout(mgr, "/emails/aliases")}
		w := httptest.NewRecorder()
		if err := aliasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if strings.Contains(w.Body.String(), `action="/emails/aliases/new"`) {
			t.Error("no sheet without domains")
		}
	})
}

func TestRenderAliasNewPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AliasNewPageData{LayoutData: baseLayout(mgr, "/emails/aliases/new"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
	w := httptest.NewRecorder()
	if err := aliasNewPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "/emails/aliases?new=1") {
		t.Error("expected the link to the sheet")
	}
}

func TestRenderAliasDetailPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with entry", func(t *testing.T) {
		entry := &AliasEntry{Source: "alias@b.com", Targets: []string{"real@b.com"}}
		data := AliasDetailPageData{LayoutData: baseLayout(mgr, "/emails/aliases/alias@b.com"), Email: "alias@b.com", Entry: entry}
		w := httptest.NewRecorder()
		if err := aliasDetailPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "real@b.com") {
			t.Error("expected target in body")
		}
	})

	t.Run("no entry", func(t *testing.T) {
		data := AliasDetailPageData{LayoutData: baseLayout(mgr, "/emails/aliases/alias@b.com"), Email: "alias@b.com", Entry: nil}
		w := httptest.NewRecorder()
		if err := aliasDetailPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No destinations configured") {
			t.Error("expected empty state")
		}
	})
}

func TestRenderAliasDeletePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with email", func(t *testing.T) {
		data := AliasDeletePageData{LayoutData: baseLayout(mgr, "/emails/aliases/delete/a@b.com"), Email: "a@b.com"}
		w := httptest.NewRecorder()
		if err := aliasDeletePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "a@b.com") {
			t.Error("expected email in body")
		}
	})

	t.Run("select alias", func(t *testing.T) {
		data := AliasDeletePageData{
			LayoutData: baseLayout(mgr, "/emails/aliases/delete"),
			AliasList:  []AliasEntry{{Source: "alias@b.com", Targets: []string{"real@b.com"}}},
		}
		w := httptest.NewRecorder()
		if err := aliasDeletePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "alias@b.com") {
			t.Error("expected alias option in body")
		}
	})
}

func TestRenderDefaultAddressPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("select domain", func(t *testing.T) {
		data := DefaultAddressPageData{LayoutData: baseLayout(mgr, "/emails/default/"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := defaultAddressPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain option in body")
		}
	})

	t.Run("with current alias", func(t *testing.T) {
		data := DefaultAddressPageData{LayoutData: baseLayout(mgr, "/emails/default/example.com"), Domain: "example.com", Current: "catchall@example.com"}
		w := httptest.NewRecorder()
		if err := defaultAddressPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "catchall@example.com") {
			t.Error("expected current alias in body")
		}
	})
}

func TestRenderDeliverabilityPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DeliverabilityPageData{LayoutData: baseLayout(mgr, "/emails/deliverability"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
	w := httptest.NewRecorder()
	if err := deliverabilityPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "example.com") {
		t.Error("expected domain row in body")
	}
}

func TestRenderDeliverabilityDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DeliverabilityDomainPageData{LayoutData: baseLayout(mgr, "/emails/deliverability/example.com"), Domain: "example.com"}
	w := httptest.NewRecorder()
	if err := deliverabilityDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "example.com") {
		t.Error("expected domain in body")
	}
}

func TestRenderFilterPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("select address", func(t *testing.T) {
		data := FilterPageData{LayoutData: baseLayout(mgr, "/emails/filter"), Addresses: []string{"a@b.com"}}
		w := httptest.NewRecorder()
		if err := filterPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "a@b.com") {
			t.Error("expected address option in body")
		}
	})

	t.Run("gui mode with parsed filters", func(t *testing.T) {
		data := FilterPageData{
			LayoutData: baseLayout(mgr, "/emails/filter/a@b.com/gui"), Email: "a@b.com", ViewMode: "gui",
			ParsedFilters: sieveparser.Parse(`if header :contains "Subject" "invoice" { fileinto "Finance"; }`),
		}
		w := httptest.NewRecorder()
		if err := filterPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "filterPage()") {
			t.Error("expected alpine filterPage() component in body")
		}
	})

	t.Run("raw mode", func(t *testing.T) {
		data := FilterPageData{
			LayoutData: baseLayout(mgr, "/emails/filter/a@b.com/raw"), Email: "a@b.com", ViewMode: "raw",
			RawContent: `if true { stop; }`, SieveFile: "/var/mail/b.com/a/home/.dovecot.sieve",
		}
		w := httptest.NewRecorder()
		if err := filterPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "if true { stop; }") {
			t.Error("expected raw content in body")
		}
	})
}

func TestRenderImportPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := struct{ web.LayoutData }{baseLayout(mgr, "/emails/import")}
	w := httptest.NewRecorder()
	if err := importPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderConfirmImportPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ConfirmImportPageData{
		LayoutData:   baseLayout(mgr, "/emails/import"),
		ValidUsers:   []ImportRow{{Email: "a@b.com", Password: "pw", Quota: "1G"}},
		InvalidUsers: []ImportRow{{Email: "c@d.com", DomainValid: false}},
	}
	w := httptest.NewRecorder()
	if err := confirmImportPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "a@b.com") || !strings.Contains(body, "c@d.com") {
		t.Error("expected both valid and invalid rows in body")
	}
}
