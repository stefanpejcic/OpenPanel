// Package domains implements the domains list/add/delete page and its
// satellite management pages (suspend, docroot, access logs, capitalize
// display, virtual host editor, redirects, SSL).
package domains

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires all domain management routes onto mux. "domains" gates
// the base list/add/delete flow, always available once the domains module
// itself is enabled; the satellite pages are individually gated to match
// the sidebar's own feature-conditional links.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(feature string, h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, feature)(h)
	}

	mux.Handle("GET /domains", requireLogin("domains", func(w http.ResponseWriter, r *http.Request) { handleDomainsPage(a, w, r) }))
	mux.Handle("/domains/new", requireLogin("domains", func(w http.ResponseWriter, r *http.Request) { handleDomainsNew(a, w, r) }))
	mux.Handle("/domains/delete", requireLogin("domains", func(w http.ResponseWriter, r *http.Request) { handleDeleteDomain(a, w, r) }))

	mux.Handle("/domains/suspend", requireLogin("domain_suspend", func(w http.ResponseWriter, r *http.Request) { handleSuspendDomain(a, w, r) }))
	mux.Handle("/domains/unsuspend", requireLogin("domain_suspend", func(w http.ResponseWriter, r *http.Request) { handleUnsuspendDomain(a, w, r) }))

	mux.Handle("/domains/docroot", requireLogin("docroot", func(w http.ResponseWriter, r *http.Request) { handleDomainDocroot(a, w, r) }))

	mux.Handle("GET /domains/log", requireLogin("domain_logs", func(w http.ResponseWriter, r *http.Request) { handleViewDomainAccessLog(a, w, r) }))
	// "GET /domains/log/{domain_name...}" already covers the bare
	// "/domains/log/" case (domain_name resolves to ""), matching the
	// established wildcard-route pattern used throughout this port.
	mux.Handle("GET /domains/log/{domain_name...}", requireLogin("domain_logs", func(w http.ResponseWriter, r *http.Request) { handleViewDomainAccessLog(a, w, r) }))

	mux.Handle("GET /domains/capitalize", requireLogin("capitalize_domains", func(w http.ResponseWriter, r *http.Request) {
		handleDisplayCapitalizedDomains(a, w, r)
	}))
	mux.Handle("/domains/capitalize/{domain}", requireLogin("capitalize_domains", func(w http.ResponseWriter, r *http.Request) {
		handleCapitalizeDomains(a, w, r)
	}))

	mux.Handle("/domains/vhosts", requireLogin("edit_vhost", func(w http.ResponseWriter, r *http.Request) { handleEditVhosts(a, w, r) }))

	mux.Handle("POST /domains/redirect/delete", requireLogin("redirects", func(w http.ResponseWriter, r *http.Request) { handleDeleteRedirect(a, w, r) }))
	mux.Handle("/domains/redirect", requireLogin("redirects", func(w http.ResponseWriter, r *http.Request) { handleSetRedirect(a, w, r) }))

	mux.Handle("/domains/ssl", requireLogin("ssl", func(w http.ResponseWriter, r *http.Request) { handleDomainCustomSSL(a, w, r) }))
	mux.Handle("GET /domains/dns/tlsa-hash/{domain}", requireLogin("ssl", func(w http.ResponseWriter, r *http.Request) { handleGetTLSAHash(a, w, r) }))
}
