package phpapp

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var installPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"manager/php_install.html",
)

// InstallPageData is manager/php_install.html's template context.
type InstallPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Install PHP Application")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InstallPageData{LayoutData: layout, Domains: domains}
	if err := installPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHPAPP - install template render error: %v", err)
	}
}
