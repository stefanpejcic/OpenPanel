package cmsapp

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

func IsValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// EnsureContainerRunning starts the container if it isn't already running, polling briefly for it to come up
func EnsureContainerRunning(ctx context.Context, userContext, container string) bool {
	if docker.IsServiceRunning(ctx, userContext, container) {
		return true
	}
	docker.StartOrStopContainer(ctx, userContext, container, "activate", "detached")
	const attempts = 15
	for i := 0; i < attempts; i++ {
		time.Sleep(2 * time.Second)
		if docker.IsServiceRunning(ctx, userContext, container) {
			return true
		}
	}
	return false
}

// EmitCleanupDatabase drops a failed install's database and user, reporting on the NDJSON stream
func EmitCleanupDatabase(ctx context.Context, userContext, dbName, dbUser, dbHost string, emit func(map[string]any)) {
	_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		emit(map[string]any{"error": "Cleanup: failed to drop database/user: " + execErr.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: dropped database `" + dbName + "` and user `" + dbUser + "`"})
}

func EscapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// CompareVersions compares two dotted numeric versions ("34.0.3" vs "9.2.1"), returns >0 if a > b
func CompareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	for i := 0; i < len(partsA) || i < len(partsB); i++ {
		var na, nb int
		if i < len(partsA) {
			na, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			nb, _ = strconv.Atoi(partsB[i])
		}
		if na != nb {
			return na - nb
		}
	}
	return 0
}

// RequestParams reads the domain/docroot query params a site action needs, checks ownership and resolves the PHP container to run it in
func RequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
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

// EmitCleanupFiles clears a failed install's files but leaves installPath itself, it may be the domain's docroot
func EmitCleanupFiles(ctx context.Context, userContext, phpContainer, installPath string, emit func(map[string]any)) {
	if err := installmanifest.ClearContentsViaContainer(ctx, userContext, phpContainer, installPath); err != nil {
		emit(map[string]any{"status": "Cleanup: failed to remove files from " + installPath + ": " + err.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: removed files from " + installPath})
}

// ShellLogs serves the output of script run in the site's PHP container with the docroot as $1, emptyMsg when it prints nothing
func ShellLogs(script, emptyMsg string) Handler {
	return func(a *appctx.App, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, _, userContext, err := auth.Injected(a, r)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		_, docroot, phpContainer, ok := RequestParams(ctx, a, r, userID, userContext)
		if !ok {
			http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
			return
		}

		argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script, "sh", docroot)
		out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if runErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(out)
			return
		}
		if len(strings.TrimSpace(string(out))) == 0 {
			_, _ = w.Write([]byte(emptyMsg))
			return
		}
		_, _ = w.Write(out)
	}
}
