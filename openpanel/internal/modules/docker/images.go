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
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleContainersChangeImage changes a service's image tag, or, with no service in the path, shows the picker of services to change
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
		if !imageChangeable(service) {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "The image of %(service)s can't be changed.", "service", service), "/containers/image/change")
			return
		}
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			value := strings.TrimSpace(r.Form.Get("new_tag"))
			if !imageTagRE.MatchString(value) {
				flashAndRedirect(a, w, r, "error", "Invalid image tag.", fmt.Sprintf("/containers/image/change/%s", service))
				return
			}
			envVar, _ := imageTagVar(userContext, service)

			result := StartOrStopContainer(ctx, userContext, service, "deactivate", "")
			if result.Success {
				SetEnvValue(userContext, envVar, value)
				_ = logger.RecordUserAction(a.Config, username, fmt.Sprintf("changed image tag for %s to %s", service, value), reqip.ClientIP(r))
				flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Successfully changed image tag for %(service)s to %(value)s!", "service", service, "value", value), "/containers/image/change")
				return
			}
			flashAndRedirect(a, w, r, "error", "Failed to stop the service in order to delete old image.", fmt.Sprintf("/containers/image/change/%s", service))
			return
		}

		envVar, defaultTag := imageTagVar(userContext, service)
		currentVersion, _ := GetEnvValue(userContext, envVar)
		if currentVersion == "" {
			currentVersion = defaultTag
		}

		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, []any{service, currentVersion})
			return
		}
		renderChangeImagePage(a, w, r, service, currentVersion, dockerHubPage(userContext, service))
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
