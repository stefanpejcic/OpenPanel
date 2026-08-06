// Command openpanel is the OpenPanel HTTP server.
package main

import (
	"context"
	"crypto/sha256"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/csrf"

	openpanel "gist.github.com/stefanpejcic/openpanel"
	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	defaultListenAddr = ":2083"
	// defaultStaticOverrideDir is where an admin can drop replacement
	// copies of the few user-editable static files (custom.css, custom.js,
	// robots.txt, security.txt); everything else is served from the
	// binary's embedded static/ tree. Empty disables overrides.
	defaultStaticOverrideDir = "/etc/openpanel/openpanel/static"
)

var licenseErrorPage = web.MustLoadPage("system/license_error.html")

func main() {
	a, err := appctx.New()
	if err != nil {
		fatalLogger.Fatalf("BOOTSTRAP - failed to initialize app: %v", err)
	}
	defer a.Close()

	configureLogging(a)
	setGOMAXPROCSFromCgroup()
	runStartupTasks()

	// AES-256 key for CSRF token encryption, derived from the panel secret
	// so the on-disk key file doesn't need to be exactly 32 bytes.
	csrfKey := sha256.Sum256(a.SecretKey)
	csrfMiddleware := csrf.Protect(csrfKey[:],
		csrf.Path("/"),
		csrf.CookieName("OPENPANEL_CSRF"),
		csrf.Secure(false),
		// Matches the "csrf_token" hidden input name every template
		// renders - gorilla/csrf's own default field name is
		// "gorilla.csrf.Token", which none of the forms use.
		csrf.FieldName("csrf_token"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Static assets are embedded in the binary; a handful of
	// admin-editable files (custom.css/js, robots.txt, security.txt) are
	// served from disk instead when present there.
	staticOverrideDir := envOrDefault("STATIC_OVERRIDE_DIR", defaultStaticOverrideDir)
	staticFS, err := fs.Sub(openpanel.Static, "static")
	if err != nil {
		log.Fatalf("BOOTSTRAP - failed to load embedded static assets: %v", err)
	}
	staticAssets, err := web.NewStaticAssets(staticFS, staticOverrideDir)
	if err != nil {
		log.Fatalf("BOOTSTRAP - failed to load embedded static assets: %v", err)
	}
	a.CustomCSS, a.CustomJS = staticAssets.CustomCSS, staticAssets.CustomJS
	log.Printf("BOOTSTRAP - static assets embedded (override dir %s: custom_css=%v custom_js=%v)",
		staticOverrideDir, a.CustomCSS, a.CustomJS)
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticAssets.Handler()))
	// Only these two root-level filenames are served this way; anything
	// else 404s.
	mux.HandleFunc("GET /robots.txt", staticAssets.ServeRootFile("robots.txt"))
	mux.HandleFunc("GET /security.txt", staticAssets.ServeRootFile("security.txt"))

	modules.RegisterAll(mux, a)

	// block_blacklisted_user_agents (config-gated, off by default) isn't
	// ported yet; not needed for any route registered so far.
	var handler http.Handler = mux
	handler = auth.CheckLicense(a, func(w http.ResponseWriter, r *http.Request) {
		renderLicenseError(a, w, r)
	})(handler)
	handler = auth.LoadUser(a)(handler)
	handler = exemptAPIFromCSRF(csrfMiddleware, handler)

	listenAddr := envOrDefault("LISTEN_ADDR", defaultListenAddr)
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	certFile, keyFile, useTLS := tlsCertPaths(listenAddr)
	go func() {
		log.Printf("BOOTSTRAP - listening on %s (tls=%v)", listenAddr, useTLS)
		var serveErr error
		if useTLS {
			serveErr = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			serveErr = srv.ListenAndServe()
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			fatalLogger.Fatalf("BOOTSTRAP - server error: %v", serveErr)
		}
	}()
	go flushRedisCache()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Print("BOOTSTRAP - shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("BOOTSTRAP - shutdown error: %v", err)
	}
}

// exemptAPIFromCSRF exempts /api/* and /mcp routes from CSRF protection.
// gorilla/csrf has no built-in path exemption, so this wraps it manually
// instead: those routes authenticate via a Bearer JWT/MCP token, not the
// session cookie CSRF protection exists to guard, so there's nothing for
// it to protect on this path family anyway.
func exemptAPIFromCSRF(csrfMiddleware func(http.Handler) http.Handler, next http.Handler) http.Handler {
	protected := csrfMiddleware(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/mcp" {
			next.ServeHTTP(w, r)
			return
		}
		protected.ServeHTTP(w, r)
	})
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func renderLicenseError(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	sessionLocale, _ := sess.Values["locale"].(string)
	locale := a.I18n.ResolveLocale(r.Context(), sessionLocale, "", r.Header.Get("Accept-Language"))

	data := map[string]any{"T": a.I18n.Translator(locale)}
	if err := licenseErrorPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("APP - license_error template render error: %v", err)
	}
}
