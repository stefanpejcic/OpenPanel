package autoinstaller

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

var autoinstallerPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "manager/autoinstaller.html")...)

// AutoinstallerPageData is manager/autoinstaller.html's template context.
type AutoinstallerPageData struct {
	web.LayoutData
	Domains []appctx.Domain
	Counts  map[string]int

	// UpsellAllowed/UpsellPlanName/UpsellURL back the disabled-tile hover
	// tooltip: when a module key is in UpsellAllowed, the tile offers an
	// upgrade CTA naming UpsellPlanName/UpsellURL instead of the plain
	// "contact your administrator" message.
	UpsellAllowed  map[string]bool
	UpsellPlanName string
	UpsellURL      string
}

func renderAutoinstallerPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, counts map[string]int, upsell upsellData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Auto Installer")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AutoinstallerPageData{
		LayoutData: layout, Domains: domains, Counts: counts,
		UpsellAllowed: upsell.Allowed, UpsellPlanName: upsell.PlanName, UpsellURL: upsell.URL,
	}
	if err := autoinstallerPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("AUTOINSTALLER - template render error: %v", err)
	}
}
