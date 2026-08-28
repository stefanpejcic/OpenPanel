package appinstall

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterPM2API wires the pm2 API routes onto mux. GET .../logs and the
// three POST .../start|stop|restart routes all share a site-name prefix
// with a literal suffix - Go's http.ServeMux requires a "{...}" wildcard
// to be the final segment, so each method gets one "{rest...}" catch-all
// and the dispatch funcs below strip the known suffix by hand to recover
// per-suffix routing. apiregistry.Add still records each logical route
// separately for /api/endpoints.
func RegisterPM2API(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Add("GET /api/pm2/{site_name}/logs")
	mux.Handle("GET /api/pm2/{rest...}", auth.RequireAPI(a, "pm2")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiPM2GetDispatch(a, w, r) })))

	apiregistry.Add("POST /api/pm2/{site_name}/start")
	apiregistry.Add("POST /api/pm2/{site_name}/stop")
	apiregistry.Add("POST /api/pm2/{site_name}/restart")
	mux.Handle("POST /api/pm2/{rest...}", auth.RequireAPI(a, "pm2")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiPM2PostDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "pm2", "PATCH /api/pm2/{site_name...}", func(w http.ResponseWriter, r *http.Request) { apiPM2Update(a, w, r) })
	apiregistry.Handle(mux, a, "pm2", "DELETE /api/pm2/{site_name...}", func(w http.ResponseWriter, r *http.Request) { apiPM2Delete(a, w, r) })
}

// apiPM2GetDispatch dispatches GET /api/pm2/{site_name}/logs.
func apiPM2GetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if strings.HasSuffix(rest, "/logs") {
		r.SetPathValue("site_name", strings.TrimSuffix(rest, "/logs"))
		apiPM2Logs(a, w, r)
		return
	}
	http.NotFound(w, r)
}

// apiPM2PostDispatch dispatches POST /api/pm2/{site_name}/start|stop|restart.
func apiPM2PostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/start"):
		r.SetPathValue("site_name", strings.TrimSuffix(rest, "/start"))
		apiPM2Action(a, w, r, "start")
	case strings.HasSuffix(rest, "/stop"):
		r.SetPathValue("site_name", strings.TrimSuffix(rest, "/stop"))
		apiPM2Action(a, w, r, "stop")
	case strings.HasSuffix(rest, "/restart"):
		r.SetPathValue("site_name", strings.TrimSuffix(rest, "/restart"))
		apiPM2Action(a, w, r, "restart")
	default:
		http.NotFound(w, r)
	}
}

func writeAPIPM2JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiPM2Logs mirrors api_pm2_logs().
func apiPM2Logs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	siteName := r.PathValue("site_name")

	containerName, _, _ := strings.Cut(strings.ToLower(strings.ReplaceAll(siteName, "/", "-")), "_")
	domainRoot, _, _ := strings.Cut(siteName, "/")

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	_, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lines := 100
	if v := r.URL.Query().Get("lines"); v != "" {
		if parsed, perr := strconv.Atoi(v); perr == nil {
			lines = parsed
		}
	}

	body, status := docker.FetchContainerLog(ctx, a, userContext, containerName, lines)
	if status != http.StatusOK {
		writeAPIPM2JSON(w, status, map[string]string{"error": body})
		return
	}

	writeAPIPM2JSON(w, http.StatusOK, map[string]any{"container": containerName, "lines": strings.Split(body, "\n")})
}

// apiPM2Action mirrors _pm2_action().
func apiPM2Action(a *appctx.App, w http.ResponseWriter, r *http.Request, action string) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	siteName := strings.ToLower(r.PathValue("site_name"))
	domainRoot, _, _ := strings.Cut(siteName, "/")

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if undeletableServices[siteName] || strings.HasPrefix(siteName, "php-fpm-") {
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "Service '" + siteName + "' cannot be modified"})
		return
	}

	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	composeFile := "/home/" + userContext + "/docker-compose.yml"

	run := func(argv []string) ([]byte, error) {
		return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	}

	var out []byte
	var runErr error
	switch action {
	case "stop":
		out, runErr = run(podmanmanager.PodmanComposeArgv("-f", composeFile, "down", siteName))
	case "start":
		out, runErr = run(podmanmanager.PodmanComposeArgv("-f", composeFile, "up", "-d", siteName))
	case "restart":
		if out, runErr = run(podmanmanager.PodmanComposeArgv("-f", composeFile, "pull", siteName)); runErr == nil {
			if out, runErr = run(podmanmanager.PodmanComposeArgv("-f", composeFile, "down", siteName)); runErr == nil {
				out, runErr = run(podmanmanager.PodmanComposeArgv("-f", composeFile, "up", "-d", siteName))
			}
		}
	}
	if runErr != nil {
		writeAPIPM2JSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, action+"ed PM2 application "+siteName, reqip.ClientIP(r))
	writeAPIPM2JSON(w, http.StatusOK, map[string]string{"message": "Application " + siteName + " " + action + "ed successfully"})
}

// apiPM2Update mirrors api_pm2_update().
func apiPM2Update(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	siteName := r.PathValue("site_name")
	siteNameLower := strings.ToLower(siteName)
	domainRoot, _, _ := strings.Cut(siteNameLower, "/")

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if undeletableServices[siteNameLower] || strings.HasPrefix(siteNameLower, "php-fpm-") {
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "Service '" + siteNameLower + "' cannot be modified"})
		return
	}

	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var appType, nameForManager, containerName sql.NullString
	row := a.DB.QueryRowContext(ctx, "SELECT type, site_name, container FROM sites WHERE container LIKE ?", "%"+siteNameLower+"%")
	if scanErr := row.Scan(&appType, &nameForManager, &containerName); scanErr != nil {
		writeAPIPM2JSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}

	kind, ok := kindByAppType(appType.String)
	if !ok {
		writeAPIPM2JSON(w, http.StatusBadRequest, map[string]string{"error": "Only NodeJS, Python, or Ruby applications can be updated"})
		return
	}

	var body struct {
		Version      string `json:"version"`
		Requirements string `json:"requirements"`
		StartupFile  string `json:"startup_file"`
		CustomCmd    string `json:"custom_cmd"`
		Workdir      string `json:"workdir"`
		CPU          string `json:"cpu"`
		RAM          string `json:"ram"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	version := strings.TrimSpace(body.Version)
	requirements := normalizeRequirements(body.Requirements)
	startupFile := strings.TrimSpace(body.StartupFile)
	customCmd := strings.TrimSpace(body.CustomCmd)
	workdir := strings.TrimSpace(body.Workdir)
	cpu := strings.TrimSpace(body.CPU)
	ram := strings.TrimSpace(body.RAM)

	var errs []string
	if version != "" && !isValidVersion(version) {
		errs = append(errs, "Invalid version format")
	}
	if !isValidRequirements(requirements) {
		errs = append(errs, `Requirements must be empty or "1"`)
	}
	if startupFile != "" && !isValidStartupFile(startupFile) {
		errs = append(errs, "Invalid startup file path")
	}
	if customCmd != "" && !isValidCustomCommand(customCmd) {
		errs = append(errs, "Invalid custom startup command (no path traversal)")
	}
	if workdir != "" && !isValidWorkdir(workdir) {
		errs = append(errs, "Invalid workdir path")
	}
	if cpu != "" && !isPositiveNumber(cpu) {
		errs = append(errs, "CPU limit must be a positive number")
	}
	if ram != "" && !isPositiveNumber(ram) {
		errs = append(errs, "RAM limit must be a positive number")
	}
	if len(errs) > 0 {
		writeAPIPM2JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	ramValue := ram
	if ram != "" && !strings.HasSuffix(strings.ToUpper(ramValue), "G") {
		ramValue += "G"
	}

	prefix := strings.ToUpper(siteNameLower) + "_" + kind.PyOrNode + "_"
	envFile := "/home/" + userContext + "/.env"
	if !fileExists(envFile) {
		writeAPIPM2JSON(w, http.StatusInternalServerError, map[string]string{"error": ".env file not found"})
		return
	}

	content, _ := os.ReadFile(envFile)
	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		switch {
		case version != "" && strings.HasPrefix(line, prefix+"TAG="):
			newLines = append(newLines, prefix+`TAG="`+version+`"`)
			_, _ = a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE container = ?", version, containerName.String)
		case strings.HasPrefix(line, prefix+"REQUIREMENTS="):
			newLines = append(newLines, prefix+`REQUIREMENTS="`+requirements+`"`)
		case startupFile != "" && strings.HasPrefix(line, prefix+"STARTUP_FILE="):
			newLines = append(newLines, prefix+`STARTUP_FILE="`+startupFile+`"`)
		case workdir != "" && strings.HasPrefix(line, prefix+"WORKDIR="):
			newLines = append(newLines, prefix+`WORKDIR="`+workdir+`"`)
		case cpu != "" && strings.HasPrefix(line, prefix+"CPU="):
			newLines = append(newLines, prefix+`CPU="`+cpu+`"`)
		case ram != "" && strings.HasPrefix(line, prefix+"RAM="):
			newLines = append(newLines, prefix+`RAM="`+ramValue+`"`)
		case customCmd != "" && strings.HasPrefix(line, prefix+"CUSTOM_CMD="):
			newLines = append(newLines, prefix+`CUSTOM_CMD="`+customCmd+`"`)
		default:
			newLines = append(newLines, line)
		}
	}
	_ = os.WriteFile(envFile, []byte(strings.Join(newLines, "\n")), 0o644)

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated PM2 application settings for "+siteNameLower, reqip.ClientIP(r))
	writeAPIPM2JSON(w, http.StatusOK, map[string]string{"message": "Settings saved. Restart the application for changes to take effect."})
}

// apiPM2Delete mirrors api_pm2_delete().
func apiPM2Delete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPIPM2JSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var serviceName, appType sql.NullString
	row := a.DB.QueryRowContext(ctx, "SELECT container, type FROM sites WHERE site_name = ?", siteName)
	if scanErr := row.Scan(&serviceName, &appType); scanErr != nil {
		writeAPIPM2JSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}

	kind, ok := kindByAppType(appType.String)
	if !ok {
		writeAPIPM2JSON(w, http.StatusBadRequest, map[string]string{"error": "Only NodeJS, Python, or Ruby applications can be deleted"})
		return
	}

	docker.StartOrStopContainer(ctx, userContext, strings.ToLower(serviceName.String), "deactivate", "")

	webServerType, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	if webServerType == "apache" || webServerType == "openresty" || webServerType == "nginx" ||
		webServerType == "openlitespeed" || webServerType == "litespeed" {
		revertWebserverConfig(userContext, subdirectory, webServerType, selectedDomain, serviceName.String)
		restartArgv := podmanmanager.PodmanArgv(userContext, "restart", webServerType)
		_ = podmanmanager.Command(ctx, userContext, restartArgv).Run()
	}

	composeData, loadErr := docker.LoadCompose(userContext)
	if loadErr != nil {
		writeAPIPM2JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update docker-compose.yml: " + loadErr.Error()})
		return
	}
	serviceKey := strings.ToLower(serviceName.String)
	if services, ok := composeData["services"].(map[string]any); ok {
		if _, exists := services[serviceKey]; exists {
			delete(services, serviceKey)
			if saveErr := docker.SaveCompose(userContext, composeData); saveErr != nil {
				writeAPIPM2JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update docker-compose.yml: " + saveErr.Error()})
				return
			}
		}
	}

	envFile := "/home/" + userContext + "/.env"
	if fileExists(envFile) {
		serviceNameUp := strings.ToUpper(serviceName.String)
		commentPrefix := "# " + kind.PyOrNode + ": " + serviceNameUp
		prefix := serviceNameUp + "_" + kind.PyOrNode + "_"
		content, _ := os.ReadFile(envFile)
		lines := strings.Split(string(content), "\n")
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, prefix) || strings.HasPrefix(trimmed, commentPrefix) {
				continue
			}
			filtered = append(filtered, line)
		}
		_ = os.WriteFile(envFile, []byte(strings.Join(filtered, "\n")), 0o644)
	}

	if _, execErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE site_name = ?", siteName); execErr != nil {
		writeAPIPM2JSON(w, http.StatusInternalServerError, map[string]string{"error": "Deleted from docker but DB cleanup failed: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted PM2 application "+siteName, reqip.ClientIP(r))
	writeAPIPM2JSON(w, http.StatusOK, map[string]string{"message": "Application " + siteName + " deleted successfully"})
}
