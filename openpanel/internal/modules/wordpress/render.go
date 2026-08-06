package wordpress

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

var (
	listPage    = loadPage("manager/wp/list.html")
	installPage = loadPage("manager/wp/install.html")
)

// ListPageData is manager/wp/list.html's template context.
type ListPageData struct {
	web.LayoutData
	Domains  []appctx.Domain
	Sites    []SiteRow
	ViewMode string
}

func renderListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, sites []SiteRow, viewMode string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "WordPress Manager")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ListPageData{LayoutData: layout, Domains: domains, Sites: sites, ViewMode: viewMode}
	if err := listPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WORDPRESS - list template render error: %v", err)
	}
}

// InstallPageData is manager/wp/install.html's template context.
type InstallPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Install WordPress")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InstallPageData{LayoutData: layout, Domains: domains}
	if err := installPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WORDPRESS - install template render error: %v", err)
	}
}
