package sofawiki

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// handleRemoveSofawiki uninstalls a SofaWiki site: deletes the whole
// install directory and removes the sites row. No database to drop.
func handleRemoveSofawiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'sofawiki'`, id)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]
	subdirectory := parts[1:]

	if docroot == "" {
		flashAndRedirect(a, w, r, "error", "SofaWiki installation not found in the database", "/sites")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, selectedDomain)
	phpContainer := webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + phpVersion
	}
	installPath := docroot
	if len(subdirectory) > 0 {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + strings.Join(subdirectory, "/")
	}
	_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath)).Run()

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		flashSess(a, w, r, "error", "An error occurred during SofaWiki uninstall.")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled SofaWiki website "+siteName, reqip.ClientIP(r))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "SofaWiki uninstalled successfully"})
		return
	}
	flashAndRedirect(a, w, r, "success", "SofaWiki uninstalled successfully", "/sites")
}
