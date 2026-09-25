package mongodb

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ProcessRow is one row of mongodb/processlist.html's table.
type ProcessRow struct {
	OpID           string `json:"opid"`
	User           string `json:"user"`
	Client         string `json:"client"`
	AppName        string `json:"app_name"`
	Op             string `json:"op"`
	NS             string `json:"ns"`
	SecsRunning    string `json:"secs_running"`
	WaitingForLock bool   `json:"waiting_for_lock"`
	Command        string `json:"command"`
	Desc           string `json:"desc"`
	Killable       bool   `json:"killable"`
}

func opKillable(op mongomanager.Op) bool {
	return op.IsClient && op.Active && op.Op != "none"
}

// handleProcessList shows the operations currently running on the user's mongod.
func handleProcessList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var processList []ProcessRow
	ops, opsErr := mongomanager.CurrentOps(ctx, userContext)
	if opsErr != nil {
		flashSess(a, w, r, "error", web.Tr(a, r, "Error fetching process list: %(error)s", "error", opsErr.Error()))
	}
	for _, op := range ops {
		processList = append(processList, ProcessRow{
			OpID: strconv.FormatInt(op.OpID, 10), User: op.User, Client: op.Client, AppName: op.AppName,
			Op: op.Op, NS: op.NS, SecsRunning: strconv.FormatInt(op.SecsRunning, 10), WaitingForLock: op.WaitingForLock,
			Command: op.Command, Desc: op.Desc, Killable: opKillable(op),
		})
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"processlist": processList})
		return
	}

	renderProcessListPage(a, w, r, processList)
}

var (
	errInvalidOpID = errors.New("invalid operation ID")
	errOpNotFound  = errors.New("operation not found or no longer running")
	errOpProtected = errors.New("internal MongoDB operations can not be killed")
)

// killOp runs killOp on opid, but only if it's a running op from a client connection
func killOp(ctx context.Context, userContext, opid string) error {
	id, err := strconv.ParseInt(opid, 10, 64)
	if err != nil || id <= 0 {
		return errInvalidOpID
	}
	ops, err := mongomanager.CurrentOps(ctx, userContext)
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.OpID != id {
			continue
		}
		if !op.IsClient {
			return errOpProtected
		}
		if !opKillable(op) {
			return errOpNotFound
		}
		return mongomanager.KillOp(ctx, userContext, id)
	}
	return errOpNotFound
}

// handleKillQuery kills one running operation from the processlist page
func handleKillQuery(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	opid := r.Form.Get("opid")
	if err := killOp(r.Context(), userContext, opid); err != nil {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error killing operation %(opid)s: %(error)s", "opid", opid, "error", err.Error()), "/mongodb/processlist")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "killed MongoDB operation "+opid, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Operation %(opid)s killed.", "opid", opid), "/mongodb/processlist")
}
