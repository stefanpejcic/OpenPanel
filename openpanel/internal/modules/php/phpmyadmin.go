package php

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

const sharedPMAContainer = "phpmyadmin"

// getPMABaseURL builds the base URL phpMyAdmin is reachable at for this
// user, preferring a configured force-domain with SSL over the raw IP.
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

// pmaTokenReuseWindow bounds how long a just-minted token stays valid for
// reuse. Regenerating the token on every single GET to /phpmyadmin/ is fine
// for one request, but a second near-simultaneous hit (double-click, a
// restored tab, a link-prefetch) would overwrite the file before the first
// request's redirect is followed, so pma.php sees a token that no longer
// matches and bounces to index.php?invalid. Reusing a still-fresh token
// turns any cluster of hits within the window into one shared, valid
// redirect target instead of a race.
const pmaTokenReuseWindow = 60 * time.Second

// pmaTokenLocks serializes ensureUserToken per userContext so two
// near-simultaneous requests (double-click, a link-prefetch, a duplicate
// browser request) can't race: without a lock, both could see "no fresh
// token" at the same time, both mint different tokens, and whichever wins
// the final write invalidates the other's already-built redirect URL before
// the browser ever follows it - pmaTokenReuseWindow alone only shrinks that
// window, it doesn't close it.
var pmaTokenLocks sync.Map // userContext (string) -> *sync.Mutex

func pmaTokenLockFor(userContext string) *sync.Mutex {
	v, _ := pmaTokenLocks.LoadOrStore(userContext, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// ensureUserToken reuses a recently-written token instead of
// unconditionally minting a new one - see pmaTokenReuseWindow.
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
	// os.WriteFile alone leaves the write in this process's page cache;
	// the phpMyAdmin container reads /home/<user>/pma.token through its own
	// bind-mount of the same path, and without an explicit fsync the
	// browser can follow the redirect to pma.php before that write is
	// visible on the other side of the mount, so pma.php reads a stale
	// (empty or short) file and bounces to index.php?invalid - reproduced
	// live: an identical token succeeded on a bare retry a moment later
	// with no code change, which only makes sense as a write-visibility
	// race, not a logic bug.
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

// pmaProbeClient never follows redirects itself - probePHPMyAdminToken
// needs to inspect the FIRST redirect's Location header (?server=N vs
// ?invalid), not whatever page that target eventually renders.
var pmaProbeClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// probePHPMyAdminToken issues the exact request the browser is about to
// make and reports whether pma.php accepted the token (redirected to
// ./index.php?server=N) rather than rejecting it (./index.php?invalid or
// anything else).
func probePHPMyAdminToken(ctx context.Context, phpmyadminURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, phpmyadminURL, nil)
	if err != nil {
		return false
	}
	resp, err := pmaProbeClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	location := resp.Header.Get("Location")
	return resp.StatusCode == http.StatusFound && strings.Contains(location, "server=") && !strings.Contains(location, "invalid")
}

// pmaProbeRetries/pmaProbeRetryDelay bound how long a failed autologin
// attempt is retried server-side before giving up and redirecting the
// browser anyway. A "cold" hit (phpMyAdmin's own container/MySQL socket
// not touched in a while) has been observed taking multiple seconds to
// settle before pma.php reliably accepts a token, while a "warm" hit
// (recent prior activity) typically succeeds on the very first probe -
// this budget is sized for the cold case since that's the one that
// otherwise strands the user on index.php?invalid.
const (
	pmaProbeRetries    = 15
	pmaProbeRetryDelay = 300 * time.Millisecond
)

func probePHPMyAdminTokenWithRetry(ctx context.Context, phpmyadminURL string) bool {
	for attempt := 0; attempt < pmaProbeRetries; attempt++ {
		if probePHPMyAdminToken(ctx, phpmyadminURL) {
			return true
		}
		time.Sleep(pmaProbeRetryDelay)
	}
	return false
}

func isPMAContainerRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "podman", "inspect", "-f", "{{.State.Running}}", sharedPMAContainer).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(strings.ToLower(string(out))) == "true"
}

// handlePHPMyAdminRedirect mints (or reuses) an autologin token and
// redirects the browser to phpMyAdmin, probing the token server-side first
// so the user isn't sent to a dead end.
func handlePHPMyAdminRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	log.Printf("PHPMYADMIN - DEBUG - redirect request from %s: user=%s context=%s ua=%q", reqip.ClientIP(r), currentUsername, userContext, r.UserAgent())

	if !isPMAContainerRunning(ctx) {
		renderPHPMyAdminUnavailablePage(a, w, r, "Please contact support.", http.StatusServiceUnavailable)
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

	token, tokenErr := ensureUserToken(ctx, a, userContext)
	if tokenErr != nil {
		log.Printf("PHPMYADMIN - DEBUG - ensureUserToken failed for context=%s: %v", userContext, tokenErr)
		renderPHPMyAdminUnavailablePage(a, w, r, "Failed to generate autologin token.", http.StatusInternalServerError)
		return
	}
	log.Printf("PHPMYADMIN - DEBUG - wrote token for context=%s", userContext)

	pmaBaseURL := getPMABaseURL(ctx, a, currentUsername)
	phpmyadminURL := pmaBaseURL + "/pma.php?user=" + userContext + "&token=" + token
	if db := r.URL.Query().Get("db"); db != "" {
		phpmyadminURL += "&db=" + db
	}

	// pma.php's validation of a token this same handler just wrote a
	// moment ago has been observed to intermittently fail on the very
	// next read (bouncing to index.php?invalid) even though the token
	// file's content, ownership, and permissions are already correct on
	// this host at that instant - a cross-container read race of some
	// kind between this process's write and the phpMyAdmin container's
	// own view of the same bind-mounted path. Since pma.php's token check
	// is a plain re-readable file compare (not one-time-use - verified
	// live, the same token validates successfully on repeated reads), a
	// server-side probe of the exact URL we're about to send the browser
	// to is a safe, harmless way to confirm it actually works before
	// committing the user to that redirect, retrying past the race
	// instead of sending them to a dead end.
	if ok := probePHPMyAdminTokenWithRetry(ctx, phpmyadminURL); !ok {
		log.Printf("PHPMYADMIN - DEBUG - token probe kept failing for context=%s after retries, redirecting anyway (best effort)", userContext)
	}

	log.Printf("PHPMYADMIN - DEBUG - redirecting context=%s to %s", userContext, strings.Replace(phpmyadminURL, token, "REDACTED", 1))

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "opened phpMyAdmin", ipAddress)
	http.Redirect(w, r, phpmyadminURL, http.StatusFound)
}

// handlePHPMyAdminLoginLink redirects to phpMyAdmin's manual login form
// for this user's context.
func handlePHPMyAdminLoginLink(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
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
