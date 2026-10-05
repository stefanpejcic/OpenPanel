package cmsapp

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

type Handler func(a *appctx.App, w http.ResponseWriter, r *http.Request)

// remover is the app's own Remove if it has one, the shared HandleRemove otherwise
func (app *App) remover() Handler {
	if app.Remove != nil {
		return app.Remove
	}
	return app.HandleRemove
}

func (app *App) cloner() Handler {
	if app.Clone != nil {
		return app.Clone
	}
	if app.CanClone {
		return app.HandleClone
	}
	return nil
}

func (app *App) apiCloner() Handler {
	if app.APIClone != nil {
		return app.APIClone
	}
	if app.APICloneForm != nil {
		return app.HandleAPIClone
	}
	return nil
}

// Register wires the app's UI routes, gated behind its feature flag, nil handlers are left out
func (app *App) Register(mux *http.ServeMux, a *appctx.App) {
	route := func(pattern string, h Handler) {
		if h == nil {
			return
		}
		mux.Handle(pattern, auth.RequireLogin(a, app.Slug)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { h(a, w, r) })))
	}
	p := "/" + app.Slug
	route(p+"/install", app.HandleInstallPage)
	route("POST "+p+"/remove", app.remover())
	route("GET "+p+"/login", app.Login)
	route("POST "+p+"/cache", app.Cache)
	route("GET "+p+"/logs", app.Logs)
	route("GET "+p+"/versions", app.Versions)
	route(p+"/maintenance", app.Maintenance)
	route("GET "+p+"/backup/get_dates/{selected_domain...}", app.HandleGetBackupDates)
	route("POST "+p+"/backup/restore/{selected_domain...}", app.HandleRestoreBackup)
	route("POST "+p+"/backup/run/{selected_domain...}", app.HandleRunBackup)
	route("POST "+p+"/clone", app.cloner())
	route("POST "+p+"/update", app.Update)
}

// RegisterAPI wires the app's /api/<slug>/* routes
func (app *App) RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	route := func(pattern string, h Handler) {
		if h == nil {
			return
		}
		apiregistry.Handle(mux, a, app.Slug, pattern, func(w http.ResponseWriter, r *http.Request) { h(a, w, r) })
	}
	p := "/api/" + app.Slug
	route("POST "+p+"/install", app.HandleAPIInstall)
	route("DELETE "+p+"/sites/{site_id}", app.HandleAPIRemove)
	clonePath := "POST " + p + "/sites/{site_id}/clone"
	if app.APIClonePath != "" {
		clonePath = app.APIClonePath
	}
	route(clonePath, app.apiCloner())
	route("POST "+p+"/sites/{site_id}/update", app.APIUpdate)
	route("POST "+p+"/sites/{site_id}/cache", app.APICache)
}
