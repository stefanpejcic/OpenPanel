package wordpress

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func wordpressBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "update", Label: t.Get("Update"), Confirm: t.Get("Update WordPress core on the selected sites?")},
		{Key: "backup", Label: t.Get("Backup"), Confirm: t.Get("Back up the files and database of the selected sites?")},
		{Key: "detach", Label: t.Get("Detach"), Confirm: t.Get("Detach the selected sites from WordPress Manager? Files and databases are kept.")},
		{Key: "delete", Label: t.Get("Uninstall"), Confirm: t.Get("Uninstall the selected sites, including their files and databases? This cannot be undone."), Danger: true},
	}
}

type wpSiteRef struct {
	id      int
	docroot string
}

// wordpressSiteRefs maps each of the user's site names to its ID and full docroot (domain docroot plus the sub-folder, like /sites does)
func wordpressSiteRefs(a *appctx.App, r *http.Request, userID int) map[string]wpSiteRef {
	refs := map[string]wpSiteRef{}
	rows, err := a.DB.QueryContext(r.Context(), `
		SELECT sites.id, sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE domains.user_id = ?`, userID)
	if err != nil {
		return refs
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, docroot string
		if rows.Scan(&id, &name, &docroot) != nil {
			continue
		}
		if _, folder, ok := strings.Cut(name, "/"); ok && folder != "" {
			docroot = strings.TrimRight(docroot, "/") + "/" + folder
		}
		refs[name] = wpSiteRef{id: id, docroot: docroot}
	}
	return refs
}

// handleWordPressBulk replays each site through the same routes the per-site buttons and /sites bulk use
func handleWordPressBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	refs := wordpressSiteRefs(a, r, userID)
	web.ServeBulkDispatch(a, mux, w, r, wordpressBulkActions(web.RequestTranslator(a, r)), func(action, _, site string) (*web.BulkCall, web.BulkResult) {
		ref, ok := refs[site]
		if !ok {
			return web.Skip("Site not found.")
		}
		id := url.Values{"id": {strconv.Itoa(ref.id)}}
		switch action {
		case "update":
			return web.Call(web.BulkCall{Method: http.MethodGet, Path: "/wordpress/wp-cli/update_now", Query: url.Values{"website": {site}, "docroot": {ref.docroot}}})
		case "backup":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/wordpress/backup/run/" + site, Query: url.Values{"docroot": {ref.docroot}, "backup_database": {"true"}, "backup_files": {"true"}}})
		case "detach":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/wordpress/detach", Form: id})
		case "delete":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/wordpress/remove", Form: id})
		}
		return web.Skip("Unknown bulk action.")
	})
}
