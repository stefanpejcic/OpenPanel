package websites

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func loadPage(files ...string) *web.Page {
	return web.LoadPage(append([]string{"partials/site_manager_shared.html"}, files...)...)
}

var sitesPage = loadPage("manager/sites.html")

// SitesPageData is manager/sites.html's template context.
type SitesPageData struct {
	web.LayoutData
	Domains  []appctx.Domain
	Groups   []SiteGroup
	ViewMode string
	// without a key the speed column links to where the key is set instead of asking PageSpeed
	HasPagespeedKey bool
}

func renderSitesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, groups []SiteGroup, viewMode string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Websites")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := SitesPageData{LayoutData: layout, Domains: domains, Groups: groups, ViewMode: viewMode}
	if _, _, userContext, err := auth.Injected(a, r); err == nil {
		data.HasPagespeedKey = getPagespeedInsightsAPIKey(a, r, userContext) != ""
	}
	if err := sitesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - sites template render error: %v", err)
	}
}
