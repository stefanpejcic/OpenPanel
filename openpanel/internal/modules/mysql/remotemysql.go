package mysql

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var (
	identifiedByPasswordRE = regexp.MustCompile(`IDENTIFIED BY PASSWORD '[^']*'`)
	identifiedByRE         = regexp.MustCompile(`IDENTIFIED BY '[^']*'`)
)

// cloneMySQLGrants copies a user's grants from one host to another,
// stripping any embedded credential clause (the new host keeps whatever
// password it was created with).
func cloneMySQLGrants(ctx context.Context, userContext, dbUser, sourceHost, targetHost string) {
	rows, err := mysqlmanager.Exec(ctx, userContext, "SHOW GRANTS FOR '"+dbUser+"'@'"+sourceHost+"'", "")
	if err != nil {
		return
	}
	for _, row := range rows {
		grantStmt := toStringCell(row[0])
		if strings.Contains(grantStmt, "GRANT USAGE ON *.*") && !strings.Contains(grantStmt, "WITH GRANT OPTION") {
			continue
		}
		grantStmt = identifiedByPasswordRE.ReplaceAllString(grantStmt, "")
		grantStmt = identifiedByRE.ReplaceAllString(grantStmt, "")
		grantStmt = strings.ReplaceAll(grantStmt, "`"+dbUser+"`@`"+sourceHost+"`", "`"+dbUser+"`@`"+targetHost+"`")
		grantStmt = strings.ReplaceAll(grantStmt, "'"+dbUser+"'@'"+sourceHost+"'", "'"+dbUser+"'@'"+targetHost+"'")
		_, _ = mysqlmanager.Exec(ctx, userContext, grantStmt, "")
	}
}

// RemoteUserAccess is one row of remote_mysql.html's per-user access table.
type RemoteUserAccess struct {
	Username string
	Hosts    []string
}

var mysqlPortRE = regexp.MustCompile(`^(127\.0\.0\.1:)?(\d+):3306$`)

// handleRemoteMySQL shows remote-access status/port and per-user allowed hosts.
func handleRemoteMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	const mysqlPort = 3306
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)

	mysqlRemotePortOriginal := webserver.GetEnvFileValue(userContext, "MYSQL_PORT")
	m := mysqlPortRE.FindStringSubmatch(mysqlRemotePortOriginal)
	if m == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	remoteMySQLPortOnly := m[2]

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		ipAddress := reqip.ClientIP(r)
		portChanged := false
		switch r.Form.Get("action") {
		case "enable":
			docker.SetEnvValue(userContext, "MYSQL_PORT", remoteMySQLPortOnly+":3306")
			_ = logger.RecordUserAction(a.Config, currentUsername, "enabled remote MySQL", ipAddress)
			flashSess(a, w, r, "success", "Remote MySQL access is now enabled.")
			portChanged = true
		case "disable":
			if strings.Contains(mysqlRemotePortOriginal, "127.0.0.1") {
				flashSess(a, w, r, "info", "Remote MySQL access is already disabled.")
			} else {
				docker.SetEnvValue(userContext, "MYSQL_PORT", "127.0.0.1:"+remoteMySQLPortOnly+":3306")
				_ = logger.RecordUserAction(a.Config, currentUsername, "disabled remote MySQL", ipAddress)
				flashSess(a, w, r, "success", "Remote MySQL access is now disabled.")
				portChanged = true
			}
		}
    if portChanged {
        realServiceName, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE") // "sql" -> "mariadb"/"mysql"
        result := docker.RestartContainer(ctx, userContext, realServiceName)
        if !result.Success {
            flashSess(a, w, r, "error", "Port changed but the MySQL service failed to restart. Try restarting it manually from Services.")
        }
    } else {
        docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
    }
}

	mysqlRemotePortOriginal = webserver.GetEnvFileValue(userContext, "MYSQL_PORT")
	remoteMySQLDisplay := "ON"
	if strings.Contains(mysqlRemotePortOriginal, "127.0.0.1") {
		remoteMySQLDisplay = "OFF"
	}

	var userAccess []RemoteUserAccess
	if remoteMySQLDisplay == "ON" {
		rows, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT User, Host FROM mysql.user WHERE User NOT IN ("+restricted.usersSQL+") ORDER BY User, Host", "")
		if execErr != nil {
			flashSess(a, w, r, "error", "Could not load MySQL users yet - the database may still be starting. Please refresh in a moment.")
		} else {
			var order []string
			grouped := map[string][]string{}
			for _, row := range rows {
				u, h := toStringCell(row[0]), toStringCell(row[1])
				if _, ok := grouped[u]; !ok {
					order = append(order, u)
				}
				grouped[u] = append(grouped[u], h)
			}
			for _, u := range order {
				userAccess = append(userAccess, RemoteUserAccess{Username: u, Hosts: grouped[u]})
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"remote_mysql_display": remoteMySQLDisplay, "server_ip": serverIP, "container_port": remoteMySQLPortOnly,
			"mysql_version": mysqlVersion, "mysql_port": mysqlPort, "user_access": userAccess,
		})
		return
	}

	renderRemoteMySQLPage(a, w, r, mysqlVersion, serverIP, remoteMySQLPortOnly, remoteMySQLDisplay, mysqlPort, userAccess)
}

// handleRemoteMySQLAccessAdd grants a user remote access from a new host.
func handleRemoteMySQLAccessAdd(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	dbUser := strings.TrimSpace(r.Form.Get("db_user"))
	dbHost := strings.TrimSpace(r.Form.Get("db_host"))
	password := r.Form.Get("password")

	switch {
	case dbUser == "" || !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "A valid username is required - alphanumeric characters and '_' only.", "/mysql/remote-mysql")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/mysql/remote-mysql")
		return
	case dbHost == "" || !validators.IsValidHost(dbHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/remote-mysql")
		return
	case password == "":
		flashAndRedirect(a, w, r, "error", "Password is required.", "/mysql/remote-mysql")
		return
	}

	escapedPassword := escapeMySQLString(password)

	existing, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT Host FROM mysql.user WHERE User = '"+dbUser+"'", "")
	if execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error adding access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}
	existingHosts := make([]string, 0, len(existing))
	for _, row := range existing {
		existingHosts = append(existingHosts, toStringCell(row[0]))
	}
	for _, h := range existingHosts {
		if h == dbHost {
			flashAndRedirect(a, w, r, "error", "User "+dbUser+" already has access from host "+dbHost+".", "/mysql/remote-mysql")
			return
		}
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error adding access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}
	if len(existingHosts) > 0 {
		cloneMySQLGrants(ctx, userContext, dbUser, existingHosts[0], dbHost)
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error adding access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "granted MySQL remote access for user "+dbUser+" from host "+dbHost, ipAddress)
	flashSess(a, w, r, "success", "Successfully granted access for "+dbUser+" from "+dbHost)
	http.Redirect(w, r, "/mysql/remote-mysql", http.StatusFound)
}

// handleRemoteMySQLAccessEdit changes the host a remote-access entry is valid from.
func handleRemoteMySQLAccessEdit(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	dbUser := strings.TrimSpace(r.Form.Get("db_user"))
	oldHost := strings.TrimSpace(r.Form.Get("old_host"))
	newHost := strings.TrimSpace(r.Form.Get("new_host"))

	switch {
	case dbUser == "" || !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "A valid username is required.", "/mysql/remote-mysql")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mysql/remote-mysql")
		return
	case !validators.IsValidHost(oldHost) || !validators.IsValidHost(newHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/remote-mysql")
		return
	case oldHost == newHost:
		flashAndRedirect(a, w, r, "info", "No changes made.", "/mysql/remote-mysql")
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "RENAME USER '"+dbUser+"'@'"+oldHost+"' TO '"+dbUser+"'@'"+newHost+"'", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error updating access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error updating access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL remote access host for user "+dbUser+" from "+oldHost+" to "+newHost, ipAddress)
	flashSess(a, w, r, "success", "Successfully updated access for "+dbUser)
	http.Redirect(w, r, "/mysql/remote-mysql", http.StatusFound)
}

// handleRemoteMySQLAccessDelete revokes a user's remote-access entry for one host.
func handleRemoteMySQLAccessDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	dbUser := strings.TrimSpace(r.Form.Get("db_user"))
	dbHost := strings.TrimSpace(r.Form.Get("db_host"))

	switch {
	case dbUser == "" || !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "A valid username is required.", "/mysql/remote-mysql")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be deleted.", "/mysql/remote-mysql")
		return
	case !validators.IsValidHost(dbHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/remote-mysql")
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error removing access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error removing access: "+execErr.Error(), "/mysql/remote-mysql")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "removed MySQL remote access for user "+dbUser+" from host "+dbHost, ipAddress)
	flashSess(a, w, r, "success", "Successfully removed access for "+dbUser+" from "+dbHost)
	http.Redirect(w, r, "/mysql/remote-mysql", http.StatusFound)
}
