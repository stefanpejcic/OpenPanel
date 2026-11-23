package account

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
)

func TestMCPModifyTokenHidesDestructiveTools(t *testing.T) {
	s := withScope(newTestMCPServer(nil, false, "mysql"), mcptokens.Scope{Mode: mcptokens.ModeModify})
	_, result := mcpDo(t, s, "tools/list", nil)
	names := listedNames(result)
	if !names["mysql_databases_post"] {
		t.Error("modify token lost a plain create tool")
	}
	for name := range names {
		if tool := getMCPToolRegistry()[name]; tool != nil && tool.Method == "DELETE" {
			t.Errorf("modify token lists %s", name)
		}
	}
}

func TestMCPModifyTokenRefusesDelete(t *testing.T) {
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/mysql/databases/{db_name}", func(http.ResponseWriter, *http.Request) { called = true })
	s := withScope(newTestMCPServer(mux, true, "mysql"), mcptokens.Scope{Mode: mcptokens.ModeModify})
	var deleteTool string
	for _, name := range mcpToolOrder {
		if tool := getMCPToolRegistry()[name]; tool.Method == "DELETE" && strings.HasPrefix(tool.Path, "/api/mysql/databases/") {
			deleteTool = name
			break
		}
	}
	if deleteTool == "" {
		t.Fatal("no mysql delete tool in the docs")
	}
	_, result := mcpDo(t, s, "tools/call", map[string]any{"name": mcpCallToolName, "arguments": map[string]any{"name": deleteTool, "arguments": map[string]any{"db_name": "shop"}}})
	if called || result["isError"] != true {
		t.Fatalf("called=%v result=%v", called, result)
	}
	if msg := toolText(t, result)["error"]; !strings.Contains(msg.(string), "delete") {
		t.Errorf("error = %v", msg)
	}
}

func TestMCPFeatureLimitedToken(t *testing.T) {
	s := withScope(newTestMCPServer(nil, true, "mysql", "domains", "dns"), mcptokens.Scope{Mode: mcptokens.ModeFull, Features: []string{"dns"}})
	_, result := mcpDo(t, s, "tools/list", nil)
	names := listedNames(result)
	if names["mysql_databases_get"] || names["domains_get"] || !names["dns_domain"] {
		t.Errorf("limited list wrong: %v", names)
	}
	_, result = mcpDo(t, s, "tools/call", map[string]any{"name": mcpSearchToolName, "arguments": map[string]any{"query": "mysql database"}})
	for name := range listedNames(toolText(t, result)) {
		if !strings.HasPrefix(name, "dns") {
			t.Errorf("search returned %s to a dns-only token", name)
		}
	}
	_, result = mcpDo(t, s, "tools/call", map[string]any{"name": "domains_get"})
	if result["isError"] != true || !strings.Contains(toolText(t, result)["error"].(string), "isn't allowed") {
		t.Errorf("domains_get on a dns-only token: %v", result)
	}
}

func TestMCPFeatureOptions(t *testing.T) {
	all := mcpFeatureOptions(nil)
	seen := map[string]bool{}
	for _, o := range all {
		if o.Feature == "api" || o.Feature == "mcp" {
			t.Errorf("%s offered as a limit", o.Feature)
		}
		if seen[o.Feature] {
			t.Errorf("%s listed twice", o.Feature)
		}
		seen[o.Feature] = true
	}
	if !seen["dns"] || !seen["nodejs"] {
		t.Fatalf("expected dns and nodejs in %v", all)
	}
	only := mcpFeatureOptions(map[string]bool{"dns": true})
	if len(only) != 1 || only[0].Feature != "dns" || only[0].Label != "DNS Zone Editor" {
		t.Errorf("filtered options = %v", only)
	}
	if bad := unknownMCPFeatures([]string{"dns", "nope"}); len(bad) != 1 || bad[0] != "nope" {
		t.Errorf("unknown = %v", bad)
	}
}

func TestRenderMCPPageScopes(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := MCPPageData{
		LayoutData:     baseLayout(mgr, "/account/mcp"),
		MCPURL:         "https://panel.example.com/mcp",
		FeatureOptions: mcpFeatureOptions(map[string]bool{"dns": true, "domains": true}),
		FeatureLabels:  mcpFeatureLabels(),
		Tokens: []mcptokens.Token{
			{ID: 1, Name: "Deployer", TokenPrefix: "op_mcp_ab", CreatedAt: "2026-11-23 10:00:00", Mode: mcptokens.ModeModify, Features: []string{"dns"}},
		},
	}
	w := httptest.NewRecorder()
	if err := mcpPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	// upstream playwright checks input[name="read_only"], so there must be exactly one
	if n := strings.Count(body, `name="read_only"`); n != 1 {
		t.Errorf("%d read_only inputs", n)
	}
	for _, want := range []string{`name="allow_destructive"`, `name="scope_form"`, `value="dns"`, `value="domains"`, "No deletes", "DNS Zone Editor"} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %s", want)
		}
	}
	if strings.Contains(body, `value="mysql"`) {
		t.Error("offered a feature the plan doesn't have")
	}
}

func TestRenderMCPPageClientSnippets(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(newToken string) string {
		t.Helper()
		data := MCPPageData{LayoutData: baseLayout(mgr, "/account/mcp"), MCPURL: "https://panel.example.com:2083/mcp", NewToken: newToken}
		w := httptest.NewRecorder()
		if err := mcpPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	body := render("")
	for _, id := range []string{"claude-code", "claude-desktop", "codex", "cursor", "vscode", "gemini"} {
		if !strings.Contains(body, `id="`+id+`-snippet"`) || !strings.Contains(body, `<option value="`+id+`">`) {
			t.Errorf("no %s snippet or option", id)
		}
	}
	if strings.Count(body, "&lt;your-token&gt;") != 6 {
		t.Errorf("want the placeholder in all 6 snippets, got %d", strings.Count(body, "&lt;your-token&gt;"))
	}
	if !strings.Contains(body, `http_headers = { Authorization = &#34;Bearer &lt;your-token&gt;&#34; }`) && !strings.Contains(body, `http_headers = { Authorization = "Bearer &lt;your-token&gt;" }`) {
		t.Error("codex snippet missing")
	}

	body = render("op_mcp_newtoken123")
	// the banner plus all six snippets
	if n := strings.Count(body, "op_mcp_newtoken123"); n != 7 {
		t.Errorf("new token shows up %d times, want 7", n)
	}
	if strings.Contains(body, "&lt;your-token&gt;") {
		t.Error("placeholder left in after generating a token")
	}
}
