package account

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

const notifyConfigFilePath = "/etc/openpanel/openpanel/conf/openpanel.config"

// userEmail is a notification for the panel user, OpenAdmin shows Details as a table with bold values and Tips under it
type userEmail struct {
	Subject, Text, Tips string
	Details             []emailDetail
}

type emailDetail struct {
	Label string `json:"label"`
	Value string `json:"value"`
	URL   string `json:"url,omitempty"`
	Flag  string `json:"flag,omitempty"`
}

// securityEmail adds when, where and from what browser, so the user can tell if it was them
func securityEmail(a *appctx.App, r *http.Request, subject, text, ifNotYou string) userEmail {
	ip := reqip.ClientIP(r)
	country := getCountryCode(a, r.Context(), ip)
	// OpenAdmin shows the panel's /static/flags/<code>.png next to a 2 letter country code
	flag := ""
	if len(country) == 2 {
		flag = strings.ToLower(country)
	}
	return userEmail{Subject: subject, Text: text, Tips: ifNotYou, Details: []emailDetail{
		{Label: "Time", Value: time.Now().Format("2006-01-02 15:04:05 MST")},
		{Label: "IP address", Value: ip},
		{Label: "Country", Value: country, Flag: flag},
		{Label: "Browser", Value: browserFromUserAgent(r.UserAgent())},
	}}
}

// notificationsFilePath is the one key=value-per-line preferences file per user.
func notificationsFilePath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/notifications.yaml"
}

// parseNotificationsFile is a simple "k=v" line parser, distinct from notifications.go's own parser (which also skips "#" comment lines)
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

// notificationDefault is used for keys the file doesn't have, username changes aren't on the page so they stay on
func notificationDefault(key string) int {
	for _, d := range notificationDefs {
		if d.Key == key {
			return int(d.Default[0] - '0')
		}
	}
	if key == "notify_username_change" {
		return 1
	}
	return 0
}

// getNotificationPreference reads one key from a user's notifications.yaml-shaped preferences file, cached for 5 minutes
func getNotificationPreference(ctx context.Context, a *appctx.App, username, key string) int {
	cacheKey := "get_from_file_value:" + username + ":" + key
	v, _ := cache.Memoize(ctx, a.Cache, cacheKey, 300*time.Second, func() (int, error) {
		content, err := os.ReadFile(notificationsFilePath(username))
		if err != nil {
			return notificationDefault(key), nil
		}
		prefs := parseNotificationsFile(string(content))
		if raw, ok := prefs[key]; ok {
			if n, convErr := strconv.Atoi(raw); convErr == nil {
				return n, nil
			}
		}
		return notificationDefault(key), nil
	})
	return v
}

// loginIPIsKnown reports whether ip is already in the user's .lastlogin history, call it before the new login is appended
func loginIPIsKnown(username, ip string) bool {
	data, err := os.ReadFile("/etc/openpanel/openpanel/core/users/" + username + "/.lastlogin")
	if err != nil || ip == "" {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "IP: "+ip+" -") {
			return true
		}
	}
	return false
}

// notifyLogin sends the login email, for an IP seen before only when notify_account_login_for_known_netblock is on
func notifyLogin(a *appctx.App, ctx context.Context, userID int, username string, message userEmail, knownIP bool) {
	if knownIP && getNotificationPreference(ctx, a, username, "notify_account_login_for_known_netblock") != 1 {
		return
	}
	checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_account_login", message)
}

// checkIfUserShouldBeNotified fires an async notification email if "notifications" is enabled and the user opted into key (or key is the always-on "notify_always"). username is passed by the caller rather than re-derived, since some callers (like the username-change flow) already know a value that may not match InjectData's cached current_username yet.
func checkIfUserShouldBeNotified(a *appctx.App, ctx context.Context, userID int, username, key string, message userEmail) {
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

// generateRandomTokenOnce rewrites the *existing* mail_security_token= line in openpanel.config in place - a no-op if that line isn't already there, never appends a missing key
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

// notifyUserOfChange looks up the recipient email if not already known, then POSTs to the local "openadmin" service's /send_email; meant to be called via `go notifyUserOfChange(...)`
func notifyUserOfChange(a *appctx.App, username string, message userEmail, currentEmail string) {
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
	details, _ := json.Marshal(message.Details)
	form := url.Values{
		"transient": {token},
		"recipient": {currentEmail},
		"subject":   {message.Subject},
		"body":      {message.Text},
		"type":      {"user"},
		"details":   {string(details)},
		"tips":      {message.Tips},
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
