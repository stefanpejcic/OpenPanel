// Package ipblocker is a per-user IP/CIDR blocklist backed by the
// opencli user-block_ip wrapper.
package ipblocker

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

func injected(a *appctx.App, r *http.Request) (username string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", err
	}
	username, _ = data["current_username"].(string)
	return username, nil
}

// normalizeIP: a "/" means CIDR (host bits are zeroed out to the network
// address), otherwise a plain address. Returns ok=false for anything that
// fails to parse, which callers treat as silently skippable.
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
