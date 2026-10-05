// Package ipblocker is a per-user IP/CIDR blocklist backed by the opencli user-block_ip wrapper.
package ipblocker

import (
	"net/http"
	"net/netip"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// normalizeIP: a "/" means CIDR (host bits are zeroed out to the network address), otherwise a plain address - returns ok=false for anything that fails to parse, which callers treat as silently skippable
func normalizeIP(ip string) (string, bool) {
	if strings.Contains(ip, "/") {
		prefix, err := netip.ParsePrefix(ip)
		if err != nil {
			return "", false
		}
		return prefix.Masked().String(), true
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return "", false
	}
	return addr.String(), true
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	web.WriteJSON(w, status, map[string]string{"error": msg})
}
