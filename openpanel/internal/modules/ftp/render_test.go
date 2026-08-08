package ftp

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "ftp": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderFTPAccountsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := FTPAccountsPageData{LayoutData: baseLayout(mgr, "/ftp"), ServerIP: "1.2.3.4", DedicatedIP: "Unknown", FTPHost: "1.2.3.4", FTPPort: "21"}
		w := httptest.NewRecorder()
		if err := ftpAccountsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No accounts yet.") {
			t.Error("expected empty-state message")
		}
		if !strings.Contains(w.Body.String(), "1.2.3.4") {
			t.Error("expected server IP fallback since dedicated IP is Unknown")
		}
	})

	t.Run("with accounts", func(t *testing.T) {
		data := FTPAccountsPageData{
			LayoutData: baseLayout(mgr, "/ftp"), ServerIP: "1.2.3.4", DedicatedIP: "9.9.9.9",
			FTPHost: "9.9.9.9", FTPPort: "21",
			Accounts: []Account{{Username: "bob@example.com", Path: "/var/www/html/", UID: "1000", GID: "1000"}},
		}
		w := httptest.NewRecorder()
		if err := ftpAccountsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "bob@example.com") {
			t.Error("expected account row in rendered page")
		}
		if !strings.Contains(body, "9.9.9.9") {
			t.Error("expected dedicated IP to be used when set")
		}
	})
}

func TestRenderFTPConnectionsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no connections", func(t *testing.T) {
		data := FTPConnectionsPageData{LayoutData: baseLayout(mgr, "/ftp/connections")}
		w := httptest.NewRecorder()
		if err := ftpConnectionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No active connections") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with connections", func(t *testing.T) {
		data := FTPConnectionsPageData{LayoutData: baseLayout(mgr, "/ftp/connections"), ConnectionsOutput: "bob@example.com 10.0.0.1", HasConnections: true}
		w := httptest.NewRecorder()
		if err := ftpConnectionsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "bob@example.com 10.0.0.1") {
			t.Error("expected connections output in rendered page")
		}
	})
}

func TestRenderNewFTPAccountPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no domains", func(t *testing.T) {
		data := FTPNewPageData{LayoutData: baseLayout(mgr, "/ftp/new")}
		w := httptest.NewRecorder()
		if err := ftpNewPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet") {
			t.Error("expected no-domains empty state")
		}
	})

	t.Run("with domains", func(t *testing.T) {
		data := FTPNewPageData{LayoutData: baseLayout(mgr, "/ftp/new"), Domains: []domainOption{{DomainURL: "example.com", Docroot: "/var/www/html/example.com"}}}
		w := httptest.NewRecorder()
		if err := ftpNewPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "example.com") {
			t.Error("expected domain option in rendered page")
		}
	})
}

func TestRenderFTPUsernamePages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := FTPUsernamePageData{LayoutData: baseLayout(mgr, "/ftp/password/bob"), Username: "bob@example.com"}

	w := httptest.NewRecorder()
	if err := ftpPasswordPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render password page: %v", err)
	}
	if !strings.Contains(w.Body.String(), "bob@example.com") {
		t.Error("expected username in rendered password page")
	}

	w2 := httptest.NewRecorder()
	if err := ftpPathPage.Render(w2, 200, data); err != nil {
		t.Fatalf("Render path page: %v", err)
	}
	if !strings.Contains(w2.Body.String(), "bob@example.com") {
		t.Error("expected username in rendered path page")
	}
}

func TestIsValidUsername(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"bob@example.com", true},
		{"", false},
		{"nodomain", false},
		{"two@@ats.com", false},
		{"@example.com", false},
		{"bob@", false},
		{"bob!@example.com", false},
		{"bob@exam ple.com", false},
	}
	for _, c := range cases {
		if got := isValidUsername(c.in); got != c.want {
			t.Errorf("isValidUsername(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLoadAccounts(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/users.list"
	content := "bob@example.com|pass|/var/www/html/|1000|1000\n" +
		"alice@example.com|pass|/var/www/html/alice\n" +
		"malformed-line\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	accounts := loadAccounts(path)
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d: %+v", len(accounts), accounts)
	}
	if accounts[0].Username != "bob@example.com" || accounts[0].UID != "1000" {
		t.Errorf("unexpected first account: %+v", accounts[0])
	}
	if accounts[1].Username != "alice@example.com" || accounts[1].UID != "" {
		t.Errorf("unexpected second account: %+v", accounts[1])
	}
}
