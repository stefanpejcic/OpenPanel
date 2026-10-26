package mediawiki

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

// unpackMediaWikiUpdateArchive extracts the new release tarball to a scratch dir, then replaces every top-level entry in installPath except LocalSettings.php (the live config maintenance/install.php wrote) and images/ (uploaded files) - MediaWiki's own documented manual-update procedure copies those two forward from the old tree into the new one, equivalent to just never touching them in place here
func unpackMediaWikiUpdateArchive(ctx context.Context, archivePath, installPath string) error {
	tmpDir := installPath + ".update-tmp"
	defer tenantfs.RemoveAll(tmpDir)
	if err := tenantfs.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	if err := tenantfs.ExtractTar(ctx, archivePath, tmpDir, 1); err != nil {
		return err
	}
	return tenantfs.ReplaceContents(tmpDir, installPath, "LocalSettings.php", "images")
}

// handleMediaWikiUpdate updates an existing MediaWiki install in place: download+replace the release tarball's code (preserving LocalSettings.php and images/), then run maintenance/update.php --quick - MediaWiki's own documented manual-update procedure, streaming NDJSON progress like install does
func handleMediaWikiUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	// docroot becomes a host path (hostOSPath below) that ReplaceContents and ChownR write to, so a "../.." in it could climb out of the account volume - reject anything that isn't a plain path under /var/www/html
	if !appkit.SafeDocroot(docroot) {
		emit(map[string]any{"error": "invalid docroot"})
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

	version, latestErr := latestMediaWikiVersion(ctx)
	if latestErr != nil {
		emit(map[string]any{"error": "Could not determine latest MediaWiki version: " + latestErr.Error()})
		return
	}
	branch := mediawikiBranchForVersion(version)
	if branch == "" {
		emit(map[string]any{"error": "Could not determine MediaWiki release branch for version " + version})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(docroot, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	archiveName := "mediawiki-" + version + ".tar.gz"
	archiveDir := "/etc/openpanel/mediawiki/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := tenantfs.Stat(archivePath); statErr != nil {
		downloadURL := "https://releases.wikimedia.org/mediawiki/" + branch + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = tenantfs.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading MediaWiki " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Replacing core files (preserving LocalSettings.php and images/)"})
	if unpackErr := unpackMediaWikiUpdateArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting MediaWiki archive: " + unpackErr.Error()})
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = tenantfs.ChownR(hostOSPath, uid, uid)
	}

	emit(map[string]any{"status": "Running maintenance/update.php"})
	updateArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/maintenance/update.php", "--quick")
	out, runErr := podmanmanager.Command(ctx, userContext, updateArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "maintenance/update.php failed: " + strings.TrimSpace(string(out))})
		return
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET version = ? WHERE site_name = ? AND type = 'mediawiki'", version, selectedDomain); execErr != nil {
		emit(map[string]any{"error": "Update completed, but failed to record new version: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated MediaWiki website "+selectedDomain, reqip.ClientIP(r))
	emit(map[string]any{"status": "Update completed!", "version": version})
}
