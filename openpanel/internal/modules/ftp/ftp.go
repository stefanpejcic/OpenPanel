// Package ftp implements FTP sub-account list/create/delete/password/path
// management, backed by the `opencli ftp-*` CLI tools and the FTP
// container's own /etc/openpanel/ftp/users/ flat-file account list.
package ftp

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

const ftpContainerName = "openadmin_ftp"

// Register wires the FTP account routes onto mux, gated behind the "ftp"
// feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "ftp")(h)
	}

	mux.Handle("GET /ftp", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFTPAccounts(a, w, r) }))
	mux.Handle("/ftp/connections", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleListFTPConnections(a, w, r) }))
	mux.Handle("/ftp/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAddFTPAccount(a, w, r) }))
	mux.Handle("POST /ftp/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteFTPAccount(a, w, r) }))
	mux.Handle("/ftp/password/{username}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleChangeFTPPassword(a, w, r, r.PathValue("username"))
	}))
	mux.Handle("/ftp/path/{username}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleChangeFTPPath(a, w, r, r.PathValue("username"))
	}))
	mux.Handle("GET /ftp/configuration/{type}/{account}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleFTPConfiguration(a, w, r, r.PathValue("type"), r.PathValue("account"))
	}))
}

// legalUsernameChars is the set of characters allowed in an FTP username.
const legalUsernameChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_@."

// isValidUsername reports whether username has exactly one "@", a
// non-empty local part and domain, and only characters drawn from the
// legal set.
func isValidUsername(username string) bool {
	if username == "" || !strings.Contains(username, "@") {
		return false
	}
	if strings.Count(username, "@") != 1 {
		return false
	}
	local, domain, _ := strings.Cut(username, "@")
	if local == "" || domain == "" {
		return false
	}
	for _, c := range username {
		if !strings.ContainsRune(legalUsernameChars, c) {
			return false
		}
	}
	return true
}

// isFTPContainerRunning reports whether the FTP container is currently
// running.
func isFTPContainerRunning(ctx context.Context, containerName string) bool {
	out, err := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+containerName, "--format", "{{.Names}}").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == containerName {
			return true
		}
	}
	return false
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirectToAccounts(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/ftp", http.StatusFound)
}

func requireFTPRunning(a *appctx.App, w http.ResponseWriter, r *http.Request) bool {
	if isFTPContainerRunning(r.Context(), ftpContainerName) {
		return true
	}
	flashAndRedirectToAccounts(a, w, r, "info", "FTP service is not yet started, please contact Administrator to enable it.")
	return false
}

// Account is one users.list row.
type Account struct {
	Username, Path, UID, GID string
}

// loadAccounts parses an FTP users.list file into Account records.
func loadAccounts(usersListFile string) []Account {
	content, err := os.ReadFile(usersListFile)
	if err != nil {
		return nil
	}

	var accounts []Account
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		switch {
		case len(parts) >= 5:
			accounts = append(accounts, Account{Username: parts[0], Path: parts[2], UID: parts[3], GID: parts[4]})
		case len(parts) == 3:
			accounts = append(accounts, Account{Username: parts[0], Path: parts[2]})
		default:
			continue
		}
	}
	return accounts
}

// handleFTPAccounts renders the FTP accounts list page.
func handleFTPAccounts(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	dedicatedIP := "Unknown"
	if ip, ok := appctx.ReadDedicatedIPFromFile(currentUsername); ok {
		dedicatedIP = ip
	}
	serverIP := a.GetCachedIPForUserOrPublicIPv4(r.Context(), currentUsername)

	userFolder := "/etc/openpanel/ftp/users/" + userContext
	usersListFile := userFolder + "/users.list"
	_ = os.MkdirAll(userFolder, 0o755)

	accounts := loadAccounts(usersListFile)

	renderFTPAccountsPage(a, w, r, serverIP, dedicatedIP, accounts)
}

// handleListFTPConnections renders the page listing currently active FTP
// connections for the user.
func handleListFTPConnections(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !requireFTPRunning(a, w, r) {
		return
	}

	out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-connections", currentUsername).Output()
	if cmdErr != nil {
		flashAndRedirectToAccounts(a, w, r, "info", "No currently active FTP connections")
		return
	}

	renderFTPConnectionsPage(a, w, r, string(out))
}

// handleAddFTPAccount handles both the new-FTP-account form page and its
// submission.
func handleAddFTPAccount(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !requireFTPRunning(a, w, r) {
		return
	}

	userDomains, _ := a.AllDomainsForUser(ctx, userID)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		ftpUsername := r.Form.Get("username")
		ftpPassword := r.Form.Get("password")
		ftpDomain := r.Form.Get("domain")
		ftpPath := r.Form.Get("path")

		if ftpUsername == "" {
			flashAndRedirect(a, w, r, "error", "FTP username not provided.", "/ftp/new")
			return
		}
		if ftpDomain == "" {
			flashAndRedirect(a, w, r, "error", "FTP domain not provided.", "/ftp/new")
			return
		}
		if !a.CheckDomainBelongsToUser(ctx, userID, ftpDomain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}
		if ftpPath == "" {
			flashAndRedirect(a, w, r, "error", "FTP path not provided.", "/ftp/new")
			return
		}
		if ftpPassword == "" {
			flashAndRedirect(a, w, r, "error", "FTP password not provided.", "/ftp/new")
			return
		}

		const realPath = "/var/www/html/"
		if !strings.HasPrefix(ftpPath, realPath) {
			flashAndRedirect(a, w, r, "error", "The FTP path must start with "+realPath, "/ftp/new")
			return
		}

		ftpUsername = ftpUsername + "@" + ftpDomain
		if !isValidUsername(ftpUsername) {
			flashAndRedirect(a, w, r, "error", "Username "+ftpUsername+" contains invalid characters, only allowed: BCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_@.", "/ftp/new")
			return
		}

		out, cmdErr := exec.CommandContext(ctx, "opencli", "ftp-add", ftpUsername, ftpPassword, ftpPath, currentUsername).CombinedOutput()
		if cmdErr != nil {
			flashAndRedirect(a, w, r, "error", "Error creating FTP user "+ftpUsername+": "+string(out), "/ftp/new")
			return
		}
		if strings.Contains(string(out), "Success: FTP user") {
			_ = logger.RecordUserAction(a.Config, currentUsername, "created FTP account "+ftpUsername, reqip.ClientIP(r))
			_ = a.Cache.Delete(ctx, "count_ftp_accounts:"+userContext)
			flashAndRedirectToAccounts(a, w, r, "success", "FTP account "+ftpUsername+" created successfully.")
			return
		}
		flashAndRedirect(a, w, r, "error", "Failed to create FTP account "+ftpUsername+": "+string(out), "/ftp/new")
		return
	}

	renderNewFTPAccountPage(a, w, r, userDomains)
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// handleDeleteFTPAccount deletes an FTP account by username.
func handleDeleteFTPAccount(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	usernameToDelete := r.Form.Get("username")

	if !requireFTPRunning(a, w, r) {
		return
	}

	out, _ := exec.CommandContext(r.Context(), "opencli", "ftp-delete", usernameToDelete, currentUsername).CombinedOutput()
	if strings.Contains(string(out), "Success") {
		_ = logger.RecordUserAction(a.Config, currentUsername, "deleted FTP account "+usernameToDelete, reqip.ClientIP(r))
		_ = a.Cache.Delete(r.Context(), "count_ftp_accounts:"+userContext)
		flashAndRedirectToAccounts(a, w, r, "success", "FTP account "+usernameToDelete+" deleted successfully!")
		return
	}
	flashAndRedirectToAccounts(a, w, r, "error", "Failed to delete FTP account "+usernameToDelete+". Output: "+string(out))
}

// handleChangeFTPPassword handles both the change-password form page and
// its submission for one FTP account.
func handleChangeFTPPassword(a *appctx.App, w http.ResponseWriter, r *http.Request, username string) {
	if !requireFTPRunning(a, w, r) {
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newPassword := r.Form.Get("new_password")
		if newPassword == "" {
			flashAndRedirectToAccounts(a, w, r, "error", "Missing new password")
			return
		}

		currentUsername, _, err := injected(a, r)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-password", username, newPassword, currentUsername).CombinedOutput()
		if cmdErr != nil {
			flashAndRedirectToAccounts(a, w, r, "error", "Error changing FTP password: "+string(out))
			return
		}
		if strings.Contains(string(out), "Success: FTP user") {
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed password for FTP account "+username, reqip.ClientIP(r))
			flashAndRedirectToAccounts(a, w, r, "success", "FTP password changed successfully for user "+username)
		} else {
			flashAndRedirectToAccounts(a, w, r, "error", "Failed to change FTP password: "+string(out))
		}
		return
	}

	renderFTPPasswordPage(a, w, r, username)
}

// handleChangeFTPPath handles both the change-path form page and its
// submission for one FTP account.
func handleChangeFTPPath(a *appctx.App, w http.ResponseWriter, r *http.Request, username string) {
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !requireFTPRunning(a, w, r) {
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		usernameToModify := r.Form.Get("username")
		newPath := r.Form.Get("new_path")
		const allowedPath = "/var/www/html/"
		if !strings.HasPrefix(newPath, allowedPath) {
			newPath = allowedPath + strings.TrimLeft(newPath, "/")
		}

		out, cmdErr := exec.CommandContext(r.Context(), "opencli", "ftp-path", usernameToModify, newPath, currentUsername).CombinedOutput()
		if cmdErr != nil {
			flashAndRedirectToAccounts(a, w, r, "error", "Error changing FTP path for user "+usernameToModify+": "+string(out))
			return
		}
		if strings.Contains(string(out), "Success: FTP path for user") {
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed path for FTP account "+usernameToModify, reqip.ClientIP(r))
			flashAndRedirectToAccounts(a, w, r, "success", "FTP path changed successfully for user "+usernameToModify)
		} else {
			flashAndRedirectToAccounts(a, w, r, "error", "Failed to change FTP path for user "+usernameToModify+": "+string(out))
		}
		return
	}

	renderFTPPathPage(a, w, r, username)
}

// handleFTPConfiguration generates a Cyberduck/FileZilla bookmark file for
// one account and streams it back as a download.
func handleFTPConfiguration(a *appctx.App, w http.ResponseWriter, r *http.Request, clientType, account string) {
	if clientType != "cyberduck" && clientType != "filezilla" {
		flashAndRedirectToAccounts(a, w, r, "danger", "Invalid configuration type, please use cyberduck or filezilla only!")
		return
	}

	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	usersListFile := "/etc/openpanel/ftp/users/" + userContext + "/users.list"
	content, readErr := os.ReadFile(usersListFile)
	if readErr != nil {
		flashAndRedirectToAccounts(a, w, r, "danger", "User has no FTP accounts.")
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
		flashAndRedirectToAccounts(a, w, r, "danger", "FTP account not found.")
		return
	}

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

	var content2, filename, mimeType string
	switch clientType {
	case "cyberduck":
		content2 = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<bookmark>\n    <hostname>" + ftpHost + "</hostname>\n    <username>" + accountData.Username + "</username>\n    <protocol>ftp</protocol>\n    <path>" + accountData.Path + "</path>\n</bookmark>"
		filename = account + "_cyberduck.ftpbookmark"
		mimeType = "application/xml"
	case "filezilla":
		content2 = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<FileZilla3>\n    <Servers>\n        <Server>\n            <Host>" + ftpHost + "</Host>\n            <Port>21</Port>\n            <Protocol>0</Protocol>\n            <Type>0</Type>\n            <User>" + accountData.Username + "</User>\n            <Logontype>2</Logontype>\n            <EncodingType>Auto</EncodingType>\n            <Name>" + account + "</Name>\n            <RemoteDir>" + accountData.Path + "</RemoteDir>\n            <UsePassive>1</UsePassive>\n            <BypassProxy>0</BypassProxy>\n            <Encryption>0</Encryption>\n        </Server>\n    </Servers>\n</FileZilla3>"
		filename = account + "_filezilla.xml"
		mimeType = "application/xml"
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded "+clientType+" configuration for FTP account "+account, reqip.ClientIP(r))

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	_, _ = w.Write([]byte(content2))
}

// domainOption is one ftp_new.html <select> entry.
type domainOption struct {
	DomainURL, Docroot string
}

func domainOptions(list []appctx.Domain) []domainOption {
	opts := make([]domainOption, len(list))
	for i, d := range list {
		opts[i] = domainOption{DomainURL: d.DomainURL, Docroot: d.Docroot}
	}
	return opts
}
