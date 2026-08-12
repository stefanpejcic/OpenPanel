package phpapp

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// countUserWebsites counts sites owned by any of this user's domains,
// capped at 1000 - same query appinstall.countUserWebsites uses.
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
func HandleInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed.")
	} else if r.Method == http.MethodPost {
		HandleInstall(a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderInstallPage(a, w, r, domains)
}

// HandleInstall drives the NDJSON-streamed install of a PHP/Composer
// project into an existing domain's docroot. Unlike appinstall.HandleInstall
// this never touches docker-compose.yml, .env container variables, or a
// webserver reverse-proxy config - the domain's existing vhost already
// routes to the php-fpm container this uses.
func HandleInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	if domainID == "" {
		emit(map[string]any{"error": "Missing required field: domain"})
		return
	}

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	lockDir := "/etc/openpanel/openpanel/core/users/" + currentUsername
	_ = os.MkdirAll(lockDir, 0o755)
	lockPath := lockDir + "/krompir.lock"
	if lockErr := os.WriteFile(lockPath, nil, 0o644); lockErr != nil {
		emit(map[string]any{"error": "Error creating " + lockPath + ": " + lockErr.Error()})
		return
	}
	defer os.Remove(lockPath)

	var topDomain string
	var docrootNull, phpVersionNull sql.NullString
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_id = ?", domainID)
	if scanErr := row.Scan(&topDomain, &docrootNull, &phpVersionNull); scanErr != nil {
		emit(map[string]any{"error": "Domain not found"})
		return
	}
	docroot := docrootNull.String

	if !a.CheckDomainBelongsToUser(ctx, userID, topDomain) {
		return
	}

	emit(map[string]any{"status": "Validating provided data"})
	subdirectory := strings.ToLower(strings.ReplaceAll(r.FormValue("subdirectory"), " ", ""))
	initialProject := strings.TrimSpace(r.FormValue("initial_project"))
	autorunComposerInstall := normalizeCheckbox(r.FormValue("autorun_composer_install"))
	composerOptimizeAutoloader := normalizeCheckbox(r.FormValue("composer_optimize_autoloader"))

	if !isValidSubdirectory(subdirectory) {
		emit(map[string]any{"error": "Invalid subdirectory."})
		return
	}
	if !isValidInitialProject(initialProject) {
		emit(map[string]any{"error": "Initial project must be empty, an https:// archive URL ending in .zip/.tar.gz/.tgz/.tar, or a Composer package name (vendor/package)."})
		return
	}

	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, topDomain)
	if phpVersion == "" || phpVersion == "/" {
		emit(map[string]any{"error": "This domain has no PHP version assigned. Set one from PHP Selector first."})
		return
	}
	_ = phpVersionNull // php_version snapshot is looked up live above instead; column kept for future use

	installPath := docroot
	selectedDomain := topDomain
	if subdirectory != "" {
		// docroot may already end in "/" (e.g. the "/var/www/html/"
		// fallback used when a domain has no docroot of its own set) - a
		// naive "docroot + / + subdirectory" join can then produce a
		// double slash, which breaks `composer create-project`'s target
		// path resolution inside the container.
		installPath = strings.TrimSuffix(docroot, "/") + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
	if !ensureContainerRunning(ctx, userContext, phpContainer) {
		emit(map[string]any{"error": "PHP container failed to start. Please check it from Services."})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(installPath, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	// Composer's own --working-dir flag, not `podman exec -w` - the
	// vendored image's `composer` binary is itself a wrapper
	// (/usr/local/aliases/composer -> `exec in-app /usr/local/bin/composer`)
	// whose `in-app` layer resets the process's actual working directory
	// regardless of `-w`/an explicit `cd`, so composer sees /var/www/html
	// no matter what podman exec's workdir was set to. --working-dir is a
	// composer CLI arg, untouched by that wrapper, and works correctly.
	composerBase := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"), "--working-dir="+installPath)

	if initialProject != "" {
		if isArchiveURL(initialProject) {
			// Unlike composer create-project (which creates its own target
			// directory), extraction needs somewhere to extract into.
			if mkErr := os.MkdirAll(hostOSPath, 0o755); mkErr != nil {
				emit(map[string]any{"error": "Error creating document root: " + mkErr.Error()})
				return
			}
			emit(map[string]any{"status": "Downloading and extracting " + initialProject})
			if extractErr := downloadAndExtractInitialProject(ctx, initialProject, hostOSPath); extractErr != nil {
				emit(map[string]any{"error": "Error extracting initial project: " + extractErr.Error()})
				return
			}
		} else {
			// Deliberately NOT pre-creating hostOSPath here: composer
			// create-project makes its own target directory, and a host-side
			// os.MkdirAll immediately beforehand raced its container-side
			// mkdir over the rootless bind mount in testing - composer would
			// intermittently fail to create vendor/ inside a directory this
			// process had *just* created microseconds earlier, even though
			// the directory was otherwise identical (same owner/mode) to one
			// composer created unassisted. Let composer own the whole path.
			emit(map[string]any{"status": "Creating Composer project " + initialProject})
			argv := append(append([]string{}, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer")...),
				"create-project", initialProject, installPath, "--no-interaction")
			out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
			appendComposerLog(currentUsername, selectedDomain, "create-project", out)
			if runErr != nil {
				emit(map[string]any{"error": "composer create-project failed: " + strings.TrimSpace(string(out))})
				return
			}
		}
	} else if autorunComposerInstall {
		// No initial_project means the directory (with its own
		// composer.json) is expected to already exist - composer install
		// can't materialize one from nothing.
		if _, statErr := os.Stat(hostOSPath); statErr != nil {
			emit(map[string]any{"error": "Directory " + installPath + " does not exist. Set an initial project, or create the directory (with a composer.json) first."})
			return
		}
	}

	if autorunComposerInstall {
		emit(map[string]any{"status": "Running composer install"})
		args := []string{"install", "--no-interaction"}
		if composerOptimizeAutoloader {
			args = append(args, "--optimize-autoloader")
		}
		argv := append(append([]string{}, composerBase...), args...)
		out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
		appendComposerLog(currentUsername, selectedDomain, "install", out)
		if runErr != nil {
			emit(map[string]any{"error": "composer install failed: " + strings.TrimSpace(string(out))})
			return
		}
	}

	emit(map[string]any{"status": "Saving application settings"})
	prefix := phpAppEnvPrefix(selectedDomain)
	docker.SetEnvValue(userContext, prefix+"INITIAL_PROJECT", initialProject)
	docker.SetEnvValue(userContext, prefix+"AUTORUN_COMPOSER_INSTALL", boolEnvValue(autorunComposerInstall))
	docker.SetEnvValue(userContext, prefix+"COMPOSER_OPTIMIZE_AUTOLOADER", boolEnvValue(composerOptimizeAutoloader))
	docker.SetEnvValue(userContext, prefix+"WORKDIR", installPath)

	emit(map[string]any{"status": "Installation finished, adding the new application to SiteManager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, version, type, path) VALUES (?, ?, ?, 'PHP', ?)",
		selectedDomain, domainID, phpVersion, subdirectory); insertErr != nil {
		emit(map[string]any{"error": "An error occurred: " + insertErr.Error()})
		return
	}

	emit(map[string]any{"status": "New PHP application setup completed!"})
	_ = logger.RecordUserAction(a.Config, currentUsername, "created a new PHP application on domain "+selectedDomain, ipAddress)
}

// ensureContainerRunning starts container if it's not already running,
// polling briefly for it to come up (mirrors wordpress.waitForWPAvailable's
// shape, but keyed off container status rather than a WP-CLI probe).
func ensureContainerRunning(ctx context.Context, userContext, container string) bool {
	if docker.IsServiceRunning(ctx, userContext, container) {
		return true
	}
	docker.StartOrStopContainer(ctx, userContext, container, "activate", "detached")
	const attempts = 15
	for i := 0; i < attempts; i++ {
		time.Sleep(2 * time.Second)
		if docker.IsServiceRunning(ctx, userContext, container) {
			return true
		}
	}
	return false
}

func normalizeCheckbox(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "on", "true", "yes":
		return true
	default:
		return false
	}
}

func boolEnvValue(b bool) string {
	if b {
		return "1"
	}
	return ""
}

// phpAppEnvPrefix derives the .env key prefix for a PHP app's settings from
// its site name (domain, optionally "/subdir") - there's no dedicated
// container to key these off of the way appinstall keys CPU/RAM/etc off a
// service name, so this is a synthetic key namespace instead.
func phpAppEnvPrefix(siteName string) string {
	return docker.ServiceKeyPrefix(strings.ReplaceAll(siteName, "/", "_")) + "_PHP_"
}
