package mediawiki

import (
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleMediaWikiLogin generates a one-time admin login link - unlike Drupal's `drush uli`, MediaWiki core ships no CLI command for this, so this mirrors joomla/cli.go's handleJoomlaLogin: a small token table (created here lazily, isolated from MediaWiki's own schema) plus a login helper PHP file deployed into the docroot at install time (see login_php.go) that verifies the token then binds an admin User to the request's session through MediaWiki's own User::setCookies() API
func handleMediaWikiLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dbInfo := extractMediaWikiDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]
	prefix := dbInfo["database_prefix"]

	// the token table is created and named with MediaWiki's own configured table prefix (e.g. "mw_openpanel_login_tokens"), not because it's a MediaWiki-managed table but because login_php.go reads it back through MediaWiki's own query builder, which auto-prepends $wgDBprefix to every bare table name
	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `"+prefix+"openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT ug_user FROM `"+prefix+"user_groups` WHERE ug_group = 'sysop' LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No active administrator (sysop) account found"})
		return
	}
	userIDStr := mysqlmanager.ToString(rows[0][0])

	token := appkit.RandomString(32)
	tokenHash := appkit.SHA256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `"+prefix+"openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
			tokenHash+"', "+userIDStr+", UNIX_TIMESTAMP() + "+strconv.Itoa(ttlSeconds)+")", dbName)
	if insErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": insErr.Error()})
		return
	}

	loginLink := "https://" + domain + "/" + openpanelLoginFileName + "?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for MediaWiki admin: "+maskedLink, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
