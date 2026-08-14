package drupal

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

var (
	removeDBNameRE = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`'username'\s*=>\s*'([^']*)'`)
)

// stripPHPCommentLines drops every line whose trimmed content starts with
// "*" - settings.php ships a large /** ... */ documentation block with
// placeholder 'database' => 'database_name' style example lines that would
// otherwise match removeDBNameRE/removeDBUserRE before the real
// $databases['default'] array drush appends near the end of the file.
func stripPHPCommentLines(content string) string {
	var codeLines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	return strings.Join(codeLines, "\n")
}

// handleRemoveDrupal fully uninstalls a Drupal site: drops the database
// and user (parsed out of settings.php, same regex-scrape approach
// wordpress/manage.go uses for wp-config.php), deletes the whole install
// directory, and removes the sites row. Unlike phpapp's no-op delete, this
// matches WordPress's full-cleanup uninstall behavior per this feature's
// scope.
func handleRemoveDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		WHERE sites.id = ? AND sites.type = 'drupal'`, id)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]
	subdirectory := parts[1:]

	if docroot == "" {
		flashAndRedirect(a, w, r, "error", "Drupal installation not found in the database", "/sites")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	realInstallPath := strings.TrimPrefix(docroot, "/var/www/html/")
	if len(subdirectory) > 0 {
		realInstallPath = filepath.Join(append([]string{realInstallPath}, subdirectory...)...)
	}
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + realInstallPath
	settingsFile := filepath.Join(volume, "web", "sites", "default", "settings.php")
	content, _ := os.ReadFile(settingsFile)
	codeOnly := stripPHPCommentLines(string(content))

	dbNameMatch := removeDBNameRE.FindStringSubmatch(codeOnly)
	dbUserMatch := removeDBUserRE.FindStringSubmatch(codeOnly)
	if dbNameMatch != nil && dbUserMatch != nil {
		dbName, dbUser := dbNameMatch[1], dbUserMatch[1]
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
		invalidateMySQLCaches(ctx, a, userContext, currentUsername)
	} else {
		flashSess(a, w, r, "warning", "Database name or user not found in settings.php")
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
		flashSess(a, w, r, "error", "An error occurred during Drupal uninstall.")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled Drupal website "+siteName, reqip.ClientIP(r))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Drupal uninstalled successfully"})
		return
	}
	flashAndRedirect(a, w, r, "success", "Drupal uninstalled successfully", "/sites")
}
