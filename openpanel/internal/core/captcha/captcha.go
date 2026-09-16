// Package captcha implements the login page's CAPTCHA: resolving the
// configured provider's widget info and verifying a submitted response
// token against Google reCAPTCHA or Cloudflare Turnstile. This is built
// directly into the panel - not an external plugin - so it needs no
// separate install step and nothing to exec.
package captcha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// Widget is what the login page needs to render the active provider's widget - Provider is "" when none is configured, which the caller treats as "don't show a captcha".
type Widget struct {
	Provider  string
	FieldName string
	SiteKey   string
}

// GetWidget resolves the configured captcha provider (openpanel.config's captcha_provider) and its public site key.
func GetWidget(cfg config.Config) Widget {
	switch cfg.Get("captcha_provider", "") {
	case "google":
		return Widget{Provider: "google", FieldName: "g-recaptcha-response", SiteKey: cfg.Get("recaptcha_site_key", "")}
	case "turnstile":
		return Widget{Provider: "turnstile", FieldName: "cf-turnstile-response", SiteKey: cfg.Get("turnstile_site_key", "")}
	case "custom":
		return Widget{Provider: "custom", FieldName: "custom-captcha", SiteKey: cfg.Get("custom_captcha_site_key", "")}
	default:
		return Widget{}
	}
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// Verify checks a submitted response token against the configured provider. Returns true when no provider is configured - nothing to enforce - or, for google/turnstile, false on a network/provider failure (fail closed, matching each provider's own siteverify contract).
func Verify(ctx context.Context, cfg config.Config, token, ip string) bool {
	switch cfg.Get("captcha_provider", "") {
	case "google":
		if token == "" {
			return false
		}
		return verifySiteverify(ctx, "https://www.google.com/recaptcha/api/siteverify", cfg.Get("recaptcha_secret_key", ""), token, ip)
	case "turnstile":
		if token == "" {
			return false
		}
		return verifySiteverify(ctx, "https://challenges.cloudflare.com/turnstile/v0/siteverify", cfg.Get("turnstile_secret_key", ""), token, "")
	case "custom":
		if token == "" {
			return false
		}
		return verifyCustom(token)
	default:
		return true
	}
}

// verifySiteverify posts to a Google/Cloudflare-style siteverify endpoint (shared response shape between the two providers).
func verifySiteverify(ctx context.Context, endpoint, secret, token, ip string) bool {
	form := url.Values{"secret": {secret}, "response": {token}}
	if ip != "" {
		form.Set("remoteip", ip)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}
	return result.Success
}

// verifyCustom is a placeholder for operators who want their own CAPTCHA provider - replace this with your own verification logic.
func verifyCustom(token string) bool {
	return true
}
