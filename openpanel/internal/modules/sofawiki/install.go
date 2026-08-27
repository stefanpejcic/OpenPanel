package sofawiki

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
)

// sofawikiSourceZip is the only source SofaWiki ships: a plain branch
// archive, no tagged releases (confirmed against
// github.com/bellenuit/sofawiki - "There aren't any releases here") and no
// composer.json (not a Composer package).
const sofawikiSourceZip = "https://github.com/bellenuit/sofawiki/archive/refs/heads/master.zip"

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

// phpVersionAbove reports whether version (e.g. "7.2", "8.5") is newer
// than maxMajor.maxMinor. An unparseable version is treated as too new,
// so an unexpected format fails safe (blocks the install with a clear
// message) rather than silently proceeding.
func phpVersionAbove(version string, maxMajor, maxMinor int) bool {
	major, minor := 0, 0
	if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
		return true
	}
	if major != maxMajor {
		return major > maxMajor
	}
	return minor > maxMinor
}

// handleInstallStream drives a SofaWiki install end to end, streaming
// NDJSON progress events to the client: download+extract the master
// branch archive into the docroot, fix ownership, record the site. There
// is no database and no CLI installer to drive - see sofawiki.go's
// package doc comment for why the install wizard itself is left for the
// site owner to complete in their browser.
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

	// SofaWiki's own async self-cron (inc/async.php, called unconditionally
	// on every request) does fsockopen($url['host'], ...) where 'host' is
	// undefined because $swBaseHrefFolder is empty here (this Apache+
	// PHP-FPM setup, via mod_proxy_fcgi, never sets the legacy SCRIPT_URI
	// var SofaWiki relies on to populate it). fsockopen() then returns
	// false, and fwrite(false, ...) is only a warning on PHP <=7.4 but a
	// fatal TypeError on PHP 8+ (stricter internal-function typing) -
	// confirmed via a live install + real browser-path request against
	// this exact deployment (stack trace: inc/async.php:28 fwrite()).
	if !isLitespeed && phpVersionAbove(phpVersion, 7, 4) {
		emit(map[string]any{"error": "SofaWiki requires PHP 7.4 or older on this server (it fatal-errors on PHP 8+ due to a self-check in inc/async.php), but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version (or install into a subdirectory using PHP 7.4 or older) and try again."})
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

	emit(map[string]any{"status": "Downloading and extracting SofaWiki"})
	out, runErr := runSofawikiExtract(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
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
		selectedDomain, domainID, adminEmail, "master", "sofawiki"); insertErr != nil {
		emit(map[string]any{"error": "SofaWiki installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed SofaWiki on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "SofaWiki installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "SofaWiki installation completed! Visit the site to finish setup (folder rights, then the SofaWiki setup wizard)."})
}

// runSofawikiExtract downloads the master branch archive and extracts it
// into installPath. GitHub's archive wraps everything in a top-level
// "sofawiki-master/" directory, so this extracts to a scratch location
// first and moves that directory into place - installPath itself must not
// already exist (checked by the caller) since `mv` won't merge into it.
func runSofawikiExtract(ctx context.Context, userContext, phpContainer, installPath string) ([]byte, error) {
	scratch := "/tmp/openpanel-sofawiki-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	script := `set -e
rm -rf ` + scratch + ` ` + scratch + `.zip
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.zip ` + sofawikiSourceZip + `
unzip -q ` + scratch + `.zip -d ` + scratch + `
mv ` + scratch + `/sofawiki-master ` + installPath + `
rm -rf ` + scratch + ` ` + scratch + `.zip`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
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
