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

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// tinyFileManagerSourceFile is the only source TinyFileManager ships: a single PHP file on the master branch, no tagged releases and no composer.json - always installs current master
const tinyFileManagerSourceFile = "https://raw.githubusercontent.com/prasathmani/tinyfilemanager/master/tinyfilemanager.php"

// tinyFileManagerVersion is a static placeholder recorded in the sites table - no real versioning upstream, same convention tinyphotogallery uses for its own "main" branch install
const tinyFileManagerVersion = "latest"

// tinyFileManagerAuthUsersRE matches the entire default $auth_users = array( ... ); block near the top of the downloaded file, from the literal opener through the first ");" that follows - neither a bcrypt hash nor the trailing comments upstream ever contain ");", so the non-greedy match is safe
var tinyFileManagerAuthUsersRE = regexp.MustCompile(`(?s)\$auth_users\s*=\s*array\(.*?\);`)

// handleInstallStream drives a TinyFileManager install end to end over NDJSON: download tinyfilemanager.php, hash the admin password inside the target php container (matching that container's own password_verify()), rewrite the default $auth_users array down to just the one admin account, fix ownership, record the site - no database and no CLI installer
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { web.WriteNDJSON(w, flusher, canFlush, v) }

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
	if err := appkit.CreateLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}
	defer appkit.RemoveLockFile(currentUsername)

	dom, found, dbErr := appkit.LookupDomainByID(ctx, a, domainID)
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
	if !cmsapp.IsValidSubdirectory(subdirectory) {
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
	if !cmsapp.EnsureContainerRunning(ctx, userContext, phpContainer) {
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
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Hashing admin password"})
	passwordHash, hashErr := hashTinyFileManagerPassword(ctx, userContext, phpContainer, adminPassword)
	if hashErr != nil {
		emit(map[string]any{"error": "Failed to hash admin password: " + hashErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Writing admin credentials"})
	filePath := filepath.Join(hostOSPath, "tinyfilemanager.php")
	if confErr := writeTinyFileManagerAuthUsers(filePath, adminUsername, passwordHash); confErr != nil {
		emit(map[string]any{"error": "Failed to write admin credentials: " + confErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

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
	web.Flash(a, w, r, "success", web.Tr(a, r, "TinyFileManager installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "TinyFileManager installation completed!", "admin_user": adminUsername})
}

// runTinyFileManagerInstall downloads tinyfilemanager.php from the master branch into installPath - that's the entire upstream install procedure for a single-file app with no build step
func runTinyFileManagerInstall(ctx context.Context, userContext, phpContainer, installPath string) ([]byte, error) {
	script := `set -e
mkdir -p ` + installPath + `
curl -sL -o ` + installPath + `/tinyfilemanager.php ` + tinyFileManagerSourceFile

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

// hashTinyFileManagerPassword shells out to password_hash() inside the same php-fpm container the site will run in, rather than reimplementing bcrypt in Go, guaranteeing the same hash format that container's password_verify() will accept - same technique dokuwiki/install.go uses; the password is passed as a separate argv element, not interpolated into a shell string
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

// writeTinyFileManagerAuthUsers reads the just-downloaded tinyfilemanager.php from the host OS path, replaces its default $auth_users block with a single admin entry, and writes it back - runs host-side, mirroring the htmlVolume-prefix pattern tinyphotogallery/sofawiki use
func writeTinyFileManagerAuthUsers(filePath, adminUsername, passwordHash string) error {
	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return readErr
	}

	replacement := "$auth_users = array(\n    '" + cmsapp.EscapePHPSingleQuoted(adminUsername) + "' => '" + cmsapp.EscapePHPSingleQuoted(passwordHash) + "'\n);"

	// regexp.ReplaceAll interprets "$" in its replacement as a submatch reference, which would mangle "$auth_users" above, so splice the match location manually instead
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
