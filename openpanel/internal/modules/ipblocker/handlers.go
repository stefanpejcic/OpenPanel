package ipblocker

import (
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// handleIPBlocker mirrors ip_blocker.py's ip_blocker().
func handleIPBlocker(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		var validIPs []string
		for _, line := range strings.Split(r.Form.Get("ips"), "\n") {
			ip := strings.TrimSpace(line)
			if ip == "" {
				continue
			}
			if normalized, ok := normalizeIP(ip); ok {
				validIPs = append(validIPs, normalized)
			}
		}

		argv := []string{"user-block_ip", username}
		var logAction, message string
		if len(validIPs) > 0 {
			logAction = "blocked IP addresses using IP Blocker"
			message = "IP addresses have been successfully added to blocklist and can no longer access websites"
			argv = append(argv, "--list="+strings.Join(validIPs, " "))
		} else {
			logAction = "removed all blocked IPs using IP Blocker"
			message = "All IP addresses have been successfully removed from blocklist and can now access websites"
			argv = append(argv, "--delete-all")
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		cmd := exec.CommandContext(r.Context(), "opencli", argv...)
		if runErr := cmd.Run(); runErr == nil {
			_ = logger.RecordUserAction(a.Config, username, logAction, reqip.ClientIP(r))
			flash.Add(sess, "success", message)
		} else {
			flash.Add(sess, "error", "Failed to block IP addresses - please try again")
		}
		_ = a.Sessions.Save(r, w, sess)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(r.Context(), "opencli", "user-block_ip", username)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve IPs: "+stderr.String())
		return
	}

	var ips []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			ips = append(ips, line)
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, ips)
		return
	}

	renderIPBlockerPage(a, w, r, ips)
}
