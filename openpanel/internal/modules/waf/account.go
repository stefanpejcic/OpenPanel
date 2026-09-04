package waf

import (
	"fmt"
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// wafDisabledMarkerPath is the per-account switch for whether WAF is
// applied to newly created domains - not to be confused with a single
// domain's own SecRuleRemoveById/SecRuleRemoveByTag exclusions (waf.go).
func wafDisabledMarkerPath(userContext string) string {
	return fmt.Sprintf("/home/%s/waf.disabled", userContext)
}

// AccountWAFEnabled reports whether WAF is enabled by default for new
// domains on this account: enabled unless wafDisabledMarkerPath exists.
func AccountWAFEnabled(userContext string) bool {
	_, err := os.Stat(wafDisabledMarkerPath(userContext))
	return os.IsNotExist(err)
}

// SetAccountWAFEnabled creates or removes wafDisabledMarkerPath to match
// enabled, the same on/off switch AccountWAFEnabled reads.
func SetAccountWAFEnabled(userContext string, enabled bool) error {
	path := wafDisabledMarkerPath(userContext)
	if enabled {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, nil, 0o644)
}

// injectedContext is like injected() (waf.go) but also returns the
// account's context, needed for /home/<context>/-rooted paths like
// wafDisabledMarkerPath - injected() only returns current_username since
// none of its other callers need context.
func injectedContext(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

// RegisterAccountAPI wires the account-wide WAF on/off toggle onto mux -
// JSON only, no page of its own (the onboarding wizard is its first
// caller; a settings page can call the same GET/POST later).
func RegisterAccountAPI(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "waf")(h)
	}
	mux.Handle("/waf/account", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFAccountToggle(a, w, r) }))
}

func handleWAFAccountToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		action := r.Form.Get("action")
		if action != "enable" && action != "disable" {
			writeJSONError(w, http.StatusBadRequest, "action must be 'enable' or 'disable'")
			return
		}
		enabled := action == "enable"
		if err := SetAccountWAFEnabled(userContext, enabled); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to update WAF: "+err.Error())
			return
		}
		verb := "disabled"
		if enabled {
			verb = "enabled"
		}
		_ = logger.RecordUserAction(a.Config, username, verb+" WAF for new domains", reqip.ClientIP(r))
	}

	writeJSON(w, http.StatusOK, map[string]bool{"enabled": AccountWAFEnabled(userContext)})
}
