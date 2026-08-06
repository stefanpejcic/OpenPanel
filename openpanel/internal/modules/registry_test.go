package modules

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// testApp builds a minimal *appctx.App good enough to exercise route
// wiring and any handler paths that don't touch the DB (e.g. the login
// page's GET request) - a lightweight substitute for appctx.New(), which
// requires reading real files under /etc/openpanel that aren't writable in
// a sandboxed test environment.
func testApp(t *testing.T) *appctx.App {
	t.Helper()
	c := cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))
	t.Cleanup(func() { c.Close() })

	return &appctx.App{
		Sessions:       session.NewStore([]byte("test-secret-key-for-registry-test")),
		Cache:          c,
		I18n:           i18n.NewManager(t.TempDir(), c),
		EnabledModules: []string{"dashboard", "websites", "docker", "services", "filemanager", "disk_usage", "inodes", "fix_permissions", "malware_scan", "trash", "ftp", "backup_wizard", "backups", "domains", "php", "dns", "dynamic_dns", "redis", "memcached", "elasticsearch", "opensearch", "valkey", "varnish", "mysql", "mysql_conf", "mysql_import", "mysql_processlist", "mysql_root_password", "remote_mysql", "emails", "email_aliases", "email_default", "email_deliverability", "email_export", "email_filters", "email_import", "webmail", "crons", "info", "usage", "process_manager", "ip_blocker", "webserver_conf", "waf", "account", "locale", "twofa", "passkeys", "notifications", "favorites", "sessions", "activity", "login_history", "mcp", "api", "postgresql", "postgresql_conf", "postgresql_import", "remote_postgresql", "python", "nodejs", "autoinstaller"},
	}
}

// TestRegisterAllWiresRoutes is an end-to-end smoke test of the module
// registry: real HTTP requests through the real mux, verifying the routes
// that don't need a live database/redis actually respond as app.py's
// equivalent routes would.
func TestRegisterAllWiresRoutes(t *testing.T) {
	a := testApp(t)
	mux := http.NewServeMux()
	RegisterAll(mux, a)

	t.Run("GET /login renders the login form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `name="username"`) {
			t.Errorf("expected login form in response body, got:\n%s", w.Body.String())
		}
	})

	t.Run("GET /openpanel redirects to /", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openpanel", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/" {
			t.Errorf("Location = %q, want /", got)
		}
	})

	t.Run("GET /this-route-does-not-exist 404s, even without a session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/this-route-does-not-exist", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404; body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /dashboard without a session redirects to /login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302; body: %s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/login" {
			t.Errorf("Location = %q, want /login", got)
		}
	})

	t.Run("GET / without a session redirects to /login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302; body: %s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/login" {
			t.Errorf("Location = %q, want /login (RequireLogin should reject an unauthenticated request)", got)
		}
	})

	t.Run("GET /logout clears session and redirects to /login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/logout", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/login" {
			t.Errorf("Location = %q, want /login", got)
		}
	})

	for _, path := range []string{
		"/account", "/settings", "/account/language", "/account/2fa",
		"/account/passkeys", "/account/notifications", "/account/favorites",
		"/account/sessions", "/account/activity", "/account/login-history",
		"/account/mcp", "/account/api",
		"/postgresql", "/postgresql/new", "/postgresql/users", "/postgresql/user",
		"/postgresql/wizard", "/postgresql/assign", "/postgresql/remove",
		"/postgresql/configuration", "/postgresql/import", "/postgresql/remote-postgresql",
		"/python/install", "/nodejs/install", "/auto-installer",
	} {
		t.Run("GET "+path+" without a session redirects to /login", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302; body: %s", w.Code, w.Body.String())
			}
			if got := w.Header().Get("Location"); got != "/login" {
				t.Errorf("Location = %q, want /login", got)
			}
		})
	}
}
