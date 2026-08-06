package webserverconf

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

var webserverConfPage = loadPage("system/apache_nginx_conf_editor.html")

// WebserverConfPageData is system/apache_nginx_conf_editor.html's
// template context.
type WebserverConfPageData struct {
	web.LayoutData
	Service           string
	Path              string
	CanRestoreDefault bool
	FileContent       string
}

func renderWebserverConfPage(a *appctx.App, w http.ResponseWriter, r *http.Request, pageTitle, service, path string, canRestoreDefault bool, fileContent string) {
	layout, _, err := web.BuildLayoutData(a, w, r, pageTitle)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WebserverConfPageData{
		LayoutData: layout, Service: service, Path: path,
		CanRestoreDefault: canRestoreDefault, FileContent: fileContent,
	}
	if err := webserverConfPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSERVERCONF - template render error: %v", err)
	}
}
