package appinstall

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// pm2BackupManifest is what settings.json in a backup folder holds - the
// exact same fields the Update tab manages, plus the Env Vars tab's
// custom KEY=VALUE list. There is deliberately no container_name/image/
// volumes/networks field: since applyPM2Settings only ever writes into
// the site's own already-verified container (looked up from the sites
// table, never from the backup file), a hand-edited settings.json has no
// way to repoint this restore at a different container, bind a port, or
// change a volume/network - those fields simply aren't inputs to any
// code path a restore goes through.
type pm2BackupManifest struct {
	pm2Settings
	EnvVars string `json:"env_vars"`
}

var pm2BackupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// RegisterPM2Backup wires the three /pm2/backup* routes onto mux.
func RegisterPM2Backup(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "helpers")(h)
	}
	mux.Handle("POST /pm2/backup/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2CreateBackup(a, w, r) }))
	mux.Handle("GET /pm2/backups/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2ListBackups(a, w, r) }))
	mux.Handle("POST /pm2/restore/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2RestoreBackup(a, w, r) }))
}

// lookupPM2Site resolves the {site_name} path segment (actually the
// container/service identifier, matching every other /pm2/* route's own
// convention) to the owning user's sites-table row, the same query
// handlePM2Action already uses.
func lookupPM2Site(a *appctx.App, userID int, pathSiteName string) (pm2SiteLookup, Kind, bool) {
	var lookup pm2SiteLookup
	row := a.DB.QueryRow(`
		SELECT sites.type, sites.site_name, sites.container
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.container LIKE ? AND domains.user_id = ?`, "%"+strings.ToLower(pathSiteName)+"%", userID)
	if scanErr := row.Scan(&lookup.Type, &lookup.SiteName, &lookup.Container); scanErr != nil {
		return lookup, Kind{}, false
	}
	kind, ok := kindByAppType(lookup.Type)
	return lookup, kind, ok
}

// pm2InstallPaths resolves a site's docroot, both as the container-visible
// "/var/www/html/..." path and as its real host-filesystem path under the
// user's html_data volume.
func pm2InstallPaths(a *appctx.App, userContext, siteName string) (installPath, installHostPath string, ok bool) {
	domain, subdir, _ := strings.Cut(siteName, "/")
	var docrootBase string
	if scanErr := a.DB.QueryRow("SELECT docroot FROM domains WHERE domain_url = ?", domain).Scan(&docrootBase); scanErr != nil {
		return "", "", false
	}
	installPath = docrootBase
	if subdir != "" {
		installPath += "/" + subdir
	}
	hostBase := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	installHostPath = strings.Replace(installPath, "/var/www/html/", hostBase, 1)
	return installPath, installHostPath, true
}

func pm2BackupsDir(userContext, siteName string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + siteName
}

// getCurrentPM2EnvVars mirrors websites.getCurrentEnvVars (unreachable
// from here - that package imports this one) closely enough for a
// backup's purposes: the service's current `environment:` list, one
// "KEY=VALUE" per line.
func getCurrentPM2EnvVars(userContext, containerName string) string {
	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		return ""
	}
	services, ok := composeData["services"].(map[string]any)
	if !ok {
		return ""
	}
	svc, ok := services[strings.ToLower(containerName)].(map[string]any)
	if !ok {
		return ""
	}
	envList, ok := svc["environment"].([]any)
	if !ok {
		return ""
	}
	lines := make([]string, 0, len(envList))
	for _, e := range envList {
		if s, ok := e.(string); ok {
			lines = append(lines, s)
		}
	}
	return strings.Join(lines, "\n")
}

// handlePM2CreateBackup archives the app's docroot and records its
// current Update-tab settings + custom env vars into settings.json -
// mirrors every CMS backups.go's handleXRunBackup in directory layout
// (backups/<site_name>/<timestamp>/...) so the same restore/list UI
// pattern applies here too.
func handlePM2CreateBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lookup, kind, ok := lookupPM2Site(a, userID, r.PathValue("site_name"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}
	_, installHostPath, ok := pm2InstallPaths(a, userContext, lookup.SiteName)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Domain not found"})
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDir := pm2BackupsDir(userContext, lookup.SiteName) + "/" + timestamp
	if mkErr := os.MkdirAll(backupDir, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}

	if runErr := exec.CommandContext(ctx, "tar", "-czf", backupDir+"/files.tar.gz", "-C", installHostPath, ".").Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to archive files: " + runErr.Error()})
		return
	}

	prefix := strings.ToUpper(lookup.Container) + "_" + kind.PyOrNode + "_"
	getVal := func(key string) string {
		v, _ := docker.GetEnvValue(userContext, prefix+key)
		return v
	}
	manifest := pm2BackupManifest{
		pm2Settings: pm2Settings{
			Version:      getVal("TAG"),
			Requirements: getVal("REQUIREMENTS"),
			StartupFile:  getVal("STARTUP_FILE"),
			CustomCmd:    getVal("CUSTOM_CMD"),
			Workdir:      getVal("WORKDIR"),
			CPU:          getVal("CPU"),
			RAM:          strings.TrimSuffix(strings.TrimSuffix(getVal("RAM"), "g"), "G"),
			PIDs:         getVal("PIDS"),
			GitRepoURL:   getVal("GIT_URL"),
		},
		EnvVars: getCurrentPM2EnvVars(userContext, lookup.Container),
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if writeErr := os.WriteFile(backupDir+"/settings.json", manifestBytes, 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created a backup for application "+lookup.SiteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "date": timestamp})
}

// handlePM2ListBackups returns every backup timestamp available for the
// site, newest first.
func handlePM2ListBackups(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	_, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lookup, _, ok := lookupPM2Site(a, userID, r.PathValue("site_name"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}

	entries, readErr := os.ReadDir(pm2BackupsDir(userContext, lookup.SiteName))
	if readErr != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	dates := []string{}
	for _, entry := range entries {
		if entry.IsDir() && pm2BackupDateRE.MatchString(entry.Name()) {
			dates = append(dates, entry.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	writeJSON(w, http.StatusOK, dates)
}

// handlePM2RestoreBackup extracts a backup's files and re-applies its
// settings.json - every setting goes through the exact same
// validatePM2Settings/applyPM2Settings path a live Update-tab submission
// does (see pm2Settings' doc comment for why that's what actually makes
// a hand-edited backup file safe to restore), and the env_vars list gets
// the same per-line KEY=VALUE check handlePM2EnvVars already enforces.
// The whole restore is rejected up front if anything fails validation -
// nothing is extracted or written until the backup's settings have
// already passed.
func handlePM2RestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	backupDate := r.URL.Query().Get("backup_date")
	if !pm2BackupDateRE.MatchString(backupDate) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
		return
	}

	lookup, kind, ok := lookupPM2Site(a, userID, r.PathValue("site_name"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}
	_, installHostPath, ok := pm2InstallPaths(a, userContext, lookup.SiteName)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Domain not found"})
		return
	}
	backupDir := pm2BackupsDir(userContext, lookup.SiteName) + "/" + backupDate

	manifestBytes, readErr := os.ReadFile(backupDir + "/settings.json")
	if readErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Backup settings not found."})
		return
	}
	var manifest pm2BackupManifest
	if jsonErr := json.Unmarshal(manifestBytes, &manifest); jsonErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Backup settings.json is not valid JSON: " + jsonErr.Error()})
		return
	}

	if errs := validatePM2Settings(manifest.pm2Settings); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Backup settings failed validation, refusing to restore.", "details": errs})
		return
	}
	var envLines []string
	for _, rawLine := range strings.Split(manifest.EnvVars, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !envVarLineRE.MatchString(line) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Backup env_vars contains an invalid line: " + line})
			return
		}
		envLines = append(envLines, line)
	}

	containerName := strings.ToLower(lookup.Container)

	if info, statErr := os.Stat(backupDir + "/files.tar.gz"); statErr == nil && !info.IsDir() {
		if mkErr := os.MkdirAll(installHostPath, 0o755); mkErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
			return
		}
		if runErr := exec.CommandContext(ctx, "tar", "-xzf", backupDir+"/files.tar.gz", "-C", installHostPath).Run(); runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to extract files: " + runErr.Error()})
			return
		}
		_ = exec.CommandContext(ctx, "chown", "-R", userContext+":"+userContext, installHostPath).Run()
	}

	if applyErr := applyPM2Settings(a, ctx, userContext, containerName, kind, manifest.pm2Settings); applyErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to apply settings: " + applyErr.Error()})
		return
	}

	composeData, loadErr := docker.LoadCompose(userContext)
	if loadErr == nil {
		if services, ok := composeData["services"].(map[string]any); ok {
			if svc, svcOK := services[containerName].(map[string]any); svcOK {
				if len(envLines) == 0 {
					delete(svc, "environment")
				} else {
					envAny := make([]any, len(envLines))
					for i, l := range envLines {
						envAny[i] = l
					}
					svc["environment"] = envAny
				}
				_ = docker.SaveCompose(userContext, composeData)
			}
		}
	}

	// -f composeFile explicitly - podman-compose has no reliable default
	// working directory here (podmanmanager.Command doesn't set cmd.Dir),
	// confirmed live: down/up without -f silently did nothing (no error
	// surfaced, container just never came back up) while the identical
	// command with -f worked immediately.
	composeFile := "/home/" + userContext + "/docker-compose.yml"
	downArgv := podmanmanager.PodmanComposeArgv("-f", composeFile, "down", containerName)
	downOut, _ := podmanmanager.Command(ctx, userContext, downArgv).CombinedOutput()
	upArgv := podmanmanager.PodmanComposeArgv("-f", composeFile, "up", "-d", containerName)
	upOut, upErr := podmanmanager.Command(ctx, userContext, upArgv).CombinedOutput()

	_ = logger.RecordUserAction(a.Config, currentUsername, "restored backup ("+backupDate+") for application "+lookup.SiteName, reqip.ClientIP(r))
	if upErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "success",
			"warning": "Backup restored, but the container failed to restart automatically - restart it manually from the Overview tab: " +
				strings.TrimSpace(string(downOut)) + " " + strings.TrimSpace(string(upOut)),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
