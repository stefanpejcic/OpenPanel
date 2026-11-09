package wordpress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// "Back up before automatic updates": the panel turns WordPress's own updater off and runs the same updates itself once a day, behind a backup

const (
	autoUpdateMarker    = ".panel-auto-update"
	autoUpdateStateName = "auto-update.json"
	autoUpdateKeep      = 3
	autoUpdateEvery     = 24 * time.Hour
	autoUpdateTick      = time.Hour
	autoUpdateFirstTick = 10 * time.Minute
	autoUpdateTimeout   = 20 * time.Minute
)

type autoUpdateResult struct {
	When int64 `json:"when"`
	// Stamp is When on the server's clock in the backup folder format, so the page shows both times the same way
	Stamp   string   `json:"stamp"`
	Backup  string   `json:"backup,omitempty"`
	Updated []string `json:"updated"`
	Error   string   `json:"error,omitempty"`
	Note    string   `json:"note,omitempty"`
}

// autoUpdateState is backups/<site>/auto-update.json
type autoUpdateState struct {
	Enabled    bool              `json:"enabled"`
	LastRun    int64             `json:"last_run"`
	LastResult *autoUpdateResult `json:"last_result,omitempty"`
	// PrevDisabled is AUTOMATIC_UPDATER_DISABLED as wp-config.php had it before the switch went on, empty when it wasn't defined
	PrevDisabled string `json:"prev_disabled,omitempty"`
}

func autoUpdateStatePath(backupsRoot string) string {
	return filepath.Join(backupsRoot, autoUpdateStateName)
}

// readAutoUpdateState reads the site's state file, a missing file is the default (off, never run)
func readAutoUpdateState(backupsRoot string) (autoUpdateState, error) {
	var st autoUpdateState
	b, err := tenantfs.ReadFile(autoUpdateStatePath(backupsRoot))
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return autoUpdateState{}, errors.New("auto-update.json is not valid JSON")
	}
	return st, nil
}

func writeAutoUpdateState(backupsRoot string, st autoUpdateState) error {
	if err := tenantfs.MkdirAll(backupsRoot, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	path := autoUpdateStatePath(backupsRoot)
	if err := tenantfs.WriteFile(path+".tmp", append(b, '\n'), 0o644); err != nil {
		return err
	}
	return tenantfs.Rename(path+".tmp", path)
}

// autoUpdateStateMu keeps the switch and a running update from overwriting each other's changes to the file
var autoUpdateStateMu sync.Mutex

func changeAutoUpdateState(backupsRoot string, change func(*autoUpdateState)) (autoUpdateState, error) {
	autoUpdateStateMu.Lock()
	defer autoUpdateStateMu.Unlock()
	st, err := readAutoUpdateState(backupsRoot)
	if err != nil {
		// a broken file shouldn't lock the switch, it gets rewritten
		st = autoUpdateState{}
	}
	change(&st)
	return st, writeAutoUpdateState(backupsRoot, st)
}

func autoUpdateDue(st autoUpdateState, now time.Time) bool {
	return st.Enabled && now.Sub(time.Unix(st.LastRun, 0)) >= autoUpdateEvery
}

// ---------------------- POLICY ---------------------- //

// wpUpdatePolicy is what the site's wp-config.php lets WordPress update on its own
type wpUpdatePolicy struct {
	Core    string // false, minor or true
	Plugins bool
	Themes  bool
}

// defaultUpdatePolicy is WordPress's behaviour with none of the constants set: minor core releases, plugins and themes by their own switch
var defaultUpdatePolicy = wpUpdatePolicy{Core: "minor"}

// parseUpdatePolicy reads `wp config list WP_AUTO_UPDATE_ --format=json`, development values like beta or rc count as minor
func parseUpdatePolicy(out []byte) (wpUpdatePolicy, error) {
	p := defaultUpdatePolicy
	var raw []struct {
		Name  string `json:"name"`
		Value any    `json:"value"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &raw); err != nil {
		return p, errors.New("unexpected wp-cli output")
	}
	for _, c := range raw {
		v := strings.ToLower(strings.Trim(anyString(c.Value), `'" `))
		on := v == "true" || v == "1"
		switch c.Name {
		case "WP_AUTO_UPDATE_CORE":
			switch {
			case on:
				p.Core = "true"
			case v == "false" || v == "0" || v == "":
				p.Core = "false"
			default:
				p.Core = "minor"
			}
		case "WP_AUTO_UPDATE_PLUGINS":
			p.Plugins = on
		case "WP_AUTO_UPDATE_THEMES":
			p.Themes = on
		}
	}
	return p, nil
}

// coreCheckArgs asks wp-cli for the core updates the policy allows, nil when core is left alone
func coreCheckArgs(p wpUpdatePolicy) []string {
	switch p.Core {
	case "true":
		return []string{"core", "check-update", "--format=json", "--force-check"}
	case "minor":
		return []string{"core", "check-update", "--minor", "--format=json", "--force-check"}
	}
	return nil
}

func coreUpdateArgs(p wpUpdatePolicy) []string {
	switch p.Core {
	case "true":
		return []string{"core", "update"}
	case "minor":
		return []string{"core", "update", "--minor"}
	}
	return nil
}

// parseCoreCheck returns the newest version from `wp core check-update --format=json`, empty when it's up to date
func parseCoreCheck(out []byte) (string, error) {
	text := strings.TrimSpace(string(out))
	if text == "" || strings.HasPrefix(text, "Success:") {
		return "", nil
	}
	var raw []struct {
		Version any `json:"version"`
	}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return "", errors.New("unexpected wp-cli output")
	}
	if len(raw) == 0 {
		return "", nil
	}
	return anyString(raw[0].Version), nil
}

// pickAutoItems is every updatable item when the policy says all, otherwise the ones whose own auto-update switch is on
func pickAutoItems(items []wpItem, all bool) (names, titles []string) {
	for _, it := range items {
		if it.HasUpdate && (all || it.AutoUpdate) {
			names = append(names, it.Name)
			titles = append(titles, it.Title)
		}
	}
	return names, titles
}

// autoUpdatePlan is the updates one run will make
type autoUpdatePlan struct {
	CoreArgs     []string
	CoreVersion  string
	Plugins      []string
	PluginTitles []string
	Themes       []string
	ThemeTitles  []string
}

func (p autoUpdatePlan) empty() bool {
	return p.CoreArgs == nil && len(p.Plugins) == 0 && len(p.Themes) == 0
}

// commands lists the wp-cli runs for the plan, core first
func (p autoUpdatePlan) commands() [][]string {
	skip := []string{"--skip-plugins", "--skip-themes"}
	var cmds [][]string
	if p.CoreArgs != nil {
		cmds = append(cmds, append(append([]string{}, p.CoreArgs...), skip...))
	}
	if len(p.Plugins) > 0 {
		cmds = append(cmds, append(append([]string{"plugin", "update"}, p.Plugins...), skip...))
	}
	if len(p.Themes) > 0 {
		cmds = append(cmds, append(append([]string{"theme", "update"}, p.Themes...), skip...))
	}
	return cmds
}

// ---------------------- RUN ---------------------- //

// autoUpdateRunner is the site as the updater sees it, tests swap in a fake
type autoUpdateRunner interface {
	wp(ctx context.Context, timeout time.Duration, args ...string) (string, error)
	backup(ctx context.Context) (string, error)
}

func buildAutoUpdatePlan(ctx context.Context, run autoUpdateRunner) (autoUpdatePlan, error) {
	var plan autoUpdatePlan
	policy := defaultUpdatePolicy
	out, err := run.wp(ctx, time.Minute, "config", "list", "WP_AUTO_UPDATE_", "--format=json", "--skip-plugins", "--skip-themes")
	switch {
	case err == nil:
		if policy, err = parseUpdatePolicy([]byte(out)); err != nil {
			return plan, fmt.Errorf("Couldn't read the update settings: %w", err)
		}
	case !strings.Contains(err.Error(), "No matching entries"):
		return plan, fmt.Errorf("Couldn't read the update settings: %w", err)
	}

	if args := coreCheckArgs(policy); args != nil {
		out, err := run.wp(ctx, 2*time.Minute, append(args, "--skip-plugins", "--skip-themes")...)
		if err != nil {
			return plan, fmt.Errorf("Couldn't check for WordPress updates: %w", err)
		}
		version, err := parseCoreCheck([]byte(out))
		if err != nil {
			return plan, fmt.Errorf("Couldn't check for WordPress updates: %w", err)
		}
		if version != "" {
			plan.CoreArgs, plan.CoreVersion = coreUpdateArgs(policy), version
		}
	}

	for _, kind := range []string{"plugin", "theme"} {
		out, err := run.wp(ctx, 90*time.Second, kind, "list", "--format=json", wpListFields, "--skip-plugins", "--skip-themes")
		if err != nil {
			return plan, fmt.Errorf("Couldn't list %ss: %w", kind, err)
		}
		items, err := parseWPItems([]byte(out))
		if err != nil {
			return plan, fmt.Errorf("Couldn't list %ss: %w", kind, err)
		}
		if kind == "plugin" {
			plan.Plugins, plan.PluginTitles = pickAutoItems(items, policy.Plugins)
		} else {
			plan.Themes, plan.ThemeTitles = pickAutoItems(items, policy.Themes)
		}
	}
	return plan, nil
}

// runAutoUpdate backs the site up and makes the updates its policy allows, nothing is updated when the backup fails
func runAutoUpdate(ctx context.Context, run autoUpdateRunner, now time.Time) autoUpdateResult {
	res := autoUpdateResult{When: now.Unix(), Stamp: now.Format("2006-01-02_15-04-05"), Updated: []string{}}
	plan, err := buildAutoUpdatePlan(ctx, run)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if plan.empty() {
		res.Note = "Nothing to update."
		return res
	}
	dir, err := run.backup(ctx)
	if err != nil {
		res.Error = "Backup failed, updates skipped: " + err.Error()
		return res
	}
	res.Backup = filepath.Base(dir)

	var errs []string
	for _, cmd := range plan.commands() {
		_, err := run.wp(ctx, 10*time.Minute, cmd...)
		switch cmd[0] {
		case "core":
			if err != nil {
				errs = append(errs, "WordPress: "+err.Error())
				continue
			}
			_, _ = run.wp(ctx, 5*time.Minute, "core", "update-db", "--skip-plugins", "--skip-themes")
			res.Updated = append(res.Updated, "WordPress "+plan.CoreVersion)
		case "plugin":
			if err != nil {
				errs = append(errs, "Plugins: "+err.Error())
				continue
			}
			res.Updated = append(res.Updated, plan.PluginTitles...)
		case "theme":
			if err != nil {
				errs = append(errs, "Themes: "+err.Error())
				continue
			}
			res.Updated = append(res.Updated, plan.ThemeTitles...)
		}
	}
	res.Error = strings.Join(errs, " ")
	return res
}

// pruneAutoBackups removes the oldest backups the updater made beyond keep, folders without its marker are never touched
func pruneAutoBackups(backupsRoot string, keep int) []string {
	entries, err := tenantfs.ReadDir(backupsRoot)
	if err != nil {
		return nil
	}
	var tagged []string
	for _, e := range entries {
		if !e.IsDir() || !backupDateRE.MatchString(e.Name()) {
			continue
		}
		if info, statErr := tenantfs.Lstat(filepath.Join(backupsRoot, e.Name(), autoUpdateMarker)); statErr == nil && info.Mode().IsRegular() {
			tagged = append(tagged, e.Name())
		}
	}
	if len(tagged) <= keep {
		return nil
	}
	sort.Strings(tagged)
	var removed []string
	for _, name := range tagged[:len(tagged)-keep] {
		if tenantfs.RemoveAll(filepath.Join(backupsRoot, name)) == nil {
			removed = append(removed, name)
		}
	}
	return removed
}

// siteUpdater runs the updater against a real site
type siteUpdater struct {
	s *wpToolSite
	a *appctx.App
}

func (u siteUpdater) wp(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	return u.s.wp(ctx, timeout, args...)
}

// backup tags the folder so pruning can tell it from the user's own backups
func (u siteUpdater) backup(ctx context.Context) (string, error) {
	dir, err := u.s.backupFull(ctx, u.a)
	if err != nil {
		return dir, err
	}
	if err := tenantfs.WriteFile(filepath.Join(dir, autoUpdateMarker), []byte("made by the panel's automatic updater\n"), 0o644); err != nil {
		log.Printf("WP AUTO-UPDATE - %s: couldn't tag backup %s: %v", u.s.Site, filepath.Base(dir), err)
		return dir, nil
	}
	if removed := pruneAutoBackups(wpBackupsRoot(u.s.UserContext, u.s.Site), autoUpdateKeep); len(removed) > 0 {
		log.Printf("WP AUTO-UPDATE - %s: removed old automatic backups %s", u.s.Site, strings.Join(removed, ", "))
	}
	return dir, nil
}

var autoUpdateRunning sync.Map

func autoUpdateKey(s *wpToolSite) string { return s.UserContext + "|" + s.Site }

func claimAutoUpdate(s *wpToolSite) bool {
	_, busy := autoUpdateRunning.LoadOrStore(autoUpdateKey(s), struct{}{})
	return !busy
}

func autoUpdateBusy(s *wpToolSite) bool {
	_, busy := autoUpdateRunning.Load(autoUpdateKey(s))
	return busy
}

// runSiteAutoUpdate does one claimed run and records it, panics end up as the run's error
func runSiteAutoUpdate(a *appctx.App, s *wpToolSite, trigger string) {
	defer autoUpdateRunning.Delete(autoUpdateKey(s))
	root := wpBackupsRoot(s.UserContext, s.Site)
	now := time.Now()
	if _, err := changeAutoUpdateState(root, func(st *autoUpdateState) { st.LastRun = now.Unix() }); err != nil {
		log.Printf("WP AUTO-UPDATE - %s: couldn't save state: %v", s.Site, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), autoUpdateTimeout)
	defer cancel()
	var res autoUpdateResult
	func() {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("WP AUTO-UPDATE - %s: panic: %v", s.Site, p)
				res = autoUpdateResult{When: now.Unix(), Stamp: now.Format("2006-01-02_15-04-05"), Updated: []string{}, Error: "The automatic update stopped on an internal error."}
			}
		}()
		res = runAutoUpdate(ctx, siteUpdater{s: s, a: a}, now)
	}()
	if res.Error == "" && ctx.Err() != nil {
		res.Error = "Stopped after " + autoUpdateTimeout.String() + "."
	}

	if _, err := changeAutoUpdateState(root, func(st *autoUpdateState) { st.LastResult = &res }); err != nil {
		log.Printf("WP AUTO-UPDATE - %s: couldn't save the result: %v", s.Site, err)
	}
	if res.Backup != "" {
		forgetWPUpdates(context.Background(), a, s)
	}

	summary := autoUpdateSummary(res)
	log.Printf("WP AUTO-UPDATE - %s (%s, %s): %s", s.Site, s.Username, trigger, summary)
	_ = logger.RecordUserAction(a.Config, s.Username, "automatic update ("+trigger+") for WordPress website "+s.Site+": "+summary, "")
}

func autoUpdateSummary(res autoUpdateResult) string {
	var parts []string
	if res.Backup != "" {
		parts = append(parts, "backup "+res.Backup)
	}
	if len(res.Updated) > 0 {
		parts = append(parts, "updated "+strings.Join(res.Updated, ", "))
	}
	if res.Note != "" {
		parts = append(parts, res.Note)
	}
	if res.Error != "" {
		parts = append(parts, "error: "+res.Error)
	}
	return strings.Join(parts, "; ")
}

// ---------------------- SCHEDULER ---------------------- //

var autoUpdaterOnce sync.Once

// startAutoUpdater looks every hour for sites with the switch on whose last run is a day old, and runs them one at a time
func startAutoUpdater(a *appctx.App) {
	autoUpdaterOnce.Do(func() {
		go func() {
			timer := time.NewTimer(autoUpdateFirstTick)
			for range timer.C {
				autoUpdateSweep(a)
				timer.Reset(autoUpdateTick)
			}
		}()
	})
}

type autoUpdateSite struct {
	Site, Username, UserContext string
	UserID                      int
}

func listAutoUpdateSites(ctx context.Context, a *appctx.App) ([]autoUpdateSite, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT s.site_name, d.user_id, u.username, COALESCE(u.server, '')
		FROM sites s
		JOIN domains d ON d.domain_id = s.domain_id
		JOIN users u ON u.id = d.user_id
		WHERE s.type = 'wordpress'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sites []autoUpdateSite
	for rows.Next() {
		var site autoUpdateSite
		if err := rows.Scan(&site.Site, &site.UserID, &site.Username, &site.UserContext); err != nil {
			return nil, err
		}
		if site.UserContext == "" || strings.HasPrefix(site.Username, "SUSPENDED_") || !validSiteName(site.Site) {
			continue
		}
		sites = append(sites, site)
	}
	return sites, rows.Err()
}

func autoUpdateSweep(a *appctx.App) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("WP AUTO-UPDATE - sweep panic: %v", p)
		}
	}()
	if a == nil || a.DB == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	sites, err := listAutoUpdateSites(ctx, a)
	cancel()
	if err != nil {
		log.Printf("WP AUTO-UPDATE - couldn't list WordPress sites: %v", err)
		return
	}
	for _, site := range sites {
		st, err := readAutoUpdateState(wpBackupsRoot(site.UserContext, site.Site))
		if err != nil || !autoUpdateDue(st, time.Now()) {
			continue
		}
		resolveCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		s, _, err := resolveToolSiteFor(resolveCtx, a, site.UserID, site.Username, site.UserContext, site.Site)
		cancel()
		if err != nil {
			log.Printf("WP AUTO-UPDATE - %s: skipped: %v", site.Site, err)
			continue
		}
		if claimAutoUpdate(s) {
			runSiteAutoUpdate(a, s, "daily")
		}
	}
}

// ---------------------- HANDLERS ---------------------- //

type autoUpdateView struct {
	autoUpdateState
	Running bool `json:"running"`
}

func handleWPToolAutoUpdateState(w http.ResponseWriter, s *wpToolSite) {
	st, err := readAutoUpdateState(wpBackupsRoot(s.UserContext, s.Site))
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, autoUpdateView{autoUpdateState: st, Running: autoUpdateBusy(s)})
}

func constNotDefined(err error) bool {
	return err != nil && strings.Contains(err.Error(), "is not defined")
}

// handleWPToolAutoUpdateSet turns WordPress's own updater off while the panel runs the updates, and puts AUTOMATIC_UPDATER_DISABLED back as it was when switched off
func handleWPToolAutoUpdateSet(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	ctx := r.Context()
	root := wpBackupsRoot(s.UserContext, s.Site)
	on := r.FormValue("state") == "on"
	current, _ := readAutoUpdateState(root)

	prev := current.PrevDisabled
	if on && !current.Enabled {
		out, err := s.wp(ctx, time.Minute, "config", "get", "AUTOMATIC_UPDATER_DISABLED", "--type=constant", "--skip-plugins", "--skip-themes")
		switch {
		case err == nil:
			prev = "false"
			if v := strings.ToLower(strings.TrimSpace(out)); v == "1" || v == "true" {
				prev = "true"
			}
		case constNotDefined(err):
			prev = ""
		default:
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	var err error
	switch {
	case on:
		_, err = s.wp(ctx, time.Minute, "config", "set", "AUTOMATIC_UPDATER_DISABLED", "true", "--raw", "--type=constant", "--skip-plugins", "--skip-themes")
	case prev == "true" || prev == "false":
		_, err = s.wp(ctx, time.Minute, "config", "set", "AUTOMATIC_UPDATER_DISABLED", prev, "--raw", "--type=constant", "--skip-plugins", "--skip-themes")
	default:
		if _, err = s.wp(ctx, time.Minute, "config", "delete", "AUTOMATIC_UPDATER_DISABLED", "--type=constant", "--skip-plugins", "--skip-themes"); constNotDefined(err) {
			err = nil
		}
	}
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	st, err := changeAutoUpdateState(root, func(st *autoUpdateState) {
		st.Enabled = on
		st.PrevDisabled = prev
		if !on {
			st.PrevDisabled = ""
		}
	})
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Couldn't save the setting: " + err.Error()})
		return
	}
	if on {
		s.record(a, r, "turned on backups before automatic updates (AUTOMATIC_UPDATER_DISABLED set)")
	} else {
		s.record(a, r, "turned off backups before automatic updates (WordPress updates itself again)")
	}
	web.WriteJSON(w, http.StatusOK, autoUpdateView{autoUpdateState: st, Running: autoUpdateBusy(s)})
}

func handleWPToolAutoUpdateRun(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	st, err := readAutoUpdateState(wpBackupsRoot(s.UserContext, s.Site))
	if err != nil || !st.Enabled {
		web.WriteJSON(w, http.StatusConflict, map[string]string{"error": "Turn on \"Back up before automatic updates\" first."})
		return
	}
	if !claimAutoUpdate(s) {
		web.WriteJSON(w, http.StatusConflict, map[string]string{"error": "An automatic update is already running for this site."})
		return
	}
	s.record(a, r, "started an automatic update run")
	go runSiteAutoUpdate(a, s, "run now")
	web.WriteJSON(w, http.StatusAccepted, autoUpdateView{autoUpdateState: st, Running: true})
}
