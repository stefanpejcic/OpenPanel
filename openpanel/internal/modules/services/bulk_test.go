package services

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestServicesListBulkBar(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/services/")
	layout.BulkActions = servicesBulkActions(layout.T)
	data := servicesListData(layout, []string{"nginx", "redis"}, []docker.ContainerStatus{{State: "running"}, {State: "exited"}}, "nginx", "mysql")
	w := httptest.NewRecorder()
	if err := servicesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"/services/bulk", "#services-table", `data-bulk-action="restart"`, `data-bulk-action="start"`, `data-bulk-action="stop"`, `value="nginx" data-bulk-skip="start"`, `value="redis" data-bulk-skip="restart,stop"`} {
		if !strings.Contains(body, want) {
			t.Errorf("services list missing %q", want)
		}
	}
}

func TestRunServiceBulkActionRejects(t *testing.T) {
	r := httptest.NewRequest("POST", "/services/bulk", nil)
	if res := runServiceBulkAction(r, "restart", "apache", "testuser", []string{"nginx"}); res.OK || res.Message != "Service not found." {
		t.Errorf("service outside the list = %+v, want not found", res)
	}
	if res := runServiceBulkAction(r, "delete", "nginx", "testuser", []string{"nginx"}); res.OK || res.Message != "Unknown bulk action." {
		t.Errorf("unknown action = %+v", res)
	}
}
