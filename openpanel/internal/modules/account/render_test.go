package account

import (
	"database/sql"
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apidocs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// baseLayout is a minimal web.LayoutData for rendering the authenticated
// pages in this package (account settings, locale, 2FA, passkeys, etc.) -
// distinct from loginPageData above, which is the pre-session login page's
// own standalone context.
func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "account": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

// TestRenderLoginPage exercises the actual embedded template through
// html/template's executor - the only way to catch a broken {{template
// "content" .}} wiring, a typo'd field name, or an accidental package-load
// panic (like the go:embed "_login.html" gotcha this test would have
// caught) before a real request hits it.
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
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered login page missing %q\n--- body ---\n%s", want, body)
		}
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
		if strings.Contains(body, `id="username"`) {
			t.Error("did not expect username field when PermitUsernameChangeByUser is false")
		}
	})

	t.Run("username change permitted", func(t *testing.T) {
		data := AccountPageData{LayoutData: baseLayout(mgr, "/account"), CurrentEmail: "test@example.com", PermitUsernameChangeByUser: true}
		w := httptest.NewRecorder()
		if err := accountPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `id="username"`) {
			t.Error("expected username field when PermitUsernameChangeByUser is true")
		}
	})
}

func TestRenderLocalePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no current locale", func(t *testing.T) {
		data := LocalePageData{LayoutData: baseLayout(mgr, "/account/language"), Locales: []string{"en", "de"}}
		w := httptest.NewRecorder()
		if err := localePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `value="en"`) || !strings.Contains(body, `value="de"`) {
			t.Error("expected both locale options in body")
		}
		if strings.Contains(body, "Current") {
			t.Error("did not expect a Current marker when no locale is selected")
		}
	})

	t.Run("with current locale", func(t *testing.T) {
		data := LocalePageData{LayoutData: baseLayout(mgr, "/account/language"), Locales: []string{"en", "de"}, Current: "de"}
		w := httptest.NewRecorder()
		if err := localePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "(Current)") {
			t.Error("expected the current locale to be marked")
		}
	})
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
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := PasskeysPageData{LayoutData: baseLayout(mgr, "/account/passkeys")}
		w := httptest.NewRecorder()
		if err := passkeysPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No passkeys registered yet.") {
			t.Error("expected empty state text")
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
		Notifications: []NotificationPref{
			{Key: "notify_password_change", Value: "1", Label: " password change"},
			{Key: "notify_account_login", Value: "0", Label: " account login"},
		},
	}
	w := httptest.NewRecorder()
	if err := notificationsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="notify_password_change"`) || !strings.Contains(body, `name="notify_account_login"`) {
		t.Error("expected both preference checkboxes in body")
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
				{Timestamp: "2026-08-03 10:00:00", IP: "1.2.3.4", User: "bob", Action: "changed password"},
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

func TestRenderAPIDocsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	data := APIDocsPageData{
		LayoutData:    baseLayout(mgr, "/account/api"),
		EndpointsJSON: template.JS(apidocs.EndpointsJSON), //nolint:gosec // test uses the real embedded documentation data
	}
	w := httptest.NewRecorder()
	if err := apiDocsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="api-endpoints-data"`) {
		t.Error("expected embedded endpoints data script tag")
	}
	if !strings.Contains(body, "API Reference") {
		t.Error("expected page title")
	}
	if len(apidocs.EndpointsJSON) == 0 {
		t.Fatal("apidocs.EndpointsJSON is empty")
	}
	if !strings.Contains(body, `"group"`) {
		t.Error("expected endpoint group JSON embedded in page")
	}
}
