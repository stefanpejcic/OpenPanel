package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterVarnish wires the varnish page and stats routes onto mux.
func RegisterVarnish(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "varnish")(h)
	}
	mux.Handle("/cache/varnish", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleVarnish(a, w, r) }))
	mux.Handle("GET /cache/varnish/stats", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleVarnishStats(a, w, r) }))
}

// getVarnishStats caches parsed varnishstat output for 5s, keyed per
// userContext since this panel runs many independent per-user containers
// on one host - a single shared cache key would let two different hosting
// accounts' varnish containers serve each other's cached stats.
func getVarnishStats(ctx context.Context, a *appctx.App, userContext string) map[string]float64 {
	stats, _ := cache.Memoize(ctx, a.Cache, "varnish_stats:"+userContext, 5*time.Second, func() (map[string]float64, error) {
		return computeVarnishStats(ctx, userContext), nil
	})
	return stats
}

func computeVarnishStats(ctx context.Context, userContext string) map[string]float64 {
	argv := podmanmanager.PodmanArgv(userContext, "exec", "varnish", "varnishstat", "-1", "-j")
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return map[string]float64{}
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(out, &raw); err != nil {
		return map[string]float64{}
	}

	flat := map[string]float64{}
	for k, v := range raw {
		var entry struct {
			Value *float64 `json:"value"`
		}
		if json.Unmarshal(v, &entry) == nil && entry.Value != nil {
			flat[k] = *entry.Value
		}
	}
	return flat
}

// handleVarnishStats returns computed cache-hit ratio, traffic, backend
// health, and memory metrics derived from varnishstat.
func handleVarnishStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !docker.IsServiceRunning(ctx, userContext, "varnish") {
		writeJSON(w, http.StatusOK, map[string]any{"status": "stopped", "message": "Varnish is not running"})
		return
	}

	flat := getVarnishStats(ctx, a, userContext)
	hits, misses := flat["MAIN.cache_hit"], flat["MAIN.cache_miss"]
	pass, hitmiss := flat["MAIN.cache_hitpass"], flat["MAIN.cache_hitmiss"]
	total := hits + misses + pass + hitmiss
	hitRatio := 0.0
	if total > 0 {
		hitRatio = round4(hits / total)
	}

	backendFail, backendReq := flat["MAIN.backend_fail"], flat["MAIN.backend_req"]
	backendHealth := "healthy"
	if backendFail >= 10 {
		backendHealth = "degraded"
	}

	errorRate := 0.0
	if backendReq != 0 {
		errorRate = backendFail / backendReq
	}
	efficiencyScore := math.Max(0, math.Round((hitRatio*100)-(errorRate*50)))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "running",
		"cache":  map[string]any{"hit_ratio": hitRatio, "hits": hits, "misses": misses, "pass": pass},
		"traffic": map[string]any{
			"requests_total": flat["MAIN.client_req"], "connections": flat["MAIN.client_conn"],
		},
		"backend": map[string]any{
			"requests": backendReq, "failures": backendFail, "retries": flat["MAIN.backend_retry"], "health": backendHealth,
		},
		"memory": map[string]any{
			"objects": flat["MAIN.n_object"], "evictions": flat["MAIN.n_lru_nuked"],
		},
		"performance": map[string]any{"efficiency_score": efficiencyScore, "error_rate": round4(errorRate)},
	})
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

// toggleProxyHTTPPort comments or uncomments the PROXY_HTTP_PORT line in a
// user's .env file, enabling or disabling the port varnish listens on.
func toggleProxyHTTPPort(userContext, state string) error {
	path := "/home/" + userContext + "/.env"
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(rewriteProxyHTTPPortLines(string(content), state)), 0o644)
}

// rewriteProxyHTTPPortLines is toggleProxyHTTPPort()'s pure line-rewriting
// half, split out for testability without touching the filesystem.
func rewriteProxyHTTPPortLines(content, state string) string {
	lines := strings.SplitAfter(content, "\n")
	var out strings.Builder
	for _, line := range lines {
		if !strings.Contains(line, "PROXY_HTTP_PORT=") {
			out.WriteString(line)
			continue
		}
		switch state {
		case "on":
			out.WriteString(strings.TrimLeft(line, "# "))
		case "off":
			if !strings.HasPrefix(strings.TrimSpace(line), "#") {
				out.WriteString("#" + line)
			} else {
				out.WriteString(line)
			}
		default:
			out.WriteString(line)
		}
	}
	return out.String()
}

// proxyPortSwapPair picks the old/new port-variable pair to swap when
// toggling varnish on or off.
func proxyPortSwapPair(state string) (old, replacement string, err error) {
	switch state {
	case "on":
		return "${HTTP_PORT}", "${PROXY_HTTP_PORT}", nil
	case "off":
		return "${PROXY_HTTP_PORT}", "${HTTP_PORT}", nil
	default:
		return "", "", fmt.Errorf("state must be either 'on' or 'off'")
	}
}

// swapWebserverComposePort swaps the active webserver's port-mapping
// variable in place. podman-compose can't resolve a ${VAR} nested inside
// another ${VAR:-default}, so docker-compose.yml keeps the active
// webserver's port mapping flat rather than using a fallback expression -
// the swap is scoped to the block between the webserver's own
// `container_name:` line and its `HTTPS_PORT` line, so only that
// service's port mapping is touched.
func swapWebserverComposePort(userContext, webserver, state string) error {
	old, replacement, err := proxyPortSwapPair(state)
	if err != nil {
		return err
	}

	path := "/home/" + userContext + "/docker-compose.yml"
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(rewriteComposePortBlock(string(content), webserver, old, replacement)), 0o644)
}

// rewriteComposePortBlock is swapWebserverComposePort()'s pure block-scoped
// rewrite, split out for testability without touching the filesystem.
func rewriteComposePortBlock(content, webserver, old, replacement string) string {
	lines := strings.SplitAfter(content, "\n")

	inBlock := false
	for i, line := range lines {
		if strings.HasSuffix(strings.TrimRight(line, " \t\r\n"), "container_name: "+webserver) {
			inBlock = true
		}
		if inBlock && strings.Contains(line, old) {
			lines[i] = strings.ReplaceAll(line, old, replacement)
		}
		if inBlock && strings.Contains(line, "HTTPS_PORT") {
			inBlock = false
		}
	}
	return strings.Join(lines, "")
}

// handleVarnish serves the varnish page and handles its enable/disable/
// per-domain-toggle form actions.
func handleVarnish(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	const service = "varnish"

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		action := r.Form.Get("action")
		webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
		ipAddress := reqip.ClientIP(r)

		switch action {
		case "enable":
			if !docker.IsServiceRunning(ctx, userContext, service) {
				_ = toggleProxyHTTPPort(userContext, "on")
				_ = swapWebserverComposePort(userContext, webserver, "on")
				docker.ComposeContainer(ctx, userContext, webserver, "stop")

				// Checks the actual running state rather than sniffing the
				// activate command's stdout for the word "started" - real
				// podman-compose output doesn't reliably contain it (see the
				// identical fix in internal/modules/php/extensions.go).
				result := docker.StartOrStopContainer(ctx, userContext, service, "activate", "run")
				if !result.Success || !docker.IsServiceRunning(ctx, userContext, service) {
					_ = swapWebserverComposePort(userContext, webserver, "off")
					docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
					_ = toggleProxyHTTPPort(userContext, "off")
					flashAndRedirect(a, w, r, "error", fmt.Sprintf("Failed to start %s: %s", service, result.Message), "/cache/varnish")
					return
				}

				docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
				_ = logger.RecordUserAction(a.Config, currentUsername, "enabled Varnish", ipAddress)
				flashSess(a, w, r, "success", "Varnish caching is now enabled.")
			}

		case "disable":
			if docker.IsServiceRunning(ctx, userContext, service) {
				_ = toggleProxyHTTPPort(userContext, "off")
				_ = swapWebserverComposePort(userContext, webserver, "off")
				docker.ComposeContainer(ctx, userContext, webserver, "stop")
				docker.ComposeContainer(ctx, userContext, service, "stop")

				result := docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "run")
				_ = logger.RecordUserAction(a.Config, currentUsername, "disabled Varnish", ipAddress)

				if !result.Success {
					flashAndRedirect(a, w, r, "error", fmt.Sprintf("Failed to start %s after disabling %s: %s", webserver, service, result.Message), "/cache/varnish")
					return
				}
				flashSess(a, w, r, "success", "Varnish caching is now disabled.")
			}

		case "domain":
			domainName := r.Form.Get("domain_name")
			if idx := strings.Index(domainName, "/"); idx != -1 {
				domainName = domainName[:idx]
			}
			newStatus := r.Form.Get("varnish_action")

			if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
				http.Error(w, "You do not own this domain.", http.StatusForbidden)
				return
			}

			what, verb := "off", "disabled"
			if newStatus == "On" {
				what, verb = "on", "enabled"
			}
			_ = logger.RecordUserAction(a.Config, currentUsername, verb+" Varnish caching for domain "+domainName, ipAddress)
			_ = exec.CommandContext(ctx, "opencli", "domains-varnish", domainName, what).Run()
			flashSess(a, w, r, "success", fmt.Sprintf("Varnish cache is now %s for domain %s", newStatus, domainName))
		}
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	varnishStatus := make(map[string]string, len(domains))
	for _, d := range domains {
		content, readErr := os.ReadFile("/etc/openpanel/caddy/domains/" + d.DomainURL + ".conf")
		if readErr != nil {
			varnishStatus[d.DomainURL] = "Unknown"
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
			varnishStatus[d.DomainURL] = "Off"
		} else {
			varnishStatus[d.DomainURL] = "On"
		}
	}

	status := docker.GetContainerStatus(ctx, userContext, service)
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable", "domain"}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": varnishStatus, "varnish_status": varnishStatus, "actions": actions,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderVarnishPage(a, w, r, status, varnishStatus, domains)
}
