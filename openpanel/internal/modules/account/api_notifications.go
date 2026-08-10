package account

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterNotificationsAPI wires the /api/account/notifications routes
// onto mux, gated behind the "notifications" feature flag.
func RegisterNotificationsAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "notifications", "GET /api/account/notifications", func(w http.ResponseWriter, r *http.Request) { apiNotificationsGet(a, w, r) })
	apiregistry.Handle(mux, a, "notifications", "PUT /api/account/notifications", func(w http.ResponseWriter, r *http.Request) { apiNotificationsUpdate(a, w, r) })
}

type apiNotificationPref struct {
	Key   string `json:"key"`
	Value bool   `json:"value"`
	Label string `json:"label"`
}

func toAPINotificationPrefs(prefs []NotificationPref) []apiNotificationPref {
	out := make([]apiNotificationPref, len(prefs))
	for i, p := range prefs {
		out[i] = apiNotificationPref{Key: p.Key, Value: p.Value == "1", Label: p.Label}
	}
	return out
}

// apiNotificationsGet returns the caller's notification preferences.
func apiNotificationsGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"preferences": toAPINotificationPrefs(readNotificationsPrefs(username))})
}

// apiNotificationsUpdate applies a partial set of preference changes:
// {"preferences": {"notify_account_login": true, ...}}. Keys not present
// in the body are left unchanged, unlike the web form (which submits every
// checkbox at once).
func apiNotificationsUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	var body struct {
		Preferences map[string]bool `json:"preferences"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil || len(body.Preferences) == 0 {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "preferences object is required"})
		return
	}

	prefs := readNotificationsPrefs(username)
	weShouldNotifyUser := false
	for i := range prefs {
		newVal, present := body.Preferences[prefs[i].Key]
		if !present {
			continue
		}
		oldVal := prefs[i].Value
		if newVal {
			prefs[i].Value = "1"
		} else {
			prefs[i].Value = "0"
		}
		for _, critical := range criticalNotificationKeys {
			if prefs[i].Key == critical && oldVal == "1" && prefs[i].Value == "0" {
				weShouldNotifyUser = true
			}
		}
	}

	var sb strings.Builder
	for _, p := range prefs {
		sb.WriteString(p.Key + "=" + p.Value + "\n")
	}
	path := notificationsFilePath(username)
	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if writeErr := os.WriteFile(path, []byte(sb.String()), 0o644); writeErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "changed notification preferences for the account via API", reqip.ClientIP(r))
	if weShouldNotifyUser {
		message := "Notification preferences changed for account " + username + "\n Notification preferences have been changed for your account <b>" + username +
			"</b>.<br><br> Review the new notification settings from <b>Account > Notifications</b> page."
		checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_always", message)
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"preferences": toAPINotificationPrefs(prefs)})
}
