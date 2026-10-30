package wordpress

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// wpToolSite is one WordPress install resolved from the site name, everything taken from the panel DB rather than the request
type wpToolSite struct {
	Site, Domain, Docroot, HostPath     string
	UserID                              int
	Username, UserContext, PHPContainer string
	IsLitespeed                         bool
}

// resolveToolSite checks ownership and works out the docroot, host path and php container for site (domain or domain/folder)
func resolveToolSite(a *appctx.App, r *http.Request, site string) (*wpToolSite, int, error) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("internal error")
	}
	return resolveToolSiteFor(r.Context(), a, userID, username, userContext, site)
}

// resolveToolSiteFor is resolveToolSite for a known account, the automatic updater uses it outside a request
func resolveToolSiteFor(ctx context.Context, a *appctx.App, userID int, username, userContext, site string) (*wpToolSite, int, error) {
	if !validSiteName(site) {
		return nil, http.StatusBadRequest, errors.New("Invalid site.")
	}
	domain, folder, _ := strings.Cut(site, "/")
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		return nil, http.StatusForbidden, errors.New("You do not own this domain.")
	}
	dom, found, dbErr := lookupDomainByURL(ctx, a, domain)
	if dbErr != nil || !found || !dom.Docroot.Valid {
		return nil, http.StatusNotFound, errors.New("Domain not found.")
	}
	docroot := strings.TrimSuffix(dom.Docroot.String, "/")
	if folder != "" {
		docroot += "/" + strings.Trim(folder, "/")
	}
	if _, ok := sanitizeWPCLIPath(docroot); !ok || strings.Contains(docroot, "..") {
		return nil, http.StatusBadRequest, errors.New("Invalid document root.")
	}
	htmlVolume := filepath.Join("/home", userContext, "docker-data", "volumes", userContext+"_html_data", "_data")
	hostPath, ok := wpCLIRealPath(htmlVolume, docroot)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Invalid document root.")
	}

	s := &wpToolSite{Site: site, Domain: domain, Docroot: docroot, HostPath: hostPath, UserID: userID, Username: username, UserContext: userContext}
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		s.IsLitespeed = true
		s.PHPContainer, ok = sanitizeName(webServer)
	} else {
		s.PHPContainer, ok = sanitizeName("php-fpm-" + php.GetPHPVForDomain(ctx, a, userContext, domain))
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.New("Could not work out the PHP container for this site.")
	}
	return s, 0, nil
}

// wp runs one wp-cli command against the site, args never go through a shell
func (s *wpToolSite) wp(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	argv := append(podmanmanager.BuildWPCLIBaseCommand(s.UserContext, s.PHPContainer), "--path="+s.Docroot, "--allow-root")
	argv = append(argv, args...)
	cmd := podmanmanager.Command(runCtx, s.UserContext, argv)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if runCtx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("wp-cli timed out after %s", timeout)
	}
	if err != nil {
		return string(out), errors.New(wpCLIErrorText(stderr.String(), string(out)))
	}
	return string(out), nil
}

// wpCLIErrorText picks the "Error: ..." line wp-cli prints, falling back to the last line of output
func wpCLIErrorText(stderr, stdout string) string {
	for _, src := range []string{stderr, stdout} {
		for _, line := range strings.Split(src, "\n") {
			if msg, ok := strings.CutPrefix(strings.TrimSpace(line), "Error: "); ok && msg != "" {
				return truncateMsg(msg)
			}
		}
	}
	if s := strings.TrimSpace(stderr); s != "" {
		return truncateMsg(lastLine(s))
	}
	if s := strings.TrimSpace(stdout); s != "" {
		return truncateMsg(lastLine(s))
	}
	return "wp-cli failed"
}

func truncateMsg(s string) string {
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

var wpToolReadActions = map[string]bool{"updates": true, "database": true, "security": true, "autoupdate": true, "overview": true, "debug-log": true, "maintenance-page": true}

var wpToolWriteActions = map[string]bool{
	"update": true, "auto-update": true, "db-optimize": true, "db-repair": true, "db-clean": true,
	"search-replace": true, "config-set": true, "config-chmod": true,
	"core-update": true, "autoupdate-set": true, "autoupdate-run": true, "debug-log-clear": true,
	"maintenance-page-save": true,
}

// handleWPTools serves the Updates/Maintenance/Security tab endpoints, reads are GET and anything that changes the site is POST (so it carries the CSRF token)
func handleWPTools(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	action := r.PathValue("action")
	switch {
	case wpToolReadActions[action] && r.Method == http.MethodGet:
	case wpToolWriteActions[action] && r.Method == http.MethodPost:
	case wpToolReadActions[action] || wpToolWriteActions[action]:
		web.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	default:
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Unknown action"})
		return
	}

	s, status, err := resolveToolSite(a, r, strings.TrimSpace(r.FormValue("site")))
	if err != nil {
		web.WriteJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	switch action {
	case "updates":
		web.WriteJSON(w, http.StatusOK, cachedWPUpdates(r.Context(), a, s, r.FormValue("refresh") == "1"))
	case "update":
		handleWPToolUpdate(a, w, r, s)
	case "core-update":
		handleWPToolCoreUpdate(a, w, r, s)
	case "autoupdate":
		handleWPToolAutoUpdateState(w, s)
	case "autoupdate-set":
		handleWPToolAutoUpdateSet(a, w, r, s)
	case "autoupdate-run":
		handleWPToolAutoUpdateRun(a, w, r, s)
	case "auto-update":
		handleWPToolAutoUpdate(a, w, r, s)
	case "database":
		handleWPToolDatabase(w, r, s)
	case "db-optimize", "db-repair":
		handleWPToolTableMaint(a, w, r, s, strings.TrimPrefix(action, "db-"))
	case "db-clean":
		handleWPToolClean(a, w, r, s)
	case "search-replace":
		handleWPToolSearchReplace(a, w, r, s)
	case "security":
		handleWPToolSecurity(w, r, s)
	case "overview":
		handleWPToolOverview(w, r, s)
	case "config-set":
		handleWPToolConfigSet(a, w, r, s)
	case "config-chmod":
		handleWPToolConfigChmod(a, w, r, s)
	case "debug-log":
		handleWPToolDebugLog(w, r, s)
	case "debug-log-clear":
		handleWPToolDebugLogClear(a, w, r, s)
	case "maintenance-page":
		handleWPToolMaintenancePage(a, w, r, s)
	case "maintenance-page-save":
		handleWPToolMaintenancePageSave(a, w, r, s)
	}
}

func (s *wpToolSite) record(a *appctx.App, r *http.Request, action string) {
	_ = logger.RecordUserAction(a.Config, s.Username, action+" for WordPress website "+s.Site+" using WP Manager", reqip.ClientIP(r))
}

// ---------------------- UPDATES ---------------------- //

// wpItem is one plugin or theme row from `wp plugin|theme list --format=json`
type wpItem struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	Version       string `json:"version"`
	UpdateVersion string `json:"update_version"`
	HasUpdate     bool   `json:"has_update"`
	AutoUpdate    bool   `json:"auto_update"`
	Manageable    bool   `json:"manageable"`
}

const wpListFields = "--fields=name,title,status,version,update,update_version,auto_update"

// parseWPItems reads plugin/theme list JSON, "update" is "available", "none", "version higher than expected" or false for must-use plugins
func parseWPItems(out []byte) ([]wpItem, error) {
	var raw []struct {
		Name          string `json:"name"`
		Title         string `json:"title"`
		Status        string `json:"status"`
		Version       any    `json:"version"`
		Update        any    `json:"update"`
		UpdateVersion any    `json:"update_version"`
		AutoUpdate    any    `json:"auto_update"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &raw); err != nil {
		return nil, errors.New("unexpected wp-cli output")
	}
	items := make([]wpItem, 0, len(raw))
	for _, it := range raw {
		item := wpItem{
			Name: it.Name, Title: it.Title, Status: it.Status,
			Version: anyString(it.Version), UpdateVersion: anyString(it.UpdateVersion),
			AutoUpdate: anyString(it.AutoUpdate) == "on",
		}
		if item.Title == "" {
			item.Title = item.Name
		}
		// must-use plugins and drop-ins can't be updated or auto-updated, odd names never reach wp-cli
		item.Manageable = it.Status != "must-use" && it.Status != "dropin" && validItemName(it.Name)
		item.HasUpdate = anyString(it.Update) == "available" && item.Manageable
		items = append(items, item)
	}
	return items, nil
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	}
	return ""
}

func validItemName(name string) bool {
	return name != "" && len(name) <= 200 && !strings.HasPrefix(name, "-") && wpCLINameRE.MatchString(name)
}

type wpUpdatesResult struct {
	CheckedAt     int64    `json:"checked_at"`
	Plugins       []wpItem `json:"plugins"`
	Themes        []wpItem `json:"themes"`
	PluginUpdates int      `json:"plugin_updates"`
	ThemeUpdates  int      `json:"theme_updates"`
	PluginError   string   `json:"plugin_error,omitempty"`
	ThemeError    string   `json:"theme_error,omitempty"`
}

func countUpdates(items []wpItem) int {
	n := 0
	for _, it := range items {
		if it.HasUpdate {
			n++
		}
	}
	return n
}

func fetchWPUpdates(ctx context.Context, s *wpToolSite) wpUpdatesResult {
	res := wpUpdatesResult{CheckedAt: time.Now().Unix(), Plugins: []wpItem{}, Themes: []wpItem{}}
	var wg sync.WaitGroup
	list := func(kind string, items *[]wpItem, errText *string) {
		defer wg.Done()
		out, err := s.wp(ctx, 90*time.Second, kind, "list", "--format=json", wpListFields, "--skip-plugins", "--skip-themes")
		if err != nil {
			*errText = err.Error()
			return
		}
		parsed, parseErr := parseWPItems([]byte(out))
		if parseErr != nil {
			*errText = parseErr.Error()
			return
		}
		*items = parsed
	}
	wg.Add(2)
	go list("plugin", &res.Plugins, &res.PluginError)
	go list("theme", &res.Themes, &res.ThemeError)
	wg.Wait()
	res.PluginUpdates, res.ThemeUpdates = countUpdates(res.Plugins), countUpdates(res.Themes)
	return res
}

const wpUpdatesTTL = 10 * time.Minute

var errIncompleteUpdates = errors.New("incomplete")

func wpUpdatesKey(s *wpToolSite) string { return "wp_updates:" + s.UserContext + ":" + s.Site }

// cachedWPUpdates keeps the list for a few minutes so the Overview tiles and the Updates tab share one wp-cli run, failed lists aren't cached
func cachedWPUpdates(ctx context.Context, a *appctx.App, s *wpToolSite, refresh bool) wpUpdatesResult {
	if a.Cache == nil {
		return fetchWPUpdates(ctx, s)
	}
	if refresh {
		_ = a.Cache.Delete(ctx, wpUpdatesKey(s))
	}
	var partial wpUpdatesResult
	res, err := cache.Memoize(ctx, a.Cache, wpUpdatesKey(s), wpUpdatesTTL, func() (wpUpdatesResult, error) {
		partial = fetchWPUpdates(ctx, s)
		if partial.PluginError != "" || partial.ThemeError != "" {
			return partial, errIncompleteUpdates
		}
		return partial, nil
	})
	if err != nil {
		return partial
	}
	return res
}

func forgetWPUpdates(ctx context.Context, a *appctx.App, s *wpToolSite) {
	if a.Cache != nil {
		_ = a.Cache.Delete(ctx, wpUpdatesKey(s))
	}
}

// itemKindAndNames reads kind=plugin|theme plus one or more name fields
func itemKindAndNames(r *http.Request) (string, []string, error) {
	kind := r.FormValue("kind")
	if kind != "plugin" && kind != "theme" {
		return "", nil, errors.New("kind must be plugin or theme")
	}
	_ = r.ParseForm()
	var names []string
	seen := map[string]bool{}
	for _, raw := range r.Form["name"] {
		for _, n := range strings.Split(raw, ",") {
			n = strings.TrimSpace(n)
			if n == "" || seen[n] {
				continue
			}
			if !validItemName(n) {
				return "", nil, errors.New("Invalid " + kind + " name.")
			}
			seen[n] = true
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return "", nil, errors.New("Nothing selected.")
	}
	if len(names) > 100 {
		return "", nil, errors.New("Too many items at once.")
	}
	return kind, names, nil
}

// wpJob answers with one JSON object, or with stream=1 with NDJSON {"phase": ...} lines ending in the result, so the Updates sheet can say what it's doing
type wpJob struct {
	w        http.ResponseWriter
	stream   bool
	flusher  http.Flusher
	canFlush bool
}

func newWPJob(w http.ResponseWriter, r *http.Request) *wpJob {
	j := &wpJob{w: w, stream: r.FormValue("stream") == "1"}
	if j.stream {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("X-Accel-Buffering", "no")
		j.flusher, j.canFlush = w.(http.Flusher)
	}
	return j
}

func (j *wpJob) phase(p string) {
	if j.stream {
		web.WriteNDJSON(j.w, j.flusher, j.canFlush, map[string]any{"phase": p})
	}
}

func (j *wpJob) finish(status int, body map[string]any) {
	if j.stream {
		body["done"] = true
		web.WriteNDJSON(j.w, j.flusher, j.canFlush, body)
		return
	}
	web.WriteJSON(j.w, status, body)
}

// backupGateError is a pre-update backup that failed, so nothing was updated
type backupGateError struct{ err error }

func (e *backupGateError) Error() string {
	return "Backup failed, nothing was updated: " + e.err.Error()
}
func (e *backupGateError) Unwrap() error { return e.err }

// backupThenUpdate takes the backup first when asked and only runs update once it worked
func backupThenUpdate(ctx context.Context, withBackup bool, backup, update func(context.Context) (string, error), phase func(string)) (backupName, out string, err error) {
	if withBackup {
		phase("backup")
		dir, backupErr := backup(ctx)
		if backupErr != nil {
			return "", "", &backupGateError{backupErr}
		}
		backupName = filepath.Base(dir)
	}
	phase("update")
	out, err = update(ctx)
	return backupName, out, err
}

// backupFull takes a database + files backup and removes the folder again when it failed half way
func (s *wpToolSite) backupFull(ctx context.Context, a *appctx.App) (string, error) {
	dir, err := runWPBackup(ctx, a, s.UserContext, s.Site, s.Docroot, true, true)
	if err != nil && dir != "" {
		_ = tenantfs.RemoveAll(dir)
	}
	return dir, err
}

// updateJobCtx outlives the request, a browser leaving half way through shouldn't stop an update
func updateJobCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.Context()), 40*time.Minute)
}

// runUpdateJob runs update behind the optional backup and answers with the outcome
func runUpdateJob(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite, what string, update func(context.Context) (string, error)) {
	job := newWPJob(w, r)
	ctx, cancel := updateJobCtx(r)
	defer cancel()
	backupName, out, err := backupThenUpdate(ctx, r.FormValue("backup") == "1",
		func(ctx context.Context) (string, error) { return s.backupFull(ctx, a) }, update, job.phase)
	if backupName != "" {
		s.record(a, r, "generated full backup "+backupName+" before updating "+what)
	}
	var gate *backupGateError
	if !errors.As(err, &gate) {
		forgetWPUpdates(ctx, a, s)
	}
	if err != nil {
		job.finish(http.StatusInternalServerError, map[string]any{"error": err.Error(), "backup": backupName, "backup_failed": gate != nil})
		return
	}
	s.record(a, r, "updated "+what)
	job.finish(http.StatusOK, map[string]any{"message": lastLine(out), "backup": backupName})
}

func handleWPToolUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	kind, names, err := itemKindAndNames(r)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	args := append(append([]string{kind, "update"}, names...), "--skip-plugins", "--skip-themes")
	runUpdateJob(a, w, r, s, kind+"s "+strings.Join(names, ", "), func(ctx context.Context) (string, error) {
		return s.wp(ctx, 10*time.Minute, args...)
	})
}

// handleWPToolCoreUpdate updates WordPress itself, then its database like wp-admin would on the next visit
func handleWPToolCoreUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	runUpdateJob(a, w, r, s, "WordPress core", func(ctx context.Context) (string, error) {
		out, err := s.wp(ctx, 10*time.Minute, "core", "update", "--skip-plugins", "--skip-themes")
		if err == nil {
			_, _ = s.wp(ctx, 5*time.Minute, "core", "update-db", "--skip-plugins", "--skip-themes")
		}
		return out, err
	})
}

func handleWPToolAutoUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	kind, names, err := itemKindAndNames(r)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	verb := "disable"
	if r.FormValue("state") == "on" {
		verb = "enable"
	}
	args := append([]string{kind, "auto-updates", verb}, names...)
	out, runErr := s.wp(r.Context(), 90*time.Second, append(args, "--skip-plugins", "--skip-themes")...)
	forgetWPUpdates(r.Context(), a, s)
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}
	s.record(a, r, verb+"d auto-updates for "+kind+" "+strings.Join(names, ", "))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": lastLine(out)})
}

// ---------------------- DATABASE ---------------------- //

// likePrefix escapes the table prefix for LIKE, wp_ would otherwise also match wpx
func likePrefix(prefix string) string {
	return strings.ReplaceAll(prefix, "_", `\_`) + "%"
}

func siteTables(ctx context.Context, s *wpToolSite) (dbName, prefix string, tables []string, err error) {
	dbName, prefix, err = getWPConfigDBInfo(s.HostPath)
	if err != nil {
		return "", "", nil, err
	}
	rows, err := mysqlmanager.Exec(ctx, s.UserContext, "SHOW TABLES LIKE '"+likePrefix(prefix)+"'", dbName)
	if err != nil {
		return "", "", nil, err
	}
	for _, row := range rows {
		tables = append(tables, mysqlmanager.ToString(row[0]))
	}
	return dbName, prefix, tables, nil
}

func handleWPToolDatabase(w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	ctx := r.Context()
	dbName, prefix, tables, err := siteTables(ctx, s)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	resp := map[string]any{"database": dbName, "tables": len(tables), "revisions": nil, "expired_transients": nil, "size_bytes": nil}
	if rows, qErr := mysqlmanager.Exec(ctx, s.UserContext, "SELECT SUM(data_length + index_length) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name LIKE '"+likePrefix(prefix)+"'", dbName); qErr == nil && len(rows) == 1 && len(rows[0]) == 1 {
		resp["size_bytes"] = mysqlmanager.ToInt(rows[0][0])
	}
	// both counts straight from SQL, booting WordPress twice just to count rows is slow
	q := "SELECT (SELECT COUNT(*) FROM `" + prefix + "posts` WHERE post_type = 'revision'), " +
		"(SELECT COUNT(*) FROM `" + prefix + "options` WHERE (option_name LIKE '\\_transient\\_timeout\\_%' OR option_name LIKE '\\_site\\_transient\\_timeout\\_%') AND CAST(option_value AS UNSIGNED) < UNIX_TIMESTAMP())"
	if rows, qErr := mysqlmanager.Exec(ctx, s.UserContext, q, dbName); qErr == nil && len(rows) == 1 && len(rows[0]) == 2 {
		resp["revisions"] = mysqlmanager.ToInt(rows[0][0])
		resp["expired_transients"] = mysqlmanager.ToInt(rows[0][1])
	} else if qErr != nil {
		resp["count_error"] = qErr.Error()
	}
	web.WriteJSON(w, http.StatusOK, resp)
}

type tableMaintSummary struct {
	Tables int      `json:"tables"`
	OK     int      `json:"ok"`
	Notes  []string `json:"notes"`
	Errors []string `json:"errors"`
}

// summarizeTableMaint folds OPTIMIZE/REPAIR TABLE rows (Table, Op, Msg_type, Msg_text) into counts and distinct messages
func summarizeTableMaint(rows [][]any) tableMaintSummary {
	sum := tableMaintSummary{Notes: []string{}, Errors: []string{}}
	seenTable, seenNote := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		table, msgType, msg := mysqlmanager.ToString(row[0]), strings.ToLower(mysqlmanager.ToString(row[2])), mysqlmanager.ToString(row[3])
		if !seenTable[table] {
			seenTable[table] = true
			sum.Tables++
		}
		switch msgType {
		case "status":
			if strings.EqualFold(msg, "OK") || strings.EqualFold(msg, "Table is already up to date") {
				sum.OK++
			} else if !seenNote[msg] {
				seenNote[msg] = true
				sum.Notes = append(sum.Notes, msg)
			}
		case "error":
			sum.Errors = append(sum.Errors, table+": "+msg)
		default:
			if !seenNote[msg] {
				seenNote[msg] = true
				sum.Notes = append(sum.Notes, msg)
			}
		}
	}
	return sum
}

func quoteTables(tables []string) string {
	quoted := make([]string, len(tables))
	for i, t := range tables {
		quoted[i] = "`" + strings.ReplaceAll(t, "`", "``") + "`"
	}
	return strings.Join(quoted, ", ")
}

// handleWPToolTableMaint runs OPTIMIZE or REPAIR TABLE over the site's tables, `wp db optimize` needs mysqlcheck which the php containers don't ship
func handleWPToolTableMaint(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite, op string) {
	ctx := r.Context()
	dbName, _, tables, err := siteTables(ctx, s)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if len(tables) == 0 {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No tables found for this site."})
		return
	}
	stmt := "OPTIMIZE TABLE "
	if op == "repair" {
		stmt = "REPAIR TABLE "
	}
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	rows, err := mysqlmanager.Exec(runCtx, s.UserContext, stmt+quoteTables(tables), dbName)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sum := summarizeTableMaint(rows)
	s.record(a, r, "ran "+op+" on "+strconv.Itoa(sum.Tables)+" database tables")
	web.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("%d of %d tables OK.", sum.OK, sum.Tables), "summary": sum})
}

var revisionIDsRE = regexp.MustCompile(`^\d+$`)

// parseIDs reads `wp post list --format=ids`, a single space-separated line
func parseIDs(out string) []string {
	var ids []string
	for _, f := range strings.Fields(out) {
		if revisionIDsRE.MatchString(f) {
			ids = append(ids, f)
		}
	}
	return ids
}

func handleWPToolClean(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	ctx := r.Context()
	what := r.FormValue("what")
	doRevisions, doTransients := what == "revisions" || what == "all", what == "transients" || what == "all"
	if !doRevisions && !doTransients {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "what must be revisions, transients or all"})
		return
	}
	var msgs []string
	if doRevisions {
		out, err := s.wp(ctx, 2*time.Minute, "post", "list", "--post_type=revision", "--format=ids", "--skip-plugins", "--skip-themes")
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Listing revisions failed: " + err.Error()})
			return
		}
		ids := parseIDs(out)
		deleted := 0
		for start := 0; start < len(ids); start += 200 {
			chunk := ids[start:min(start+200, len(ids))]
			if _, err := s.wp(ctx, 5*time.Minute, append(append([]string{"post", "delete"}, chunk...), "--force", "--skip-plugins", "--skip-themes")...); err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Deleted %d revisions, then: %s", deleted, err.Error())})
				return
			}
			deleted += len(chunk)
		}
		msgs = append(msgs, fmt.Sprintf("Deleted %d post revisions.", deleted))
	}
	if doTransients {
		out, err := s.wp(ctx, 2*time.Minute, "transient", "delete", "--expired", "--skip-plugins", "--skip-themes")
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Deleting expired transients failed: " + err.Error()})
			return
		}
		msgs = append(msgs, strings.TrimPrefix(lastLine(out), "Success: "))
	}
	s.record(a, r, "cleaned "+what+" from the database")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": strings.Join(msgs, " ")})
}

// ---------------------- SEARCH & REPLACE ---------------------- //

type srRow struct {
	Table        string `json:"table"`
	Column       string `json:"column"`
	Replacements int    `json:"replacements"`
}

var (
	srDryTotalRE = regexp.MustCompile(`(\d+) replacements? to be made`)
	srRunTotalRE = regexp.MustCompile(`Made (\d+) replacements?`)
)

// parseSearchReplace reads the tab-separated report plus the closing Success line of `wp search-replace`
func parseSearchReplace(out string) (rows []srRow, total int, ok bool) {
	rows = []srRow{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if m := srDryTotalRE.FindStringSubmatch(line); m != nil {
			total, _ = strconv.Atoi(m[1])
			ok = true
			continue
		}
		if m := srRunTotalRE.FindStringSubmatch(line); m != nil {
			total, _ = strconv.Atoi(m[1])
			ok = true
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 3 || cols[0] == "Table" {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(cols[2]))
		if err != nil {
			continue
		}
		rows = append(rows, srRow{Table: cols[0], Column: cols[1], Replacements: n})
	}
	return rows, total, ok
}

func validSRString(s string) bool {
	return s != "" && len(s) <= 500 && !strings.HasPrefix(s, "-") && !strings.ContainsAny(s, "\x00\r\n")
}

// srDryRuns remembers which find/replace pairs had a dry run, a real run is refused without one
var srDryRuns = struct {
	sync.Mutex
	m map[string]time.Time
}{m: map[string]time.Time{}}

const srDryRunTTL = 15 * time.Minute

func srKey(s *wpToolSite, from, to string) string {
	return s.UserContext + "\x00" + s.Site + "\x00" + from + "\x00" + to
}

func rememberDryRun(key string, now time.Time) {
	srDryRuns.Lock()
	defer srDryRuns.Unlock()
	for k, t := range srDryRuns.m {
		if now.Sub(t) > srDryRunTTL {
			delete(srDryRuns.m, k)
		}
	}
	srDryRuns.m[key] = now
}

// takeDryRun reports whether key had a recent dry run and uses it up
func takeDryRun(key string, now time.Time) bool {
	srDryRuns.Lock()
	defer srDryRuns.Unlock()
	t, ok := srDryRuns.m[key]
	delete(srDryRuns.m, key)
	return ok && now.Sub(t) <= srDryRunTTL
}

func handleWPToolSearchReplace(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	from, to := r.FormValue("from"), r.FormValue("to")
	if !validSRString(from) || !validSRString(to) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Both values are required, up to 500 characters, on one line and not starting with a dash."})
		return
	}
	if from == to {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "The replacement is the same as the search text."})
		return
	}
	dryRun := r.FormValue("dry_run") != "0"
	key := srKey(s, from, to)
	if !dryRun && !takeDryRun(key, time.Now()) {
		web.WriteJSON(w, http.StatusConflict, map[string]string{"error": "Run a dry run for these values first."})
		return
	}
	args := []string{"search-replace", from, to, "--all-tables-with-prefix", "--skip-columns=guid", "--report-changed-only", "--skip-plugins", "--skip-themes"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	out, err := s.wp(r.Context(), 10*time.Minute, args...)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	rows, total, ok := parseSearchReplace(out)
	if !ok {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unexpected wp-cli output: " + truncateMsg(lastLine(out))})
		return
	}
	if dryRun {
		rememberDryRun(key, time.Now())
	} else {
		_, _ = s.wp(r.Context(), time.Minute, "cache", "flush", "--skip-plugins", "--skip-themes")
		s.record(a, r, fmt.Sprintf("ran search and replace (%d replacements)", total))
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"dry_run": dryRun, "total": total, "rows": rows})
}

// ---------------------- SECURITY ---------------------- //

// parseConfigConstants reads `wp config list ... --format=json`, value comes back as a bool, number or string
func parseConfigConstants(out []byte) (map[string]bool, error) {
	var raw []struct {
		Name  string `json:"name"`
		Value any    `json:"value"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &raw); err != nil {
		return nil, errors.New("unexpected wp-cli output")
	}
	vals := map[string]bool{}
	for _, c := range raw {
		switch strings.ToLower(strings.Trim(anyString(c.Value), `'" `)) {
		case "true", "1":
			vals[c.Name] = true
		default:
			vals[c.Name] = false
		}
	}
	return vals, nil
}

type wpSecurityResult struct {
	FileEditOff   *bool  `json:"file_edit_off"`
	DebugOn       *bool  `json:"debug_on"`
	ConfigError   string `json:"config_error,omitempty"`
	ConfigMode    string `json:"config_mode,omitempty"`
	ConfigModeOK  bool   `json:"config_mode_ok"`
	CanFixMode    bool   `json:"can_fix_mode"`
	XMLRPCFile    bool   `json:"xmlrpc_file"`
	XMLRPCBlocked bool   `json:"xmlrpc_blocked"`
	Review        int    `json:"review"`
}

// configModeOK is true once wp-config.php has no group or world bits
func configModeOK(mode uint32) bool { return mode&0o077 == 0 }

func (res *wpSecurityResult) countReview() {
	res.Review = 0
	if res.FileEditOff != nil && !*res.FileEditOff {
		res.Review++
	}
	if res.DebugOn != nil && *res.DebugOn {
		res.Review++
	}
	if res.ConfigMode != "" && !res.ConfigModeOK {
		res.Review++
	}
	if res.XMLRPCFile && !res.XMLRPCBlocked {
		res.Review++
	}
}

func handleWPToolSecurity(w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	res := wpSecurityResult{}
	out, err := s.wp(r.Context(), 60*time.Second, "config", "list", "DISALLOW_FILE_EDIT", "WP_DEBUG", "--strict", "--format=json", "--fields=name,value", "--skip-plugins", "--skip-themes")
	vals := map[string]bool{}
	switch {
	case err != nil && strings.Contains(err.Error(), "No matching entries"):
		// neither constant is in wp-config.php
	case err != nil:
		res.ConfigError = err.Error()
	default:
		if vals, err = parseConfigConstants([]byte(out)); err != nil {
			res.ConfigError = err.Error()
		}
	}
	if res.ConfigError == "" {
		off, debug := vals["DISALLOW_FILE_EDIT"], vals["WP_DEBUG"]
		res.FileEditOff, res.DebugOn = &off, &debug
	}

	if info, statErr := tenantfs.Stat(filepath.Join(s.HostPath, "wp-config.php")); statErr == nil {
		perm := uint32(info.Mode().Perm())
		res.ConfigMode = fmt.Sprintf("%04o", perm)
		res.ConfigModeOK = configModeOK(perm)
		// lsphp may not run as the file's owner, so 0600 could lock the site out there
		res.CanFixMode = !res.ConfigModeOK && !s.IsLitespeed
	}
	if _, statErr := tenantfs.Stat(filepath.Join(s.HostPath, "xmlrpc.php")); statErr == nil {
		res.XMLRPCFile = true
		if conf, readErr := tenantfs.ReadFile("/etc/openpanel/caddy/domains/" + s.Domain + ".conf"); readErr == nil {
			res.XMLRPCBlocked = strings.Contains(string(conf), "wp_manager_xmlrpc")
		}
	}
	res.countReview()
	web.WriteJSON(w, http.StatusOK, res)
}

var wpConfigSettable = map[string]bool{"DISALLOW_FILE_EDIT": true, "WP_DEBUG": true}

func handleWPToolConfigSet(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	name := r.FormValue("name")
	if !wpConfigSettable[name] {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported constant"})
		return
	}
	value := "false"
	if r.FormValue("state") == "on" {
		value = "true"
	}
	if _, err := s.wp(r.Context(), 60*time.Second, "config", "set", name, value, "--raw", "--type=constant", "--skip-plugins", "--skip-themes"); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.record(a, r, "set "+name+" to "+value+" in wp-config.php")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": name + " set to " + value + "."})
}

func handleWPToolConfigChmod(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	if s.IsLitespeed {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Change wp-config.php permissions from the File Manager on LiteSpeed."})
		return
	}
	configPath := filepath.Join(s.HostPath, "wp-config.php")
	info, err := tenantfs.Lstat(configPath)
	if err != nil || !info.Mode().IsRegular() {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "wp-config.php not found."})
		return
	}
	if err := tenantfs.Chmod(configPath, 0o600); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.record(a, r, "set wp-config.php permissions to 0600")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "wp-config.php is now 0600."})
}

// ---------------------- OVERVIEW ---------------------- //

type wpOverviewResult struct {
	AdminUser   string `json:"admin_user"`
	AdminCount  int    `json:"admin_count"`
	AdminError  string `json:"admin_error,omitempty"`
	Core        string `json:"core"` // false, minor or true, what WordPress may update by itself
	Plugins     bool   `json:"plugins"`
	Themes      bool   `json:"themes"`
	PolicyError string `json:"policy_error,omitempty"`
	PanelRuns   bool   `json:"panel_runs"` // the panel runs the updates with a backup first
}

// adminSummary picks the first administrator by login for the Site details card
func adminSummary(users []wpUser) (first string, count int) {
	for _, u := range users {
		if u.Role != "administrator" {
			continue
		}
		count++
		if first == "" || u.Username < first {
			first = u.Username
		}
	}
	return first, count
}

// handleWPToolOverview is the Overview tab's Site details: admin user from SQL, auto-update settings from wp-config.php
func handleWPToolOverview(w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	res := wpOverviewResult{Core: defaultUpdatePolicy.Core}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		users, err := getWPUsers(r.Context(), s.UserContext, s.HostPath, "administrator")
		if err != nil {
			res.AdminError = err.Error()
			return
		}
		res.AdminUser, res.AdminCount = adminSummary(users)
	}()
	out, err := s.wp(r.Context(), time.Minute, "config", "list", "WP_AUTO_UPDATE_", "--format=json", "--skip-plugins", "--skip-themes")
	switch {
	case err == nil:
		if p, perr := parseUpdatePolicy([]byte(out)); perr == nil {
			res.Core, res.Plugins, res.Themes = p.Core, p.Plugins, p.Themes
		} else {
			res.PolicyError = perr.Error()
		}
	case !strings.Contains(err.Error(), "No matching entries"):
		res.PolicyError = err.Error()
	}
	if st, stErr := readAutoUpdateState(wpBackupsRoot(s.UserContext, s.Site)); stErr == nil {
		res.PanelRuns = st.Enabled
	}
	wg.Wait()
	web.WriteJSON(w, http.StatusOK, res)
}
