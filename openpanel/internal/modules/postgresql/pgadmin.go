package postgresql

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

const (
	sharedPGAContainer = "pgadmin"
	pgaInternalURL     = "http://127.0.0.1:8889"
)

// getPGABaseURL resolves the externally-reachable pgAdmin URL for the
// current user, preferring the panel's forced domain (with SSL) over the
// server's bare IP.
func getPGABaseURL(ctx context.Context, a *appctx.App, currentUsername string) string {
	if a.ForceDomain != "" {
		dynamicIP := sysinfo.FetchPublicIP(ctx, a.Cache)
		if a.ForceDomain != dynamicIP && sysinfo.HasSSL(ctx, a.Cache, a.ForceDomain) {
			return "https://" + a.ForceDomain + ":2054"
		}
	}
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
	return "http://" + serverIP + ":8889"
}

func isPGAContainerRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "podman", "inspect", "-f", "{{.State.Running}}", sharedPGAContainer).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(strings.ToLower(string(out))) == "true"
}

// pgaCookie extracts a named cookie from the client's cookie jar.
func pgaCookie(jar *cookiejar.Jar, name string) string {
	u, _ := url.Parse(pgaInternalURL)
	for _, c := range jar.Cookies(u) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// pgaLogin authenticates against the shared pgAdmin container's own REST
// API and returns the cookie jar carrying its session, or nil on failure.
func pgaLogin(ctx context.Context, email, password string) *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, pgaInternalURL+"/login", nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(getReq)
	if err != nil {
		log.Printf("PGADMIN - login request failed for %s: %v", email, err)
		return nil
	}
	resp.Body.Close()

	csrfToken := pgaCookie(jar, "CSRF-Token")

	form := url.Values{"email": {email}, "password": {password}, "csrf_token": {csrfToken}}
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pgaInternalURL+"/authenticate/login", strings.NewReader(form.Encode()))
	if err != nil {
		return nil
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if csrfToken != "" {
		postReq.Header.Set("X-CSRFToken", csrfToken)
	}
	authResp, err := client.Do(postReq)
	if err != nil {
		log.Printf("PGADMIN - login request failed for %s: %v", email, err)
		return nil
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK || pgaCookie(jar, "pga4_session") == "" {
		log.Printf("PGADMIN - login failed for %s: HTTP %d", email, authResp.StatusCode)
		return nil
	}
	return jar
}

// ensurePGAdminAccount idempotently provisions this user's pgAdmin login
// and their Postgres server entry (connected over their own unix socket).
// Returns an authenticated cookie jar, or nil on failure.
func ensurePGAdminAccount(ctx context.Context, userContext string) *cookiejar.Jar {
	email := webserver.GetEnvFileValue(userContext, "PGADMIN_MAIL")
	password := webserver.GetEnvFileValue(userContext, "PGADMIN_PW")
	pgUser := webserver.GetEnvFileValue(userContext, "POSTGRES_USER")
	if pgUser == "" {
		pgUser = "postgres"
	}
	pgPassword := webserver.GetEnvFileValue(userContext, "POSTGRES_PASSWORD")

	if email == "" || password == "" || pgPassword == "" {
		log.Printf("PGADMIN - missing PGADMIN_MAIL/PGADMIN_PW/POSTGRES_PASSWORD for %s", userContext)
		return nil
	}

	// 1) create the pgAdmin account if it doesn't exist yet - ignore
	// failures caused by the account already existing, same
	// accept-and-continue idempotency as php/phpmyadmin.go's
	// ensureUserToken().
	_ = exec.CommandContext(ctx, "podman", "exec", sharedPGAContainer, "python3", "/pgadmin4/setup.py", "add-user", email, password).Run()

	jar := pgaLogin(ctx, email, password)
	if jar == nil {
		return nil
	}
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	// 2) register their Postgres server if it isn't there yet
	socketDir := "/home/" + userContext + "/sockets/postgres"

	listReq, err := http.NewRequestWithContext(ctx, http.MethodGet, pgaInternalURL+"/browser/server_groups/1/servers/", nil)
	if err == nil {
		if listResp, listErr := client.Do(listReq); listErr == nil {
			body := new(bytes.Buffer)
			_, _ = body.ReadFrom(listResp.Body)
			listResp.Body.Close()
			alreadyRegistered := listResp.StatusCode == http.StatusOK && strings.Contains(body.String(), userContext)

			if !alreadyRegistered {
				csrfToken := pgaCookie(jar, "CSRF-Token")
				payload, _ := json.Marshal(map[string]any{
					"name": userContext, "host": socketDir, "port": 5432,
					"username": pgUser, "maintenance_db": "postgres",
					"save_password": true, "password": pgPassword, "connect_now": false,
				})
				regReq, regErr := http.NewRequestWithContext(ctx, http.MethodPost, pgaInternalURL+"/browser/server_groups/1/servers/", bytes.NewReader(payload))
				if regErr == nil {
					regReq.Header.Set("Content-Type", "application/json")
					if csrfToken != "" {
						regReq.Header.Set("X-CSRFToken", csrfToken)
					}
					if regResp, regDoErr := client.Do(regReq); regDoErr == nil {
						regResp.Body.Close()
					} else {
						log.Printf("PGADMIN - failed to register server for %s: %v", userContext, regDoErr)
					}
				}
			}
		} else {
			log.Printf("PGADMIN - failed to register server for %s: %v", userContext, listErr)
		}
	}

	return jar
}

// handlePGAdminRedirect provisions the user's pgAdmin account/session if
// needed, then redirects the browser into the shared pgAdmin instance.
func handlePGAdminRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !isPGAContainerRunning(ctx) {
		renderPGAdminUnavailablePage(a, w, r, "Please contact support.", http.StatusServiceUnavailable)
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")

	jar := ensurePGAdminAccount(ctx, userContext)
	if jar == nil {
		log.Printf("PGADMIN - could not provision pgAdmin account for context: %s", userContext)
		renderPGAdminUnavailablePage(a, w, r, "Failed to provision pgAdmin access.", http.StatusInternalServerError)
		return
	}

	pgaBaseURL := getPGABaseURL(ctx, a, currentUsername)
	pgadminURL := pgaBaseURL + "/"
	if qs := r.URL.RawQuery; qs != "" {
		pgadminURL += "?" + qs
	}

	log.Printf("PGADMIN - redirecting %s to pgAdmin: %s", userContext, pgadminURL)
	_ = logger.RecordUserAction(a.Config, currentUsername, "opened pgAdmin", reqip.ClientIP(r))

	for _, name := range []string{"pga4_session", "CSRF-Token"} {
		if v := pgaCookie(jar, name); v != "" {
			http.SetCookie(w, &http.Cookie{Name: name, Value: v, Path: "/", HttpOnly: name == "pga4_session"})
		}
	}
	http.Redirect(w, r, pgadminURL, http.StatusFound)
}
