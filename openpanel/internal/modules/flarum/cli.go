package flarum

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleFlarumCacheClear runs `php flarum cache:clear` - Flarum's console does have this command (Flarum\Foundation\Console\CacheClearCommand, registered in ConsoleServiceProvider), unlike the install/login commands this module can't use, see flarum.go's package doc comment
func handleFlarumCacheClear(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", "--workdir", docroot, phpContainer, "php", "flarum", "cache:clear")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "php flarum cache:clear failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Flarum cache for "+domain, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleFlarumLogs returns the tail of storage/logs/flarum.log as plain text - Flarum has no `watchdog:show`-style console command like Drupal, so this reads its Laravel-style daily log file directly off the host filesystem instead of exec-ing into the container, same live-read-from-disk approach internal/modules/websites uses for every other CMS's version/DB info
func handleFlarumLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, docroot, _, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
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

	// tail the last 300 lines, same order of magnitude as drush watchdog:show's default --count, and this file has no built-in rotation/size cap the way a DB-backed log would
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	const maxLines = 300
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	_, _ = w.Write([]byte(strings.Join(lines, "\n")))
}
