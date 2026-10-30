package postgresql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestShortPGVersion(t *testing.T) {
	cases := map[string]string{"17.4": "17.4", "16.2 (Debian 16.2-1.pgdg120+2)": "16.2", "": ""}
	for in, want := range cases {
		if got := shortPGVersion(in); got != want {
			t.Errorf("shortPGVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderDatabasesServerInfo(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DatabasesPageData{
		LayoutData:        baseLayout(mgr, "/postgresql"),
		ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
		Databases:         []DatabaseRow{{Database: "app"}, {Database: "shop"}},
		ServerInfo:        ServerInfo{ServerVersion: "17.4", Connections: 2, HasInfo: true}, DBLimit: 10,
	}
	w := httptest.NewRecorder()
	if err := databasesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"PostgreSQL 17.4 · 2 connections", "2 <small>/ 10</small>", "width:20%"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}
