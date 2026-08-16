package moodle

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

var (
	removeDBNameRE = regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`CFG->dbuser\s*=\s*'([^']*)'`)
)

// handleRemoveMoodle fully uninstalls a Moodle site: removes the per-minute
// admin/cli/cron.php job registered at install time, drops the database and
// user (parsed out of the approot's config.php), deletes the docroot
// symlink plus its backing approot/dataroot directories, and removes the
// sites row. Mirrors prestashop/manage.go's handleRemovePrestashop, with
// the config.php/approot lookup adjusted for Moodle's public/-split layout
// (see moodle.go's package doc comment).
func handleRemoveMoodle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		WHERE sites.id = ? AND sites.type = 'moodle'`, id)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]

	if docroot == "" {
		flashAndRedirect(a, w, r, "error", "Moodle installation not found in the database", "/sites")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	slug := siteSlug(siteName)
	approotHostPath := filepath.Join(htmlVolume, slug+"_moodleapp")
	datarootHostPath := filepath.Join(htmlVolume, slug+"_moodledata")

	realInstallPath := strings.TrimPrefix(docroot, "/var/www/html/")
	if len(parts) > 1 {
		realInstallPath = filepath.Join(append([]string{realInstallPath}, parts[1:]...)...)
	}
	hostOSPath := filepath.Join(htmlVolume, realInstallPath)

	configFile := filepath.Join(approotHostPath, "config.php")
	content, _ := os.ReadFile(configFile)

	dbNameMatch := removeDBNameRE.FindStringSubmatch(string(content))
	dbUserMatch := removeDBUserRE.FindStringSubmatch(string(content))
	if dbNameMatch != nil && dbUserMatch != nil {
		dbName, dbUser := dbNameMatch[1], dbUserMatch[1]
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
		invalidateMySQLCaches(ctx, a, userContext, currentUsername)
	} else {
		flashSess(a, w, r, "warning", "Database name or user not found in config.php")
	}

	_ = crons.RemoveJobByComment(ctx, userContext, moodleCronComment(siteName))

	_ = os.Remove(hostOSPath)
	_ = os.RemoveAll(approotHostPath)
	_ = os.RemoveAll(datarootHostPath)

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		flashSess(a, w, r, "error", "An error occurred during Moodle uninstall.")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled Moodle website "+siteName, reqip.ClientIP(r))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Moodle uninstalled successfully"})
		return
	}
	flashAndRedirect(a, w, r, "success", "Moodle uninstalled successfully", "/sites")
}
