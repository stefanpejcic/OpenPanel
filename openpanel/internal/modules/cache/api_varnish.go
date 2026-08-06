package cache

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterVarnishAPI wires the varnish JSON API routes onto mux.
func RegisterVarnishAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "varnish", "GET /api/cache/varnish", func(w http.ResponseWriter, r *http.Request) { apiVarnishStatus(a, w, r) })
	apiregistry.Handle(mux, a, "varnish", "POST /api/cache/varnish", func(w http.ResponseWriter, r *http.Request) { apiVarnishAction(a, w, r) })
	apiregistry.Handle(mux, a, "varnish", "GET /api/cache/varnish/stats", func(w http.ResponseWriter, r *http.Request) { handleVarnishStats(a, w, r) })
	apiregistry.Handle(mux, a, "varnish", "GET /api/cache/varnish/domains", func(w http.ResponseWriter, r *http.Request) { apiVarnishDomainsList(a, w, r) })
	apiregistry.Handle(mux, a, "varnish", "POST /api/cache/varnish/domains/{domain}", func(w http.ResponseWriter, r *http.Request) { apiVarnishDomainToggle(a, w, r) })
}

// varnishDomainStatuses reports each of the user's domains' varnish
// on/off state, derived from whether its webserver config still proxies
// to the plain HTTPS backend uncommented.
func varnishDomainStatuses(a *appctx.App, r *http.Request, userID int) map[string]string {
	domains, _ := a.AllDomainsForUser(r.Context(), userID)
	result := make(map[string]string, len(domains))
	for _, d := range domains {
		content, readErr := os.ReadFile("/etc/openpanel/caddy/domains/" + d.DomainURL + ".conf")
		if readErr != nil {
			result[d.DomainURL] = "Unknown"
			continue
		}
		hasUncommentedHTTPS := false
		for _, line := range strings.Split(string(content), "\n") {
			stripped := strings.TrimSpace(line)
			if strings.Contains(stripped, "reverse_proxy https://") && !strings.HasPrefix(stripped, "#") {
				hasUncommentedHTTPS = true
				break
			}
		}
		if hasUncommentedHTTPS {
			result[d.DomainURL] = "Off"
		} else {
			result[d.DomainURL] = "On"
		}
	}
	return result
}

// apiVarnishStatus returns varnish's container state, health, available
// actions, and per-domain toggle statuses.
func apiVarnishStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	_, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	status := docker.GetContainerStatus(r.Context(), userContext, "varnish")
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable", "domain"}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "varnish", "container_state": status.State, "health_status": status.Health,
		"actions": actions, "domain_statuses": varnishDomainStatuses(a, r, userID),
	})
}

// apiVarnishAction enables, disables, or restarts varnish, swapping the
// active webserver's proxy port and stopping/starting containers as
// needed around the transition.
func apiVarnishAction(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ctx := r.Context()
	ip := reqip.ClientIP(r)

	var body struct {
		Action string `json:"action"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := strings.ToLower(strings.TrimSpace(body.Action))

	switch action {
	case "restart":
		docker.ComposeContainer(ctx, userContext, "varnish", "restart")
		_ = logger.RecordUserAction(a.Config, currentUsername, "restarted service varnish", ip)
		writeJSON(w, http.StatusOK, map[string]string{"message": "varnish restarted"})
		return

	case "enable":
		webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
		if docker.IsServiceRunning(ctx, userContext, "varnish") {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Varnish is already running"})
			return
		}
		_ = toggleProxyHTTPPort(userContext, "on")
		docker.ComposeContainer(ctx, userContext, webserver, "stop")
		result := docker.StartOrStopContainer(ctx, userContext, "varnish", "activate", "run")
		if !result.Success || !docker.IsServiceRunning(ctx, userContext, "varnish") {
			docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
			_ = toggleProxyHTTPPort(userContext, "off")
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to start varnish: " + result.Message})
			return
		}
		docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
		_ = logger.RecordUserAction(a.Config, currentUsername, "enabled Varnish", ip)
		writeJSON(w, http.StatusOK, map[string]string{"message": "Varnish enabled"})
		return

	case "disable":
		webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
		if !docker.IsServiceRunning(ctx, userContext, "varnish") {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Varnish is not running"})
			return
		}
		_ = toggleProxyHTTPPort(userContext, "off")
		docker.ComposeContainer(ctx, userContext, webserver, "stop")
		docker.ComposeContainer(ctx, userContext, "varnish", "stop")
		result := docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "run")
		_ = logger.RecordUserAction(a.Config, currentUsername, "disabled Varnish", ip)
		if !result.Success {
			writeJSON(w, http.StatusMultiStatus, map[string]string{"warning": "Varnish disabled but failed to restart " + webserver + ": " + result.Message})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Varnish disabled"})
		return

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action. Valid: enable, disable, restart"})
	}
}

// apiVarnishDomainsList returns every domain's varnish on/off status.
func apiVarnishDomainsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	writeJSON(w, http.StatusOK, map[string]any{"domain_statuses": varnishDomainStatuses(a, r, userID)})
}

// apiVarnishDomainToggle turns varnish caching on or off for one domain.
func apiVarnishDomainToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	currentUsername, _, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	status := strings.TrimSpace(body.Status)
	if status != "On" && status != "Off" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be On or Off"})
		return
	}

	what, verb := "off", "disabled"
	if status == "On" {
		what, verb = "on", "enabled"
	}
	_ = exec.CommandContext(ctx, "opencli", "domains-varnish", domain, what).Run()
	_ = logger.RecordUserAction(a.Config, currentUsername, verb+" Varnish for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"domain": domain, "varnish": status})
}
