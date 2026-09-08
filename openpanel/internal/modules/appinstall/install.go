package appinstall

import (
	"context"
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

// containerStartPollAttempts/containerStartPollInterval bound how long the post-install readiness check waits for a freshly started container to report State.Running=true
const (
	containerStartPollAttempts = 30
	containerStartPollInterval = 2 * time.Second
)

// containerFailureDetail inspects a container that never reached State.Running=true and returns a human-readable reason (OOM-killed, exit code, podman start error, or a log tail), or "" if nothing specific was found. Must be called before the container is `rm -f`'d, which takes its logs and state with it.
func containerFailureDetail(ctx context.Context, userContext, serviceName string) string {
	inspectArgv := podmanmanager.PodmanArgv(userContext, "inspect", "-f",
		"{{.State.Status}}\t{{.State.ExitCode}}\t{{.State.OOMKilled}}\t{{.State.Error}}", serviceName)
	inspectOut, inspectErr := podmanmanager.Command(ctx, userContext, inspectArgv).Output()

	var reason string
	if inspectErr == nil {
		fields := strings.SplitN(strings.TrimSpace(string(inspectOut)), "\t", 4)
		if len(fields) == 4 {
			status, exitCode, oomKilled, stateErr := fields[0], fields[1], fields[2], fields[3]
			switch {
			case oomKilled == "true":
				reason = "Container was killed for using too much memory (out of memory). Increase the memory limit allocated to this app."
			case stateErr != "" && stateErr != "<no value>":
				reason = "Error: " + stateErr
			case status == "exited" && exitCode != "0" && exitCode != "":
				reason = "Container exited with code " + exitCode + "."
			}
		}
	}

	logsArgv := podmanmanager.PodmanArgv(userContext, "logs", "--tail", "20", serviceName)
	if logsOut, logsErr := podmanmanager.Command(ctx, userContext, logsArgv).CombinedOutput(); logsErr == nil {
		if tail := strings.TrimSpace(string(logsOut)); tail != "" {
			const maxTailLen = 2000
			if len(tail) > maxTailLen {
				tail = "…" + tail[len(tail)-maxTailLen:]
			}
			if reason != "" {
				reason += "\n"
			}
			reason += "Last container output:\n" + tail
		}
	}

	return reason
}

// vendoredComposeServiceIndent is the leading-space indent the vendored SERVICE.yml templates hardcode for their own top-level service key
const vendoredComposeServiceIndent = 2

// composeServiceIndent detects the indent actually used by service keys under "services:", from the file's lines following that line - the indent of the first non-blank line, which is always a service key.
// Can't be assumed fixed: a vendored file starts at 2 spaces, but once docker.SaveCompose rewrites it, yaml.v3 always re-indents to 4. Defaults to vendoredComposeServiceIndent if there's no non-blank line.
func composeServiceIndent(linesAfterServices []string) int {
	for _, line := range linesAfterServices {
		trimmed := strings.TrimRight(line, "\n")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		return len(trimmed) - len(strings.TrimLeft(trimmed, " "))
	}
	return vendoredComposeServiceIndent
}

// indentComposeService re-indents a rendered SERVICE.yml template (whose service key sits at vendoredComposeServiceIndent) to match existingIndent, the indent its siblings actually use under "services:" - see composeServiceIndent. Getting this wrong isn't just cosmetic: a deeper-indented sibling is invalid YAML, and this broke every fresh install in production before it was fixed.
func indentComposeService(serviceStr string, existingIndent int) string {
	pad := ""
	if existingIndent > vendoredComposeServiceIndent {
		pad = strings.Repeat(" ", existingIndent-vendoredComposeServiceIndent)
	}
	return pad + strings.ReplaceAll(strings.TrimSuffix(serviceStr, "\n"), "\n", "\n"+pad) + "\n"
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// countUserWebsites counts sites owned by any of this user's domains, capped at 1000
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

// HandleInstallPage renders the install form and handles the early over-limit check for a POST; the streaming install itself is handled separately by HandleInstall
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

// HandleInstall drives the NDJSON-streamed install of one app-type service. Some failure branches deliberately end the stream early with a bare `return` instead of an {"error":...} event - once streaming has started, there's no clean way to signal a mid-stream failure besides stopping, so the browser just sees it stall until the front-end's 5-minute hang timer fires, matching how the UI treats any interrupted install stream.
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

	// a root-level install gets a catch-all `ProxyPass /` in the vhost, which Apache matches before any other path (mod_proxy matches by file order, not specificity) - so a root app would silently break every subdirectory site on the domain, and the reverse breaks too (a subdirectory app behind an existing root app is unreachable). Neither can be fixed by reordering, so both cases are blocked outright.
	if subdirectory == "" {
		var otherSitesCount int
		if countErr := a.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM sites WHERE site_name LIKE ?", topDomain+"/%").Scan(&otherSitesCount); countErr == nil && otherSitesCount > 0 {
			emit(map[string]any{"error": "This domain already has " + strconv.Itoa(otherSitesCount) + " other site(s) installed in a subdirectory. Installing " + kind.DisplayAppType + " at the domain root would make those unreachable (a root app's reverse proxy catches every request on the domain, including existing subdirectories) - install into a subdirectory instead, or remove the other site(s) first."})
			return
		}
	} else {
		var rootAppType string
		if scanErr := a.DB.QueryRowContext(ctx, "SELECT type FROM sites WHERE site_name = ? AND type IN ('NodeJS','Python','Ruby')", topDomain).Scan(&rootAppType); scanErr == nil {
			emit(map[string]any{"error": "This domain already has a " + rootAppType + " application installed at its root. That root app's reverse proxy catches every request on the domain, so a new site in a subdirectory would never actually be reachable - remove the root application first, or choose a different domain."})
			return
		}
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

	// the vendored template hardcodes "pids: 100" - swap it for an env-var placeholder before the SERVICE_NAME replace below, so PIDs becomes as editable as CPU/RAM post-install (see handlePM2Update in pm2.go)
	templateStr = strings.ReplaceAll(templateStr, "pids: 100", `pids: "${SERVICE_NAME_`+kind.PyOrNode+`_PIDS:-100}"`)

	resolvedCommand := buildAppRunCommand(kind, requirements, customCmd, startupFile, gitRepoURL)

	nestedCommandPattern := "${SERVICE_NAME_" + kind.PyOrNode + "_REQUIREMENTS:+" + kind.InstallToken + " &&} " +
		"${SERVICE_NAME_" + kind.PyOrNode + "_CUSTOM_CMD:-" + kind.RunToken + " ${SERVICE_NAME_" + kind.PyOrNode + "_STARTUP_FILE:-" + kind.DefaultStartupFile + "}}"

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
		if strings.HasPrefix(line, "services:") {
			insertPosition = i + 1
			break
		}
	}
	if insertPosition == -1 {
		emit(map[string]any{"error": "'services:' section not found in docker-compose.yml."})
		return
	}
	for _, line := range composeLines {
		if strings.Contains(line, serviceName+":") {
			emit(map[string]any{"error": "Service '" + serviceName + "' already exists, please use a different name."})
			return
		}
	}

	// anchored on "services:" (not "networks:") so this lands as a child of services: regardless of the file's top-level key order - docker.SaveCompose's YAML round-trip can reorder keys, and anchoring on "networks:" used to silently corrupt the file once that assumption broke. See composeServiceIndent/indentComposeService for why the insert also has to match the file's real indent.
	indentedServiceStr := indentComposeService(newServiceStr, composeServiceIndent(composeLines[insertPosition:]))

	newLines := make([]string, 0, len(composeLines)+1)
	newLines = append(newLines, composeLines[:insertPosition]...)
	newLines = append(newLines, "\n"+indentedServiceStr+"\n")
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
	startResult := docker.StartOrStopContainer(ctx, userContext, serviceName, "activate", "")
	if !startResult.Success {
		// `podman-compose up` itself failed - no container was ever created, so containerFailureDetail below has nothing to inspect. Report the real reason now rather than falling through to the poll loop, which would just time out.
		_ = copyFile(composeBackup, composeFile)
		_ = copyFile(envBackup, envFile)
		_ = os.Remove(lockPath)
		emit(map[string]any{"error": "Container failed to start. " + startResult.Message})
		return
	}

	emit(map[string]any{"status": "Checking if container started..."})
	// a single fixed sleep-then-check would false-negative here: the entrypoint typically runs npm/pip install first, plus an image pull on first use, both easily exceeding a few seconds, and State.Running stays false the whole time it's mid-install/mid-pull, not true-with-a-delay. Poll generously instead.
	isRunning := false
	inspectArgv := podmanmanager.PodmanArgv(userContext, "inspect", "-f", "{{.State.Running}}", serviceName)
	for attempt := 0; attempt < containerStartPollAttempts; attempt++ {
		time.Sleep(containerStartPollInterval)
		// podmanmanager.Command, not a bare exec.CommandContext - a per-user context needs CONTAINER_HOST pointed at that user's podman.sock, which only podmanmanager.Command's env sets. Without it this was silently inspecting root's default podman instance instead, so it could never observe State.Running=true even though the container was actually fine.
		inspectOut, inspectErr := podmanmanager.Command(ctx, userContext, inspectArgv).Output()
		if inspectErr == nil && strings.TrimSpace(string(inspectOut)) == "true" {
			isRunning = true
			break
		}
	}

	if !isRunning {
		// grab the reason before the container is torn away below - `podman rm -f` takes its logs and inspect state with it, so this is the only chance to see why it actually failed
		failureDetail := containerFailureDetail(ctx, userContext, serviceName)
		_ = copyFile(composeBackup, composeFile)
		_ = copyFile(envBackup, envFile)
		rmArgv := podmanmanager.PodmanArgv(userContext, "rm", "-f", serviceName)
		_ = podmanmanager.Command(ctx, userContext, rmArgv).Run()
		_ = os.Remove(lockPath)
		errMsg := "Container failed to start."
		if failureDetail != "" {
			errMsg += " " + failureDetail
		} else {
			errMsg += " Please check allocated resources."
		}
		emit(map[string]any{"error": errMsg})
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

// formatPyFloat formats a float for the .env values written here, keeping at least one digit after the decimal point - unlike strconv.FormatFloat's -1 precision, which drops a bare ".0"
func formatPyFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
