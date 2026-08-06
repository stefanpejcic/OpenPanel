// Package websites (this file) implements the "Live Preview" button's
// temporary .openpanel.org-style preview link.
package websites

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

const temporaryLinkFallback = "https://preview.openpanel.org/index.php"

// temporaryLinkSetting reads the `temporary_links` config value, memoized
// for 2 hours since it's rarely changed.
func temporaryLinkSetting(ctx context.Context, a *appctx.App) string {
	setting, _ := cache.Memoize(ctx, a.Cache, "temporary_links_setting", 2*time.Hour, func() (string, error) {
		return a.Config.Get("temporary_links", ""), nil
	})
	return setting
}

// handleTemporaryLink requests a preview link from the temporary-links
// service for the given domain. The response is memoized for 800s (~13m)
// per user+domain since the upstream link doesn't change that often and
// the request is relatively expensive.
func handleTemporaryLink(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Domain name not provided.", http.StatusBadRequest)
		return
	}

	apexDomain, _ := splitDomainAndFolder(domain)
	if !a.CheckDomainBelongsToUser(ctx, userID, apexDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	cacheKey := "temporary_link:" + currentUsername + ":" + domain
	result, _ := cache.Memoize(ctx, a.Cache, cacheKey, 800*time.Second, func() (map[string]any, error) {
		serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
		link := temporaryLinkSetting(ctx, a)
		if link == "" {
			link = temporaryLinkFallback
		}

		form := url.Values{"domain": {domain}, "ip": {serverIP}}
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, link, strings.NewReader(form.Encode()))
		if reqErr != nil {
			return map[string]any{"error": reqErr.Error()}, nil
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, doErr := client.Do(req)
		if doErr != nil {
			return map[string]any{"error": doErr.Error()}, nil
		}
		defer resp.Body.Close()

		if loc := resp.Header.Get("Location"); loc != "" {
			return map[string]any{"link": loc}, nil
		}

		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		return map[string]any{
			"error": "No redirect link found", "url": link, "sent_data": map[string]string{"domain": domain, "ip": serverIP},
			"response_status_code": resp.StatusCode, "response_text": string(body[:n]),
		}, nil
	})

	status := http.StatusOK
	if _, isError := result["error"]; isError {
		if _, hasStatusCode := result["response_status_code"]; hasStatusCode {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}
	}
	writeJSON(w, status, result)
}
