package mediawiki

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleInstallStream drives a MediaWiki install end to end over NDJSON: download the packaged tarball from releases.wikimedia.org, extract it directly into the docroot (flat, like Joomla - no separate web/ or public/ subdirectory to reconcile), create a MySQL database, run MediaWiki's own `maintenance/install.php` non-interactive CLI installer, deploy the admin login helper, then register a per-minute `maintenance/runJobs.php` job via crons.AddJob (MediaWiki's job queue - link tables, search index, email - does not run without it)
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	in, ok := cmsapp.StartInstall(a, w, r)
	if !ok {
		return
	}
	defer appkit.RemoveLockFile(in.Username)
	ctx := in.Ctx
	userID := in.UserID
	currentUsername := in.Username
	userContext := in.UserContext
	emit := in.Emit
	ipAddress := in.IP
	domainID := in.DomainID
	dom := in.Domain
	subdirectory := in.Subdirectory

	siteName := web.FormOr(r, "site_name", "MediaWiki")
	adminUsername := web.FormOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16) + "!A1"
	}
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "mediawiki_" + strings.ToLower(appkit.RandomString(6))
	}
	dbUser := strings.ToLower(r.FormValue("db_user"))
	if dbUser == "" {
		dbUser = strings.ToLower(appkit.RandomString(10))
	}
	dbPassword := r.FormValue("db_password")
	if dbPassword == "" {
		dbPassword = appkit.RandomString(16)
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
	if mkErr := os.MkdirAll(hostOSPath, 0o755); mkErr != nil {
		emit(map[string]any{"error": "Error creating document root: " + mkErr.Error()})
		return
	}

	version := strings.TrimSpace(r.FormValue("mediawiki_version"))
	if version == "" {
		var latestErr error
		version, latestErr = latestMediaWikiVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest MediaWiki version: " + latestErr.Error()})
			return
		}
	}
	branch := mediawikiBranchForVersion(version)
	if branch == "" {
		emit(map[string]any{"error": "Could not determine MediaWiki release branch for version " + version})
		return
	}

	archiveName := "mediawiki-" + version + ".tar.gz"
	archiveDir := "/etc/openpanel/mediawiki/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://releases.wikimedia.org/mediawiki/" + branch + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			_ = os.Remove(archivePath)
			emit(map[string]any{"error": "Error downloading MediaWiki " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + installPath})
	if runErr := exec.CommandContext(ctx, "tar", "-xzf", archivePath, "-C", hostOSPath, "--strip-components=1").Run(); runErr != nil {
		emit(map[string]any{"error": "Error extracting MediaWiki archive: " + runErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// MediaWiki's minimum required PHP version climbs across branches (1.39 LTS wants 7.4.3+, 1.46 wants 8.3+) - without this check an incompatible domain PHP version fails deep inside install.php with an opaque parse error instead of a clear message
	if minPHP := minPHPVersionFromComposerJSON(hostOSPath); minPHP != "" && cmsapp.CompareVersions(phpVersion, minPHP) < 0 {
		emit(map[string]any{"error": "MediaWiki " + version + " requires PHP " + minPHP + " or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version (PHP Settings) or pick an older MediaWiki version, then try again."})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// host-side chown, not `podman exec ... chown` - the archive was extracted host-side, so the files are owned by this process's own real unmapped UID, and a rootless container's "root" cannot chown files outside its own user-namespace UID mapping
	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr != nil {
		emit(map[string]any{"error": "Could not determine file owner: " + uidErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}
	uidStr := strconv.Itoa(uid)
	if chownErr := exec.Command("chown", "-R", uidStr+":"+uidStr, hostOSPath).Run(); chownErr != nil {
		emit(map[string]any{"error": "Could not set file ownership: " + chownErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	mysqlVersion := mysql.GetMySQLVersion(ctx, a, userContext)
	if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
		emit(map[string]any{"status": "Starting " + mysqlVersion + " container.."})
		docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
	}
	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"error": "The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			return
		}
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan." + a.UpgradeMessageForUser(ctx, userID)})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Creating database " + dbName + " and user " + dbUser})
	const dbHost = "%"
	queries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dbUser + "'@'" + dbHost + "' IDENTIFIED BY '" + dbPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dbName + "`.* TO '" + dbUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			appkit.InvalidateMySQLCaches(ctx, a, userContext, currentUsername)
			emit(map[string]any{"error": "Error creating MySQL database and user: " + execErr.Error()})
			emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
			return
		}
	}
	appkit.InvalidateMySQLCaches(ctx, a, userContext, currentUsername)

	scriptPath := ""
	if subdirectory != "" {
		scriptPath = "/" + subdirectory
	}

	emit(map[string]any{"status": "Running MediaWiki CLI installer (maintenance/install.php)"})
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		installPath+"/maintenance/install.php",
		"--dbtype=mysql",
		"--dbserver="+mysql.AppDBHost(userContext, mysqlVersion),
		"--dbname="+dbName,
		"--dbuser="+dbUser,
		"--dbpass="+dbPassword,
		"--dbprefix=mw_",
		"--server=https://"+dom.DomainURL,
		"--scriptpath="+scriptPath,
		"--lang=en",
		"--pass="+adminPassword,
		siteName,
		adminUsername)
	out, runErr := podmanmanager.Command(ctx, userContext, installArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "MediaWiki CLI installer failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Deploying admin login helper"})
	loginFilePath := filepath.Join(hostOSPath, openpanelLoginFileName)
	if writeErr := os.WriteFile(loginFilePath, []byte(openpanelLoginPHP), 0o644); writeErr == nil {
		_ = os.Chown(loginFilePath, uid, uid)
	}

	emit(map[string]any{"status": "Registering cron job (maintenance/runJobs.php, every minute)"})
	cronComment := mediawikiCronComment(selectedDomain)
	cronCommand := "php " + installPath + "/maintenance/runJobs.php --maxjobs=50"
	if cronErr := crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true); cronErr != nil {
		emit(map[string]any{"status": "Warning: MediaWiki installed, but the cron job could not be registered automatically: " + cronErr.Error() + " - add it manually from Cron Jobs: schedule '0 * * * * *', container '" + phpContainer + "', command '" + cronCommand + "'."})
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "mediawiki"); insertErr != nil {
		emit(map[string]any{"error": "MediaWiki installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed MediaWiki on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "MediaWiki installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "MediaWiki installation completed!"})
}

// emitCleanupFiles clears the failed install's files but leaves installPath itself - it's the domain's docroot, not ours to delete. Host-side removal, not `podman exec ... rm`, since once the archive is extracted host-side (owned by this process's own real UID) a rootless container's "root" can't write into a directory it doesn't own outside its user-namespace mapping, so `podman exec rm -rf` would silently fail
func emitCleanupFiles(_ context.Context, userContext, _, installPath string, emit func(map[string]any)) {
	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(installPath, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)
	if err := installmanifest.ClearContents(hostOSPath); err != nil {
		emit(map[string]any{"status": "Cleanup: failed to remove files from " + installPath + ": " + err.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: removed files from " + installPath})
}
