package flarum

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// flarumRequestParams pulls the domain/docroot query params every handler
// in this file needs, splits the main domain out of a possible
// subdirectory suffix, verifies ownership, and resolves the PHP container
// to exec the flarum console inside - mirrors drupal/drush.go's
// drushRequestParams.
func flarumRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
	domain = r.URL.Query().Get("domain")
	docroot = r.URL.Query().Get("docroot")
	if domain == "" || docroot == "" {
		return "", "", "", false
	}

	mainDomain := domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		mainDomain = domain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, mainDomain) {
		return "", "", "", false
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, mainDomain)
	phpContainer = webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + phpVersion
	}
	return domain, docroot, phpContainer, true
}

// handleFlarumCacheClear runs `php flarum cache:clear` - Flarum's console
// does have this command (Flarum\Foundation\Console\CacheClearCommand,
// registered in ConsoleServiceProvider), unlike the install/login commands
// this module can't use - see flarum.go's package doc comment.
func handleFlarumCacheClear(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := flarumRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", "--workdir", docroot, phpContainer, "php", "flarum", "cache:clear")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "php flarum cache:clear failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Flarum cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleFlarumLogs returns the tail of storage/logs/flarum.log as plain
// text. Flarum has no `watchdog:show`-style console command the way
// Drupal does, so this reads its Laravel-style daily log file directly off
// the host filesystem instead of exec-ing into the container - same
// live-read-from-disk approach internal/modules/websites uses for every
// other CMS's version/DB info.
func handleFlarumLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, docroot, _, ok := flarumRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(docroot, "/var/www/html/") {
		http.Error(w, "invalid docroot", http.StatusBadRequest)
		return
	}

	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/")
	logPath := filepath.Join(mappedDir, "storage", "logs", "flarum.log")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	content, readErr := os.ReadFile(logPath)
	if readErr != nil {
		_, _ = w.Write([]byte("No log file found yet at storage/logs/flarum.log."))
		return
	}

	// Tail the last 300 lines - the same order of magnitude as drush
	// watchdog:show's default --count, and this file has no built-in
	// rotation/size cap the way a DB-backed log would.
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	const maxLines = 300
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	_, _ = w.Write([]byte(strings.Join(lines, "\n")))
}
