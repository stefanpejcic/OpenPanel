package account

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// NotificationPref is one key=value line from a user's notifications.yaml, kept in on-disk order since the page rewrites the file in the same order it read it
type NotificationPref struct {
	Key   string
	Value string
	Label string
}

// notificationLabel mirrors notifications.html's display transform: strips the "notify" prefix for shipped keys, or Title Case otherwise (dead in practice, every shipped key starts with "notify")
func notificationLabel(key string) string {
	if strings.HasPrefix(key, "notify") {
		return strings.ReplaceAll(key[len("notify"):], "_", " ")
	}
	return titleCase(strings.ReplaceAll(key, "_", " "))
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func readNotificationsPrefs(username string) []NotificationPref {
	content, err := os.ReadFile(notificationsFilePath(username))
	if err != nil {
		return nil
	}
	var prefs []NotificationPref
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		prefs = append(prefs, NotificationPref{Key: key, Value: strings.TrimSpace(v), Label: notificationLabel(key)})
	}
	return prefs
}

// criticalNotificationKeys always fire a notification themselves when flipped from '1' to '0', so the user doesn't silently lose visibility into a security-relevant setting being turned off
var criticalNotificationKeys = []string{
	"notify_account_login_notification_disabled",
	"notify_contact_address_change_notification_disabled",
	"notify_password_change_notification_disabled",
	"notify_twofactorauth_change_notification_disabled",
}

// alertWasDisabled is true when a watched alert like notify_password_change was switched off while its _notification_disabled key was on
func alertWasDisabled(oldValues, newValues map[string]string) bool {
	for _, key := range criticalNotificationKeys {
		watched := strings.TrimSuffix(key, "_notification_disabled")
		if oldValues[key] == "1" && oldValues[watched] == "1" && newValues[watched] == "0" {
			return true
		}
	}
	return false
}

// clearNotificationPrefsCache drops the cached values so a saved change applies right away
func clearNotificationPrefsCache(ctx context.Context, a *appctx.App, username string, prefs []NotificationPref) {
	for _, p := range prefs {
		_ = a.Cache.Delete(ctx, "get_from_file_value:"+username+":"+p.Key)
	}
}

// handleAccountNotifications views or updates a user's notification preferences.
func handleAccountNotifications(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	prefs := readNotificationsPrefs(username)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()

		originalValues := make(map[string]string, len(prefs))
		for _, p := range prefs {
			originalValues[p.Key] = p.Value
		}

		for i := range prefs {
			if _, checked := r.Form[prefs[i].Key]; checked {
				prefs[i].Value = "1"
			} else {
				prefs[i].Value = "0"
			}
		}

		newValues := make(map[string]string, len(prefs))
		for _, p := range prefs {
			newValues[p.Key] = p.Value
		}

		weShouldNotifyUser := alertWasDisabled(originalValues, newValues)

		var sb strings.Builder
		for _, p := range prefs {
			sb.WriteString(p.Key + "=" + p.Value + "\n")
		}
		path := notificationsFilePath(username)
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr == nil {
			_ = os.WriteFile(path, []byte(sb.String()), 0o644)
		}

		clearNotificationPrefsCache(ctx, a, username, prefs)
		_ = logger.RecordUserAction(a.Config, username, "changed notification preferences for the account", reqip.ClientIP(r))

		if weShouldNotifyUser {
			message := "Notification preferences changed for account " + username + "\n Notification preferences have been changed for your account <b>" + username +
				"</b>.<br><br> Review the new notification settings from <b>Account > Notifications</b> page."
			checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_always", message)
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, "success", "Notification preferences updated successfully!")
		_ = a.Sessions.Save(r, w, sess)
	}

	renderNotificationsPage(a, w, r, prefs)
}

// RegisterNotifications wires the notification-preferences route onto mux, gated behind the "notifications" feature flag
func RegisterNotifications(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "notifications")(h)
	}
	mux.Handle("/account/notifications", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountNotifications(a, w, r) }))
}
