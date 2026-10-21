package matomo

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

// unpackMatomoUpdateArchive is unpackMatomoArchive's update-time sibling: replaces every top-level entry from the new release except config/ (config.ini.php holds live DB credentials), matching how nextcloud/update.go's unpackNextcloudUpdateArchive preserves config/data
func unpackMatomoUpdateArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".update-tmp"
	defer tenantfs.RemoveAll(tmpDir)
	_ = tenantfs.RemoveAll(tmpDir)
	if err := tenantfs.ExtractZipSubdir(ctx, archivePath, "matomo", tmpDir); err != nil {
		return err
	}
	return tenantfs.ReplaceContents(tmpDir, destDir, "config")
}

// handleMatomoUpdate updates an existing Matomo install in place: downloads the latest release, replaces core files (preserving config/), then runs `console core:update -n` to bring the database schema up to date - Matomo's documented manual-update procedure (unlike a fresh install which has no CLI equivalent, an existing install can be updated via console since core:update doesn't need the interactive wizard), streaming NDJSON progress like install does
func handleMatomoUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	version, latestErr := latestMatomoVersion(ctx)
	if latestErr != nil {
		emit(map[string]any{"error": "Could not determine latest Matomo version: " + latestErr.Error()})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(docroot, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	archiveName := "matomo-" + version + ".zip"
	archiveDir := "/etc/openpanel/matomo/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := tenantfs.Stat(archivePath); statErr != nil {
		downloadURL := "https://github.com/matomo-org/matomo/releases/download/" + version + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = tenantfs.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading Matomo " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Replacing core files (preserving config)"})
	if unpackErr := unpackMatomoUpdateArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting Matomo archive: " + unpackErr.Error()})
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = tenantfs.ChownR(hostOSPath, uid, uid)
	}

	emit(map[string]any{"status": "Running database updates (console core:update)"})
	updateArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/console", "core:update", "-n")
	out, runErr := podmanmanager.Command(ctx, userContext, updateArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "console core:update failed: " + strings.TrimSpace(string(out))})
		return
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'matomo'", version, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated Matomo website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": version})
}
