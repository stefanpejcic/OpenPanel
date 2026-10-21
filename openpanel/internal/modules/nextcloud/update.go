package nextcloud

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// unpackNextcloudUpdateArchive is unpackNextcloudArchive's update-time sibling: replaces every top-level entry except config/ and data/, deleting the old copy first to mirror Nextcloud's documented rsync --delete update behavior
func unpackNextcloudUpdateArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".update-tmp"
	defer tenantfs.RemoveAll(tmpDir)
	_ = tenantfs.RemoveAll(tmpDir)
	if err := tenantfs.ExtractZipSubdir(ctx, archivePath, "nextcloud", tmpDir); err != nil {
		return err
	}
	return tenantfs.ReplaceContents(tmpDir, destDir, "config", "data")
}

// handleNextcloudUpdate updates an install in place: downloads the latest release, replaces core files, then runs occ upgrade wrapped in maintenance mode - streams NDJSON progress like install does
func handleNextcloudUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { web.WriteNDJSON(w, flusher, canFlush, v) }

	selectedDomain := r.URL.Query().Get("domain")
	docroot := r.URL.Query().Get("docroot")
	if selectedDomain == "" || docroot == "" {
		emit(map[string]any{"error": "Missing required domain/docroot."})
		return
	}
	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		emit(map[string]any{"error": "You do not own this domain."})
		return
	}

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	if err := appkit.CreateLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}
	defer appkit.RemoveLockFile(currentUsername)

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
	if !cmsapp.EnsureContainerRunning(ctx, userContext, phpContainer) {
		emit(map[string]any{"error": "PHP container failed to start. Please check it from Services."})
		return
	}

	version, latestErr := latestNextcloudVersion(ctx)
	if latestErr != nil {
		emit(map[string]any{"error": "Could not determine latest Nextcloud version: " + latestErr.Error()})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(docroot, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	archiveName := "nextcloud-" + version + ".zip"
	archiveDir := "/etc/openpanel/nextcloud/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := tenantfs.Stat(archivePath); statErr != nil {
		downloadURL := "https://download.nextcloud.com/server/releases/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = tenantfs.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading Nextcloud " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Enabling maintenance mode"})
	maintOnArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/occ", "maintenance:mode", "--on")
	_, _ = podmanmanager.Command(ctx, userContext, maintOnArgv).CombinedOutput()

	emit(map[string]any{"status": "Replacing core files (preserving config and data)"})
	if unpackErr := unpackNextcloudUpdateArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting Nextcloud archive: " + unpackErr.Error()})
		maintOffArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/occ", "maintenance:mode", "--off")
		_, _ = podmanmanager.Command(ctx, userContext, maintOffArgv).CombinedOutput()
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = tenantfs.ChownR(hostOSPath, uid, uid)
	}

	emit(map[string]any{"status": "Running occ upgrade"})
	upgradeArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/occ", "upgrade")
	out, runErr := podmanmanager.Command(ctx, userContext, upgradeArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "occ upgrade failed: " + strings.TrimSpace(string(out))})
		maintOffArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/occ", "maintenance:mode", "--off")
		_, _ = podmanmanager.Command(ctx, userContext, maintOffArgv).CombinedOutput()
		return
	}

	emit(map[string]any{"status": "Disabling maintenance mode"})
	maintOffArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/occ", "maintenance:mode", "--off")
	_, _ = podmanmanager.Command(ctx, userContext, maintOffArgv).CombinedOutput()

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'nextcloud'", version, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated Nextcloud website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": version})
}
