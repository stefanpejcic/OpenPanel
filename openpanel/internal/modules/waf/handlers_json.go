package waf

import (
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// handleWAFJSONForDomain returns recent check/block counts for a domain,
// polled by waf.html's status column.
func handleWAFJSONForDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domain := firstPathSegment(r.PathValue("domain"))

	userID, _ := auth.UserID(r)
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	seconds := 60
	if s := r.URL.Query().Get("seconds"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			seconds = v
		}
	}

	stats := readWAFLogs(wafLogPath(domain), seconds)
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "seconds": seconds, "checks": stats.Checks, "blocks": stats.Blocks,
	})
}
