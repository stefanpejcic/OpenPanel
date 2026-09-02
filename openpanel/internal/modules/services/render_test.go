package services

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "services": true, "docker": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderServicesListPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	data := ServicesPageData{LayoutData: baseLayout(mgr, "/services/"), Services: []string{"redis", "nginx"}}
	if err := servicesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Choose Service") || !strings.Contains(body, `value="redis"`) || !strings.Contains(body, `value="nginx"`) {
		t.Errorf("rendered services list page missing expected content:\n%s", body)
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

// TestStatusColorLabelStopping guards against "stopping" - libpod's real
// State.Status for a container mid-shutdown - falling through to the
// generic "Unknown" fallback. Confirmed live: a page render landing in
// that few-second window (between a container being told to stop and it
// actually reaching "exited"/"removing") showed literal "Unknown" with no
// mapped label at all.
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
