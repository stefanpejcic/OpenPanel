package services

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "services": true, "docker": true, "process_manager": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderServicesListPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	names := []string{"nginx", "mysql", "php-fpm-8.3", "redis", "memcached", "elasticsearch"}
	statuses := []docker.ContainerStatus{{State: "running", Health: "healthy", Started: time.Now().Add(-50 * time.Hour)}, {State: "running", Health: "none", Started: time.Now().Add(-3 * time.Hour)}, {State: "running"}, {State: "exited"}, {State: "running", Health: "starting"}, {State: "not_found"}}
	data := servicesListData(baseLayout(mgr, "/services/"), names, statuses, "nginx", "mysql")
	if err := servicesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"4 of 6 running", `data-href="/services/redis"`, "PHP 8.3", "Object cache", `value="enable"`, `value="restart"`, `name="service" value="nginx"`, "nginx · mysql", "op-status is-success", "op-status is-muted", "op-status is-warning", "2 days", "3 hours", `data-stat="pids"`} {
		if !strings.Contains(body, want) {
			t.Errorf("services list missing %q", want)
		}
	}
	// running first, then by name
	if i, j := strings.Index(body, `data-service="php-fpm-8.3"`), strings.Index(body, `data-service="redis"`); i < 0 || j < 0 || i > j {
		t.Error("running services should be listed before stopped ones")
	}
	if i, j := strings.Index(body, `data-service="memcached"`), strings.Index(body, `data-service="mysql"`); i < 0 || j < 0 || i > j {
		t.Error("services should be sorted by name within a group")
	}
	if strings.Contains(body, "service-select") {
		t.Error("services list still renders the old picker")
	}
}

func TestRenderServiceDetailPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("running", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ServicesPageData{
			LayoutData: baseLayout(mgr, "/services/nginx"), Service: "nginx", UserContext: "testuser",
			IsRunning: true, StatusKey: "running", StatusColor: "emerald-500", StatusLabel: "Running",
			StatusMapJSON: StatusMapJSON(mgr.Translator("en")),
			ActionValue:   "disable", ActionLabel: "Click to Disable",
			Kind: "Web server", StatusTone: "is-success", Uptime: "2 days",
		}
		if err := servicesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"nginx", "Click to Disable", `value="disable"`, "containerStats()", "/json/containers/log/nginx"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered running-service page missing %q", want)
			}
		}
	})

	t.Run("not_found shows enable button and no container stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ServicesPageData{
			LayoutData: baseLayout(mgr, "/services/redis"), Service: "redis", UserContext: "testuser",
			IsRunning: false, StatusKey: "not_found", StatusColor: "gray-400", StatusLabel: "Disabled",
			StatusMapJSON: StatusMapJSON(mgr.Translator("en")),
			ActionValue:   "enable", ActionLabel: "Click to Enable",
		}
		if err := servicesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Click to Enable") || !strings.Contains(body, `value="enable"`) {
			t.Error("expected enable button for a not_found service")
		}
		if strings.Contains(body, "containerStats()") {
			t.Error("didn't expect container stats widget for a non-running service")
		}
	})
}

func TestStatusKeyFor(t *testing.T) {
	cases := []struct {
		state, health, want string
	}{
		{"running", "healthy", "healthy"},
		{"running", "unhealthy", "unhealthy"},
		{"running", "starting", "starting"},
		{"running", "none", "running"},
		{"exited", "none", "exited"},
		{"not_found", "none", "not_found"},
	}
	for _, c := range cases {
		if got := StatusKeyFor(c.state, c.health); got != c.want {
			t.Errorf("StatusKeyFor(%q, %q) = %q, want %q", c.state, c.health, got, c.want)
		}
	}
}

// TestStatusColorLabelStopping guards against "stopping" (libpod's real State.Status mid-shutdown) falling through to the generic "Unknown" fallback - used to happen for real when a render landed before the container reached "exited"/"removing"
func TestStatusColorLabelStopping(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	_, label := StatusColorLabel(mgr.Translator("en"), StatusKeyFor("stopping", "none"))
	if label == "Unknown" {
		t.Errorf("StatusColorLabel for \"stopping\" fell through to the generic Unknown fallback")
	}
}

func TestFilterServices(t *testing.T) {
	all := []string{"nginx", "apache", "mysql", "mariadb", "redis", "phpmyadmin"}
	got := FilterServices(all, "nginx", "mysql")
	want := map[string]bool{"nginx": true, "mysql": true, "redis": true, "phpmyadmin": true}
	if len(got) != len(want) {
		t.Fatalf("FilterServices() = %v, want services matching %v", got, want)
	}
	for _, s := range got {
		if !want[s] {
			t.Errorf("FilterServices() unexpectedly included %q", s)
		}
	}
	for s := range want {
		found := false
		for _, g := range got {
			if g == s {
				found = true
			}
		}
		if !found {
			t.Errorf("FilterServices() missing expected service %q", s)
		}
	}
}
