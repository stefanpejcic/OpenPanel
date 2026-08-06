package backups

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// Register wires the backups routes onto mux, gated behind the "backups"
// feature flag.
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
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func envFilePath(userContext string) string {
	return "/home/" + userContext + "/backup.env"
}

// parseFormGroups pulls out every "values[KEY]"/"settings[KEY]" field into
// key->value. http.Request.PostForm is a map and doesn't preserve field
// order, so callers that need stable ordering read the updated keys off
// the env file's own line order instead.
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

// handleBackupSettings serves and updates the backup.env key/value form
// for the user's currently configured backup target.
func handleBackupSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path := envFilePath(userContext)

	if r.Method == http.MethodPost {
		if _, statErr := os.Stat(path); statErr != nil {
			_ = os.WriteFile(path, nil, 0o644)
		}

		if isBackupInProgress(r.Context(), userContext) {
			flashAndRedirect(a, w, r, "error", "A backup is currently in progress. Please wait until it finishes before saving settings. Or if you want to interrupt the backup process: stop and start the backup service.", r.URL.String())
			return
		}

		if postErr := saveBackupSettings(a, r, path, userContext, currentUsername); postErr != nil {
			flashSess(a, w, r, "error", "Error updating settings.")
		} else {
			flashSess(a, w, r, "success", "Settings updated successfully.")
		}
	}

	renderBackupSettingsFromFile(a, w, r, path)
}

func saveBackupSettings(a *appctx.App, r *http.Request, path, userContext, currentUsername string) error {
	content, err := os.ReadFile(path)
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
	for k, v := range updated {
		envDict[k] = v
	}

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

	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0o644); err != nil {
		return err
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated backup settings", reqip.ClientIP(r))

	ctx := r.Context()
	if isBackupContainerRunning(ctx, userContext) {
		docker.StartOrStopContainer(ctx, userContext, "backup", "deactivate", "")
	}
	docker.StartOrStopContainer(ctx, userContext, "backup", "activate", "")

	return nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func renderBackupSettingsFromFile(a *appctx.App, w http.ResponseWriter, r *http.Request, path string) {
	if _, err := os.Stat(path); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}

	entries, err := parseUncommentedEnv(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
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
		writeJSON(w, http.StatusOK, backupSettingsJSON(target, errMsg, grouped, values, settingsKV))
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

// handleBackupTarget switches (or reports) which backup destination
// section (s3/webdav/ssh/azure/dropbox) is active, by commenting out the
// other sections' keys in backup.env.
func handleBackupTarget(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path := envFilePath(userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		target := r.Form.Get("target")

		if _, ok := sectionKeys[target]; !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup target: " + target})
			return
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
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

		if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0o644); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "switched backup target to '"+target+"'", reqip.ClientIP(r))
		flashSess(a, w, r, "success", "Backup target switched to '"+target+"' successfully.")
	}

	if _, err := os.Stat(path); err != nil {
		renderBackupDestinationsPage(a, w, r, "")
		return
	}

	content, _ := os.ReadFile(path)
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
	active := ""
	if len(matched) == 1 {
		active = matched[0]
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"active": active, "targets": sectionOrder})
		return
	}

	renderBackupDestinationsPage(a, w, r, active)
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

// handleBackupsPage serves the backups landing page. Registered for GET
// and POST, but the handler body never branches on method, so POST
// behaves identically to GET - intentional, not an oversight.
func handleBackupsPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path := envFilePath(userContext)

	if _, statErr := os.Stat(path); statErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
		return
	}

	entries, err := parseUncommentedEnv(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "env file not found"})
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
		writeJSON(w, http.StatusOK, payload)
		return
	}

	renderBackupsPage(a, w, r, target, hasAnyCredentialMarker(values), serviceActive)
}

// hasAnyCredentialMarker reports whether any of the well-known
// per-destination credential keys (DROPBOX_APP_KEY, AWS_ACCESS_KEY_ID,
// WEBDAV_USERNAME, SSH_HOST_NAME, AZURE_STORAGE_ACCOUNT_NAME) has a value,
// used to decide whether the settings form should show as "configured".
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
