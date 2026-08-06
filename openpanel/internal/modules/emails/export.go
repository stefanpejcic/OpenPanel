package emails

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleEmailExport streams the current user's mailboxes as a CSV
// download (email, password, quota columns; password is always blank).
func handleEmailExport(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)
	currentEmailsList := GetEmailList(ctx, a, userID, currentUsername, userDomains)

	type row struct{ email, quota string }
	var rows []row
	for _, line := range currentEmailsList {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) >= 6 {
			rows = append(rows, row{email: parts[1], quota: parts[5]})
		} else if len(parts) >= 2 {
			rows = append(rows, row{email: parts[1], quota: ""})
		}
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="mailboxes-`+currentUsername+`.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"email", "password", "quota"})
	for _, rrow := range rows {
		_ = cw.Write([]string{rrow.email, "", rrow.quota})
	}
	cw.Flush()

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "exported mailbox list ("+strconv.Itoa(len(rows))+" addresses)", ipAddress)
}
