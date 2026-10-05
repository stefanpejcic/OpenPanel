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
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// sofawikiSourceZip is the only source SofaWiki ships: a plain branch archive, no tagged releases and no composer.json
const sofawikiSourceZip = "https://github.com/bellenuit/sofawiki/archive/refs/heads/master.zip"

// phpVersionAbove reports whether version is newer than maxMajor.maxMinor - an unparseable version is treated as too new, so an unexpected format fails safe rather than silently proceeding
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

// handleInstallStream drives a SofaWiki install end to end over NDJSON: download+extract the master branch archive into the docroot, fix ownership, record the site - no database and no CLI installer, see sofawiki.go's package doc comment for why the setup wizard is left to the site owner
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	in, ok := cmsapp.StartInstall(a, w, r)
	if !ok {
		return
	}
	defer appkit.RemoveLockFile(in.Username)
	ctx := in.Ctx
	currentUsername := in.Username
	userContext := in.UserContext
	emit := in.Emit
	ipAddress := in.IP
	domainID := in.DomainID
	dom := in.Domain
	subdirectory := in.Subdirectory

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

	// inc/async.php's fsockopen self-check fails here since $swBaseHrefFolder never gets populated on this Apache+PHP-FPM/mod_proxy_fcgi setup - fwrite(false, ...) is just a warning on PHP <=7.4 but a fatal TypeError on PHP 8+
	if !isLitespeed && phpVersionAbove(phpVersion, 7, 4) {
		emit(map[string]any{"error": "SofaWiki requires PHP 7.4 or older on this server (it fatal-errors on PHP 8+ due to a self-check in inc/async.php), but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version (or install into a subdirectory using PHP 7.4 or older) and try again."})
		return
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

	emit(map[string]any{"status": "Downloading and extracting SofaWiki"})
	out, runErr := runSofawikiExtract(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
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
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, "master", "sofawiki"); insertErr != nil {
		emit(map[string]any{"error": "SofaWiki installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed SofaWiki on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "SofaWiki installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "SofaWiki installation completed! Visit the site to finish setup (folder rights, then the SofaWiki setup wizard)."})
}

// runSofawikiExtract downloads the master branch archive and extracts it into installPath - GitHub wraps everything in "sofawiki-master/", so this extracts to a scratch location first then cp -a's the contents into installPath, which works whether installPath already exists (root install) or not (subdirectory install)
// index.php also hardcodes ini_set('display_errors', 1), so legacy warnings like inc/async.php's fsockopen self-check get printed into the response body and the WAF blocks it as PHP error disclosure - patched to 0 after extraction
func runSofawikiExtract(ctx context.Context, userContext, phpContainer, installPath string) ([]byte, error) {
	scratch := "/tmp/openpanel-sofawiki-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	script := `set -e
rm -rf ` + scratch + ` ` + scratch + `.zip
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.zip ` + sofawikiSourceZip + `
unzip -q ` + scratch + `.zip -d ` + scratch + `
mkdir -p ` + installPath + `
cp -a ` + scratch + `/sofawiki-master/. ` + installPath + `/
rm -rf ` + scratch + ` ` + scratch + `.zip
sed -i "s/\(ini_set([\"']display_errors[\"'], \)1/\10/g" ` + installPath + `/index.php`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}
