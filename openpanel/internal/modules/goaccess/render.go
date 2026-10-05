package goaccess

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func loadPage(files ...string) *web.Page {
	return web.LoadPage(append([]string{"domains/_shared.html"}, files...)...)
}

var goaccessPage = loadPage("domains/goaccess.html")

// GoaccessPageData is domains/goaccess.html's template context.
type GoaccessPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderGoaccessSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Visitor Statistics")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := GoaccessPageData{LayoutData: layout, Domains: domains}
	if err := goaccessPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("GOACCESS - template render error: %v", err)
	}
}
