package php

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const sharedPMAContainer = "phpmyadmin"

// getPMABaseURL builds the base URL phpMyAdmin is reachable at for this user, preferring a configured force-domain with SSL over the raw IP
func getPMABaseURL(ctx context.Context, a *appctx.App, currentUsername string) string {
	if a.ForceDomain != "" {
		dynamicIP := sysinfo.FetchPublicIP(ctx, a.Cache)
		if a.ForceDomain != dynamicIP && sysinfo.HasSSL(ctx, a.Cache, a.ForceDomain) {
			return "https://" + a.ForceDomain + ":2053"
		}
	}
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
	return "http://" + serverIP + ":8888"
}

// pmaTokenReuseWindow bounds how long a just-minted token stays valid for reuse - a second near-simultaneous hit would otherwise overwrite the token file before the first request's redirect is followed, bouncing pma.php to index.php?invalid
const pmaTokenReuseWindow = 60 * time.Second

// pmaTokenLocks serializes ensureUserToken per userContext so two near-simultaneous requests can't both mint a token and race each other's redirect URL - pmaTokenReuseWindow alone only shrinks that window, it doesn't close it
var pmaTokenLocks sync.Map // userContext (string) -> *sync.Mutex

func pmaTokenLockFor(userContext string) *sync.Mutex {
	v, _ := pmaTokenLocks.LoadOrStore(userContext, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// ensureUserToken reuses a recently-written token instead of unconditionally minting a new one, see pmaTokenReuseWindow
func ensureUserToken(ctx context.Context, a *appctx.App, userContext string) (string, error) {
	lock := pmaTokenLockFor(userContext)
	lock.Lock()
	defer lock.Unlock()

	tokenFile := "/home/" + userContext + "/pma.token"
	if info, err := os.Stat(tokenFile); err == nil && time.Since(info.ModTime()) < pmaTokenReuseWindow {
		if existing, err := os.ReadFile(tokenFile); err == nil && len(existing) == 64 {
			return string(existing), nil
		}
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	// os.WriteFile alone leaves the write in this process's page cache - without an explicit fsync the phpMyAdmin container can read a stale token through its bind-mount before the write is visible on the other side
	f, err := os.OpenFile(tokenFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := f.Write([]byte(token)); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	uid, err := a.GetUID(ctx, userContext)
	if err != nil {
		return "", err
	}
	_ = os.Chown(tokenFile, uid, uid)
	return token, nil
}

// mysqld creates its socket a few seconds after compose up returns and pma.php rejects the login until it exists
const pmaSocketWait = 30 * time.Second

func waitForMySQLSocket(ctx context.Context, userContext string) bool {
	sock := "/home/" + userContext + "/sockets/mysqld/mysqld.sock"
	deadline := time.Now().Add(pmaSocketWait)
	for {
		conn, err := net.DialTimeout("unix", sock, time.Second)
		if err == nil {
			conn.Close()
			return true
		}
		// socket is there but we can't connect as this uid, pma.php only checks it exists anyway
		if errors.Is(err, os.ErrPermission) {
			return true
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func isPMAContainerRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "podman", "inspect", "-f", "{{.State.Running}}", sharedPMAContainer).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(strings.ToLower(string(out))) == "true"
}

// errPMAUnavailable means the shared phpMyAdmin container isn't running
var errPMAUnavailable = errors.New("phpMyAdmin is not running")

// buildPHPMyAdminAutologinURL mints (or reuses) an autologin token and returns the pma.php URL for it once the user's mysql socket is up
func buildPHPMyAdminAutologinURL(ctx context.Context, a *appctx.App, currentUsername, userContext, db string) (string, string, error) {
	start := time.Now()
	if !isPMAContainerRunning(ctx) {
		return "", "", errPMAUnavailable
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
	if !waitForMySQLSocket(ctx, userContext) {
		log.Printf("PHPMYADMIN - DEBUG - mysql socket not ready for context=%s after %s, returning url anyway", userContext, pmaSocketWait)
	}

	token, err := ensureUserToken(ctx, a, userContext)
	if err != nil {
		log.Printf("PHPMYADMIN - DEBUG - ensureUserToken failed for context=%s: %v", userContext, err)
		return "", "", err
	}
	log.Printf("PHPMYADMIN - DEBUG - wrote token for context=%s in %s", userContext, time.Since(start))

	phpmyadminURL := getPMABaseURL(ctx, a, currentUsername) + "/pma.php?user=" + userContext + "&token=" + token
	if db != "" {
		phpmyadminURL += "&db=" + url.QueryEscape(db)
	}

	return phpmyadminURL, token, nil
}

// handlePHPMyAdminRedirect sends the browser straight into phpMyAdmin via an autologin link
func handlePHPMyAdminRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	log.Printf("PHPMYADMIN - DEBUG - redirect request from %s: user=%s context=%s ua=%q", reqip.ClientIP(r), currentUsername, userContext, r.UserAgent())

	phpmyadminURL, token, err := buildPHPMyAdminAutologinURL(ctx, a, currentUsername, userContext, r.URL.Query().Get("db"))
	if errors.Is(err, errPMAUnavailable) {
		renderPHPMyAdminUnavailablePage(a, w, r, web.Tr(a, r, "Please contact support."), http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		renderPHPMyAdminUnavailablePage(a, w, r, web.Tr(a, r, "Failed to generate autologin token."), http.StatusInternalServerError)
		return
	}

	log.Printf("PHPMYADMIN - DEBUG - redirecting context=%s to %s", userContext, strings.Replace(phpmyadminURL, token, "REDACTED", 1))

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "opened phpMyAdmin", ipAddress)
	http.Redirect(w, r, phpmyadminURL, http.StatusFound)
}

// RegisterPHPMyAdminAPI wires the phpMyAdmin autologin API route onto mux, gated by the same "phpmyadmin" feature as the web routes
func RegisterPHPMyAdminAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "phpmyadmin", "GET /api/phpmyadmin", func(w http.ResponseWriter, r *http.Request) { apiPHPMyAdminAutologin(a, w, r) })
}

// apiPHPMyAdminAutologin returns a ready-to-open phpMyAdmin autologin url instead of redirecting, so API clients can hand it to a browser
func apiPHPMyAdminAutologin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	phpmyadminURL, _, err := buildPHPMyAdminAutologinURL(ctx, a, currentUsername, userContext, r.URL.Query().Get("db"))
	if errors.Is(err, errPMAUnavailable) {
		writeJSONError(w, http.StatusServiceUnavailable, "phpMyAdmin is not available, please contact support")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate autologin token")
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "generated phpMyAdmin autologin link via API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"url":       phpmyadminURL,
		"login_url": getPMABaseURL(ctx, a, currentUsername) + "/index.php?manual=" + userContext,
	})
}

// handlePHPMyAdminLoginLink redirects to phpMyAdmin's manual login form for this user's context
func handlePHPMyAdminLoginLink(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pmaBaseURL := getPMABaseURL(ctx, a, currentUsername)
	phpmyadminURL := pmaBaseURL + "/index.php?manual=" + userContext

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "opened phpMyAdmin login form", ipAddress)
	http.Redirect(w, r, phpmyadminURL, http.StatusFound)
}
