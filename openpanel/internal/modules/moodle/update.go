package moodle

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

// unpackMoodleUpdateArchive is unpackMoodleArchive's update-time sibling:
// extracts the new packaged tarball to a scratch dir, then replaces every
// top-level entry in approotHostPath except config.php - the tarball only
// ships config-dist.php (a template), never a real config.php, so this
// would never actually clash on a fresh install, but the update path must
// not overwrite the live one install.go itself wrote to approot's root
// (see unpackMoodleArchive's own comment on that layout).
func unpackMoodleUpdateArchive(ctx context.Context, archivePath, approotHostPath string) error {
	tmpDir := approotHostPath + ".update-tmp"
	defer os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "tar", "-xzf", archivePath, "-C", tmpDir, "--strip-components=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return &execError{msg: strings.TrimSpace(string(out)), err: err}
	}
	script := `set -e
cd "$1"
for f in .[!.]* ..?* *; do
  [ -e "$f" ] || continue
  [ "$f" = "config.php" ] && continue
  rm -rf "$2/$f"
  mv "$f" "$2/"
done
`
	cmd = exec.CommandContext(ctx, "sh", "-c", script, "sh", tmpDir, approotHostPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return &execError{msg: strings.TrimSpace(string(out)), err: err}
	}
	return nil
}

// handleMoodleUpdate updates an existing Moodle install in place:
// maintenance mode on, download+replace the packaged tarball's code
// (preserving config.php), then run admin/cli/upgrade.php - Moodle's own
// documented manual-update procedure. Streams NDJSON progress like install
// does.
func handleMoodleUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	if selectedDomain == "" {
		emit(map[string]any{"error": "Missing required domain."})
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

	version, latestErr := latestMoodleVersion(ctx)
	if latestErr != nil {
		emit(map[string]any{"error": "Could not determine latest Moodle version: " + latestErr.Error()})
		return
	}
	branch := moodleBranch(version)
	if branch == "" {
		emit(map[string]any{"error": "Could not determine Moodle packaging branch for version " + version})
		return
	}

	slug := siteSlug(selectedDomain)
	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	approotHostPath := filepath.Join(htmlVolume, slug+"_moodleapp")
	approotContainerPath := "/var/www/html/" + slug + "_moodleapp"

	archiveName := "moodle-" + version + ".tgz"
	archiveDir := "/etc/openpanel/moodle/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://download.moodle.org/download.php/direct/stable" + branch + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading Moodle " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Enabling maintenance mode"})
	maintOnArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approotContainerPath+"/admin/cli/maintenance.php", "--enable")
	_, _ = podmanmanager.Command(ctx, userContext, maintOnArgv).CombinedOutput()

	emit(map[string]any{"status": "Replacing core files (preserving config.php)"})
	if unpackErr := unpackMoodleUpdateArchive(ctx, archivePath, approotHostPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting Moodle archive: " + unpackErr.Error()})
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = exec.Command("chown", "-R", strconv.Itoa(uid)+":"+strconv.Itoa(uid), approotHostPath).Run()
	}

	emit(map[string]any{"status": "Running admin/cli/upgrade.php"})
	upgradeArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approotContainerPath+"/admin/cli/upgrade.php", "--non-interactive")
	out, runErr := podmanmanager.Command(ctx, userContext, upgradeArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "admin/cli/upgrade.php failed: " + strings.TrimSpace(string(out))})
		maintOffArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approotContainerPath+"/admin/cli/maintenance.php", "--disable")
		_, _ = podmanmanager.Command(ctx, userContext, maintOffArgv).CombinedOutput()
		return
	}

	emit(map[string]any{"status": "Disabling maintenance mode"})
	maintOffArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approotContainerPath+"/admin/cli/maintenance.php", "--disable")
	_, _ = podmanmanager.Command(ctx, userContext, maintOffArgv).CombinedOutput()

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'moodle'", version, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated Moodle website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": version})
}
