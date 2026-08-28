package dokuwiki

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

// dokuwikiStableTarball is DokuWiki's version-agnostic "always current
// stable" download, confirmed live to resolve to a real dated release
// (2026-07-14b "Mort" as of this module being written) wrapped in a single
// top-level dokuwiki-<version>/ directory inside the tarball.
const dokuwikiStableTarball = "https://download.dokuwiki.org/src/dokuwiki/dokuwiki-stable.tgz"

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

func formOr(r *http.Request, key, def string) string {
	if v := r.FormValue(key); v != "" {
		return v
	}
	return def
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

// phpVersionBelow reports whether version (e.g. "7.2", "8.5") is older
// than wantMajor.wantMinor - the opposite guard SofaWiki needs (DokuWiki
// needs a PHP floor, not a ceiling). An unparseable version is treated as
// too old, so an unexpected format fails safe.
func phpVersionBelow(version string, wantMajor, wantMinor int) bool {
	major, minor := 0, 0
	if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
		return true
	}
	if major != wantMajor {
		return major < wantMajor
	}
	return minor < wantMinor
}

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// hashDokuwikiPassword shells out to `php -r 'echo password_hash(...)'`
// inside the same php-fpm container the site will run in, rather than
// reimplementing bcrypt in Go - guarantees byte-for-byte the same hash
// format DokuWiki's own install.php produces (confirmed live: PassHash's
// hash_bcrypt() is a thin wrapper around PHP's crypt()/password_hash with
// the standard $2y$ prefix), so there's no risk of a subtly incompatible
// hash locking the generated admin account out. Password is passed as a
// separate argv element (not interpolated into a shell string), so it's
// safe regardless of its contents.
func hashDokuwikiPassword(ctx context.Context, userContext, phpContainer, password string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", "-r",
		"echo password_hash($argv[1], PASSWORD_BCRYPT);", "--", password)
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

// handleInstallStream drives a DokuWiki install end to end, streaming
// NDJSON progress events: download+extract the current stable tarball,
// write conf/local.php + conf/users.auth.php + conf/acl.auth.php directly
// (replicating exactly what DokuWiki's own install.php wizard would have
// written for a single-admin, ACL-enabled site - see dokuwiki.go's package
// doc comment), fix ownership, record the site, and remove install.php
// since it's never needed and DokuWiki's own docs recommend deleting it.
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

	adminUser := strings.ToLower(strings.TrimSpace(formOr(r, "admin_user", "admin")))
	if adminUser == "" || strings.ContainsAny(adminUser, " :\t\n") {
		emit(map[string]any{"error": "Invalid admin username."})
		return
	}
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16)
	}
	adminFullName := formOr(r, "admin_full_name", "Administrator")
	siteTitle := formOr(r, "site_title", dom.DomainURL)

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

	// DokuWiki needs PHP 7.4+ - the opposite constraint of SofaWiki's PHP
	// 8-fatals ceiling. Confirmed against a live copy of the current
	// stable release; older releases needed even less, but this server
	// only ever offers 7.4+ containers anyway.
	if !isLitespeed && phpVersionBelow(phpVersion, 7, 4) {
		emit(map[string]any{"error": "DokuWiki requires PHP 7.4 or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version and try again."})
		return
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

	emit(map[string]any{"status": "Downloading and extracting DokuWiki"})
	installedVersion, out, runErr := runDokuwikiExtract(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Hashing admin password"})
	passwordHash, hashErr := hashDokuwikiPassword(ctx, userContext, phpContainer, adminPassword)
	if hashErr != nil {
		emit(map[string]any{"error": "Failed to hash admin password: " + hashErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Writing DokuWiki configuration"})
	if confErr := writeDokuwikiConfig(ctx, userContext, phpContainer, installPath, dokuwikiConfigParams{
		Title:      siteTitle,
		AdminUser:  adminUser,
		AdminHash:  passwordHash,
		AdminName:  adminFullName,
		AdminEmail: formOr(r, "admin_email", "admin@"+dom.DomainURL),
	}); confErr != nil {
		emit(map[string]any{"error": "Failed to write configuration: " + confErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, installedVersion, "dokuwiki"); insertErr != nil {
		emit(map[string]any{"error": "DokuWiki installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed DokuWiki on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "DokuWiki installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "DokuWiki installation completed!", "admin_user": adminUser, "admin_password": adminPassword})
}

// runDokuwikiExtract downloads the current stable tarball and extracts it
// into installPath, returning the installed version string (read from the
// extracted VERSION file). The tarball wraps everything in a single
// top-level dokuwiki-<version>/ directory (confirmed live), so this
// extracts to a scratch location first and copies that directory's
// *contents* into installPath - same reasoning as sofawiki's
// runSofawikiExtract for why `cp -a` beats `mv` here (a root, no-
// subdirectory install's installPath already exists as the domain's
// docroot, even though empty).
func runDokuwikiExtract(ctx context.Context, userContext, phpContainer, installPath string) (version string, out []byte, err error) {
	scratch := "/tmp/openpanel-dokuwiki-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	script := `set -e
rm -rf ` + scratch + ` ` + scratch + `.tgz
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.tgz ` + dokuwikiStableTarball + `
tar -xzf ` + scratch + `.tgz -C ` + scratch + `
SRC=$(find ` + scratch + ` -mindepth 1 -maxdepth 1 -type d | head -1)
mkdir -p ` + installPath + `
cp -a "$SRC"/. ` + installPath + `/
cat "$SRC"/VERSION
rm -rf ` + scratch + ` ` + scratch + `.tgz
rm -f ` + installPath + `/install.php`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	out, err = podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if err != nil {
		return "", out, err
	}
	version = strings.TrimSpace(string(out))
	if version == "" {
		version = "unknown"
	}
	return version, out, nil
}

type dokuwikiConfigParams struct {
	Title      string
	AdminUser  string
	AdminHash  string
	AdminName  string
	AdminEmail string
}

// writeDokuwikiConfig writes conf/local.php, conf/users.auth.php and
// conf/acl.auth.php directly inside the container, byte-for-byte
// replicating what DokuWiki's own install.php wizard's store_data()
// function writes for a single-admin, ACL-enabled, open-read wiki -
// confirmed against that function's source on a live extracted copy
// rather than guessed. AdminHash must already be a valid DokuWiki
// password hash (see hashDokuwikiPassword) - this function only writes
// files, it never handles a clear-text password.
func writeDokuwikiConfig(ctx context.Context, userContext, phpContainer, installPath string, p dokuwikiConfigParams) error {
	localPHP := "<?php\n" +
		"$conf['title'] = '" + escapePHPSingleQuoted(p.Title) + "';\n" +
		"$conf['lang'] = 'en';\n" +
		"$conf['license'] = 'cc-by-sa';\n" +
		"$conf['useacl'] = 1;\n" +
		"$conf['superuser'] = '@admin';\n"

	usersAuth := "# users.auth.php\n" +
		"# <?php exit()?>\n" +
		"# Don't modify the lines above\n" +
		"#\n" +
		"# Auto-generated by OpenPanel install\n" +
		"#\n" +
		"# Format:\n" +
		"# login:passwordhash:Real Name:email:groups,comma,separated\n\n" +
		p.AdminUser + ":" + p.AdminHash + ":" + escapeAuthField(p.AdminName) + ":" + escapeAuthField(p.AdminEmail) + ":admin,user\n"

	aclAuth := "# acl.auth.php\n" +
		"# <?php exit()?>\n" +
		"# Don't modify the lines above\n" +
		"#\n" +
		"# Auto-generated by OpenPanel install\n\n" +
		"*               @ALL          8\n"

	script := `set -e
mkdir -p ` + installPath + `/conf
cat > ` + installPath + `/conf/local.php << 'OPENPANEL_EOF'
` + localPHP + `OPENPANEL_EOF
cat > ` + installPath + `/conf/users.auth.php << 'OPENPANEL_EOF'
` + usersAuth + `OPENPANEL_EOF
cat > ` + installPath + `/conf/acl.auth.php << 'OPENPANEL_EOF'
` + aclAuth + `OPENPANEL_EOF`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// escapeAuthField strips the colon delimiter users.auth.php's line format
// depends on, since a stray colon in a display name/email would silently
// shift every field after it.
func escapeAuthField(value string) string {
	return strings.ReplaceAll(value, ":", "")
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
