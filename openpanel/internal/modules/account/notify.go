package account

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

const notifyConfigFilePath = "/etc/openpanel/openpanel/conf/openpanel.config"

// notificationsFilePath is the one key=value-per-line preferences file per user.
func notificationsFilePath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/notifications.yaml"
}

// parseNotificationsFile is a simple "k=v" line parser (distinct from
// notifications.go's own parser, which additionally skips "#" comment
// lines - see the comment there for why the two aren't unified).
func parseNotificationsFile(content string) map[string]string {
	prefs := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		prefs[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return prefs
}

// getNotificationPreference reads one key from a user's
// notifications.yaml-shaped preferences file, cached for 5 minutes (see
// notifications.go for the same format written by the notifications
// settings page).
func getNotificationPreference(ctx context.Context, a *appctx.App, username, key string) int {
	cacheKey := "get_from_file_value:" + username + ":" + key
	v, _ := cache.Memoize(ctx, a.Cache, cacheKey, 300*time.Second, func() (int, error) {
		path := notificationsFilePath(username)
		content, err := os.ReadFile(path)
		if err != nil {
			return 0, nil
		}
		prefs := parseNotificationsFile(string(content))
		if raw, ok := prefs[key]; ok {
			if n, convErr := strconv.Atoi(raw); convErr == nil {
				return n, nil
			}
		}
		return 0, nil
	})
	return v
}

// checkIfUserShouldBeNotified fires an async notification email if the
// "notifications" feature is
// enabled for the account and the user has opted into this particular
// key (or key is the always-on "notify_always" sentinel). username is
// passed explicitly by the caller (rather than re-derived here) because
// some callers - the username-change flow in particular - already know a
// value that may not match InjectData's own cached current_username yet.
func checkIfUserShouldBeNotified(a *appctx.App, ctx context.Context, userID int, username, key, message string) {
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		return
	}

	userAllowedSlice, _ := data["user_allowed"].([]string)
	userAllowed := make(map[string]bool, len(userAllowedSlice))
	for _, m := range userAllowedSlice {
		userAllowed[m] = true
	}
	if !userAllowed["notifications"] {
		return
	}

	email, _ := data["current_email"].(string)

	fire := func() {
		go notifyUserOfChange(a, username, message, email)
	}

	if key == "notify_always" {
		fire()
		return
	}
	if getNotificationPreference(ctx, a, username, key) == 1 {
		fire()
	}
}

// generateRandomTokenOnce rewrites the *existing* mail_security_token=
// line in openpanel.config in place - a no-op if that line isn't already
// present, it never appends a missing key.
func generateRandomTokenOnce() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 64)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	token := string(b)

	content, err := os.ReadFile(notifyConfigFilePath)
	if err != nil {
		return token
	}

	lines := strings.Split(string(content), "\n")
	newValue := "mail_security_token=" + token
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(line, "mail_security_token=") {
			lines[i] = newValue
			changed = true
		}
	}
	if changed {
		_ = os.WriteFile(notifyConfigFilePath, []byte(strings.Join(lines, "\n")), 0o644)
	}
	return token
}

// notifyUserOfChange looks up the recipient email if not already known,
// then POSTs to the local "openadmin" service's /send_email. Meant to be
// called via `go notifyUserOfChange(...)` - the caller doesn't wait for it.
func notifyUserOfChange(a *appctx.App, username, message, currentEmail string) {
	ctx := context.Background()

	if currentEmail == "" {
		var email sql.NullString
		row := a.DB.QueryRowContext(ctx, "SELECT email FROM users WHERE username = ?", username)
		if err := row.Scan(&email); err == nil {
			currentEmail = email.String
		}
	}

	token := generateRandomTokenOnce()

	adminPort := sysinfo.GetOpenAdminPort()
	forceDomain := sysinfo.GetOpenPanelDomain(ctx, a.Cache)
	useHTTPS := forceDomain != "" && sysinfo.HasSSL(ctx, a.Cache, forceDomain)

	protocol, domain := "http", sysinfo.FetchPublicIP(ctx, a.Cache)
	if useHTTPS {
		protocol, domain = "https", forceDomain
	}

	targetURL := protocol + "://" + domain + ":" + adminPort + "/send_email"
	subject := message
	if idx := strings.Index(message, "\n"); idx != -1 {
		subject = message[:idx]
	}

	form := url.Values{
		"transient": {token},
		"recipient": {currentEmail},
		"subject":   {subject},
		"body":      {message},
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.PostForm(targetURL, form)
	if err != nil {
		log.Printf("NOTIFICATIONS - Error sending notification to %s: %v", targetURL, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("NOTIFICATIONS - Notification email sent successfully to user: %s", username)
	} else {
		log.Printf("NOTIFICATIONS - Failed to send notification. Status code: %d", resp.StatusCode)
	}
}
