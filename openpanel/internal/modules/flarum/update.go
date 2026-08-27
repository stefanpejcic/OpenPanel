package flarum

import (
	"encoding/json"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// parseFlarumCoreVersion reads flarum/core's resolved version out of
// composer.lock content already fetched via podman exec cat, mirroring
// websites.go's getFlarumVersion (which reads the same file from the host
// bind mount) without needing this package to import that one.
func parseFlarumCoreVersion(lockContent []byte) string {
	var lock struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
	}
	if json.Unmarshal(lockContent, &lock) != nil {
		return "Unknown"
	}
	for _, pkg := range lock.Packages {
		if pkg.Name == "flarum/core" {
			return pkg.Version
		}
	}
	return "Unknown"
}

// handleFlarumUpdate updates an existing Flarum install in place: composer
// update of flarum/core (and its transient deps), then `php flarum migrate`
// and `php flarum cache:clear`, matching Flarum's own documented Composer
// update procedure. Streams NDJSON progress like install does. Reuses the
// same absolute-flarum-script-path workaround install.go needs (the PHP
// wrapper doesn't reliably resolve relative script paths).
func handleFlarumUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	currentVersion := "Unknown"
	if lockOut, lockErr := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", docroot+"/composer.lock")).Output(); lockErr == nil {
		currentVersion = parseFlarumCoreVersion(lockOut)
	}

	// Only flarum/core 2.x requires PHP 8.1+ (see install.go's identical
	// guard) - an install already sitting on the 1.x line (the common
	// case, since "latest" resolves to 1.x until a stable 2.0.0 ships)
	// must still be able to take routine 1.x patch updates on older PHP.
	if !isLitespeed && strings.HasPrefix(currentVersion, "2.") && phpVersionBelow(phpVersion, 8, 1) {
		emit(map[string]any{"error": "Flarum 2.x requires PHP 8.1 or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version and try again."})
		return
	}

	// Pin to the latest *stable* numeric release explicitly rather than
	// an unconstrained "flarum/core" bump - composer.json's own
	// minimum-stability:beta (inherited from flarum/flarum's own
	// recommended-project template) would otherwise let this silently
	// jump to a 2.0.0 pre-release the moment one ships, exactly like
	// install.go's ^2.0.0 + --stability=beta combo used to force one
	// even though none is stable yet (confirmed live against the real
	// tags feed: v2.0.0-rc.7 is still the newest 2.x tag).
	updateTarget := "flarum/core"
	if latestVersion, verErr := latestFlarumVersion(ctx); verErr == nil {
		updateTarget = "flarum/core:^" + latestVersion
	}

	emit(map[string]any{"status": "Running composer require (" + updateTarget + ")"})
	// flarum/core is a transitive dependency pulled in by flarum/flarum,
	// not a direct "require" entry in composer.json - "composer update
	// flarum/core" refuses to touch a package composer.json doesn't name
	// directly ("Run composer require flarum/core instead"), confirmed
	// live against a real install. "composer require" both adds/bumps the
	// constraint and installs it in one step, which is what's needed here.
	composerArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"),
		"--working-dir="+docroot, "require", updateTarget, "--with-all-dependencies", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, composerArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer require failed: " + strings.TrimSpace(string(out))})
		return
	}

	flarumScript := docroot + "/flarum"

	emit(map[string]any{"status": "Running database migrations (flarum migrate)"})
	migrateArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", flarumScript, "migrate")
	out, runErr = podmanmanager.Command(ctx, userContext, migrateArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "flarum migrate failed: " + strings.TrimSpace(string(out))})
		return
	}

	emit(map[string]any{"status": "Clearing cache"})
	cacheArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", flarumScript, "cache:clear")
	_, _ = podmanmanager.Command(ctx, userContext, cacheArgv).CombinedOutput()

	emit(map[string]any{"status": "Checking installed version"})
	newVersion := "Unknown"
	lockOut, lockErr := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", docroot+"/composer.lock")).Output()
	if lockErr == nil {
		newVersion = parseFlarumCoreVersion(lockOut)
	}

	if newVersion != "" && newVersion != "Unknown" {
		if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'flarum'", newVersion, selectedDomain); execErr != nil {
			emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
			return
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated Flarum website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": newVersion})
}
