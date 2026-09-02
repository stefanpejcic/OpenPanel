// Package docker handles container list/create/edit/delete, image
// management, and logs, all built on top of the podmanmanager CLI layer
// (internal/core/podmanmanager) and this package's own status.go/compose.go
// helpers. MySQL/webserver switching, changing a service's image tag, and
// the interactive web terminal live in this same package but are wired up
// by their own Register* functions (RegisterChangeDB, RegisterChangeWS,
// RegisterChangeImage, RegisterTerminal), each gated behind its own feature
// flag independently of "docker" - see registry.go.
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
	mux.Handle("/containers/new", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleAddContainer(a, w, r)
	}))
	mux.Handle("/containers/edit/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleEditContainer(a, w, r)
	}))
	mux.Handle("/containers/delete/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDeleteContainer(a, w, r)
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
	for _, action := range []string{"start", "stop", "restart", "cpu", "ram", "pids"} {
		action := action
		mux.Handle("POST /containers/"+action+"/{container_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
			handleManageContainer(a, w, r, action)
		}))
	}
}

// RegisterServicesJSON wires GET /json/services onto mux. It backs
// containers.html plus base.html's site-wide fetchServiceData helper, which
// the services module's own service cards (system/services.html) and
// cache's redis/memcached widgets also call - so it must work whenever
// either "docker" or "services" is enabled, not just docker. That's why
// it's split out from Register into its own function, gated on both
// feature names, and registered once from modules.RegisterAll rather than
// from either module's own Register (which would either miss the other
// module, or double-register and panic if both are enabled).
func RegisterServicesJSON(mux *http.ServeMux, a *appctx.App) {
	mux.Handle("GET /json/services", auth.RequireLogin(a, "docker", "services")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleServicesStats(a, w, r)
	})))
}

// RegisterTerminal wires the interactive web terminal's routes onto mux,
// gated behind its own "terminal" feature flag rather than "docker" - it
// grants shell access inside a user's containers, a materially bigger
// privilege than the rest of the docker module's container lifecycle
// management, so admins can enable/disable it independently.
func RegisterTerminal(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "terminal")(h)
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

// RegisterChangeImage wires the "change a service's image tag" routes onto
// mux, gated behind its own "change_image" feature flag.
func RegisterChangeImage(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "change_image")(h)
	}

	mux.Handle("GET /containers/image/change", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersChangeImage(a, w, r)
	}))
	mux.Handle("/containers/image/change/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersChangeImage(a, w, r)
	}))
}

// RegisterChangeWS wires the webserver-swap routes onto mux, gated behind
// its own "change_ws" feature flag.
func RegisterChangeWS(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "change_ws")(h)
	}

	mux.Handle("/containers/webserver", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersWebserver(a, w, r)
	}))
}

// RegisterChangeDB wires the MySQL/MariaDB-swap routes onto mux, gated
// behind its own "change_db" feature flag.
func RegisterChangeDB(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "change_db")(h)
	}

	mux.Handle("/containers/mysql", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleContainersMySQL(a, w, r)
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
