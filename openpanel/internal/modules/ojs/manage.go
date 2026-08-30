package ojs

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// handleRemoveOJS fully uninstalls an OJS site: removes the per-minute
// scheduler cron job registered at install time, drops the database and
// user (parsed out of the approot's config.inc.php), deletes the docroot
// symlink plus its backing approot/files directories, and removes the sites
// row. Mirrors moodle/manage.go's handleRemoveMoodle, adjusted for
// config.inc.php's INI syntax (see config.go) and OJS having no per-site
// table prefix.
func handleRemoveOJS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		WHERE sites.id = ? AND sites.type = 'ojs'`, id)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]

	if docroot == "" {
		flashAndRedirect(a, w, r, "error", "OJS installation not found in the database", "/sites")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	slug := siteSlug(siteName)
	approotHostPath := filepath.Join(htmlVolume, slug+"_ojsapp")
	filesHostPath := filepath.Join(htmlVolume, slug+"_ojsfiles")

	realInstallPath := strings.TrimPrefix(docroot, "/var/www/html/")
	if len(parts) > 1 {
		realInstallPath = filepath.Join(append([]string{realInstallPath}, parts[1:]...)...)
	}
	hostOSPath := filepath.Join(htmlVolume, realInstallPath)

	dbInfo := extractOJSDatabaseInfoForLogin(userContext, siteName)
	if dbInfo["error"] == "" && dbInfo["database_name"] != "" {
		dbName := dbInfo["database_name"]
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
		if dbUser := dbInfo["database_username"]; dbUser != "" {
			_, _ = mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
		}
		invalidateMySQLCaches(ctx, a, userContext, currentUsername)
	} else {
		flashSess(a, w, r, "warning", "Database name or user not found in config.inc.php")
	}

	_ = crons.RemoveJobByComment(ctx, userContext, ojsCronComment(siteName))

	_ = os.Remove(hostOSPath)
	_ = os.RemoveAll(approotHostPath)
	_ = os.RemoveAll(filesHostPath)

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		flashSess(a, w, r, "error", "An error occurred during OJS uninstall.")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled OJS website "+siteName, reqip.ClientIP(r))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "OJS uninstalled successfully"})
		return
	}
	flashAndRedirect(a, w, r, "success", "OJS uninstalled successfully", "/sites")
}
