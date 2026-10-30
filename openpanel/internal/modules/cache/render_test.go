package cache

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "redis": true, "varnish": true, "docker": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
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

func TestRenderCacheServiceVariants(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, c := range []struct {
		def          serviceDef
		state, key   string
		running      bool
		addr, ping   string
		wantRestart  bool
		wantStatsBox bool
	}{
		{memcachedDef, "running", "healthy", true, "memcached:11211", "echo stats | nc memcached 11211", true, true},
		{opensearchDef, "running", "unhealthy", true, "http://opensearch:9200", "curl -s http://opensearch:9200", true, true},
		{valkeyDef, "exited", "exited", false, "tcp://valkey:6379", "valkey-cli -h valkey -p 6379 ping", true, false},
		{elasticsearchDef, "not_found", "not_found", false, "http://elasticsearch:9200", "curl -s http://elasticsearch:9200", false, false},
	} {
		t.Run(c.def.Name, func(t *testing.T) {
			layout := baseLayout(mgr, "/cache/"+c.def.Name)
			layout.Title = c.def.Title
			data := CachePageData{
				LayoutData: layout, Service: c.def.Name, Description: c.def.Description, Port: c.def.Port,
				ContainerState: c.state, StatusKey: c.key, StatusLabel: c.key, IsRunning: c.running,
			}
			if got := data.ConnectAddress(); got != c.addr {
				t.Errorf("ConnectAddress = %q, want %q", got, c.addr)
			}
			if got := data.PingCommand(); got != c.ping {
				t.Errorf("PingCommand = %q, want %q", got, c.ping)
			}
			w := httptest.NewRecorder()
			if err := cacheServicePage.Render(w, 200, data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			if strings.Contains(body, `value="restart"`) != c.wantRestart {
				t.Errorf("restart button shown = %v, want %v", !c.wantRestart, c.wantRestart)
			}
			if strings.Contains(body, "containerStats") != c.wantStatsBox {
				t.Errorf("container stats shown = %v, want %v", !c.wantStatsBox, c.wantStatsBox)
			}
			if !strings.Contains(body, "op-status "+data.StatusTone()) {
				t.Errorf("expected status pill with %s", data.StatusTone())
			}
		})
	}
}

func TestCacheConnectHelpers(t *testing.T) {
	cases := []struct {
		def        serviceDef
		addr, ping string
	}{
		{redisDef, "tcp://redis:6379", "redis-cli -h redis -p 6379 ping"},
		{valkeyDef, "tcp://valkey:6379", "valkey-cli -h valkey -p 6379 ping"},
		{memcachedDef, "memcached:11211", "echo stats | nc memcached 11211"},
		{elasticsearchDef, "http://elasticsearch:9200", "curl -s http://elasticsearch:9200"},
		{opensearchDef, "http://opensearch:9200", "curl -s http://opensearch:9200"},
	}
	for _, c := range cases {
		d := CachePageData{Service: c.def.Name, Port: c.def.Port}
		if got := d.ConnectAddress(); got != c.addr {
			t.Errorf("%s ConnectAddress = %q, want %q", c.def.Name, got, c.addr)
		}
		if got := d.PingCommand(); got != c.ping {
			t.Errorf("%s PingCommand = %q, want %q", c.def.Name, got, c.ping)
		}
	}
	for key, want := range map[string]string{"running": "is-success", "healthy": "is-success", "exited": "is-danger", "not_found": "is-muted", "starting": "is-warning"} {
		if got := (CachePageData{StatusKey: key}).StatusTone(); got != want {
			t.Errorf("StatusTone(%s) = %q, want %q", key, got, want)
		}
	}
}

// every generic service renders its own connection card when running
func TestRenderGenericServicesRunning(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, def := range []serviceDef{memcachedDef, opensearchDef} {
		layout := baseLayout(mgr, "/cache/"+def.Name)
		layout.Title = def.Title
		data := CachePageData{
			LayoutData: layout, Service: def.Name, Description: def.Description, Port: def.Port,
			ContainerState: "running", StatusKey: "healthy", StatusColor: "emerald-500", StatusLabel: "Running",
			IsRunning: true,
		}
		w := httptest.NewRecorder()
		if err := cacheServicePage.Render(w, 200, data); err != nil {
			t.Fatalf("%s Render: %v", def.Name, err)
		}
		body := w.Body.String()
		for _, want := range []string{data.ConnectAddress(), `value="restart"`, "is-success"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: expected %q in body", def.Name, want)
			}
		}
	}
}
