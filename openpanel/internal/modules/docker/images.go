package docker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// imageInUse reports whether any container (running or stopped) currently
// references imageRef.
func imageInUse(ctx context.Context, userContext, imageRef string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "ps", "-a", "--format", "{{.Image}}")
	cmd := podmanmanager.Command(ctx, userContext, argv)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == imageRef {
			return true
		}
	}
	return false
}

// handleContainersImage serves the cup.json image-update report,
// refreshed via `opencli docker-images` on POST or when no cached report
// exists yet.
func handleContainersImage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	filePath := homePath(userContext, "docker-data", "cup", "cup.json")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		if r.Form.Get("action") == "delete" {
			imageRef := strings.TrimSpace(r.Form.Get("image"))
			switch {
			case imageRef == "":
				flashAndRedirect(a, w, r, "error", "No image specified.", "/containers/image/")
			case imageInUse(ctx, userContext, imageRef):
				flashAndRedirect(a, w, r, "error", fmt.Sprintf("Cannot delete %s: it is used by a container.", imageRef), "/containers/image/")
			default:
				removeImage(ctx, userContext, imageRef)
				_ = logger.RecordUserAction(a.Config, username, "deleted docker image "+imageRef, reqip.ClientIP(r))
				flashAndRedirect(a, w, r, "success", fmt.Sprintf("Image %s deleted successfully.", imageRef), "/containers/image/")
			}
			return
		}
	}

	var data string
	var lastModified string

	if info, err := os.Stat(filePath); err == nil {
		if raw, err := os.ReadFile(filePath); err == nil {
			data = string(raw)
			lastModified = info.ModTime().Format("2006-01-02 15:04:05")
		}
	}

	if r.Method == http.MethodPost || data == "" {
		cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		out, err := exec.CommandContext(cmdCtx, "opencli", "docker-images", userContext).Output()
		if err == nil {
			data = string(out)
			lastModified = time.Now().Format("2006-01-02 15:04:05")
		} else {
			data = ""
			lastModified = ""
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, []any{data, lastModified})
		return
	}

	renderImagesPage(a, w, r, data, lastModified)
}

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
				flashAndRedirect(a, w, r, "success", fmt.Sprintf("Successfully changed image tag for %s to %s!", service, value), "/containers/image/")
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
