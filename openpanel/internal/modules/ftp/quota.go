package ftp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// quotaUsage is one quota.usage row, the FTP container rewrites that file every minute for accounts with a quota
type quotaUsage struct {
	UsedKB int64
	Over   bool
}

// loadQuotaUsage reads name|used_kb|quota_mb|over rows from the quota.usage file next to users.list
func loadQuotaUsage(usersListFile string) map[string]quotaUsage {
	content, err := tenantfs.ReadFile(filepath.Join(filepath.Dir(usersListFile), "quota.usage"))
	if err != nil {
		return nil
	}
	usage := map[string]quotaUsage{}
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Split(strings.TrimSpace(line), "|")
		if len(parts) < 4 || parts[0] == "" {
			continue
		}
		used, _ := strconv.ParseInt(parts[1], 10, 64)
		usage[parts[0]] = quotaUsage{UsedKB: used, Over: parts[3] == "1"}
	}
	return usage
}

// parseQuotaMB reads the 6th users.list field, anything that isn't a positive number is unlimited
func parseQuotaMB(field string) int {
	n, err := strconv.Atoi(strings.TrimSpace(field))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// HasQuota is false for unlimited accounts.
func (acc Account) HasQuota() bool { return acc.QuotaMB > 0 }

// QuotaLabel is the quota for display, e.g. "500 MB" or "2 GB".
func (acc Account) QuotaLabel() string { return formatMB(acc.QuotaMB) }

// UsedLabel is the disk usage for display, empty until the FTP container has checked the account.
func (acc Account) UsedLabel() string {
	if !acc.HasUsage {
		return ""
	}
	return formatKB(acc.UsedKB)
}

// UsedPercent is the share of the quota in use, capped at 100 for the progress bar.
func (acc Account) UsedPercent() int {
	if !acc.HasQuota() || !acc.HasUsage {
		return 0
	}
	pct := int(acc.UsedKB * 100 / (int64(acc.QuotaMB) * 1024))
	if pct > 100 {
		return 100
	}
	return pct
}

// QuotaInput is the quota as a value for the form's number and unit fields.
func (acc Account) QuotaInput() (string, string) {
	if !acc.HasQuota() {
		return "", "M"
	}
	if acc.QuotaMB%1024 == 0 {
		return strconv.Itoa(acc.QuotaMB / 1024), "G"
	}
	return strconv.Itoa(acc.QuotaMB), "M"
}

func formatMB(mb int) string {
	if mb >= 1024 {
		if mb%1024 == 0 {
			return fmt.Sprintf("%d GB", mb/1024)
		}
		return fmt.Sprintf("%.1f GB", float64(mb)/1024)
	}
	return fmt.Sprintf("%d MB", mb)
}

func formatKB(kb int64) string {
	switch {
	case kb < 1024:
		return fmt.Sprintf("%d KB", kb)
	case kb < 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(kb)/1024)
	default:
		return fmt.Sprintf("%.2f GB", float64(kb)/1024/1024)
	}
}

var quotaInputRe = regexp.MustCompile(`^([0-9]+)\s*([MG])?B?$`)

// normalizeQuota turns form or API input into what opencli ftp-quota takes, empty and 0 mean unlimited
func normalizeQuota(value, unit string) (string, bool) {
	v := strings.ToUpper(strings.TrimSpace(value))
	if v == "" || v == "UNLIMITED" {
		return "unlimited", true
	}
	m := quotaInputRe.FindStringSubmatch(v)
	if m == nil {
		return "", false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n > 1024*1024*100 {
		return "", false
	}
	if n == 0 {
		return "unlimited", true
	}
	u := m[2]
	if u == "" {
		u = strings.ToUpper(strings.TrimSpace(unit))
	}
	if u != "G" {
		u = "M"
	}
	return strconv.Itoa(n) + u, true
}

// rawQuotaString accepts the API's quota as a number (500) or a string ("2G"), null and missing mean unlimited
func rawQuotaString(raw json.RawMessage) string {
	v := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if v == "null" {
		return ""
	}
	return v
}

// findAccount returns the users.list entry for username, if this user has it
func findAccount(userContext, username string) (Account, bool) {
	for _, acc := range loadAccounts("/etc/openpanel/ftp/users/" + userContext + "/users.list") {
		if acc.Username == username {
			return acc, true
		}
	}
	return Account{}, false
}

func runFTPQuota(r *http.Request, username, quota, currentUsername string) (string, bool) {
	out, err := exec.CommandContext(r.Context(), "opencli", "ftp-quota", username, quota, currentUsername).CombinedOutput()
	return strings.TrimSpace(string(out)), err == nil && strings.Contains(string(out), "Success:")
}

// handleChangeFTPQuota handles both the change-quota form page and its submission for one FTP account
func handleChangeFTPQuota(a *appctx.App, w http.ResponseWriter, r *http.Request, username string) {
	if !isValidUsername(username) {
		flashAndRedirectToAccounts(a, w, r, "error", "Invalid FTP username.")
		return
	}
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !requireFTPRunning(a, w, r) {
		return
	}
	account, ok := findAccount(userContext, username)
	if !ok {
		flashAndRedirectToAccounts(a, w, r, "error", "FTP account not found.")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		quota, valid := normalizeQuota(r.Form.Get("quota"), r.Form.Get("quota_unit"))
		if !valid {
			flashAndRedirectToAccounts(a, w, r, "error", "Invalid quota, use a whole number of MB or GB.")
			return
		}
		out, success := runFTPQuota(r, username, quota, currentUsername)
		if !success {
			flashAndRedirectToAccounts(a, w, r, "error", web.Tr(a, r, "Failed to change FTP quota for user %(username)s: %(output)s", "username", username, "output", out))
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "set quota "+quota+" for FTP account "+username, reqip.ClientIP(r))
		if quota == "unlimited" {
			flashAndRedirectToAccounts(a, w, r, "success", web.Tr(a, r, "FTP quota removed for user %(username)s", "username", username))
		} else {
			flashAndRedirectToAccounts(a, w, r, "success", web.Tr(a, r, "FTP quota changed successfully for user %(username)s", "username", username))
		}
		return
	}

	renderFTPQuotaPage(a, w, r, account)
}

// apiFTPQuota sets or removes an FTP account's disk quota.
func apiFTPQuota(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !apiFTPServiceCheck(w, r) {
		return
	}
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username := r.PathValue("username")
	if !isValidUsername(username) {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid FTP username"})
		return
	}

	var body struct {
		Quota json.RawMessage `json:"quota"`
	}
	raw := ""
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr == nil {
		raw = rawQuotaString(body.Quota)
	} else {
		_ = r.ParseForm()
		raw = r.Form.Get("quota")
	}
	quota, valid := normalizeQuota(raw, "M")
	if !valid {
		writeAPIFTPJSON(w, http.StatusBadRequest, map[string]string{"error": "quota must be a number of MB, a value like 500M or 2G, or 'unlimited'"})
		return
	}

	out, success := runFTPQuota(r, username, quota, currentUsername)
	if !success {
		writeAPIFTPJSON(w, http.StatusInternalServerError, map[string]string{"error": out})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "set quota "+quota+" for FTP account "+username, reqip.ClientIP(r))
	writeAPIFTPJSON(w, http.StatusOK, map[string]string{"message": out, "quota": quota})
}
