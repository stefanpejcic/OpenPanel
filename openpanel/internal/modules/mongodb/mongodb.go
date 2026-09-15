// Package mongodb implements database/user CRUD, the creation wizard, role assignment, and
// import - all built on internal/core/mongomanager's per-user Mongo client.
package mongodb

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// restrictedUsers is the set of built-in/system Mongo user names that users are never allowed to create, edit, or delete
var restrictedUsers = map[string]bool{"admin": true, "root": true}

func isRestrictedUser(name string) bool {
	return restrictedUsers[strings.ToLower(name)]
}

func isRestrictedDatabase(name string) bool {
	return mongomanager.IsSystemDatabase(name)
}

// mongoRoles are the database-scoped roles selectable from the assign/wizard forms.
var mongoRoles = []string{"read", "readWrite", "dbAdmin", "dbOwner"}

func isValidRole(role string) bool {
	for _, r := range mongoRoles {
		if r == role {
			return true
		}
	}
	return false
}

// mongoContainerStatusDetail builds the explanatory text shown in the databases table's
// empty-state row while the service isn't running/healthy - mirrors postgresContainerStatusDetail
// so both templates give the same guidance, returns "" for the running+healthy case
func mongoContainerStatusDetail(containerState, healthStatus string) string {
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

// flashSess adds a flash message without redirecting - several handlers here fall through to the same GET rendering logic below on error rather than redirecting
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

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// checkMongoInsideContainer makes a single connection attempt - no retry/polling loop, that behavior belongs to the install wizard, out of scope here
func checkMongoInsideContainer(ctx context.Context, userContext string) bool {
	return mongomanager.Ping(ctx, userContext)
}

// rolesDisplay joins a user's {role, db} pairs into "role@db, role@db" for the users table.
func rolesDisplay(roles []mongomanager.UserRole) string {
	parts := make([]string, 0, len(roles))
	for _, r := range roles {
		parts = append(parts, r.Role+"@"+r.DB)
	}
	return strings.Join(parts, ", ")
}

// formatSize renders a byte count as a human-readable MB/KB/B string for the databases table.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return strconv.FormatFloat(float64(bytes)/(1024*1024), 'f', 2, 64) + " MB"
	case bytes >= 1024:
		return strconv.FormatFloat(float64(bytes)/1024, 'f', 2, 64) + " KB"
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}
