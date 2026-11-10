package emails

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sieveparser"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AccountsPageData{
		LayoutData: baseLayout(mgr, "/emails"),
		Rows: []EmailListRow{
			parseEmailListRow("* info@demo.rs ( 0 / 2.0G ) [45%]"),
		},
		TotalCount: 1,
	}
	w := httptest.NewRecorder()
	if err := accountsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "info@demo.rs") {
		t.Error("expected address in body")
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
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := NewEmailPageData{
		LayoutData:           baseLayout(mgr, "/emails/new"),
		Domains:              []appctx.Domain{{DomainURL: "example.com"}},
		MaxEmailQuotaNumeric: 5, AllocatedUnit: "G",
	}
	w := httptest.NewRecorder()
	if err := newEmailPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "example.com") {
		t.Error("expected domain option in body")
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

func TestRenderAliasNewPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AliasNewPageData{LayoutData: baseLayout(mgr, "/emails/aliases/new"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
	w := httptest.NewRecorder()
	if err := aliasNewPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "example.com") {
		t.Error("expected domain option in body")
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

	t.Run("domain list", func(t *testing.T) {
		data := DefaultAddressPageData{
			LayoutData: baseLayout(mgr, "/emails/default/"),
			Domains:    []appctx.Domain{{DomainURL: "example.com"}},
			Picks:      []web.DomainPick{{Domain: "example.com", Detail: "catchall@example.com"}},
		}
		w := httptest.NewRecorder()
		if err := defaultAddressPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `href="/emails/default/example.com"`) || !strings.Contains(body, "catchall@example.com") {
			t.Error("expected domain row with its catch-all target")
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
		data := FilterPageData{LayoutData: baseLayout(mgr, "/emails/filter"), Addresses: []string{"a@b.com"}, Picks: []web.DomainPick{{Domain: "a@b.com", Badge: "Active", Tone: "success"}}}
		w := httptest.NewRecorder()
		if err := filterPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `href="/emails/filter/a@b.com"`) || !strings.Contains(body, " address</p>") {
			t.Error("expected address row with an edit link and the address count")
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

func TestRenderAccountsPageCreated(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(created CreatedMailbox, webmail bool) string {
		layout := baseLayout(mgr, "/emails")
		layout.UserAllowed["webmail"] = webmail
		w := httptest.NewRecorder()
		if err := accountsPage.Render(w, 200, AccountsPageData{LayoutData: layout, Created: created}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	body := render(CreatedMailbox{Address: "info@demo.rs", Welcome: true, Conn: mailConnectionFor(true, "mail.demo.rs")}, true)
	for _, want := range []string{"Connect a mail app", "mail.demo.rs", "993", "465", "SSL/TLS", "welcome email", "/emails/configuration/apple/info@demo.rs", "/webmail/info@demo.rs"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}

	for _, unwanted := range []string{"STARTTLS", "?ssl=", "sent unencrypted"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unexpected %q with SSL on", unwanted)
		}
	}

	body = render(CreatedMailbox{Address: "info@demo.rs", Conn: mailConnectionFor(false, "1.2.3.4")}, false)
	for _, want := range []string{"1.2.3.4", "143", "587", "No encryption", "sent unencrypted"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q with no SSL", want)
		}
	}
	for _, unwanted := range []string{"STARTTLS", ">993<", ">465<", "SSL/TLS"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unexpected %q with no SSL", unwanted)
		}
	}
	if strings.Contains(body, "welcome email") {
		t.Error("welcome line shown without welcome=1")
	}
	if strings.Contains(body, "/webmail/info@demo.rs") {
		t.Error("webmail button shown without the webmail feature")
	}
	if strings.Contains(render(CreatedMailbox{}, true), "Connect a mail app") {
		t.Error("section shown with nothing created")
	}
}

func TestCreatedMailboxFromRequest(t *testing.T) {
	owned := map[string]bool{"demo.rs": true}
	cases := map[string]CreatedMailbox{
		"created=a%40demo.rs&welcome=1": {Address: "a@demo.rs", Welcome: true},
		"created=a%40demo.rs":           {Address: "a@demo.rs"},
		"created=a%40other.rs":          {},
		"created=%3Cscript%3E":          {},
		"":                              {},
	}
	for query, want := range cases {
		r := httptest.NewRequest("GET", "/emails?"+query, nil)
		if got := createdMailboxFromRequest(r, owned); got != want {
			t.Errorf("%q: got %+v, want %+v", query, got, want)
		}
	}
}
