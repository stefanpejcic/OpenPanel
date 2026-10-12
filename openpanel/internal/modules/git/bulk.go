package git

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func gitBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "pull", Label: t.Get("Pull Updates"), Confirm: t.Get("Pull updates for the selected repositories? Automatic ones are deployed too."), Icon: "bi-cloud-download"},
		{Key: "deploy", Label: t.Get("Deploy"), Confirm: t.Get("Deploy the selected repositories now?"), Icon: "bi-rocket-takeoff"},
		{Key: "mode", Label: t.Get("Change deployment mode"), Confirm: t.Get("Deployment mode for the selected repositories:"), Icon: "bi-sliders",
			Input: &web.BulkInput{Type: "select", Options: []web.BulkOption{
				{Value: DeployAutomatic, Label: t.Get("Automatic")},
				{Value: DeployManual, Label: t.Get("Manual")},
				{Value: DeployNone, Label: t.Get("No deployment")},
			}}},
		{Key: "delete", Label: t.Get("Remove"), Confirm: t.Get("Remove the selected repositories? Files already deployed to the website stay in place."), Danger: true},
	}
}

// handleGitBulk replays each selected repo through /git/action.json, the same path the site tab uses
func handleGitBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	web.ServeBulkDispatch(a, mux, w, r, gitBulkActions(web.RequestTranslator(a, r)), func(action, value, name string) (*web.BulkCall, web.BulkResult) {
		repo, ok := getRepo(userContext, name)
		if !ok {
			return web.Skip("Repository not found, reload the page.")
		}
		form := url.Values{"repo": {name}}
		switch action {
		case "pull":
			if repo.Source != SourceRemote {
				return web.Skip("Only remote repositories can be pulled.")
			}
			form.Set("action", "pull")
		case "deploy":
			if repo.DeployMode == DeployNone {
				return web.Skip("Deployment is turned off for this repository.")
			}
			form.Set("action", "deploy")
		case "mode":
			form.Set("action", "settings")
			form.Set("deploy_mode", value)
		case "delete":
			form.Set("action", "delete")
		}
		return web.Call(web.BulkCall{Label: name, Method: http.MethodPost, Path: "/git/action.json", Form: form})
	})
}
