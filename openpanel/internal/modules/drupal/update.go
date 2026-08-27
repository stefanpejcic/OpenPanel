package drupal

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// handleDrupalUpdate updates an existing Drupal install in place: composer
// update of drupal/core-recommended (and its transient deps), then drush's
// own database-schema-update and cache-rebuild steps, matching Drupal's
// documented Composer-based update procedure. Streams NDJSON progress like
// install does. No automatic backup - the UI tells the user to take one
// from the Backups tab first, since a DB dump strategy that's safe to run
// unattended here would need the same per-CMS table-discovery logic
// backups.go already has, and running it silently before every update
// hides how large/slow that step can be from the user.
func handleDrupalUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	emit(map[string]any{"status": "Running composer update (drupal/core-recommended)"})
	composerArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"),
		"--working-dir="+docroot, "update", "drupal/core-recommended", "drupal/core-composer-scaffold", "drupal/core-project-message",
		"--with-all-dependencies", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, composerArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer update failed: " + strings.TrimSpace(string(out))})
		return
	}

	emit(map[string]any{"status": "Running database updates (drush updatedb)"})
	drushBin := docroot + "/vendor/bin/drush"
	updatedbArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, drushBin, "updatedb", "--root="+docroot, "-y")
	out, runErr = podmanmanager.Command(ctx, userContext, updatedbArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "drush updatedb failed: " + strings.TrimSpace(string(out))})
		return
	}

	emit(map[string]any{"status": "Rebuilding cache"})
	cacheArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, drushBin, "cache:rebuild", "--root="+docroot)
	_, _ = podmanmanager.Command(ctx, userContext, cacheArgv).CombinedOutput()

	emit(map[string]any{"status": "Checking installed version"})
	statusArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, drushBin, "status", "--root="+docroot, "--field=drupal-version")
	versionOut, _ := podmanmanager.Command(ctx, userContext, statusArgv).Output()
	newVersion := strings.TrimSpace(string(versionOut))

	if newVersion != "" {
		if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'drupal'", newVersion, selectedDomain); execErr != nil {
			emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
			return
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated Drupal website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": newVersion})
}
