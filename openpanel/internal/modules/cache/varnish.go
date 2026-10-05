package cache

import (
	"context"
	"encoding/json"
	"errors"
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
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterVarnish wires the varnish page and stats routes onto mux.
func RegisterVarnish(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "varnish")(h)
	}
	mux.Handle("/cache/varnish", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleVarnish(a, w, r) }))
	mux.Handle("GET /cache/varnish/stats", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleVarnishStats(a, w, r) }))
}

// getVarnishStats caches parsed varnishstat output for 5s, keyed per userContext - a single shared cache key would let two different hosting accounts' varnish containers serve each other's cached stats
func getVarnishStats(ctx context.Context, a *appctx.App, userContext string) map[string]float64 {
	stats, _ := cache.Memoize(ctx, a.Cache, "varnish_stats:"+userContext, 5*time.Second, func() (map[string]float64, error) {
		stats := computeVarnishStats(ctx, userContext)
		// don't cache a failed read, right after a restart it would show zeros
		if len(stats) == 0 {
			return stats, errors.New("no varnish stats")
		}
		return stats, nil
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

// handleVarnishStats returns computed cache-hit ratio, traffic, backend health, and memory metrics derived from varnishstat
func handleVarnishStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !docker.IsServiceRunning(ctx, userContext, "varnish") {
		web.WriteJSON(w, http.StatusOK, map[string]any{"status": "stopped", "message": "Varnish is not running"})
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

	web.WriteJSON(w, http.StatusOK, map[string]any{
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

// StartVarnish moves the webserver behind varnish and starts it, rolling back on failure - webserverFailed tells which step broke
func StartVarnish(opCtx context.Context, userContext string) (webserverFailed bool, errMsg string) {
	const service = "varnish"
	webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	_ = docker.ToggleProxyHTTPPort(userContext, "on")
	_ = docker.SwapAllWebserversComposePort(userContext, "on")
	// not docker.ComposeContainer(webserver, "stop") - see ForceRemoveContainer's doc comment for why "podman-compose down" isn't safe here (it cascades through depends_on and takes php-fpm down with it)
	docker.ForceRemoveContainer(opCtx, userContext, webserver)

	// the webserver has to come up BEFORE varnish, not after: varnish's VCL resolves the webserver's container-network hostname at startup, and if that name isn't registered yet the VCL compile fails and it exits ("Backend host could not be resolved", exit code 2). It recovers via restart:unless-stopped once the webserver is up, but a crash-looping container responds slowly to a later `podman rm -f` (~10s stall), long enough to blow past the request and abandon the webserver recreation mid-flight.
	if !restartWebserverAfterVarnishToggle(opCtx, userContext, webserver).Success {
		_ = docker.SwapAllWebserversComposePort(userContext, "off")
		restartWebserverAfterVarnishToggle(opCtx, userContext, webserver)
		_ = docker.ToggleProxyHTTPPort(userContext, "off")
		return true, "could not bring " + webserver + " back up with the Varnish proxy port"
	}

	// checks the actual running state rather than sniffing the activate command's stdout for "started" - real podman-compose output doesn't reliably contain it (see the identical fix in php/extensions.go). Polls instead of checking once: varnish's entrypoint rewrites its VCL before exec'ing varnishd, and the image may still need pulling on a first run, both can take a few seconds longer than `podman-compose up` takes to return - a single immediate check misreported that as a start failure (issue #1091).
	result := docker.StartOrStopContainer(opCtx, userContext, service, "activate", "run")
	if !result.Success || !docker.WaitForServiceRunning(opCtx, userContext, service) {
		_ = docker.SwapAllWebserversComposePort(userContext, "off")
		restartWebserverAfterVarnishToggle(opCtx, userContext, webserver)
		_ = docker.ToggleProxyHTTPPort(userContext, "off")
		if result.Message == "" {
			result.Message = "container did not reach a running state"
		}
		return false, result.Message
	}
	return false, ""
}

// DomainVarnishStatus reads the domain's caddy file: "On" when traffic goes through varnish, "Off" when straight to the webserver
func DomainVarnishStatus(domain string) string {
	content, err := os.ReadFile("/etc/openpanel/caddy/domains/" + domain + ".conf")
	if err != nil {
		return "Unknown"
	}
	for _, line := range strings.Split(string(content), "\n") {
		stripped := strings.TrimSpace(line)
		if strings.Contains(stripped, "reverse_proxy https://") && !strings.HasPrefix(stripped, "#") {
			return "Off"
		}
	}
	return "On"
}

// VarnishHitStats returns the account-wide varnish hit ratio plus raw hit/miss/object counters
func VarnishHitStats(ctx context.Context, a *appctx.App, userContext string) (hitRatio, hits, misses, objects float64) {
	flat := getVarnishStats(ctx, a, userContext)
	hits, misses = flat["MAIN.cache_hit"], flat["MAIN.cache_miss"]
	total := hits + misses + flat["MAIN.cache_hitpass"] + flat["MAIN.cache_hitmiss"]
	if total > 0 {
		hitRatio = round4(hits / total)
	}
	return hitRatio, hits, misses, flat["MAIN.n_object"]
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

// handleVarnish serves the varnish page and handles its enable/disable/per-domain-toggle form actions
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
		outputJSON := r.URL.Query().Get("output") == "json"

		// enable/disable run several sequential podman operations that can outlast an impatient client, and r.Context() cancels the moment the client gives up - which would SIGKILL whatever podman command is running via exec.CommandContext, abandoning the operation mid-way (this is exactly how a webserver was once left permanently missing after a Varnish disable that reported success). A background context with its own generous timeout keeps this running to completion regardless of the client.
		opCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		switch action {
		case "enable":
			if !docker.IsServiceRunning(ctx, userContext, service) {
				webserverFailed, errMsg := StartVarnish(opCtx, userContext)
				if errMsg != "" {
					msg := web.Tr(a, r, "Failed to start %(service)s: %(message)s", "service", service, "message", errMsg)
					if webserverFailed {
						msg = web.Tr(a, r, "Failed to start %(webserver)s: could not bring it back up with the Varnish proxy port", "webserver", webserver)
					}
					if outputJSON {
						web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
						return
					}
					web.FlashRedirect(a, w, r, "error", msg, "/cache/varnish")
					return
				}

				_ = logger.RecordUserAction(a.Config, currentUsername, "enabled Varnish", ipAddress)
				web.Flash(a, w, r, "success", "Varnish caching is now enabled.")
			}

		case "disable":
			if docker.IsServiceRunning(ctx, userContext, service) {
				_ = docker.ToggleProxyHTTPPort(userContext, "off")
				_ = docker.SwapAllWebserversComposePort(userContext, "off")
				// not docker.ComposeContainer(..., "stop") - see ForceRemoveContainer's doc comment for why "podman-compose down" isn't safe here (it cascades through depends_on and takes php-fpm down with it)
				docker.ForceRemoveContainer(opCtx, userContext, webserver)
				docker.ForceRemoveContainer(opCtx, userContext, service)

				result := restartWebserverAfterVarnishToggle(opCtx, userContext, webserver)
				_ = logger.RecordUserAction(a.Config, currentUsername, "disabled Varnish", ipAddress)

				if !result.Success {
					msg := web.Tr(a, r, "Failed to start %(webserver)s after disabling %(service)s: %(message)s", "webserver", webserver, "service", service, "message", result.Message)
					if outputJSON {
						web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
						return
					}
					web.FlashRedirect(a, w, r, "error", msg, "/cache/varnish")
					return
				}
				web.Flash(a, w, r, "success", "Varnish caching is now disabled.")
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
			web.Flash(a, w, r, "success", web.Tr(a, r, "Varnish cache is now %(new_status)s for domain %(domain_name)s", "new_status", newStatus, "domain_name", domainName))
		}
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	varnishStatus := make(map[string]string, len(domains))
	for _, d := range domains {
		varnishStatus[d.DomainURL] = DomainVarnishStatus(d.DomainURL)
	}

	status := docker.GetContainerStatus(ctx, userContext, service)
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable", "domain"}
	}

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"status": varnishStatus, "varnish_status": varnishStatus, "actions": actions,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderVarnishPage(a, w, r, status, varnishStatus, domains)
}
