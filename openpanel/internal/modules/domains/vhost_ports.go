package domains

import (
	"regexp"
	"slices"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// VhostFileInfo is what the VHosts list shows about each domain's file
type VhostFileInfo struct {
	Lines int
	Ports []string // like "80" and "443 SSL", empty when the file names none (LiteSpeed)
	SSL   bool
}

var (
	apacheVhostRE = regexp.MustCompile(`(?i)^\s*<VirtualHost\s+[^>]*:(\d+)\s*>`)
	nginxListenRE = regexp.MustCompile(`^\s*listen\s+([^;]+);`)
	portRE        = regexp.MustCompile(`(?:^|:)(\d+)$`)
)

// readVhostFiles reads every domain's vhost file once for the list page
func readVhostFiles(userContext string, domainsList []appctx.Domain) map[string]VhostFileInfo {
	out := make(map[string]VhostFileInfo, len(domainsList))
	for _, d := range domainsList {
		if content := readVhostContent(userContext, d.DomainURL); content != "" {
			out[d.DomainURL] = parseVhostFile(content)
		}
	}
	return out
}

// parseVhostFile counts lines and picks the listening ports out of an Apache, Nginx/OpenResty or LiteSpeed vhost
func parseVhostFile(content string) VhostFileInfo {
	var info VhostFileInfo
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	info.Lines = len(lines)
	add := func(port string, ssl bool) {
		label := port
		if ssl || port == "443" {
			label, info.SSL = port+" SSL", true
		}
		if !slices.Contains(info.Ports, label) {
			info.Ports = append(info.Ports, label)
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if m := apacheVhostRE.FindStringSubmatch(line); m != nil {
			add(m[1], false)
			continue
		}
		if m := nginxListenRE.FindStringSubmatch(line); m != nil {
			fields := strings.Fields(m[1])
			if p := portRE.FindStringSubmatch(fields[0]); p != nil {
				add(p[1], slices.Contains(fields[1:], "ssl"))
			}
			continue
		}
		// LiteSpeed vhosts have no ports, only an SSL block
		if strings.HasPrefix(trimmed, "vhssl") {
			info.SSL = true
		}
	}
	return info
}
