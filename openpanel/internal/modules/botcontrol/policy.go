package botcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// Policy represents the per-domain AI crawler enforcement mode.
type Policy string

const (
	PolicyAllow     Policy = "allow"
	PolicyRateLimit Policy = "ratelimit"
	PolicyBlock     Policy = "block"
)

var (
	botBlockRE     = regexp.MustCompile(`(?m)^[ \t]*import[ \t]+botcontrol_block\b.*$`)
	botRateLimitRE = regexp.MustCompile(`(?m)^[ \t]*import[ \t]+botcontrol_ratelimit\b.*$`)
	botAllowRE     = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*botcontrol:[ \t]*allow\b.*$`)
	wafAnchorRE    = regexp.MustCompile(`(?m)^([ \t]*)(coraza_waf\b)`)
	proxyAnchorRE  = regexp.MustCompile(`(?m)^([ \t]*)(reverse_proxy\b|import[ \t]+reverse_proxy_upstream\b)`)
)

// reloadCaddy reloads caddy and returns caddy's own error so the user sees why a config was rejected.
var reloadCaddy = func(ctx context.Context) error {
	out, err := exec.CommandContext(ctx, "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").CombinedOutput()
	if err != nil {
		if msg := caddyErrorLine(string(out)); msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

// caddyErrorLine pulls caddy's reason out of its output, older caddy prints "Error: ...", newer logs it as a JSON error line
func caddyErrorLine(out string) string {
	msg := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Error: ") {
			return strings.TrimPrefix(line, "Error: ")
		}
		var entry struct {
			Level string `json:"level"`
			Msg   string `json:"msg"`
		}
		if json.Unmarshal([]byte(line), &entry) == nil && entry.Level == "error" && entry.Msg != "" {
			msg = entry.Msg
		}
	}
	if msg != "" {
		return msg
	}
	return strings.TrimSpace(out)
}

// errNoAnchor means the domain config has no marker, WAF or upstream line to put the directive next to.
var errNoAnchor = errors.New("no place for the bot control rule in the domain's Caddy config")

// caddyDomainsDir is a var so tests can point it at a temp dir.
var caddyDomainsDir = "/etc/openpanel/caddy/domains/"

func domainConfigPath(domainName string) string {
	return caddyDomainsDir + domainName + ".conf"
}

func validPolicy(p Policy) bool {
	return p == PolicyAllow || p == PolicyRateLimit || p == PolicyBlock
}

// policyFromConfig reads the active bot-control policy from a Caddy domain config.
func policyFromConfig(content string) Policy {
	if botBlockRE.MatchString(content) {
		return PolicyBlock
	}
	if botRateLimitRE.MatchString(content) {
		return PolicyRateLimit
	}
	return PolicyAllow
}

// directiveForPolicy returns the Caddy line to place inside the domain's route block.
func directiveForPolicy(domainName string, p Policy) string {
	switch p {
	case PolicyBlock:
		return "import botcontrol_block"
	case PolicyRateLimit:
		return "import botcontrol_ratelimit " + domainName
	default:
		return "# botcontrol: allow"
	}
}

// botAnyLineRE catches any botcontrol line, so ones left outside a site block can be dropped
var botAnyLineRE = regexp.MustCompile(`^[ \t]*(import[ \t]+botcontrol_(block|ratelimit)\b|#[ \t]*botcontrol:)`)

// dropTopLevelBotLines removes botcontrol lines outside any site block, at the top level they'd break the whole Caddyfile
func dropTopLevelBotLines(content string) string {
	lines := strings.Split(content, "\n")
	out := lines[:0]
	depth, inRaw := 0, false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if depth == 0 && !inRaw && botAnyLineRE.MatchString(line) {
			continue
		}
		out = append(out, line)
		if !inRaw && strings.HasPrefix(trimmed, "#") {
			continue
		}
		for _, c := range line {
			switch {
			case c == '`':
				inRaw = !inRaw
			case inRaw:
			case c == '{':
				depth++
			case c == '}' && depth > 0:
				depth--
			}
		}
	}
	return strings.Join(out, "\n")
}

// rewritePolicy swaps the botcontrol line or inserts one before the WAF/upstream, errNoAnchor if there's nowhere to put it.
func rewritePolicy(content, domainName string, p Policy) (string, error) {
	directive := directiveForPolicy(domainName, p)
	content = dropTopLevelBotLines(content)

	switch {
	case botBlockRE.MatchString(content):
		return botBlockRE.ReplaceAllString(content, "    "+directive), nil
	case botRateLimitRE.MatchString(content):
		return botRateLimitRE.ReplaceAllString(content, "    "+directive), nil
	case botAllowRE.MatchString(content):
		return botAllowRE.ReplaceAllString(content, "    "+directive), nil
	case wafAnchorRE.MatchString(content):
		return wafAnchorRE.ReplaceAllString(content, "${1}"+directive+"\n\n${1}${2}"), nil
	case proxyAnchorRE.MatchString(content):
		return proxyAnchorRE.ReplaceAllString(content, "${1}"+directive+"\n\n${1}${2}"), nil
	case p == PolicyAllow:
		return content, nil
	default:
		return "", errNoAnchor
	}
}

// getDomainPolicy reads the domain's Caddy config and returns its current Policy.
func getDomainPolicy(domainName string) Policy {
	content, err := tenantfs.ReadFile(domainConfigPath(domainName))
	if err != nil {
		return PolicyAllow
	}
	return policyFromConfig(string(content))
}

// setDomainPolicy puts the old file back if caddy rejects the change, otherwise caddy won't start next time.
func setDomainPolicy(ctx context.Context, domainName string, p Policy) error {
	path := domainConfigPath(domainName)
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	updated, err := rewritePolicy(string(content), domainName, p)
	if err != nil {
		return err
	}
	if updated == string(content) {
		return nil
	}
	if err := tenantfs.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	if err := reloadCaddy(ctx); err != nil {
		if rerr := tenantfs.WriteFile(path, content, 0o644); rerr != nil {
			return fmt.Errorf("caddy rejected the change (%v) and restoring %s failed: %v", err, path, rerr)
		}
		return fmt.Errorf("caddy rejected the change, nothing was changed: %v", err)
	}
	return nil
}
