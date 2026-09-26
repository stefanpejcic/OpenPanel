package web

import (
	"fmt"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// RequestTranslator resolves the same locale BuildLayoutData does, for text built outside a page render like flash messages
func RequestTranslator(a *appctx.App, r *http.Request) i18n.Translator {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	sessionLocale, _ := sess.Values["locale"].(string)
	userLocale := ""
	if sessionLocale == "" {
		if userID, ok := auth.UserID(r); ok && userID != 0 {
			if details, err := a.GetUserDetailsWithPlan(r.Context(), userID); err == nil {
				userLocale = i18n.UserLocale(details.Context)
			}
		}
	}
	return a.I18n.Translator(a.I18n.ResolveLocale(r.Context(), sessionLocale, userLocale, r.Header.Get("Accept-Language")))
}

// Tr translates msg into the request's locale and fills its %(name)s placeholders from name/value pairs
func Tr(a *appctx.App, r *http.Request, msg string, kv ...any) string {
	if msg == "" {
		return "" // gettext hands back the catalog header for an empty msgid
	}
	pairs := make([]string, len(kv))
	for i, v := range kv {
		pairs[i] = fmt.Sprint(v)
	}
	return RequestTranslator(a, r).Get(msg, pairs...)
}
