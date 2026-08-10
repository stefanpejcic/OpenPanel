package processmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

type killRequest struct {
	PIDToKill string `json:"pid_to_kill"`
	Container string `json:"container"`
}

// handleProcessManager serves the process manager page. GET lists (or,
// with ?output=json, dumps) every process running across the user's
// containers; POST terminates one.
func handleProcessManager(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	processes, procErr := getPodmanProcesses(r.Context(), userContext)
	valid := make(map[string]bool, len(processes))
	for _, p := range processes {
		valid[processKey(p.Container, p.PID)] = true
	}
	if procErr != nil {
		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, "error", procErr.Error())
		_ = a.Sessions.Save(r, w, sess)
		processes = nil
	}

	if r.Method == http.MethodPost {
		var payload killRequest
		_ = json.NewDecoder(r.Body).Decode(&payload)

		if payload.PIDToKill == "" || payload.Container == "" {
			writeJSON(w, http.StatusOK, map[string]any{"success": false, "error_message": "No PID provided"})
			return
		}

		if !valid[processKey(payload.Container, payload.PIDToKill)] {
			writeJSON(w, http.StatusOK, map[string]any{"success": false, "error_message": "PID not found or not allowed"})
			return
		}

		pidInt, atoiErr := strconv.Atoi(payload.PIDToKill)
		if atoiErr != nil || pidInt <= 0 {
			writeJSON(w, http.StatusOK, map[string]any{"success": false, "error_message": "Invalid PID"})
			return
		}

		argv := podmanmanager.PodmanArgv(userContext, "exec", "--user", "root", payload.Container, "kill", "-9", strconv.Itoa(pidInt))
		cmd := podmanmanager.Command(r.Context(), userContext, argv)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if runErr := cmd.Run(); runErr != nil {
			writeJSON(w, http.StatusOK, map[string]any{"success": false, "error_message": stderr.String()})
			return
		}

		_ = logger.RecordUserAction(a.Config, username,
			fmt.Sprintf("terminated process %d in container %s using Process Manager", pidInt, payload.Container),
			reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Process killed successfully"})
		return
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, processes)
		return
	}

	renderProcessManagerPage(a, w, r, filterDisplayable(processes))
}

// filterDisplayable drops entrypoint/healthcheck noise rows before
// rendering the table.
func filterDisplayable(processes []Process) []Process {
	var out []Process
	for _, p := range processes {
		if isDisplayableCmd(p.CMD) {
			out = append(out, p)
		}
	}
	return out
}
