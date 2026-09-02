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
				_ = docker.ToggleProxyHTTPPort(userContext, "on")
				_ = docker.SwapAllWebserversComposePort(userContext, "on")
				// Not docker.ComposeContainer(webserver, "stop") - see
				// ForceRemoveContainer's doc comment for why "podman-compose
				// down" isn't safe to use here (it cascades through
				// depends_on and takes php-fpm down with it).
				docker.ForceRemoveContainer(ctx, userContext, webserver)

				// Checks the actual running state rather than sniffing the
				// activate command's stdout for the word "started" - real
				// podman-compose output doesn't reliably contain it (see the
				// identical fix in internal/modules/php/extensions.go).
				// Polls instead of checking once: varnish's entrypoint copies
				// and rewrites its VCL file before exec'ing varnishd, and on
				// an account that's never run varnish before the image may
				// still need to be pulled - both can take a few seconds past
				// when `podman-compose up` returns, and a single immediate
				// check was misreporting that as a start failure (issue
				// #1091), rolling back and surfacing the compose command's
				// raw stdout (a container ID) as a bogus "error" message.
				result := docker.StartOrStopContainer(ctx, userContext, service, "activate", "run")
				if !result.Success || !docker.WaitForServiceRunning(ctx, userContext, service) {
					_ = docker.SwapAllWebserversComposePort(userContext, "off")
					docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
					_ = docker.ToggleProxyHTTPPort(userContext, "off")
					flashAndRedirect(a, w, r, "error", fmt.Sprintf("Failed to start %s: %s", service, result.Message), "/cache/varnish")
					return
				}

				docker.StartOrStopContainer(ctx, userContext, webserver, "activate", "")
				_ = logger.RecordUserAction(a.Config, currentUsername, "enabled Varnish", ipAddress)
				flashSess(a, w, r, "success", "Varnish caching is now enabled.")
			}

		case "disable":
			if docker.IsServiceRunning(ctx, userContext, service) {
				_ = docker.ToggleProxyHTTPPort(userContext, "off")
				_ = docker.SwapAllWebserversComposePort(userContext, "off")
				// Not docker.ComposeContainer(..., "stop") - see
				// ForceRemoveContainer's doc comment for why "podman-compose
				// down" isn't safe to use here (it cascades through
				// depends_on and takes php-fpm down with it).
				docker.ForceRemoveContainer(ctx, userContext, webserver)
				docker.ForceRemoveContainer(ctx, userContext, service)

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
