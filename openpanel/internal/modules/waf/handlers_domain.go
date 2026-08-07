package waf

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

var ruleIDRE = regexp.MustCompile(`^\d+$`)

// reloadCaddy mirrors the `podman exec caddy caddy reload --config
// /etc/caddy/Caddyfile` call every WAF config change makes to apply it.
func reloadCaddy(ctx context.Context) error {
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

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	flashSess(a, w, r, category, message)
	http.Redirect(w, r, path, http.StatusFound)
}

func handleWAFDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domainName := firstPathSegment(r.PathValue("domain"))

	userID, _ := auth.UserID(r)
	username, err := injected(a, r)
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
				flashAndRedirect(a, w, r, "error", "Error: IDs should only contain numbers separated by spaces.", "/server/waf/"+domainName)
				return
			}
		}
		// removed_tags has no meaningful validation: the pattern this was
		// checked against (`^[\w\W]+$`) matches any non-empty string, so
		// every split value already passes.

		removedRules = append([]string{excludedRuleID}, filterOut(removedRules, excludedRuleID)...)
		removedTags = append([]string{excludedTag}, filterOut(removedTags, excludedTag)...)

		if content, readErr := os.ReadFile(configFilePath); readErr == nil {
			newContent := rewriteDirectivesBlock(string(content), removedRules, removedTags)
			if writeErr := os.WriteFile(configFilePath, []byte(newContent), 0o644); writeErr == nil {
				if reloadErr := reloadCaddy(r.Context()); reloadErr == nil {
					_ = logger.RecordUserAction(a.Config, username, "updated WAF rules for domain "+domainName, reqip.ClientIP(r))
					flashSess(a, w, r, "success", "WAF configuration updated.")
				} else {
					log.Printf("WAF - Error occurred during updating WAF settings: %v", reloadErr)
					flashSess(a, w, r, "error", "Error updating WAF!")
				}
			} else {
				log.Printf("WAF - Error occurred during updating WAF settings: %v", writeErr)
				flashSess(a, w, r, "error", "Error updating WAF!")
			}
		} else {
			log.Printf("WAF - Configuration file: %s does not exist.", configFilePath)
			flashSess(a, w, r, "error", "Configuration file not found.")
		}
	}

	// GET (and the POST fallthrough): re-read the file after any change.
	status := "Not Found"
	var removedRules, removedTags []string

	if content, readErr := os.ReadFile(configFilePath); readErr == nil {
		contentStr := string(content)
		switch {
		case strings.Contains(contentStr, "SecRuleEngine On"):
			status = "On"
		case strings.Contains(contentStr, "SecRuleEngine Off"):
			status = "Off"
		default:
			status = "Unknown"
		}

		foundRule, foundTag := false, false
		for _, line := range strings.Split(contentStr, "\n") {
			line = strings.TrimSpace(line)
			if !foundRule && strings.HasPrefix(line, "SecRuleRemoveById") {
				ids := strings.Fields(line)[1:]
				if len(ids) > 0 && ids[0] == excludedRuleID {
					ids = ids[1:]
				}
				removedRules = append(removedRules, ids...)
				foundRule = true
			} else if !foundTag && strings.HasPrefix(line, "SecRuleRemoveByTag") {
				tags := strings.Fields(line)[1:]
				if len(tags) > 0 && strings.EqualFold(tags[0], excludedTag) {
					tags = tags[1:]
				}
				removedTags = append(removedTags, tags...)
				foundTag = true
			}
			if foundRule && foundTag {
				break
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"domain": domainName, "status": status, "removed_rules": removedRules, "removed_tags": removedTags,
		})
		return
	}

	renderWAFDomainPage(a, w, r, domainName, status, removedRules, removedTags)
}

// readLinesKeepEnds splits content into lines; each element (except
// possibly the last) keeps its trailing "\n".
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

// rewriteDirectivesBlock mirrors the line-buffering loop in
// server_settings_waf_for_domain(): finds the `directives \`...\“ block
// in a Caddy per-domain conf, strips any existing
// SecRuleRemoveById/SecRuleRemoveByTag lines from it, and appends fresh
// ones built from removedRules/removedTags.
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
				filtered = append(filtered, "            SecRuleRemoveByTag "+strings.Join(removedTags, " ")+"\n")
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
