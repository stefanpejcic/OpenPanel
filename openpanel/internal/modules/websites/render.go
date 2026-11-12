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
	// summary tiles above the list
	SiteCount      int
	ContainerCount int
	WebsitesLimit  int
	WebsitesPct    int
	FreeDomains    []string
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
	fillSitesSummary(a, r, &data)
	if _, _, userContext, err := auth.Injected(a, r); err == nil {
		data.HasPagespeedKey = getPagespeedInsightsAPIKey(a, r, userContext) != ""
	}
	if err := sitesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - sites template render error: %v", err)
	}
}

// fillSitesSummary counts sites and containers, finds the domains with no site and reads the plan's website limit
func fillSitesSummary(a *appctx.App, r *http.Request, data *SitesPageData) {
	used := map[int]bool{}
	for _, g := range data.Groups {
		for _, s := range g.Sites {
			data.SiteCount++
			used[s.DomainID] = true
			if s.Container != "" {
				data.ContainerCount++
			}
		}
	}
	for _, d := range data.Domains {
		if !used[d.DomainID] {
			data.FreeDomains = append(data.FreeDomains, d.DomainURL)
		}
	}
	if a == nil || a.DB == nil {
		return
	}
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return
	}
	planID, _ := injected["hosting_plan"].(int)
	if plan, err := a.QueryPlanDetailsByID(r.Context(), planID); err == nil {
		data.WebsitesLimit = web.AtoiDefault(plan.WebsitesLimit, 0)
	}
	if data.WebsitesLimit > 0 {
		data.WebsitesPct = min(100, data.SiteCount*100/data.WebsitesLimit)
	}
}
