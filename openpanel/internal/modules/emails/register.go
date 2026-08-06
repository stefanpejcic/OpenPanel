package emails

import (
	"context"
	"net/http"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// RegisterAccounts wires the email account routes onto mux.
func RegisterAccounts(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "emails")(h)
	}

	mux.Handle("GET /emails/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsNew(a, w, r) }))
	mux.Handle("POST /emails/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsNew(a, w, r) }))

	mux.Handle("GET /emails", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))
	mux.Handle("POST /emails", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))
	mux.Handle("DELETE /emails", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))
	mux.Handle("GET /emails/edit/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))
	mux.Handle("POST /emails/edit/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))
	mux.Handle("DELETE /emails/edit/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmails(a, w, r) }))

	mux.Handle("GET /emails/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsDelete(a, w, r) }))
	mux.Handle("GET /emails/delete/{address}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsDelete(a, w, r) }))

	mux.Handle("GET /emails/info/{address}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsServerInfo(a, w, r) }))
	mux.Handle("GET /emails/configuration/{type}/{account}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailConfiguration(a, w, r) }))
}

// RegisterAliases wires the email alias routes onto mux.
func RegisterAliases(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_aliases")(h)
	}

	mux.Handle("GET /emails/aliases", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliases(a, w, r) }))
	mux.Handle("GET /emails/aliases/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasNew(a, w, r) }))
	mux.Handle("POST /emails/aliases/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasNew(a, w, r) }))
	mux.Handle("GET /emails/aliases/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasDeletePage(a, w, r) }))
	mux.Handle("GET /emails/aliases/delete/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasDeletePage(a, w, r) }))
	mux.Handle("GET /emails/aliases/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasDetail(a, w, r) }))
	mux.Handle("POST /emails/aliases/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasDetail(a, w, r) }))
	mux.Handle("DELETE /emails/aliases/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAliasDetail(a, w, r) }))
}

// RegisterDefault wires the default (catch-all) alias routes onto mux.
func RegisterDefault(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_default")(h)
	}

	mux.Handle("GET /emails/default/", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDefaultAlias(a, w, r) }))
	mux.Handle("POST /emails/default/", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDefaultAlias(a, w, r) }))
	mux.Handle("GET /emails/default/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDefaultAlias(a, w, r) }))
	mux.Handle("POST /emails/default/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDefaultAlias(a, w, r) }))
}

// RegisterDeliverability wires the email deliverability routes onto mux.
func RegisterDeliverability(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_deliverability")(h)
	}

	mux.Handle("GET /emails/deliverability", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailsDeliverability(a, w, r) }))
	mux.Handle("GET /emails/deliverability/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailDeliverabilityDomain(a, w, r) }))
}

// RegisterExport wires the email export route onto mux.
func RegisterExport(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_export")(h)
	}

	mux.Handle("GET /emails/export", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEmailExport(a, w, r) }))
}

// RegisterFilters wires the mail filter (Sieve) routes onto mux.
func RegisterFilters(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_filters")(h)
	}

	mux.Handle("GET /emails/filter", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFiltersForUser(a, w, r) }))
	mux.Handle("GET /emails/filter/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilterForEmail(a, w, r) }))
	mux.Handle("GET /emails/filter/{email}/gui", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilterGUI(a, w, r) }))
	mux.Handle("POST /emails/filter/{email}/gui", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilterGUI(a, w, r) }))
	mux.Handle("GET /emails/filter/{email}/raw", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilterRaw(a, w, r) }))
	mux.Handle("POST /emails/filter/{email}/raw", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilterRaw(a, w, r) }))
}

// RegisterImport wires the email import routes onto mux.
func RegisterImport(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "email_import")(h)
	}

	mux.Handle("GET /emails/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleImportEmails(a, w, r) }))
	mux.Handle("POST /emails/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleImportEmails(a, w, r) }))
	mux.Handle("POST /emails/import/confirm", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleConfirmEmailImport(a, w, r) }))
}

var webmailInitOnce sync.Once

// RegisterWebmail wires the webmail routes onto mux, and ensures the
// webmail master user exists once - only when "webmail" is an enabled
// module, since that's the only time this function is called at all.
func RegisterWebmail(mux *http.ServeMux, a *appctx.App) {
	webmailInitOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				// Never let a webmail setup failure abort the rest of
				// RegisterAll.
			}
		}()
		ensureMasterUser(context.Background())
	})

	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "webmail")(h)
	}

	mux.Handle("GET /webmail/{email}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebmailLogin(a, w, r) }))
	mux.Handle("GET /webmail/", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebmailLogin(a, w, r) }))
}
