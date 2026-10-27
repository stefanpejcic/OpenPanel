package appinstall

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// markers around everything an app proxy adds to an nginx vhost, so removing it puts the file back byte for byte
const (
	nginxProxyStart = "# OPENPANEL-PROXY-START:"
	nginxProxyEnd   = "# OPENPANEL-PROXY-END:"
	nginxProxyOrig  = "# OPENPANEL-PROXY-ORIG:"
)

// nginxProxy is one app proxy for addNginxProxy, an empty Resolver falls back to a plain proxy_pass that nginx resolves at startup
type nginxProxy struct {
	Service      string
	Port         int
	Subdirectory string
	Resolver     string
	Headers      []string
}

var nginxRootLocationRE = regexp.MustCompile(`(?m)^([ \t]*)location[ \t]+/[ \t]*\{`)

// exact matches like = /favicon.ico beat ^~ /, so a root app has to take those over too
var nginxExactLocationRE = regexp.MustCompile(`(?m)^([ \t]*)location[ \t]+=[ \t]*\S+[ \t]*\{`)

// the subdirectory ends up in a location and a rewrite regex, so only plain path characters
var nginxSubdirRE = regexp.MustCompile(`^[a-z0-9._~-]+(/[a-z0-9._~-]+)*$`)

// addNginxProxy proxies p in every server block: a root app takes over location / and the exact-match locations (old blocks kept as comments), a subdirectory app gets its own location next to location /
func addNginxProxy(conf string, p nginxProxy) (string, error) {
	if strings.Contains(conf, nginxProxyStart+p.Service+"\n") {
		return conf, nil
	}
	sub := strings.Trim(p.Subdirectory, "/")
	if sub != "" && (!nginxSubdirRE.MatchString(sub) || strings.Contains("/"+sub+"/", "/./") || strings.Contains("/"+sub+"/", "/../")) {
		return "", fmt.Errorf("subdirectory %q can't be used in an nginx location", p.Subdirectory)
	}
	openers := nginxRootLocationRE.FindAllStringSubmatchIndex(conf, -1)
	if len(openers) == 0 {
		return "", errors.New("no location / block found in vhost configuration")
	}
	if sub == "" {
		openers = append(openers, nginxExactLocationRE.FindAllStringSubmatchIndex(conf, -1)...)
		sort.Slice(openers, func(i, j int) bool { return openers[i][0] < openers[j][0] })
	}

	var b strings.Builder
	last := 0
	for _, m := range openers {
		start, indent := m[0], conf[m[2]:m[3]]
		if start < last {
			// nested in a block that's already replaced
			continue
		}
		b.WriteString(conf[last:start])
		lines := []string{nginxProxyStart + p.Service}
		if nginxRootLocationRE.MatchString(conf[start:m[1]]) {
			lines = append(lines, nginxProxyLocation(p, sub)...)
		}
		last = start
		if sub == "" {
			end, err := nginxBlockEnd(conf, m[1]-1)
			if err != nil {
				return "", err
			}
			for _, l := range strings.SplitAfter(conf[start:end], "\n") {
				if l != "" {
					lines = append(lines, nginxProxyOrig+strings.TrimSuffix(l, "\n"))
				}
			}
			last = end
		}
		lines = append(lines, nginxProxyEnd+p.Service)
		for _, l := range lines {
			b.WriteString(indent + l + "\n")
		}
	}
	b.WriteString(conf[last:])
	return b.String(), nil
}

// nginxProxyLocation is the location block for p, ^~ so the vhost's .php and static file regex locations can't grab the app's URLs
func nginxProxyLocation(p nginxProxy, sub string) []string {
	host := p.Service
	if p.Port != 0 {
		host += ":" + strconv.Itoa(p.Port)
	}
	path := "/"
	if sub != "" {
		path = "/" + sub + "/"
	}
	lines := []string{"location ^~ " + path + " {"}
	switch {
	case p.Resolver != "":
		// looked up per request, a stopped app then only 502s its own URLs instead of keeping nginx from starting
		lines = append(lines, "    resolver "+p.Resolver+" valid=10s ipv6=off;", "    set $openpanel_upstream "+host+";")
		if sub != "" {
			lines = append(lines, "    rewrite ^"+regexp.QuoteMeta(path)+"(.*)$ /$1 break;")
		}
		lines = append(lines, "    proxy_pass http://$openpanel_upstream;")
	case sub != "":
		lines = append(lines, "    proxy_pass http://"+host+"/;")
	default:
		lines = append(lines, "    proxy_pass http://"+host+";")
	}
	for _, h := range p.Headers {
		if h = strings.TrimSpace(h); h != "" {
			lines = append(lines, "    "+h)
		}
	}
	return append(lines, "}")
}

// nginxBlockEnd returns where the block opened at conf[open] ends, its closing line included, skipping braces in comments and quotes
func nginxBlockEnd(conf string, open int) (int, error) {
	depth := 0
	for i := open; i < len(conf); i++ {
		switch c := conf[i]; c {
		case '#':
			nl := strings.IndexByte(conf[i:], '\n')
			if nl == -1 {
				return 0, errors.New("location / block is not closed in vhost configuration")
			}
			i += nl
		case '"', '\'':
			for i++; i < len(conf) && conf[i] != c; i++ {
				if conf[i] == '\\' {
					i++
				}
			}
		case '{':
			depth++
		case '}':
			depth--
			if depth > 0 {
				continue
			}
			rest := conf[i+1:]
			if nl := strings.IndexByte(rest, '\n'); nl != -1 {
				rest = rest[:nl+1]
			}
			if strings.TrimSpace(rest) != "" {
				return 0, errors.New("location / block shares its closing line with other config")
			}
			return i + 1 + len(rest), nil
		}
	}
	return 0, errors.New("location / block is not closed in vhost configuration")
}

// restoreNginxProxy turns a marked block back into the location / it replaced, or nothing for a subdirectory one
func restoreNginxProxy(region string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(region, "\n") {
		if _, orig, ok := strings.Cut(line, nginxProxyOrig); ok {
			b.WriteString(orig)
		}
	}
	return b.String()
}

// nginxResolver is the gateway of the account's www network, where aardvark-dns answers for its containers, or "" if podman can't tell
func nginxResolver(ctx context.Context, userContext string) string {
	argv := podmanmanager.PodmanArgv(userContext, "network", "inspect", userContext+"_www", "--format", "{{range .Subnets}}{{.Gateway}} {{end}}")
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return ""
	}
	for _, f := range strings.Fields(string(out)) {
		if ip := net.ParseIP(f); ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}
