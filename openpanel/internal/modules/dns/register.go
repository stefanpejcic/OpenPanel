package dns

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the DNS zone editor routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "dns")(h)
	}

	mux.Handle("GET /domains/edit-dns-zone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditDNSZone(a, w, r) }))
	mux.Handle("GET /domains/edit-dns-zone/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditDNSZone(a, w, r) }))

	mux.Handle("POST /domains/dns/update-record/{row_id}/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleUpdateDNSRecord(a, w, r) }))
	mux.Handle("POST /domains/dns/delete-record/{rowId}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDNSRecord(a, w, r) }))
	mux.Handle("POST /domains/dns/add-record/", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAddDNSRecord(a, w, r) }))

	mux.Handle("POST /domains/save-dns-zone/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSaveDNSZone(a, w, r) }))
	mux.Handle("POST /domains/restart-dns-zone/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRestartDNSZone(a, w, r) }))
	mux.Handle("GET /domains/export-dns-zone/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleExportDNSZone(a, w, r) }))
}
