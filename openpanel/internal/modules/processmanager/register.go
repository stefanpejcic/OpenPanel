package processmanager

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Register wires the process manager page route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "process_manager")(h)
	}
	mux.Handle("/process-manager", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleProcessManager(a, w, r) }))
	mux.Handle("POST /process-manager/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, processBulkActions(web.RequestTranslator(a, r)), func(_, _, key string) (*web.BulkCall, web.BulkResult) {
			container, pid, ok := strings.Cut(key, "/")
			if !ok {
				return web.Skip("Invalid process.")
			}
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/process-manager", JSON: map[string]string{"pid_to_kill": pid, "container": container}})
		})
	}))
}

func processBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "kill", Label: t.Get("Kill"), Confirm: t.Get("Kill the selected processes?"), Danger: true},
	}
}
