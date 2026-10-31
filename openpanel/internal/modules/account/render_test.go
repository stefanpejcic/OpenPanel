package account

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// baseLayout is a minimal web.LayoutData for rendering the authenticated pages in this package, distinct from loginPageData's pre-session standalone context
func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "account": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

// exercises the actual embedded template through html/template's executor - the only way to catch broken wiring, a typo'd field name, or a package-load panic before a real request hits it
func TestRenderLoginPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := loginPageData{
		Title:         "Login",
		BrandName:     "",
		CSRFToken:     "test-csrf-token",
		Locales:       localeOptions([]string{"en", "de"}),
		PasswordReset: "yes",
		T:             mgr.Translator("en"),
	}

	w := httptest.NewRecorder()
	if err := loginPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	for _, want := range []string{
		`name="username"`,
		`name="password"`,
		`test-csrf-token`,
		`Sign In`,
		`data-value="en"`,
		`data-value="de"`,
		`id="locale-search"`,
		`data-search="Deutsch German de"`,
		`/static/flags/gb.png`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered login page missing %q\n--- body ---\n%s", want, body)
		}
	}
}

func TestRenderLoginPageDarkLogo(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	base := loginPageData{Title: "Login", CSRFToken: "x", Locales: localeOptions([]string{"en"}), T: mgr.Translator("en"), Logo: "/light.svg"}

	render := func(d loginPageData) string {
		w := httptest.NewRecorder()
		if err := loginPage.Render(w, 200, d); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	// no dark logo: one logo, no mode classes
	body := render(base)
	if !strings.Contains(body, `src="/light.svg"`) || strings.Contains(body, "op-logo-dark") {
		t.Errorf("without logo_dark the card should show only the logo")
	}

	// card: both logos, CSS shows the one for the mode
	d := base
	d.LogoDark = "/dark.svg"
	body = render(d)
	for _, want := range []string{`src="/light.svg" class="h-16 op-logo-light"`, `src="/dark.svg" class="h-16 op-logo-dark"`} {
		if !strings.Contains(body, want) {
			t.Errorf("card login missing %q", want)
		}
	}

	// split: the brand panel is always dark, so it uses the dark logo
	d.LoginLayout = "split"
	body = render(d)
	if !strings.Contains(body, `<aside class="op-auth-side dark">`) || !strings.Contains(body[strings.Index(body, "op-auth-side"):strings.Index(body, "op-auth-main")], `src="/dark.svg"`) {
		t.Errorf("split login brand panel should use the dark logo")
	}
}

func TestRenderLoginPageTwofa(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := loginPageData{
		Title:        "Login",
		CSRFToken:    "tok",
		TwofaEnabled: true,
		UserID:       42,
		T:            mgr.Translator("en"),
	}

	w := httptest.NewRecorder()
	if err := loginPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, `name="twofa_code"`) {
		t.Errorf("expected 2FA code field in rendered output, got:\n%s", body)
	}
	if !strings.Contains(body, `value="42"`) {
		t.Errorf("expected user_id=42 hidden field, got:\n%s", body)
	}
	if strings.Contains(body, `name="password"`) {
		t.Error("password field should not render when TwofaEnabled is true")
	}
}

func TestRenderLoginPageErrorMessage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := loginPageData{
		Title:        "Login",
		CSRFToken:    "tok",
		ErrorMessage: "Invalid password. Please try again.",
		T:            mgr.Translator("en"),
	}

	w := httptest.NewRecorder()
	if err := loginPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "Invalid password. Please try again.") {
		t.Error("expected error message to appear in rendered output")
	}
}

func TestRenderAccountPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("username change not permitted", func(t *testing.T) {
		data := AccountPageData{LayoutData: baseLayout(mgr, "/account"), CurrentEmail: "test@example.com"}
		w := httptest.NewRecorder()
		if err := accountPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "test@example.com") {
			t.Error("expected current email in body")
		}
		// the username shows read-only and isn't submitted
		if strings.Contains(body, `name="username"`) || !strings.Contains(body, "readonly") {
			t.Error("expected a read-only username without a form field when PermitUsernameChangeByUser is false")
		}
	})

	t.Run("username change permitted", func(t *testing.T) {
		data := AccountPageData{LayoutData: baseLayout(mgr, "/account"), CurrentEmail: "test@example.com", PermitUsernameChangeByUser: true}
		w := httptest.NewRecorder()
		if err := accountPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `name="username"`) {
			t.Error("expected username field when PermitUsernameChangeByUser is true")
		}
	})

	t.Run("member since and passkeys", func(t *testing.T) {
		layout := baseLayout(mgr, "/account")
		layout.HostingPlanName = "Developer plus"
		layout.UserAllowed = map[string]bool{"dashboard": true, "account": true, "passkeys": true}
		since := time.Date(2026, 3, 14, 9, 30, 0, 0, time.UTC)
		for _, tc := range []struct {
			count int
			want  []string
			not   string
		}{
			{2, []string{"Developer plus · member since 14 Mar 2026", `<dt>Passkeys</dt><dd><a href="/account/passkeys">2</a></dd>`}, ""},
			{0, []string{`<a href="/account/passkeys">None</a>`}, ""},
			{-1, nil, "<dt>Passkeys</dt>"},
		} {
			w := httptest.NewRecorder()
			if err := accountPage.Render(w, 200, AccountPageData{LayoutData: layout, MemberSince: since, PasskeyCount: tc.count}); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			for _, want := range tc.want {
				if !strings.Contains(body, want) {
					t.Errorf("count %d: missing %q", tc.count, want)
				}
			}
			if tc.not != "" && strings.Contains(body, tc.not) {
				t.Errorf("count %d: unexpected %q", tc.count, tc.not)
			}
		}
	})

	t.Run("no registered date", func(t *testing.T) {
		if got := (AccountPageData{}).MemberSinceLabel(); got != "" {
			t.Errorf("MemberSinceLabel() = %q, want empty", got)
		}
	})
}

func TestRenderLocalePage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := LocalePageData{LayoutData: baseLayout(mgr, "/account/language"), Locales: localeOptions([]string{"en", "de", "sr"}), Current: "de"}
	w := httptest.NewRecorder()
	if err := localePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{`value="en"`, `value="de"`, "Deutsch", "German", "/static/flags/gb.png", "/static/flags/rs.png", `action="/account/language"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
	cardOf := func(code string) string {
		i := strings.Index(body, `value="`+code+`"`)
		if i < 0 {
			return ""
		}
		return body[i : i+strings.Index(body[i:], "</button>")]
	}
	if de := cardOf("de"); !strings.Contains(de, `aria-pressed="true"`) || !strings.Contains(de, "op-locale-on") {
		t.Error("expected the current locale's card to be highlighted")
	}
	if en := cardOf("en"); strings.Contains(en, `aria-pressed="true"`) || strings.Contains(en, "op-locale-on") {
		t.Error("only the current locale should be highlighted")
	}
	if !strings.Contains(body, `class="op-locales"`) || strings.Count(body, `name="locale"`) != 3 {
		t.Error("expected one card per locale in the locale grid")
	}
}

func TestLocaleOptionsFlags(t *testing.T) {
	for code, flag := range map[string]string{"en": "gb", "sr": "rs", "sv": "se", "ne": "np", "zh": "cn", "de": "de", "xx": "xx"} {
		if got := localeOptions([]string{code})[0].FlagCode; got != flag {
			t.Errorf("%s: flag %q, want %q", code, got, flag)
		}
	}
	if o := localeOptions([]string{"xx"})[0]; o.Native != "XX" {
		t.Errorf("unknown locale should fall back to its code, got %+v", o)
	}
}

func TestRenderTwofaPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("disabled, no setup started", func(t *testing.T) {
		data := TwofaPageData{LayoutData: baseLayout(mgr, "/account/2fa")}
		w := httptest.NewRecorder()
		if err := twofaPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Click to enable 2FA") {
			t.Error("expected enable button")
		}
		if strings.Contains(body, "qrcode") {
			t.Error("did not expect QR setup UI before setup starts")
		}
	})

	t.Run("setup started", func(t *testing.T) {
		data := TwofaPageData{LayoutData: baseLayout(mgr, "/account/2fa"), OTPSecret: "ABCDEFGHIJKLMNOP"}
		w := httptest.NewRecorder()
		if err := twofaPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "ABCDEFGHIJKLMNOP") || !strings.Contains(body, "Click to Cancel") {
			t.Error("expected OTP secret and cancel button during setup")
		}
	})

	t.Run("enabled", func(t *testing.T) {
		data := TwofaPageData{LayoutData: baseLayout(mgr, "/account/2fa"), TwofaEnabled: true, OTPSecret: "ABCDEFGHIJKLMNOP"}
		w := httptest.NewRecorder()
		if err := twofaPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Click to disable 2FA") {
			t.Error("expected disable button when enabled")
		}
	})
}

func TestRenderPasskeysPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := PasskeysPageData{LayoutData: baseLayout(mgr, "/account/passkeys")}
		w := httptest.NewRecorder()
		if err := passkeysPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "No passkeys registered yet.") {
			t.Error("expected empty state text")
		}
		for _, want := range []string{`x-data="opAddPasskey(`, `aria-labelledby="passkey-add-title"`, `id="passkey-name"`, `maxlength="100"`, `@click="show()"`} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
		if strings.Contains(body, "prompt(") {
			t.Error("the native prompt should be gone")
		}
	})

	t.Run("unavailable", func(t *testing.T) {
		data := PasskeysPageData{LayoutData: baseLayout(mgr, "/account/passkeys"), UnavailableReason: "Passkeys need HTTPS."}
		w := httptest.NewRecorder()
		if err := passkeysPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Passkeys need HTTPS.") || strings.Contains(body, `id="passkey-name"`) {
			t.Error("expected the reason and no add sheet")
		}
	})

	t.Run("with passkeys", func(t *testing.T) {
		data := PasskeysPageData{
			LayoutData: baseLayout(mgr, "/account/passkeys"),
			Passkeys:   []passkeyRow{{ID: 5, Name: "MacBook Touch ID", CreatedAt: "2026-08-01 10:00:00"}},
		}
		w := httptest.NewRecorder()
		if err := passkeysPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "MacBook Touch ID") || !strings.Contains(body, "/account/passkeys/5/delete") {
			t.Error("expected passkey row and delete form in body")
		}
	})
}

func TestRenderNotificationsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := NotificationsPageData{
		LayoutData: baseLayout(mgr, "/account/notifications"),
	}
	data.Notifications = []NotificationPref{{Key: "notify_password_change", Value: "1"}, {Key: "notify_account_login", Value: "0"}}
	data.Cards = buildNotificationCards(data.Notifications, nil)
	w := httptest.NewRecorder()
	if err := notificationsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="notify_password_change"`) || !strings.Contains(body, `name="notify_account_login"`) {
		t.Error("expected both preference checkboxes in body")
	}
	if !strings.Contains(body, "/docs/panel/account/notifications/#password-changed") {
		t.Error("expected docs link")
	}
}

func TestRenderFavoritesPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("feature disabled", func(t *testing.T) {
		layout := baseLayout(mgr, "/account/favorites")
		layout.UserAllowed = map[string]bool{"dashboard": true}
		data := FavoritesPageData{LayoutData: layout}
		w := httptest.NewRecorder()
		if err := favoritesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "not enabled by the Administrator") {
			t.Error("expected disabled-feature notice")
		}
	})

	t.Run("with favorites", func(t *testing.T) {
		layout := baseLayout(mgr, "/account/favorites")
		layout.UserAllowed = map[string]bool{"dashboard": true, "favorites": true}
		data := FavoritesPageData{
			LayoutData: layout,
			Favorites:  []FavoriteRow{{Name: "Dashboard", Link: "dashboard"}},
		}
		w := httptest.NewRecorder()
		if err := favoritesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Dashboard") || !strings.Contains(body, `href="/dashboard"`) {
			t.Error("expected favorite row with leading-slash link in body")
		}
	})
}

func TestRenderActiveSessionsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := ActiveSessionsPageData{LayoutData: baseLayout(mgr, "/account/sessions")}
		w := httptest.NewRecorder()
		if err := activeSessionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("with sessions and activity link", func(t *testing.T) {
		layout := baseLayout(mgr, "/account/sessions")
		layout.UserAllowed = map[string]bool{"dashboard": true, "activity": true}
		data := ActiveSessionsPageData{
			LayoutData: layout,
			Sessions:   []ActiveSession{{SessionToken: "abc123", IPAddress: "1.2.3.4", CreatedAt: "2026-08-03 10:00:00", ExpiresIn: "30m"}},
		}
		w := httptest.NewRecorder()
		if err := activeSessionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "1.2.3.4") || !strings.Contains(body, "/account/sessions/terminate/abc123") {
			t.Error("expected session row and terminate form in body")
		}
		if !strings.Contains(body, "/account/activity?search=1.2.3.4") {
			t.Error("expected activity log link when 'activity' is allowed")
		}
	})
}

func TestRenderActivityPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no logs", func(t *testing.T) {
		data := ActivityPageData{LayoutData: baseLayout(mgr, "/account/activity"), ActivityPageResult: ActivityPageResult{TotalPages: 1}}
		w := httptest.NewRecorder()
		if err := activityPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No logs yet.") {
			t.Error("expected empty state text")
		}
	})

	t.Run("with rows and pagination", func(t *testing.T) {
		result := ActivityPageResult{
			Rows: []ActivityLogRow{
				{Timestamp: "2026-08-03 10:00:00", IP: "1.2.3.4", User: "bob", Action: "changed password", Kind: ActivityKindSecurity},
				{Timestamp: "2026-08-03 10:01:00", IP: "1.2.3.4", User: "bob", Action: "deleted domain example.com", Kind: ActivityKindDanger},
			},
			Page: 1, ItemsPerPage: 100, TotalPages: 3, TotalLines: 250,
			PageEntries: buildPageEntries(1, 3),
		}
		data := ActivityPageData{LayoutData: baseLayout(mgr, "/account/activity"), ActivityPageResult: result}
		w := httptest.NewRecorder()
		if err := activityPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "bob") || !strings.Contains(body, "changed password") {
			t.Error("expected activity row in body")
		}
		if !strings.Contains(body, "Next") {
			t.Error("expected a Next link when not on the last page")
		}
		if !strings.Contains(body, "activity-kind--security") || !strings.Contains(body, "activity-kind--danger") {
			t.Error("expected kind icons for each row")
		}
	})

	t.Run("filters are kept in links", func(t *testing.T) {
		lines := []string{"2026-08-03 10:00:00  1.2.3.4 User bob deleted domain a.com"}
		a := &appctx.App{Config: config.Config{"activity_items_per_page": "1"}}
		result := paginateActivityLog(a, append(lines, lines...), ActivityFilter{Kind: ActivityKindDanger, From: "2026-08-01"}, 1)
		data := ActivityPageData{LayoutData: baseLayout(mgr, "/account/activity"), ActivityPageResult: result}
		w := httptest.NewRecorder()
		if err := activityPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `href="/account/activity?from=2026-08-01&amp;page=2&amp;type=danger"`) {
			t.Error("expected pagination link to keep filters")
		}
		if !strings.Contains(body, `dateRange(&#34;2026-08-01&#34;`) && !strings.Contains(body, `dateRange("2026-08-01"`) {
			t.Error("expected date picker prefilled with the from date")
		}
		if !strings.Contains(body, "Clear filters") || !strings.Contains(body, "Last 7 days") {
			t.Error("expected clear link and translated presets")
		}
	})

	t.Run("search term", func(t *testing.T) {
		result := ActivityPageResult{SearchTerm: "1.2.3.4", TotalPages: 1}
		data := ActivityPageData{LayoutData: baseLayout(mgr, "/account/activity"), ActivityPageResult: result}
		w := httptest.NewRecorder()
		if err := activityPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Search results for:") {
			t.Error("expected search-results header")
		}
	})
}

func TestRenderLoginHistoryPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := LoginHistoryPageData{LayoutData: baseLayout(mgr, "/account/login-history")}
		w := httptest.NewRecorder()
		if err := loginHistoryPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("with records and activity link", func(t *testing.T) {
		layout := baseLayout(mgr, "/account/login-history")
		layout.UserAllowed = map[string]bool{"dashboard": true, "activity": true}
		data := LoginHistoryPageData{
			LayoutData:    layout,
			LastLoginData: []appctx.LastLogin{{IP: "1.2.3.4", CountryCode: "US", LoginTime: "2026-08-03 10:00:00"}},
		}
		w := httptest.NewRecorder()
		if err := loginHistoryPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "1.2.3.4") || !strings.Contains(body, "/static/flags/us.png") {
			t.Error("expected IP and lowercased flag image path in body")
		}
		if !strings.Contains(body, "/account/activity?search=1.2.3.4") {
			t.Error("expected activity log link")
		}
	})
}

func TestRenderMCPPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no tokens", func(t *testing.T) {
		data := MCPPageData{LayoutData: baseLayout(mgr, "/account/mcp"), MCPURL: "https://panel.example.com/mcp"}
		w := httptest.NewRecorder()
		if err := mcpPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "No MCP tokens yet.") {
			t.Error("expected empty state text")
		}
		if !strings.Contains(body, "https://panel.example.com/mcp") {
			t.Error("expected mcp url in connection snippets")
		}
	})

	t.Run("with tokens and new token banner", func(t *testing.T) {
		data := MCPPageData{
			LayoutData: baseLayout(mgr, "/account/mcp"),
			MCPURL:     "https://panel.example.com/mcp",
			NewToken:   "op_mcp_abc123",
			Tokens: []mcptokens.Token{
				{
					ID: 1, Name: "Laptop", TokenPrefix: "op_mcp_ab", CreatedAt: "2026-08-03 10:00:00",
					LastUsedAt: sql.NullString{}, ReadOnly: true, ExpiresAt: sql.NullString{},
				},
				{
					ID: 2, Name: "Server", TokenPrefix: "op_mcp_cd", CreatedAt: "2026-08-01 09:00:00",
					LastUsedAt: sql.NullString{String: "2026-08-04 12:00:00", Valid: true}, ReadOnly: false,
					ExpiresAt: sql.NullString{String: "2026-09-01 00:00:00", Valid: true},
				},
			},
		}
		w := httptest.NewRecorder()
		if err := mcpPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "op_mcp_abc123") {
			t.Error("expected new token value in banner")
		}
		if !strings.Contains(body, "Laptop") || !strings.Contains(body, "Server") {
			t.Error("expected both token names in table")
		}
		if !strings.Contains(body, "/account/mcp/1/revoke") || !strings.Contains(body, "/account/mcp/2/revoke") {
			t.Error("expected revoke form actions for each token")
		}
	})
}

func TestLoginLocaleButtonShowsCurrent(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := loginPageData{Title: "Login", CSRFToken: "x", Locales: localeOptions([]string{"en", "de"}), T: mgr.Translator("de")}
	w := httptest.NewRecorder()
	if err := loginPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	btn := body[strings.Index(body, `id="locale-btn"`):strings.Index(body, `id="locale-list"`)]
	if !strings.Contains(btn, "Deutsch") || !strings.Contains(btn, "/static/flags/de.png") || strings.Contains(btn, "English") {
		t.Errorf("button should show the current language, got %s", btn)
	}
	if !strings.Contains(body, `data-value="de" data-search="Deutsch German de" aria-selected="true"`) {
		t.Error("current language should be marked selected in the list")
	}

	data.T = mgr.Translator("xx")
	w = httptest.NewRecorder()
	_ = loginPage.Render(w, 200, data)
	if !strings.Contains(w.Body.String(), "Change Language") {
		t.Error("unknown current locale should fall back to the Change Language label")
	}
}

func TestAlertWasDisabled(t *testing.T) {
	on := map[string]string{"notify_password_change": "1", "notify_password_change_notification_disabled": "1"}
	cases := []struct {
		name string
		old  map[string]string
		new  map[string]string
		want bool
	}{
		{"watched alert switched off", on, map[string]string{"notify_password_change": "0", "notify_password_change_notification_disabled": "1"}, true},
		{"both switched off together", on, map[string]string{"notify_password_change": "0", "notify_password_change_notification_disabled": "0"}, true},
		{"only the meta key switched off", on, map[string]string{"notify_password_change": "1", "notify_password_change_notification_disabled": "0"}, false},
		{"meta key was already off", map[string]string{"notify_password_change": "1", "notify_password_change_notification_disabled": "0"}, map[string]string{"notify_password_change": "0", "notify_password_change_notification_disabled": "0"}, false},
		{"nothing changed", on, on, false},
	}
	for _, c := range cases {
		if got := alertWasDisabled(c.old, c.new); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// field errors sit under their field, other messages in an alert coloured by kind
func TestRenderLoginPageMessagePlacement(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	cases := []struct {
		name string
		data loginPageData
		want []string
		not  []string
	}{
		{"wrong password", loginPageData{Username: "bob", ErrorMessage: "Invalid password. Please try again.", PasswordError: true},
			[]string{`id="password-error"`, `aria-describedby="password-error"`, `value="bob"`}, []string{`role="alert"`}},
		{"unknown user", loginPageData{Username: "bo", ErrorMessage: "Unrecognized account. Please check username.", UnrecognizedError: true},
			[]string{`id="username-error"`, `aria-invalid="true" aria-describedby="username-error"`}, []string{`role="alert"`, `id="password-error"`}},
		{"suspended", loginPageData{ErrorMessage: "Your account is suspended. Please contact support.", ErrorTone: "warning"},
			[]string{`class="op-alert is-warning" role="alert"`}, []string{`id="username-error"`}},
		{"captcha", loginPageData{ErrorMessage: "Captcha verification failed. Please try again."},
			[]string{`class="op-alert is-danger" role="alert"`}, nil},
		{"reset done", loginPageData{FirstFlash: &flash.Message{Category: "success", Text: "Your password has been reset. Please log in."}},
			[]string{`class="op-alert is-success" role="alert"`, "Your password has been reset."}, nil},
		{"wrong 2fa code", loginPageData{TwofaEnabled: true, UserID: 7, ErrorMessage: "Invalid 2FA code. Please try again."},
			[]string{`id="twofa-error"`, `autocomplete="one-time-code"`, `href="/login"`}, []string{`role="alert"`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.data.Title, tc.data.CSRFToken, tc.data.T = "Login", "tok", mgr.Translator("en")
			w := httptest.NewRecorder()
			if err := loginPage.Render(w, 200, tc.data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			for _, s := range tc.want {
				if !strings.Contains(body, s) {
					t.Errorf("expected %q", s)
				}
			}
			for _, s := range tc.not {
				if strings.Contains(body, s) {
					t.Errorf("did not expect %q", s)
				}
			}
		})
	}
}

func TestRenderLoginPageLayouts(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	for layout, wantSplit := range map[string]bool{"": false, "card": false, "split": true} {
		w := httptest.NewRecorder()
		data := loginPageData{Title: "Login", CSRFToken: "tok", LoginLayout: layout, T: mgr.Translator("en")}
		if err := loginPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render %q: %v", layout, err)
		}
		body := w.Body.String()
		if got := strings.Contains(body, `class="op-auth op-scope is-split"`) && strings.Contains(body, `class="op-auth-side dark"`); got != wantSplit {
			t.Errorf("layout %q: split markup present = %v, want %v", layout, got, wantSplit)
		}
		if !strings.Contains(body, `name="username"`) {
			t.Errorf("layout %q: sign-in form missing", layout)
		}
	}
}
