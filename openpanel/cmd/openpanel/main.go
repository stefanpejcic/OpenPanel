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
	// also tlsCertPaths' fallback (startup.go) when `opencli port` can't be read
	defaultAdminPort  = "2083"
	defaultListenAddr = ":" + defaultAdminPort
	// where an admin can drop replacement static files (custom.css/js, robots.txt, security.txt); empty disables overrides
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

	// derive the AES-256 CSRF key from the panel secret so the on-disk key file doesn't need to be exactly 32 bytes
	csrfKey := sha256.Sum256(a.SecretKey)
	csrfMiddleware := csrf.Protect(csrfKey[:],
		csrf.Path("/"),
		csrf.CookieName("OPENPANEL_CSRF"),
		csrf.Secure(false),
		// matches the "csrf_token" name our templates use, not gorilla/csrf's default "gorilla.csrf.Token"
		csrf.FieldName("csrf_token"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// static assets are embedded, but a few admin-editable files are served from disk instead when present there
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
	// only these two root-level files are served this way, anything else 404s
	mux.HandleFunc("GET /robots.txt", staticAssets.ServeRootFile("robots.txt"))
	mux.HandleFunc("GET /security.txt", staticAssets.ServeRootFile("security.txt"))

	modules.RegisterAll(mux, a)

	// block_blacklisted_user_agents isn't ported yet, not needed for any route we have so far
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
	go flushRedisCache(a.Cache)

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

// exemptAPIFromCSRF skips CSRF checks for /api/* and /mcp - gorilla/csrf has no built-in path exemption, and those routes use a Bearer token instead of the session cookie anyway.
func exemptAPIFromCSRF(csrfMiddleware func(http.Handler) http.Handler, next http.Handler) http.Handler {
	protected := csrfMiddleware(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/mcp" {
			next.ServeHTTP(w, r)
			return
		}
		// gorilla/csrf >=1.7.3 rejects plain-HTTP Referers by default, but this panel usually runs plain HTTP behind Caddy's TLS-terminating proxy (or bare http://ip:port before a domain is set up), so we have to tell it explicitly when the request isn't really HTTPS
		if requestScheme(r) != "https" {
			r = csrf.PlaintextHTTPRequest(r)
		}
		protected.ServeHTTP(w, r)
	})
}

// requestScheme returns the client-facing scheme, trusting X-Forwarded-Proto first since Caddy terminates TLS in front of us and r.TLS is nil even over https
func requestScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(strings.TrimSpace(strings.SplitN(proto, ",", 2)[0]))
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
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
