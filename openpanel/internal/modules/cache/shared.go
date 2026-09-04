package cache

import (
	"context"
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
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

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// flashSess adds a flash message without redirecting - the POST branches
// here fall through to the same GET rendering logic below rather than
// redirecting, so a successful varnish/generic-service action re-renders
// the page directly instead of round-tripping through a redirect.
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

// restartWebserverAfterVarnishToggle brings the webserver back up after a
// Varnish enable/disable removed it (to pick up the swapped port mapping),
// verifying it's actually running rather than trusting `podman-compose
// up`'s exit code alone - confirmed live, that command can exit 0 while
// the webserver still isn't running: podman-compose's own config-hash
// reconciliation touches every service with a stale hash on each compose
// invocation (not just the one named, even with --no-deps), and a
// transient "container name already in use" on one of those unrelated
// services can occasionally leave the actually-requested one never
// created. One retry (a fresh force-remove + activate cycle) clears that
// transient state; only report failure if the webserver still isn't up
// after that.
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
