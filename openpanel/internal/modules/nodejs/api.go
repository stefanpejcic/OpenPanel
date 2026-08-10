package nodejs

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// RegisterAPI wires POST /api/nodejs/install onto mux, gated behind the
// "nodejs" feature flag. Reuses appinstall.HandleInstall as-is: it's
// already pure form-value-in/NDJSON-out with no session or flash usage,
// so it needs no API-specific variant. Input is the same form-encoded
// fields the web install form posts (domain_id, service_name,
// startup_file, cpu_limit, mem_limit, port, subdirectory, version,
// custom_cmd, requirements, git_repo_url); the response streams
// newline-delimited JSON status/error events as the install progresses.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "nodejs", "POST /api/nodejs/install", func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstall(appinstall.NodeJS, a, w, r)
	})
}
