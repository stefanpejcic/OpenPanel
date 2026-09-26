package crons

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the crons routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "crons")(h)
	}

	mux.Handle("GET /cronjobs/log", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsLog(a, w, r) }))
	mux.Handle("GET /cronjobs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobs(a, w, r) }))
	mux.Handle("GET /cronjobs/editor", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsView(a, w, r, "code") }))
	mux.Handle("GET /cronjobs/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsView(a, w, r, "logs") }))
	mux.Handle("GET /cronjobs/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsNew(a, w, r) }))
	mux.Handle("POST /cronjobs/save", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSaveCronjob(a, w, r) }))
	mux.Handle("POST /cronjobs/edit", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditCronjob(a, w, r) }))
	mux.Handle("POST /cronjobs/timezone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronTimeZone(a, w, r) }))
	mux.Handle("POST /cronjobs/toggle", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleToggleCronjob(a, w, r) }))
	mux.Handle("POST /cronjobs/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteCronjob(a, w, r) }))
	mux.Handle("POST /cronjobs/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsBulk(a, mux, w, r) }))
	mux.Handle("GET /ws/cronjobs/run", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCronjobsRunWS(a, w, r) }))
}
