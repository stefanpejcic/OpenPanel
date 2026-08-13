package wordpress

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

// RegisterAPI wires the WordPress API routes onto mux. Several
// sub-resources share a <domain> prefix with a literal suffix - Go's
// http.ServeMux requires a "{...}" wildcard to be the final segment, so
// GET/POST get a "{rest...}" catch-all where needed and the dispatch funcs
// below strip the known suffix by hand to recover per-suffix routing.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "wordpress", "GET /api/wordpress", func(w http.ResponseWriter, r *http.Request) { apiWordPressList(a, w, r) })
	apiregistry.Handle(mux, a, "wordpress", "GET /api/wordpress/secure", func(w http.ResponseWriter, r *http.Request) { apiWordPressSecureRules(a, w, r) })

	apiregistry.Add("GET /api/wordpress/{domain}/backups")
	apiregistry.Add("GET /api/wordpress/{domain}/secure")
	mux.Handle("GET /api/wordpress/{rest...}", auth.RequireAPI(a, "wordpress")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiWordPressGetDispatch(a, w, r) })))

	apiregistry.Add("POST /api/wordpress/{domain}/backups")
	apiregistry.Add("POST /api/wordpress/{domain}/restore")
	mux.Handle("POST /api/wordpress/{rest...}", auth.RequireAPI(a, "wordpress")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiWordPressPostDispatch(a, w, r) })))

	apiregistry.Add("PUT /api/wordpress/{domain}/secure")
	mux.Handle("PUT /api/wordpress/{rest...}", auth.RequireAPI(a, "wordpress")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiWordPressPutDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "wordpress", "DELETE /api/wordpress/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiWordPressRemove(a, w, r) })
	apiregistry.Handle(mux, a, "wordpress", "POST /api/wordpress/sites/{site_id}/detach", func(w http.ResponseWriter, r *http.Request) { apiWordPressDetach(a, w, r) })

	apiregistry.Handle(mux, a, "wordpress", "POST /api/wordpress/reload", func(w http.ResponseWriter, r *http.Request) { apiWordPressReload(a, w, r) })

	apiregistry.Handle(mux, a, "wordpress", "POST /api/wp-cli/{action}", func(w http.ResponseWriter, r *http.Request) { apiWPCLI(a, w, r) })

	apiregistry.Handle(mux, a, "wordpress", "POST /api/wordpress/install", func(w http.ResponseWriter, r *http.Request) { apiWordPressInstall(a, w, r) })
	apiregistry.Handle(mux, a, "wordpress", "POST /api/wordpress/clone", func(w http.ResponseWriter, r *http.Request) { apiWordPressClone(a, w, r) })
	apiregistry.Handle(mux, a, "wordpress", "GET /api/wordpress/scan", func(w http.ResponseWriter, r *http.Request) { apiWordPressScan(a, w, r) })
}

func writeAPIWPJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiWordPressGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/backups"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/backups"))
		apiWordPressBackupList(a, w, r)
	case strings.HasSuffix(rest, "/secure"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/secure"))
		apiWordPressSecureGet(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

func apiWordPressPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/backups"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/backups"))
		apiWordPressBackupRun(a, w, r)
	case strings.HasSuffix(rest, "/restore"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/restore"))
		apiWordPressRestore(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

func apiWordPressPutDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if strings.HasSuffix(rest, "/secure") {
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/secure"))
		apiWordPressSecureSet(a, w, r)
		return
	}
	http.NotFound(w, r)
}

// ── List ─────────────────────────────────────────────────────────────────

// apiWordPressList mirrors api_wordpress_list().
func apiWordPressList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT site_name, domain_id, admin_email, version, created_date, type, id
		FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)
		AND type = 'WordPress'
	`, userID)
	if execErr != nil {
		writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	defer rows.Close()

	type siteEntry struct {
		SiteName    *string `json:"site_name"`
		DomainID    *int64  `json:"domain_id"`
		AdminEmail  *string `json:"admin_email"`
		Version     *string `json:"version"`
		CreatedDate *string `json:"created_date"`
		Type        *string `json:"type"`
		ID          *int64  `json:"id"`
	}
	sites := []siteEntry{}
	for rows.Next() {
		var siteName, adminEmail, version, createdDate, typ sql.NullString
		var domainID, id sql.NullInt64
		if scanErr := rows.Scan(&siteName, &domainID, &adminEmail, &version, &createdDate, &typ, &id); scanErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		entry := siteEntry{}
		if siteName.Valid {
			entry.SiteName = &siteName.String
		}
		if domainID.Valid {
			entry.DomainID = &domainID.Int64
		}
		if adminEmail.Valid {
			entry.AdminEmail = &adminEmail.String
		}
		if version.Valid {
			entry.Version = &version.String
		}
		if createdDate.Valid {
			entry.CreatedDate = &createdDate.String
		}
		if typ.Valid {
			entry.Type = &typ.String
		}
		if id.Valid {
			entry.ID = &id.Int64
		}
		sites = append(sites, entry)
	}
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"sites": sites, "count": len(sites)})
}

// ── Backups ──────────────────────────────────────────────────────────────

// apiWordPressBackupList mirrors api_wordpress_backup_list(): reuses
// handleGetBackupDates's directory-scan logic (via the capture writer, since
// its bare-array response shape differs from the API's {domain, backups}
// envelope).
func apiWordPressBackupList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _, _ := strings.Cut(domain, "/")
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	r.SetPathValue("selected_domain", domain)

	rec := newAPICaptureWriter()
	handleGetBackupDates(a, rec, r)
	if rec.status >= 400 {
		writeAPIWPJSON(w, rec.status, json.RawMessage(rec.body.Bytes()))
		return
	}
	var dates []backupDateInfo
	_ = json.Unmarshal(rec.body.Bytes(), &dates)
	if dates == nil {
		dates = []backupDateInfo{}
	}
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"domain": domain, "backups": dates})
}

// apiWordPressBackupRun mirrors api_wordpress_backup_run(): resolves
// docroot when not provided, then delegates to handleRunBackup (which reads
// docroot/backup_database/backup_files from the query string, so those are
// copied over from the JSON body here).
func apiWordPressBackupRun(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _, _ := strings.Cut(domain, "/")
	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var body struct {
		BackupDatabase *bool  `json:"backup_database"`
		BackupFiles    *bool  `json:"backup_files"`
		Docroot        string `json:"docroot"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	backupDatabase := body.BackupDatabase == nil || *body.BackupDatabase
	backupFiles := body.BackupFiles == nil || *body.BackupFiles
	docroot := strings.TrimSpace(body.Docroot)

	if docroot == "" {
		dom, found, dbErr := lookupDomainByURL(ctx, a, domainRoot)
		if dbErr != nil {
			writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": dbErr.Error()})
			return
		}
		if !found {
			writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Domain not found"})
			return
		}
		docroot = dom.Docroot.String
	}
	if !backupDatabase && !backupFiles {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "Select at least one of backup_database or backup_files"})
		return
	}

	q := r.URL.Query()
	q.Set("docroot", docroot)
	q.Set("backup_database", boolStr(backupDatabase))
	q.Set("backup_files", boolStr(backupFiles))
	r.URL.RawQuery = q.Encode()
	r.SetPathValue("selected_domain", domain)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	rec := newAPICaptureWriter()
	handleRunBackup(a, rec, r)
	if rec.status >= 400 {
		writeAPIWPJSON(w, rec.status, json.RawMessage(rec.body.Bytes()))
		return
	}
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"message": "Backup completed successfully", "timestamp": timestamp})
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// apiCaptureWriter buffers a reused UI handler's response instead of
// streaming it to the client, so the API wrapper can inspect the outcome
// (status code, plain-text vs JSON body) and translate it into the proper
// JSON envelope, rather than letting the UI handler's own plain-text
// success body leak into an API response.
type apiCaptureWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newAPICaptureWriter() *apiCaptureWriter {
	return &apiCaptureWriter{header: make(http.Header), status: http.StatusOK}
}

func (c *apiCaptureWriter) Header() http.Header         { return c.header }
func (c *apiCaptureWriter) WriteHeader(status int)      { c.status = status }
func (c *apiCaptureWriter) Write(b []byte) (int, error) { return c.body.Write(b) }

// apiWordPressRestore mirrors api_wordpress_restore(): resolves
// docroot/php_version when not provided, then delegates to
// handleRestoreBackup (which reads backup_date/docroot/php_version from the
// query string).
func apiWordPressRestore(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _, _ := strings.Cut(domain, "/")
	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var body struct {
		BackupDate string `json:"backup_date"`
		Docroot    string `json:"docroot"`
		PHPVersion string `json:"php_version"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	backupDate := strings.TrimSpace(body.BackupDate)
	if backupDate == "" {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "backup_date is required"})
		return
	}
	docroot := strings.TrimSpace(body.Docroot)
	phpVersion := strings.TrimSpace(body.PHPVersion)

	if docroot == "" || phpVersion == "" {
		dom, found, dbErr := lookupDomainByURL(ctx, a, domainRoot)
		if dbErr != nil {
			writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": dbErr.Error()})
			return
		}
		if !found {
			writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Domain not found"})
			return
		}
		if docroot == "" {
			docroot = strings.TrimPrefix(strings.TrimPrefix(dom.Docroot.String, "/var/www/html/"), "/")
		}
		if phpVersion == "" {
			phpVersion = dom.PHPVersion.String
		}
	}

	q := r.URL.Query()
	q.Set("backup_date", backupDate)
	q.Set("docroot", docroot)
	q.Set("php_version", phpVersion)
	r.URL.RawQuery = q.Encode()
	r.SetPathValue("selected_domain", domain)

	rec := newAPICaptureWriter()
	handleRestoreBackup(a, rec, r)
	if rec.status >= 400 {
		writeAPIWPJSON(w, rec.status, json.RawMessage(rec.body.Bytes()))
		return
	}

	// handleRestoreBackup's success path writes a plain-text summary rather
	// than JSON (it's shared with the UI) - translate it into the same
	// {message, restored} / {error} shapes api_wordpress_restore() returns.
	text := rec.body.String()
	if strings.HasPrefix(text, "No files to restore") {
		writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "No backup files found to restore"})
		return
	}
	var restored []string
	if strings.Contains(text, "files") {
		restored = append(restored, "files")
	}
	if strings.Contains(text, "database") {
		restored = append(restored, "database")
	}
	writeAPIWPJSON(w, http.StatusOK, map[string]any{
		"message":  "Backup restored: " + strings.Join(restored, " and "),
		"restored": restored,
	})
}

// ── Secure / Hardening ───────────────────────────────────────────────────

// apiWordPressSecureRules mirrors api_wordpress_secure_rules().
func apiWordPressSecureRules(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	out, runErr := exec.CommandContext(r.Context(), "opencli", "websites-secure", "--list-available-rules").CombinedOutput()
	if runErr != nil {
		writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": string(out)})
		return
	}
	rules := wpManagerRuleRE.FindAllString(string(out), -1)
	if rules == nil {
		rules = []string{}
	}
	sort.Strings(rules)
	writeAPIWPJSON(w, http.StatusOK, map[string][]string{"rules": rules})
}

// apiWordPressSecureGet mirrors api_wordpress_secure_get().
func apiWordPressSecureGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _, _ := strings.Cut(domain, "/")
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	configPath := "/etc/openpanel/caddy/domains/" + domainRoot + ".conf"
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Config file not found"})
		return
	}
	matches := wpManagerRuleRE.FindAllString(string(content), -1)
	seen := map[string]bool{}
	active := []string{}
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			active = append(active, m)
		}
	}
	sort.Strings(active)
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"domain": domainRoot, "active_rules": active})
}

var apiWPManagerRuleFullRE = regexp.MustCompile(`^wp_manager_\w+$`)

// apiWordPressSecureSet mirrors api_wordpress_secure_set().
func apiWordPressSecureSet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _, _ := strings.Cut(domain, "/")
	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var body struct {
		DisableAll bool     `json:"disable_all"`
		Rules      []string `json:"rules"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	cmd := []string{"websites-secure", domainRoot}
	var logAction string
	if body.DisableAll || len(body.Rules) == 0 {
		cmd = append(cmd, "--disable-all")
		logAction = "disabled WP hardening rules for " + domainRoot
	} else {
		var valid []string
		for _, rule := range body.Rules {
			if apiWPManagerRuleFullRE.MatchString(rule) {
				valid = append(valid, rule)
			}
		}
		cmd = append(cmd, "--rules="+strings.Join(valid, " "))
		logAction = "enabled WP hardening rules for " + domainRoot + ": " + strings.Join(valid, ", ")
	}

	out, runErr := exec.CommandContext(ctx, "opencli", cmd...).CombinedOutput()
	if runErr != nil {
		writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"message": "Hardening rules applied", "command": cmd})
}

// ── Remove / Detach ──────────────────────────────────────────────────────

// apiWordPressRemove doesn't reuse the UI's handleRemoveWordPress: that
// one reports outcomes via session flash messages and a redirect, while
// this needs its own distinct status-code contract (403/404/500) instead
// of flash-and-redirect - so this reimplements the same DB lookup /
// wp-config.php credential scrape / DB+user drop / file cleanup sequence
// directly, reusing only the low-level pieces (removeDBNameRE/
// removeDBUserRE, wordpressFiles, mysqlmanager.Exec, cache invalidation).
func apiWordPressRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteID := r.PathValue("site_id")

	var fullSiteName, docroot sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ?`, siteID)
	if scanErr := row.Scan(&fullSiteName, &docroot); scanErr != nil {
		writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	siteName := fullSiteName.String
	parts := strings.Split(siteName, "/")
	domainRoot := parts[0]
	subdirs := parts[1:]

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if !docroot.Valid || docroot.String == "" {
		writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "WordPress installation not found in database"})
		return
	}

	realPath := strings.TrimPrefix(docroot.String, "/var/www/html/")
	if len(subdirs) > 0 {
		realPath = filepath.Join(append([]string{realPath}, subdirs...)...)
	}
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + realPath
	wpConfigFile := filepath.Join(volume, "wp-config.php")

	if content, readErr := os.ReadFile(wpConfigFile); readErr == nil {
		dbNameMatch := removeDBNameRE.FindStringSubmatch(string(content))
		dbUserMatch := removeDBUserRE.FindStringSubmatch(string(content))
		if dbNameMatch != nil && dbUserMatch != nil {
			dbName, dbUser := dbNameMatch[1], dbUserMatch[1]
			_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
			_, _ = mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
			invalidateMySQLCaches(ctx, a, userContext, currentUsername)
		}
	}

	var toDelete []string
	for _, item := range wordpressFiles {
		itemPath := filepath.Join(volume, item)
		if _, statErr := os.Stat(itemPath); statErr == nil {
			toDelete = append(toDelete, itemPath)
		}
	}
	if len(toDelete) > 0 {
		_ = exec.CommandContext(ctx, "rm", append([]string{"-rf"}, toDelete...)...).Run()
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", siteID); delErr != nil {
		writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}
	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled WordPress website for "+siteName, reqip.ClientIP(r))
	writeAPIWPJSON(w, http.StatusOK, map[string]string{"message": "WordPress uninstalled successfully"})
}

// apiWordPressDetach mirrors api_wordpress_detach(): standalone for the
// same reason as apiWordPressRemove above.
func apiWordPressDetach(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteID := r.PathValue("site_id")

	var siteName sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? LIMIT 1`, siteID)
	if scanErr := row.Scan(&siteName); scanErr != nil {
		writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	domainRoot, _, _ := strings.Cut(siteName.String, "/")
	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", siteID); delErr != nil {
		writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}
	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "detached WordPress website "+siteName.String, reqip.ClientIP(r))
	writeAPIWPJSON(w, http.StatusOK, map[string]string{"message": "WordPress installation detached from manager"})
}

// ── Reload ───────────────────────────────────────────────────────────────

// apiWordPressReload mirrors api_wordpress_reload(): standalone rather than
// reusing handleReloadWordPressData, since that UI handler only ever writes
// a fixed plain-text banner (no per-site detail), while the API contract
// returns the full `updated` list with each site's refreshed admin_email
// and version - reuses the same walkForWPConfig/checkSiteAlreadyExistsForUser/
// phpContainerForUser helpers handleReloadWordPressData itself is built on.
func apiWordPressReload(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	const baseDirectory = "/var/www/html/"
	phpContainer := phpContainerForUser(userContext)
	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

	type updatedSite struct {
		SiteName   string `json:"site_name"`
		AdminEmail string `json:"admin_email"`
		Version    string `json:"version"`
	}
	updated := []updatedSite{}

	walkForWPConfig(baseDirectory, func(root string) {
		siteURLArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "option", "get", "siteurl")
		out, _ := podmanmanager.Command(ctx, userContext, siteURLArgv).Output()
		siteURL := strings.TrimSpace(string(out))
		siteName := strings.TrimPrefix(strings.TrimPrefix(siteURL, "http://"), "https://")

		domainName := siteName
		if idx := strings.Index(siteName, "/"); idx != -1 {
			domainName = siteName[:idx]
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			return
		}
		if !checkSiteAlreadyExistsForUser(ctx, a, domainName) {
			return
		}

		emailArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "option", "get", "admin_email")
		emailOut, emailErr := podmanmanager.Command(ctx, userContext, emailArgv).Output()
		adminEmail := strings.TrimSpace(string(emailOut))
		if emailErr != nil || !strings.Contains(adminEmail, "@") {
			return
		}

		versionArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "core", "version")
		versionOut, versionErr := podmanmanager.Command(ctx, userContext, versionArgv).Output()
		if versionErr != nil {
			return
		}
		version := strings.TrimSpace(string(versionOut))

		if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET admin_email = ?, version = ? WHERE site_name = ?", adminEmail, version, siteName); execErr != nil {
			return
		}
		_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
		updated = append(updated, updatedSite{SiteName: siteName, AdminEmail: adminEmail, Version: version})
	})

	_ = logger.RecordUserAction(a.Config, currentUsername, "reloaded WordPress data from filesystem", reqip.ClientIP(r))
	writeAPIWPJSON(w, http.StatusOK, map[string]any{"message": "Reload completed", "updated": updated, "count": len(updated)})
}

// ── WP-CLI ───────────────────────────────────────────────────────────────

// apiWPCLIActions is a deliberately separate, smaller action set from the
// UI's own wp-cli passthrough (handleWPCLI).
var apiWPCLIActions = map[string][]string{
	"core_update":       {"core", "update", "--allow-root", "--skip-themes"},
	"core_update_check": {"core", "check-update", "--allow-root", "--skip-themes"},
	"plugin_update_all": {"plugin", "update", "--all", "--allow-root", "--skip-themes"},
	"theme_update_all":  {"theme", "update", "--all", "--allow-root", "--skip-themes"},
	"list_plugins":      {"plugin", "list", "--format=json", "--allow-root", "--skip-themes"},
	"list_themes":       {"theme", "list", "--format=json", "--allow-root", "--skip-themes"},
	"cache_flush":       {"cache", "flush", "--allow-root", "--skip-themes"},
	"cron_run":          {"cron", "event", "run", "--due-now", "--allow-root", "--skip-themes"},
}

var apiWPCLISafeRE = regexp.MustCompile(`^[A-Za-z0-9_\-./]+$`)

// apiWPCLI mirrors api_wp_cli().
func apiWPCLI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	action := strings.ToLower(r.PathValue("action"))

	var body struct {
		Domain     string `json:"domain"`
		PHPVersion string `json:"php_version"`
		Docroot    string `json:"docroot"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	domainParam := strings.TrimSpace(body.Domain)
	phpVersion := strings.TrimSpace(body.PHPVersion)
	domainDirectory := strings.TrimSpace(body.Docroot)

	if domainParam == "" || !apiWPCLISafeRE.MatchString(domainParam) {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or missing domain parameter"})
		return
	}

	domain, subdirectory, hasSubdir := strings.Cut(domainParam, "/")
	if !hasSubdir {
		subdirectory = ""
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIWPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	if phpVersion == "" || domainDirectory == "" {
		dom, found, dbErr := lookupDomainByURL(ctx, a, domain)
		if dbErr != nil {
			writeAPIWPJSON(w, http.StatusInternalServerError, map[string]string{"error": dbErr.Error()})
			return
		}
		if !found {
			writeAPIWPJSON(w, http.StatusNotFound, map[string]string{"error": "Domain not found"})
			return
		}
		if domainDirectory == "" {
			domainDirectory = dom.Docroot.String
		}
		if phpVersion == "" {
			phpVersion = dom.PHPVersion.String
		}
	}

	if subdirectory != "" && domainDirectory != "" {
		normalizedDir := strings.TrimSuffix(domainDirectory, "/")
		normalizedSub := strings.TrimSuffix(subdirectory, "/")
		if !strings.HasSuffix(normalizedDir, normalizedSub) {
			domainDirectory = strings.TrimSuffix(domainDirectory, "/") + "/" + subdirectory
		}
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	var phpContainer string
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = webServer
	} else if phpVersion != "" {
		phpContainer = "php-fpm-" + phpVersion
	} else {
		phpContainer = "php-fpm-8.1"
	}

	subcmd, ok := apiWPCLIActions[action]
	if !ok {
		keys := make([]string, 0, len(apiWPCLIActions))
		for k := range apiWPCLIActions {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]any{"error": "Unknown action. Allowed: " + strings.Join(keys, ", ")})
		return
	}

	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)
	cmd := append(append([]string{}, wpBase...), subcmd...)
	cmd = append(cmd, "--path="+domainDirectory)

	runCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	out, runErr := podmanmanager.Command(runCtx, userContext, cmd).CombinedOutput()
	if runCtx.Err() != nil {
		writeAPIWPJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "WP-CLI command timed out after 120 seconds"})
		return
	}

	returnCode := 0
	if runErr != nil {
		if exitErr, isExit := runErr.(interface{ ExitCode() int }); isExit {
			returnCode = exitErr.ExitCode()
		} else {
			returnCode = -1
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "ran wp-cli "+action+" for "+domainParam, reqip.ClientIP(r))
	writeAPIWPJSON(w, http.StatusOK, map[string]any{
		"action": action, "domain": domainParam, "stdout": strings.TrimSpace(string(out)), "stderr": "", "returncode": returnCode,
	})
}

// ── Install / Clone / Scan ───────────────────────────────────────────────

// withWPForm clones r as a POST carrying the given values as both Form and
// PostForm, so a UI handler that reads r.FormValue(...)/r.Form after its
// own (no-op, since r.Form is already set) parse call sees exactly the
// fields the API's JSON body supplied.
func withWPForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// apiWordPressInstall delegates straight to handleInstallPage (which
// itself calls handleInstallStream on POST): same website-limit check,
// same MySQL-ensure-running step, same NDJSON progress stream written
// directly to the response - just fed from the API's JSON body instead of
// a UI form post.
func apiWordPressInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID         string `json:"domain_id"`
		AdminEmail       string `json:"admin_email"`
		WebsiteName      string `json:"website_name"`
		SiteDescription  string `json:"site_description"`
		AdminUsername    string `json:"admin_username"`
		AdminPassword    string `json:"admin_password"`
		WordPressVersion string `json:"wordpress_version"`
		Subdirectory     string `json:"subdirectory"`
		DBName           string `json:"db_name"`
		DBUser           string `json:"db_user"`
		DBPassword       string `json:"db_password"`
		DBPrefix         string `json:"db_prefix"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.DomainID == "" || body.AdminUsername == "" || body.AdminPassword == "" {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "domain_id, admin_username and admin_password are required"})
		return
	}

	form := url.Values{
		"domain_id": {body.DomainID}, "admin_email": {body.AdminEmail}, "website_name": {body.WebsiteName},
		"site_description": {body.SiteDescription}, "admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword},
		"wordpress_version": {body.WordPressVersion}, "subdirectory": {body.Subdirectory},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword}, "db_prefix": {body.DBPrefix},
	}
	handleInstallPage(a, w, withWPForm(r, form))
}

// apiWordPressClone delegates straight to handleCloneWordPress, which
// already writes a JSON response as-is.
func apiWordPressClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		SourceDomain         string `json:"source_domain"`
		TargetDomain         string `json:"target_domain"`
		SourceDB             string `json:"source_db"`
		SourceFolder         string `json:"source_folder"`
		Subdirectory         string `json:"subdirectory"`
		TargetDB             string `json:"target_db"`
		TargetDBUser         string `json:"target_db_user"`
		TargetDBUserPassword string `json:"target_db_user_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIWPJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	form := url.Values{
		"source_domain": {body.SourceDomain}, "target_domain": {body.TargetDomain},
		"source_db": {body.SourceDB}, "source_folder": {body.SourceFolder}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
	}
	handleCloneWordPress(a, w, withWPForm(r, form))
}

// apiWordPressScan mirrors handleScanWordPress's filesystem walk (finding
// WP installs not yet tracked in the sites table and inserting them - the
// same "checkSiteAlreadyExistsForUser" gate, just inverted from
// apiWordPressReload's "only touch what's already tracked"), returning a
// structured JSON list instead of a plain-text summary.
func apiWordPressScan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lockPath := lockFilePath(currentUsername)
	if info, statErr := os.Stat(lockPath); statErr == nil {
		if time.Since(info.ModTime()) < time.Minute {
			writeAPIWPJSON(w, http.StatusConflict, map[string]string{"error": "A WordPress installation is currently running"})
			return
		}
		_ = os.Remove(lockPath)
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "initiated scan for WordPress installations", reqip.ClientIP(r))

	wwwBaseDirectory := "/var/www/html/"
	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	phpContainer := phpContainerForUser(userContext)
	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

	type foundSite struct {
		ConfigFile string `json:"config_file"`
		Domain     string `json:"domain"`
		Email      string `json:"admin_email"`
		Version    string `json:"version"`
	}
	installations := []foundSite{}

	walkForWPConfig(baseDirectory, func(root string) {
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		configFilePath := strings.TrimPrefix(containerRoot, "/")

		siteURLArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "option", "get", "siteurl")
		out, runErr := podmanmanager.Command(ctx, userContext, siteURLArgv).Output()
		if runErr != nil {
			return
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		siteURL := ""
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				siteURL = strings.TrimSpace(l)
			}
		}
		siteName := strings.TrimPrefix(strings.TrimPrefix(siteURL, "http://"), "https://")
		domainName := siteName
		if idx := strings.Index(siteName, "/"); idx != -1 {
			domainName = siteName[:idx]
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			return
		}
		if checkSiteAlreadyExistsForUser(ctx, a, siteName) {
			return
		}

		emailArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "option", "get", "admin_email")
		emailOut, emailErr := podmanmanager.Command(ctx, userContext, emailArgv).Output()
		adminEmail := strings.TrimSpace(string(emailOut))
		if emailErr != nil || !strings.Contains(adminEmail, "@") {
			return
		}

		versionArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "core", "version")
		versionOut, versionErr := podmanmanager.Command(ctx, userContext, versionArgv).Output()
		if versionErr != nil {
			return
		}
		version := strings.TrimSpace(string(versionOut))

		domainID, _ := getDomainID(ctx, a, domainName)
		if _, insertErr := a.DB.ExecContext(ctx, "INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
			siteName, domainID, adminEmail, version, "WordPress"); insertErr != nil {
			return
		}
		_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))

		installations = append(installations, foundSite{ConfigFile: configFilePath, Domain: domainName, Email: adminEmail, Version: version})
	})

	writeAPIWPJSON(w, http.StatusOK, map[string]any{"message": "Scan completed", "installations": installations, "count": len(installations)})
}
