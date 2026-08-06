package account

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// handleLoginHistory shows recent login attempts for this account. The
// underlying a.GetLastLoginData is already cached (600s), so this doesn't
// add another cache layer on top.
func handleLoginHistory(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	for i, j := 0, len(lastLoginData)-1; i < j; i, j = i+1, j-1 {
		lastLoginData[i], lastLoginData[j] = lastLoginData[j], lastLoginData[i]
	}

	renderLoginHistoryPage(a, w, r, lastLoginData)
}

// RegisterLoginHistory wires the login-history route onto mux, gated
// behind the "login_history" feature flag.
func RegisterLoginHistory(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "login_history")(h)
	}
	mux.Handle("/account/login-history", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleLoginHistory(a, w, r) }))
}
