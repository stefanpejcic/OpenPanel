// Package reqip extracts the client's IP, preferring Cloudflare's CF-Connecting-IP header (the panel usually sits behind Cloudflare/Caddy) over the raw connection address.
package reqip

import (
	"net"
	"net/http"
)

func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
