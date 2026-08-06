// Package goaccess serves the pre-rendered GoAccess HTML report per domain,
// generated externally (opencli/cron, once every 24h per the UI copy) and
// simply read from disk here.
package goaccess

import (
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func injected(a *appctx.App, r *http.Request) (username string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", err
	}
	username, _ = data["current_username"].(string)
	return username, nil
}

// handleDomainStats shows the domain picker when no domain_name is given,
// or reads and serves that domain's pre-rendered GoAccess HTML report file
// as-is.
func handleDomainStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainName := r.PathValue("domain_name")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderGoaccessSelectPage(a, w, r, domainsList)
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	currentUsername, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	logFilePath := "/var/log/caddy/stats/" + currentUsername + "/" + domainName + ".html"
	content, readErr := os.ReadFile(logFilePath)
	if readErr != nil {
		flashAndRedirect(a, w, r, "error", "Stats file for domain "+domainName+" not found. Data is generated every 24h.", "/domains/log")
		return
	}

	// goaccess_single.html has no {% extends %} - the report is a
	// complete standalone HTML document, not wrapped in the panel's own
	// layout, so this writes the file's bytes directly rather than going
	// through a web.Page.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(content)
}
