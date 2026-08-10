package emails

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterEmailExportAPI wires the email export API route onto mux, gated
// behind the same "email_export" feature as the web UI's /emails/export.
func RegisterEmailExportAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "email_export", "GET /api/emails/export", func(w http.ResponseWriter, r *http.Request) { apiEmailExport(a, w, r) })
}

// apiEmailExport streams the current user's mailboxes as a CSV download
// (email, password, quota columns; password is always blank), mirroring
// the web UI's /emails/export.
func apiEmailExport(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	_ = logger.RecordUserAction(a.Config, currentUsername, "exported mailbox list ("+strconv.Itoa(len(rows))+" addresses) via API", reqip.ClientIP(r))
}
