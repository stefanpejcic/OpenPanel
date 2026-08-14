package drupal

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
	listPage    = loadPage("manager/drupal_list.html")
	installPage = loadPage("manager/drupal_install.html")
)

// ListPageData is manager/drupal_list.html's template context.
type ListPageData struct {
	web.LayoutData
	Sites []SiteRow
}

func renderListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, sites []SiteRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Drupal Manager")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ListPageData{LayoutData: layout, Sites: sites}
	if err := listPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DRUPAL - list template render error: %v", err)
	}
}

// InstallPageData is manager/drupal_install.html's template context.
type InstallPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Install Drupal")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InstallPageData{LayoutData: layout, Domains: domains}
	if err := installPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DRUPAL - install template render error: %v", err)
	}
}
