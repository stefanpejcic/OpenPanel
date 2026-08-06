// Package websites (this file) fetches site screenshots from a remote
// screenshot API rather than rendering them locally - there's no local
// headless-Chromium rendering pipeline here, so a `screenshots` config
// value that isn't a remote URL (unset or "local") falls back to
// OpenPanel's own hosted API (api.openpanel.com). The API returns a
// base64 PNG in a JSON envelope, which is decoded once and cached to disk
// under screenshotCacheDir; that cached file is served on every view
// (indefinitely, no TTL) until the user explicitly regenerates it via the
// POST action the "screenshot" partial's refresh button already fires.
package websites

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	screenshotFetchTimeout = 30 * time.Second
)

// screenshotAPIBase resolves the `screenshots` config value: a configured,
// non-"local" value is used as-is; anything else (unset or "local", the
// value that used to select a local-Playwright rendering path before that
// pipeline existed here) falls back to OpenPanel's hosted API.
func screenshotAPIBase(a *appctx.App) string {
	setting := strings.TrimSpace(a.Config.Get("screenshots", ""))
	if setting != "" && setting != "local" {
		return strings.TrimRight(setting, "/")
	}
	return screenshotFallbackAPI
}

// screenshotCachePath is the MD5-hash-named cache file for a domain,
// keyed on the raw domain (+ optional /subfolder) path segment rather than
// a scheme-qualified URL - the remote API, not this process, decides how
// to fetch the page.
func screenshotCachePath(domain string) string {
	sum := md5.Sum([]byte(domain)) //nolint:gosec // content-addressed cache key, not a security boundary
	return filepath.Join(screenshotCacheDir, hex.EncodeToString(sum[:])+".png")
}

// fetchAndCacheScreenshot fetches the API's {"base64": "..."} envelope,
// decodes it, and writes the PNG to the local cache.
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

// handleScreenshot: GET serves the cached screenshot (generating it first
// on a cache miss), POST always regenerates it - the screenshot partial's
// refresh button POSTs then re-fetches the GET URL with a cache-busting
// query string.
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
