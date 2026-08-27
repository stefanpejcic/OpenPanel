// Package mysql implements database/user CRUD, the creation wizard,
// privilege management, export/import, table optimize/repair, per-service
// configuration, remote access, and the process list - all built on
// internal/core/mysqlmanager's per-user connection pool.
package mysql

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// flashSess adds a flash message without redirecting - several handlers
// here fall through to the same GET rendering logic below on error rather
// than redirecting.
func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// stripQuotes strips a single layer of matching quotes, used here on raw
// config values before splitting them into restricted-name lists.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
			(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// restrictedNames holds the admin-editable, space-separated lists of
// usernames/database names that are off-limits to account-level MySQL
// operations, parsed once from config, not per-request.
type restrictedNames struct {
	users     []string
	usersSQL  string
	databases []string
	dbsSQL    string
}

var restricted restrictedNames

// loadRestrictedNames parses the mysql_restricted_usernames/
// mysql_restricted_databases config values, called once from Register().
func loadRestrictedNames(a *appctx.App) {
	usersRaw := stripQuotes(a.Config.Get("mysql_restricted_usernames",
		"mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck"))
	restricted.users = splitTrimQuotes(usersRaw)
	restricted.usersSQL = sqlQuotedList(restricted.users)

	dbsRaw := stripQuotes(a.Config.Get("mysql_restricted_databases",
		"information_schema performance_schema mysql phpmyadmin sys mariadb.sys"))
	restricted.databases = splitTrimQuotes(dbsRaw)
	restricted.dbsSQL = sqlQuotedList(restricted.databases)
}

func splitTrimQuotes(raw string) []string {
	var out []string
	for _, f := range strings.Fields(raw) {
		f = strings.Trim(strings.TrimSpace(f), `"'`)
		out = append(out, f)
	}
	return out
}

func sqlQuotedList(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = "'" + n + "'"
	}
	return strings.Join(quoted, ", ")
}

func isRestrictedUser(name string) bool {
	lower := strings.ToLower(name)
	for _, u := range restricted.users {
		if u == lower {
			return true
		}
	}
	return false
}

func isRestrictedDatabase(name string) bool {
	for _, d := range restricted.databases {
		if d == name {
			return true
		}
	}
	return false
}

// mysqlStartupTime is the max SELECT 1 retry attempts when retry=true,
// read once at Register() time from the mysql_startup_time config value.
var mysqlStartupTime = 10

// mysqlImportMaxSizeGB is the max upload size for a database import,
// read once at Register() time from the mysql_import_max_size_gb config value.
var mysqlImportMaxSizeGB = "1"

func loadTuningConfig(a *appctx.App) {
	if v, err := strconv.Atoi(a.Config.Get("mysql_startup_time", "10")); err == nil {
		mysqlStartupTime = v
	}
	mysqlImportMaxSizeGB = a.Config.Get("mysql_import_max_size_gb", "1")
}

// escapeMySQLString applies the minimal escaping needed for a value
// interpolated into a single-quoted SQL string literal (backslash first,
// so it doesn't double-escape the quote's own escape).
func escapeMySQLString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// GetMySQLVersion returns the configured MySQL server type (mysql/mariadb),
// memoized 6h.
func GetMySQLVersion(ctx context.Context, a *appctx.App, userContext string) string {
	version, _ := cache.Memoize(ctx, a.Cache, "get_mysql_version:"+userContext, 6*time.Hour, func() (string, error) {
		return webserver.GetEnvFileValue(userContext, "MYSQL_TYPE"), nil
	})
	return version
}

// CheckMySQLInsideContainer confirms the per-user mysqld actually answers
// SELECT 1, optionally retrying once a second for up to mysqlStartupTime
// attempts (the wp-install-style call site).
func CheckMySQLInsideContainer(ctx context.Context, userContext string, retry bool) bool {
	attempt := 0
	for {
		if _, err := mysqlmanager.Exec(ctx, userContext, "SELECT 1", ""); err == nil {
			return true
		}
		attempt++
		if retry && attempt < mysqlStartupTime {
			time.Sleep(time.Second)
			continue
		}
		return false
	}
}

// CheckMySQLNotTemporary is the fallback check once
// CheckMySQLInsideContainer's own retries are exhausted: it just keeps
// polling SELECT 1 a bit longer (up to 30s) and uses the database as soon
// as it answers. Previously this scraped the container's logs for
// mysqld's "Temporary server stopped" line (printed partway through a
// crash-recovery/first-boot sequence, before the real server starts) -
// dropped because a container that's simply been up and healthy for a
// while never logs that line at all (no bootstrap ever happened), so that
// check produced a false "not ready" for a database that was actually
// fine the whole time. mysqlVersion is unused now but kept in the
// signature to avoid churning every CMS installer's call site.
func CheckMySQLNotTemporary(ctx context.Context, userContext, mysqlVersion string) bool {
	const attempts = 6
	for i := 0; i < attempts; i++ {
		if _, err := mysqlmanager.Exec(ctx, userContext, "SELECT 1", ""); err == nil {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// resolveContainerHealth wraps docker.GetContainerStatus()'s Health field
// to patch over one specific staleness window: Podman's HEALTHCHECK
// only re-runs on its own interval, so a container can already be answering
// queries fine (e.g. right after this same request just ran a CREATE
// DATABASE through mysqlmanager) while the cached health status still reads
// "starting"/"unhealthy" for a few more seconds. Rather than trust that
// stale field for the "should I even try to query" gate, do one live SELECT
// 1 and treat that as authoritative - this is the same "check real runtime
// state instead of a racy signal" pattern already used for the PHP
// extensions/info start-check and Varnish's enable-action check.
func resolveContainerHealth(ctx context.Context, userContext string, status docker.ContainerStatus) string {
	if status.Health == "healthy" || status.State != "running" {
		return status.Health
	}
	if CheckMySQLInsideContainer(ctx, userContext, false) {
		return "healthy"
	}
	return status.Health
}

// toStringCell converts one mysqlmanager.Exec result cell to a string. The
// MySQL driver hands back []byte for VARCHAR/TEXT columns when scanned into
// `any` (not string), so every textual column read via Exec needs this
// rather than a bare type assertion.
func toStringCell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

// toFloatCell converts one mysqlmanager.Exec result cell (typically a
// ROUND(...) aggregate) to a float64, tolerating the driver's []byte
// encoding the same way mysqlmanager.ToInt does for integers.
func toFloatCell(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

// mysqlContainerStatusDetail builds the explanatory text shown in the
// databases/users table's empty-state row while the service isn't
// running/healthy - computed once server-side instead of duplicating a
// 9-branch chain in two templates. Returns "" for the
// running+healthy case, where the real table rows render instead. Plain
// (untranslated) English text, matching every flash message in this
// codebase - translation isn't wired up for either yet.
func mysqlContainerStatusDetail(containerState, healthStatus string) string {
	switch containerState {
	case "not_found":
		return "Service installation is underway and may take up to one minute."
	case "created":
		return "Service has been created but not yet started. Start it from the Services page."
	case "restarting":
		return "Service is restarting. This may indicate a configuration error or a crash — check the logs on the Services page."
	case "paused":
		return "Service is paused. Resume it from the Services page."
	case "exited":
		return "Service has stopped. This is usually caused by resource exhaustion — increase resource limits for this service and then restart it from the Services page."
	case "removing":
		return "Service is being deleted. Please wait for the process to complete."
	case "dead":
		return "Service has crashed and cannot recover. This is usually caused by resource exhaustion — check the logs on the Services page."
	case "running":
		switch healthStatus {
		case "healthy":
			return ""
		case "unhealthy":
			return "Service is unhealthy. Try restarting it from Services, or contact your administrator."
		case "starting":
			return "Service is active, but the databases are still initializing."
		default:
			return "Unable to retrieve service status. Please contact your administrator."
		}
	default:
		return "Unable to retrieve service status. Please contact your administrator."
	}
}

// isSystemMySQLDatabase reports whether name is a built-in system database:
// these get a "System Database" badge instead of import/export/
// optimize/repair/delete actions, and are excluded from the
// zero-user-databases health-toast warning.
func isSystemMySQLDatabase(name string) bool {
	switch name {
	case "information_schema", "mysql", "phpmyadmin", "performance_schema", "sys", "mariadb.sys":
		return true
	default:
		return false
	}
}

// mysqlWarningFlashMessage builds the shorter, service-name-interpolated
// flash message shown above the databases/users table - distinct from
// mysqlContainerStatusDetail()'s longer table-body text, even though both
// branch on the same two fields.
func mysqlWarningFlashMessage(mysqlVersion, containerState, healthStatus string) string {
	switch {
	case containerState == "not_found":
		return mysqlVersion + " service is not yet installed. Starting it in the background.."
	case containerState != "running":
		return mysqlVersion + " service is not accessible. Current status: " + containerState + "."
	case healthStatus == "unhealthy":
		return mysqlVersion + " is running but still initializing. If it stays in this state for more than 60s, check the logs."
	case healthStatus == "starting":
		return mysqlVersion + " is starting. It is not ready to handle queries yet."
	case healthStatus != "healthy":
		return mysqlVersion + " is running, but health status is unknown. It may or may not be ready to handle queries."
	default:
		return ""
	}
}
