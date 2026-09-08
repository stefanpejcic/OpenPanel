// Package websites (this file) fetches site screenshots from a remote screenshot API rather than rendering them locally - there's no local headless-Chromium pipeline here, so a `screenshots` config value that isn't a remote URL falls back to OpenPanel's own hosted API. The API returns a base64 PNG, decoded once and cached to disk under screenshotCacheDir; that cached file is served on every view (indefinitely, no TTL) until the user regenerates it via the "screenshot" partial's refresh button.
package websites

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

const (
	screenshotCacheDir     = "/etc/openpanel/wordpress/screenshots"
	screenshotFallbackAPI  = "https://api.openpanel.com/screenshots"
	screenshotFetchTimeout = 5 * time.Second
)

// screenshotAPIBase resolves the `screenshots` config value: a configured, non-"local" value is used as-is; anything else (unset or "local", the old local-Playwright selector) falls back to OpenPanel's hosted API.
func screenshotAPIBase(a *appctx.App) string {
	setting := strings.TrimSpace(a.Config.Get("screenshots", ""))
	if setting != "" && setting != "local" {
		return strings.TrimRight(setting, "/")
	}
	return screenshotFallbackAPI
}

// screenshotCachePath is the MD5-hash-named cache file for a domain, keyed on the raw domain (+ optional /subfolder) path segment rather than a scheme-qualified URL, since the remote API decides how to fetch the page, not this process.
func screenshotCachePath(domain string) string {
	sum := md5.Sum([]byte(domain)) //nolint:gosec // content-addressed cache key, not a security boundary
	return filepath.Join(screenshotCacheDir, hex.EncodeToString(sum[:])+".png")
}

// fetchAndCacheScreenshot fetches the API's {"base64": "..."} envelope, decodes it, and writes the PNG to the local cache.
func fetchAndCacheScreenshot(ctx context.Context, a *appctx.App, domain string) error {
	apiURL := screenshotAPIBase(a) + "/" + domain

	fetchCtx, cancel := context.WithTimeout(ctx, screenshotFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("screenshot API returned status %d", resp.StatusCode)
	}

	var body struct {
		Base64 string `json:"base64"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	imgBytes, err := base64.StdEncoding.DecodeString(body.Base64)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(screenshotCacheDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(screenshotCachePath(domain), imgBytes, 0o644)
}

// TriggerScreenshotGeneration kicks off screenshot generation for a freshly installed site in the background and returns immediately - every CMS/app install.go calls this right after its own "INSERT INTO sites" succeeds, so the screenshot is already cached by the time the install-complete redirect lands the user on /website. Uses a detached context (5s cap) instead of the request's own, since the install's HTTP response has already been sent by the time that would get cancelled - a slow/unreachable API just falls back to the page's own on-demand fetch.
func TriggerScreenshotGeneration(a *appctx.App, domain string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := fetchAndCacheScreenshot(ctx, a, domain); err != nil {
			log.Printf("WEBSITES - background screenshot generation for %s failed: %v", domain, err)
		}
	}()
}

// handleScreenshot: GET serves the cached screenshot (generating it first on a cache miss), POST always regenerates it - the screenshot partial's refresh button POSTs then re-fetches the GET URL with a cache-busting query string.
func handleScreenshot(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.PathValue("domain")
	apexDomain, _ := splitDomainAndFolder(domain)
	if !a.CheckDomainBelongsToUser(ctx, userID, apexDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	cachePath := screenshotCachePath(domain)

	if r.Method == http.MethodPost {
		if err := fetchAndCacheScreenshot(ctx, a, domain); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		http.ServeFile(w, r, cachePath)
		return
	}

	if _, statErr := os.Stat(cachePath); statErr != nil {
		if err := fetchAndCacheScreenshot(ctx, a, domain); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	http.ServeFile(w, r, cachePath)
}
