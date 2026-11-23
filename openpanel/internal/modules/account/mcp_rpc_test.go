package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
)

func newTestMCPServer(dispatcher http.Handler, search bool, features ...string) *mcpServer {
	set := map[string]bool{}
	for _, f := range features {
		set[f] = true
	}
	return &mcpServer{
		dispatcher: dispatcher,
		limiter:    &mcpRateLimiter{limit: 1000, win: map[string]*window{}},
		search:     search,
		version:    func(*http.Request) string { return "9.9.9" },
		features:   func(*http.Request) map[string]bool { return set },
		scope:      func(*http.Request) mcptokens.Scope { return mcptokens.Scope{Mode: mcptokens.ModeFull} },
	}
}

func withScope(s *mcpServer, scope mcptokens.Scope) *mcpServer {
	s.scope = func(*http.Request) mcptokens.Scope { return scope }
	return s
}

// mcpDo posts one json-rpc message and returns the status and the decoded "result" (or the whole reply on error)
func mcpDo(t *testing.T, s *mcpServer, method string, params map[string]any) (int, map[string]any) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	rec := httptest.NewRecorder()
	s.handle(rec, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(string(body))))
	var reply map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &reply)
	if result, ok := reply["result"].(map[string]any); ok {
		return rec.Code, result
	}
	return rec.Code, reply
}

func listedNames(result map[string]any) map[string]bool {
	names := map[string]bool{}
	tools, _ := result["tools"].([]any)
	for _, raw := range tools {
		if tool, ok := raw.(map[string]any); ok {
			names[fmt.Sprint(tool["name"])] = true
		}
	}
	return names
}

// toolText decodes the json a tool result carries in its text block
func toolText(t *testing.T, result map[string]any) map[string]any {
	t.Helper()
	content, _ := result["content"].([]any)
	if len(content) == 0 {
		t.Fatalf("no content in %v", result)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(fmt.Sprint(content[0].(map[string]any)["text"])), &out); err != nil {
		t.Fatalf("tool text is not json: %v", err)
	}
	return out
}

func TestMCPInitializeNegotiatesVersion(t *testing.T) {
	s := newTestMCPServer(nil, true)
	_, result := mcpDo(t, s, "initialize", map[string]any{"protocolVersion": "2025-03-26"})
	if result["protocolVersion"] != "2025-03-26" {
		t.Errorf("known version not echoed: %v", result["protocolVersion"])
	}
	_, result = mcpDo(t, s, "initialize", map[string]any{"protocolVersion": "2099-01-01"})
	if result["protocolVersion"] != mcpProtocolVersions[0] {
		t.Errorf("unknown version should fall back to %s, got %v", mcpProtocolVersions[0], result["protocolVersion"])
	}
	if result["instructions"] == "" || result["instructions"] == nil {
		t.Error("initialize has no instructions")
	}
	if info, _ := result["serverInfo"].(map[string]any); info["version"] != "9.9.9" {
		t.Errorf("serverInfo version = %v", info["version"])
	}
}

func TestMCPPingAndNotifications(t *testing.T) {
	s := newTestMCPServer(nil, true)
	if code, reply := mcpDo(t, s, "ping", nil); code != http.StatusOK || reply["error"] != nil {
		t.Errorf("ping: %d %v", code, reply)
	}
	if code, _ := mcpDo(t, s, "notifications/progress", nil); code != http.StatusAccepted {
		t.Errorf("notification answered %d, want 202", code)
	}
}

func TestMCPListSearchModeShowsOnlyAllowedDirectTools(t *testing.T) {
	s := newTestMCPServer(nil, true, "mysql")
	_, result := mcpDo(t, s, "tools/list", nil)
	names := listedNames(result)
	for _, want := range []string{mcpSearchToolName, mcpCallToolName, "mysql_databases_get"} {
		if !names[want] {
			t.Errorf("%s missing from the list", want)
		}
	}
	if names["domains_get"] {
		t.Error("domains_get listed without the domains feature")
	}
	if len(names) > len(mcpDirectTools)+2 {
		t.Errorf("search mode listed %d tools", len(names))
	}
}

func TestMCPDirectToolsExist(t *testing.T) {
	registry := getMCPToolRegistry()
	for _, name := range mcpDirectTools {
		tool, ok := registry[name]
		if !ok {
			t.Errorf("direct tool %s is not in the api docs", name)
			continue
		}
		if tool.Method != "GET" && !strings.HasPrefix(name, "deploy") && name != "pm2_domain_restart" {
			t.Errorf("direct tool %s is %s, only the deploy tools may change things from the short list", name, tool.Method)
		}
	}
}

func TestMCPListFullModeFiltersByFeature(t *testing.T) {
	s := newTestMCPServer(nil, false, "mysql")
	_, result := mcpDo(t, s, "tools/list", nil)
	names := listedNames(result)
	if names[mcpSearchToolName] || !names["mysql_databases_post"] || names["domains_post"] {
		t.Errorf("full list wrong: search=%v mysql=%v domains=%v", names[mcpSearchToolName], names["mysql_databases_post"], names["domains_post"])
	}
}

func TestMCPToolAllowedFor(t *testing.T) {
	registry := getMCPToolRegistry()
	write, read := registry["mysql_databases_post"], registry["mysql_databases_get"]
	if write == nil || read == nil {
		t.Fatal("mysql tools missing")
	}
	mysql := map[string]bool{"mysql": true}
	ro := mcptokens.Scope{Mode: mcptokens.ModeReadOnly}
	full := mcptokens.Scope{Mode: mcptokens.ModeFull}
	if write.allowedFor(mcpAccess{mysql, ro}) || !read.allowedFor(mcpAccess{mysql, ro}) {
		t.Error("read-only token should keep GET tools and drop the rest")
	}
	if !write.allowedFor(mcpAccess{nil, full}) {
		t.Error("unknown features should not hide tools")
	}
	if read.allowedFor(mcpAccess{map[string]bool{"domains": true}, full}) {
		t.Error("tool allowed without its feature")
	}
}

func TestMCPAnnotations(t *testing.T) {
	registry := getMCPToolRegistry()
	get, _ := registry["mysql_databases_get"].Schema["annotations"].(map[string]any)
	if get["readOnlyHint"] != true {
		t.Errorf("GET tool annotations = %v", get)
	}
	for _, name := range mcpToolOrder {
		tool := registry[name]
		ann, _ := tool.Schema["annotations"].(map[string]any)
		if tool.Method == "DELETE" && ann["destructiveHint"] != true {
			t.Errorf("%s is a DELETE without destructiveHint", name)
		}
		if strings.Contains(strings.ToLower(fmt.Sprint(tool.Schema["description"])), "multipart/form-data") {
			t.Errorf("%s needs a file upload and should not be a tool", name)
		}
	}
}

func TestMCPSearchTools(t *testing.T) {
	s := newTestMCPServer(nil, true, "mysql", "domains")
	_, result := mcpDo(t, s, "tools/call", map[string]any{"name": mcpSearchToolName, "arguments": map[string]any{"query": "create mysql database", "limit": 5}})
	found := toolText(t, result)
	tools, _ := found["tools"].([]any)
	if len(tools) == 0 || len(tools) > 5 {
		t.Fatalf("got %d tools", len(tools))
	}
	names := listedNames(found)
	if !names["mysql_databases_post"] {
		t.Errorf("mysql_databases_post not in the results: %v", names)
	}
	for name := range names {
		if !strings.HasPrefix(name, "mysql") && !strings.HasPrefix(name, "domains") {
			t.Errorf("%s returned without its feature", name)
		}
	}

	_, result = mcpDo(t, s, "tools/call", map[string]any{"name": mcpSearchToolName, "arguments": map[string]any{"query": "zzzzqq"}})
	found = toolText(t, result)
	if areas, _ := found["areas"].([]any); len(areas) == 0 {
		t.Errorf("no areas offered on an empty result: %v", found)
	}
}

func TestMCPCallToolSendsFormAndReadsStream(t *testing.T) {
	var gotDomain, gotType string
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/nodejs/install", func(w http.ResponseWriter, r *http.Request) {
		gotDomain, gotType = r.FormValue("domain_id"), r.Header.Get("Content-Type")
		_, _ = w.Write([]byte("{\"status\":\"Pulling image\"}\n{\"error\":\"port is taken\"}\n"))
	})
	s := newTestMCPServer(mux, true, "nodejs")
	_, result := mcpDo(t, s, "tools/call", map[string]any{"name": mcpCallToolName, "arguments": map[string]any{
		"name":      "nodejs_install",
		"arguments": map[string]any{"body": map[string]any{"domain_id": 42}},
	}})
	if gotDomain != "42" || gotType != "application/x-www-form-urlencoded" {
		t.Errorf("handler saw domain_id=%q content-type=%q", gotDomain, gotType)
	}
	if result["isError"] != true {
		t.Error("a stream ending in an error line should be an error result")
	}
	out := toolText(t, result)
	last, _ := out["result"].(map[string]any)
	if last["error"] != "port is taken" {
		t.Errorf("last event = %v", out["result"])
	}
	if progress, _ := out["progress"].([]any); len(progress) != 1 {
		t.Errorf("progress = %v", out["progress"])
	}
}

func TestMCPCallStreamSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/wordpress/install", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = fmt.Fprintf(w, "{\"status\":\"Downloading\"}\n{\"status\":\"done\",\"domain_id\":%q}\n", fmt.Sprint(body["domain_id"]))
	})
	s := newTestMCPServer(mux, true, "wordpress")
	_, result := mcpDo(t, s, "tools/call", map[string]any{"name": "wordpress_install", "arguments": map[string]any{"body": map[string]any{"domain_id": "7"}}})
	if result["isError"] != false {
		t.Errorf("isError = %v", result["isError"])
	}
	last, _ := toolText(t, result)["result"].(map[string]any)
	if last["domain_id"] != "7" {
		t.Errorf("json body did not reach the handler: %v", last)
	}
}

func TestMCPCallBlockedWithoutFeature(t *testing.T) {
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/domains", func(http.ResponseWriter, *http.Request) { called = true })
	s := newTestMCPServer(mux, true, "mysql")
	_, result := mcpDo(t, s, "tools/call", map[string]any{"name": "domains_get"})
	if called || result["isError"] != true {
		t.Errorf("called=%v isError=%v", called, result["isError"])
	}
	if msg := fmt.Sprint(toolText(t, result)["error"]); !strings.Contains(msg, "domains") {
		t.Errorf("error doesn't name the feature: %s", msg)
	}
}

func TestMCPToolNamesUnchangedByUploadFilter(t *testing.T) {
	registry := getMCPToolRegistry()
	// /api/emails/import is an upload, its sibling tools must keep the names they had before it was hidden
	for _, name := range []string{"mysql_databases_get", "mysql_databases_post", "sites", "domains_get"} {
		if registry[name] == nil {
			t.Errorf("%s is gone", name)
		}
	}
}

func TestMCPDescribeToolUsesTypedSchemas(t *testing.T) {
	tool := mcpDescribeTool("things_post", "things", docEndpoint{
		Method: "POST", Path: "/api/things", Description: "Create a thing.",
		Body:        json.RawMessage(`{"name":"a"}`),
		BodySchema:  json.RawMessage(`{"type":"object","x-content-type":"application/x-www-form-urlencoded","properties":{"name":{"type":"string","description":"Thing name."}},"required":["name"]}`),
		QuerySchema: json.RawMessage(`{"type":"object","properties":{"dry_run":{"type":"boolean","description":"Only check."}}}`),
	})
	if !tool.Form {
		t.Error("x-content-type form should make the call form-encoded")
	}
	input := tool.Schema["inputSchema"].(map[string]any)
	props := input["properties"].(map[string]any)
	body := props["body"].(map[string]any)
	if _, leaked := body["x-content-type"]; leaked {
		t.Error("x-content-type leaked into the tool schema")
	}
	if name, _ := body["properties"].(map[string]any)["name"].(map[string]any); name["type"] != "string" {
		t.Errorf("typed body lost its properties: %v", body)
	}
	if req := input["required"].([]string); len(req) != 1 || req[0] != "body" {
		t.Errorf("required = %v, want [body]", req)
	}
	if q := props["query"].(map[string]any); q["properties"] == nil {
		t.Errorf("typed query missing: %v", q)
	}

	none := mcpDescribeTool("things_start", "things", docEndpoint{
		Method: "POST", Path: "/api/things/start", Description: "Start.",
		BodySchema: json.RawMessage(`{"type":"object","properties":{}}`),
	})
	if _, ok := none.Schema["inputSchema"].(map[string]any)["properties"].(map[string]any)["body"]; ok {
		t.Error("a route documented as taking no body still asks for one")
	}

	untyped := mcpDescribeTool("old_post", "things", docEndpoint{Method: "POST", Path: "/api/old", Description: "Old."})
	if b := untyped.Schema["inputSchema"].(map[string]any)["properties"].(map[string]any)["body"].(map[string]any); b["additionalProperties"] != true {
		t.Errorf("untyped route should keep the free-form body: %v", b)
	}
}

func TestMCPRealTypedTools(t *testing.T) {
	registry := getMCPToolRegistry()
	create := registry["mysql_databases_post"]
	if create == nil {
		t.Fatal("mysql_databases_post missing")
	}
	input := create.Schema["inputSchema"].(map[string]any)
	body, _ := input["properties"].(map[string]any)["body"].(map[string]any)
	name, _ := body["properties"].(map[string]any)["name"].(map[string]any)
	if name["type"] != "string" || name["pattern"] == nil {
		t.Errorf("mysql_databases_post body isn't typed: %v", body)
	}
	if req := input["required"].([]string); len(req) != 1 || req[0] != "body" {
		t.Errorf("required = %v", req)
	}

	// revoking a grant reads a json body, so the DELETE tool has to offer one
	for _, n := range mcpToolOrder {
		tool := registry[n]
		if tool.Method == "DELETE" && tool.Path == "/api/mysql/grants" {
			if _, ok := tool.Schema["inputSchema"].(map[string]any)["properties"].(map[string]any)["body"]; !ok {
				t.Error("DELETE /api/mysql/grants tool has no body")
			}
		}
	}
}

func TestMCPHidesCredentialRoutes(t *testing.T) {
	for _, tool := range getMCPToolRegistry() {
		if strings.HasSuffix(tool.Path, "/wp-info") || tool.Path == "/api/phpmyadmin" {
			t.Errorf("%s hands back credentials and should not be a tool", tool.Name)
		}
	}
}

func TestMCPSearchPrefersTheNamedGroup(t *testing.T) {
	found := mcpSearchTools("wordpress plugins update", 5, mcpAccess{scope: mcptokens.Scope{Mode: mcptokens.ModeFull}})
	tools, _ := found["tools"].([]any)
	for _, raw := range tools {
		name := fmt.Sprint(raw.(map[string]any)["name"])
		if tool := getMCPToolRegistry()[name]; tool != nil && tool.Group != "WordPress" && !strings.Contains(tool.Path, "wp") {
			t.Errorf("%s (%s) ranked in the top 5 for a WordPress search", name, tool.Group)
		}
	}
	if names := listedNames(found); !names["wp_cli_action"] {
		t.Errorf("wp_cli_action not found: %v", names)
	}
}
