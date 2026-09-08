// Package apiregistry tracks every /api/* route as it's registered, so /api/endpoints can list them - http.ServeMux has no route-iteration API of its own.
package apiregistry

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Endpoint is one entry of the /api/endpoints introspection response.
type Endpoint struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

var (
	mu        sync.Mutex
	endpoints []Endpoint
)

// Add records one route. pattern is the net/http mux syntax ("GET /api/mysql/databases" or just the path for all methods); defaults to GET if no verb is given.
func Add(pattern string) {
	method := ""
	path := pattern
	if i := strings.IndexByte(pattern, ' '); i != -1 {
		method, path = pattern[:i], pattern[i+1:]
	}
	methods := []string{"GET"}
	if method != "" {
		methods = []string{method}
	}

	mu.Lock()
	defer mu.Unlock()
	for i, e := range endpoints {
		if e.Path == path {
			endpoints[i].Methods = append(endpoints[i].Methods, methods...)
			return
		}
	}
	endpoints = append(endpoints, Endpoint{Path: path, Methods: methods})
}

// All returns every recorded endpoint, sorted by path.
func All() []Endpoint {
	mu.Lock()
	out := make([]Endpoint, len(endpoints))
	copy(out, endpoints)
	mu.Unlock()

	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Handle registers pattern on mux behind auth.RequireAPI(a, featureName) and records it for /api/endpoints in one call
func Handle(mux *http.ServeMux, a *appctx.App, featureName, pattern string, h http.HandlerFunc) {
	Add(pattern)
	mux.Handle(pattern, auth.RequireAPI(a, featureName)(h))
}
