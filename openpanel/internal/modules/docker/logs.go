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

// FetchContainerLog runs `podman logs` for the given service, cached 60s, and returns (body, httpStatus).
func FetchContainerLog(ctx context.Context, a *appctx.App, userContext, serviceName string, tail int) (string, int) {
	type result struct {
		Body   string
		Status int
	}
	key := "_fetch_container_log:" + userContext + ":" + serviceName + ":" + strconv.Itoa(tail)
	r, _ := cache.Memoize(ctx, a.Cache, key, 60*time.Second, func() (result, error) {
		body, status := runPodmanLogs(ctx, userContext, "logs", "--tail", strconv.Itoa(tail), serviceName)
		return result{Body: body, Status: status}, nil
	})
	return r.Body, r.Status
}

// FetchContainerLogSince is uncached and skips lines older than since, podman keeps logs from before a restart
func FetchContainerLogSince(ctx context.Context, userContext, serviceName string, tail int, since time.Duration) (string, int) {
	return runPodmanLogs(ctx, userContext, "logs", "--since", strconv.FormatInt(int64(since.Seconds()), 10)+"s", "--tail", strconv.Itoa(tail), serviceName)
}

func runPodmanLogs(ctx context.Context, userContext string, args ...string) (string, int) {
	argv := podmanmanager.PodmanArgv(userContext, args...)
	cmdCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := podmanmanager.Command(cmdCtx, userContext, argv).CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "Error: " + string(out), http.StatusInternalServerError
		}
		return "Unhandled error: " + err.Error(), http.StatusInternalServerError
	}
	return string(out), http.StatusOK
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
