package waf

import (
	"context"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var ruleIDRE = regexp.MustCompile(`^\d+$`)

// tagRE matches a CRS tag such as attack-sqli or paranoia-level/1, nothing that could end the quoted SecLang argument or the Caddyfile backtick token
var tagRE = regexp.MustCompile(`^[A-Za-z0-9._/:-]{1,128}$`)

// writeConfAndReload writes a domain conf and reloads the shared caddy, a conf caddy rejects is put back so it can't break reloads for everyone
func writeConfAndReload(ctx context.Context, path string, old []byte, newContent string) error {
	if err := tenantfs.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return err
	}
	if err := reloadCaddy(ctx); err != nil {
		_ = tenantfs.WriteFile(path, old, 0o644)
		_ = reloadCaddy(ctx)
		return errCaddyReload
	}
	return nil
}

// reloadCaddy mirrors the `podman exec caddy caddy reload --config /etc/caddy/Caddyfile` call every WAF config change makes to apply it
var reloadCaddy = func(ctx context.Context) error {
	return exec.CommandContext(ctx, "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run()
}

func filterOut(items []string, exclude string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != exclude {
			out = append(out, item)
		}
	}
	return out
}

func handleWAFDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domainName := firstPathSegment(r.PathValue("domain"))

	userID, _ := auth.UserID(r)
	_, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainName) {
		log.Printf("WAF - Domain %s is not owned by the user.", domainName)
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	configFilePath := domainConfigPath(domainName)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		removedRules := strings.Fields(r.Form.Get("removed_rules"))
		removedTags := strings.Fields(r.Form.Get("removed_tags"))

		for _, rule := range removedRules {
			if !ruleIDRE.MatchString(rule) {
				web.FlashRedirect(a, w, r, "error", "Error: IDs should only contain numbers separated by spaces.", "/server/waf/"+domainName)
				return
			}
			// the single-rule path (disableRuleForDomain) refuses these; enforce it here too, or a
			// tenant could remove the CRS anomaly-scoring rules and turn off blocking for the domain
			// while the WAF still shows as enabled.
			if isScoringRule(rule) {
				web.FlashRedirect(a, w, r, "error", errScoringRule.Error(), "/server/waf/"+domainName)
				return
			}
		}
		for _, tag := range removedTags {
			if !tagRE.MatchString(tag) {
				web.FlashRedirect(a, w, r, "error", "Error: tags can only contain letters, numbers and . _ / : -", "/server/waf/"+domainName)
				return
			}
		}

		removedRules = append([]string{excludedRuleID}, filterOut(removedRules, excludedRuleID)...)
		removedTags = append([]string{excludedTag}, filterOut(removedTags, excludedTag)...)

		if content, readErr := tenantfs.ReadFile(configFilePath); readErr == nil {
			newContent := rewriteDirectivesBlock(string(content), removedRules, removedTags)
			if writeErr := writeConfAndReload(r.Context(), configFilePath, content, newContent); writeErr == nil {
				_ = logger.RecordUserAction(a.Config, username, "updated WAF rules for domain "+domainName, reqip.ClientIP(r))
				web.Flash(a, w, r, "success", "WAF configuration updated.")
			} else {
				log.Printf("WAF - Error occurred during updating WAF settings: %v", writeErr)
				web.Flash(a, w, r, "error", "Error updating WAF!")
			}
		} else {
			log.Printf("WAF - Configuration file: %s does not exist.", configFilePath)
			web.Flash(a, w, r, "error", "Configuration file not found.")
		}
	}

	// GET (and the POST fallthrough): re-read the file after any change.
	status := "Not Found"
	var removedRules, removedTags, activeProfiles []string
	level := defaultLevel

	if content, readErr := tenantfs.ReadFile(configFilePath); readErr == nil {
		contentStr := string(content)
		status = engineStatus(contentStr)

		removedRules, removedTags = parseWAFRemovals(contentStr)
		activeProfiles = parseWAFProfiles(contentStr)
		level = parseWAFLevel(contentStr)
	}

	detected := detectProfiles(r.Context(), a, domainName)

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"domain": domainName, "status": status, "removed_rules": removedRules, "removed_tags": removedTags,
			"profiles": activeProfiles, "detected_profiles": detected, "level": level,
		})
		return
	}

	profiles, suggested := buildProfileViews(activeProfiles, detected)
	renderWAFDomainPage(a, w, r, WAFDomainPageData{
		Domain: domainName, Status: status, RemovedRules: removedRules, RemovedTags: removedTags,
		Profiles: profiles, Suggested: suggested, Stats: readWAFLogs(wafLogPath(domainName), 86400), Level: level,
		Summary: recentSummary(domainName),
	})
}

// readLinesKeepEnds splits content into lines; each element (except possibly the last) keeps its trailing "\n"
func readLinesKeepEnds(content string) []string {
	if content == "" {
		return nil
	}
	var lines []string
	start := 0
	for i, c := range content {
		if c == '\n' {
			lines = append(lines, content[start:i+1])
			start = i + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

// rewriteDirectivesBlock mirrors the line-buffering loop in server_settings_waf_for_domain(): finds the directives block in a Caddy per-domain conf, strips existing SecRuleRemoveById/SecRuleRemoveByTag lines, and appends fresh ones built from removedRules/removedTags
func rewriteDirectivesBlock(content string, removedRules, removedTags []string) string {
	lines := readLinesKeepEnds(content)
	var newLines []string
	insideDirectives := false
	var buffer []string

	for _, line := range lines {
		stripped := strings.TrimSpace(line)

		if !insideDirectives && strings.HasPrefix(stripped, "directives `") {
			insideDirectives = true
			buffer = []string{line}
			continue
		}

		if insideDirectives {
			if stripped == "`" {
				var filtered []string
				for _, l := range buffer {
					s := strings.TrimSpace(l)
					if strings.HasPrefix(s, "SecRuleRemoveById") || strings.HasPrefix(s, "SecRuleRemoveByTag") {
						continue
					}
					filtered = append(filtered, l)
				}
				filtered = append(filtered, "            SecRuleRemoveById "+strings.Join(removedRules, " ")+"\n")
				for _, tag := range removedTags {
					if !tagRE.MatchString(tag) {
						continue
					}
					filtered = append(filtered, "            SecRuleRemoveByTag \""+tag+"\"\n")
				}
				filtered = append(filtered, line)
				newLines = append(newLines, filtered...)
				buffer = nil
				insideDirectives = false
			} else {
				buffer = append(buffer, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	return strings.Join(newLines, "")
}
