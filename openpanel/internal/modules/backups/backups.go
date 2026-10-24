package backups

import (
	"log"
	"maps"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Register wires the backups routes onto mux, gated behind the "backups" feature flag
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "backups")(h)
	}

	mux.Handle("/backups/settings", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupSettings(a, w, r) }))
	mux.Handle("/backups/destination", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupTarget(a, w, r) }))
	mux.Handle("/backups", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupsPage(a, w, r) }))
	mux.Handle("/backups/list", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleListBackupsFromDestination(a, w, r) }))
	mux.Handle("POST /backups/restore", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRestoreFromBackup(a, w, r) }))
	mux.Handle("POST /backups/download", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDownloadBackup(a, w, r) }))
	mux.Handle("GET /backups/files", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupFiles(a, w, r) }))
	mux.Handle("POST /backups/download-file", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDownloadBackupFile(a, w, r) }))
	mux.Handle("POST /backups/download-items", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDownloadItems(a, w, r) }))
	mux.Handle("POST /backups/queue", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleQueueRestore(a, w, r) }))
	mux.Handle("GET /backups/jobs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRestoreJobs(a, w, r) }))
}

func envFilePath(userContext string) string {
	return "/home/" + userContext + "/backup.env"
}

// parseFormGroups pulls out every "values[KEY]"/"settings[KEY]" field into key->value. PostForm is a map and doesn't preserve field order, so callers needing stable ordering read the updated keys off the env file's own line order instead.
func parseFormGroups(r *http.Request, group string) map[string]string {
	out := map[string]string{}
	prefix := group + "["
	for key, values := range r.PostForm {
		if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, "]") && len(values) > 0 {
			out[key[len(prefix):len(key)-1]] = values[0]
		}
	}
	return out
}

// handleBackupSettings serves and updates the backup.env key/value form for the user's currently configured backup target
func handleBackupSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if AdminManagedBackups(userContext) {
		respondBackupsAdminManaged(a, w, r)
		return
	}
	path := envFilePath(userContext)

	if r.Method == http.MethodPost {
		if _, statErr := tenantfs.Stat(path); statErr != nil {
			_ = tenantfs.WriteFile(path, nil, 0o644)
		}

		status, msg := updateBackupSettings(a, r, path, userContext, currentUsername)
		// the API and the onboarding wizard read the status, a refused or failed save used to come back as the unchanged settings
		if status != http.StatusOK && r.URL.Query().Get("output") == "json" {
			web.WriteJSON(w, status, map[string]string{"error": msg})
			return
		}
		switch status {
		case http.StatusConflict:
			web.FlashRedirect(a, w, r, "error", msg, r.URL.String())
			return
		case http.StatusOK:
			web.Flash(a, w, r, "success", msg)
		default:
			web.Flash(a, w, r, "error", msg)
		}
	}

	renderBackupSettingsFromFile(a, w, r, path)
}

// updateBackupSettings saves the posted settings unless a backup is running, returns the status and message to show
func updateBackupSettings(a *appctx.App, r *http.Request, path, userContext, currentUsername string) (int, string) {
	if isBackupInProgress(r.Context(), userContext) {
		return http.StatusConflict, "A backup is currently in progress. Please wait until it finishes before saving settings. Or if you want to interrupt the backup process: stop and start the backup service."
	}
	if err := saveBackupSettings(a, r, path, userContext, currentUsername); err != nil {
		return http.StatusInternalServerError, "Error updating settings."
	}
	return http.StatusOK, "Settings updated successfully."
}

func saveBackupSettings(a *appctx.App, r *http.Request, path, userContext, currentUsername string) error {
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	envDict := map[string]string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "=") {
			continue
		}
		key, val, _ := strings.Cut(trimmed, "=")
		envDict[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}

	_ = r.ParseForm()
	updated := map[string]string{}
	for k, v := range parseFormGroups(r, "values") {
		updated[k] = v
	}
	for k, v := range parseFormGroups(r, "settings") {
		updated[k] = v
	}
	before := maps.Clone(envDict)
	for k, v := range updated {
		envDict[k] = v
	}
	pinSSHHostKey(before, envDict, updated, fetchHostKey)

	var newLines []string
	existing := map[string]bool{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "=") {
			newLines = append(newLines, line)
			continue
		}
		key := strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0])
		if _, ok := updated[key]; ok {
			newLines = append(newLines, key+"="+envDict[key])
			existing[key] = true
		} else {
			newLines = append(newLines, line)
		}
	}
	for k := range updated {
		if !existing[k] {
			newLines = append(newLines, k+"="+envDict[k])
		}
	}

	if err := tenantfs.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0o644); err != nil {
		return err
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated backup settings", reqip.ClientIP(r))

	ctx := r.Context()
	if isBackupContainerRunning(ctx, userContext) {
		docker.StartOrStopContainer(ctx, userContext, "backup", "deactivate", "")
	}
	docker.StartOrStopContainer(ctx, userContext, "backup", "activate", "")
	// compose down backup takes docker-proxy with it and the up is --no-deps, without the proxy every backup fails before it starts
	docker.StartComposeServiceIfNotRunning(ctx, userContext, "docker-proxy")

	return nil
}

// pinSSHHostKey stores the SSH destination's host key when there's none yet or the host changed, docker-volume-backup won't connect without it
func pinSSHHostKey(before, after, updated map[string]string, fetch func(map[string]string) (string, error)) {
	host := unquoteEnv(after["SSH_HOST_NAME"])
	if host == "" {
		return
	}
	if typed := unquoteEnv(updated["SSH_HOST_PUBLIC_KEY"]); typed != "" {
		return
	}
	moved := unquoteEnv(before["SSH_HOST_NAME"]) != host || unquoteEnv(before["SSH_PORT"]) != unquoteEnv(after["SSH_PORT"])
	if unquoteEnv(after["SSH_HOST_PUBLIC_KEY"]) != "" && !moved {
		return
	}
	key, err := fetch(map[string]string{"SSH_HOST_NAME": host, "SSH_PORT": unquoteEnv(after["SSH_PORT"]), "SSH_USER": unquoteEnv(after["SSH_USER"])})
	if err != nil {
		log.Printf("BACKUPS - could not read the host key of %s: %v", host, err)
		return
	}
	value := `"` + key + `"`
	updated["SSH_HOST_PUBLIC_KEY"] = value
	after["SSH_HOST_PUBLIC_KEY"] = value
}

func renderBackupSettingsFromFile(a *appctx.App, w http.ResponseWriter, r *http.Request, path string) {
	if _, err := tenantfs.Stat(path); err != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}

	entries, err := parseUncommentedEnv(path)
	if err != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}
	grouped := groupBySections(entries)

	var target, errMsg string
	var values, settingsKV []KV
	switch len(grouped.MatchedSections) {
	case 1:
		target = grouped.MatchedSections[0]
		values = grouped.SectionValues[target]
		settingsKV = grouped.Settings
	case 0:
		errMsg = "no backup target configured"
		settingsKV = grouped.Settings
	default:
		errMsg = "multiple backup targets configured"
		settingsKV = grouped.Settings
	}

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, backupSettingsJSON(target, errMsg, grouped, values, settingsKV))
		return
	}

	renderBackupSettingsPage(a, w, r, errMsg, target, values, settingsKV)
}

func backupSettingsJSON(target, errMsg string, grouped EnvSections, values, settingsKV []KV) map[string]any {
	settingsMap := kvToMap(settingsKV)
	switch {
	case target != "":
		return map[string]any{"target": target, "values": kvToMap(values), "settings": settingsMap}
	case errMsg == "no backup target configured":
		return map[string]any{"error": errMsg, "settings": settingsMap}
	default:
		valuesBySection := map[string]map[string]string{}
		for _, section := range grouped.MatchedSections {
			valuesBySection[section] = kvToMap(grouped.SectionValues[section])
		}
		return map[string]any{"error": errMsg, "targets": grouped.MatchedSections, "values": valuesBySection, "settings": settingsMap}
	}
}

func kvToMap(entries []KV) map[string]string {
	m := make(map[string]string, len(entries))
	for _, kv := range entries {
		m[kv.Key] = kv.Value
	}
	return m
}

// handleBackupTarget switches (or reports) which backup destination section (s3/webdav/ssh/azure/dropbox) is active, by commenting out the other sections' keys in backup.env
func handleBackupTarget(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if AdminManagedBackups(userContext) {
		respondBackupsAdminManaged(a, w, r)
		return
	}
	path := envFilePath(userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		target := r.Form.Get("target")

		if _, ok := sectionKeys[target]; !ok {
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup target: " + target})
			return
		}
		content, readErr := tenantfs.ReadFile(path)
		if readErr != nil {
			web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
			return
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		var newLines []string
		for _, line := range lines {
			stripped := strings.TrimSpace(line)
			if stripped == "" || !strings.Contains(stripped, "=") {
				newLines = append(newLines, line)
				continue
			}
			keyPart := strings.TrimSpace(strings.SplitN(strings.TrimLeft(stripped, "#"), "=", 2)[0])

			switch {
			case isSectionKey(target, keyPart):
				newLines = append(newLines, strings.TrimSpace(strings.TrimLeft(stripped, "#")))
			case belongsToOtherSection(target, keyPart):
				newLines = append(newLines, "# "+strings.TrimSpace(strings.TrimLeft(stripped, "#")))
			default:
				newLines = append(newLines, line)
			}
		}

		if err := tenantfs.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0o644); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "switched backup target to '"+target+"'", reqip.ClientIP(r))
		web.Flash(a, w, r, "success", web.Tr(a, r, "Backup target switched to '%(target)s' successfully.", "target", target))
	}

	if _, err := tenantfs.Stat(path); err != nil {
		renderBackupDestinationsPage(a, w, r, "")
		return
	}

	active := ActiveBackupTarget(userContext)

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{"active": active, "targets": sectionOrder})
		return
	}

	renderBackupDestinationsPage(a, w, r, active)
}

// ActiveBackupTarget is the one destination with uncommented keys in backup.env, or "" when none or several are
func ActiveBackupTarget(userContext string) string {
	content, err := tenantfs.ReadFile(envFilePath(userContext))
	if err != nil {
		return ""
	}
	uncommentedKeys := map[string]bool{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		key := strings.TrimSpace(strings.SplitN(line, "=", 2)[0])
		uncommentedKeys[key] = true
	}

	var matched []string
	for _, section := range sectionOrder {
		for _, key := range sectionKeys[section] {
			if uncommentedKeys[key] {
				matched = append(matched, section)
				break
			}
		}
	}
	if len(matched) == 1 {
		return matched[0]
	}
	return ""
}

func belongsToOtherSection(target, key string) bool {
	for _, section := range sectionOrder {
		if section == target {
			continue
		}
		if isSectionKey(section, key) {
			return true
		}
	}
	return false
}

// handleBackupsPage serves the backups landing page. Registered for GET and POST, but never branches on method, so POST behaves identically to GET - intentional, not an oversight.
func handleBackupsPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	adminManaged := AdminManagedBackups(userContext)
	// nothing to configure on an admin-managed account, so the restore page is the landing page
	if adminManaged && r.URL.Query().Get("output") != "json" {
		http.Redirect(w, r, "/backups/list", http.StatusFound)
		return
	}
	path := envFilePath(userContext)

	if _, statErr := tenantfs.Stat(path); statErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}

	entries, err := parseUncommentedEnv(path)
	if err != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}
	grouped := groupBySections(entries)
	serviceActive := isBackupContainerRunning(r.Context(), userContext)

	var target string
	var values []KV
	if len(grouped.MatchedSections) == 1 {
		target = grouped.MatchedSections[0]
		values = grouped.SectionValues[target]
	}

	if r.URL.Query().Get("output") == "json" {
		// admin-managed accounts don't get destination/settings details (which include stored credentials) here - just enough to know a backup target exists and whether the service is running
		if adminManaged {
			web.WriteJSON(w, http.StatusOK, map[string]any{"admin_managed": true, "configured": target != "", "service_active": serviceActive})
			return
		}
		var errMsg string
		switch len(grouped.MatchedSections) {
		case 0:
			errMsg = "no backup target configured"
		case 1:
		default:
			errMsg = "multiple backup targets configured"
		}
		payload := backupSettingsJSON(target, errMsg, grouped, values, grouped.Settings)
		payload["service_active"] = serviceActive
		web.WriteJSON(w, http.StatusOK, payload)
		return
	}

	renderBackupsPage(a, w, r, target, hasAnyCredentialMarker(values), serviceActive, adminManaged)
}

// hasAnyCredentialMarker reports whether any well-known per-destination credential key has a value, used to decide whether the settings form should show as "configured"
func hasAnyCredentialMarker(values []KV) bool {
	markers := map[string]bool{
		"DROPBOX_APP_KEY": true, "AWS_ACCESS_KEY_ID": true, "WEBDAV_USERNAME": true,
		"SSH_HOST_NAME": true, "AZURE_STORAGE_ACCOUNT_NAME": true,
	}
	for _, kv := range values {
		if markers[kv.Key] && kv.Value != "" {
			return true
		}
	}
	return false
}
