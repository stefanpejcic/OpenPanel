// Package cmsclone holds the clone-workflow steps shared by every CMS's clone.go (validation, site-limit check, destination lookup, DB dump/create, chown, and the finalize tail).
// Each CMS package still owns its own file-copy and config-rewrite steps, since those genuinely differ per CMS.
package cmsclone

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

var (
	domainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	dbNameRE = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// ValidDomain/ValidDB/ValidDocroot replace the identical validation helpers every CMS clone.go used to define for itself.
// ValidDocroot accepts the "/var/www/html/..." absolute form clone handlers actually use, unlike WordPress's own stricter validateDocroot().
func ValidDomain(name string) bool { return name != "" && domainRE.MatchString(name) }
func ValidDB(name string) bool     { return name != "" && dbNameRE.MatchString(name) }
func ValidDocroot(path string) bool {
	return path != "" && !strings.Contains(path, "..") && strings.HasPrefix(path, "/var/www/html/")
}

// WithinSiteLimit reports whether the user can clone one more site under their plan's WebsitesLimit ("" or non-numeric means unlimited)
func WithinSiteLimit(ctx context.Context, a *appctx.App, userID, currentSiteCount int) bool {
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	limit := 0
	if v, err := strconv.Atoi(plan.WebsitesLimit); err == nil {
		limit = v
	}
	return limit == 0 || currentSiteCount < limit
}

// ResolveDestination looks up the destination domain's row and joins dstFolder onto its docroot for a subdirectory clone
func ResolveDestination(ctx context.Context, a *appctx.App, dstDomain, dstFolder string) (domainID int, docroot, phpVersion, dstDomainWithSubdir string, ok bool) {
	row := a.DB.QueryRowContext(ctx, "SELECT domain_id, docroot, php_version FROM domains WHERE domain_url = ?", dstDomain)
	if err := row.Scan(&domainID, &docroot, &phpVersion); err != nil {
		return 0, "", "", "", false
	}
	dstDomainWithSubdir = dstDomain
	if dstFolder != "" {
		docroot = filepath.Join(docroot, dstFolder)
		dstDomainWithSubdir = dstDomain + "/" + dstFolder
	}
	return domainID, docroot, phpVersion, dstDomainWithSubdir, true
}

// SelectDumpCommand picks the mysqldump/mariadb-dump invocation for the user's MYSQL_TYPE, or errors if it's neither
func SelectDumpCommand(userContext string) (dumpCmd, mysqlVersion string, err error) {
	mysqlVersion = webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	switch mysqlVersion {
	case "mysql":
		return "mysqldump --column-statistics=0 --set-gtid-purged=OFF", mysqlVersion, nil
	case "mariadb":
		return "mariadb-dump --gtid", mysqlVersion, nil
	default:
		return "", mysqlVersion, &unsupportedMySQLTypeError{mysqlVersion}
	}
}

// unsupportedMySQLTypeError carries the "Unsupported MYSQL_TYPE: ..." message text
type unsupportedMySQLTypeError struct{ mysqlVersion string }

func (e *unsupportedMySQLTypeError) Error() string {
	return "Unsupported MYSQL_TYPE: " + e.mysqlVersion
}

// ChownRecursive chown -R's each path to userContext's container UID, doing nothing on a lookup failure. Takes a variadic list since Moodle needs it for both approot and dataroot.
func ChownRecursive(ctx context.Context, userContext string, paths ...string) {
	uid, err := podmanmanager.GetUID(userContext)
	if err != nil {
		return
	}
	idStr := strconv.Itoa(uid)
	for _, p := range paths {
		_ = exec.CommandContext(ctx, "chown", "-R", idStr+":"+idStr, p).Run()
	}
}

// DumpStageFailed reports whether err came from the dump-pipe step (which callers report differently than a DB setup failure)
func DumpStageFailed(err error) bool {
	_, ok := err.(*dumpStageError)
	return ok
}

type dumpStageError struct{ err error }

func (e *dumpStageError) Error() string { return e.err.Error() }
func (e *dumpStageError) Unwrap() error { return e.err }

// CreateDatabaseAndDump creates dstDB/dstDBUser, grants privileges, and pipes a dump of srcDB into dstDB.
// escapedPassword is returned for callers that need it for a config-file rewrite step. Use DumpStageFailed(err) to tell which error response shape to write.
func CreateDatabaseAndDump(ctx context.Context, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword string) (escapedPassword string, err error) {
	escapedPassword = strings.ReplaceAll(dstDBUserPassword, `\`, `\\`)
	escapedPassword = strings.ReplaceAll(escapedPassword, `'`, `\'`)

	cloneQueries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dstDB + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dstDBUser + "'@'%' IDENTIFIED BY '" + escapedPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dstDB + "`.* TO '" + dstDBUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range cloneQueries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			return escapedPassword, execErr
		}
	}

	// srcDB/dstDB are already validated by ValidDB, so no quoting here - backticks would actually break this, since bash -c treats an unescaped backtick as command substitution, not SQL quoting
	dumpTablesCmd := dumpCmd + " --single-transaction --quick " + srcDB + " | " + mysqlVersion + " " + dstDB
	fullDBArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, "bash", "-c", dumpTablesCmd)
	if runErr := podmanmanager.Command(ctx, userContext, fullDBArgv).Run(); runErr != nil {
		return escapedPassword, &dumpStageError{runErr}
	}
	return escapedPassword, nil
}

// FinalizeParams bundles what FinalizeSite needs to invalidate the cache, upsert the sites row, log the activity, and write the success response
type FinalizeParams struct {
	App             *appctx.App
	WriteJSON       func(w http.ResponseWriter, status int, v any)
	UserID          int
	Username        string
	CMSDisplayName  string // e.g. "Joomla" - used in the activity-log message
	CMSType         string // e.g. "joomla" - the sites.type value
	ProvidedDomain  string
	DstDomainWithSubdir string
	DomainID        int
	AdminEmail      string
	Version         string
	SrcPath, DstPath, DstDB string
}

// FinalizeSite runs the cache-invalidate + sites-table upsert + activity-log + success-response tail every clone ends with, writing the error response itself and returning false on a DB failure
func FinalizeSite(ctx context.Context, w http.ResponseWriter, r *http.Request, p FinalizeParams) bool {
	_ = p.App.Cache.Delete(ctx, "get_user_websites:"+strconv.Itoa(p.UserID))

	if _, err := p.App.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		p.DstDomainWithSubdir, p.DomainID, p.AdminEmail, p.Version, p.CMSType); err != nil {
		p.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": err.Error()})
		return false
	}

	_ = logger.RecordUserAction(p.App.Config, p.Username, "cloned "+p.CMSDisplayName+" website from "+p.ProvidedDomain+" to "+p.DstDomainWithSubdir, reqip.ClientIP(r))

	p.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "success", "source": p.ProvidedDomain, "target": p.DstDomainWithSubdir,
		"source_path": p.SrcPath, "target_path": p.DstPath, "target_db": p.DstDB,
	})
	return true
}
