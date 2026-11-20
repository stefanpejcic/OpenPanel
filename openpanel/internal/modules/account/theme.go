package account

import (
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleAccountTheme saves the theme the user picked, only when allow_user_theme=yes
func handleAccountTheme(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !web.AllowUserTheme {
		http.Error(w, "Switching themes is turned off on this server.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	folder := strings.TrimSpace(r.Form.Get("theme"))
	userContext := auth.UserContext(r)
	if !web.SaveUserTheme(userContext, folder) {
		http.Error(w, "unknown theme", http.StatusBadRequest)
		return
	}
	userID, _ := auth.UserID(r)
	data, _ := a.InjectData(r.Context(), userID)
	username, _ := data["current_username"].(string)
	_ = logger.RecordUserAction(a.Config, username, "changed theme to "+folder, reqip.ClientIP(r))

	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, "success", web.Tr(a, r, "Theme changed."))
	_ = a.Sessions.Save(r, w, sess)

	// back to the page they were on, never to another site
	redirectTo := "/dashboard"
	if ref, err := url.Parse(r.Referer()); err == nil && ref.Path != "" && (ref.Host == "" || ref.Host == r.Host) {
		redirectTo = ref.RequestURI()
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// RegisterTheme wires the theme picker's POST route, every user has the dashboard feature so the picker works for all of them
func RegisterTheme(mux *http.ServeMux, a *appctx.App) {
	mux.Handle("POST /account/theme", auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleAccountTheme(a, w, r)
	})))
}
