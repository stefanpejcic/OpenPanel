package ipblocker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the ip-blocker API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "ip_blocker", "GET /api/ip-blocker", func(w http.ResponseWriter, r *http.Request) { apiIPBlockerList(a, w, r) })
	apiregistry.Handle(mux, a, "ip_blocker", "POST /api/ip-blocker", func(w http.ResponseWriter, r *http.Request) { apiIPBlockerAdd(a, w, r) })
	apiregistry.Handle(mux, a, "ip_blocker", "DELETE /api/ip-blocker", func(w http.ResponseWriter, r *http.Request) { apiIPBlockerClear(a, w, r) })
}

// validateIPs mirrors _validate_ips().
func validateIPs(raw []string) (valid, invalid []string) {
	for _, ip := range raw {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if normalized, ok := normalizeIP(ip); ok {
			valid = append(valid, normalized)
		} else {
			invalid = append(invalid, ip)
		}
	}
	return valid, invalid
}

// apiIPBlockerList mirrors api_ip_blocker_list().
func apiIPBlockerList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
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
	if ips == nil {
		ips = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"blocked_ips": ips})
}

// apiIPBlockerAdd mirrors api_ip_blocker_add().
func apiIPBlockerAdd(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		IPs any `json:"ips"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	var raw []string
	switch v := body.IPs.(type) {
	case string:
		raw = strings.Split(v, "\n")
	case []any:
		for _, e := range v {
			if s, ok := e.(string); ok {
				raw = append(raw, s)
			}
		}
	}

	validIPs, invalidIPs := validateIPs(raw)
	if invalidIPs == nil {
		invalidIPs = []string{}
	}

	if len(validIPs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "No valid IPs or CIDRs provided", "invalid": invalidIPs})
		return
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(r.Context(), "opencli", "user-block_ip", username, "--list="+strings.Join(validIPs, " "))
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to block IPs", "details": stderr.String()})
		return
	}

	countStr := strconv.Itoa(len(validIPs))
	_ = logger.RecordUserAction(a.Config, username, "blocked "+countStr+" IP(s) via API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"blocked": validIPs, "invalid": invalidIPs,
		"message": countStr + " IP(s) added to blocklist",
	})
}

// apiIPBlockerClear mirrors api_ip_blocker_clear().
func apiIPBlockerClear(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(r.Context(), "opencli", "user-block_ip", username, "--delete-all")
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to remove IPs", "details": stderr.String()})
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "removed all blocked IPs via API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "All blocked IPs removed"})
}
