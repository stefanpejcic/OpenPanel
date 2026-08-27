package appinstall

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
)

// containerStartPollAttempts/containerStartPollInterval bound how long the
// post-install readiness check waits for a freshly started container to
// report State.Running=true - see the call site for why a single fixed
// delay isn't enough for language-runtime stacks.
const (
	containerStartPollAttempts = 30
	containerStartPollInterval = 2 * time.Second
)

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// countUserWebsites counts sites owned by any of this user's domains,
// capped at 1000.
func countUserWebsites(a *appctx.App, userID int) (int, error) {
	rows, err := a.DB.Query(
		"SELECT site_name FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?) LIMIT 1000", userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n, rows.Err()
}

// HandleInstallPage renders the install form and handles the early
// over-limit check for a POST; the streaming install itself is handled
// separately by HandleInstall.
func HandleInstallPage(kind Kind, a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injectedData, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefault(plan.WebsitesLimit, 0)
	websiteCount, _ := countUserWebsites(a, userID)

	if websitesLimit != 0 && websiteCount >= websitesLimit {
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed."+plan.UpgradeMessage())
	} else if r.Method == http.MethodPost {
		HandleInstall(kind, a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderInstallPage(kind, a, w, r, domains)
}

func writeNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

// HandleInstall drives the NDJSON-streamed install of one app-type
// service. Several failure branches deliberately end the stream early
// with a bare `return` instead of emitting an {"error": ...} event: once
// the response has started streaming, there's no clean way to signal a
// mid-stream failure beyond just stopping, so the browser sees a silently
// stalled response until the front-end's 5-minute hang timer fires. That
// matches how the rest of the UI already treats an interrupted install
// stream.
func HandleInstall(kind Kind, a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injectedData, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	currentUsername, _ := injectedData["current_username"].(string)
	userContext, _ := injectedData["context"].(string)

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

	ipAddress := reqip.ClientIP(r)

	domainID := r.FormValue("domain_id")
	serviceName := r.FormValue("service_name")
	if serviceName == "" || domainID == "" {
		emit(map[string]any{"error": "Missing required fields domain or service_name"})
		return
	}
	serviceName = strings.ToLower(serviceName)

	if !isValidServiceName(serviceName) {
		emit(map[string]any{"error": "Invalid service name. Use only letters, numbers, '_' and '-'."})
		return
	}
	serviceNameUp := strings.ToUpper(serviceName)

	startupFile := r.FormValue("startup_file")
	cpuLimit := getValidatedFloat(r.FormValue("cpu_limit"), "1.0")
	memLimit := getValidatedFloat(r.FormValue("mem_limit"), "1.0")
	pidsLimit := getValidatedInt(r.FormValue("pids_limit"), "100")

	var appPort int
	if portStr := r.FormValue("port"); portStr != "" {
		if p, perr := strconv.Atoi(portStr); perr == nil && p >= 1 && p <= 65535 {
			appPort = p
		}
	}

	composeFile := "/home/" + userContext + "/docker-compose.yml"
	envFile := "/home/" + userContext + "/.env"

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	lockDir := "/etc/openpanel/openpanel/core/users/" + currentUsername
	_ = os.MkdirAll(lockDir, 0o755)
	lockPath := lockDir + "/krompir.lock"
	if lockErr := os.WriteFile(lockPath, nil, 0o644); lockErr != nil {
		emit(map[string]any{"error": "Error creating " + lockPath + ": " + lockErr.Error()})
		return
	}

	var topDomain string
	var docrootNull sql.NullString
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot FROM domains WHERE domain_id = ?", domainID)
	if scanErr := row.Scan(&topDomain, &docrootNull); scanErr != nil {
		emit(map[string]any{"error": "Domain not found"})
		return
	}
	docroot := docrootNull.String

	if !a.CheckDomainBelongsToUser(ctx, userID, topDomain) {
		return
	}

	emit(map[string]any{"status": "Validating provided data"})
	subdirectory := strings.ToLower(r.FormValue("subdirectory"))
	version := r.FormValue("version")
	if version == "" {
		version = "latest"
	}
	customCmd := r.FormValue("custom_cmd")
	requirements := normalizeRequirements(r.FormValue("requirements"))
	gitRepoURL := strings.TrimSpace(r.FormValue("git_repo_url"))

	if !isValidSubdirectory(subdirectory) {
		emit(map[string]any{"error": "Invalid subdirectory."})
		return
	}
	if version != "latest" && !isValidVersion(version) {
		emit(map[string]any{"error": "Invalid version."})
		return
	}
	if startupFile != "" && !isValidStartupFile(startupFile) {
		emit(map[string]any{"error": "Invalid startup file."})
		return
	}
	if customCmd != "" && !isValidCustomCommand(customCmd) {
		emit(map[string]any{"error": "Invalid custom command."})
		return
	}
	if !isValidGitURL(gitRepoURL) {
		emit(map[string]any{"error": "Invalid git repository URL. Only https:// URLs are supported."})
		return
	}

	installPath := docroot
	selectedDomain := topDomain
	if subdirectory != "" {
		subdirectory = strings.ReplaceAll(subdirectory, " ", "")
		installPath = docroot + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	emit(map[string]any{"status": "Generating docker-compose entry for " + kind.DisplayAppType + " application: '" + serviceName + "'"})
	emit(map[string]any{"status": "Starting the actual install of " + kind.DisplayAppType + " application"})

	templatePath := "/etc/openpanel/docker/compose/" + kind.AppType + ".yml"
	templateBytes, tmplErr := os.ReadFile(templatePath)
	if tmplErr != nil {
		emit(map[string]any{"error": "Error reading/parsing template: " + tmplErr.Error()})
		return
	}
	templateStr := string(templateBytes)

	// The vendored SERVICE.yml templates hardcode "pids: 100" (not
	// editable). Swap it for an env-var placeholder here, before the
	// "SERVICE_NAME" -> serviceNameUp replace below, so it picks up that
	// same substitution and PIDs becomes as editable as CPU/RAM post-install
	// (see handlePM2Update in pm2.go).
	templateStr = strings.ReplaceAll(templateStr, "pids: 100", `pids: "${SERVICE_NAME_`+kind.PyOrNode+`_PIDS:-100}"`)

	resolvedCommand := buildAppRunCommand(kind.PyOrNode, requirements, customCmd, startupFile, gitRepoURL)

	nestedCommandPattern := "${SERVICE_NAME_" + kind.PyOrNode + "_REQUIREMENTS:+" + requirementsInstallToken(kind.PyOrNode) + " &&} " +
		"${SERVICE_NAME_" + kind.PyOrNode + "_CUSTOM_CMD:-" + defaultRunToken(kind.PyOrNode) + " ${SERVICE_NAME_" + kind.PyOrNode + "_STARTUP_FILE:-" + defaultStartupFile(kind.PyOrNode) + "}}"

	replaced := strings.ReplaceAll(templateStr, nestedCommandPattern, resolvedCommand)
	replaced = strings.ReplaceAll(replaced, "SERVICE_NAME", serviceNameUp)
	replaced = strings.ReplaceAll(replaced, "SERVICE", serviceName)
	newServiceStr := replaced

	composeBackup := composeFile + ".app_bak"
	envBackup := envFile + ".app_bak"

	composeContent, composeErr := os.ReadFile(composeFile)
	if composeErr != nil {
		emit(map[string]any{"error": "docker-compose.yml file does not exist, contact support!"})
		return
	}
	if backupErr := copyFile(composeFile, composeBackup); backupErr != nil {
		emit(map[string]any{"error": "Failed to create backup files: " + backupErr.Error()})
		return
	}
	if backupErr := copyFile(envFile, envBackup); backupErr != nil {
		emit(map[string]any{"error": "Failed to create backup files: " + backupErr.Error()})
		return
	}

	composeLines := strings.SplitAfter(string(composeContent), "\n")
	insertPosition := -1
	for i, line := range composeLines {
		if strings.HasPrefix(line, "networks:") {
			insertPosition = i
			break
		}
	}
	if insertPosition == -1 {
		emit(map[string]any{"error": "'networks:' section not found in docker-compose.yml."})
		return
	}
	for _, line := range composeLines {
		if strings.Contains(line, serviceName+":") {
			emit(map[string]any{"error": "Service '" + serviceName + "' already exists, please use a different name."})
			return
		}
	}

	newLines := make([]string, 0, len(composeLines)+1)
	newLines = append(newLines, composeLines[:insertPosition]...)
	newLines = append(newLines, "\n"+newServiceStr+"\n")
	newLines = append(newLines, composeLines[insertPosition:]...)
	if writeErr := os.WriteFile(composeFile, []byte(strings.Join(newLines, "")), 0o644); writeErr != nil {
		emit(map[string]any{"error": "Error updating compose file: " + writeErr.Error()})
		return
	}
	emit(map[string]any{"status": "Added '" + serviceName + "' to docker-compose.yml"})

	envVariables := "\n# " + kind.PyOrNode + ": " + serviceNameUp + "\n" +
		serviceNameUp + "_" + kind.PyOrNode + "_TAG=\"" + version + "\""
	if requirements == "1" {
		envVariables += "\n" + serviceNameUp + "_" + kind.PyOrNode + "_REQUIREMENTS=\"1\""
	} else {
		envVariables += "\n" + serviceNameUp + "_" + kind.PyOrNode + "_REQUIREMENTS=\"\""
	}
	envVariables += "\n" + serviceNameUp + "_" + kind.PyOrNode + "_STARTUP_FILE=\"" + startupFile + "\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_CUSTOM_CMD=\"" + customCmd + "\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_GIT_URL=\"" + gitRepoURL + "\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_WORKDIR=\"" + installPath + "\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_CPU=\"" + formatPyFloat(cpuLimit) + "\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_RAM=\"" + formatPyFloat(memLimit) + "G\"" +
		"\n" + serviceNameUp + "_" + kind.PyOrNode + "_PIDS=\"" + strconv.Itoa(pidsLimit) + "\"\n"

	if envF, openErr := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); openErr == nil {
		_, _ = envF.WriteString(envVariables)
		_ = envF.Close()
	} else {
		emit(map[string]any{"error": "Failed to update .env: " + openErr.Error()})
	}
	emit(map[string]any{"status": "Service added successfully", "success": true})

	emit(map[string]any{"status": "Starting docker container.."})
	docker.StartOrStopContainer(ctx, userContext, serviceName, "activate", "")

	emit(map[string]any{"status": "Checking if container started..."})
	// A single fixed 3s sleep-then-check is enough for stacks that don't do
	// any work at container start, but a NodeJS/Python app's entrypoint
	// typically runs `npm install`/`pip install` before the process is
	// actually up - and if this is the first time this particular
	// language version's image is used, podman also has to pull it first.
	// Both can comfortably exceed 3s, and a container that's still
	// mid-install/mid-pull reports State.Running=false right up until it
	// finishes, not true-with-a-delay - so a single early check is a false
	// negative, not a slow-but-correct one. Poll instead, generously,
	// before concluding it actually failed.
	isRunning := false
	inspectArgv := podmanmanager.PodmanArgv(userContext, "inspect", "-f", "{{.State.Running}}", serviceName)
	for attempt := 0; attempt < containerStartPollAttempts; attempt++ {
		time.Sleep(containerStartPollInterval)
		// podmanmanager.Command, not a bare exec.CommandContext - a
		// per-user context talks to podman over --remote, which needs
		// CONTAINER_HOST pointed at that user's own podman.sock (set by
		// podmanmanager.Command's env, not inherited from this process's
		// environment). Without it, this was silently inspecting nothing
		// (root's default podman instance, where the service doesn't
		// exist), so it could never observe State.Running=true no matter
		// how long the poll loop waited - the container was actually up
		// fine the whole time, this just never noticed.
		inspectOut, inspectErr := podmanmanager.Command(ctx, userContext, inspectArgv).Output()
		if inspectErr == nil && strings.TrimSpace(string(inspectOut)) == "true" {
			isRunning = true
			break
		}
	}

	if !isRunning {
		_ = copyFile(composeBackup, composeFile)
		_ = copyFile(envBackup, envFile)
		rmArgv := podmanmanager.PodmanArgv(userContext, "rm", "-f", serviceName)
		_ = podmanmanager.Command(ctx, userContext, rmArgv).Run()
		_ = os.Remove(lockPath)
		emit(map[string]any{"error": "Container failed to start. Please check allocated resources."})
		return
	}

	emit(map[string]any{"status": "Container '" + serviceName + "' is running successfully."})
	_ = os.Remove(composeBackup)
	_ = os.Remove(envBackup)

	emit(map[string]any{"status": "Detecting installed webserver"})
	webServerType, _ := docker.GetEnvValue(userContext, "WEB_SERVER")

	switch {
	case webServerType == "apache":
		emit(map[string]any{"status": "Editing Apache configuration: creating reverse proxy from " + selectedDomain + " to container " + serviceName})
		editApacheConfig(userContext, topDomain, subdirectory, serviceName, appPort)
	case webServerType == "nginx":
		emit(map[string]any{"status": "Editing Nginx configuration: creating reverse proxy from " + selectedDomain + " to container " + serviceName})
		editNginxConfig(userContext, topDomain, subdirectory, serviceName, appPort)
	case strings.Contains(strings.ToLower(webServerType), "litespeed"):
		emit(map[string]any{"status": "Editing Litespeed configuration: creating reverse proxy from " + selectedDomain + " to container " + serviceName})
		if lswsErr := editLswsConfig(userContext, topDomain, subdirectory, serviceName, appPort); lswsErr != "" {
			return
		}
	case webServerType == "openresty":
		emit(map[string]any{"status": "Editing OpenResty configuration: creating reverse proxy from " + selectedDomain + " to container " + serviceName})
		editNginxConfig(userContext, topDomain, subdirectory, serviceName, appPort)
	default:
		emit(map[string]any{"status": "Unknown webserver for user, no domain conf will be edited! Create proxy manually."})
		return
	}

	emit(map[string]any{"status": "Restarting " + webServerType + " to apply the new proxy"})
	restartArgv := podmanmanager.PodmanArgv(userContext, "restart", webServerType)
	_ = podmanmanager.Command(ctx, userContext, restartArgv).Run()

	emit(map[string]any{"status": "Installation finished, adding the new application to SiteManager"})

	var insertErr error
	if appPort != 0 {
		_, insertErr = a.DB.ExecContext(ctx,
			"INSERT INTO sites (site_name, domain_id, version, type, path, container, ports) VALUES (?, ?, ?, ?, ?, ?, ?)",
			selectedDomain, domainID, version, kind.DisplayAppType, startupFile, serviceNameUp, appPort)
	} else {
		_, insertErr = a.DB.ExecContext(ctx,
			"INSERT INTO sites (site_name, domain_id, version, type, path, container) VALUES (?, ?, ?, ?, ?, ?)",
			selectedDomain, domainID, version, kind.DisplayAppType, startupFile, serviceNameUp)
	}
	if insertErr != nil {
		emit(map[string]any{"error": "An error occurred: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	emit(map[string]any{"status": "New " + kind.DisplayAppType + " application setup completed!"})
	_ = logger.RecordUserAction(a.Config, currentUsername, "created a new "+kind.DisplayAppType+" application on domain "+selectedDomain, ipAddress)

	_ = os.Remove(lockPath)
}

// requirementsInstallToken/defaultRunToken/defaultStartupFile mirror the
// literal placeholder text baked into each app-type's docker-compose
// template (python.yml/nodejs.yml), needed here only to build the exact
// nested-interpolation substring podman-compose can't resolve on its own.
func requirementsInstallToken(pyOrNode string) string {
	switch pyOrNode {
	case "NODE":
		return "npm install"
	case "RUBY":
		return "bundle install"
	default:
		return "pip install -r requirements.txt"
	}
}

func defaultRunToken(pyOrNode string) string {
	switch pyOrNode {
	case "NODE":
		return "node"
	case "RUBY":
		return "ruby"
	default:
		return "python"
	}
}

func defaultStartupFile(pyOrNode string) string {
	switch pyOrNode {
	case "NODE":
		return "index.js"
	case "RUBY":
		return "app.rb"
	default:
		return "app.py"
	}
}

// formatPyFloat formats a float for the .env values written here,
// keeping at least one digit after the decimal point - unlike
// strconv.FormatFloat's -1 precision, which drops a bare ".0".
func formatPyFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
