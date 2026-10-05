package nextcloud

import (
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleNextcloudCacheClean clears the preview cache and distributed memcache if configured - no single "clear all" CLI command exists and APCu can't be cleared without restarting php-fpm, so this is what's actually safe to act on
func handleNextcloudCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`rm -rf "$1"/data/appdata_*/preview/* 2>/dev/null; php "$1/occ" memcache:distributed:clear 2>/dev/null; true`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Nextcloud cache for "+domain, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleNextcloudLogin generates a one-time admin login link, mirrors joomla/opencart's approach with a lazily-created token table plus the login helper PHP deployed at install time (see login_php.go)
func handleNextcloudLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.URL.Query().Get("domain")
	docroot := r.URL.Query().Get("docroot")
	if domain == "" || docroot == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required"})
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

	dbInfo := extractNextcloudDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]
	prefix := dbInfo["database_prefix"]
	if prefix == "" {
		prefix = "oc_"
	}

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `"+prefix+"openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id VARCHAR(64) NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT uid FROM `"+prefix+"group_user` WHERE gid = 'admin' ORDER BY uid ASC LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No admin account found"})
		return
	}
	userIDStr := mysqlmanager.ToString(rows[0][0])

	token := appkit.RandomString(32)
	tokenHash := appkit.SHA256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `"+prefix+"openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
			tokenHash+"', '"+userIDStr+"', UNIX_TIMESTAMP() + "+strconv.Itoa(ttlSeconds)+")", dbName)
	if insErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": insErr.Error()})
		return
	}

	loginLink := "https://" + domain + "/" + openpanelLoginFileName + "?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for Nextcloud admin: "+maskedLink, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
