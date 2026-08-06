package wordpress

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

const opLoginTokenTTL = 300 * time.Second

// openpanelMuPluginPHP mirrors OPENPANEL_MU_PLUGIN_PHP verbatim.
const openpanelMuPluginPHP = `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next autologin from WP Manager.
 */
add_action('init', function () {
    if (!isset($_GET['op_login'])) {
        return;
    }

    $token = isset($_GET['op_login']) ? sanitize_text_field(wp_unslash($_GET['op_login'])) : '';
    if (!preg_match('/^[a-f0-9]{64}$/', $token)) {
        wp_die('Invalid login token.', 'OpenPanel Login', ['response' => 400]);
    }

    global $wpdb;
    $option_name = 'op_login_' . hash('sha256', $token);

    // Raw query on purpose: bypasses any persistent object cache
    $row = $wpdb->get_row(
        $wpdb->prepare("SELECT option_value FROM {$wpdb->options} WHERE option_name = %s LIMIT 1", $option_name)
    );

    if (!$row) {
        wp_die('This login link is invalid or has already been used.', 'OpenPanel Login', ['response' => 403]);
    }

    // One-time use: delete immediately
    $wpdb->delete($wpdb->options, ['option_name' => $option_name]);

    $data = maybe_unserialize($row->option_value);
    if (!is_array($data) || empty($data['user_id']) || empty($data['expires'])) {
        wp_die('This login link is invalid.', 'OpenPanel Login', ['response' => 403]);
    }

    if (time() > (int) $data['expires']) {
        wp_die('This login link has expired.', 'OpenPanel Login', ['response' => 403]);
    }

    $user = get_user_by('id', (int) $data['user_id']);
    if (!$user || !user_can($user, 'manage_options')) {
        wp_die('User not found or not an administrator.', 'OpenPanel Login', ['response' => 403]);
    }

    wp_set_current_user($user->ID);
    wp_set_auth_cookie($user->ID);
    do_action('wp_login', $user->user_login, $user);

    wp_safe_redirect(admin_url());
    exit;
});
`

var (
	wpCLINameRE = regexp.MustCompile(`^[A-Za-z0-9_\-.]+$`)
	wpCLIPathRE = regexp.MustCompile(`^[A-Za-z0-9_\-/.]+$`)
)

// sanitizeName/sanitizePath/sanitizePHPVersion/sanitizeAdminUser mirror
// wp_cli()'s in-file sanitize_name()/sanitize_path()/etc. Python's versions
// call bare `abort` (a reference to the function object, not `abort(...)`) -
// a no-op that never actually rejects anything, so tainted values flow
// straight into shelled-out wp-cli/podman commands unsanitized. That's a
// real command-injection-adjacent bug in an endpoint whose whole job is to
// build shell commands from these values, not a behavior worth preserving,
// so these actually reject invalid input here.
func sanitizeName(name string) (string, bool) {
	if name == "" || !wpCLINameRE.MatchString(name) {
		return "", false
	}
	return name, true
}

func sanitizeWPCLIPath(p string) (string, bool) {
	if p == "" || !wpCLIPathRE.MatchString(p) {
		return "", false
	}
	return p, true
}

func sanitizePHPVersion(v string) (string, bool) {
	if v == "" {
		return "", true
	}
	if !wpCLINameRE.MatchString(v) {
		return "", false
	}
	return v, true
}

func sanitizeAdminUser(u string) (string, bool) {
	if u == "" || !wpCLINameRE.MatchString(u) {
		return "", false
	}
	return u, true
}

// wpCLIParams mirrors `request.get_json(silent=True) or request.form or
// request.args`: a JSON body if present, else the merged form/query values.
func wpCLIParams(r *http.Request) map[string]string {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		body, readErr := io.ReadAll(r.Body)
		if readErr == nil {
			var m map[string]any
			if json.Unmarshal(body, &m) == nil && len(m) > 0 {
				out := make(map[string]string, len(m))
				for k, v := range m {
					out[k] = fmt.Sprint(v)
				}
				return out
			}
		}
	}
	_ = r.ParseForm()
	out := make(map[string]string, len(r.Form))
	for k := range r.Form {
		out[k] = r.Form.Get(k)
	}
	return out
}

// getWPConfigDBInfo mirrors get_wp_config_db_info().
func getWPConfigDBInfo(realPath string) (dbName, tablePrefix string, err error) {
	wpConfigFile := filepath.Join(realPath, "wp-config.php")
	content, readErr := os.ReadFile(wpConfigFile)
	if readErr != nil {
		return "", "", fmt.Errorf("wp-config.php not found")
	}
	dbNameMatch := dbNameRE.FindStringSubmatch(string(content))
	tablePrefixMatch := tablePrefixValueRE.FindStringSubmatch(string(content))
	if dbNameMatch == nil || tablePrefixMatch == nil {
		return "", "", fmt.Errorf("wp-config.php not found")
	}
	return dbNameMatch[1], tablePrefixMatch[1], nil
}

type wpUser struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Registered string `json:"registered"`
	Role       string `json:"role"`
}

// getWPUsers mirrors get_wp_users(): reads users/usermeta directly, no
// wp-cli round trip.
func getWPUsers(ctx context.Context, userContext, realPath, roleFilter string) ([]wpUser, error) {
	dbName, tablePrefix, err := getWPConfigDBInfo(realPath)
	if err != nil {
		return nil, err
	}

	usersTable := tablePrefix + "users"
	metaTable := tablePrefix + "usermeta"
	capKey := tablePrefix + "capabilities"

	userRows, execErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT ID, user_login, user_email, user_registered FROM `"+usersTable+"` ORDER BY user_login", dbName)
	if execErr != nil {
		return nil, execErr
	}

	roleMap := map[string]string{}
	if len(userRows) > 0 {
		roleRows, roleErr := mysqlmanager.Exec(ctx, userContext,
			"SELECT user_id, meta_value FROM `"+metaTable+"` WHERE meta_key = '"+capKey+"'", dbName)
		if roleErr == nil {
			for _, row := range roleRows {
				uid := toStringCell(row[0])
				metaValue := toStringCell(row[1])
				role := "unknown"
				switch {
				case strings.Contains(metaValue, "administrator"):
					role = "administrator"
				case strings.Contains(metaValue, "editor"):
					role = "editor"
				case strings.Contains(metaValue, "author"):
					role = "author"
				case strings.Contains(metaValue, "contributor"):
					role = "contributor"
				case strings.Contains(metaValue, "subscriber"):
					role = "subscriber"
				}
				roleMap[uid] = role
			}
		}
	}

	var users []wpUser
	for _, row := range userRows {
		uid := toStringCell(row[0])
		role := roleMap[uid]
		if role == "" {
			role = "unknown"
		}
		if roleFilter != "" && role != roleFilter {
			continue
		}
		id, _ := strconv.Atoi(uid)
		users = append(users, wpUser{
			ID: id, Username: toStringCell(row[1]), Email: toStringCell(row[2]),
			Registered: toStringCell(row[3]), Role: role,
		})
	}
	return users, nil
}

// phpSerializeAssoc mirrors php_serialize_assoc(): only string/int/bool
// values, matching the Go call sites (user_id, expires) exactly.
func phpSerializeAssoc(d map[string]any, order []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "a:%d:{", len(d))
	for _, k := range order {
		fmt.Fprintf(&b, "s:%d:\"%s\";", len(k), k)
		switch v := d[k].(type) {
		case bool:
			if v {
				b.WriteString("b:1;")
			} else {
				b.WriteString("b:0;")
			}
		case int:
			fmt.Fprintf(&b, "i:%d;", v)
		case string:
			fmt.Fprintf(&b, "s:%d:\"%s\";", len(v), v)
		}
	}
	b.WriteString("}")
	return b.String()
}

func safeRunWPCLI(ctx context.Context, userContext string, argv []string) (stdout string, err error) {
	runCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	out, runErr := podmanmanager.Command(runCtx, userContext, argv).CombinedOutput()
	return string(out), runErr
}

// handleWPCLI mirrors wp_cli().
func handleWPCLI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	action := strings.ToLower(r.PathValue("action"))
	params := wpCLIParams(r)

	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domainParam := params["domain"]
	verifyType := params["type"]
	if verifyType == "" {
		verifyType = "core"
	}
	phpVersion := params["php_version"]
	domainDirectory := params["docroot"]
	adminUsername := params["admin_user"]

	if domainParam == "" {
		http.Error(w, "Domain is not provided.", http.StatusForbidden)
		return
	}
	domainParam, ok := sanitizeWPCLIPath(domainParam)
	if !ok {
		http.Error(w, "Invalid domain parameter.", http.StatusBadRequest)
		return
	}

	var domain, subdirectory string
	if idx := strings.Index(domainParam, "/"); idx != -1 {
		domain = domainParam[:idx]
		subdirectory = domainParam[idx+1:]
		if domain, ok = sanitizeName(domain); !ok {
			http.Error(w, "Invalid domain.", http.StatusBadRequest)
			return
		}
		if subdirectory, ok = sanitizeWPCLIPath(subdirectory); !ok {
			http.Error(w, "Invalid subdirectory.", http.StatusBadRequest)
			return
		}
	} else {
		if domain, ok = sanitizeName(domainParam); !ok {
			http.Error(w, "Invalid domain.", http.StatusBadRequest)
			return
		}
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if phpVersion == "" || domainDirectory == "" {
		var dbDocroot, dbPHPVersion string
		row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_url = ?", domain)
		var unusedDomain string
		if scanErr := row.Scan(&unusedDomain, &dbDocroot, &dbPHPVersion); scanErr != nil {
			http.Error(w, "Domain not found", http.StatusNotFound)
			return
		}
		if domainDirectory == "" {
			domainDirectory = dbDocroot
		}
		if phpVersion == "" {
			phpVersion = dbPHPVersion
		}
	}

	if subdirectory != "" && domainDirectory != "" {
		normalizedDir := filepath.Clean(domainDirectory)
		normalizedSub := filepath.Clean(subdirectory)
		if !strings.HasSuffix(normalizedDir, normalizedSub) {
			domainDirectory = filepath.Clean(filepath.Join(domainDirectory, subdirectory))
		}
	}

	domainDirectory, ok = sanitizeWPCLIPath(domainDirectory)
	if !ok {
		http.Error(w, "Invalid document root.", http.StatusBadRequest)
		return
	}

	const wwwPrefix = "/var/www/html/"
	relativeDocroot := strings.TrimPrefix(domainDirectory, wwwPrefix)
	hostosPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + relativeDocroot

	phpVersion, ok = sanitizePHPVersion(phpVersion)
	if !ok {
		http.Error(w, "Invalid php_version.", http.StatusBadRequest)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	var phpContainer string
	if isLitespeed {
		if phpContainer, ok = sanitizeName(webServer); !ok {
			http.Error(w, "Invalid webserver.", http.StatusBadRequest)
			return
		}
	} else {
		if phpVersion == "" {
			http.Error(w, "php_version could not be determined", http.StatusBadRequest)
			return
		}
		if phpContainer, ok = sanitizeName("php-fpm-" + phpVersion); !ok {
			http.Error(w, "Invalid php_version.", http.StatusBadRequest)
			return
		}
	}

	htmlVolume := filepath.Join("/home", userContext, "docker-data", "volumes", userContext+"_html_data", "_data")
	var docrootWithoutWWW string
	if strings.HasPrefix(domainDirectory, wwwPrefix) {
		docrootWithoutWWW = domainDirectory[len(wwwPrefix):]
	} else {
		docrootWithoutWWW = strings.TrimPrefix(filepath.Clean(domainDirectory), "/")
	}
	realPath := filepath.Clean(filepath.Join(htmlVolume, docrootWithoutWWW))

	switch action {
	case "login":
		handleWPCLILogin(a, w, r, currentUsername, userContext, domain, realPath, adminUsername)
		return
	case "users":
		roleFilter := r.URL.Query().Get("role")
		users, usersErr := getWPUsers(ctx, userContext, realPath, roleFilter)
		if usersErr != nil {
			status := http.StatusInternalServerError
			if usersErr.Error() == "wp-config.php not found" {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": usersErr.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "domain": domain, "filter_role": roleFilter, "count": len(users), "users": users,
		})
		return
	}

	allowedActions := map[string]map[string][]string{
		"verify-checksum":  {"core": {"core", "verify-checksums"}, "plugins": {"plugin", "verify-checksums", "--all"}},
		"download":         {"core": {"core", "download", "--force", "--skip-content"}, "plugins": {"plugin", "list", "--field=name"}},
		"shuffle-salts":    {"cmd": {"config", "shuffle-salts"}},
		"cache":            {"flush": {"cache", "flush"}, "type": {"cache", "type"}},
		"maintenance-mode": {"base": {"maintenance-mode"}},
	}

	subActions, actionOK := allowedActions[action]
	if !actionOK {
		http.Error(w, "Unsupported action", http.StatusBadRequest)
		return
	}

	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)
	baseDocker := append(append([]string{}, wpBase...), "--path="+domainDirectory, "--allow-root")

	switch action {
	case "maintenance-mode":
		fullPath := filepath.Join(hostosPath, ".maintenance")
		if r.Method == http.MethodGet {
			if _, statErr := os.Stat(fullPath); statErr == nil {
				writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
			} else {
				writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
			}
			return
		}

		act := strings.ToLower(params["subcmd"])
		if act != "enable" && act != "disable" {
			http.Error(w, "Use 'enable' or 'disable'", http.StatusBadRequest)
			return
		}
		if act == "enable" {
			_ = os.WriteFile(fullPath, []byte("<?php $upgrading = time(); ?>"), 0o644)
		} else {
			_ = os.Remove(fullPath)
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, act+"d maintenance mode", reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + act + "d successfully."})
		return

	case "cache":
		if r.Method == http.MethodPost {
			_, _ = safeRunWPCLI(ctx, userContext, append(append([]string{}, baseDocker...), subActions["flush"]...))
			_ = logger.RecordUserAction(a.Config, currentUsername, "flushed cache", reqip.ClientIP(r))
			writeJSON(w, http.StatusOK, map[string]string{"message": "Cache flushed successfully."})
			return
		}
		out, _ := safeRunWPCLI(ctx, userContext, append(append([]string{}, baseDocker...), subActions["type"]...))
		out = strings.TrimSpace(out)
		if out == "" {
			out = "none"
		}
		writeJSON(w, http.StatusOK, map[string]string{"type": out})
		return
	}

	var subcmd []string
	if action == "verify-checksum" || action == "download" {
		sc, typeOK := subActions[verifyType]
		if !typeOK {
			http.Error(w, "Invalid type", http.StatusBadRequest)
			return
		}
		subcmd = sc
	} else {
		subcmd = subActions["cmd"]
	}

	cmd := append(append([]string{}, baseDocker...), subcmd...)
	cmd = append(cmd, "--skip-themes", "--skip-plugins")
	out, _ := safeRunWPCLI(ctx, userContext, cmd)
	_ = logger.RecordUserAction(a.Config, currentUsername, "Executed command: wp "+strings.Join(subcmd, " ")+" for website "+domainParam+" using WP Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Success", "output": strings.TrimSpace(out)})
}

// handleWPCLILogin mirrors the "login" action branch of wp_cli(): generates
// a one-time autologin link for an administrator via the mu-plugin above.
func handleWPCLILogin(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, domain, realPath, adminUsername string) {
	ctx := r.Context()

	admins, usersErr := getWPUsers(ctx, userContext, realPath, "administrator")
	if usersErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Admin username could not be obtained", "details": "Please try login manually to wp-admin."})
		return
	}
	if len(admins) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No administrator users found"})
		return
	}

	var matched *wpUser
	if adminUsername == "" {
		matched = &admins[0]
	} else {
		sanitized, ok := sanitizeAdminUser(adminUsername)
		if !ok {
			http.Error(w, "Invalid admin username.", http.StatusBadRequest)
			return
		}
		adminUsername = sanitized
		for i := range admins {
			if admins[i].Username == adminUsername {
				matched = &admins[i]
				break
			}
		}
		if matched == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": adminUsername + " is not an administrator on this site"})
			return
		}
	}

	muFilePath := filepath.Join(realPath, "wp-content", "mu-plugins", "openpanel-login.php")
	muPluginDir := filepath.Dir(muFilePath)
	if mkErr := os.MkdirAll(muPluginDir, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed deploying login handler", "details": mkErr.Error()})
		return
	}
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = os.Chown(muPluginDir, uid, uid)
		if writeErr := os.WriteFile(muFilePath, []byte(openpanelMuPluginPHP), 0o644); writeErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed deploying login handler", "details": writeErr.Error()})
			return
		}
		_ = os.Chown(muFilePath, uid, uid)
	} else if writeErr := os.WriteFile(muFilePath, []byte(openpanelMuPluginPHP), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed deploying login handler", "details": writeErr.Error()})
		return
	}

	dbName, tablePrefix, dbErr := getWPConfigDBInfo(realPath)
	if dbErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": dbErr.Error()})
		return
	}

	optionsTable := tablePrefix + "options"
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	optionName := "op_login_" + fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expires := time.Now().Add(opLoginTokenTTL).Unix()
	optionValue := phpSerializeAssoc(map[string]any{"user_id": matched.ID, "expires": int(expires)}, []string{"user_id", "expires"})

	siteURLRows, rowsErr := mysqlmanager.Exec(ctx, userContext, "SELECT option_value FROM `"+optionsTable+"` WHERE option_name = 'siteurl' LIMIT 1", dbName)
	if rowsErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": rowsErr.Error()})
		return
	}
	siteURL := "https://" + domain
	if len(siteURLRows) > 0 {
		siteURL = toStringCell(siteURLRows[0][0])
	}
	siteURL = strings.TrimSuffix(siteURL, "/")

	insertQuery := "INSERT INTO `" + optionsTable + "` (option_name, option_value, autoload) VALUES ('" + optionName + "', '" + optionValue + "', 'no') ON DUPLICATE KEY UPDATE option_value = '" + optionValue + "'"
	if _, execErr := mysqlmanager.Exec(ctx, userContext, insertQuery, dbName); execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to create login link", "details": execErr.Error()})
		return
	}

	loginLink := siteURL + "/?op_login=" + token
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	} else {
		maskedLink = "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for wp-admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
