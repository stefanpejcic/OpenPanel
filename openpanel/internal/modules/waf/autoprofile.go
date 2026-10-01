package waf

import (
	"context"
	"log"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// profileKeyForSiteType maps a sites.type value to its profile key, "" when there's no profile for that app
func profileKeyForSiteType(siteType string) string {
	siteType = strings.ToLower(strings.TrimSpace(siteType))
	for _, p := range profileCatalog {
		for _, t := range p.SiteTypes {
			if t == siteType {
				return p.Key
			}
		}
	}
	return ""
}

// ownerAllowsWAF is true when the domain's owner has the waf module on their plan, so they can see and change the profile later
var ownerAllowsWAF = func(ctx context.Context, a *appctx.App, domain string) bool {
	if a == nil || a.DB == nil {
		return false
	}
	var userID int
	if err := a.DB.QueryRowContext(ctx, "SELECT user_id FROM domains WHERE domain_url = ?", domain).Scan(&userID); err != nil {
		return false
	}
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		return false
	}
	list, _ := data["user_allowed"].([]string)
	return containsString(list, "waf")
}

// EnableProfileForNewSite turns on the firewall profile for an app right after it's installed or imported, the level and other profiles stay as they are
func EnableProfileForNewSite(a *appctx.App, siteName, siteType string) {
	key := profileKeyForSiteType(siteType)
	if key == "" || !fileExists(pluginFile(key, "before")) {
		return
	}
	domain := firstPathSegment(siteName)
	// installers stream their response, so don't hang this on the request context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !ownerAllowsWAF(ctx, a, domain) {
		return
	}
	current := currentProfiles(domain)
	if containsString(current, key) {
		return
	}
	if _, _, err := setDomainProfiles(ctx, domain, append(current, key), ""); err != nil {
		log.Printf("WAF - auto profile %s for %s: %v", key, domain, err)
	}
}
