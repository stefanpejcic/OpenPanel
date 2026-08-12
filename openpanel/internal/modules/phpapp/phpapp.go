// Package phpapp installs a Composer-based PHP project into an existing
// domain's docroot, run inside whichever shared php-fpm-<version> container
// that domain's PHP version already points to. Unlike the NodeJS/Python
// app installer (internal/modules/appinstall), this never creates a
// dedicated container, edits docker-compose.yml, or touches a webserver
// reverse-proxy config - the domain's existing vhost already routes to the
// right php-fpm container. Settings are still persisted the same way
// appinstall persists CPU/RAM/etc - as .env keys under a synthetic prefix
// derived from the site name, since there's no per-app container to key
// them off of.
package phpapp

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires the PHP app install/manage routes onto mux, gated behind
// the already-enabled "php" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "php")(h)
	}
	mux.Handle("GET /php/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { HandleInstallPage(a, w, r) }))
	mux.Handle("POST /php/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { HandleInstallPage(a, w, r) }))

	mux.Handle("POST /php/manage/composer-install/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleComposerAction(a, w, r, "install")
	}))
	mux.Handle("POST /php/manage/composer-update/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleComposerAction(a, w, r, "update")
	}))
	mux.Handle("GET /php/manage/logs/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleComposerLogs(a, w, r) }))
	mux.Handle("POST /php/manage/delete/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDelete(a, w, r) }))
}

// RegisterAPI wires POST /api/php/install onto mux - same pattern as
// nodejs.RegisterAPI/python.RegisterAPI, reusing HandleInstall as-is since
// it's already pure form-value-in/NDJSON-out with no session/flash usage.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php", "POST /api/php/install", func(w http.ResponseWriter, r *http.Request) {
		HandleInstall(a, w, r)
	})
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	flashSess(a, w, r, category, message)
	http.Redirect(w, r, path, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

func injectedContext(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}
