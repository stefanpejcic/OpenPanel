package matomo

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

// unpackMatomoUpdateArchive is unpackMatomoArchive's update-time sibling:
// replaces every top-level entry from the new release except config/
// (config.ini.php holds live DB credentials), matching how
// nextcloud/update.go's unpackNextcloudUpdateArchive preserves config/data.
func unpackMatomoUpdateArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".update-tmp"
	script := `set -e
rm -rf "$2"
mkdir -p "$2"
unzip -q "$1" -d "$2"
cd "$2/matomo"
for f in .[!.]* ..?* *; do
  [ -e "$f" ] || continue
  case "$f" in
    config) continue ;;
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

// handleMatomoUpdate updates an existing Matomo install in place: downloads
// the latest release, replaces core files (preserving config/), then runs
// `console core:update -n` to bring the database schema up to date -
// Matomo's own documented manual-update procedure (unlike a fresh install,
// which has no CLI equivalent - see install.go's comment - an *existing*
// install can be updated via console since core:update doesn't need the
// interactive wizard). Streams NDJSON progress like install does.
func handleMatomoUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://github.com/matomo-org/matomo/releases/download/" + version + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
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
		_ = exec.Command("chown", "-R", strconv.Itoa(uid)+":"+strconv.Itoa(uid), hostOSPath).Run()
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
