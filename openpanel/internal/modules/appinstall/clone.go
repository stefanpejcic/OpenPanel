package appinstall

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterPM2Clone wires the /pm2/clone route onto mux.
func RegisterPM2Clone(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "helpers")(h)
	}
	mux.Handle("POST /pm2/clone/{site_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePM2Clone(a, w, r) }))
}

// readPrefixedEnvValues reads envFile and returns every PREFIX+KEY="value"
// line's KEY -> value, quotes stripped - the read-side counterpart of the
// PREFIX+KEY="value" lines handlePM2Update/HandleInstall write.
func readPrefixedEnvValues(envFile, prefix string) map[string]string {
	values := map[string]string{}
	content, err := os.ReadFile(envFile)
	if err != nil {
		return values
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimPrefix(line, prefix)
		key, value, found := strings.Cut(rest, "=")
		if !found {
			continue
		}
		values[key] = strings.Trim(value, `"`)
	}
	return values
}

// handlePM2Clone copies an existing Python/NodeJS/Ruby app's files and
// settings (version, requirements, custom command, git URL, resource
// limits) onto a new domain/subdirectory, for quick clone/staging copies.
// It reuses HandleInstall itself (a direct Go call, not a new network
// request) to create the new container/compose entry/webserver proxy/DB
// row exactly as a fresh install would, then copies the source docroot's
// files into the new one - HandleInstall never touches docroot files
// itself (the app's own code is always added after install, normally via
// git deploy or the file manager), so the copy step can run right after
// without racing or overwriting anything HandleInstall wrote.
func handlePM2Clone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	siteName := strings.ToLower(r.PathValue("site_name"))

	var lookup pm2SiteLookup
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.type, sites.site_name, sites.container
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.container LIKE ? AND domains.user_id = ?`, "%"+siteName+"%", userID)
	if scanErr := row.Scan(&lookup.Type, &lookup.SiteName, &lookup.Container); scanErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
		return
	}
	kind, ok := kindByAppType(lookup.Type)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Only NodeJS, Python, or Ruby applications can be cloned"})
		return
	}

	targetDomainID := r.FormValue("target_domain_id")
	targetSubdirectory := strings.ToLower(strings.TrimSpace(r.FormValue("target_subdirectory")))
	targetServiceName := strings.ToLower(strings.TrimSpace(r.FormValue("target_service_name")))
	startupFile := strings.TrimSpace(r.FormValue("startup_file"))
	if targetDomainID == "" || targetServiceName == "" || startupFile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Target domain, service name, and startup file are required"})
		return
	}

	var targetTopDomain string
	var targetDocrootNull sql.NullString
	if scanErr := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot FROM domains WHERE domain_id = ?", targetDomainID).
		Scan(&targetTopDomain, &targetDocrootNull); scanErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Target domain not found"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, targetTopDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	targetDocroot := targetDocrootNull.String

	srcDomain, srcSubdirectory, _ := strings.Cut(lookup.SiteName, "/")
	var srcDocrootNull sql.NullString
	if scanErr := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", srcDomain).Scan(&srcDocrootNull); scanErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Source domain not found"})
		return
	}
	srcDocroot := srcDocrootNull.String
	srcPath := srcDocroot
	if srcSubdirectory != "" {
		srcPath += "/" + srcSubdirectory
	}

	// Read the source app's current settings to carry over onto the clone
	// - the same PREFIX+KEY="value" lines handlePM2Update writes.
	srcPrefix := strings.ToUpper(lookup.Container) + "_" + kind.PyOrNode + "_"
	envFile := "/home/" + userContext + "/.env"
	src := readPrefixedEnvValues(envFile, srcPrefix)

	targetPath := targetDocroot
	if targetSubdirectory != "" {
		targetPath += "/" + targetSubdirectory
	}
	targetSiteName := targetTopDomain
	if targetSubdirectory != "" {
		targetSiteName += "/" + targetSubdirectory
	}

	const wwwBase = "/var/www/html/"
	hostBase := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	srcHostPath := strings.Replace(srcPath, wwwBase, hostBase, 1)
	dstHostPath := strings.Replace(targetPath, wwwBase, hostBase, 1)

	// Copy the source app's files into place *before* triggering the
	// install below - HandleInstall only considers the install successful
	// once the container is actually observed running (polled for a few
	// seconds after start), and a Ruby/Python/NodeJS container immediately
	// exits if its startup file doesn't exist yet, which HandleInstall
	// would then treat as a failed install and roll everything back. A
	// fresh (non-clone) install normally avoids this either by pointing
	// git_repo_url at a repo (auto-cloned into the container at startup)
	// or by the user uploading files after install completes - a clone's
	// whole point is to already have real files, so they need to land
	// before HandleInstall's own container-start check runs, not after.
	if info, statErr := os.Stat(srcHostPath); statErr != nil || !info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Source folder not found on disk: " + srcPath})
		return
	}
	if mkErr := os.MkdirAll(dstHostPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to prepare target folder: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcHostPath+"/.", dstHostPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy application files: " + cpErr.Error()})
		return
	}
	_ = exec.CommandContext(ctx, "chown", "-R", userContext+":"+userContext, dstHostPath).Run()

	installForm := url.Values{
		"domain_id":    {targetDomainID},
		"service_name": {targetServiceName},
		"subdirectory": {targetSubdirectory},
		"startup_file": {startupFile},
		"version":      {orDefault(src["TAG"], "latest")},
		"requirements": {src["REQUIREMENTS"]},
		"custom_cmd":   {src["CUSTOM_CMD"]},
		"git_repo_url": {src["GIT_URL"]},
		"cpu_limit":    {orDefault(src["CPU"], "1.0")},
		"mem_limit":    {orDefault(strings.TrimSuffix(src["RAM"], "G"), "1.0")},
		"pids_limit":   {orDefault(src["PIDS"], "100")},
	}

	installReq := httptest.NewRequest(http.MethodPost, "/"+kind.AppType+"/install", strings.NewReader(installForm.Encode()))
	installReq = installReq.WithContext(ctx)
	installReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	installRec := httptest.NewRecorder()
	HandleInstall(kind, a, installRec, installReq)

	installOutput := installRec.Body.String()
	if strings.Contains(installOutput, `"error"`) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clone failed: " + firstNDJSONError(installOutput)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned "+kind.DisplayAppType+" application from "+lookup.SiteName+" to "+targetSiteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "target": targetSiteName})
}

func orDefault(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// firstNDJSONError pulls the value out of the first `{"error":"..."}` line
// HandleInstall's ndjson stream emitted, for a one-line summary instead of
// dumping the whole stream back to the clone form.
func firstNDJSONError(ndjson string) string {
	for _, line := range strings.Split(ndjson, "\n") {
		if !strings.Contains(line, `"error"`) {
			continue
		}
		if _, after, found := strings.Cut(line, `"error":"`); found {
			if end := strings.Index(after, `"`); end != -1 {
				return after[:end]
			}
		}
		return line
	}
	return "unknown error"
}
