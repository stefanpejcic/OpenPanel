package prestashop

import (
	"context"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// prestashopRequestParams pulls the domain/docroot query params every
// handler in this file needs, splits the main domain out of a possible
// subdirectory suffix, verifies ownership, and resolves the PHP container -
// shared by cache/logs/login, mirroring opencart/nextcloud's
// {opencart,nextcloud}RequestParams.
func prestashopRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
	domain = r.URL.Query().Get("domain")
	docroot = r.URL.Query().Get("docroot")
	if domain == "" || docroot == "" {
		return "", "", "", false
	}

	mainDomain := domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		mainDomain = domain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, mainDomain) {
		return "", "", "", false
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, mainDomain)
	phpContainer = webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + phpVersion
	}
	return domain, docroot, phpContainer, true
}

// handlePrestashopCacheClean clears PrestaShop's Symfony prod cache via its
// bundled console (confirmed live: `php bin/console cache:clear --env=prod`
// works against a real install and is the standard, documented way -
// PrestaShop has no separate admin-tab-only cache mechanism beyond this).
func handlePrestashopCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := prestashopRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/bin/console", "cache:clear", "--env=prod", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared PrestaShop cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handlePrestashopLogs tails the newest file under var/logs/ (PrestaShop's
// Symfony logger writes one file per environment per day, e.g.
// prod-2026-08-15.log - there's no single fixed filename the way
// OpenCart/Nextcloud have, so this picks whichever sorts newest by mtime).
func handlePrestashopLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, docroot, phpContainer, ok := prestashopRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`f=$(ls -t "$1"/var/logs/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		_, _ = w.Write([]byte("No log entries yet."))
		return
	}
	_, _ = w.Write(out)
}

// handlePrestashopLogin generates a one-time admin login link. PrestaShop
// core ships no CLI command for this, so this mirrors joomla/opencart's
// approach: a small token table (created here lazily, isolated from
// PrestaShop's own schema) plus a login helper PHP file deployed into the
// (randomly-named) admin directory at install time (see login_php.go) that
// verifies the token then binds the admin employee to a real session
// through the same cookie fields PrestaShop's own login controller sets
// after a successful password check.
func handlePrestashopLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.URL.Query().Get("domain")
	docroot := r.URL.Query().Get("docroot")
	if domain == "" || docroot == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required"})
		return
	}
	mainDomain := domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		mainDomain = domain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, mainDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	dbInfo := extractPrestashopDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]
	prefix := dbInfo["database_prefix"]
	if prefix == "" {
		prefix = "ps_"
	}

	adminDir, dirErr := findAdminDir(mappedDocroot(userContext, docroot))
	if dirErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not locate the admin directory"})
		return
	}

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `"+prefix+"openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	// id_profile = 1 is PrestaShop's default seeded "SuperAdmin" profile -
	// confirmed live against a fresh 8.2.7 install's ps_employee row.
	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT id_employee FROM `"+prefix+"employee` WHERE id_profile = 1 AND active = 1 ORDER BY id_employee ASC LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No admin account found"})
		return
	}
	employeeIDStr := toStringCell(rows[0][0])

	token := generateRandomString(32)
	tokenHash := sha256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `"+prefix+"openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
			tokenHash+"', '"+employeeIDStr+"', UNIX_TIMESTAMP() + "+itoa(ttlSeconds)+")", dbName)
	if insErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": insErr.Error()})
		return
	}

	loginLink := "https://" + domain + "/" + adminDir + "/" + openpanelLoginFileName + "?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for PrestaShop admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}

// mappedDocroot converts a container docroot path (/var/www/html/...) into
// its host-side equivalent under the user's html_data volume - the same
// mapping install.go computes as hostOSPath, needed here too since
// findAdminDir reads the directory listing straight off disk rather than
// through podman exec.
func mappedDocroot(userContext, docroot string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/")
}
