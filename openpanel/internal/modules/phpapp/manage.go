package phpapp

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// composerLogPath returns the host path a site's captured Composer run
// output is written to and read back from - same base directory
// convention as the install lock file (per-user metadata under
// /etc/openpanel/openpanel/core/users/<username>/).
func composerLogPath(username, siteName string) string {
	safe := strings.ReplaceAll(siteName, "/", "_")
	return "/etc/openpanel/openpanel/core/users/" + username + "/php-app-logs/" + safe + ".log"
}

// appendComposerLog records one Composer run's output, prefixed with a
// timestamp/action header, so the manage page's Logs tab has a running
// history rather than just the last run.
func appendComposerLog(username, siteName, action string, output []byte) {
	path := composerLogPath(username, siteName)
	_ = os.MkdirAll(path[:strings.LastIndex(path, "/")], 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	header := fmt.Sprintf("\n=== composer %s - %s ===\n", action, time.Now().UTC().Format(time.RFC3339))
	_, _ = f.WriteString(header)
	_, _ = f.Write(output)
	if len(output) == 0 || output[len(output)-1] != '\n' {
		_, _ = f.WriteString("\n")
	}
}

// resolvePHPSite looks up a site by name, confirms it belongs to the
// current user, and resolves the php-fpm container currently serving it.
// Returns ("", "", false) with the response already written on any failure.
func resolvePHPSite(a *appctx.App, w http.ResponseWriter, r *http.Request) (userContext, phpContainer, installPath, username string, ok bool) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	username, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return "", "", "", "", false
	}

	siteName := r.PathValue("site_name")
	domain := siteName
	if idx := strings.Index(siteName, "/"); idx != -1 {
		domain = siteName[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return "", "", "", "", false
	}

	prefix := phpAppEnvPrefix(siteName)
	installPathVal, _ := docker.GetEnvValue(userContext, prefix+"WORKDIR")
	if installPathVal == "" {
		http.NotFound(w, r)
		return "", "", "", "", false
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
	phpContainer = webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + phpVersion
	}

	return userContext, phpContainer, installPathVal, username, true
}

// handleComposerAction runs `composer install` or `composer update` for an
// already-installed PHP app, inside the domain's current php-fpm container.
func handleComposerAction(a *appctx.App, w http.ResponseWriter, r *http.Request, action string) {
	ctx := r.Context()
	userContext, phpContainer, installPath, username, ok := resolvePHPSite(a, w, r)
	if !ok {
		return
	}
	siteName := r.PathValue("site_name")

	_ = r.ParseForm()
	args := []string{action, "--no-interaction"}
	if action == "update" && normalizeCheckbox(r.FormValue("optimize_autoloader")) {
		args = append(args, "--optimize-autoloader")
	}

	if !ensureContainerRunning(ctx, userContext, phpContainer) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "PHP container is not running."})
		return
	}

	// See install.go's composerBase comment: --working-dir, not `exec -w`.
	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"), "--working-dir="+installPath)
	argv = append(argv, args...)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	appendComposerLog(username, siteName, action, out)

	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "composer " + action + " failed: " + strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "ran composer "+action+" for "+siteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "composer " + action + " completed successfully.", "output": string(out)})
}

// handleComposerLogs returns the captured Composer run history for a site.
func handleComposerLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, _, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userID, _ := auth.UserID(r)
	siteName := r.PathValue("site_name")
	domain := siteName
	if idx := strings.Index(siteName, "/"); idx != -1 {
		domain = siteName[:idx]
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(composerLogPath(username, siteName))
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("No Composer runs recorded yet."))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}

// handleDelete removes a PHP app's `sites` row and its .env settings.
// Docroot files and the database, if any, are left untouched - same as the
// NodeJS/Python delete flow's "All website data such as files and database
// remains" behavior - since there's no dedicated container to tear down.
func handleDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	username, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	siteName := r.PathValue("site_name")
	domain := siteName
	if idx := strings.Index(siteName, "/"); idx != -1 {
		domain = siteName[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if _, execErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE site_name = ? AND type = 'PHP'", siteName); execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	prefix := phpAppEnvPrefix(siteName)
	env := docker.LoadEnvFile(userContext)
	for k := range env {
		if strings.HasPrefix(k, prefix) {
			delete(env, k)
		}
	}
	_ = docker.SaveEnvFile(userContext, env)

	_ = logger.RecordUserAction(a.Config, username, "deleted PHP application "+siteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "PHP application removed."})
}
