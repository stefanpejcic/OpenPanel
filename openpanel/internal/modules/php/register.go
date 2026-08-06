package php

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires all PHP, PHP ini/options/extensions, and phpMyAdmin
// routes onto mux.
//
// Go's net/http.ServeMux only allows a wildcard to span an entire path
// segment (see its Patterns doc), so it can't match a version embedded
// inside a literal "php<version>" segment (e.g. /php/php8.2/options) -
// every version-scoped route here instead captures the whole segment as a
// wildcard and the handler unwraps it with phpVersionFromSegment()/
// phpVersionFromIniSegment().
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(feature string, h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, feature)(h)
	}

	mux.Handle("/php/default", requireLogin("php", func(w http.ResponseWriter, r *http.Request) { handleDefaultPHPVersion(a, w, r) }))
	mux.Handle("/php/domains", requireLogin("php", func(w http.ResponseWriter, r *http.Request) { handlePHPDomains(a, w, r) }))
	mux.Handle("GET /php/{phpversion}/info", requireLogin("php", func(w http.ResponseWriter, r *http.Request) { handlePHPInfo(a, w, r) }))

	mux.Handle("/php/php_ini_editor", requireLogin("php_ini", func(w http.ResponseWriter, r *http.Request) { handlePHPIniEditor(a, w, r, "") }))
	mux.Handle("/php/{phpiniversion}/editor", requireLogin("php_ini", func(w http.ResponseWriter, r *http.Request) {
		handlePHPIniEditor(a, w, r, r.PathValue("phpiniversion"))
	}))

	mux.Handle("/php/options", requireLogin("php_options", func(w http.ResponseWriter, r *http.Request) { handlePHPOptions(a, w, r, "") }))
	mux.Handle("/php/{phpversion}/options", requireLogin("php_options", func(w http.ResponseWriter, r *http.Request) {
		handlePHPOptions(a, w, r, r.PathValue("phpversion"))
	}))

	mux.Handle("GET /php/extensions", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) { handlePHPExtensionsSelect(a, w, r) }))
	mux.Handle("/php/{phpversion}/extensions", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) {
		handlePHPExtensions(a, w, r, r.PathValue("phpversion"))
	}))
	mux.Handle("GET /php/{phpversion}/available-extensions", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) {
		handlePHPAvailableExtensions(a, w, r, r.PathValue("phpversion"))
	}))
	mux.Handle("/php/{phpversion}/extensions-history", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) {
		handlePHPExtensionsHistory(a, w, r, r.PathValue("phpversion"))
	}))
	mux.Handle("POST /php/{phpversion}/install-extensions", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) {
		handlePHPInstallExtensions(a, w, r, r.PathValue("phpversion"))
	}))
	mux.Handle("GET /php/{phpversion}/install-extensions/status", requireLogin("php_extensions", func(w http.ResponseWriter, r *http.Request) {
		handlePHPInstallExtensionsStatus(a, w, r, r.PathValue("phpversion"))
	}))

	mux.Handle("GET /phpmyadmin/", requireLogin("phpmyadmin", func(w http.ResponseWriter, r *http.Request) { handlePHPMyAdminRedirect(a, w, r) }))
	mux.Handle("GET /mysql/phpmyadmin", requireLogin("phpmyadmin", func(w http.ResponseWriter, r *http.Request) { handlePHPMyAdminRedirect(a, w, r) }))
	mux.Handle("GET /phpmyadmin/link", requireLogin("phpmyadmin", func(w http.ResponseWriter, r *http.Request) { handlePHPMyAdminLoginLink(a, w, r) }))
}
