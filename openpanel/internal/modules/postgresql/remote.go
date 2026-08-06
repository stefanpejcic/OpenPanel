package postgresql

import (
	"net/http"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var postgresPortRE = regexp.MustCompile(`^(127\.0\.0\.1:)?(\d+):5432$`)

// handleRemotePostgres toggles PostgreSQL's exposed port between bound to
// 127.0.0.1 (disabled) and 0.0.0.0 (enabled for remote access).
func handleRemotePostgres(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)

	postgresRemotePortOriginal := webserver.GetEnvFileValue(userContext, "POSTGRES_PORT")
	m := postgresPortRE.FindStringSubmatch(postgresRemotePortOriginal)
	if m == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	remotePostgresPortOnly := m[2]

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		ipAddress := reqip.ClientIP(r)
		switch r.Form.Get("action") {
		case "enable":
			docker.SetEnvValue(userContext, "POSTGRES_PORT", remotePostgresPortOnly+":5432")
			_ = logger.RecordUserAction(a.Config, currentUsername, "enabled remote PostgreSQL", ipAddress)
			flashSess(a, w, r, "success", "Remote PostgreSQL access is now enabled.")
		case "disable":
			if strings.Contains(postgresRemotePortOriginal, "127.0.0.1") {
				flashSess(a, w, r, "info", "Remote PostgreSQL access is already disabled.")
			} else {
				docker.SetEnvValue(userContext, "POSTGRES_PORT", "127.0.0.1:"+remotePostgresPortOnly+":5432")
				_ = logger.RecordUserAction(a.Config, currentUsername, "disabled remote PostgreSQL", ipAddress)
				flashSess(a, w, r, "success", "Remote PostgreSQL access is now disabled.")
			}
		}
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")
	}

	postgresRemotePortOriginal = webserver.GetEnvFileValue(userContext, "POSTGRES_PORT")
	display := "ON"
	if strings.Contains(postgresRemotePortOriginal, "127.0.0.1") {
		display = "OFF"
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"remote_postgresql_display": display, "server_ip": serverIP,
			"container_port": remotePostgresPortOnly, "postgres_port": 5432,
		})
		return
	}

	renderRemotePostgresPage(a, w, r, serverIP, remotePostgresPortOnly, display)
}
