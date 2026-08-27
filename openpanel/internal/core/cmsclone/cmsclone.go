// Package cmsclone holds the clone-workflow steps that are identical
// across every CMS's clone.go (joomla, drupal, moodle, prestashop,
// opencart, nextcloud, mediawiki, matomo): input validation, the
// site-limit check, destination-domain lookup, DB dump-command selection,
// DB create+dump-pipe, ownership chown, and the final
// cache-invalidate/sites-table-insert/activity-log/success-response tail.
//
// What stays local to each CMS package: the docroot/file-copy step (shapes
// differ - Moodle's approot+dataroot+symlink split is nothing like a plain
// "cp -a docroot" the other seven do) and the config-file rewrite step
// (regex/file-format differs per CMS), since those are the genuinely
// CMS-specific parts of a clone.
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

// ValidDomain/ValidDB/ValidDocroot replace the identical
// cloneValidateDomain/cloneValidateDB/cloneValidateDocroot helpers every
// CMS clone.go used to define for itself. ValidDocroot intentionally
// accepts the "/var/www/html/..." absolute-path form every clone handler's
// .Docroot/source_folder actually uses - WordPress's own separate
// validateDocroot() rejects a leading "/", which would reject the real
// values these forms submit; not replicated here.
func ValidDomain(name string) bool { return name != "" && domainRE.MatchString(name) }
func ValidDB(name string) bool     { return name != "" && dbNameRE.MatchString(name) }
func ValidDocroot(path string) bool {
	return path != "" && !strings.Contains(path, "..") && strings.HasPrefix(path, "/var/www/html/")
}

// WithinSiteLimit reports whether a user with the given current site count
// can clone one more, per their hosting plan's WebsitesLimit ("" or
// non-numeric means unlimited, matching every pre-refactor caller's
// atoiDefault(plan.WebsitesLimit, 0) fallback).
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

// ResolveDestination looks up the destination domain's row and joins
// dstFolder onto its docroot when a subdirectory clone was requested - the
// "SELECT domain_id, docroot, php_version ... dstDomainWithSubdir" block
// every clone.go repeated verbatim.
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

// SelectDumpCommand picks the mysqldump/mariadb-dump invocation for the
// user's configured MYSQL_TYPE, or a non-nil err (with the exact "Unsupported
// MYSQL_TYPE: ..." message every clone.go wrote) when it's neither.
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

// unsupportedMySQLTypeError carries the exact message text every clone.go
// wrote into its "Unsupported MYSQL_TYPE: ..." response.
type unsupportedMySQLTypeError struct{ mysqlVersion string }

func (e *unsupportedMySQLTypeError) Error() string {
	return "Unsupported MYSQL_TYPE: " + e.mysqlVersion
}

// ChownRecursive chown -R's each path to userContext's container UID,
// silently doing nothing on lookup failure - the
// "if uid, uidErr := podmanmanager.GetUID(...); uidErr == nil { chown }"
// block every clone.go repeated (Moodle repeated it twice, once per
// approot/dataroot path, which is why this takes a variadic list).
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

// DumpStageFailed distinguishes CreateDatabaseAndDump's two error shapes:
// callers used map[string]any{"status":"error","details":err} for a DB
// setup failure, but map[string]any{"status":"error","step":"command_failed"}
// (no error text) for a dump-pipe failure - this reports which one err
// came from.
func DumpStageFailed(err error) bool {
	_, ok := err.(*dumpStageError)
	return ok
}

type dumpStageError struct{ err error }

func (e *dumpStageError) Error() string { return e.err.Error() }
func (e *dumpStageError) Unwrap() error { return e.err }

// CreateDatabaseAndDump creates dstDB/dstDBUser, grants privileges, and
// pipes a dump of srcDB into dstDB - the CREATE DATABASE/CREATE
// USER/GRANT/FLUSH queries plus the dump-pipe podman exec every clone.go
// repeated identically. escapedPassword is returned for callers that also
// need it for a config-file rewrite step (every CMS but Joomla/Nextcloud
// does). Check DumpStageFailed(err) to tell which of the two original
// error response shapes to write.
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

	// srcDB/dstDB are validated against ^[a-zA-Z0-9_]+$ by ValidDB before
	// this is ever called, so no identifier quoting is needed here -
	// backticks would be actively wrong: bash -c interprets an unescaped
	// backtick as command substitution, not passed-through SQL quoting,
	// which silently truncated this exact command when it was first copied
	// from WordPress's clone (verified live: bash treated `srcDB` as "run
	// srcDB as a command").
	dumpTablesCmd := dumpCmd + " --single-transaction --quick " + srcDB + " | " + mysqlVersion + " " + dstDB
	fullDBArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, "bash", "-c", dumpTablesCmd)
	if runErr := podmanmanager.Command(ctx, userContext, fullDBArgv).Run(); runErr != nil {
		return escapedPassword, &dumpStageError{runErr}
	}
	return escapedPassword, nil
}

// FinalizeParams bundles what FinalizeSite needs to invalidate the sites
// cache, upsert the sites-table row, record the activity-log entry, and
// write the final success response.
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

// FinalizeSite runs the cache-invalidate + sites-table upsert +
// activity-log + success-response tail every clone.go ended with. On a DB
// error it writes the error response itself and returns false; on success
// it writes the {"status":"success",...} response and returns true.
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
