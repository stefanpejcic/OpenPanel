package wordpress

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the WordPress list + install routes onto mux.
// Backups, clone/remove/detach/reload/scan and wp-cli passthrough are
// added to this same Register separately.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := auth.RequireLogin(a, "wordpress")

	mux.Handle("GET /wordpress", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleListWordPress(a, w, r)
	})))
	mux.Handle("/wordpress/install", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInstallPage(a, w, r)
	})))

	mux.Handle("GET /wordpress/backup/get_dates/{selected_domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleGetBackupDates(a, w, r)
	})))
	mux.Handle("GET /wordpress/backup/restore/{selected_domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRestoreBackup(a, w, r)
	})))
	mux.Handle("GET /wordpress/backup/run/{selected_domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRunBackup(a, w, r)
	})))

	mux.Handle("POST /wordpress/clone", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleCloneWordPress(a, w, r)
	})))
	mux.Handle("POST /wordpress/remove", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRemoveWordPress(a, w, r)
	})))
	mux.Handle("POST /wordpress/detach", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDetachWordPress(a, w, r)
	})))
	mux.Handle("GET /wordpress/reload_data", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleReloadWordPressData(a, w, r)
	})))
	mux.Handle("GET /wordpress/scan", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleScanWordPress(a, w, r)
	})))
	mux.Handle("/wordpress/secure/{provided_domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWordPressSecure(a, w, r)
	})))
	mux.Handle("/wp-cli/{action}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWPCLI(a, w, r)
	})))
}
