package waf

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// DomainWAFRow is the extra per-domain data shown in the /server/waf table
type DomainWAFRow struct {
	Status   string
	Level    string
	Profiles []string
	Blocks   int
	// profiles matching apps installed on the domain that aren't on yet
	Recommended []string
	// filled dots under the level name, 1 compatibility, 2 standard, 3 strict, 0 custom
	LevelStrength int
	LevelDots     []bool
	// untranslated keys for the status/level/profile filters
	LevelKey    string
	ProfileKeys []string
}

func domainWAFRow(domain string, detected map[string][]string) DomainWAFRow {
	row := DomainWAFRow{Status: "Not Found", Level: defaultLevel}
	content, err := os.ReadFile(domainConfigPath(domain))
	if err != nil {
		return row
	}
	c := string(content)
	row.Status = engineStatus(c)
	row.Level = parseWAFLevel(c)
	active := parseWAFProfiles(c)
	row.ProfileKeys = active
	row.Recommended = recommendedNames(active, detected)
	for _, key := range active {
		if p, ok := profileByKey(key); ok {
			row.Profiles = append(row.Profiles, p.Name)
		}
	}
	row.Blocks = readWAFLogs(wafLogPath(domain), 86400).Blocks
	return row
}

// setWAFForDomains sets SecRuleEngine on every given domain conf with a single caddy reload at the end, monitor-only domains included; returns how many files changed and which failed
func setWAFForDomains(ctx context.Context, domains []string, enabled bool) (changed int, failed []string, err error) {
	to := "Off"
	if enabled {
		to = "On"
	}
	for _, d := range domains {
		path := domainConfigPath(d)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			failed = append(failed, d)
			continue
		}
		updated := setEngine(string(content), to)
		if updated == string(content) {
			continue
		}
		if writeErr := os.WriteFile(path, []byte(updated), 0o644); writeErr != nil {
			failed = append(failed, d)
			continue
		}
		changed++
	}
	if changed > 0 {
		err = reloadCaddy(ctx)
	}
	return changed, failed, err
}

// handleWAFAll handles POST /server/waf/all, the "Firewall for all domains" switch: applies to every domain the user owns and saves it as the default for new ones
func handleWAFAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	action := r.Form.Get("action")
	if action != "enable" && action != "disable" {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Invalid action."), "/server/waf")
		return
	}
	enabled := action == "enable"

	if err := SetAccountWAFEnabled(userContext, enabled); err != nil {
		log.Printf("WAF - saving default for new domains: %v", err)
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Could not save the default for new domains."), "/server/waf")
		return
	}

	userID, _ := auth.UserID(r)
	domains, _ := a.AllDomainsForUser(r.Context(), userID)
	names := make([]string, 0, len(domains))
	for _, d := range domains {
		names = append(names, d.DomainURL)
	}
	_, failed, reloadErr := setWAFForDomains(r.Context(), names, enabled)

	verb := "disabled"
	if enabled {
		verb = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, username, verb+" WAF for all domains and new domains", reqip.ClientIP(r))

	switch {
	case reloadErr != nil:
		log.Printf("WAF - reload after changing all domains: %v", reloadErr)
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Settings saved but reloading the web server failed."), "/server/waf")
	case len(failed) > 0:
		web.FlashRedirect(a, w, r, "warning", web.Tr(a, r, "Could not update: %(domains)s", "domains", strings.Join(failed, ", ")), "/server/waf")
	case enabled:
		web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "The firewall is now on for all your domains and for new domains."), "/server/waf")
	default:
		web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "The firewall is now off for all your domains and for new domains."), "/server/waf")
	}
}
