package ipblocker

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

var ipBlockerPage = loadPage("system/ip_blocker.html")

// IPBlockerPageData is system/ip_blocker.html's template context.
type IPBlockerPageData struct {
	web.LayoutData
	IPs []string
}

func renderIPBlockerPage(a *appctx.App, w http.ResponseWriter, r *http.Request, ips []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "IP Blocker")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := IPBlockerPageData{LayoutData: layout, IPs: ips}
	if err := ipBlockerPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("IPBLOCKER - template render error: %v", err)
	}
}
