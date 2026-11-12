package mongodb

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRenderDatabasesUsersAndSize(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DatabasesPageData{
		LayoutData:        baseLayout(mgr, "/mongodb"),
		ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
		Databases: []DatabaseRow{
			{Database: "shop", SizeBytes: 2 * 1024 * 1024}, {Database: "blog", SizeBytes: 512 * 1024},
			{Database: "admin", SizeBytes: 50 * 1024 * 1024, IsSystem: true},
		},
		UserCount: 3, HasUsers: true,
	}
	if got := data.TotalSize(); got != "2.50 MB" {
		t.Errorf("TotalSize = %q", got)
	}
	if got := data.LargestDatabase(); got != "shop" {
		t.Errorf("LargestDatabase = %q", got)
	}
	w := httptest.NewRecorder()
	if err := databasesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"2.50 MB", "Largest: shop", `<div class="op-kpi-value">3</div>`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}

func TestShouldCountUsers(t *testing.T) {
	for _, c := range []struct {
		state string
		dbs   int
		want  bool
	}{
		{"running", 0, true},
		{"running", 2, true},
		{"exited", 0, false},
		{"", 0, false},
		{"exited", 1, true},
	} {
		if got := shouldCountUsers(c.state, c.dbs); got != c.want {
			t.Errorf("shouldCountUsers(%q, %d) = %v, want %v", c.state, c.dbs, got, c.want)
		}
	}
}
