package websites

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the websites module's routes onto mux: the /sites listing
// page and its side JSON endpoints. The /website CMS-type dispatcher is
// registered separately once its per-type templates are ready.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := auth.RequireLogin(a, "websites")

	mux.Handle("GET /sites", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleListSites(a, w, r)
	})))
	mux.Handle("GET /sites/scan", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleSitesScan(a, w, r)
	})))
	mux.Handle("POST /sites/detach", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleSitesDetach(a, w, r)
	})))
	mux.Handle("POST /sites/bulk", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleSitesBulk(a, mux, w, r)
	})))
	mux.Handle("GET /sites/updates", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleSitesUpdates(a, w, r)
	})))
	mux.Handle("/website", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWebsiteDispatch(a, w, r)
	})))

	mux.Handle("GET /json/favicon/{domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFavicon(a, w, r)
	})))
	mux.Handle("GET /json/database-size", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDatabaseSize(a, w, r)
	})))
	mux.Handle("/json/screenshot/{domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleScreenshot(a, w, r)
	})))
	mux.Handle("GET /domains/temporary-link", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleTemporaryLink(a, w, r)
	})))
	mux.Handle("GET /json/visitors/{domain}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleVisitors(a, w, r)
	})))

	mux.Handle("/json/safebrowsing/{domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleGoogleSafeBrowsing(a, w, r)
	})))
	mux.Handle("/json/page_speed/{domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlePageSpeed(a, w, r)
	})))
	mux.Handle("/json/wp_vulnerability/{domain...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWPVulnerability(a, w, r)
	})))
	mux.Handle("POST /pm2/install/{install_type}/{selected_domain}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInstallPackages(a, w, r)
	})))
	mux.Handle("GET /website/wp_info/{site_name...}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWebsiteWPInfo(a, w, r)
	})))
	mux.Handle("/wordpress/wp-cli/{action}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWordPressWPCLI(a, w, r)
	})))
}
