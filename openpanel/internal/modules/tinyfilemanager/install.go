package tinyfilemanager

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
)

// tinyFileManagerSourceFile is the only source TinyFileManager ships: a
// single PHP file on the master branch, no tagged releases and no
// composer.json - always installs current master.
const tinyFileManagerSourceFile = "https://raw.githubusercontent.com/prasathmani/tinyfilemanager/master/tinyfilemanager.php"

// tinyFileManagerVersion is a static placeholder recorded in the sites
// table - there is no real versioning upstream (no tags/releases), same
// convention tinyphotogallery uses for its own "main" branch install.
const tinyFileManagerVersion = "latest"

// tinyFileManagerAuthUsersRE matches the entire default
// `$auth_users = array( ... );` block near the top of the downloaded
// file, from the literal `$auth_users = array(` through the first `);`
// that follows - confirmed against a live download of the file that the
// block spans two sample lines like:
//
//	$auth_users = array(
//	    'admin' => '$2y$10$...', //admin@123
//	    'user' => '$2y$10$...' //12345
//	);
//
// Neither a bcrypt hash nor the trailing comments upstream ever contain
// ");", so the non-greedy match below is safe to stop at the first one.
var tinyFileManagerAuthUsersRE = regexp.MustCompile(`(?s)\$auth_users\s*=\s*array\(.*?\);`)

// handleInstallPage renders the install form / checks the plan's site
// limit for a GET, and hands POST off to handleInstallStream.
func handleInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefault(plan.WebsitesLimit, 0)
	websiteCount, _ := countUserWebsites(a, userID)

	if websitesLimit != 0 && websiteCount >= websitesLimit {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/x-ndjson")
			flusher, canFlush := w.(http.Flusher)
			writeNDJSON(w, flusher, canFlush, map[string]any{"error": "You have reached the maximum number of sites allowed." + plan.UpgradeMessage()})
			return
		}
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed."+plan.UpgradeMessage())
	} else if r.Method == http.MethodPost {
		handleInstallStream(a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderInstallPage(a, w, r, domains)
}

// ensureContainerRunning starts the container if it isn't already running,
// polling briefly for it to come up.
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

// handleInstallStream drives a TinyFileManager install end to end,
// streaming NDJSON progress events to the client: download
// tinyfilemanager.php, hash the provided admin password inside the target
// php container (so the hash format matches what that container's own
// password_verify() will accept), rewrite the file's default sample
// $auth_users array down to just the one admin account provided, fix
// ownership, record the site. There is no database and no CLI installer
// to drive.
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

	ipAddress := reqip.ClientIP(r)
	domainID := r.FormValue("domain_id")
	if domainID == "" {
		emit(map[string]any{"error": "Missing required field: domain"})
		return
	}

	adminUsername := strings.TrimSpace(r.FormValue("admin_username"))
	adminPassword := r.FormValue("admin_password")
	if adminUsername == "" || adminPassword == "" {
		emit(map[string]any{"error": "Admin username and password are required."})
		return
	}

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	if err := createLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}
	defer removeLockFile(currentUsername)

	dom, found, dbErr := lookupDomainByID(ctx, a, domainID)
	if dbErr != nil {
		emit(map[string]any{"error": "An error occurred fetching docroot for domain from database."})
		return
	}
	if !found {
		emit(map[string]any{"error": "Domain not found"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, dom.DomainURL) {
		return
	}

	emit(map[string]any{"status": "Validating provided data"})
	subdirectory := strings.ToLower(strings.ReplaceAll(r.FormValue("subdirectory"), " ", ""))
	if !isValidSubdirectory(subdirectory) {
		emit(map[string]any{"error": "Invalid subdirectory."})
		return
	}

	docroot := dom.Docroot.String
	selectedDomain := dom.DomainURL
	installPath := docroot
	if subdirectory != "" {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpVersion := dom.PHPVersion.String
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

	if _, statErr := os.Stat(hostOSPath); statErr == nil {
		if entries, readErr := os.ReadDir(hostOSPath); readErr == nil && len(entries) > 0 {
			emit(map[string]any{"error": "Directory " + installPath + " already exists and is not empty."})
			return
		}
	}

	emit(map[string]any{"status": "Downloading TinyFileManager"})
	out, runErr := runTinyFileManagerInstall(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Hashing admin password"})
	passwordHash, hashErr := hashTinyFileManagerPassword(ctx, userContext, phpContainer, adminPassword)
	if hashErr != nil {
		emit(map[string]any{"error": "Failed to hash admin password: " + hashErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Writing admin credentials"})
	filePath := filepath.Join(hostOSPath, "tinyfilemanager.php")
	if confErr := writeTinyFileManagerAuthUsers(filePath, adminUsername, passwordHash); confErr != nil {
		emit(map[string]any{"error": "Failed to write admin credentials: " + confErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	adminEmail := "admin@" + dom.DomainURL
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, tinyFileManagerVersion, "tinyfilemanager"); insertErr != nil {
		emit(map[string]any{"error": "TinyFileManager installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed TinyFileManager on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "TinyFileManager installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "TinyFileManager installation completed!", "admin_user": adminUsername})
}

// runTinyFileManagerInstall downloads tinyfilemanager.php from the master
// branch into installPath - that is the entire upstream install procedure
// per the project's README (a single-file app, no build step).
func runTinyFileManagerInstall(ctx context.Context, userContext, phpContainer, installPath string) ([]byte, error) {
	script := `set -e
mkdir -p ` + installPath + `
curl -sL -o ` + installPath + `/tinyfilemanager.php ` + tinyFileManagerSourceFile

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

// hashTinyFileManagerPassword shells out to
// `php -r 'echo password_hash($argv[1], PASSWORD_DEFAULT);'` inside the
// same php-fpm container the site will run in, rather than reimplementing
// bcrypt in Go - guarantees byte-for-byte the same hash format that
// container's own password_verify() call (inside tinyfilemanager.php)
// will accept, the same technique dokuwiki/install.go uses for its own
// admin account. Password is passed as a separate argv element (not
// interpolated into a shell string), so it's safe regardless of its
// contents.
func hashTinyFileManagerPassword(ctx context.Context, userContext, phpContainer, password string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", "-r",
		"echo password_hash($argv[1], PASSWORD_DEFAULT);", "--", password)
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return "", err
	}
	hash := strings.TrimSpace(string(out))
	if hash == "" {
		return "", fmt.Errorf("empty hash returned")
	}
	return hash, nil
}

// writeTinyFileManagerAuthUsers reads the just-downloaded
// tinyfilemanager.php from the host OS path (bind-mounted from the
// container), replaces its entire default `$auth_users = array(...);`
// block (two sample users) with a single entry for the provided admin
// account, and writes the file back. Runs on the host side, not inside
// the container, mirroring the htmlVolume-prefix pattern
// tinyphotogallery/sofawiki use for host-side file operations.
func writeTinyFileManagerAuthUsers(filePath, adminUsername, passwordHash string) error {
	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return readErr
	}

	replacement := "$auth_users = array(\n    '" + escapePHPSingleQuoted(adminUsername) + "' => '" + escapePHPSingleQuoted(passwordHash) + "'\n);"

	// Note: regexp.ReplaceAll interprets "$" in its replacement argument as
	// a submatch reference (see Expand), which would mangle the literal
	// "$auth_users" text above - so this splices the match location
	// manually instead of using ReplaceAll.
	loc := tinyFileManagerAuthUsersRE.FindIndex(content)
	if loc == nil {
		return fmt.Errorf("could not find $auth_users array in downloaded file")
	}
	newContent := make([]byte, 0, len(content)-((loc[1]-loc[0])-len(replacement)))
	newContent = append(newContent, content[:loc[0]]...)
	newContent = append(newContent, []byte(replacement)...)
	newContent = append(newContent, content[loc[1]:]...)

	return os.WriteFile(filePath, newContent, 0o644)
}

// escapePHPSingleQuoted escapes value for embedding inside a PHP
// single-quoted string literal - only backslash and single-quote need
// escaping in that context. Same helper as drupal/clone.go's.
func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// emitCleanupFiles removes a failed install's partially-created directory.
func emitCleanupFiles(ctx context.Context, userContext, phpContainer, installPath string, emit func(map[string]any)) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath)
	if err := podmanmanager.Command(ctx, userContext, argv).Run(); err != nil {
		emit(map[string]any{"status": "Cleanup: failed to remove " + installPath + ": " + err.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: removed " + installPath})
}

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}
