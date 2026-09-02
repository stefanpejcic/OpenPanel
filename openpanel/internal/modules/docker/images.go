package docker

import (
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleContainersChangeImage changes a service's image tag, or, with no
// service in the path, shows the picker of services to change.
func handleContainersChangeImage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	if service != "" {
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			value := r.Form.Get("new_tag")

			result := StartOrStopContainer(ctx, userContext, service, "deactivate", "")
			if result.Success {
				SetEnvValue(userContext, service+"_VERSION", value)
				_ = logger.RecordUserAction(a.Config, username, fmt.Sprintf("changed image tag for %s to %s", service, value), reqip.ClientIP(r))
				flashAndRedirect(a, w, r, "success", fmt.Sprintf("Successfully changed image tag for %s to %s!", service, value), "/containers/image/change")
				return
			}
			flashAndRedirect(a, w, r, "error", "Failed to stop the service in order to delete old image.", fmt.Sprintf("/containers/image/change/%s", service))
			return
		}

		var currentVersion string
		switch {
		case service == "phpmyadmin":
			currentVersion, _ = GetEnvValue(userContext, "PMA_VERSION")
		case service == userContext:
			currentVersion, _ = GetEnvValue(userContext, "OS")
		case service == "mariadb":
			currentVersion, _ = GetEnvValue(userContext, "MYSQL_VERSION")
		default:
			currentVersion, _ = GetEnvValue(userContext, strings.ToUpper(service)+"_VERSION")
		}

		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, []any{service, currentVersion})
			return
		}
		renderChangeImagePage(a, w, r, service, currentVersion)
		return
	}

	composeData, err := podmanmanager.LoadComposeConfig(ctx, userContext)
	if err != nil {
		composeData = map[string]any{"error": "Failed to fetch container data", "details": err.Error()}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, composeData)
		return
	}
	renderChangeImageSelectPage(a, w, r, composeData)
}
