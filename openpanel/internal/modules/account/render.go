package account

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// pageFiles is the standard authenticated-layout template set, shared by every page in this file - unlike loginPage in login.go, which renders standalone with no session/sidebar yet
var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var (
	accountPage        = loadPage("user/account.html")
	localePage         = loadPage("user/locale.html")
	twofaPage          = loadPage("user/twofa_settings.html")
	passkeysPage       = loadPage("user/passkeys.html")
	notificationsPage  = loadPage("user/notifications.html")
	favoritesPage      = loadPage("user/favorites.html")
	activeSessionsPage = loadPage("user/active_sessions.html")
	activityPage       = loadPage("user/activity.html")
	loginHistoryPage   = loadPage("user/loginlog.html")
	mcpPage            = loadPage("user/mcp.html")
	apiSwaggerPage     = loadPage("user/api_swagger.html")
)

// AccountPageData is user/account.html's template context.
type AccountPageData struct {
	web.LayoutData
	PermitUsernameChangeByUser bool
	CurrentEmail               string
}

func renderAccountPage(a *appctx.App, w http.ResponseWriter, r *http.Request, permitUsernameChange string) {
	layout, injected, err := web.BuildLayoutData(a, w, r, "Settings")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	currentEmail, _ := injected["current_email"].(string)
	data := AccountPageData{LayoutData: layout, PermitUsernameChangeByUser: permitUsernameChange == "yes", CurrentEmail: currentEmail}
	if err := accountPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - settings template render error: %v", err)
	}
}

// LocalePageData is user/locale.html's template context.
type LocalePageData struct {
	web.LayoutData
	Locales []localeOption
	Current string
}

func renderLocalePage(a *appctx.App, w http.ResponseWriter, r *http.Request, locales []string, current string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change Language")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if current == "" {
		current = layout.T.Locale()
	}
	data := LocalePageData{LayoutData: layout, Locales: localeOptions(locales), Current: current}
	if err := localePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - locale template render error: %v", err)
	}
}

// TwofaPageData is user/twofa_settings.html's template context.
type TwofaPageData struct {
	web.LayoutData
	TwofaEnabled bool
	OTPSecret    string
	TwofaIssuer  string
}

func renderTwofaPage(a *appctx.App, w http.ResponseWriter, r *http.Request, twofaEnabled bool, otpSecret string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Two-Factor Authentication")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := TwofaPageData{LayoutData: layout, TwofaEnabled: twofaEnabled, OTPSecret: otpSecret, TwofaIssuer: twofaIssuerName(a, r.Context())}
	if err := twofaPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - twofa template render error: %v", err)
	}
}

// PasskeysPageData is user/passkeys.html's template context.
type PasskeysPageData struct {
	web.LayoutData
	Passkeys          []passkeyRow
	UnavailableReason string
}

func renderPasskeysPage(a *appctx.App, w http.ResponseWriter, r *http.Request, passkeys []passkeyRow, unavailableReason string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Passkeys")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PasskeysPageData{LayoutData: layout, Passkeys: passkeys, UnavailableReason: unavailableReason}
	if err := passkeysPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - passkeys template render error: %v", err)
	}
}

// NotificationsPageData is user/notifications.html's template context.
type NotificationsPageData struct {
	web.LayoutData
	Notifications []NotificationPref
	Cards         []NotificationCard
	Email         string
}

// NotificationCard is one event on the page, Subs are its extra options shown under the switch
type NotificationCard struct {
	Key, Title, Description, Group, DocsURL string
	On                                      bool
	Subs                                    []NotificationCard
}

func buildNotificationCards(prefs []NotificationPref) []NotificationCard {
	values := make(map[string]bool, len(prefs))
	for _, p := range prefs {
		values[p.Key] = p.Value == "1"
	}
	var cards []NotificationCard
	for _, d := range notificationDefs {
		c := NotificationCard{Key: d.Key, Title: d.Title, Description: d.Description, Group: d.Group, On: values[d.Key]}
		if d.Parent == "" {
			c.DocsURL = "https://openpanel.com/docs/panel/account/notifications/#" + d.Anchor
			cards = append(cards, c)
			continue
		}
		for i := range cards {
			if cards[i].Key == d.Parent {
				cards[i].Subs = append(cards[i].Subs, c)
			}
		}
	}
	return cards
}

func renderNotificationsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, prefs []NotificationPref, email string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Notification Preferences")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := NotificationsPageData{LayoutData: layout, Notifications: prefs, Cards: buildNotificationCards(prefs), Email: email}
	if err := notificationsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - notifications template render error: %v", err)
	}
}

// FavoritesPageData is user/favorites.html's template context.
type FavoritesPageData struct {
	web.LayoutData
	Favorites []FavoriteRow
}

func renderFavoritesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, favorites []FavoriteRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Favorites")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FavoritesPageData{LayoutData: layout, Favorites: favorites}
	if err := favoritesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - favorites template render error: %v", err)
	}
}

// ActiveSessionsPageData is user/active_sessions.html's template context.
type ActiveSessionsPageData struct {
	web.LayoutData
	Sessions []ActiveSession
}

func renderActiveSessionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, sessionsList []ActiveSession) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Active Sessions")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ActiveSessionsPageData{LayoutData: layout, Sessions: sessionsList}
	if err := activeSessionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - active sessions template render error: %v", err)
	}
}

// ActivityPageData is user/activity.html's template context.
type ActivityPageData struct {
	web.LayoutData
	ActivityPageResult
}

// DateLabels are the translated preset names the date range picker needs in JS
func (d ActivityPageData) DateLabels() map[string]string {
	return map[string]string{
		"todayShort": d.T.Get("Today"),
		"today":      d.T.Get("Today"),
		"last7":      d.T.Get("Last 7 days"),
		"last30":     d.T.Get("Last 30 days"),
		"last3m":     d.T.Get("Last 3 months"),
		"last6m":     d.T.Get("Last 6 months"),
		"ytd":        d.T.Get("Year to date"),
	}
}

func renderActivityPage(a *appctx.App, w http.ResponseWriter, r *http.Request, result ActivityPageResult) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Activity Log")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ActivityPageData{LayoutData: layout, ActivityPageResult: result}
	if err := activityPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - activity template render error: %v", err)
	}
}

// LoginHistoryPageData is user/loginlog.html's template context.
type LoginHistoryPageData struct {
	web.LayoutData
	LastLoginData []appctx.LastLogin
}

func renderLoginHistoryPage(a *appctx.App, w http.ResponseWriter, r *http.Request, lastLoginData []appctx.LastLogin) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Login History")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := LoginHistoryPageData{LayoutData: layout, LastLoginData: lastLoginData}
	if err := loginHistoryPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - login history template render error: %v", err)
	}
}

// MCPPageData is user/mcp.html's template context.
type MCPPageData struct {
	web.LayoutData
	Tokens   []mcptokens.Token
	MCPURL   string
	NewToken string
}

func renderMCPPage(a *appctx.App, w http.ResponseWriter, r *http.Request, tokens []mcptokens.Token, mcpURL, newToken string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "AI Assistant (MCP)")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := MCPPageData{LayoutData: layout, Tokens: tokens, MCPURL: mcpURL, NewToken: newToken}
	if err := mcpPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - mcp template render error: %v", err)
	}
}

// APISwaggerPageData is user/api_swagger.html's template context.
type APISwaggerPageData struct {
	web.LayoutData
	Token string
}

func renderAPISwaggerPage(a *appctx.App, w http.ResponseWriter, r *http.Request, token string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "API")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := APISwaggerPageData{LayoutData: layout, Token: token}
	if err := apiSwaggerPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("ACCOUNT - api swagger template render error: %v", err)
	}
}
