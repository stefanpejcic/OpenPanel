package domains

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{
		"dashboard": true, "domains": true, "docroot": true, "redirects": true, "ssl": true,
		"edit_vhost": true, "domain_suspend": true, "capitalize_domains": true, "filemanager": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderDomainsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := DomainsPageData{LayoutData: baseLayout(mgr, "/domains")}
		w := httptest.NewRecorder()
		if err := domainsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet.") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with rows", func(t *testing.T) {
		rows := []DomainRow{
			{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/example.com", PHPVersion: "8.2", SiteCount: 2, HTTPS: "Automatic", Status: "Not Suspended"},
			{DomainID: 2, DomainURL: "blog.example.com", IsSubdomain: true, Status: "Suspended", SuspendComment: "unpaid invoice"},
		}
		data := DomainsPageData{LayoutData: baseLayout(mgr, "/domains"), Domains: rows, TotalDomains: 2, TotalPages: 1, CurrentPage: 1}
		w := httptest.NewRecorder()
		if err := domainsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "example.com") {
			t.Error("expected domain row")
		}
		if !strings.Contains(body, "unpaid invoice") {
			t.Error("expected suspend comment for suspended domain")
		}
	})
}

func TestRenderNewDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := NewDomainPageData{LayoutData: baseLayout(mgr, "/domains/new")}
	w := httptest.NewRecorder()
	if err := newDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "addDomainForm") {
		t.Error("expected add-domain form")
	}
}

func TestRenderDeleteDomainPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no blockers", func(t *testing.T) {
		data := DeleteDomainPageData{LayoutData: baseLayout(mgr, "/domains/delete"), DomainURL: "example.com", Docroot: "/var/www/html/example.com"}
		w := httptest.NewRecorder()
		if err := deleteDomainPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `>Delete Domain <svg`) {
			t.Error("expected delete button when no sites/subdomains block it")
		}
	})

	t.Run("blocked by sites", func(t *testing.T) {
		data := DeleteDomainPageData{
			LayoutData: baseLayout(mgr, "/domains/delete"), DomainURL: "example.com",
			SiteCount: 1, Sites: []SiteRow{{SiteName: "example.com"}},
		}
		w := httptest.NewRecorder()
		if err := deleteDomainPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if strings.Contains(w.Body.String(), `>Delete Domain <svg`) {
			t.Error("expected delete button hidden when sites exist")
		}
	})
}

func TestRenderSuspendUnsuspendPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("suspend selector", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/suspend"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := suspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain selector option")
		}
	})

	t.Run("suspend confirm", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/suspend"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := suspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "confirm-suspend") {
			t.Error("expected confirm button")
		}
	})

	t.Run("unsuspend confirm", func(t *testing.T) {
		data := DomainSelectorPageData{LayoutData: baseLayout(mgr, "/domains/unsuspend"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := unsuspendPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "confirm-unsuspend") {
			t.Error("expected confirm button")
		}
	})
}

func TestRenderDocrootPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DocrootPageData{LayoutData: baseLayout(mgr, "/domains/docroot"), DomainName: "example.com", Docroot: "/var/www/html/example.com"}
	w := httptest.NewRecorder()
	if err := docrootPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "/var/www/html/example.com") {
		t.Error("expected current docroot value")
	}
}

func TestRenderDomainLogsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := DomainLogsPageData{LayoutData: baseLayout(mgr, "/domains/log"), Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := domainLogsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain selector")
		}
	})

	t.Run("with entries", func(t *testing.T) {
		entries := []AccessLogEntry{{TS: 1700000000, Status: 200, Request: AccessLogRequest{ClientIP: "1.2.3.4", Method: "GET", URI: "/index.html"}}}
		data := DomainLogsPageData{
			LayoutData: baseLayout(mgr, "/domains/log/example.com"), DomainName: "example.com",
			JSONLogs: entries, CurrentPage: 1, TotalPages: 1, TotalLines: 1, ItemsPerPage: 1000,
		}
		w := httptest.NewRecorder()
		if err := domainLogsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "1.2.3.4") || !strings.Contains(body, "/index.html") {
			t.Error("expected log entry fields")
		}
	})
}

func TestRenderCapitalizePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CapitalizePageData{LayoutData: baseLayout(mgr, "/domains/capitalize/example.com"), DomainURL: "example.com", CapitalizedDomain: "ExampleDomain.com"}
	w := httptest.NewRecorder()
	if err := capitalizePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "ExampleDomain.com") {
		t.Error("expected capitalized domain value")
	}
}

func TestRenderVhostPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("selector", func(t *testing.T) {
		data := VhostPageData{LayoutData: baseLayout(mgr, "/domains/vhosts"), WebServerPreference: "nginx", Domains: []appctx.Domain{{DomainURL: "example.com"}}}
		w := httptest.NewRecorder()
		if err := vhostPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Nginx") {
			t.Error("expected capitalized web server name")
		}
	})

	t.Run("editor", func(t *testing.T) {
		data := VhostPageData{LayoutData: baseLayout(mgr, "/domains/vhosts"), DomainName: "example.com", WebServerPreference: "nginx", VhostContent: "server { listen 80; }"}
		w := httptest.NewRecorder()
		if err := vhostPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "server { listen 80; }") {
			t.Error("expected vhost content in textarea")
		}
	})
}

func TestRenderRedirectPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("create", func(t *testing.T) {
		data := RedirectPageData{LayoutData: baseLayout(mgr, "/domains/redirect"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := redirectPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if strings.Contains(w.Body.String(), "Delete Redirect") {
			t.Error("expected no delete button when no redirect exists yet")
		}
	})

	t.Run("edit", func(t *testing.T) {
		data := RedirectPageData{LayoutData: baseLayout(mgr, "/domains/redirect"), DomainName: "example.com", RedirectURL: "https://target.example"}
		w := httptest.NewRecorder()
		if err := redirectPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Delete Redirect") {
			t.Error("expected delete button when redirect exists")
		}
		if !strings.Contains(body, "https://target.example") {
			t.Error("expected redirect URL value")
		}
	})
}

func TestRenderSSLPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no certificate", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), DomainName: "example.com"}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No Certificate!") {
			t.Error("expected no-certificate state")
		}
	})

	t.Run("autossl", func(t *testing.T) {
		data := SSLPageData{LayoutData: baseLayout(mgr, "/domains/ssl"), DomainName: "example.com", CurrentSetting: "autossl", Keys: "cert info here"}
		w := httptest.NewRecorder()
		if err := sslPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Auto SSL") {
			t.Error("expected Auto SSL status")
		}
	})
}
