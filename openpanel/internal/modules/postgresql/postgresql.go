// Package postgresql implements database/user CRUD, the creation wizard,
// privilege assignment, import, per-service configuration, remote access,
// and the process list - all built on internal/core/postgresmanager's
// per-(user, database) connection pool.
package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// restrictedUsers is the set of built-in/system PostgreSQL role names that
// users are never allowed to create, edit, or delete.
var restrictedUsers = map[string]bool{
	"postgres": true, "pg_signal_backend": true, "pg_read_all_data": true,
	"pg_write_all_data": true, "pg_monitor": true, "pg_read_all_settings": true,
	"pg_read_server_files": true, "pg_write_server_files": true,
	"pg_execute_server_program": true, "replication": true,
}

func isRestrictedUser(name string) bool {
	return restrictedUsers[strings.ToLower(name)]
}

// restrictedDatabases is the set of built-in/system database names that
// users are never allowed to create, rename, or drop.
var restrictedDatabases = map[string]bool{
	"postgres": true, "template0": true, "template1": true,
	"information_schema": true, "pg_catalog": true,
}

func isRestrictedDatabase(name string) bool {
	return restrictedDatabases[name]
}

// filteredRoles is the longer role-name exclusion list used by pg_roles
// queries - a superset of restrictedUsers, including roles that were never
// registration-blocked but shouldn't show up in the regular "show system
// users" toggle.
var filteredRoles = []string{
	"postgres", "pg_signal_backend", "pg_read_all_data", "pg_write_all_data",
	"pg_read_all_settings", "pg_read_server_files", "pg_write_server_files",
	"pg_execute_server_program", "pg_monitor", "pg_database_owner",
	"pg_read_all_stats", "pg_stat_scan_tables", "pg_checkpoint", "pg_maintain",
	"pg_use_reserved_connections", "pg_create_subscription",
}

func filteredRolesSQL() string {
	out := ""
	for i, r := range filteredRoles {
		if i > 0 {
			out += ","
		}
		out += "'" + r + "'"
	}
	return out
}

func isSystemDatabase(name string) bool {
	switch name {
	case "postgres", "template0", "public", "template1":
		return true
	default:
		return false
	}
}

func isSystemUser(name string) bool {
	for _, r := range filteredRoles {
		if r == name {
			return true
		}
	}
	return false
}

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

// toStringCell converts one postgresmanager.Exec result cell to a string.
// lib/pq hands back native Go strings for text columns (unlike the MySQL
// driver's []byte), but this stays tolerant of both.
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

// genericCellString stringifies whatever concrete type the driver hands
// back for an arbitrary pg_stat_activity column (int64, time.Time, bool,
// ...), for building the pipe-delimited process list line. Unlike
// toStringCell (strict, text-column-only), this accepts any type.
func genericCellString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

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

// checkPostgresInsideContainer makes a single connection attempt (no
// retry/polling loop - that behavior belongs to the install wizard, out of
// scope here).
func checkPostgresInsideContainer(ctx context.Context, userContext string) bool {
	_, err := postgresmanager.Exec(ctx, userContext, "SELECT 1", "postgres")
	return err == nil
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
