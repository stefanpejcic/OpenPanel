package processmanager

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var processManagerPage = loadPage("system/process_manager.html")

// ProcessManagerPageData is system/process_manager.html's template context.
type ProcessManagerPageData struct {
	web.LayoutData
	ProcessData []Process
}

func renderProcessManagerPage(a *appctx.App, w http.ResponseWriter, r *http.Request, processes []Process) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Process Manager")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ProcessManagerPageData{LayoutData: layout, ProcessData: processes}
	if err := processManagerPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PROCESSMANAGER - template render error: %v", err)
	}
}
