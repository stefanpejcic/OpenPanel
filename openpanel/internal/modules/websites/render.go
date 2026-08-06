package websites

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
	"partials/site_manager_shared.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var sitesPage = loadPage("manager/sites.html")

// SitesPageData is manager/sites.html's template context.
type SitesPageData struct {
	web.LayoutData
	Domains  []appctx.Domain
	Groups   []SiteGroup
	ViewMode string
}

func renderSitesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, groups []SiteGroup, viewMode string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Websites")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := SitesPageData{LayoutData: layout, Domains: domains, Groups: groups, ViewMode: viewMode}
	if err := sitesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - sites template render error: %v", err)
	}
}
