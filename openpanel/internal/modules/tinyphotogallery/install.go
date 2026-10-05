package tinyphotogallery

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// tinyPhotoGallerySourceFile is the only source TinyPhotoGallery ships: a single PHP file on the main branch, no tagged releases and no composer.json
const tinyPhotoGallerySourceFile = "https://raw.githubusercontent.com/stefanpejcic/tinyphotogallery/main/index.php"

// tinyPhotoGalleryVersion is a static placeholder recorded in the sites table - no real versioning upstream, same convention sofawiki uses for its own "master" branch install
const tinyPhotoGalleryVersion = "main"

// handleInstallStream drives a TinyPhotoGallery install end to end over NDJSON: download index.php, create an empty photos/ folder next to it, fix ownership, record the site - no database, no admin account, no CLI installer; install is complete the moment the two filesystem items exist
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

	emit(map[string]any{"status": "Downloading TinyPhotoGallery"})
	out, runErr := runTinyPhotoGalleryInstall(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download failed: " + strings.TrimSpace(string(out))})
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
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, "admin@"+dom.DomainURL, tinyPhotoGalleryVersion, "tinyphotogallery"); insertErr != nil {
		emit(map[string]any{"error": "TinyPhotoGallery installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed TinyPhotoGallery on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "TinyPhotoGallery installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "TinyPhotoGallery installation completed! Visit the site to start uploading photos."})
}

// runTinyPhotoGalleryInstall downloads index.php from the main branch into installPath and creates an empty photos/ folder next to it - that's the entire upstream install procedure
func runTinyPhotoGalleryInstall(ctx context.Context, userContext, phpContainer, installPath string) ([]byte, error) {
	script := `set -e
mkdir -p ` + installPath + `
curl -sL -o ` + installPath + `/index.php ` + tinyPhotoGallerySourceFile + `
mkdir -p ` + installPath + `/photos`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}
