package ftp

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the FTP module's API routes onto mux. PATCH
// .../password and .../path share a username prefix with a literal
// suffix - Go's http.ServeMux requires a "{...}" wildcard to be the final
// segment, so both are merged into one "{rest...}" catch-all and
// apiFTPPatchDispatch strips the known suffix by hand to recover the
// username and dispatch to the right handler. apiregistry.Add still
// records each logical route separately for /api/endpoints.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "ftp", "GET /api/ftp", func(w http.ResponseWriter, r *http.Request) { apiFTPList(a, w, r) })
	apiregistry.Handle(mux, a, "ftp", "POST /api/ftp", func(w http.ResponseWriter, r *http.Request) { apiFTPCreate(a, w, r) })
	apiregistry.Handle(mux, a, "ftp", "DELETE /api/ftp/{username...}", func(w http.ResponseWriter, r *http.Request) { apiFTPDelete(a, w, r) })

	apiregistry.Add("PATCH /api/ftp/{username}/password")
	apiregistry.Add("PATCH /api/ftp/{username}/path")
	mux.Handle("PATCH /api/ftp/{rest...}", auth.RequireAPI(a, "ftp")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiFTPPatchDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "ftp", "GET /api/ftp/connections", func(w http.ResponseWriter, r *http.Request) { apiFTPConnections(a, w, r) })
	apiregistry.Handle(mux, a, "ftp", "GET /api/ftp/configuration/{config_type}/{account...}", func(w http.ResponseWriter, r *http.Request) { apiFTPConfiguration(a, w, r) })
}

// apiFTPPatchDispatch dispatches PATCH /api/ftp/{username}/password|path.
func apiFTPPatchDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/password"):
		r.SetPathValue("username", strings.TrimSuffix(rest, "/password"))
		apiFTPPassword(a, w, r)
	case strings.HasSuffix(rest, "/path"):
		r.SetPathValue("username", strings.TrimSuffix(rest, "/path"))
		apiFTPPath(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

func writeAPIFTPJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiFTPServiceCheck writes an error response and returns false if the FTP
// container isn't running.
func apiFTPServiceCheck(w http.ResponseWriter, r *http.Request) bool {
	if isFTPContainerRunning(r.Context(), ftpContainerName) {
		return true
	}
	writeAPIFTPJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "FTP service is not running. Contact the administrator to enable it."})
	return false
}

// apiServerIP returns the FTP host to advertise for a user: their
// dedicated IP if one is configured, otherwise the shared server IP.
func apiServerIP(a *appctx.App, r *http.Request, currentUsername string) string {
	ftpHost := a.GetCachedIPForUserOrPublicIPv4(r.Context(), currentUsername)
	dedicatedIPPath := "/etc/openpanel/openpanel/core/users/" + currentUsername + "/ip.json"
	if data, err := os.ReadFile(dedicatedIPPath); err == nil {
		var parsed struct {
			IP string `json:"ip"`
		}
		if json.Unmarshal(data, &parsed) == nil && parsed.IP != "" {
			ftpHost = parsed.IP
		}
	}
	return ftpHost
}

type apiAccountEntry struct {
	Username string  `json:"username"`
	Path     string  `json:"path"`
	UID      *string `json:"uid"`
	GID      *string `json:"gid"`
}

// apiFTPList returns the FTP accounts configured for the current user.
func apiFTPList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	usersListFile := "/etc/openpanel/ftp/users/" + userContext + "/users.list"
	accounts := loadAccounts(usersListFile)
	entries := make([]apiAccountEntry, 0, len(accounts))
	for _, acc := range accounts {
		e := apiAccountEntry{Username: acc.Username, Path: acc.Path}
		if acc.UID != "" {
			uid := acc.UID
			e.UID = &uid
		}
		if acc.GID != "" {
			gid := acc.GID
			e.GID = &gid
		}
		entries = append(entries, e)
	}

	writeAPIFTPJSON(w, http.StatusOK, map[string]any{
		"accounts":  entries,
		"server_ip": a.GetCachedIPForUserOrPublicIPv4(r.Context(), currentUsername),
		"ftp_host":  apiServerIP(a, r, currentUsername),
	})
}

// apiFTPCreate creates a new FTP account scoped to a domain the caller
// owns.
func apiFTPCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Domain   string `json:"domain"`
		Path     string `json:"path"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Username = r.Form.Get("username")
		body.Password = r.Form.Get("password")
		body.Domain = r.Form.Get("domain")
		body.Path = r.Form.Get("path")
	}
	ftpUser := strings.TrimSpace(body.Username)
	ftpPassword := strings.TrimSpace(body.Password)
	ftpDomain := strings.TrimSpace(body.Domain)
	ftpPath := strings.TrimSpace(body.Path)

	if ftpUser == "" || ftpPassword == "" || ftpDomain == "" || ftpPath == "" {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "username, password, domain, and path are required"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, ftpDomain) {
		writeAPIFTPJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	const allowedPath = "/var/www/html/"
	if !strings.HasPrefix(ftpPath, allowedPath) {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "Path must start with " + allowedPath})
		return
	}

	fullUsername := ftpUser + "@" + ftpDomain
	if !isValidUsername(fullUsername) {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "Username contains invalid characters (allowed: A-Z a-z 0-9 - _ @ .)"})
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "ftp-add", fullUsername, ftpPassword, ftpPath, currentUsername).CombinedOutput()
	if cmdErr != nil {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}
	if !strings.Contains(string(out), "Success: FTP user") {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create FTP account: " + string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created FTP account "+fullUsername, reqip.ClientIP(r))
	_ = a.Cache.Delete(ctx, "count_ftp_accounts:"+userContext)
	writeAPIFTPJSON(w, http.StatusCreated, map[string]string{"message": "FTP account " + fullUsername + " created", "username": fullUsername})
}

// apiFTPDelete removes an FTP account.
func apiFTPDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username := r.PathValue("username")

	out, _ := exec.CommandContext(r.Context(), "opencli", "ftp-delete", username, currentUsername).CombinedOutput()
	if !strings.Contains(string(out), "Success") {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete " + username + ": " + string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted FTP account "+username, reqip.ClientIP(r))
	_ = a.Cache.Delete(r.Context(), "count_ftp_accounts:"+userContext)
	writeAPIFTPJSON(w, http.StatusOK, map[string]string{"message": "FTP account " + username + " deleted"})
}

// apiFTPPassword changes an FTP account's password.
func apiFTPPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username := r.PathValue("username")

	var body struct {
		Password string `json:"password"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Password = r.Form.Get("password")
	}
	newPassword := strings.TrimSpace(body.Password)
	if newPassword == "" {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "password is required"})
		return
	}

	out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-password", username, newPassword, currentUsername).CombinedOutput()
	if cmdErr != nil {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}
	if !strings.Contains(string(out), "Success: FTP user") {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to change password: " + string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed password for FTP account "+username, reqip.ClientIP(r))
	writeAPIFTPJSON(w, http.StatusOK, map[string]string{"message": "Password changed for " + username})
}

// apiFTPPath changes an FTP account's root path.
func apiFTPPath(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username := r.PathValue("username")

	var body struct {
		Path string `json:"path"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Path = r.Form.Get("path")
	}
	newPath := strings.TrimSpace(body.Path)
	if newPath == "" {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "path is required"})
		return
	}

	const allowedPath = "/var/www/html/"
	if !strings.HasPrefix(newPath, allowedPath) {
		newPath = allowedPath + strings.TrimLeft(newPath, "/")
	}

	out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-path", username, newPath, currentUsername).CombinedOutput()
	if cmdErr != nil {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}
	if !strings.Contains(string(out), "Success: FTP path for user") {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to change path: " + string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed path for FTP account "+username, reqip.ClientIP(r))
	writeAPIFTPJSON(w, http.StatusOK, map[string]string{"message": "Path changed to " + newPath, "path": newPath})
}

// apiFTPConnections returns the current user's active FTP connections.
func apiFTPConnections(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-connections", currentUsername).Output()
	if cmdErr != nil {
		writeAPIFTPJSON(w, http.StatusOK, map[string]string{"connections": "", "message": "No active FTP connections"})
		return
	}
	writeAPIFTPJSON(w, http.StatusOK, map[string]string{"connections": string(out)})
}

// apiFTPConfiguration generates a client config file (Cyberduck or
// FileZilla) for an FTP account.
func apiFTPConfiguration(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	configType := r.PathValue("config_type")
	if configType != "cyberduck" && configType != "filezilla" {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid type (cyberduck, filezilla)"})
		return
	}

	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	account := r.PathValue("account")

	usersListFile := "/etc/openpanel/ftp/users/" + userContext + "/users.list"
	content, readErr := os.ReadFile(usersListFile)
	if readErr != nil {
		writeAPIFTPJSON(w, http.StatusNotFound, map[string]string{"error": "No FTP accounts found"})
		return
	}

	var accountData *Account
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 && parts[0] == account {
			accountData = &Account{Username: parts[0], Path: parts[2]}
			break
		}
	}
	if accountData == nil {
		writeAPIFTPJSON(w, http.StatusNotFound, map[string]string{"error": "FTP account not found"})
		return
	}

	ftpHost := apiServerIP(a, r, currentUsername)

	var xmlContent, filename string
	if configType == "cyberduck" {
		xmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<bookmark>\n    <hostname>" + ftpHost + "</hostname>\n    <username>" + accountData.Username + "</username>\n    <protocol>ftp</protocol>\n    <path>" + accountData.Path + "</path>\n</bookmark>"
		filename = account + "_cyberduck.ftpbookmark"
	} else {
		xmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<FileZilla3>\n    <Servers>\n        <Server>\n            <Host>" + ftpHost + "</Host>\n            <Port>21</Port>\n            <Protocol>0</Protocol>\n            <Type>0</Type>\n            <User>" + accountData.Username + "</User>\n            <Logontype>2</Logontype>\n            <EncodingType>Auto</EncodingType>\n            <Name>" + account + "</Name>\n            <RemoteDir>" + accountData.Path + "</RemoteDir>\n            <UsePassive>1</UsePassive>\n        </Server>\n    </Servers>\n</FileZilla3>"
		filename = account + "_filezilla.xml"
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded "+configType+" config for FTP account "+account, reqip.ClientIP(r))

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, _ = w.Write([]byte(xmlContent))
}
