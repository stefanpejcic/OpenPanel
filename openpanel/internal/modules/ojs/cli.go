package ojs

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

// ojsRequestParams pulls the domain/docroot query params every CLI-backed
// handler in this file needs, splits the main domain out of a possible
// subdirectory suffix, verifies ownership, and resolves the PHP container -
// mirrors moodle/cli.go's moodleRequestParams.
func ojsRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
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

// ojsApprootContainerPath returns the approot's container-visible path for
// a given site (domain) - docroot itself is a symlink into it (see ojs.go's
// package doc comment), but tools/*.php scripts live at its root, so
// cache-clear/logs/update all need this path instead of docroot.
func ojsApprootContainerPath(domain string) string {
	return "/var/www/html/" + siteSlug(domain) + "_ojsapp"
}

func ojsFilesContainerPath(domain string) string {
	return "/var/www/html/" + siteSlug(domain) + "_ojsfiles"
}

// handleOJSCacheClean clears OJS's on-disk template/data cache (the "cache/"
// directory under the approot - OJS ships no dedicated CLI cache-purge
// command the way Moodle/Joomla do, so this just empties the directory
// directly, the same technique the PKP community documents for
// troubleshooting a stuck install).
func handleOJSCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := ojsRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	approot := ojsApprootContainerPath(domain)
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`find "$1/cache" -mindepth 1 -not -name index.html -exec rm -rf {} +`, "sh", approot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared OJS cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Caches purged successfully."})
}

// handleOJSLogs tails any *.log file it can find under the approot's cache/
// directory or the files directory - OJS ships no dedicated flat
// application log file by default (log_errors goes to the PHP SAPI's own
// configured error_log, not a per-site file), so this surfaces whatever a
// plugin/PHP itself may have written there instead, consistent in spirit
// with every other module's Logs tab, with an honest fallback message when
// nothing is found.
func handleOJSLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := ojsRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	approot := ojsApprootContainerPath(domain)
	filesDir := ojsFilesContainerPath(domain)
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`f=$(ls -t "$1"/cache/*.log "$2"/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "sh", approot, filesDir)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		_, _ = w.Write([]byte("No log file found. OJS logs PHP-level errors through PHP's own configured error_log, not a per-site file - check the PHP container's own logs from Services for PHP-level errors."))
		return
	}
	_, _ = w.Write(out)
}

// handleOJSLogin generates a one-time site-admin login link. Like Joomla,
// OJS core ships no CLI command for this, so this mirrors
// joomla/cli.go's handleJoomlaLogin: a small token table (created here
// lazily, isolated from OJS's own schema) plus a login helper PHP file
// deployed into the approot at install time (see login_php.go) that
// verifies the token then binds a site-admin User to the OJS session
// through PKP\security\Validation's own registerUserSession() helper.
func handleOJSLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dbInfo := extractOJSDatabaseInfoForLogin(userContext, domain)
	if dbInfo["error"] != "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	// ROLE_ID_SITE_ADMIN = 1 (lib/pkp/classes/security/Role.php); a user's
	// roles live in user_user_groups -> user_groups.role_id, not a simple
	// is_admin column (confirmed by reading
	// lib/pkp/classes/migration/install/RolesAndUserGroupsMigration.php and
	// CommonMigration.php directly).
	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT u.user_id FROM `users` u "+
			"JOIN `user_user_groups` uug ON uug.user_id = u.user_id "+
			"JOIN `user_groups` ug ON ug.user_group_id = uug.user_group_id "+
			"WHERE u.disabled = 0 AND ug.role_id = 1 LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No active site administrator account found"})
		return
	}
	userIDStr := toStringCell(rows[0][0])

	token := generateRandomString(32)
	tokenHash := sha256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
			tokenHash+"', "+userIDStr+", UNIX_TIMESTAMP() + "+itoa(ttlSeconds)+")", dbName)
	if insErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": insErr.Error()})
		return
	}

	loginLink := "https://" + domain + "/" + openpanelLoginFileName + "?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for OJS site admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
