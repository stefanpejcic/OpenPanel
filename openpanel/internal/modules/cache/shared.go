package cache

import (
	"context"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

func cacheInjected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

// restartWebserverAfterVarnishToggle brings the webserver back up after a Varnish enable/disable removed it, verifying it's actually running rather than trusting `podman-compose up`'s exit code alone - that command can exit 0 while the webserver still isn't running, since podman-compose's config-hash reconciliation touches every stale-hash service on each invocation and a transient "container name already in use" elsewhere can leave the requested one never created. One retry clears that; only report failure if it's still not up after that.
func restartWebserverAfterVarnishToggle(ctx context.Context, userContext, webserver string) docker.StartStopResult {
	result := docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "run")
	if result.Success && docker.WaitForServiceRunning(ctx, userContext, webserver) {
		return result
	}

	docker.ForceRemoveContainer(ctx, userContext, webserver)
	result = docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "run")
	if result.Success && !docker.WaitForServiceRunning(ctx, userContext, webserver) {
		return docker.StartStopResult{Success: false, Message: "container did not reach a running state"}
	}
	return result
}
