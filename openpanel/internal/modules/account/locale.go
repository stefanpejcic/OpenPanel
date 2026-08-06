package account

import (
	"net/http"
	"os"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// handleAccountLocale sets the user's preferred UI language.
func handleAccountLocale(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)

	locales := a.I18n.AvailableLocales(ctx)
	current := i18n.UserLocale(userContext)
	if current != "" && !contains(locales, current) {
		current = ""
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		locale := r.Form.Get("locale")
		if locale != "" {
			sess, _ := a.Sessions.Get(r, session.CookieName)

			var message string
			if oldLocale, ok := sess.Values["locale"].(string); ok && oldLocale != "" {
				message = "Language changed from " + oldLocale + " to " + locale
			} else {
				message = "Language changed to " + locale
			}

			sess.Values["locale"] = locale
			flash.Add(sess, "success", message)
			_ = a.Sessions.Save(r, w, sess)

			_ = logger.RecordUserAction(a.Config, username, "changed language to "+locale, reqip.ClientIP(r))

			userLocaleFile := filepath.Join("/home", userContext, "locale")
			if mkdirErr := os.MkdirAll(filepath.Dir(userLocaleFile), 0o755); mkdirErr == nil {
				_ = os.WriteFile(userLocaleFile, []byte(locale), 0o644)
			}

			redirectTo := r.Referer()
			if redirectTo == "" {
				redirectTo = "/account/language"
			}
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
	}

	renderLocalePage(a, w, r, locales, current)
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// RegisterLocale wires the locale-preference route onto mux, gated behind
// the "locale" feature flag.
func RegisterLocale(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "locale")(h)
	}
	mux.Handle("/account/language", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountLocale(a, w, r) }))
}
