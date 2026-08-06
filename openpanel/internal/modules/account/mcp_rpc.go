package account

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apidocs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// This file implements the /mcp JSON-RPC endpoint: a minimal MCP
// Streamable HTTP transport (initialize, tools/list, tools/call). One tool
// is generated per documented /api/ route+method (from apidocs.EndpointsJSON,
// the same data /account/api's docs page renders), and calling a tool just
// replays the equivalent HTTP request against the real route by dispatching
// it straight through the app's own http.ServeMux, forwarding the caller's
// own Authorization header - so this file never duplicates the business
// logic already living in each module's api.go.

const mcpProtocolVersion = "2025-06-18"

// mcpExcludedToolPaths are routes that don't make sense as agent-facing
// tools (auth/introspection).
var mcpExcludedToolPaths = map[string]bool{
	"/api/login":     true,
	"/api/endpoints": true,
}

var mcpPathParamRE = regexp.MustCompile(`<(?:[^:<>]+:)?([^<>]+)>`)

type docEndpoint struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Description string          `json:"description"`
	Body        json.RawMessage `json:"body,omitempty"`
}

type docGroup struct {
	Group     string        `json:"group"`
	Feature   featureField  `json:"feature"`
	Endpoints []docEndpoint `json:"endpoints"`
}

// featureField handles endpoint doc groups whose "feature" is usually a
// single string but is a list for the one group gating on either of two
// features (Node.js / Python Apps -> ["nodejs", "python"]).
type featureField []string

func (f *featureField) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*f = []string{single}
		return nil
	}
	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return err
	}
	*f = multi
	return nil
}

func (f featureField) String() string {
	return strings.Join([]string(f), "' or '")
}

type mcpPathParam struct {
	Token    string
	Name     string
	JSONType string
}

type mcpTool struct {
	Name       string
	Method     string
	Path       string
	PathParams []mcpPathParam
	Schema     map[string]any
}

var (
	mcpRegistryOnce sync.Once
	mcpRegistry     map[string]*mcpTool
	mcpToolOrder    []string
)

// converterJSONType maps a path-param converter name to a JSON Schema type.
func converterJSONType(converterName string) string {
	switch converterName {
	case "int":
		return "integer"
	case "float":
		return "number"
	default:
		return "string"
	}
}

// mcpDescribeTool builds the MCP tool schema for one documented API endpoint.
func mcpDescribeTool(name, method, path, feature, description string, body json.RawMessage) *mcpTool {
	var pathParams []mcpPathParam
	for _, m := range mcpPathParamRE.FindAllStringSubmatch(path, -1) {
		raw, paramName := m[0], m[1]
		converter := "default"
		if idx := strings.IndexByte(raw, ':'); idx != -1 {
			converter = raw[1:idx]
		}
		pathParams = append(pathParams, mcpPathParam{Token: raw, Name: paramName, JSONType: converterJSONType(converter)})
	}

	if description == "" {
		description = method + " " + path
	}
	fullDescription := description + " (requires the '" + feature + "' feature to be enabled on the account)"

	properties := map[string]any{}
	var required []string
	for _, p := range pathParams {
		properties[p.Name] = map[string]any{
			"type":        p.JSONType,
			"description": "Path parameter '" + p.Name + "'",
		}
		required = append(required, p.Name)
	}
	if required == nil {
		required = []string{}
	}

	switch method {
	case "POST", "PUT", "PATCH":
		bodyDescription := "JSON request body."
		if len(body) > 0 {
			bodyDescription += " Example: " + string(body)
		}
		properties["body"] = map[string]any{
			"type":                 "object",
			"description":          bodyDescription,
			"additionalProperties": true,
		}
	case "GET", "DELETE":
		properties["query"] = map[string]any{
			"type":                 "object",
			"description":          "Optional query string parameters as key/value pairs.",
			"additionalProperties": true,
		}
	}

	return &mcpTool{
		Name: name, Method: method, Path: path, PathParams: pathParams,
		Schema: map[string]any{
			"name":        name,
			"description": fullDescription,
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}
}

// mcpPathSlug derives a stable, readable, unique-per-path tool-name base
// from a route path.
func mcpPathSlug(path string) string {
	slug := mcpPathParamRE.ReplaceAllString(path, "$1")
	slug = strings.TrimPrefix(slug, "/api/")
	slug = strings.NewReplacer("/", "_", "-", "_").Replace(slug)
	return strings.Trim(slug, "_")
}

// getMCPToolRegistry builds the tool registry once from apidocs.EndpointsJSON
// (the same data source /account/api's docs page renders), skipping
// excluded paths, and disambiguating tool names by appending the lowercased
// method only when the same path is documented under more than one HTTP method.
func getMCPToolRegistry() map[string]*mcpTool {
	mcpRegistryOnce.Do(func() {
		var groups []docGroup
		if err := json.Unmarshal(apidocs.EndpointsJSON, &groups); err != nil {
			mcpRegistry = map[string]*mcpTool{}
			return
		}

		type rawEntry struct {
			method, path, feature, description string
			body                               json.RawMessage
		}
		var entries []rawEntry
		slugMethodCount := map[string]int{}
		for _, g := range groups {
			for _, ep := range g.Endpoints {
				if mcpExcludedToolPaths[ep.Path] {
					continue
				}
				entries = append(entries, rawEntry{ep.Method, ep.Path, g.Feature.String(), ep.Description, ep.Body})
				slugMethodCount[mcpPathSlug(ep.Path)]++
			}
		}

		mcpRegistry = make(map[string]*mcpTool, len(entries))
		for _, e := range entries {
			slug := mcpPathSlug(e.path)
			name := slug
			if slugMethodCount[slug] > 1 {
				name = slug + "_" + strings.ToLower(e.method)
			}
			tool := mcpDescribeTool(name, e.method, e.path, e.feature, e.description, e.body)
			mcpRegistry[name] = tool
			mcpToolOrder = append(mcpToolOrder, name)
		}
	})
	return mcpRegistry
}

// mcpCallTool substitutes path params, then dispatches the equivalent HTTP
// request straight through dispatcher (mux wrapped in auth.LoadUser, so the
// replayed request's Authorization header actually gets resolved into a
// user identity - raw mux.ServeHTTP skips that middleware entirely) rather
// than over the network, forwarding the caller's Authorization header.
func mcpCallTool(dispatcher http.Handler, tool *mcpTool, arguments map[string]any, authHeader string) (any, int) {
	body, hasBody := arguments["body"]
	query, hasQuery := arguments["query"]

	actualPath := tool.Path
	for _, p := range tool.PathParams {
		v, ok := arguments[p.Name]
		if !ok {
			return map[string]string{"error": "Missing required path parameter '" + p.Name + "'"}, 400
		}
		actualPath = strings.Replace(actualPath, p.Token, (&url.URL{Path: toStringArg(v)}).EscapedPath(), 1)
	}

	if hasQuery {
		if qMap, ok := query.(map[string]any); ok && len(qMap) > 0 {
			values := url.Values{}
			for k, v := range qMap {
				values.Set(k, toStringArg(v))
			}
			actualPath += "?" + values.Encode()
		}
	}

	var reqBody io.Reader
	if hasBody && body != nil {
		encoded, _ := json.Marshal(body)
		reqBody = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(tool.Method, actualPath, reqBody)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", authHeader)

	rec := httptest.NewRecorder()
	dispatcher.ServeHTTP(rec, req)

	var payload any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		payload = map[string]string{"raw": rec.Body.String()}
	}
	return payload, rec.Code
}

func toStringArg(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		encoded, _ := json.Marshal(t)
		return string(encoded)
	}
}

// mcpRateLimiter is a fixed-window limit keyed per-account (mcp:<user_id>)
// rather than per-IP, since one /mcp call can fan out into any of the
// underlying /api/ actions - a shared IP shouldn't throttle everyone on it,
// and a token used from a new IP shouldn't get a fresh allowance.
type mcpRateLimiter struct {
	limit int
	mu    sync.Mutex
	win   map[string]*window
}

func newMCPRateLimiter(a *appctx.App) *mcpRateLimiter {
	limit := atoiDefault(a.Config.Get("mcp_ratelimit", ""), 60)
	if limit <= 0 {
		limit = 60
	}
	return &mcpRateLimiter{limit: limit, win: map[string]*window{}}
}

func (l *mcpRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w, ok := l.win[key]
	if !ok || now.Sub(w.start) >= time.Minute {
		w = &window{start: now}
		l.win[key] = w
	}
	w.count++
	return w.count <= l.limit
}

func mcpRateLimitKey(r *http.Request) string {
	if userID, ok := auth.UserID(r); ok && userID != 0 {
		return "mcp:" + strconv.Itoa(userID)
	}
	return reqip.ClientIP(r)
}

type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

func jsonRPCResult(id, result any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
}

func jsonRPCError(id any, code int, message string) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}}
}

// handleMCPEndpoint dispatches one JSON-RPC request per MCP's Streamable HTTP transport.
func handleMCPEndpoint(dispatcher http.Handler, limiter *mcpRateLimiter, a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !limiter.allow(mcpRateLimitKey(r)) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate_limit_exceeded", "message": "Too many requests. Please try again later."})
		return
	}

	var payload jsonRPCRequest
	raw, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(raw, &payload); err != nil || strings.TrimSpace(string(raw)) == "" || bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(jsonRPCError(nil, -32600, "Invalid Request: expected a single JSON-RPC object, batch requests are not supported"))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch payload.Method {
	case "initialize":
		result := map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "openpanel", "version": "1.0"},
		}
		_ = json.NewEncoder(w).Encode(jsonRPCResult(payload.ID, result))

	case "notifications/initialized", "notifications/cancelled":
		w.WriteHeader(http.StatusAccepted)

	case "tools/list":
		readOnly := auth.MCPReadOnly(r)
		registry := getMCPToolRegistry()
		tools := make([]any, 0, len(registry))
		for _, name := range mcpToolOrder {
			t := registry[name]
			if readOnly && t.Method != "GET" && t.Method != "HEAD" {
				continue
			}
			tools = append(tools, t.Schema)
		}
		_ = json.NewEncoder(w).Encode(jsonRPCResult(payload.ID, map[string]any{"tools": tools}))

	case "tools/call":
		name, _ := payload.Params["name"].(string)
		tool, ok := getMCPToolRegistry()[name]
		if !ok {
			_ = json.NewEncoder(w).Encode(jsonRPCError(payload.ID, -32602, "Unknown tool '"+name+"'"))
			return
		}
		arguments, _ := payload.Params["arguments"].(map[string]any)
		responseBody, status := mcpCallTool(dispatcher, tool, arguments, r.Header.Get("Authorization"))
		text, _ := json.Marshal(responseBody)
		result := map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(text)}},
			"isError": status >= 400,
		}
		_ = json.NewEncoder(w).Encode(jsonRPCResult(payload.ID, result))

	default:
		_ = json.NewEncoder(w).Encode(jsonRPCError(payload.ID, -32601, "Unknown method '"+payload.Method+"'"))
	}
}

// RegisterMCPEndpoint wires the POST /mcp route onto mux, separate from
// RegisterMCP (the /account/mcp token-management pages) so the two can be
// reasoned about independently.
//
// tools/call replays go through mux wrapped in auth.LoadUser (not raw mux):
// the live server only runs LoadUser as part of the outer middleware chain
// built in cmd/openpanel/main.go, which a direct mux.ServeHTTP bypasses
// entirely - without it, the replayed request's Authorization header would
// never get resolved into a user identity and every tool call would 401.
func RegisterMCPEndpoint(mux *http.ServeMux, a *appctx.App) {
	limiter := newMCPRateLimiter(a)
	dispatcher := auth.LoadUser(a)(mux)
	mux.Handle("POST /mcp", auth.RequireAPI(a, "mcp")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMCPEndpoint(dispatcher, limiter, a, w, r)
	})))
}
