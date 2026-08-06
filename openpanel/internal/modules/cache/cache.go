// Package cache implements five near-identical single-container cache
// services (redis, memcached, elasticsearch, opensearch, valkey) that just
// enable/disable/restart a fixed compose service, plus varnish's richer
// per-domain reverse-cache toggle and live varnishstat metrics.
package cache

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/services"
)

// serviceDef is one of the five generic cache services' fixed identity:
// its service name, port, title, and description.
type serviceDef struct {
	Name        string
	Port        int
	Title       string
	Description string
}

var (
	redisDef = serviceDef{
		Name: "redis", Port: 6379, Title: "Redis",
		Description: "Redis (Remote Dictionary Server) is an in-memory, key/value store that is used primarily as an application cache or quick-response database.",
	}
	memcachedDef = serviceDef{
		Name: "memcached", Port: 11211, Title: "Memcached",
		Description: "A high-performance, distributed in-memory key-value caching system designed to accelerate dynamic web applications.",
	}
	elasticsearchDef = serviceDef{
		Name: "elasticsearch", Port: 9200, Title: "ElasticSearch",
		Description: "The leading distributed, RESTful, open source search and analytics engine designed for speed and reliability.",
	}
	opensearchDef = serviceDef{
		Name: "opensearch", Port: 9200, Title: "OpenSearch",
		Description: "OpenSearch is an open-source, distributed search and analytics suite for large volumes of data.",
	}
	valkeyDef = serviceDef{
		Name: "valkey", Port: 6379, Title: "Valkey",
		Description: "Valkey is an in-memory key/value store (a community-driven fork of Redis) used primarily as an application cache or fast, low-latency database.",
	}
)

func registerGenericService(mux *http.ServeMux, a *appctx.App, def serviceDef) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, def.Name)(h)
	}
	mux.Handle("/cache/"+def.Name, requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleGenericService(a, w, r, def)
	}))
}

// RegisterRedis, RegisterMemcached, RegisterElasticsearch, RegisterOpensearch
// and RegisterValkey each wire up one generic cache service's route. Kept
// as separate Registrar entries (rather than one bundled registrar, unlike
// e.g. the domains/php phases' satellite-module bundling) since each has
// its own real enabled_modules entry admins toggle independently.
func RegisterRedis(mux *http.ServeMux, a *appctx.App) { registerGenericService(mux, a, redisDef) }
func RegisterMemcached(mux *http.ServeMux, a *appctx.App) {
	registerGenericService(mux, a, memcachedDef)
}
func RegisterElasticsearch(mux *http.ServeMux, a *appctx.App) {
	registerGenericService(mux, a, elasticsearchDef)
}
func RegisterOpensearch(mux *http.ServeMux, a *appctx.App) {
	registerGenericService(mux, a, opensearchDef)
}
func RegisterValkey(mux *http.ServeMux, a *appctx.App) { registerGenericService(mux, a, valkeyDef) }

// handleGenericService implements the shared enable/disable/restart page
// for all five generic cache services - identical apart from the def.
func handleGenericService(a *appctx.App, w http.ResponseWriter, r *http.Request, def serviceDef) {
	ctx := r.Context()
	username, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		services.HandleServiceAction(a, r, r.Form.Get("action"), def.Name, userContext, username)
	}

	status := docker.GetContainerStatus(ctx, userContext, def.Name)
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable"}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"service": def.Name, "description": def.Description, "port": def.Port, "actions": actions,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderCacheServicePage(a, w, r, def, status)
}
