package docker

import (
	"context"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// FetchContainerLog runs `podman logs` for the given service, cached 60s.
// Returns (body, httpStatus).
func FetchContainerLog(ctx context.Context, a *appctx.App, userContext, serviceName string, tail int) (string, int) {
	type result struct {
		Body   string
		Status int
	}
	key := "_fetch_container_log:" + userContext + ":" + serviceName + ":" + strconv.Itoa(tail)
	r, _ := cache.Memoize(ctx, a.Cache, key, 60*time.Second, func() (result, error) {
		argv := podmanmanager.PodmanArgv(userContext, "logs", "--tail", strconv.Itoa(tail), serviceName)
		cmdCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		cmd := podmanmanager.Command(cmdCtx, userContext, argv)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return result{Body: "Error: " + string(out), Status: http.StatusInternalServerError}, nil
			}
			return result{Body: "Unhandled error: " + err.Error(), Status: http.StatusInternalServerError}, nil
		}
		return result{Body: string(out), Status: http.StatusOK}, nil
	})
	return r.Body, r.Status
}

func handleContainerLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containerName := r.PathValue("container_name")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)

	if containerName != "" {
		lines := 100
		if v := r.URL.Query().Get("lines"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				lines = n
			}
		}
		body, status := FetchContainerLog(ctx, a, userContext, containerName, lines)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
		return
	}

	var serviceNames []string
	composeData, err := podmanmanager.LoadComposeConfig(ctx, userContext)
	if err == nil {
		if services, ok := composeData["services"].(map[string]any); ok {
			for name := range services {
				serviceNames = append(serviceNames, name)
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, serviceNames)
		return
	}

	renderLogsPage(a, w, r, serviceNames)
}
