package emails

import (
	"encoding/json"
	"io"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterEmailMigrateAPI wires the IMAP migration API routes onto mux, gated behind the same "email_migrate" feature as /emails/migrate
func RegisterEmailMigrateAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "email_migrate", "GET /api/emails/migrate", func(w http.ResponseWriter, r *http.Request) { apiMigrateStatus(a, w, r) })
	apiregistry.Handle(mux, a, "email_migrate", "POST /api/emails/migrate", func(w http.ResponseWriter, r *http.Request) { apiMigrateStart(a, w, r) })
	apiregistry.Handle(mux, a, "email_migrate", "POST /api/emails/migrate/test", func(w http.ResponseWriter, r *http.Request) { apiMigrateTest(a, w, r) })
	apiregistry.Handle(mux, a, "email_migrate", "POST /api/emails/migrate/cancel", func(w http.ResponseWriter, r *http.Request) { apiMigrateCancel(a, w, r) })
	apiregistry.Handle(mux, a, "email_migrate", "GET /api/emails/migrate/log/{id}", func(w http.ResponseWriter, r *http.Request) { apiMigrateLog(a, w, r) })
}

func writeAPIMigrateErr(w http.ResponseWriter, err error) {
	status, msg := errStatus(err)
	writeAPIEmailsJSON(w, status, map[string]string{"error": msg})
}

// decodeMigrateRequest reads a JSON body, exclude_junk defaults to true like the form
func decodeMigrateRequest(w http.ResponseWriter, r *http.Request) (MigrateRequest, bool) {
	req := MigrateRequest{ExcludeJunk: true}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&req); err != nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return req, false
	}
	return req, true
}

func apiMigrateStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	st, err := migrationStatus(r.Context(), userContext)
	if err != nil {
		writeAPIMigrateErr(w, err)
		return
	}
	writeAPIEmailsJSON(w, http.StatusOK, st)
}

func apiMigrateStart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req, ok := decodeMigrateRequest(w, r)
	if !ok {
		return
	}
	job, err := startMigration(r.Context(), a, userID, username, userContext, req)
	if err != nil {
		writeAPIMigrateErr(w, err)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "started email migration from "+job.SourceUser+"@"+job.SourceHost+" to "+job.Destination, reqip.ClientIP(r))
	writeAPIEmailsJSON(w, http.StatusAccepted, job)
}

func apiMigrateTest(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req, ok := decodeMigrateRequest(w, r)
	if !ok {
		return
	}
	res, err := testMigration(r.Context(), a, userID, username, userContext, req)
	if err != nil {
		writeAPIMigrateErr(w, err)
		return
	}
	writeAPIEmailsJSON(w, http.StatusOK, res)
}

func apiMigrateCancel(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := cancelMigration(r.Context(), userContext); err != nil {
		writeAPIMigrateErr(w, err)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "cancelled email migration", reqip.ClientIP(r))
	writeAPIEmailsJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func apiMigrateLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	text, err := readJobLog(userContext, r.PathValue("id"))
	if err != nil {
		writeAPIMigrateErr(w, err)
		return
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]string{"id": r.PathValue("id"), "log": text})
}
