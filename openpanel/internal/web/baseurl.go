package web

import (
	"context"
	"net"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

// PublicBaseURL reconstructs the scheme+host+port the same way enforceAccessDomain() (internal/auth/loaduser.go) does for the canonical-domain redirect - by render time, the request's host has already passed that check
func PublicBaseURL(ctx context.Context, a *appctx.App, r *http.Request) string {
	requestHost, requestPort := r.Host, ""
	if h, p, err := net.SplitHostPort(requestHost); err == nil {
		requestHost, requestPort = h, p
	}

	host := strings.TrimSpace(a.ForceDomain)
	if host == "" {
		host = requestHost
	}

	scheme := "http"
	if sysinfo.HasSSL(ctx, a.Cache, host) {
		scheme = "https"
	}

	// a.ForcePort comes from `opencli port`, which can fail to resolve in some deployments - fall back to the port this request came in on rather than omitting it, since that's the port actually in use right now
	portSuffix := ""
	switch {
	case a.ForcePort != "":
		portSuffix = ":" + a.ForcePort
	case isNumericPort(requestPort):
		// only a real numeric port: net.SplitHostPort("x:@evil") yields port "@evil", which would put
		// attacker-controlled authority into the dynamic-DNS / git URLs built from this base
		portSuffix = ":" + requestPort
	}

	return scheme + "://" + host + portSuffix + "/"
}

// isNumericPort reports whether s is a decimal TCP port in 1..65535 - net.SplitHostPort accepts a
// non-numeric "port" such as the "@evil.example" from a "host:@evil.example" Host header.
func isNumericPort(s string) bool {
	if s == "" || len(s) > 5 {
		return false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		n = n*10 + int(c-'0')
	}
	return n >= 1 && n <= 65535
}
