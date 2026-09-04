package docker

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// hasOnlyRestrictedDatabases reports whether the currently active
// mysql/mariadb server has no real user-created databases - only the
// system ones every fresh install ships with (mysql_restricted_databases) -
// in which case it's safe to force a type switch without making the user
// go delete anything first. Any failure to check (server unreachable,
// query error) returns false, falling back to the safe "must stop it
// manually" path rather than risking a silent wipe.
func hasOnlyRestrictedDatabases(ctx context.Context, a *appctx.App, userContext string) bool {
	raw := a.Config.Get("mysql_restricted_databases", "information_schema performance_schema mysql phpmyadmin sys mariadb.sys")
	fields := strings.Fields(strings.Trim(strings.TrimSpace(raw), `"'`))
	quoted := make([]string, len(fields))
	for i, d := range fields {
		quoted[i] = "'" + strings.Trim(strings.TrimSpace(d), `"'`) + "'"
	}
	query := "SELECT COUNT(*) AS total FROM information_schema.schemata WHERE schema_name NOT IN (" + strings.Join(quoted, ", ") + ")"
	rows, err := mysqlmanager.Exec(ctx, userContext, query, "")
	if err != nil || len(rows) == 0 || len(rows[0]) == 0 {
		return false
	}
	return mysqlmanager.ToInt(rows[0][0]) == 0
}

// handleContainersMySQL swaps between mysql/mariadb, wiping the old data
// volume.
//
// Every failure branch (stop failed, start failed) returns immediately
// with a redirect carrying the specific error message, rather than
// falling through to a generic "Invalid mysql server selected" flash that
// would bury the real cause.
func handleContainersMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	mysqlType, _ := GetEnvValue(userContext, "MYSQL_TYPE")
	available := "mysql"
	if mysqlType == "mysql" {
		available = "mariadb"
	}

	if r.Method == http.MethodPost {
		outputJSON := r.URL.Query().Get("output") == "json"
		jsonErr := func(status int, msg string) {
			writeJSONError(w, status, msg)
		}

		if IsServiceRunning(ctx, userContext, mysqlType) && !hasOnlyRestrictedDatabases(ctx, a, userContext) {
			msg := fmt.Sprintf("Existing databases must first be deleted and %s container stopped in order to change mysql type.", mysqlType)
			if outputJSON {
				jsonErr(http.StatusConflict, msg)
				return
			}
			flashAndRedirect(a, w, r, "error", msg, "/containers/mysql")
			return
		}

		_ = r.ParseForm()
		newSQL := r.Form.Get("new_sql")
		if newSQL == "mysql" || newSQL == "mariadb" {
			stopResp := StartOrStopContainer(ctx, userContext, mysqlType, "deactivate", "")
			if !stopResp.Success {
				if outputJSON {
					jsonErr(http.StatusInternalServerError, stopResp.Message)
					return
				}
				flashAndRedirect(a, w, r, "error", stopResp.Message, "/containers/mysql")
				return
			}

			var flashes [][2]string
			if stopResp.Message != "" {
				flashes = append(flashes, [2]string{"warning", stopResp.Message})
			}

			deleteDockerVolume(ctx, userContext, userContext+"_mysql_data")
			StartOrStopContainer(ctx, userContext, "phpmyadmin", "deactivate", "")
			startResp := StartOrStopContainer(ctx, userContext, newSQL, "activate", "")
			if !startResp.Success {
				msg := fmt.Sprintf("Failed to start %s.", newSQL)
				if outputJSON {
					jsonErr(http.StatusInternalServerError, msg)
					return
				}
				flashes = append(flashes, [2]string{"error", msg})
				redirectWithFlashes(a, w, r, "/containers/mysql", flashes...)
				return
			}

			SetEnvValue(userContext, "MYSQL_TYPE", newSQL)
			removeImage(ctx, userContext, mysqlType)
			mysqlmanager.InvalidatePool(userContext)

			_ = logger.RecordUserAction(a.Config, username, "switched mysql type to: "+newSQL, reqip.ClientIP(r))
			successMsg := fmt.Sprintf("Successfully switched to %s!", newSQL)
			if outputJSON {
				writeJSON(w, map[string]string{"message": successMsg})
				return
			}
			flashes = append(flashes, [2]string{"success", successMsg})
			redirectWithFlashes(a, w, r, "/containers/mysql", flashes...)
			return
		}

		if outputJSON {
			jsonErr(http.StatusBadRequest, "Invalid mysql server selected")
			return
		}
		flashAndRedirect(a, w, r, "error", "Invalid mysql server selected", "/containers/mysql")
		return
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, mysqlType)
		return
	}

	renderChangeMySQLPage(a, w, r, mysqlType, available)
}

var webserverOptions = []string{"apache", "nginx", "openresty", "openlitespeed", "litespeed"}

// handleContainersWebserver swaps the active webserver container. Same
// clean-early-return choice as handleContainersMySQL above (see its
// comment): every failure branch redirects immediately with its specific
// error message.
func handleContainersWebserver(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	webserver, _ := GetEnvValue(userContext, "WEB_SERVER")
	var available []string
	for _, ws := range webserverOptions {
		if ws != webserver {
			available = append(available, ws)
		}
	}

	userDomains, domainsErr := a.AllDomainsForUser(ctx, userID)

	if r.Method == http.MethodPost {
		outputJSON := r.URL.Query().Get("output") == "json"

		if domainsErr == nil && len(userDomains) > 0 {
			msg := fmt.Sprintf("Existing domains (%d) must first be removed in order to change webserver.", len(userDomains))
			if outputJSON {
				writeJSONError(w, http.StatusConflict, msg)
				return
			}
			flashAndRedirect(a, w, r, "error", msg, "/containers/webserver")
			return
		}

		_ = r.ParseForm()
		newWebserver := r.Form.Get("new_ws")
		if newWebserver != "" && containsString(available, newWebserver) {
			// Varnish only ever proxies to whichever webserver is active, so
			// its compose block is the only one kept on PROXY_HTTP_PORT
			// (everything else, including a freshly-activated webserver,
			// defaults to the flat HTTP_PORT varnish itself listens on -
			// see SwapWebserverComposePort). Without re-pointing that swap
			// at the new webserver here, varnish and the new webserver
			// would both try to bind the same public port.
			varnishRunning := IsServiceRunning(ctx, userContext, "varnish")
			if varnishRunning {
				_ = SwapWebserverComposePort(userContext, newWebserver, "on")
				_ = SwapWebserverComposePort(userContext, webserver, "off")
			}

			stopResp := StartOrStopContainer(ctx, userContext, webserver, "deactivate", "")
			if !stopResp.Success {
				if outputJSON {
					writeJSONError(w, http.StatusInternalServerError, stopResp.Message)
					return
				}
				flashAndRedirect(a, w, r, "error", stopResp.Message, "/containers/webserver")
				return
			}

			deleteDockerVolume(ctx, userContext, userContext+"_webserver_data")
			startResp := StartOrStopContainer(ctx, userContext, newWebserver, "activate", "")
			if !startResp.Success {
				msg := fmt.Sprintf("Failed to start %s.", newWebserver)
				if outputJSON {
					writeJSONError(w, http.StatusInternalServerError, msg)
					return
				}
				flashAndRedirect(a, w, r, "error", msg, "/containers/webserver")
				return
			}

			SetEnvValue(userContext, "WEB_SERVER", newWebserver)
			removeImage(ctx, userContext, webserver)

			if varnishRunning {
				// Varnish's backend host gets baked into its container env
				// (and from there its VCL) at creation time from
				// WEB_SERVER - a plain `podman-compose restart` reuses the
				// existing container as-is and doesn't re-resolve
				// ${WEB_SERVER} (confirmed live: the baked-in env stayed on
				// the old webserver after a restart), so it has to be torn
				// down and recreated instead, same as every other swap in
				// this file.
				StartOrStopContainer(ctx, userContext, "varnish", "deactivate", "")
				StartOrStopContainer(ctx, userContext, "varnish", "activate", "run")
			}

			_ = logger.RecordUserAction(a.Config, username, "switched webserver type to: "+newWebserver, reqip.ClientIP(r))
			successMsg := fmt.Sprintf("Successfully switched to %s!", newWebserver)
			if outputJSON {
				writeJSON(w, map[string]string{"message": successMsg})
				return
			}
			flashAndRedirect(a, w, r, "success", successMsg, "/containers/webserver")
			return
		}

		if outputJSON {
			writeJSONError(w, http.StatusBadRequest, "Invalid web server selected")
			return
		}
		flashAndRedirect(a, w, r, "error", "Invalid web server selected", "/containers/webserver")
		return
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, webserver)
		return
	}

	renderChangeWebserverPage(a, w, r, webserver, available, userDomains)
}
