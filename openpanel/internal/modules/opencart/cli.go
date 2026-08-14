package opencart

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

// openCartRequestParams pulls the domain/docroot query params every
// handler in this file needs, splits the main domain out of a possible
// subdirectory suffix, verifies ownership, and resolves the PHP container -
// shared by cache/logs/login, mirroring drupal/drush.go's
// drushRequestParams and joomla/cli.go's joomlaRequestParams.
func openCartRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
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

// handleOpenCartCacheClean deletes every cache.* file under
// system/storage/cache/ - OpenCart has no CLI cache-clear command, this is
// exactly what the admin "Refresh Cache" tool button does under the hood.
func handleOpenCartCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := openCartRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`find "$1/system/storage/cache" -maxdepth 1 -name 'cache.*' -delete`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared OpenCart cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleOpenCartLogs returns the tail of system/storage/logs/error.log -
// the single file OpenCart logs PHP warnings/errors and admin activity to.
func handleOpenCartLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, docroot, phpContainer, ok := openCartRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`tail -n 300 "$1/system/storage/logs/error.log" 2>/dev/null`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		_, _ = w.Write([]byte("No log entries yet - OpenCart only writes here when a PHP warning/error occurs."))
		return
	}
	_, _ = w.Write(out)
}

// handleOpenCartLogin generates a one-time admin login link. Like Joomla,
// OpenCart core ships no CLI command for this, so this mirrors that
// module's approach: a small token table (created here lazily, isolated
// from OpenCart's own schema) plus a login helper PHP file deployed into
// the docroot at install time (see openpanel-login.php below) that
// verifies the token then binds an admin user to the OpenCart session
// through the CMS's own Session library, exactly the way
// admin/controller/common/login.php does after a real password check.
func handleOpenCartLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dbInfo := extractOpenCartDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `oc_openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT user_id FROM `oc_user` WHERE status = 1 ORDER BY user_id ASC LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No active admin account found"})
		return
	}
	userIDStr := toStringCell(rows[0][0])

	token := generateRandomString(32)
	tokenHash := sha256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `oc_openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
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
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for OpenCart admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
