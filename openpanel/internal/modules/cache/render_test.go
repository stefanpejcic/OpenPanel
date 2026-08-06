package cache

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "redis": true, "varnish": true, "docker": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderCacheServicePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("not running", func(t *testing.T) {
		layout := baseLayout(mgr, "/cache/redis")
		layout.Title = "redis"
		data := CachePageData{
			LayoutData: layout, Service: "redis", Description: redisDef.Description, Port: 6379,
			ContainerState: "not_found", StatusKey: "not_found", StatusColor: "gray-400", StatusLabel: "Disabled",
		}
		w := httptest.NewRecorder()
		if err := cacheServicePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "6379") {
			t.Error("expected port in body")
		}
		if !strings.Contains(body, "Click to Enable") {
			t.Error("expected enable button when not_found")
		}
	})

	t.Run("running", func(t *testing.T) {
		layout := baseLayout(mgr, "/cache/redis")
		layout.Title = "redis"
		data := CachePageData{
			LayoutData: layout, Service: "redis", Description: redisDef.Description, Port: 6379,
			ContainerState: "running", StatusKey: "running", StatusColor: "emerald-500", StatusLabel: "Running",
			IsRunning: true,
		}
		w := httptest.NewRecorder()
		if err := cacheServicePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Click to Disable") {
			t.Error("expected disable button when running")
		}
		if !strings.Contains(body, "containerStats") {
			t.Error("expected container stats widget when running")
		}
	})
}

func TestRenderVarnishPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/cache/varnish")
	layout.Title = "Varnish"
	data := CachePageData{
		LayoutData: layout, Service: "varnish", Description: "Varnish desc",
		ContainerState: "running", StatusKey: "running", StatusColor: "emerald-500", StatusLabel: "Running",
		IsRunning: true,
		VarnishDomains: []DomainVarnishRow{
			{DomainURL: "example.com", Status: "On"},
			{DomainURL: "other.com", Status: "Off"},
		},
	}
	w := httptest.NewRecorder()
	if err := cacheServicePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "example.com") || !strings.Contains(body, "other.com") {
		t.Error("expected domain rows in body")
	}
	if !strings.Contains(body, "varnishStats") {
		t.Error("expected varnish stats widget when running")
	}
	if strings.Contains(body, `id="port"`) {
		t.Error("varnish page should not render the TCP port card")
	}
}

func TestRenderVarnishPage_NotRunning(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CachePageData{
		LayoutData: baseLayout(mgr, "/cache/varnish"), Service: "varnish",
		ContainerState: "not_found", StatusKey: "not_found", StatusColor: "gray-400", StatusLabel: "Disabled",
		VarnishDomains: []DomainVarnishRow{{DomainURL: "example.com", Status: "Unknown"}},
	}
	w := httptest.NewRecorder()
	if err := cacheServicePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), "varnishStats") {
		t.Error("should not fetch varnish stats when the container isn't running")
	}
}
