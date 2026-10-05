// Package phpapp installs a Composer-based PHP project into an existing domain's docroot, run inside whichever shared php-fpm-<version> container that domain's PHP version already points to - unlike appinstall's NodeJS/Python installer, this never creates a dedicated container or touches docker-compose.yml/reverse-proxy config, since the domain's existing vhost already routes there
// settings are still persisted the same way appinstall persists CPU/RAM - as .env keys under a synthetic prefix derived from the site name, since there's no per-app container to key them off of
package phpapp

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the PHP app install/manage routes onto mux, gated behind the already-enabled "php" feature flag
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

// RegisterAPI wires POST /api/php/install onto mux - same pattern as nodejs.RegisterAPI/python.RegisterAPI, reusing HandleInstall as-is since it's already pure form-value-in/NDJSON-out with no session/flash usage
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php", "POST /api/php/install", func(w http.ResponseWriter, r *http.Request) {
		HandleInstall(a, w, r)
	})

	// manage-surface for already-installed Composer apps, reusing the same handlers as Register's UI routes since they already write JSON/text directly with no session/flash dependency - "{site_name...}" must be the last path segment, so the action stays a literal prefix
	apiregistry.Handle(mux, a, "php", "POST /api/php/apps/composer-install/{site_name...}", func(w http.ResponseWriter, r *http.Request) {
		handleComposerAction(a, w, r, "install")
	})
	apiregistry.Handle(mux, a, "php", "POST /api/php/apps/composer-update/{site_name...}", func(w http.ResponseWriter, r *http.Request) {
		handleComposerAction(a, w, r, "update")
	})
	apiregistry.Handle(mux, a, "php", "GET /api/php/apps/logs/{site_name...}", func(w http.ResponseWriter, r *http.Request) { handleComposerLogs(a, w, r) })
	apiregistry.Handle(mux, a, "php", "DELETE /api/php/apps/{site_name...}", func(w http.ResponseWriter, r *http.Request) { handleDelete(a, w, r) })
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
