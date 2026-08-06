// Package docker handles container list/create/edit/delete, MySQL/webserver
// switching, image management, logs, and the interactive web terminal, all
// built on top of the podmanmanager CLI layer (internal/core/podmanmanager)
// and this package's own status.go/compose.go helpers.
package docker

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires the docker module's routes onto mux, with every route
// gated behind the "docker" feature flag via auth.RequireLogin.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "docker")(h)
	}

	mux.Handle("GET /containers", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersList(a, w, r)
	}))
	// Used by both containers.html and base.html's site-wide service-status
	// widget, so it's registered unconditionally here rather than gated to
	// a single page - it needs the same podmanmanager CLI layer as the
	// rest of this package.
	mux.Handle("GET /json/services", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleServicesStats(a, w, r)
	}))
	mux.Handle("/containers/new", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleAddContainer(a, w, r)
	}))
	mux.Handle("/containers/edit/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleEditContainer(a, w, r)
	}))
	mux.Handle("/containers/delete/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDeleteContainer(a, w, r)
	}))
	mux.Handle("/containers/mysql", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersMySQL(a, w, r)
	}))
	mux.Handle("/containers/webserver", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersWebserver(a, w, r)
	}))
	mux.Handle("/containers/image/", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersImage(a, w, r)
	}))
	mux.Handle("GET /containers/image/change", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersChangeImage(a, w, r)
	}))
	mux.Handle("/containers/image/change/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersChangeImage(a, w, r)
	}))
	mux.Handle("GET /containers/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainerLogs(a, w, r)
	}))
	mux.Handle("GET /containers/logs/{container_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainerLogs(a, w, r)
	}))
	mux.Handle("GET /containers/status", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersStatus(a, w, r)
	}))
	for _, action := range []string{"start", "stop", "restart", "cpu", "ram"} {
		action := action
		mux.Handle("POST /containers/"+action+"/{container_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
			handleManageContainer(a, w, r, action)
		}))
	}
	mux.Handle("GET /containers/terminal", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDockerTerminal(a, w, r)
	}))
	mux.Handle("GET /containers/terminal/{container_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDockerTerminal(a, w, r)
	}))
	mux.Handle("GET /ws/containers/terminal/{container_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDockerTerminalWS(a, w, r)
	}))
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// flashAndRedirect queues a single flash message on the session, then
// redirects to path.
func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// redirectWithFlashes is like flashAndRedirect but for handlers that queue
// more than one flash message before their final redirect (e.g. a
// "container stopped with a warning" message followed by a separate
// success/error message once the rest of the operation finishes).
func redirectWithFlashes(a *appctx.App, w http.ResponseWriter, r *http.Request, path string, flashes ...[2]string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	for _, f := range flashes {
		flash.Add(sess, f[0], f[1])
	}
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}
