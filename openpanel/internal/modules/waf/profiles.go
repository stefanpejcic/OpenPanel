package waf

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// crsPluginsDir is where `opencli waf enable` clones the official CRS rule exclusion plugins, swapped in tests
var crsPluginsDir = "/etc/openpanel/caddy/coreruleset-plugins/"

// wafProfile is one app profile, backed by the official coreruleset/<Key>-rule-exclusions-plugin
type wafProfile struct {
	Key         string
	Name        string
	Initials    string
	Color       string
	Description string
	// sites.type values that get this profile recommended
	SiteTypes []string
}

var profileCatalog = []wafProfile{
	{Key: "wordpress", Name: "WordPress", Initials: "WP", Color: "#21759b", SiteTypes: []string{"wordpress"},
		Description: "Stops false blocks when editing posts, saving settings, using the block editor or logging in."},
	{Key: "drupal", Name: "Drupal", Initials: "Dr", Color: "#0678be", SiteTypes: []string{"drupal"},
		Description: "Stops false blocks when editing content, saving forms and changing settings in the Drupal admin."},
	{Key: "nextcloud", Name: "Nextcloud", Initials: "NC", Color: "#0082c9", SiteTypes: []string{"nextcloud"},
		Description: "Stops false blocks on file uploads, WebDAV sync, calendars, contacts and document editing."},
	{Key: "dokuwiki", Name: "DokuWiki", Initials: "DW", Color: "#7e8a35", SiteTypes: []string{"dokuwiki"},
		Description: "Stops false blocks when editing and saving wiki pages."},
	{Key: "phpbb", Name: "phpBB", Initials: "BB", Color: "#105289", SiteTypes: []string{"phpbb"},
		Description: "Stops false blocks when posting, sending private messages and using the admin panel."},
	{Key: "xenforo", Name: "XenForo", Initials: "XF", Color: "#2577b1",
		Description: "Stops false blocks when posting, using the editor and managing the forum."},
	{Key: "phpmyadmin", Name: "phpMyAdmin", Initials: "PMA", Color: "#6c78af",
		Description: "Lets you run SQL queries and import databases from a phpMyAdmin installed on this domain."},
}

func profileByKey(key string) (wafProfile, bool) {
	for _, p := range profileCatalog {
		if p.Key == key {
			return p, true
		}
	}
	return wafProfile{}, false
}

func pluginFile(key, stage string) string {
	return crsPluginsDir + key + "-rule-exclusions-plugin/plugins/" + key + "-rule-exclusions-" + stage + ".conf"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// availableProfiles is the catalog filtered to plugins actually installed on the server
func availableProfiles() []wafProfile {
	var out []wafProfile
	for _, p := range profileCatalog {
		if fileExists(pluginFile(p.Key, "before")) {
			out = append(out, p)
		}
	}
	return out
}

var pluginIncludeRE = regexp.MustCompile(`/([a-z0-9]+)-rule-exclusions-plugin/plugins/`)

// parseWAFProfiles returns the profile keys a domain conf currently includes, in catalog order
func parseWAFProfiles(contentStr string) []string {
	seen := map[string]bool{}
	for _, line := range strings.Split(contentStr, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "Include") {
			continue
		}
		if m := pluginIncludeRE.FindStringSubmatch(line); m != nil {
			seen[m[1]] = true
		}
	}
	var out []string
	for _, p := range profileCatalog {
		if seen[p.Key] {
			out = append(out, p.Key)
		}
	}
	return out
}

var (
	errUnknownProfile = errors.New("unknown profile")
	errNoRulesInclude = errors.New("WAF config for this domain has no CRS rules include")
	errProfileReload  = errors.New("the web server rejected the new WAF config, nothing was changed")
)

// rewriteProfiles swaps the level SecAction and plugin Include lines in every directives block: level, config and before go right above the CRS rules include, after goes right below it; an empty level keeps the current line as-is
func rewriteProfiles(content string, keys []string, level string) (string, error) {
	var before, after []string
	if level == "" {
		for _, line := range strings.Split(content, "\n") {
			if s := strings.TrimSpace(line); isLevelLine(s) {
				before = append(before, s)
				break
			}
		}
	} else if l, ok := levelByKey(level); ok && l.Key != defaultLevel {
		before = append(before, levelLine(l))
	}
	for _, key := range keys {
		for _, stage := range []string{"config", "before"} {
			if f := pluginFile(key, stage); fileExists(f) {
				before = append(before, "Include "+f)
			}
		}
		if f := pluginFile(key, "after"); fileExists(f) {
			after = append(after, "Include "+f)
		}
	}

	lines := readLinesKeepEnds(content)
	var out []string
	insideDirectives, foundRules := false, false

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		switch {
		case !insideDirectives && strings.HasPrefix(stripped, "directives `"):
			insideDirectives = true
			out = append(out, line)
		case insideDirectives && stripped == "`":
			insideDirectives = false
			out = append(out, line)
		case insideDirectives && (isLevelLine(stripped) || strings.HasPrefix(stripped, "Include") && pluginIncludeRE.MatchString(stripped)):
			// dropped, re-added below around the rules include
		case insideDirectives && strings.HasPrefix(stripped, "Include") && strings.Contains(stripped, "/coreruleset/rules/"):
			foundRules = true
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			for _, d := range before {
				out = append(out, indent+d+"\n")
			}
			out = append(out, line)
			for _, d := range after {
				out = append(out, indent+d+"\n")
			}
		default:
			out = append(out, line)
		}
	}
	if !foundRules {
		return content, errNoRulesInclude
	}
	return strings.Join(out, ""), nil
}

// detectedSite is an installed app on the domain that maps to a profile
type detectedSite struct {
	SiteName string
	Type     string
}

// detectProfiles maps the domain's sites (root and subfolders) to profile keys
func detectProfiles(ctx context.Context, a *appctx.App, domain string) map[string][]string {
	out := map[string][]string{}
	if a == nil || a.DB == nil {
		return out
	}
	rows, err := a.DB.QueryContext(ctx, `
		SELECT sites.site_name, sites.type FROM sites
		JOIN domains ON domains.domain_id = sites.domain_id
		WHERE domains.domain_url = ?`, domain)
	if err != nil {
		return out
	}
	defer rows.Close()
	var sites []detectedSite
	for rows.Next() {
		var s detectedSite
		if rows.Scan(&s.SiteName, &s.Type) == nil {
			sites = append(sites, s)
		}
	}
	return profilesForSites(sites)
}

// detectProfilesForUser is detectProfiles for every domain the user owns in one query, keyed by domain
func detectProfilesForUser(ctx context.Context, a *appctx.App, userID int) map[string]map[string][]string {
	out := map[string]map[string][]string{}
	if a == nil || a.DB == nil {
		return out
	}
	rows, err := a.DB.QueryContext(ctx, `
		SELECT domains.domain_url, sites.site_name, sites.type FROM sites
		JOIN domains ON domains.domain_id = sites.domain_id
		WHERE domains.user_id = ?`, userID)
	if err != nil {
		return out
	}
	defer rows.Close()
	byDomain := map[string][]detectedSite{}
	for rows.Next() {
		var domain string
		var s detectedSite
		if rows.Scan(&domain, &s.SiteName, &s.Type) == nil {
			byDomain[domain] = append(byDomain[domain], s)
		}
	}
	for domain, sites := range byDomain {
		out[domain] = profilesForSites(sites)
	}
	return out
}

// recommendedNames is the installed, not yet active profiles detected for a domain, as display names
func recommendedNames(active []string, detected map[string][]string) []string {
	_, suggested := buildProfileViews(active, detected)
	var names []string
	for _, p := range suggested {
		names = append(names, p.Name)
	}
	return names
}

func profilesForSites(sites []detectedSite) map[string][]string {
	out := map[string][]string{}
	for _, s := range sites {
		typ := strings.ToLower(strings.TrimSpace(s.Type))
		for _, p := range profileCatalog {
			for _, t := range p.SiteTypes {
				if t == typ {
					out[p.Key] = append(out[p.Key], s.SiteName)
				}
			}
		}
	}
	return out
}

// setDomainProfiles validates profile keys and level and rewrites the domain conf, an empty level keeps the current one
func setDomainProfiles(ctx context.Context, domain string, keys []string, level string) ([]string, string, error) {
	level = strings.ToLower(strings.TrimSpace(level))
	if _, ok := levelByKey(level); level != "" && !ok {
		return nil, "", errUnknownLevel
	}

	path := domainConfigPath(domain)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", errNoDomainConf
	}
	installed := map[string]bool{}
	for _, p := range availableProfiles() {
		installed[p.Key] = true
	}
	var clean []string
	seen := map[string]bool{}
	for _, k := range keys {
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" || seen[k] {
			continue
		}
		if !installed[k] {
			return nil, "", fmt.Errorf("%w: %s", errUnknownProfile, k)
		}
		seen[k] = true
		clean = append(clean, k)
	}

	newContent, err := rewriteProfiles(string(content), clean, level)
	if err != nil {
		return nil, "", err
	}
	if newContent != string(content) {
		if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
			return nil, "", err
		}
		if err := reloadCaddy(ctx); err != nil {
			// put the old conf back so a bad plugin file can't leave caddy unable to reload
			_ = os.WriteFile(path, content, 0o644)
			_ = reloadCaddy(ctx)
			return nil, "", errProfileReload
		}
	}
	return parseWAFProfiles(newContent), parseWAFLevel(newContent), nil
}

func currentProfiles(domain string) []string {
	content, err := os.ReadFile(domainConfigPath(domain))
	if err != nil {
		return nil
	}
	return parseWAFProfiles(string(content))
}

// profilesFromRequest reads the repeated "profiles" field, the page sends FormData so it's multipart and ParseForm alone would miss it
func profilesFromRequest(r *http.Request) []string {
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		_ = r.ParseForm()
	}
	return r.Form["profiles"]
}

// handleWAFProfiles handles POST /server/waf/profiles/{domain}, the profile toggles on the domain's WAF page
func handleWAFProfiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domain := firstPathSegment(r.PathValue("domain"))
	userID, _ := auth.UserID(r)
	username, err := injected(a, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, web.Tr(a, r, "internal error"))
		return
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeJSONError(w, http.StatusForbidden, web.Tr(a, r, "You do not own this domain."))
		return
	}
	profiles := profilesFromRequest(r)
	// bulk actions on /server/waf work on top of what the domain already has
	switch r.Form.Get("mode") {
	case "level":
		profiles = currentProfiles(domain)
	case "add":
		profiles = append(currentProfiles(domain), r.Form.Get("profile"))
	case "clear":
		profiles = nil
	}
	active, level, err := setDomainProfiles(r.Context(), domain, profiles, r.Form.Get("level"))
	switch {
	case errors.Is(err, errNoDomainConf):
		writeJSONError(w, http.StatusNotFound, web.Tr(a, r, err.Error()))
		return
	case errors.Is(err, errUnknownProfile):
		writeJSONError(w, http.StatusBadRequest, web.Tr(a, r, "Unknown profile: %(profile)s", "profile", strings.TrimPrefix(err.Error(), errUnknownProfile.Error()+": ")))
		return
	case err != nil:
		log.Printf("WAF - setting profiles for %s: %v", domain, err)
		writeJSONError(w, http.StatusBadRequest, web.Tr(a, r, err.Error()))
		return
	}
	label := "none"
	if len(active) > 0 {
		label = strings.Join(active, ", ")
	}
	_ = logger.RecordUserAction(a.Config, username, "set WAF level "+level+" and profiles for domain "+domain+": "+label, reqip.ClientIP(r))
	if active == nil {
		active = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"domain": domain, "profiles": active, "level": level})
}
