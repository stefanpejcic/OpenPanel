package ojs

import (
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

// handleOJSUpdate updates an existing OJS install by extracting the new
// version's tarball into a fresh sibling directory (never touching the
// live one in place), copying the old install's config.inc.php across
// unmodified (files_dir/database credentials/base_url all stay correct
// since none of those are version-specific), then atomically repointing the
// docroot symlink at the new tree and running OJS's own
// `tools/upgrade.php upgrade` (confirmed genuinely non-interactive/
// flag-based, unlike tools/install.php - see
// lib/pkp/classes/cliTool/UpgradeTool.php). On failure the symlink is
// pointed back at the untouched old tree, so a bad upgrade never leaves the
// site down.
func handleOJSUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	requestedVersion := strings.TrimSpace(r.URL.Query().Get("ojs_version"))
	var dotted string
	if requestedVersion == "" {
		latest, latestErr := latestOJSVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest OJS version: " + latestErr.Error()})
			return
		}
		dotted = latest.Dotted
	} else {
		resolved, resolveErr := findOJSVersion(ctx, requestedVersion)
		if resolveErr != nil {
			emit(map[string]any{"error": resolveErr.Error()})
			return
		}
		dotted = resolved.Dotted
	}

	slug := siteSlug(selectedDomain)
	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	approotHostPath := filepath.Join(htmlVolume, slug+"_ojsapp")
	newApprootHostPath := approotHostPath + ".new"
	newApprootContainerPath := "/var/www/html/" + slug + "_ojsapp.new"

	archiveName := "ojs-" + dotted + ".tar.gz"
	archiveDir := "/etc/openpanel/ojs/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := ojsDownloadURL(dotted)
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading OJS " + dotted + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting new version to a fresh directory"})
	_ = os.RemoveAll(newApprootHostPath)
	if unpackErr := unpackOJSArchive(ctx, archivePath, newApprootHostPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting OJS archive: " + unpackErr.Error()})
		_ = os.RemoveAll(newApprootHostPath)
		return
	}

	emit(map[string]any{"status": "Copying existing config.inc.php into the new version"})
	oldConfigPath := filepath.Join(approotHostPath, "config.inc.php")
	newConfigPath := filepath.Join(newApprootHostPath, "config.inc.php")
	configContent, readErr := os.ReadFile(oldConfigPath)
	if readErr != nil {
		emit(map[string]any{"error": "Could not read existing config.inc.php: " + readErr.Error()})
		_ = os.RemoveAll(newApprootHostPath)
		return
	}
	if writeErr := os.WriteFile(newConfigPath, configContent, 0o644); writeErr != nil {
		emit(map[string]any{"error": "Could not write config.inc.php into new version: " + writeErr.Error()})
		_ = os.RemoveAll(newApprootHostPath)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = exec.Command("chown", "-R", strconv.Itoa(uid)+":"+strconv.Itoa(uid), newApprootHostPath).Run()
	}

	emit(map[string]any{"status": "Switching web root to the new version"})
	docrootSymlinkHostPath, symlinkErr := resolveOJSDocrootSymlink(userContext, selectedDomain)
	if symlinkErr != nil {
		emit(map[string]any{"error": symlinkErr.Error()})
		_ = os.RemoveAll(newApprootHostPath)
		return
	}
	if rmErr := os.Remove(docrootSymlinkHostPath); rmErr != nil {
		emit(map[string]any{"error": "Could not remove existing web root symlink: " + rmErr.Error()})
		_ = os.RemoveAll(newApprootHostPath)
		return
	}
	if symErr := os.Symlink(newApprootContainerPath, docrootSymlinkHostPath); symErr != nil {
		emit(map[string]any{"error": "Could not point web root at the new version: " + symErr.Error()})
		// Best-effort roll back the symlink to the old (untouched) tree.
		_ = os.Symlink("/var/www/html/"+slug+"_ojsapp", docrootSymlinkHostPath)
		_ = os.RemoveAll(newApprootHostPath)
		return
	}

	emit(map[string]any{"status": "Running tools/upgrade.php upgrade"})
	upgradeArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", newApprootContainerPath+"/tools/upgrade.php", "upgrade")
	out, runErr := podmanmanager.Command(ctx, userContext, upgradeArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "tools/upgrade.php upgrade failed: " + strings.TrimSpace(string(out))})
		emit(map[string]any{"status": "Rolling back web root to the previous version"})
		_ = os.Remove(docrootSymlinkHostPath)
		_ = os.Symlink("/var/www/html/"+slug+"_ojsapp", docrootSymlinkHostPath)
		_ = os.RemoveAll(newApprootHostPath)
		return
	}

	emit(map[string]any{"status": "Finalizing: replacing previous version"})
	_ = os.RemoveAll(approotHostPath)
	if renameErr := os.Rename(newApprootHostPath, approotHostPath); renameErr != nil {
		// The upgrade already ran successfully against newApprootHostPath and
		// the symlink already points at "<slug>_ojsapp.new" on disk - this
		// rename failing just means the directory keeps its ".new" name; not
		// fatal, but repoint the symlink at the actual path it ended up at.
		emit(map[string]any{"status": "Warning: could not rename new version into place: " + renameErr.Error()})
		_ = os.Remove(docrootSymlinkHostPath)
		_ = os.Symlink(newApprootContainerPath, docrootSymlinkHostPath)
	} else {
		_ = os.Remove(docrootSymlinkHostPath)
		_ = os.Symlink("/var/www/html/"+slug+"_ojsapp", docrootSymlinkHostPath)
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'ojs'", dotted, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated OJS website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": dotted})
}

// resolveOJSDocrootSymlink returns the host-side path of the domain's
// docroot symlink (the thing install.go created pointing at
// "<slug>_ojsapp") by looking it up from the domains table, the same way
// install.go originally derived it, rather than assuming the caller already
// has it (update.go's own route only carries ?domain=, not ?docroot=).
func resolveOJSDocrootSymlink(userContext, selectedDomain string) (string, error) {
	slug := siteSlug(selectedDomain)
	approotContainerPath := "/var/www/html/" + slug + "_ojsapp"
	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"

	// docroot on disk is literally "/var/www/html/<selectedDomain>" mapped
	// onto htmlVolume - the exact relation install.go's own hostOSPath
	// construction relies on.
	hostOSPath := filepath.Join(htmlVolume, selectedDomain)
	if info, statErr := os.Lstat(hostOSPath); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return hostOSPath, nil
	}
	return "", &execError{msg: "Could not locate the OJS web root symlink for " + selectedDomain + " (expected it to point at " + approotContainerPath + ")"}
}
