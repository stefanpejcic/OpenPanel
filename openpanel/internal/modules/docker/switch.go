package docker

import (
	"fmt"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleContainersMySQL swaps between mysql/mariadb, wiping the old data
// volume.
//
// Every failure branch (stop-container failed, start-container failed)
// returns immediately with a redirect carrying the specific error
// message, rather than falling through to a generic "Invalid mysql server
// selected" flash that would bury the real cause.
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
		if IsServiceRunning(ctx, userContext, mysqlType) {
			flashAndRedirect(a, w, r, "error",
				fmt.Sprintf("Existing databases must first be deleted and %s container stopped in order to change mysql type.", mysqlType),
				"/containers/mysql")
			return
		}

		_ = r.ParseForm()
		newSQL := r.Form.Get("new_sql")
		if newSQL == "mysql" || newSQL == "mariadb" {
			stopResp := StartOrStopContainer(ctx, userContext, mysqlType, "deactivate", "")
			if !stopResp.Success {
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
				flashes = append(flashes, [2]string{"error", fmt.Sprintf("Failed to start %s.", newSQL)})
				redirectWithFlashes(a, w, r, "/containers/mysql", flashes...)
				return
			}

			SetEnvValue(userContext, "MYSQL_TYPE", newSQL)
			removeImage(ctx, userContext, mysqlType)
			mysqlmanager.InvalidatePool(userContext)

			_ = logger.RecordUserAction(a.Config, username, "switched mysql type to: "+newSQL, reqip.ClientIP(r))
			flashes = append(flashes, [2]string{"success", fmt.Sprintf("Successfully switched to %s!", newSQL)})
			redirectWithFlashes(a, w, r, "/containers/mysql", flashes...)
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
		if domainsErr == nil && len(userDomains) > 0 {
			flashAndRedirect(a, w, r, "error",
				fmt.Sprintf("Existing domains (%d) must first be removed in order to change webserver.", len(userDomains)),
				"/containers/webserver")
			return
		}

		_ = r.ParseForm()
		newWebserver := r.Form.Get("new_ws")
		if newWebserver != "" && containsString(available, newWebserver) {
			// Varnish only ever proxies to whichever webserver is currently
			// active, so its docker-compose.yml block is the only one kept
			// on PROXY_HTTP_PORT (the rest, including a freshly-activated
			// new webserver, default to the flat HTTP_PORT that varnish
			// itself listens on - see SwapWebserverComposePort). Without
			// re-pointing that swap at the new webserver here, both varnish
			// and the new webserver end up trying to bind the same public
			// port.
			varnishRunning := IsServiceRunning(ctx, userContext, "varnish")
			if varnishRunning {
				_ = SwapWebserverComposePort(userContext, newWebserver, "on")
				_ = SwapWebserverComposePort(userContext, webserver, "off")
			}

			stopResp := StartOrStopContainer(ctx, userContext, webserver, "deactivate", "")
			if !stopResp.Success {
				flashAndRedirect(a, w, r, "error", stopResp.Message, "/containers/webserver")
				return
			}

			deleteDockerVolume(ctx, userContext, userContext+"_webserver_data")
			startResp := StartOrStopContainer(ctx, userContext, newWebserver, "activate", "")
			if !startResp.Success {
				flashAndRedirect(a, w, r, "error", fmt.Sprintf("Failed to start %s.", newWebserver), "/containers/webserver")
				return
			}

			SetEnvValue(userContext, "WEB_SERVER", newWebserver)
			removeImage(ctx, userContext, webserver)

			if varnishRunning {
				// Varnish's backend host is baked into its container
				// environment (and from there into its VCL) at container
				// creation time from the WEB_SERVER env var - a plain
				// `podman-compose restart` reuses the existing container
				// as-is and does NOT re-resolve ${WEB_SERVER} (confirmed
				// live: its baked-in env stayed on the old webserver after
				// a restart), so it has to be torn down and recreated
				// instead, same as every other container swap in this
				// file.
				StartOrStopContainer(ctx, userContext, "varnish", "deactivate", "")
				StartOrStopContainer(ctx, userContext, "varnish", "activate", "run")
			}

			_ = logger.RecordUserAction(a.Config, username, "switched webserver type to: "+newWebserver, reqip.ClientIP(r))
			flashAndRedirect(a, w, r, "success", fmt.Sprintf("Successfully switched to %s!", newWebserver), "/containers/webserver")
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
