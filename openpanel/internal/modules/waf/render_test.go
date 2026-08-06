package waf

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "waf": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderWAFListPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no domains", func(t *testing.T) {
		data := WAFListPageData{LayoutData: baseLayout(mgr, "/server/waf")}
		w := httptest.NewRecorder()
		if err := wafListPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet.") {
			t.Error("expected empty state text")
		}
	})

	t.Run("with domains and issues", func(t *testing.T) {
		data := WAFListPageData{
			LayoutData: baseLayout(mgr, "/server/waf"),
			Domains:    []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "off.example.com"}},
			ModsecStatus: map[string]string{
				"example.com": "On", "off.example.com": "Off",
			},
			Issues: []WAFIssue{{ID: "waf-disabled:off.example.com", Severity: "warning", Message: "WAF is disabled for off.example.com."}},
		}
		w := httptest.NewRecorder()
		if err := wafListPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "example.com") || !strings.Contains(body, "off.example.com") {
			t.Error("expected domains in body")
		}
		if !strings.Contains(body, "reportHealthIssues") {
			t.Error("expected health toast script when issues are present")
		}
	})
}

func TestRenderWAFDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WAFDomainPageData{
		LayoutData: baseLayout(mgr, "/server/waf/example.com"), Domain: "example.com", Status: "On",
		RemovedRules: []string{"123", "456"}, RemovedTags: []string{"foo"},
	}
	w := httptest.NewRecorder()
	if err := wafDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "example.com") || !strings.Contains(body, "123 456") {
		t.Error("expected domain and removed rules in body")
	}
}

func TestRenderWAFLogSelectPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WAFLogsPageData{LayoutData: baseLayout(mgr, "/server/waf/log"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
	w := httptest.NewRecorder()
	if err := wafLogsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "example.com") || !strings.Contains(body, "/server/waf/log/${selectedVersion}") {
		t.Error("expected domain picker with plain domain-url redirect (no username prefix) in body")
	}
	if strings.Contains(body, "domain.username") {
		t.Error("did not expect a reference to the nonexistent domain.username field")
	}
}

func TestRenderWAFLogPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	logs := []json.RawMessage{json.RawMessage(`{"transaction":{"client_ip":"1.2.3.4"}}`)}
	data := WAFLogsPageData{
		LayoutData: baseLayout(mgr, "/server/waf/log/example.com"), DomainName: "example.com",
		JSONLogs: logs, CurrentPage: 1, ItemsPerPage: 1000, TotalPages: 1, TotalLines: 1,
		TotalAllowedLinesForShowAll: 10000, PageEntries: buildLogPageEntries(1, 1),
	}
	w := httptest.NewRecorder()
	if err := wafLogsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "example.com") || !strings.Contains(body, `"client_ip":"1.2.3.4"`) {
		t.Error("expected domain name and embedded log JSON in body")
	}
}
