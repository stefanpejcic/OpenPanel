// Package websites (this file) fetches site screenshots from a remote screenshot API rather than rendering them locally - there's no local headless-Chromium pipeline here, so a `screenshots` config value that isn't a remote URL falls back to OpenPanel's own hosted API. The API returns a base64 PNG, decoded once and cached to disk under screenshotCacheDir; that cached file is served on every view (indefinitely, no TTL) until the user regenerates it via the "screenshot" partial's refresh button.
package websites

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"golang.org/x/net/idna"
)

const (
	screenshotCacheDir     = "/etc/openpanel/wordpress/screenshots"
	screenshotFallbackAPI  = "https://api.openpanel.com/screenshots"
	screenshotFetchTimeout = 25 * time.Second // heavy sites take 10-15s on the API
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

// errBlankScreenshot is a capture that came back one flat colour, the API answers like that when it couldn't load the site
var errBlankScreenshot = errors.New("the screenshot came back blank, the site may not be reachable from the internet")

// blankScreenshot reports whether an image is (almost) one colour, a real page always has some text or edges
func blankScreenshot(data []byte) bool {
	// a real 1280x720 page never compresses this small, so bigger files skip the decode
	if len(data) > 64<<10 {
		return false
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}
	b := img.Bounds()
	if b.Empty() {
		return true
	}
	r0, g0, b0, _ := img.At(b.Min.X, b.Min.Y).RGBA()
	near := func(a, b uint32) bool { return a-b < 0x0a00 || b-a < 0x0a00 }
	samples, off := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y += 4 {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			r, g, bl, _ := img.At(x, y).RGBA()
			samples++
			if !near(r, r0) || !near(g, g0) || !near(bl, b0) {
				off++
			}
		}
	}
	// a few stray pixels (a cursor, a 1px border) still count as blank
	return off*1000 <= samples
}

// fetchAndCacheScreenshot fetches the API's {"base64": "..."} envelope, decodes it, and writes the PNG to the local cache.
func fetchAndCacheScreenshot(ctx context.Context, a *appctx.App, domain string) error {
	apiURL := screenshotAPIBase(a) + "/" + domain
	// the API loads the site from this server's IP, so it works before DNS points here, older APIs ignore it
	if ip := screenshotServerIP(ctx, a, domain); ip != "" {
		apiURL += "?ip=" + url.QueryEscape(ip)
	}

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

	if err := tenantfs.MkdirAll(screenshotCacheDir, 0o755); err != nil {
		return err
	}
	// a blank capture is cached too, so an unreachable site doesn't call the API on every view
	if err := tenantfs.WriteFile(screenshotCachePath(domain), imgBytes, 0o644); err != nil {
		return err
	}
	if blankScreenshot(imgBytes) {
		return errBlankScreenshot
	}
	return nil
}

// cachedScreenshotBlank reads the cached file back and checks it
func cachedScreenshotBlank(path string) bool {
	data, err := tenantfs.ReadFile(path)
	return err != nil || blankScreenshot(data)
}

// screenshotServerIP is the IP the domain's owner is served from, "" when the owner can't be found
func screenshotServerIP(ctx context.Context, a *appctx.App, domain string) string {
	host, _, _ := strings.Cut(domain, "/")
	host, err := idna.ToASCII(host)
	if err != nil {
		return ""
	}
	var username string
	if err := a.DB.QueryRowContext(ctx, "SELECT u.username FROM domains d JOIN users u ON u.id = d.user_id WHERE d.domain_url = ?", host).Scan(&username); err != nil {
		return ""
	}
	return a.GetCachedIPForUserOrPublicIPv4(ctx, username)
}

// TriggerScreenshotGeneration kicks off screenshot generation for a freshly installed site in the background and returns immediately - every CMS/app install.go calls this right after its own "INSERT INTO sites" succeeds, so the screenshot is already cached by the time the install-complete redirect lands the user on /website. Uses a detached context (screenshotFetchTimeout cap) instead of the request's own, since the install's HTTP response has already been sent by the time that would get cancelled - a slow/unreachable API just falls back to the page's own on-demand fetch.
func TriggerScreenshotGeneration(a *appctx.App, domain string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), screenshotFetchTimeout)
		defer cancel()
		if err := fetchAndCacheScreenshot(ctx, a, domain); err != nil {
			log.Printf("WEBSITES - background screenshot generation for %s failed: %v", domain, err)
		}
	}()
}

// handleScreenshot: GET serves the cached screenshot (generating it first on a cache miss), POST always regenerates it - the screenshot partial's refresh button POSTs then re-fetches the GET URL with a cache-busting query string.
func handleScreenshot(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := auth.Injected(a, r)
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

	if _, statErr := tenantfs.Stat(cachePath); statErr != nil {
		if err := fetchAndCacheScreenshot(ctx, a, domain); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}
	// the page shows "No screenshot" instead of a white box
	if cachedScreenshotBlank(cachePath) {
		http.Error(w, errBlankScreenshot.Error(), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, cachePath)
}
