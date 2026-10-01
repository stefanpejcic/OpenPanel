package domains

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// same check opencli domains-cloudflare status does
var cloudflareOnlyRE = regexp.MustCompile(`^import[ \t]+cloudflare-only([ \t]|$)`)

// handleDomainCloudflare restricts a domain to Cloudflare IPs or lifts that, via opencli domains-cloudflare
func handleDomainCloudflare(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_ = r.ParseForm()
	domainName := r.Form.Get("domain_name")

	if domainName == "" || strings.HasSuffix(domainName, ".onion") {
		flashAndRedirect(a, w, r, "error", "Invalid request. Domain name must be provided.", "/domains")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	action, done, logMsg := "disable", "Cloudflare restriction removed for %(domain)s.", "removed Cloudflare-only restriction for "
	if r.Form.Get("cloudflare_action") == "enable" {
		action, done, logMsg = "enable", "%(domain)s is now reachable only through Cloudflare.", "restricted to Cloudflare-only "
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-cloudflare", action, domainName).CombinedOutput()
	invalidateRewriteCondCache(ctx, a, domainName)
	if cmdErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = cmdErr.Error()
		}
		flashAndRedirect(a, w, r, "error", msg, "/domains")
		return
	}

	currentUsername, _, _ := injected(a, ctx, userID)
	_ = logger.RecordUserAction(a.Config, currentUsername, logMsg+domainName, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, done, "domain", domainName), "/domains")
}
