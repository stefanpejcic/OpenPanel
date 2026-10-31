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
	Domains  []appctx.Domain
	Selected string
	Summary  *StatsSummary
	// set when the selected domain has no report yet, or it couldn't be read
	NoReport bool
}

func renderGoaccessSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, selected string, summary *StatsSummary) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Visitor Statistics")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := GoaccessPageData{LayoutData: layout, Domains: domains, Selected: selected, Summary: summary, NoReport: selected != "" && summary == nil}
	if err := goaccessPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("GOACCESS - template render error: %v", err)
	}
}
