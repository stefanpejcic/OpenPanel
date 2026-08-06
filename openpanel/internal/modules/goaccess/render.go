package goaccess

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
	"domains/_shared.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var goaccessPage = loadPage("domains/goaccess.html")

// GoaccessPageData is domains/goaccess.html's template context.
type GoaccessPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderGoaccessSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "GoAccess")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := GoaccessPageData{LayoutData: layout, Domains: domains}
	if err := goaccessPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("GOACCESS - template render error: %v", err)
	}
}
