package appinstall

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// hostRE matches http(s)/ws(s)://name only as a whole host, so kuma doesn't hit kuma2
func hostRE(name string) string {
	return `(?:https?|wss?)://` + regexp.QuoteMeta(name) + `(?::\d+)?(?:[/;\s"]|$)`
}

var emptyLocationRE = regexp.MustCompile(`(?m)^[ \t]*location[ \t]+([^{\n]+?)[ \t]*\{\s*\}[ \t]*\n?`)

// stripServiceProxies drops every Apache, Nginx and LiteSpeed proxy to one of names from a vhost file
func stripServiceProxies(content string, names ...string) (string, bool) {
	out := content
	for _, name := range names {
		if name == "" {
			continue
		}
		host := hostRE(name)
		apacheRE := regexp.MustCompile(`^\s*ProxyPass(?:Reverse)?\s+\S+\s+` + host)
		nginxRE := regexp.MustCompile(`^\s*proxy_pass\s+` + host)

		lines := strings.SplitAfter(out, "\n")
		kept := make([]string, 0, len(lines))
		// headers the installer adds right after proxy_pass go with it
		inHeaders := false
		// editApacheConfig inserts right after the marker's indent, give that indent back to the marker line
		indent := ""
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if apacheRE.MatchString(line) {
				if indent == "" {
					indent = line[:len(line)-len(strings.TrimLeft(line, " "))]
				}
				continue
			}
			if indent != "" {
				line = indent + strings.TrimLeft(line, " ")
				indent = ""
			}
			if nginxRE.MatchString(line) {
				for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
					kept = kept[:len(kept)-1]
				}
				inHeaders = true
				continue
			}
			if inHeaders {
				if strings.HasPrefix(trimmed, "proxy_") && !strings.HasPrefix(trimmed, "proxy_pass") {
					continue
				}
				if trimmed == "" {
					continue
				}
				inHeaders = false
			}
			kept = append(kept, line)
		}
		out = strings.Join(kept, "")

		// a root app's ^~ and commented try_files go back to how the vhost had them
		out = strings.ReplaceAll(out, nginxRootAppHeader(name)+"\n", "location / {\n")
		out = strings.ReplaceAll(out, nginxAppTag(name), "")

		svc := regexp.QuoteMeta(name)
		lsws := regexp.MustCompile(`(?s)\n?# PM2-PROXY-START:` + svc + `\n.*?# PM2-PROXY-END:` + svc + `\n?`)
		out = lsws.ReplaceAllString(out, "\n")
	}

	// a subdirectory proxy leaves its location block empty, the root one is never removed
	out = emptyLocationRE.ReplaceAllStringFunc(out, func(m string) string {
		if sub := emptyLocationRE.FindStringSubmatch(m); sub != nil && strings.TrimSpace(sub[1]) == "/" {
			return m
		}
		return ""
	})
	return out, out != content
}

// RemoveServiceProxies cleans every domain vhost of proxies to the container and restarts the webserver if anything changed
func RemoveServiceProxies(ctx context.Context, userContext string, names ...string) []string {
	dir := filepath.Dir(vhostConfPath(userContext, "x"))
	files, _ := filepath.Glob(filepath.Join(dir, "*.conf"))
	webServerType, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	isNginx := webServerType == "nginx" || webServerType == "openresty"

	var changed []string
	originals := map[string][]byte{}
	for _, f := range files {
		content, err := tenantfs.ReadFile(f)
		if err != nil {
			continue
		}
		updated, ok := stripServiceProxies(string(content), names...)
		if !ok {
			continue
		}
		if tenantfs.WriteFile(f, []byte(updated), 0o644) != nil {
			continue
		}
		originals[f] = content
		changed = append(changed, strings.TrimSuffix(filepath.Base(f), ".conf"))
	}
	if len(changed) == 0 || webServerType == "" {
		return changed
	}

	// put the files back if nginx won't take them, a dead proxy beats a dead webserver, a stopped one is tested in a throwaway container
	if isNginx {
		if ok, _ := webserver.TestWebserverConfig(ctx, userContext, webServerType); !ok {
			for f, content := range originals {
				_ = tenantfs.WriteFile(f, content, 0o644)
			}
			return nil
		}
	}
	reloadWebserver(ctx, userContext, webServerType)
	return changed
}
