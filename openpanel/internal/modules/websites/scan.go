package websites

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

// this file adds a universal "scan for existing installations" + "detach" feature covering every CMS type this panel can install, living in the websites package (which already owns /sites) rather than as per-CMS routes since scan/detach just need filesystem detection + a DB row
// mirrors wordpress/manage.go's handleScanWordPress/handleDetachWordPress pattern as a parallel, more general implementation (that file stays as-is for the WP-only /wordpress list page) - unlike WordPress's wp-cli-mediated version this also needs a DB-host repair+connectivity-check step, since a config file found on disk may have its DB host set to 'localhost'/'127.0.0.1' which won't resolve from the separate mysql/mariadb container, so each detector rewrites it to the real container hostname and verifies connectivity via information_schema before importing
// every extractX/getXVersion helper below duplicates the equivalent function elsewhere in websites.go (or is freshly written) rather than calling into the joomla/opencart/etc packages, since those packages don't import websites and vice versa, to avoid import cycles

var scanSkipDirs = map[string]bool{"node_modules": true, ".git": true, "backups": true, "cache": true, "tmp": true, "var": true, "storage": true, "data": true}

func scanGetDomainID(a *appctx.App, domainName string) (int, bool) {
	var domainID int
	row := a.DB.QueryRow("SELECT domain_id FROM domains WHERE domain_url = ?", domainName)
	if err := row.Scan(&domainID); err != nil {
		return 0, false
	}
	return domainID, true
}

func scanCheckSiteExists(a *appctx.App, siteName string) bool {
	var id int
	row := a.DB.QueryRow("SELECT id FROM sites WHERE site_name = ?", siteName)
	return row.Scan(&id) == nil
}

// scanVerifyDB confirms dbName is actually reachable via a root-level information_schema query against the mysql/mariadb container, rather than logging in as the site's own DB user - same check style every module's backups.go already uses for table listing.
func scanVerifyDB(ctx context.Context, userContext, dbName string) bool {
	rows, err := mysqlmanager.Exec(ctx, userContext, "SELECT schema_name FROM information_schema.schemata WHERE schema_name = '"+dbName+"'", "")
	return err == nil && len(rows) > 0
}

// ---------------------- SHARED RESULT TYPE ---------------------- //

type scanFound struct {
	cmsType, path, version string
}

type scanOutcome struct {
	found   []scanFound
	skipped []string
}

func (o *scanOutcome) writeSummary(w http.ResponseWriter) {
	var summary strings.Builder
	summary.WriteString("Scan completed. Found installations:\n\n")
	for _, f := range o.found {
		summary.WriteString(strings.ToUpper(f.cmsType[:1]) + f.cmsType[1:] + " installation: " + f.path + "\n")
		summary.WriteString(strings.ToUpper(f.cmsType[:1]) + f.cmsType[1:] + " Version: " + f.version + "\n\n")
	}
	for _, s := range o.skipped {
		summary.WriteString("Skipped: " + s + "\n")
	}
	_, _ = w.Write([]byte(summary.String()))
}

// ---------------------- DETACH ---------------------- //

// handleSitesDetach is the universal counterpart to every per-CMS handleRemoveX: removes only the sites row, no file/DB deletion, so a scan can later re-import the still-live installation - works identically regardless of the site's type.
func handleSitesDetach(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing site ID."})
		return
	}

	var siteName string
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ?
		LIMIT 1`, id)
	if scanErr := row.Scan(&siteName); scanErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No data found for the provided site ID"})
		return
	}

	domainName := strings.Split(siteName, "/")[0]
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "An error occurred during detachment. Please try again."})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "detached website "+siteName, reqip.ClientIP(r))
	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	_, _ = w.Write([]byte("Detached successfully!"))
}

// ---------------------- SCAN ---------------------- //

// handleSitesScan is the universal counterpart to every per-CMS handleScanX: walks the current user's html_data volume once per CMS type they're allowed to use, detecting/repairing/importing anything found and not already tracked.
func handleSitesScan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	injectedData, _ := a.InjectData(ctx, userID)
	allowedSlice, _ := injectedData["user_allowed"].([]string)
	allowed := make(map[string]bool, len(allowedSlice))
	for _, m := range allowedSlice {
		allowed[m] = true
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "initiated scan for existing installations", reqip.ClientIP(r))

	const wwwBaseDirectory = "/var/www/html/"
	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")

	outcome := &scanOutcome{}

	if allowed["wordpress"] {
		scanWordPress(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["joomla"] {
		scanJoomla(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["opencart"] {
		scanOpenCart(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["nextcloud"] {
		scanNextcloud(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["prestashop"] {
		scanPrestashop(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["drupal"] {
		scanDrupal(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["matomo"] {
		scanMatomo(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}
	if allowed["moodle"] {
		scanMoodle(ctx, a, userID, userContext, baseDirectory, mysqlVersion, outcome)
	}
	if allowed["mediawiki"] {
		scanMediaWiki(ctx, a, userID, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion, outcome)
	}

	if len(outcome.found) > 0 {
		_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	}
	outcome.writeSummary(w)
}

// insertScannedSite is the shared final step every scanX function ends with once a config file has passed detection+repair+connectivity-check.
func insertScannedSite(ctx context.Context, a *appctx.App, siteName, domainName, cmsType, version string) error {
	domainID, _ := scanGetDomainID(a, domainName)
	adminEmail := "admin@" + domainName
	// safety net for the sites table's version column - a per-CMS version regex producing something longer than expected (happened with Moodle's full $release string) should truncate rather than silently drop the whole import
	if len(version) > 32 {
		version = version[:32]
	}
	_, insertErr := a.DB.ExecContext(ctx, "INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		siteName, domainID, adminEmail, version, cmsType)
	return insertErr
}

// siteNameAndDomain derives (siteName, domainName) from a config file's container-visible path relative to /var/www/html/ - every domain's docroot is /var/www/html/<domain>[/<subfolder>...], so the first path segment is always the domain and everything after it is the subdirectory portion of the site name.
func siteNameAndDomain(containerRoot, wwwBaseDirectory string) (siteName, domainName string) {
	relPath := strings.TrimPrefix(containerRoot, wwwBaseDirectory)
	siteName = strings.TrimSuffix(relPath, "/")
	domainName = siteName
	if idx := strings.Index(siteName, "/"); idx != -1 {
		domainName = siteName[:idx]
	}
	return siteName, domainName
}
