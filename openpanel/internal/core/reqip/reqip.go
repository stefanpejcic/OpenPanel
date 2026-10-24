// Package reqip extracts the client's IP, taking Cloudflare's CF-Connecting-IP only when the request really came through Cloudflare.
package reqip

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// cloudflareRanges is the same list as trusted_proxies in the Caddyfile, from https://www.cloudflare.com/ips/
var cloudflareRanges = mustPrefixes(
	"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22", "141.101.64.0/18",
	"108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22", "198.41.128.0/17",
	"162.158.0.0/15", "104.16.0.0/13", "104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
	"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32", "2405:8100::/32",
	"2a06:98c0::/29", "2c0f:f248::/32",
)

func mustPrefixes(cidrs ...string) []netip.Prefix {
	out := make([]netip.Prefix, len(cidrs))
	for i, c := range cidrs {
		out[i] = netip.MustParsePrefix(c)
	}
	return out
}

// ClientIP is the connection's address, or the last X-Forwarded-For hop when a local proxy like Caddy is in front, and CF-Connecting-IP only when that address is Cloudflare's
func ClientIP(r *http.Request) string {
	peer, ok := remoteAddr(r)
	if !ok {
		return r.RemoteAddr
	}
	if peer.IsLoopback() || peer.IsPrivate() {
		if hop, ok := lastForwardedHop(r); ok {
			peer = hop
		}
	}
	if fromCloudflare(peer) {
		if cf, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); err == nil {
			return cf.Unmap().String()
		}
	}
	return peer.String()
}

func remoteAddr(r *http.Request) (netip.Addr, bool) {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	a, err := netip.ParseAddr(host)
	return a.Unmap(), err == nil
}

// lastForwardedHop is the address the proxy in front of us saw, Caddy always puts it last
func lastForwardedHop(r *http.Request) (netip.Addr, bool) {
	values := r.Header.Values("X-Forwarded-For")
	if len(values) == 0 {
		return netip.Addr{}, false
	}
	hops := strings.Split(values[len(values)-1], ",")
	a, err := netip.ParseAddr(strings.TrimSpace(hops[len(hops)-1]))
	return a.Unmap(), err == nil
}

func fromCloudflare(a netip.Addr) bool {
	for _, p := range cloudflareRanges {
		if p.Contains(a) {
			return true
		}
	}
	return false
}
