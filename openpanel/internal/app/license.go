package app

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

// licenseTagRegex matches <tag>value</tag> pairs in the license server's response - Go's RE2 has no backreferences, so the closing tag is checked manually in parseLicenseResponse
var licenseTagRegex = regexp.MustCompile(`<([^<>]+)>([^<]+)</([^<>]+)>`)

func parseLicenseResponse(body string) map[string]string {
	result := map[string]string{}
	for _, m := range licenseTagRegex.FindAllStringSubmatch(body, -1) {
		tag, value, closeTag := m[1], m[2], m[3]
		if tag == closeTag {
			result[tag] = value
		}
	}
	return result
}

// checkLicenseStartup does one remote license check at boot; network/parse failures just count as "not valid"
func (a *App) checkLicenseStartup(ctx context.Context) bool {
	if a.LicenseKey == "" {
		return false
	}

	dynamicIP := sysinfo.FetchPublicIP(ctx, a.Cache)
	log.Printf("BOOTSTRAP - validating license key: %s", a.LicenseKey)

	form := url.Values{"licensekey": {a.LicenseKey}, "ip": {dynamicIP}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openpanel.com/enterprise/index.php", strings.NewReader(form.Encode()))
	if err != nil {
		log.Printf("BOOTSTRAP - license validation failed: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "OpenAdmin-License-Check/1.0")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("BOOTSTRAP - license validation failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("BOOTSTRAP - license validation failed: %v", err)
		return false
	}

	data := parseLicenseResponse(string(body))
	if data["status"] == "Active" {
		log.Print("BOOTSTRAP - license key is 'Active'")
	}
	return data["status"] == "Active"
}

// LicenseCheckPasses reports whether the current license allows access - no key or an enterprise/noc/lifetime key always passes, anything else falls back to the startup-computed LicenseValid
func (a *App) LicenseCheckPasses() bool {
	if a.LicenseKey == "" {
		return true
	}
	if strings.HasPrefix(a.LicenseKey, "enterprise") || strings.HasPrefix(a.LicenseKey, "noc") || strings.HasPrefix(a.LicenseKey, "lifetime") {
		return true
	}
	return a.LicenseValid
}
