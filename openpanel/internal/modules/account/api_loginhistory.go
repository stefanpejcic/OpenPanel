package account

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterLoginHistoryAPI wires GET /api/account/login-history onto mux,
// gated behind the "login_history" feature flag.
func RegisterLoginHistoryAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "login_history", "GET /api/account/login-history", func(w http.ResponseWriter, r *http.Request) { apiLoginHistory(a, w, r) })
}

type apiLoginHistoryEntry struct {
	IP          string `json:"ip"`
	CountryCode string `json:"country_code"`
	LoginTime   string `json:"login_time"`
}

// apiLoginHistory returns recent login attempts for the caller, most
// recent first (matching the web page's ordering).
func apiLoginHistory(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	lastLoginData, loginErr := a.GetLastLoginData(ctx, username)
	if loginErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	entries := make([]apiLoginHistoryEntry, len(lastLoginData))
	for i, l := range lastLoginData {
		entries[len(lastLoginData)-1-i] = apiLoginHistoryEntry{IP: l.IP, CountryCode: l.CountryCode, LoginTime: l.LoginTime}
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"entries": entries})
}
