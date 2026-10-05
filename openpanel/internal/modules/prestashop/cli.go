package prestashop

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

// handlePrestashopCacheClean clears PrestaShop's Symfony prod cache via its bundled console (`php bin/console cache:clear --env=prod`), the standard documented way since PrestaShop has no other cache mechanism
func handlePrestashopCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/bin/console", "cache:clear", "--env=prod", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared PrestaShop cache for "+domain, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handlePrestashopLogin generates a one-time admin login link, mirrors joomla/opencart's approach with a lazily-created token table plus the login helper PHP deployed into the randomly-named admin directory at install time (see login_php.go)
func handlePrestashopLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dbInfo := extractPrestashopDatabaseInfoForLogin(userContext, docroot)
	if dbInfo["error"] != "" {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": dbInfo["error"]})
		return
	}
	dbName := dbInfo["database_name"]
	prefix := dbInfo["database_prefix"]
	if prefix == "" {
		prefix = "ps_"
	}

	adminDir, dirErr := findAdminDir(mappedDocroot(userContext, docroot))
	if dirErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not locate the admin directory"})
		return
	}

	_, _ = mysqlmanager.Exec(ctx, userContext,
		"CREATE TABLE IF NOT EXISTS `"+prefix+"openpanel_login_tokens` ("+
			"token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL)", dbName)

	// id_profile = 1 is PrestaShop's default seeded "SuperAdmin" profile
	rows, queryErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT id_employee FROM `"+prefix+"employee` WHERE id_profile = 1 AND active = 1 ORDER BY id_employee ASC LIMIT 1", dbName)
	if queryErr != nil || len(rows) == 0 {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No admin account found"})
		return
	}
	employeeIDStr := mysqlmanager.ToString(rows[0][0])

	token := appkit.RandomString(32)
	tokenHash := appkit.SHA256Hex(token)
	const ttlSeconds = 600
	_, insErr := mysqlmanager.Exec(ctx, userContext,
		"INSERT INTO `"+prefix+"openpanel_login_tokens` (token_hash, user_id, expires) VALUES ('"+
			tokenHash+"', '"+employeeIDStr+"', UNIX_TIMESTAMP() + "+strconv.Itoa(ttlSeconds)+")", dbName)
	if insErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": insErr.Error()})
		return
	}

	loginLink := "https://" + domain + "/" + adminDir + "/" + openpanelLoginFileName + "?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for PrestaShop admin: "+maskedLink, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}

// mappedDocroot converts a container docroot path into its host-side equivalent under the user's html_data volume - same mapping install.go computes as hostOSPath, needed here since findAdminDir reads the directory listing straight off disk rather than through podman exec
func mappedDocroot(userContext, docroot string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/")
}
