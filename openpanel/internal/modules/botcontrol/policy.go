package botcontrol

import (
	"context"
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

// reloadCaddy reloads the Caddy proxy container so updated domain rules take effect.
var reloadCaddy = func(ctx context.Context) error {
	return exec.CommandContext(ctx, "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run()
}

func domainConfigPath(domainName string) string {
	return "/etc/openpanel/caddy/domains/" + domainName + ".conf"
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

// rewritePolicy replaces any existing botcontrol directive or inserts one before every WAF/upstream block.
func rewritePolicy(content, domainName string, p Policy) string {
	directive := directiveForPolicy(domainName, p)

	switch {
	case botBlockRE.MatchString(content):
		return botBlockRE.ReplaceAllString(content, "    "+directive)
	case botRateLimitRE.MatchString(content):
		return botRateLimitRE.ReplaceAllString(content, "    "+directive)
	case botAllowRE.MatchString(content):
		return botAllowRE.ReplaceAllString(content, "    "+directive)
	case wafAnchorRE.MatchString(content):
		return wafAnchorRE.ReplaceAllString(content, "${1}"+directive+"\n\n${1}${2}")
	case proxyAnchorRE.MatchString(content):
		return proxyAnchorRE.ReplaceAllString(content, "${1}"+directive+"\n\n${1}${2}")
	default:
		return strings.TrimRight(content, "\n") + "\n# botcontrol: " + string(p) + "\n"
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

// setDomainPolicy updates the domain's Caddy config and reloads Caddy.
func setDomainPolicy(ctx context.Context, domainName string, p Policy) error {
	path := domainConfigPath(domainName)
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	updated := rewritePolicy(string(content), domainName, p)
	if err := tenantfs.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	return reloadCaddy(ctx)
}
