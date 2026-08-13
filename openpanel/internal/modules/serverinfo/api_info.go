package serverinfo

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterInfoAPI wires GET /api/server/info onto mux: a single-call
// convenience endpoint combining the three buildHosting* payloads that
// /server/info's page already fetches client-side as separate
// /json/system/hosting/* (and /api/hosting/*) calls.
func RegisterInfoAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "info", "GET /api/server/info", func(w http.ResponseWriter, r *http.Request) { handleServerInfoAPI(a, w, r) })
}

func handleServerInfoAPI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)
	planID, _ := data["hosting_plan"].(int)

	writeJSON(w, http.StatusOK, map[string]any{
		"info":  buildHostingInfo(a, r, username),
		"plan":  buildHostingPlan(a, r, userContext, planID),
		"ports": buildHostingPorts(username),
	})
}
