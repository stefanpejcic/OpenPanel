package emails

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	masterUser              = "openpanel"
	mailserverContainerName = "openadmin_mailserver"
	roundcubeContainerName  = "openadmin_roundcube"
	golubContainerName      = "openadmin_golub"
	// shared with the golub container through golub.env, opencli makes it when switching to golub
	golubSSOKeyPath = "/etc/openpanel/openpanel/golub_sso.key"
	// golub refuses tokens that live longer than a minute
	golubTokenTTL = 30 * time.Second
	// its own key, the panel's secret.key signs sessions and must never end up in the roundcube container
	dovecotMasterPassPath = "/etc/openpanel/openpanel/dovecot_master.key"
)

var (
	masterPassOnce sync.Once
	masterPass     string
)

// loadDovecotMasterPass mirrors load_dovecot_master_pass().
func loadDovecotMasterPass() string {
	masterPassOnce.Do(func() {
		data, err := tenantfs.ReadFile(dovecotMasterPassPath)
		if errors.Is(err, os.ErrNotExist) {
			data, err = createDovecotMasterPass(dovecotMasterPassPath)
		}
		if err != nil {
			log.Printf("WEBMAIL - Could not read dovecot master password file %s: %v", dovecotMasterPassPath, err)
			return
		}
		masterPass = strings.TrimSpace(string(data))
	})
	return masterPass
}

// createDovecotMasterPass makes a random key on first use, O_EXCL so two starts can't write different ones
func createDovecotMasterPass(path string) ([]byte, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	key := []byte(hex.EncodeToString(buf))
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o600)
	if errors.Is(err, os.ErrExist) {
		return tenantfs.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Write(key); err != nil {
		return nil, err
	}
	return key, nil
}

func isIPv4(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if p == "" || len(p) > 3 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

// GetWebmailDomain mirrors get_webmail_domain(), memoized 1h per username.
func GetWebmailDomain(ctx context.Context, a *appctx.App, currentUsername string) string {
	domain, _ := cache.Memoize(ctx, a.Cache, "get_webmail_domain:"+currentUsername, time.Hour, func() (string, error) {
		out, err := exec.CommandContext(ctx, "opencli", "email-webmail").Output()
		if err != nil {
			log.Printf("WEBMAIL - Error executing command: %v", err)
			return "", nil //nolint:nilerr // command failure -> empty, not an error response
		}
		output := strings.TrimSpace(string(out))

		if output != "" {
			if strings.Contains(output, "OpenPanel Community edition") {
				return "http://localhost", nil
			}
			if isIPv4(output) {
				return "http://" + output + ":8080/", nil
			}
			return "https://" + output + "/", nil
		}

		if currentUsername != "" {
			if ip, ok := appctx.ReadDedicatedIPFromFile(currentUsername); ok {
				return "http://" + ip + ":8080/", nil
			}
			ip := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
			if ip != "" {
				return "http://" + ip + ":8080/", nil
			}
			return "http://localhost:8080/", nil
		}

		return "http://localhost:8080/", nil
	})
	return domain
}

// runningWebmail says which webmail client runs, roundcube, golub or "" for none
func runningWebmail(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "podman", "ps", "--format", "{{.Names}}").Output()
	if err != nil {
		log.Printf("WEBMAIL - Error checking Podman containers: %v", err)
		return ""
	}
	return webmailFromNames(string(out))
}

func webmailFromNames(names string) string {
	for _, n := range strings.Split(names, "\n") {
		switch strings.TrimSpace(n) {
		case golubContainerName:
			return "golub"
		case roundcubeContainerName:
			return "roundcube"
		}
	}
	return ""
}

// isWebmailRunning mirrors _is_webmail_running().
func isWebmailRunning(ctx context.Context) bool {
	return runningWebmail(ctx) != ""
}

// webmailAutologin returns the url that logs the browser into email, and the token in it
func webmailAutologin(ctx context.Context, client, webmailURL, email string) (string, string, error) {
	base := strings.TrimRight(webmailURL, "/")
	if client == "golub" {
		secret, err := tenantfs.ReadFile(golubSSOKeyPath)
		if err != nil {
			return "", "", err
		}
		token, err := golubSSOToken(strings.TrimSpace(string(secret)), email, time.Now())
		if err != nil {
			return "", "", err
		}
		return base + "/sso?token=" + token, token, nil
	}
	token, err := createWebmailToken(ctx, email)
	if err != nil {
		return "", "", err
	}
	return base + "/autologin.php?token=" + token, token, nil
}

// golubSSOToken signs email|expires|nonce|lang|theme the way golub verifies it, golub then logs in as email*openpanel
func golubSSOToken(secret, email string, now time.Time) (string, error) {
	if len(secret) < 32 {
		return "", errors.New("golub SSO key is missing or too short")
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := strings.Join([]string{email, strconv.FormatInt(now.Add(golubTokenTTL).Unix(), 10), hex.EncodeToString(nonce), "", ""}, "|")
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(payload))
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(payload)) + "." + enc.EncodeToString(m.Sum(nil)), nil
}

// mailserverExistsAndRunning mirrors mailserver_exists_and_running().
func mailserverExistsAndRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "podman", "ps", "-a", "--format", "{{.Names}}").Output()
	if err != nil {
		log.Printf("WEBMAIL - Error checking mailserver container status: %v", err)
		return false
	}
	found := false
	for _, name := range strings.Split(string(out), "\n") {
		if name == mailserverContainerName {
			found = true
			break
		}
	}
	if !found {
		log.Printf("WEBMAIL - Container '%s' does not exist.", mailserverContainerName)
		return false
	}

	stateOut, err := exec.CommandContext(ctx, "podman", "inspect", "-f", "{{.State.Running}}", mailserverContainerName).Output()
	if err != nil {
		log.Printf("WEBMAIL - Error checking mailserver container status: %v", err)
		return false
	}
	return strings.ToLower(strings.TrimSpace(string(stateOut))) == "true"
}

// ensureMasterUser mirrors ensure_master_user(), run once at startup (from Register(), not an unconditional package init, so importing this package in tests doesn't shell out to podman)
func ensureMasterUser(ctx context.Context) bool {
	log.Println("WEBMAIL - Ensuring dovecot master account exists for webmail autologon..")

	pass := loadDovecotMasterPass()
	if pass == "" {
		log.Println("WEBMAIL - No dovecot master password available. Skipping dovecot setup.")
		return false
	}

	if !mailserverExistsAndRunning(ctx) {
		log.Println("WEBMAIL - Mailserver is not available. Skipping dovecot setup.")
		return false
	}

	// a fresh mailserver (or one still starting up, e.g. right after boot) has no dovecot-masters.cf file yet, so `list` itself errors out - that's not a real failure, it just means no master user exists yet and `add` still needs to run
	out, listErr := exec.CommandContext(ctx, "podman", "exec", mailserverContainerName, "setup", "dovecot-master", "list").Output()
	exists := listErr == nil && strings.Contains(string(out), masterUser)

	// every update writes a new salt, the mailserver then regenerates all accounts and chown -Rs every mailbox, so skip it when the password already matches
	if exists && masterPassMatches(ctx, pass) {
		log.Println("WEBMAIL - Dovecot master user is already set up.")
		return true
	}

	var setupErr error
	action, actionPast := "add", "added"
	if exists {
		action, actionPast = "update", "updated"
		setupErr = exec.CommandContext(ctx, "podman", "exec", mailserverContainerName, "setup", "dovecot-master", "update", masterUser, pass).Run()
	} else {
		setupErr = exec.CommandContext(ctx, "podman", "exec", mailserverContainerName, "setup", "dovecot-master", "add", masterUser, pass).Run()
	}
	if setupErr != nil {
		log.Printf("WEBMAIL - Error configuring dovecot master user (%s): %v", action, setupErr)
		return false
	}
	log.Printf("WEBMAIL - Dovecot master user %s successfully.", actionPast)
	return true
}

// masterPassMatches checks the stored master hash with doveadm, which only reads
func masterPassMatches(ctx context.Context, pass string) bool {
	script := `h=$(grep "^` + masterUser + `|" /tmp/docker-mailserver/dovecot-masters.cf | cut -d"|" -f2-); [ -n "$h" ] && doveadm pw -t "$h" -p "$1" >/dev/null 2>&1`
	return exec.CommandContext(ctx, "podman", "exec", mailserverContainerName, "sh", "-c", script, "sh", pass).Run() == nil
}

// createWebmailToken mirrors _create_webmail_token().
func createWebmailToken(ctx context.Context, email string) (string, error) {
	token := randomURLToken(32) // secrets.token_urlsafe(32)

	payload := map[string]any{
		"email":     email,
		"imap_user": email + "*" + masterUser,
		"imap_pass": loadDovecotMasterPass(),
		"expires":   time.Now().UTC().Add(30 * time.Second).Format("2006-01-02T15:04:05.000000"),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	tokenFile := "/tmp/wmt_" + token

	// autologin.php only reads the file, so tokens nobody used are swept here
	// the payload holds the dovecot master password, so it goes in on stdin instead of the command line where ps would show it
	script := "find /tmp -maxdepth 1 -name 'wmt_*' -mmin +2 -delete 2>/dev/null; umask 077 && cat > \"$1\" && chown www-data:www-data \"$1\""
	c := exec.CommandContext(ctx, "podman", "exec", "-i", roundcubeContainerName, "sh", "-c", script, "sh", tokenFile)
	c.Stdin = bytes.NewReader(payloadBytes)
	if err := c.Run(); err != nil {
		return "", err
	}
	return token, nil
}

// handleWebmailLogin mirrors webmail_login().
func handleWebmailLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	client := runningWebmail(ctx)
	if client == "" {
		web.FlashRedirect(a, w, r, "info", "Webmail service is not yet started, please contact Administrator to enable it.", "/emails")
		return
	}

	webmailURL := GetWebmailDomain(ctx, a, currentUsername)

	email := r.PathValue("email")
	if email == "" {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "accessed webmail login", ipAddress)
		http.Redirect(w, r, webmailURL, http.StatusFound)
		return
	}

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	_, domain, _ := strings.Cut(email, "@")
	if !userDomains[domain] {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if loadDovecotMasterPass() == "" {
		http.Redirect(w, r, webmailURL, http.StatusFound)
		return
	}

	target, _, tokenErr := webmailAutologin(ctx, client, webmailURL, email)
	if tokenErr != nil {
		log.Printf("WEBMAIL - autologin for %s: %v", client, tokenErr)
		http.Redirect(w, r, webmailURL, http.StatusFound)
		return
	}
	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "accessed webmail for "+email, ipAddress)
	http.Redirect(w, r, target, http.StatusFound)
}
