package account

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// NotificationPref is one key=value line from a user's notifications.yaml
type NotificationPref struct {
	Key   string
	Value string
	Label string
}

type notificationDef struct {
	Key, Default, Title, Description string
	Group, Anchor                    string
	Parent                           string
}

// notificationDefs are the only keys the page shows and saves, in the order they get written back
var notificationDefs = []notificationDef{
	{Key: "notify_account_login", Default: "1", Title: "New login", Description: "Get an email every time someone logs in to your account, with the IP address, country and browser.", Group: "Security", Anchor: "new-login"},
	{Key: "notify_account_login_for_known_netblock", Default: "0", Title: "Also for IP addresses I logged in from before", Parent: "notify_account_login"},
	{Key: "notify_account_login_notification_disabled", Default: "1", Title: "Email me if login alerts get turned off", Parent: "notify_account_login"},
	{Key: "notify_password_change", Default: "1", Title: "Password changed", Description: "Get an email when the password for your account is changed.", Group: "Security", Anchor: "password-changed"},
	{Key: "notify_password_change_notification_disabled", Default: "1", Title: "Email me if this alert gets turned off", Parent: "notify_password_change"},
	{Key: "notify_contact_address_change", Default: "1", Title: "Contact email changed", Description: "Get an email when the contact email address for your account is changed.", Group: "Security", Anchor: "contact-email-changed"},
	{Key: "notify_contact_address_change_notification_disabled", Default: "1", Title: "Email me if this alert gets turned off", Parent: "notify_contact_address_change"},
	{Key: "notify_twofactorauth_change", Default: "1", Title: "Two-factor authentication changed", Description: "Get an email when two-factor authentication is turned on or off.", Group: "Security", Anchor: "two-factor-authentication-changed"},
	{Key: "notify_twofactorauth_change_notification_disabled", Default: "1", Title: "Email me if this alert gets turned off", Parent: "notify_twofactorauth_change"},
	{Key: "notify_passkey_change", Default: "1", Title: "Passkey added or removed", Description: "Get an email when a passkey is added to your account or removed from it.", Group: "Security", Anchor: "passkey-added-or-removed"},
	{Key: "notify_passkey_change_notification_disabled", Default: "1", Title: "Email me if this alert gets turned off", Parent: "notify_passkey_change"},
	{Key: "notify_api_token_change", Default: "1", Title: "API token created or revoked", Description: "Get an email when an AI Assistant (MCP) token is created or revoked for your account.", Group: "Security", Anchor: "api-token-created-or-revoked"},
	{Key: "notify_api_token_change_notification_disabled", Default: "1", Title: "Email me if this alert gets turned off", Parent: "notify_api_token_change"},
	{Key: "notify_malware_found", Default: "1", Title: "Malware found", Description: "Get an email when the scheduled malware scan finds infected files and moves them to quarantine.", Group: "Security", Anchor: "malware-found"},
	{Key: "notify_disk_limit", Default: "1", Title: "Disk space running out", Description: "Get an email when your account uses 85% of its disk space or 95% of its inodes.", Group: "Usage", Anchor: "disk-space-running-out"},
	{Key: "notify_email_quota_limit", Default: "1", Title: "Mailbox almost full", Description: "Get an email when one of your email accounts uses 90% of its quota.", Group: "Usage", Anchor: "mailbox-almost-full"},
	{Key: "notify_email_ratelimit", Default: "1", Title: "Hourly email limit reached", Description: "Get an email when your account reaches its limit of emails sent per hour and new emails are rejected.", Group: "Usage", Anchor: "hourly-email-limit-reached"},
	{Key: "notify_service_failed", Default: "1", Title: "Service stopped", Description: "Get an email when a service like MySQL or PHP stopped or ran out of memory and had to be restarted.", Group: "Usage", Anchor: "service-stopped"},
}

// readNotificationsPrefs returns every known key, a key missing from the file gets its default
func readNotificationsPrefs(username string) []NotificationPref {
	values := map[string]string{}
	if content, err := os.ReadFile(notificationsFilePath(username)); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
			if ok && !strings.HasPrefix(k, "#") {
				values[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	prefs := make([]NotificationPref, 0, len(notificationDefs))
	for _, d := range notificationDefs {
		v := values[d.Key]
		if v != "0" && v != "1" {
			v = d.Default
		}
		prefs = append(prefs, NotificationPref{Key: d.Key, Value: v, Label: d.Title})
	}
	return prefs
}

// criticalNotificationKeys always fire a notification themselves when flipped from '1' to '0', so the user doesn't silently lose visibility into a security-relevant setting being turned off
var criticalNotificationKeys = []string{
	"notify_account_login_notification_disabled",
	"notify_contact_address_change_notification_disabled",
	"notify_password_change_notification_disabled",
	"notify_twofactorauth_change_notification_disabled",
	"notify_passkey_change_notification_disabled",
	"notify_api_token_change_notification_disabled",
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

// prefsChangedMessage is the email sent when a watched alert gets turned off, listing every alert that went off
func prefsChangedMessage(a *appctx.App, r *http.Request, username string, oldValues, newValues map[string]string) userEmail {
	var off []string
	for _, d := range notificationDefs {
		if d.Parent == "" && oldValues[d.Key] == "1" && newValues[d.Key] == "0" {
			off = append(off, "- "+d.Title)
		}
	}
	return securityEmail(a, r, "Notification preferences changed for account "+username,
		"These email alerts were turned off for account "+username+":\n"+strings.Join(off, "\n")+"\n\nYou won't get emails for them anymore.",
		"If you didn't turn them off, turn them back on from Account > Notifications and change your password right away.")
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
			checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_always", prefsChangedMessage(a, r, username, originalValues, newValues))
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, "success", "Notification preferences updated successfully!")
		_ = a.Sessions.Save(r, w, sess)
	}

	email, _ := data["current_email"].(string)
	renderNotificationsPage(a, w, r, prefs, email)
}

// RegisterNotifications wires the notification-preferences route onto mux, gated behind the "notifications" feature flag
func RegisterNotifications(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "notifications")(h)
	}
	mux.Handle("/account/notifications", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountNotifications(a, w, r) }))
}
