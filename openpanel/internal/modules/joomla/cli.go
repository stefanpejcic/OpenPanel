package joomla

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

// joomlaRequestParams pulls the domain/docroot query params every
// CLI-backed handler in this file needs, splits the main domain out of a
// possible subdirectory suffix, verifies ownership, and resolves the PHP
// container to exec `cli/joomla.php` inside - shared by cache/logs so each
// handler only has its own subcommand to worry about. (Login doesn't use
// this - it talks to the DB directly and hands the browser a URL instead of
// running a CLI command.)
func joomlaRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
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

// handleJoomlaCacheClean runs `cli/joomla.php cache:clean` - Joomla's own
// system cache flush command.
func handleJoomlaCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := joomlaRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		docroot+"/cli/joomla.php", "cache:clean", "-n")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cli/joomla.php cache:clean failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Joomla cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleJoomlaLogs returns the contents of administrator/logs/*.php (the
// only logging Joomla core does by default - there's no watchdog-style
// activity log the way Drupal has one built in), as plain text, with each
// file's `<?php die('Forbidden.'); ?>` security header line stripped.
func handleJoomlaLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := joomlaRequestParams(ctx, a, r, userID, userContext)
	_ = domain
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`cd "$1/administrator/logs" 2>/dev/null && for f in *.php; do [ -e "$f" ] || continue; echo "=== $f ==="; tail -n 200 "$f" | grep -v "^<?php die"; done`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		_, _ = w.Write([]byte("No log entries yet - Joomla only writes here when a PHP warning/error occurs."))
		return
	}
	_, _ = w.Write(out)
}

// handleJoomlaLogin generates a one-time admin login link. Unlike Drupal's
// drush user:login, Joomla core ships no CLI command for this, so this
// mirrors WordPress's approach instead: a small token table (created here
// lazily, isolated from Joomla's own schema) plus a login helper PHP file
// deployed into the docroot at install time (see openpanel-login.php below)
// that verifies the token then binds an admin User to the Joomla session
// through the CMS's own Session/User APIs.
func handleJoomlaLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dbInfo := extractJoomlaDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]
	prefix := dbInfo["database_prefix"]

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `"+prefix+"openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT u.id FROM `"+prefix+"users` u "+
			"JOIN `"+prefix+"user_usergroup_map` m ON m.user_id = u.id "+
			"WHERE u.block = 0 AND m.group_id = 8 LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No active Super User account found"})
		return
	}
	userIDStr := toStringCell(rows[0][0])

	token := generateRandomString(32)
	tokenHash := sha256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `"+prefix+"openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
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
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for Joomla admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
