package dokuwiki

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// handleDokuwikiUpdate downloads the current stable tarball, compares its
// VERSION against the installed one, and - only if newer - extracts it
// over the existing install while explicitly preserving conf/, data/ and
// lib/plugins/ (installed themes/plugins live in lib/plugins too, so this
// covers both), mirroring DokuWiki's own documented manual-upgrade
// procedure. There is no CLI updater to drive (unlike drush/wp-cli), so
// this whole check-then-update flow runs as one NDJSON-streamed request,
// unlike flarum/update.go's separate always-check-first client-side call -
// version strings here are dated codenames ("2026-07-14b"), not semver,
// but they still compare correctly as plain strings since the format is
// always YYYY-MM-DD[letter].
func handleDokuwikiUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	currentVersion := "unknown"
	if out, verErr := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", docroot+"/VERSION")).Output(); verErr == nil {
		currentVersion = strings.TrimSpace(string(out))
	}

	emit(map[string]any{"status": "Downloading latest DokuWiki release"})
	scratch := "/tmp/openpanel-dokuwiki-update-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	downloadScript := `set -e
rm -rf ` + scratch + ` ` + scratch + `.tgz
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.tgz ` + dokuwikiStableTarball + `
tar -xzf ` + scratch + `.tgz -C ` + scratch + `
SRC=$(find ` + scratch + ` -mindepth 1 -maxdepth 1 -type d | head -1)
echo "$SRC"
cat "$SRC"/VERSION`
	out, runErr := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", downloadScript)).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Download failed: " + strings.TrimSpace(string(out))})
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		emit(map[string]any{"error": "Unexpected download output."})
		return
	}
	scratchExtractDir := strings.TrimSpace(lines[0])
	latestVersion := strings.TrimSpace(lines[1])

	if latestVersion != "" && currentVersion != "unknown" && latestVersion <= currentVersion {
		emit(map[string]any{"status": "Already running the latest version (" + currentVersion + ")", "version": currentVersion})
		cleanupArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", scratch, scratch+".tgz")
		_ = podmanmanager.Command(ctx, userContext, cleanupArgv).Run()
		return
	}

	emit(map[string]any{"status": "Applying update: " + currentVersion + " -> " + latestVersion})
	// Copy every top-level entry from the fresh extract over the existing
	// install EXCEPT conf/, data/ and lib/plugins/ - DokuWiki's own
	// documented manual-upgrade procedure (there is no CLI updater), so a
	// naive full overwrite would wipe the site's configuration, pages and
	// installed plugins/themes.
	applyScript := `set -e
cd "` + scratchExtractDir + `"
for entry in *; do
  if [ "$entry" = "conf" ] || [ "$entry" = "data" ]; then continue; fi
  if [ "$entry" = "lib" ]; then
    mkdir -p ` + docroot + `/lib
    for libentry in lib/*; do
      base=$(basename "$libentry")
      if [ "$base" = "plugins" ]; then continue; fi
      rm -rf ` + docroot + `/lib/"$base"
      cp -a "$libentry" ` + docroot + `/lib/
    done
    continue
  fi
  rm -rf ` + docroot + `/"$entry"
  cp -a "$entry" ` + docroot + `/
done
rm -f ` + docroot + `/install.php
rm -rf ` + scratch + ` ` + scratch + `.tgz`
	out, runErr = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", applyScript)).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Update apply failed: " + strings.TrimSpace(string(out))})
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, docroot)).Run()
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'dokuwiki'", latestVersion, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated DokuWiki website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": latestVersion})
}
