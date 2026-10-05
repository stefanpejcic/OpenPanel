package emails

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterWebmailAPI registers the webmail API routes onto mux.
func RegisterWebmailAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "webmail", "GET /api/webmail", func(w http.ResponseWriter, r *http.Request) { apiWebmailInfo(a, w, r) })
	apiregistry.Handle(mux, a, "webmail", "POST /api/webmail/{email}", func(w http.ResponseWriter, r *http.Request) { apiWebmailToken(a, w, r) })
}

// apiWebmailInfo reports whether webmail is running and its URL, if so.
func apiWebmailInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	isRunning := isWebmailRunning(ctx)
	var url any
	if isRunning {
		url = GetWebmailDomain(ctx, a, currentUsername)
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"is_running": isRunning, "webmail_url": url})
}

// apiWebmailToken issues a one-time autologin token for a webmail account owned by the current user.
func apiWebmailToken(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	email := r.PathValue("email")

	if !isValidEmail(email) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}

	_, domain, found := strings.Cut(email, "@")
	if !found {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	owned := false
	for _, d := range domains {
		if d.DomainURL == domain {
			owned = true
			break
		}
	}
	if !owned {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	if !isWebmailRunning(ctx) {
		web.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Webmail service is not running"})
		return
	}

	webmailURL := GetWebmailDomain(ctx, a, currentUsername)
	token, tokenErr := createWebmailToken(ctx, email)
	if tokenErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": tokenErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "generated webmail token for "+email, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"email": email, "token": token, "autologin_url": webmailURL + "/autologin.php?token=" + token, "webmail_url": webmailURL,
	})
}
