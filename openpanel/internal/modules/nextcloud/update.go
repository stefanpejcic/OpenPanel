package nextcloud

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// unpackNextcloudUpdateArchive is unpackNextcloudArchive's update-time
// sibling: it replaces every top-level entry from the new release except
// config/ and data/, which must survive an update untouched (config.php
// holds live DB credentials/trusted domains, data/ holds user files) -
// install.go's version has no such exclusion since a fresh install has
// neither yet. Deleting the old copy of each entry before moving the new
// one in (rather than just overwriting) mirrors Nextcloud's own documented
// manual-update rsync --delete behavior, so files removed from newer
// releases don't linger.
func unpackNextcloudUpdateArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".update-tmp"
	script := `set -e
rm -rf "$2"
mkdir -p "$2"
unzip -q "$1" "nextcloud/*" -d "$2"
cd "$2/nextcloud"
for f in .[!.]* ..?* *; do
  [ -e "$f" ] || continue
  case "$f" in
    config|data) continue ;;
  esac
  rm -rf "$3/$f"
  mv "$f" "$3/"
done
rm -rf "$2"
`
	cmd := exec.CommandContext(ctx, "sh", "-c", script, "sh", archivePath, tmpDir, destDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return &execError{msg: strings.TrimSpace(string(out)), err: err}
	}
	return nil
}

// handleNextcloudUpdate updates an existing Nextcloud install in place:
// downloads the latest release archive, replaces core files (preserving
// config/ and data/), then runs occ upgrade wrapped in maintenance mode -
// Nextcloud's own documented manual-update procedure. Streams NDJSON
// progress like install does.
func handleNextcloudUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

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
	if err := createLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}
	defer removeLockFile(currentUsername)

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
	if !ensureContainerRunning(ctx, userContext, phpContainer) {
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
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://download.nextcloud.com/server/releases/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
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
		_ = exec.Command("chown", "-R", strconv.Itoa(uid)+":"+strconv.Itoa(uid), hostOSPath).Run()
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
