package emails

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

const (
	masterUser              = "openpanel"
	mailserverContainerName = "openadmin_mailserver"
	roundcubeContainerName  = "openadmin_roundcube"
	dovecotMasterPassPath   = "/etc/openpanel/openpanel/secret.key"
)

var (
	masterPassOnce sync.Once
	masterPass     string
)

// loadDovecotMasterPass mirrors load_dovecot_master_pass().
func loadDovecotMasterPass() string {
	masterPassOnce.Do(func() {
		data, err := os.ReadFile(dovecotMasterPassPath)
		if err != nil {
			log.Printf("WEBMAIL - Could not read dovecot master password file %s: %v", dovecotMasterPassPath, err)
			return
		}
		masterPass = strings.TrimSpace(string(data))
	})
	return masterPass
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

// isWebmailRunning mirrors _is_webmail_running().
func isWebmailRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "podman", "ps", "--format", "{{.Names}}").Output()
	if err != nil {
		log.Printf("WEBMAIL - Error checking Podman containers: %v", err)
		return false
	}
	return strings.Contains(string(out), roundcubeContainerName)
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

// ensureMasterUser mirrors ensure_master_user(), run once at startup (from
// Register(), not an unconditional package init, so importing this package
// in tests doesn't shell out to podman).
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

	// A fresh mailserver (or one that hasn't finished starting up yet, e.g.
	// right after boot when this runs) has no dovecot-masters.cf file at
	// all, so `list` itself errors out - that's not a real failure, it
	// just means no master user exists yet and `add` still needs to run;
	// only errors from the actual add/update below are worth bailing on.
	out, listErr := exec.CommandContext(ctx, "podman", "exec", mailserverContainerName, "setup", "dovecot-master", "list").Output()
	exists := listErr == nil && strings.Contains(string(out), masterUser)

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
	encoded := base64.StdEncoding.EncodeToString(payloadBytes)
	tokenFile := "/tmp/wmt_" + token

	cmd := "echo " + encoded + " | base64 -d > " + tokenFile + " && chown www-data:www-data " + tokenFile
	if err := exec.CommandContext(ctx, "podman", "exec", roundcubeContainerName, "sh", "-lc", cmd).Run(); err != nil {
		return "", err
	}
	return token, nil
}

// handleWebmailLogin mirrors webmail_login().
func handleWebmailLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	if !isWebmailRunning(ctx) {
		flashAndRedirect(a, w, r, "info", "Webmail service is not yet started, please contact Administrator to enable it.", "/emails")
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

	token, tokenErr := createWebmailToken(ctx, email)
	if tokenErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "accessed webmail for "+email, ipAddress)
	http.Redirect(w, r, strings.TrimRight(webmailURL, "/")+"/autologin.php?token="+token, http.StatusFound)
}
