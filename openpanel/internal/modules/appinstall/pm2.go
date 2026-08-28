// Package appinstall (this file) handles the generic PM2-style app
// management routes (/pm2/logs, /pm2/<action>, /pm2/delete) - these work
// identically regardless of whether the underlying app is a Python or
// NodeJS install, which is why they're grouped here rather than under
// either install type specifically.
package appinstall

import (
	"context"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// undeletableAppActions is a confusingly-named variable: it's actually the
// *allowed* app_actions list, not a list of undeletable things.
var undeletableAppActions = map[string]bool{"start": true, "stop": true, "update": true, "restart": true, "envvars": true}

// undeletableServices are core panel services app_actions/app_delete
// refuse to touch even if asked.
var undeletableServices = map[string]bool{
	"elasticsearch": true, "redis": true, "valkey": true, "postgres": true, "mysql": true,
	"mariadb": true, "phpmyadmin": true, "opensearch": true, "memcached": true,
	"openresty": true, "nginx": true, "apache": true, "openlitespeed": true, "litespeed": true,
	"varnish": true, "cron": true, "backup": true, "tor": true,
}

func flashAndRedirectApp(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// RegisterPM2 wires the three /pm2/* routes onto mux. Like RegisterShared,
// these are gated on the "helpers" feature, which is unconditionally
// granted to every user - login-only in practice.
func RegisterPM2(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "helpers")(h)
	}
	mux.Handle("GET /pm2/logs/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2Logs(a, w, r) }))
	mux.Handle("POST /pm2/delete/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2Delete(a, w, r) }))
	mux.Handle("POST /pm2/{action}/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2Action(a, w, r) }))
}

// handlePM2Logs mirrors app_container_logs().
func handlePM2Logs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteName := r.PathValue("site_name")
	_, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	containerName := strings.ToLower(strings.ReplaceAll(siteName, "/", "-"))
	containerName, _, _ = strings.Cut(containerName, "_")
	if containerName == "" {
		return
	}

	lines := 100
	if v := r.URL.Query().Get("lines"); v != "" {
		if parsed, perr := strconv.Atoi(v); perr == nil {
			lines = parsed
		}
	}

	argv := podmanmanager.PodmanArgv(userContext, "ps", "-aqf", "name=^"+containerName+"$", "--no-trunc")
	out, cmdErr := podmanmanager.Command(r.Context(), userContext, argv).CombinedOutput()
	containerID := strings.TrimSpace(string(out))
	if cmdErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Docker error: " + string(out) + "\n"))
		return
	}
	if containerID == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Container " + containerName + " is not running."})
		return
	}

	// `podman logs` itself, not a raw json-file-log-driver path on the
	// host - that path assumed Docker's default logging driver layout
	// (/home/<user>/docker-data/containers/<id>/<id>-json.log), which
	// doesn't exist under podman (confirmed live: the directory itself
	// isn't even created). `podman logs` works regardless of log driver
	// and rootless/remote setup, matching how every other container
	// operation in this codebase already goes through podmanmanager
	// rather than assuming a host-visible file layout.
	logsArgv := podmanmanager.PodmanArgv(userContext, "logs", "--tail", strconv.Itoa(lines), "--timestamps", containerID)
	out, logsErr := podmanmanager.Command(r.Context(), userContext, logsArgv).CombinedOutput()
	if logsErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Error fetching logs: " + string(out) + "\n"))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(out)
}

// pm2SiteLookup mirrors the "find the app type for this site" query shared
// by app_actions() and app_delete() (each with its own slightly different
// WHERE clause).
type pm2SiteLookup struct {
	Type, SiteName, Container string
}

// handlePM2Action mirrors app_actions(): start/stop/update/restart.
func handlePM2Action(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	action := strings.ToLower(r.PathValue("action"))
	siteName := strings.ToLower(r.PathValue("site_name"))

	if !undeletableAppActions[action] {
		flashAndRedirectApp(a, w, r, "error", "Invalid action '"+action+"'. Allowed actions: start, stop, update, restart.", "/website")
		return
	}
	if undeletableServices[siteName] || strings.HasPrefix(siteName, "php-fpm-") {
		flashAndRedirectApp(a, w, r, "error", "Hacker! Service '"+siteName+"' cannot be deleted.", "/website")
		return
	}

	var lookup pm2SiteLookup
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.type, sites.site_name
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.container LIKE ? AND domains.user_id = ?`, "%"+siteName+"%", userID)
	if scanErr := row.Scan(&lookup.Type, &lookup.SiteName); scanErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Unable to detect container type for the application."})
		return
	}

	kind, ok := kindByAppType(lookup.Type)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Not a valid type, only NodeJS, Python, or Ruby applications can be edited."})
		return
	}

	composeFile := "/home/" + userContext + "/docker-compose.yml"
	nameForManager := lookup.SiteName
	ip := reqip.ClientIP(r)

	switch action {
	case "stop":
		argv := podmanmanager.PodmanComposeArgv("down", siteName)
		if runErr := podmanmanager.Command(ctx, userContext, argv).Run(); runErr != nil {
			flashAndRedirectApp(a, w, r, "error", "Failed to stop container '"+siteName+"': "+runErr.Error(), "/website?domain="+nameForManager)
			return
		}
		flashAndRedirectApp(a, w, r, "success", "Stopped container for application: '"+siteName+"'.", "/website?domain="+nameForManager)
		_ = logger.RecordUserAction(a.Config, currentUsername, "stopped container for application "+siteName, ip)
		return

	case "start":
		argv := podmanmanager.PodmanComposeArgv("up", "-d", siteName)
		if runErr := podmanmanager.Command(ctx, userContext, argv).Run(); runErr != nil {
			flashAndRedirectApp(a, w, r, "error", "Failed to start container '"+siteName+"': "+runErr.Error(), "/website?domain="+nameForManager)
			return
		}
		flashAndRedirectApp(a, w, r, "success", "Started container for application '"+siteName+"'.", "/website?domain="+nameForManager)
		_ = logger.RecordUserAction(a.Config, currentUsername, "started container for application "+siteName, ip)
		return

	case "restart":
		pullArgv := podmanmanager.PodmanComposeArgv("-f", composeFile, "pull", siteName)
		if runErr := podmanmanager.Command(ctx, userContext, pullArgv).Run(); runErr != nil {
			flashAndRedirectApp(a, w, r, "error", "Failed to restart container '"+siteName+"': "+runErr.Error(), "/website?domain="+nameForManager)
			return
		}
		downArgv := podmanmanager.PodmanComposeArgv("-f", composeFile, "down", siteName)
		if runErr := podmanmanager.Command(ctx, userContext, downArgv).Run(); runErr != nil {
			flashAndRedirectApp(a, w, r, "error", "Failed to restart container '"+siteName+"': "+runErr.Error(), "/website?domain="+nameForManager)
			return
		}
		upArgv := podmanmanager.PodmanComposeArgv("-f", composeFile, "up", "-d", siteName)
		if runErr := podmanmanager.Command(ctx, userContext, upArgv).Run(); runErr != nil {
			flashAndRedirectApp(a, w, r, "error", "Failed to restart container '"+siteName+"': "+runErr.Error(), "/website?domain="+nameForManager)
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restarted container for application "+siteName, ip)
		flashAndRedirectApp(a, w, r, "success", "Re-downloaded image and started container for application '"+siteName+"'.", "/website?domain="+nameForManager)
		return

	case "update":
		handlePM2Update(a, w, r, currentUsername, userContext, siteName, kind, nameForManager)
		return

	case "envvars":
		handlePM2EnvVars(a, w, r, currentUsername, userContext, siteName, nameForManager)
		return
	}
}

var envVarLineRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.*$`)

// handlePM2EnvVars saves the Env Vars tab's textarea (one KEY=VALUE per
// line, blank lines and #-comments ignored) as the service's
// docker-compose.yml `environment:` list - same LoadCompose/SaveCompose
// round-trip handlePM2Update already uses for command/pids, so this never
// hand-edits YAML text. A full replace, not a merge: these apps have no
// other source of truth for their env vars, so the textarea IS the
// complete set the user wants running.
func handlePM2EnvVars(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, containerName, nameForManager string) {
	_ = r.ParseForm()
	redirectPath := "/website?domain=" + nameForManager

	var envLines []string
	for _, rawLine := range strings.Split(r.FormValue("env_vars"), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !envVarLineRE.MatchString(line) {
			flashAndRedirectApp(a, w, r, "error", "Error saving: invalid line \""+line+"\" - each line must be KEY=VALUE.", redirectPath)
			return
		}
		envLines = append(envLines, line)
	}

	composeData, loadErr := docker.LoadCompose(userContext)
	if loadErr != nil {
		flashAndRedirectApp(a, w, r, "error", "Error saving: could not read docker-compose.yml.", redirectPath)
		return
	}
	services, ok := composeData["services"].(map[string]any)
	if !ok {
		flashAndRedirectApp(a, w, r, "error", "Error saving: docker-compose.yml has no services.", redirectPath)
		return
	}
	svc, ok := services[containerName].(map[string]any)
	if !ok {
		flashAndRedirectApp(a, w, r, "error", "Error saving: service '"+containerName+"' not found in docker-compose.yml.", redirectPath)
		return
	}

	envAny := make([]any, len(envLines))
	for i, l := range envLines {
		envAny[i] = l
	}
	if len(envAny) == 0 {
		delete(svc, "environment")
	} else {
		svc["environment"] = envAny
	}
	if saveErr := docker.SaveCompose(userContext, composeData); saveErr != nil {
		flashAndRedirectApp(a, w, r, "error", "Error saving: "+saveErr.Error(), redirectPath)
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited environment variables for application "+containerName, reqip.ClientIP(r))
	flashAndRedirectApp(a, w, r, "success", "Environment variables saved, make sure to restart the application for changes to take effect.", redirectPath)
}

// pm2Settings is every field the Update tab lets a user change, and the
// only shape a PM2 app's .env/docker-compose.yml service ever gets
// written from - handlePM2Update (an interactive form) and
// handlePM2RestoreBackup (a backup's settings.json, which the user could
// have hand-edited on disk) both go through validatePM2Settings and
// applyPM2Settings below, so a crafted backup file gets exactly the same
// scrutiny a live form submission would. Container name, image, ports,
// volumes, and networks are never fields here - there's no code path
// that lets either caller change them, which is what actually stops a
// restored backup from repointing this service at another container,
// binding a port, or mounting something it shouldn't, rather than any
// one specific check.
type pm2Settings struct {
	Version      string `json:"version"`
	Requirements string `json:"requirements"`
	StartupFile  string `json:"startup_file"`
	CustomCmd    string `json:"custom_cmd"`
	Workdir      string `json:"workdir"`
	CPU          string `json:"cpu"`
	RAM          string `json:"ram"`
	PIDs         string `json:"pids"`
	GitRepoURL   string `json:"git_repo_url"`
}

// validatePM2Settings returns every validation failure in s (empty means
// valid) using the exact same checks handlePM2Update always has.
func validatePM2Settings(s pm2Settings) []string {
	var errs []string
	if !isValidVersion(s.Version) {
		errs = append(errs, "Invalid version format (check tags from hub.docker.com)")
	}
	if !isValidRequirements(s.Requirements) {
		errs = append(errs, "Requirements must be empty (No) or 1 (Yes)")
	}
	if !isValidStartupFile(s.StartupFile) {
		errs = append(errs, "Startup file invalid (must start with /var/www/html/, no path traversal, ends with .js, .py, or .rb)")
	}
	if s.CustomCmd != "" && !isValidCustomCommand(s.CustomCmd) {
		errs = append(errs, "Provided custom startup command is invalid (no path traversal '..' allowed!)")
	}
	if !isValidWorkdir(s.Workdir) {
		errs = append(errs, "Workdir invalid (must start with /var/www/html/, no path traversal)")
	}
	if !isPositiveNumber(s.CPU) {
		errs = append(errs, "CPU core limit provided is not a positive integer.")
	}
	if !isPositiveNumber(s.RAM) {
		errs = append(errs, "Memory limit provided is not a positive integer.")
	}
	if !docker.IsValidPIDsLimit(s.PIDs) {
		errs = append(errs, "PIDs limit must be a positive whole number.")
	}
	if !isValidGitURL(s.GitRepoURL) {
		errs = append(errs, "Invalid git repository URL. Only https:// URLs are supported.")
	}
	return errs
}

// applyPM2Settings writes an already-validated s into containerName's
// .env entries and docker-compose.yml command/pids-limit, exactly what
// handlePM2Update always did inline - factored out so
// handlePM2RestoreBackup can reach the identical apply step. Callers
// must run validatePM2Settings first; this doesn't re-validate.
func applyPM2Settings(a *appctx.App, ctx context.Context, userContext, containerName string, kind Kind, s pm2Settings) error {
	prefix := strings.ToUpper(containerName) + "_" + kind.PyOrNode + "_"
	envFile := "/home/" + userContext + "/.env"
	if !fileExists(envFile) {
		return errors.New("environment file not found")
	}

	ramValue := s.RAM
	if !strings.HasSuffix(strings.ToUpper(ramValue), "G") {
		ramValue += "G"
	}

	docker.SetEnvValue(userContext, prefix+"TAG", s.Version)
	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE container = ?", s.Version, containerName); execErr != nil {
		return execErr
	}
	docker.SetEnvValue(userContext, prefix+"REQUIREMENTS", s.Requirements)
	docker.SetEnvValue(userContext, prefix+"STARTUP_FILE", s.StartupFile)
	docker.SetEnvValue(userContext, prefix+"WORKDIR", s.Workdir)
	docker.SetEnvValue(userContext, prefix+"CPU", s.CPU)
	docker.SetEnvValue(userContext, prefix+"RAM", ramValue)
	docker.SetEnvValue(userContext, prefix+"PIDS", s.PIDs)
	docker.SetEnvValue(userContext, prefix+"CUSTOM_CMD", s.CustomCmd)
	docker.SetEnvValue(userContext, prefix+"GIT_URL", s.GitRepoURL)

	resolvedCommand := buildAppRunCommand(kind, s.Requirements, s.CustomCmd, s.StartupFile, s.GitRepoURL)
	composeData, loadErr := docker.LoadCompose(userContext)
	if loadErr == nil {
		if services, ok := composeData["services"].(map[string]any); ok {
			if svc, svcOK := services[containerName].(map[string]any); svcOK {
				svc["command"] = `sh -c "` + resolvedCommand + `"`
				// Installs from before PIDs became editable have "pids: 100"
				// hardcoded rather than referencing prefix+PIDS - point it
				// at the env var here so a saved edit actually takes effect
				// on the next restart, not just future installs.
				if deploy, ok := svc["deploy"].(map[string]any); ok {
					if resources, ok := deploy["resources"].(map[string]any); ok {
						if limits, ok := resources["limits"].(map[string]any); ok {
							limits["pids"] = "${" + prefix + "PIDS:-100}"
						}
					}
				}
				_ = docker.SaveCompose(userContext, composeData)
			}
		}
	}
	return nil
}

func handlePM2Update(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, containerName string, kind Kind, nameForManager string) {
	_ = r.ParseForm()
	redirectPath := "/website?domain=" + nameForManager

	siteNameUp := strings.ToUpper(containerName)
	envFile := "/home/" + userContext + "/.env"

	settings := pm2Settings{
		Version:      strings.TrimSpace(r.FormValue("version")),
		Requirements: normalizeRequirements(r.FormValue("requirements")),
		StartupFile:  strings.TrimSpace(r.FormValue("startup_file")),
		CustomCmd:    strings.TrimSpace(r.FormValue("custom_cmd")),
		Workdir:      strings.TrimSpace(r.FormValue("workdir")),
		CPU:          strings.TrimSpace(r.FormValue("cpu")),
		RAM:          strings.TrimSpace(r.FormValue("ram")),
		PIDs:         strings.TrimSpace(r.FormValue("pids")),
		GitRepoURL:   strings.TrimSpace(r.FormValue("git_repo_url")),
	}

	if errs := validatePM2Settings(settings); len(errs) > 0 {
		flashAndRedirectApp(a, w, r, "error", "Error saving: "+errs[0], redirectPath)
		return
	}
	if !fileExists(envFile) {
		flashAndRedirectApp(a, w, r, "error", "Environment file not found.", redirectPath)
		return
	}

	if applyErr := applyPM2Settings(a, r.Context(), userContext, containerName, kind, settings); applyErr != nil {
		flashAndRedirectApp(a, w, r, "error", "Error saving: "+applyErr.Error(), redirectPath)
		return
	}

	flashAndRedirectApp(a, w, r, "success", "Changes saved, make sure to restart the application for changes to take effect.", redirectPath)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited container for application "+siteNameUp, reqip.ClientIP(r))
}

// handlePM2Delete mirrors app_delete().
func handlePM2Delete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	siteName := r.PathValue("site_name")
	selectedDomain, subdirectory, _ := strings.Cut(siteName, "/")

	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	var serviceName, appType string
	row := a.DB.QueryRowContext(ctx, "SELECT container, type FROM sites WHERE site_name = ?", siteName)
	if scanErr := row.Scan(&serviceName, &appType); scanErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Unable to detect service name for the application."})
		return
	}

	kind, ok := kindByAppType(appType)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Not a valid type, only NodeJS, Python, or Ruby applications can be removed."})
		return
	}

	docker.StartOrStopContainer(ctx, userContext, strings.ToLower(serviceName), "deactivate", "")

	webServerType, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	switch {
	case webServerType == "apache", webServerType == "openresty", webServerType == "nginx",
		webServerType == "openlitespeed", strings.Contains(strings.ToLower(webServerType), "litespeed"):
		revertWebserverConfig(userContext, subdirectory, webServerType, selectedDomain, serviceName)
		restartArgv := podmanmanager.PodmanArgv(userContext, "restart", webServerType)
		_ = podmanmanager.Command(ctx, userContext, restartArgv).Run()
	default:
		_, _ = w.Write([]byte("unknown"))
		return
	}

	composeData, loadErr := docker.LoadCompose(userContext)
	if loadErr == nil {
		serviceKey := strings.ToLower(serviceName)
		if services, ok := composeData["services"].(map[string]any); ok {
			if _, exists := services[serviceKey]; exists {
				delete(services, serviceKey)
				_ = docker.SaveCompose(userContext, composeData)
				flashSess(a, w, r, "success", "Application '"+siteName+"' removed successfully")
			} else {
				flashSess(a, w, r, "warning", "Service '"+serviceKey+"' not found in compose file")
			}
		}
	}

	envFile := "/home/" + userContext + "/.env"
	serviceNameUp := strings.ToUpper(serviceName)
	if fileExists(envFile) {
		content, _ := os.ReadFile(envFile)
		lines := strings.Split(string(content), "\n")
		prefix := serviceNameUp + "_" + kind.PyOrNode + "_"
		commentPrefix := "# " + kind.PyOrNode + ": " + serviceNameUp
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, prefix) || strings.HasPrefix(trimmed, commentPrefix) {
				continue
			}
			filtered = append(filtered, line)
		}
		_ = os.WriteFile(envFile, []byte(strings.Join(filtered, "\n")), 0o644)
		flashSess(a, w, r, "success", "Environment variables for '"+serviceName+"' removed from .env")
	} else {
		flashSess(a, w, r, "warning", ".env file not found at "+envFile)
	}

	if _, execErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE site_name = ?", siteName); execErr == nil {
		_ = logger.RecordUserAction(a.Config, currentUsername, "deleted application "+siteName, reqip.ClientIP(r))
		flashSess(a, w, r, "success", "Application '"+siteName+"' deleted from the database")
	} else {
		flashAndRedirectApp(a, w, r, "error", "Failed to delete application '"+siteName+"' from the database: "+execErr.Error(), "/sites")
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/sites", http.StatusFound)
}

// revertWebserverConfig's Apache branch searches for what
// editApacheConfig() actually writes ("http://{service_name}"), not the
// "http://localhost" a naive revert might assume - matching the real
// inserted lines is what makes the revert actually find and remove them.
func revertWebserverConfig(userContext, subdirectory, webServerType, domainURL, serviceName string) {
	vhostsPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
	content, err := os.ReadFile(vhostsPath)
	if err != nil {
		return
	}
	backupPath := vhostsPath + ".bak"

	switch {
	case webServerType == "apache":
		var proxyLine, reverseProxyLine string
		if subdirectory != "" {
			proxyLine = "\tProxyPass /" + subdirectory + "/ http://" + serviceName
			reverseProxyLine = "\tProxyPassReverse /" + subdirectory + "/ http://" + serviceName
		} else {
			proxyLine = "\tProxyPass / http://" + serviceName
			reverseProxyLine = "\tProxyPassReverse / http://" + serviceName
		}
		_ = os.WriteFile(backupPath, content, 0o644)

		var updated []string
		for _, line := range strings.SplitAfter(string(content), "\n") {
			if strings.Contains(line, proxyLine) || strings.Contains(line, reverseProxyLine) {
				continue
			}
			updated = append(updated, line)
		}
		if writeErr := os.WriteFile(vhostsPath, []byte(strings.Join(updated, "")), 0o644); writeErr != nil {
			_ = copyFile(backupPath, vhostsPath)
		}

	case webServerType == "nginx" || webServerType == "openresty":
		_ = os.WriteFile(backupPath, content, 0o644)

		markerStart := "location / {"
		if subdirectory != "" {
			markerStart = "location /" + subdirectory + "/ {"
		}
		const markerEnd = "proxy_set_header X-Real-IP $remote_addr;"

		var updated []string
		insideBlock := false
		skipNext := false
		for _, line := range strings.SplitAfter(string(content), "\n") {
			stripped := strings.TrimSpace(line)
			if strings.Contains(stripped, markerStart) {
				insideBlock = true
				continue
			}
			if insideBlock && strings.Contains(stripped, markerEnd) {
				insideBlock = false
				skipNext = true
				continue
			}
			if skipNext {
				skipNext = false
				continue
			}
			if !insideBlock {
				updated = append(updated, line)
			}
		}
		_ = os.WriteFile(vhostsPath, []byte(strings.Join(updated, "")), 0o644)

	case strings.Contains(strings.ToLower(webServerType), "litespeed"):
		_ = os.WriteFile(backupPath, content, 0o644)

		svc := regexp.QuoteMeta(strings.ToLower(serviceName))
		pattern := regexp.MustCompile(`(?s)\n# PM2-PROXY-START:` + svc + `.+?# PM2-PROXY-END:` + svc + `\n`)
		cleaned := strings.TrimRight(pattern.ReplaceAllString(string(content), "\n"), " \t\n\r") + "\n"

		if writeErr := os.WriteFile(vhostsPath, []byte(cleaned), 0o644); writeErr != nil {
			_ = copyFile(backupPath, vhostsPath)
		}
	}
}
