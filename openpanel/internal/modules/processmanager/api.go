package processmanager

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the process manager JSON API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "process_manager", "GET /api/process-manager", func(w http.ResponseWriter, r *http.Request) { apiProcessList(a, w, r) })
	apiregistry.Handle(mux, a, "process_manager", "DELETE /api/process-manager/{pid}", func(w http.ResponseWriter, r *http.Request) { apiProcessKill(a, w, r) })
}

// apiProcessList returns every process running across the user's
// containers.
func apiProcessList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	processes, procErr := getPodmanProcesses(r.Context(), userContext)
	if procErr != nil {
		writeJSONError(w, http.StatusInternalServerError, procErr.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"processes": processes, "count": len(processes)})
}

// apiProcessKill terminates a process by PID, using the container-scoped
// `podman exec <container> kill -9 <pid>` also used by the UI route
// (handleProcessManager) - a bare host-level `kill -9 <pid>` wouldn't hit
// the intended process, since these are container-namespaced PIDs.
func apiProcessKill(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pidStr := r.PathValue("pid")
	pidInt, atoiErr := strconv.Atoi(pidStr)
	if atoiErr != nil || pidInt <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid PID")
		return
	}

	processes, procErr := getPodmanProcesses(ctx, userContext)
	if procErr != nil {
		writeJSONError(w, http.StatusInternalServerError, procErr.Error())
		return
	}

	var container string
	found := false
	for _, p := range processes {
		if p.PID == pidStr {
			container = p.Container
			found = true
			break
		}
	}
	if !found {
		writeJSONError(w, http.StatusForbidden, "PID not found or not allowed")
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "kill", "-9", strconv.Itoa(pidInt))
	cmd := podmanmanager.Command(ctx, userContext, argv)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		writeJSONError(w, http.StatusInternalServerError, stderr.String())
		return
	}

	_ = logger.RecordUserAction(a.Config, username, fmt.Sprintf("terminated process %d via API", pidInt), reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Process %d killed successfully", pidInt)})
}
