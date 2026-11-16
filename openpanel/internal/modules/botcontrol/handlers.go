package botcontrol

import (
	"net/http"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func handleBotControlList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainsList, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows := make([]DomainPolicyRow, 0, len(domainsList))
	for _, d := range domainsList {
		rows = append(rows, DomainPolicyRow{
			DomainURL: d.DomainURL,
			Policy:    getDomainPolicy(d.DomainURL),
		})
	}
	renderListPage(a, w, r, rows)
}

func handleBotControlDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domainName := strings.TrimSpace(r.PathValue("domain"))
	if domainName == "" || !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		policy := Policy(strings.TrimSpace(r.Form.Get("policy")))
		if !validPolicy(policy) {
			web.FlashRedirect(a, w, r, "error", "Invalid bot policy.", "/server/bot-control/"+domainName)
			return
		}
		if err := setDomainPolicy(ctx, domainName, policy); err != nil {
			web.FlashRedirect(a, w, r, "error", "Failed to update bot policy: "+err.Error(), "/server/bot-control/"+domainName)
			return
		}
		_ = logger.RecordUserAction(a.Config, username, "set AI bot policy to "+string(policy)+" for "+domainName, reqip.ClientIP(r))
		redirectTo := "/server/bot-control/" + domainName
		if r.Form.Get("return_to") == "list" {
			redirectTo = "/server/bot-control"
		}
		web.FlashRedirect(a, w, r, "success", "AI & bot policy updated for "+domainName+".", redirectTo)
		return
	}

	window := r.URL.Query().Get("window")
	policy := getDomainPolicy(domainName)
	report, err := analyzeDomainTraffic(domainName, window, time.Now())
	if err != nil {
		web.Flash(a, w, r, "error", "Unable to analyze access log: "+err.Error())
	}
	renderDomainPage(a, w, r, domainName, policy, report)
}
