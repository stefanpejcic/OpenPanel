package botcontrol

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	botControlListPage   = web.LoadPage("system/bot_control.html")
	botControlDomainPage = web.LoadPage("system/bot_control_domain.html")
)

// DomainPolicyRow pairs a domain name with its active bot-control policy.
type DomainPolicyRow struct {
	DomainURL string
	Policy    Policy
}

// ListPageData is system/bot_control.html's template context.
type ListPageData struct {
	web.LayoutData
	Rows          []DomainPolicyRow
	KnownCrawlers []BotSignature
}

// DomainPageData is system/bot_control_domain.html's template context.
type DomainPageData struct {
	web.LayoutData
	Domain        string
	Policy        Policy
	Report        TrafficReport
	KnownCrawlers []BotSignature
}

func renderListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rows []DomainPolicyRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "AI & Bot Control")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ListPageData{
		LayoutData:    layout,
		Rows:          rows,
		KnownCrawlers: KnownAICrawlers(),
	}
	if err := botControlListPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BOTCONTROL - template render error: %v", err)
	}
}

func renderDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName string, policy Policy, report TrafficReport) {
	layout, _, err := web.BuildLayoutData(a, w, r, "AI & Bot Control — "+domainName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainPageData{
		LayoutData:    layout,
		Domain:        domainName,
		Policy:        policy,
		Report:        report,
		KnownCrawlers: KnownAICrawlers(),
	}
	if err := botControlDomainPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BOTCONTROL - domain template render error: %v", err)
	}
}
