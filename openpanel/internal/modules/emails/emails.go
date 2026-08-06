// Package emails ports modules/emails.py plus its siblings
// (email_aliases.py, email_default.py, email_deliverability.py,
// email_export.py, email_filters.py, email_import.py) and modules/webmail.py:
// email account CRUD, quota management, aliases, catch-all addresses,
// SPF/DKIM/DMARC deliverability checks, CSV/XLSX export/import, Sieve
// filters, and webmail single-sign-on.
package emails

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// emailRegex mirrors EMAIL_REGEX.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(s string) bool { return emailRegex.MatchString(s) }

// baseMailPath mirrors BASE_MAIL_PATH, read once from admin.ini at package
// init (matching Python's module-load-time ConfigParser read).
var baseMailPath = readEmailStorageLocation()

const adminIniPath = "/etc/openpanel/openadmin/config/admin.ini"

// readEmailStorageLocation mirrors the module-level
// `config.get("EMAIL", "email_storage_location", fallback="/var/mail")`
// read - a minimal single-purpose INI reader since this is the only place
// in the port that needs admin.ini's [EMAIL] section.
func readEmailStorageLocation() string {
	const fallback = "/var/mail"
	data, err := os.ReadFile(adminIniPath)
	if err != nil {
		return fallback
	}
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if section != "EMAIL" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "email_storage_location" {
			return strings.TrimSpace(line[idx+1:])
		}
	}
	return fallback
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// randomURLToken mirrors secrets.token_urlsafe(nBytes).
func randomURLToken(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = cryptorand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// email list cache (mirrors emails.py's get_email_list/import_user_emails)
// ---------------------------------------------------------------------------

func emailsCacheKey(username string) string { return "email_list:" + username }

func cachedEmailsFile(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/emails.yml"
}

// readEmailsFile mirrors _read_emails_file(): lines starting with '*' whose
// domain (the part after '@' up to the next space) is in userDomains.
func readEmailsFile(username string, userDomains map[string]bool) []string {
	data, err := os.ReadFile(cachedEmailsFile(username))
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "*") {
			continue
		}
		atIdx := strings.Index(line, "@")
		if atIdx == -1 {
			continue
		}
		rest := line[atIdx+1:]
		domain := rest
		if sp := strings.Index(rest, " "); sp != -1 {
			domain = rest[:sp]
		}
		if userDomains[domain] {
			out = append(out, line)
		}
	}
	return out
}

// GetEmailList mirrors get_email_list(), cached 1h.
func GetEmailList(ctx context.Context, a *appctx.App, userID int, username string, userDomains map[string]bool) []string {
	result, _ := cache.Memoize(ctx, a.Cache, emailsCacheKey(username), time.Hour, func() ([]string, error) {
		path := cachedEmailsFile(username)
		info, statErr := os.Stat(path)
		if statErr != nil || info.Size() == 0 {
			domains, _ := a.AllDomainsForUser(ctx, userID)
			ImportUserEmails(username, domainSet(domains))
		}
		return readEmailsFile(username, userDomains), nil
	})
	return result
}

func domainSet(domains []appctx.Domain) map[string]bool {
	set := make(map[string]bool, len(domains))
	for _, d := range domains {
		set[d.DomainURL] = true
	}
	return set
}

// GetEmailCount mirrors get_email_count().
func GetEmailCount(ctx context.Context, a *appctx.App, userID int, username string, userDomains map[string]bool) int {
	return len(GetEmailList(ctx, a, userID, username, userDomains))
}

// InvalidateEmailCache mirrors invalidate_email_cache().
func InvalidateEmailCache(ctx context.Context, a *appctx.App, userID int, username string) {
	_ = a.Cache.Delete(ctx, emailsCacheKey(username))
	domains, _ := a.AllDomainsForUser(ctx, userID)
	ImportUserEmails(username, domainSet(domains))
}

// ImportUserEmails mirrors import_user_emails(): (re)writes the per-user
// on-disk email cache file from `opencli email-setup email list`.
func ImportUserEmails(currentUsername string, userDomains map[string]bool) {
	path := cachedEmailsFile(currentUsername)
	if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(strings.TrimSuffix(path, "/emails.yml"), 0o755)
		_ = os.WriteFile(path, nil, 0o644)
	}

	out, err := exec.Command("opencli", "email-setup", "email", "list").Output()
	if err != nil {
		return
	}

	var filtered []string
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "*") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		atIdx := strings.Index(fields[1], "@")
		if atIdx == -1 {
			continue
		}
		domain := fields[1][atIdx+1:]
		if userDomains[domain] {
			filtered = append(filtered, line)
		}
	}

	_ = os.MkdirAll(strings.TrimSuffix(path, "/emails.yml"), 0o755)
	_ = os.WriteFile(path, []byte(strings.Join(filtered, "\n")+"\n"), 0o644)
}

// ---------------------------------------------------------------------------
// quota helpers
// ---------------------------------------------------------------------------

var quotaUnitMultipliers = map[byte]float64{'B': 1, 'K': 1024, 'M': 1024 * 1024, 'G': 1024 * 1024 * 1024, 'T': 1024 * 1024 * 1024 * 1024}

// quotaToBytes mirrors quota_to_bytes().
func quotaToBytes(value string) (float64, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return 0, fmt.Errorf("invalid quota unit")
	}
	unit := value[len(value)-1]
	mult, ok := quotaUnitMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("invalid quota unit")
	}
	num, err := strconv.ParseFloat(value[:len(value)-1], 64)
	if err != nil {
		return 0, err
	}
	return num * mult, nil
}

// emailSetQuota mirrors email_set_quota(). Returns (message, ok); ok=false
// with an empty message means "invalid quota format" (Python's None
// sentinel).
func emailSetQuota(email, gb, format, maxEmailQuota string, maxEmailQuotaNumeric float64, allocatedUnit string) (string, bool) {
	if gb == "0" {
		_ = exec.Command("opencli", "email-setup", "quota", "del", email).Run()
		return "deleted quota", true
	}

	validFormats := map[string]bool{"B": true, "k": true, "M": true, "G": true, "T": true}
	if !validFormats[format] {
		return "", false
	}

	requestedQuota := gb + format

	if maxEmailQuota != "0" {
		reqBytes, reqErr := quotaToBytes(requestedQuota)
		maxBytes, maxErr := quotaToBytes(maxEmailQuota)
		if reqErr == nil && maxErr == nil {
			if reqBytes > maxBytes {
				requestedQuota = fmt.Sprintf("%s%s", trimFloat(maxEmailQuotaNumeric), allocatedUnit)
			}
		} else {
			requestedQuota = fmt.Sprintf("%s%s", trimFloat(maxEmailQuotaNumeric), allocatedUnit)
		}
	}

	_ = exec.Command("opencli", "email-setup", "quota", "set", email, requestedQuota).Run()
	return "set quota to " + requestedQuota, true
}

// trimFloat formats a float without a trailing ".0" when it's a whole
// number, matching Python's f-string interpolation of ints/floats
// (max_email_quota_numeric is often an int from _parse_max_quota).
func trimFloat(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

var maxQuotaRE = regexp.MustCompile(`(?i)^([\d.]+)([BKMGTP])`)

// parseMaxQuota mirrors _parse_max_quota().
func parseMaxQuota(maxEmailQuota string) (float64, string) {
	if maxEmailQuota == "0" {
		return 0, "T"
	}
	if m := maxQuotaRE.FindStringSubmatch(maxEmailQuota); m != nil {
		n, _ := strconv.ParseFloat(m[1], 64)
		return n, strings.ToUpper(m[2])
	}
	n, _ := strconv.ParseFloat(maxEmailQuota, 64)
	return n, "G"
}

// getDedicatedOrSharedIP mirrors _get_dedicated_or_shared_ip(): a
// standalone, uncached IP lookup distinct from
// appctx.GetCachedIPForUserOrPublicIPv4 (its own fresh curl call, not the
// shared 1h-cached fallback chain).
func getDedicatedOrSharedIP(ctx context.Context, currentUsername string) string {
	if ip, ok := appctx.ReadDedicatedIPFromFile(currentUsername); ok {
		return ip
	}
	out, err := exec.CommandContext(ctx, "curl", "-4", "https://ip.openpanel.com").Output()
	if err != nil {
		return "mail.example.com"
	}
	ipStr := strings.TrimSpace(string(out))
	if ipStr == "" {
		return "mail.example.com"
	}
	return ipStr
}

// ---------------------------------------------------------------------------
// NEW EMAIL
// ---------------------------------------------------------------------------

// handleEmailsNew mirrors emails_new().
func handleEmailsNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	emailLimitStr, maxEmailQuota, _ := a.QueryPlanEmailMailboxLimits(ctx, planID)
	maxEmailQuotaNumeric, allocatedUnit := parseMaxQuota(maxEmailQuota)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domain := r.Form.Get("domain")
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		gb := r.Form.Get("gb")
		format := r.Form.Get("format")

		var missing []string
		if domain == "" {
			missing = append(missing, "domain")
		}
		if username == "" {
			missing = append(missing, "username")
		}
		if password == "" {
			missing = append(missing, "password")
		}
		if len(missing) > 0 {
			sess, _ := a.Sessions.Get(r, session.CookieName)
			for _, field := range missing {
				flash.Add(sess, "error", "Error: "+field+" not provided.")
			}
			_ = a.Sessions.Save(r, w, sess)
			http.Redirect(w, r, "/emails/new", http.StatusFound)
			return
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
		if !validators.IsPasswordStrongEnough(password, threshold) {
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/emails/new")
			return
		}

		emailLimit, _ := strconv.Atoi(emailLimitStr)
		if emailLimit != 0 && GetEmailCount(ctx, a, userID, currentUsername, userDomains) >= emailLimit {
			flashAndRedirect(a, w, r, "error", "Error: reached max number of email accounts on the hosting plan.", "/emails/new")
			return
		}

		email := username + "@" + domain
		args := []string{"email-setup", "email", "add", email, password}
		if maxEmailQuota != "0" {
			args = append(args, "--wait")
		}

		cmd := exec.CommandContext(ctx, "opencli", args...)
		var stderrBuf strings.Builder
		cmd.Stderr = &stderrBuf
		stdout, cmdErr := cmd.Output()
		// mirrors emails_new(): a zero-exit run with non-empty stdout is
		// ALSO treated as a failure (opencli's success case prints nothing).
		if cmdErr != nil || strings.TrimSpace(string(stdout)) != "" {
			stderrStr := strings.TrimSpace(stderrBuf.String())
			if !strings.Contains(stderrStr, "Supplied non-number argument") && !strings.Contains(stderrStr, "User doesn't exist") {
				msg := stderrStr
				if msg == "" {
					msg = "command failed"
				}
				flashAndRedirect(a, w, r, "error", "Failed to add email "+email+": "+msg, "/emails/new")
				return
			}
		}

		if gb != "" {
			msg, ok := emailSetQuota(email, gb, format, maxEmailQuota, maxEmailQuotaNumeric, allocatedUnit)
			if !ok {
				http.Error(w, "ERROR: Invalid quota format. e.g. (B (byte), k (kibibyte), M (mebibyte), G (gibibyte) or T (tebibyte))", http.StatusBadRequest)
				return
			}
			_ = msg
		} else if maxEmailQuota != "0" {
			_ = exec.CommandContext(ctx, "opencli", "email-setup", "quota", "set", email, fmt.Sprintf("%s%s", trimFloat(maxEmailQuotaNumeric), allocatedUnit)).Run()
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created email "+email, ipAddress)
		InvalidateEmailCache(ctx, a, userID, currentUsername)
		flashAndRedirect(a, w, r, "success", "Email "+email+" added successfully.", "/emails")
		return
	}

	renderNewEmailPage(a, w, r, domains, maxEmailQuotaNumeric, allocatedUnit)
}

// ---------------------------------------------------------------------------
// quota-line parsing (mirrors accounts.html's and single_account.html's
// inline Jinja parsing of one `opencli email-setup email list` line, e.g.
// "* info@demo.rs ( 0 / 2.0G ) [0%]")
// ---------------------------------------------------------------------------

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// EmailListRow is one row of emails/accounts.html's table.
type EmailListRow struct {
	Address      string
	QuotaDisplay string
	PercentVal   int
	CappedVal    int
	BarColor     string
}

// parseEmailListRow mirrors accounts.html's per-row quota parsing.
func parseEmailListRow(entry string) EmailListRow {
	parts := strings.Split(entry, " ")
	address := ""
	if len(parts) > 1 {
		address = parts[1]
	}
	quota := ""
	if len(parts) > 2 {
		quota = strings.Join(parts[2:], " ")
	}

	quotaParts := strings.Split(quota, "[")
	percentageStr := "0%"
	if len(quotaParts) > 1 {
		percentageStr = strings.Split(quotaParts[1], "]")[0]
	}

	rawPercent := "0%"
	if percentageStr != "" && isAllDigits(strings.Trim(percentageStr, "%")) {
		rawPercent = percentageStr
	}
	percentVal, _ := strconv.Atoi(strings.Trim(rawPercent, "%"))

	cappedVal := percentVal
	if cappedVal > 100 {
		cappedVal = 100
	}

	barColor := "bg-green-500"
	switch {
	case percentVal <= 80:
		barColor = "bg-green-500"
	case percentVal <= 90:
		barColor = "bg-orange-500"
	default:
		barColor = "bg-red-500"
	}

	return EmailListRow{Address: address, QuotaDisplay: quota, PercentVal: percentVal, CappedVal: cappedVal, BarColor: barColor}
}

// SingleEmailQuota is single_account.html's parsed quota-line view-model.
type SingleEmailQuota struct {
	Address        string
	QuotaDisplay   string
	PercentVal     int
	CappedVal      int
	AllocatedQuota string
	AllocatedValue string
}

// parseSingleEmailQuota mirrors single_account.html's inline parsing.
func parseSingleEmailQuota(entry string) SingleEmailQuota {
	parts := strings.Split(entry, " ")
	address := ""
	if len(parts) > 1 {
		address = parts[1]
	}
	quota := ""
	if len(parts) > 2 {
		quota = strings.Join(parts[2:], " ")
	}

	quotaParts := strings.Split(quota, "[")
	percentageStr := "0"
	if len(quotaParts) > 1 {
		percentageStr = strings.Split(quotaParts[1], "]")[0]
	}
	percentVal, _ := strconv.Atoi(strings.Trim(percentageStr, "%"))
	cappedVal := percentVal
	if cappedVal > 100 {
		cappedVal = 100
	}

	allocatedQuota := "0"
	allocatedValue := ""
	quotaMatch := strings.Split(strings.Trim(quotaParts[0], "() "), "/")
	if len(quotaMatch) > 1 {
		allocatedQuotaWithUnit := strings.TrimSpace(quotaMatch[1])
		if allocatedQuotaWithUnit != "" {
			allocatedQuota = allocatedQuotaWithUnit[:len(allocatedQuotaWithUnit)-1]
			allocatedValue = allocatedQuotaWithUnit[len(allocatedQuotaWithUnit)-1:]
		}
	}

	return SingleEmailQuota{
		Address: address, QuotaDisplay: quota, PercentVal: percentVal, CappedVal: cappedVal,
		AllocatedQuota: allocatedQuota, AllocatedValue: allocatedValue,
	}
}

// ---------------------------------------------------------------------------
// SINGLE EMAIL: LIST, EDIT, DELETE
// ---------------------------------------------------------------------------

// handleEmails mirrors emails(): GET /emails, GET/POST/DELETE /emails and
// /emails/edit/{email} (email is "" when hit at the bare /emails path,
// matching Python's default parameter).
func handleEmails(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	email := r.PathValue("email")
	if email != "" {
		handleSingleEmail(a, w, r, email, userID, currentUsername, userDomains)
		return
	}

	currentEmailsList := GetEmailList(ctx, a, userID, currentUsername, userDomains)
	renderAccountsPage(a, w, r, currentEmailsList, domains)
}

func handleSingleEmail(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string, userDomains map[string]bool) {
	ctx := r.Context()

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	_, domain, _ := strings.Cut(email, "@")
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	_, maxEmailQuota, _ := a.QueryPlanEmailMailboxLimits(ctx, planID)
	maxEmailQuotaNumeric, allocatedUnit := parseMaxQuota(maxEmailQuota)

	switch r.Method {
	case http.MethodGet:
		getSingleEmail(a, w, r, email, userID, currentUsername, userDomains, maxEmailQuotaNumeric, allocatedUnit)
	case http.MethodPost:
		postSingleEmail(a, w, r, email, userID, currentUsername, maxEmailQuota, maxEmailQuotaNumeric, allocatedUnit)
	case http.MethodDelete:
		deleteSingleEmail(a, w, r, email, userID, currentUsername)
	}
}

func getEmailRestriction(direction, email string) string {
	out, err := exec.Command("opencli", "email-setup", "email", "restrict", "list", direction).CombinedOutput()
	if err != nil {
		return "Error: " + strings.TrimSpace(string(out))
	}
	if strings.Contains(string(out), email) {
		return "REJECT"
	}
	return "ACCEPT"
}

func getSingleEmail(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string, userDomains map[string]bool, maxEmailQuotaNumeric float64, allocatedUnit string) {
	ctx := r.Context()
	sendRestriction := getEmailRestriction("send", email)
	receiveRestriction := getEmailRestriction("receive", email)

	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
	dedicatedIP := getDedicatedOrSharedIP(ctx, currentUsername)

	emailLines := GetEmailList(ctx, a, userID, currentUsername, userDomains)
	currentEmailsList := ""
	for _, line := range emailLines {
		if strings.Contains(line, email) {
			currentEmailsList = line
			break
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"title": email, "current_emails_list": currentEmailsList,
			"max_email_quota_numeric": maxEmailQuotaNumeric, "allocated_unit": allocatedUnit,
			"server_ip": serverIP, "dedicated_ip": dedicatedIP,
			"send_restriction": sendRestriction, "receive_restriction": receiveRestriction,
		})
		return
	}

	renderSingleAccountPage(a, w, r, currentEmailsList, maxEmailQuotaNumeric, allocatedUnit, serverIP, dedicatedIP, sendRestriction, receiveRestriction)
}

func postSingleEmail(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername, maxEmailQuota string, maxEmailQuotaNumeric float64, allocatedUnit string) {
	ctx := r.Context()
	_ = r.ParseForm()
	gb := r.Form.Get("gb")
	format := r.Form.Get("format")
	incoming := r.Form.Get("incoming")
	outgoing := r.Form.Get("outgoing")
	emailPassword := r.Form.Get("email_password")

	var actionsTaken []string

	if gb != "" {
		msg, ok := emailSetQuota(email, gb, format, maxEmailQuota, maxEmailQuotaNumeric, allocatedUnit)
		if !ok {
			http.Error(w, "ERROR: Invalid quota format. e.g. (B (byte), k (kibibyte), M (mebibyte), G (gibibyte) or T (tebibyte))", http.StatusBadRequest)
			return
		}
		actionsTaken = append(actionsTaken, msg)
	}

	type restrictionCmd struct {
		field, value, verb, direction, logMsg string
	}
	for _, c := range []restrictionCmd{
		{"incoming", "suspend", "add", "receive", "suspended incoming emails"},
		{"incoming", "allow", "del", "receive", "unsuspended incoming emails"},
		{"outgoing", "suspend", "add", "send", "suspended outgoing emails"},
		{"outgoing", "allow", "del", "send", "unsuspended outgoing emails"},
	} {
		formVal := incoming
		if c.field == "outgoing" {
			formVal = outgoing
		}
		if formVal == c.value {
			_ = exec.CommandContext(ctx, "opencli", "email-setup", "email", "restrict", c.verb, c.direction, email).Run()
			actionsTaken = append(actionsTaken, c.logMsg)
		}
	}

	if emailPassword != "" {
		threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
		if !validators.IsPasswordStrongEnough(emailPassword, threshold) {
			http.Error(w, "ERROR: Password does not meet the required strength.", http.StatusBadRequest)
			return
		}
		_ = exec.CommandContext(ctx, "opencli", "email-setup", "email", "update", email, emailPassword).Run()
		actionsTaken = append(actionsTaken, "updated password")
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, strings.Join(actionsTaken, "; ")+" for email "+email, ipAddress)
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	flashAndRedirect(a, w, r, "success", "Settings saved for email "+email, r.URL.Path)
}

func deleteSingleEmail(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string) {
	ctx := r.Context()

	var body struct {
		Email string `json:"email"`
	}
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.Email != "" {
			email = body.Email
		}
	}

	out, err := exec.CommandContext(ctx, "opencli", "email-setup", "email", "del", email).CombinedOutput()
	if err == nil {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "deleted email "+email, ipAddress)
		InvalidateEmailCache(ctx, a, userID, currentUsername)
		flashAndRedirect(a, w, r, "success", "Email "+email+" deleted successfully.", "/emails")
		return
	}
	flashAndRedirect(a, w, r, "error", "ERROR: Failed to delete email "+email+": "+strings.TrimSpace(string(out)), "/emails")
}

// ---------------------------------------------------------------------------
// DELETE (select-then-confirm page)
// ---------------------------------------------------------------------------

// handleEmailsDelete mirrors emails_delete().
func handleEmailsDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	address := r.PathValue("address")
	var currentEmailsList []string
	if address == "" {
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentEmailsList = GetEmailList(ctx, a, userID, currentUsername, domainSet(domains))
	} else if !isValidEmail(address) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	renderDeletePage(a, w, r, address, currentEmailsList)
}

// ---------------------------------------------------------------------------
// SERVER INFO
// ---------------------------------------------------------------------------

// handleEmailsServerInfo mirrors emails_server_info().
func handleEmailsServerInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")
	if !isValidEmail(address) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if a.ForceDomain != "" && sysinfo.HasSSL(r.Context(), a.Cache, a.ForceDomain) {
		scheme = "https"
	}

	renderInfoPage(a, w, r, address, scheme, a.ForceDomain)
}

// ---------------------------------------------------------------------------
// CONNECT DEVICES
// ---------------------------------------------------------------------------

// handleEmailConfiguration mirrors email_configuration().
func handleEmailConfiguration(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	account := r.PathValue("account")
	if !isValidEmail(account) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	configType := strings.ToLower(r.PathValue("type"))
	if configType != "thunderbird" && configType != "outlook" && configType != "apple" {
		flashAndRedirect(a, w, r, "error", "Error: Invalid configuration type.", "/emails")
		return
	}

	sslParam := strings.ToLower(r.URL.Query().Get("ssl"))
	if sslParam == "" {
		sslParam = "true"
	}
	sslEnabled := sslParam == "1" || sslParam == "true" || sslParam == "yes"

	var incomingPort, outgoingPort int
	var encryption, sslTrueFalse string
	if sslEnabled {
		incomingPort, outgoingPort, encryption, sslTrueFalse = 993, 465, "SSL/TLS", "<true />"
	} else {
		incomingPort, outgoingPort, encryption, sslTrueFalse = 143, 587, "STARTTLS", "<false />"
	}

	mailHost := getDedicatedOrSharedIP(ctx, currentUsername)
	username, domain, _ := strings.Cut(account, "@")

	var content, filename, mimetype string

	switch configType {
	case "thunderbird":
		content = fmt.Sprintf(
			"[InternetEmail]\nEmailAddress=%s\nUserName=%s\nIncomingServer=%s\nIncomingProtocol=IMAP\nIncomingPort=%d\nIncomingSecurity=%s\nOutgoingServer=%s\nOutgoingPort=%d\nOutgoingSecurity=%s\nAuthentication=normal password\nDescription=Thunderbird Profile\nNote=Password will be requested on first login.\n",
			account, account, mailHost, incomingPort, encryption, mailHost, outgoingPort, encryption,
		)
		filename = username + "_thunderbird_profile.txt"
		mimetype = "text/plain"

	case "outlook":
		content = fmt.Sprintf(
			"Outlook Email Configuration:\n\nEmail Address: %s\nUsername: %s\n\nIncoming Server (IMAP): %s\nPort: %d\nEncryption: %s\n\nOutgoing Server (SMTP): %s\nPort: %d\nEncryption: %s\n\nAuthentication: Normal Password\nNote: Password is not stored. You will be prompted on first send/receive.\n",
			account, account, mailHost, incomingPort, encryption, mailHost, outgoingPort, encryption,
		)
		filename = username + "_outlook_profile.txt"
		mimetype = "text/plain"

	default: // apple
		payloadUUID := uuid.New().String()
		innerUUID := uuid.New().String()
		outerUUID := uuid.New().String()
		payloadIdentifierBase := mailHost + "." + domain
		description := account + " Secure Email Setup"
		content = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>PayloadContent</key>
        <array>
            <dict>
                <key>EmailAccountDescription</key>
                <string>%s</string>
                <key>EmailAccountName</key>
                <string>%s</string>
                <key>EmailAccountType</key>
                <string>EmailTypeIMAP</string>
                <key>EmailAddress</key>
                <string>%s</string>
                <key>IncomingMailServerAuthentication</key>
                <string>EmailAuthPassword</string>
                <key>IncomingMailServerHostName</key>
                <string>%s</string>
                <key>IncomingMailServerPortNumber</key>
                <integer>%d</integer>
                <key>IncomingMailServerUseSSL</key>
                %s
                <key>IncomingMailServerUsername</key>
                <string>%s</string>
                <key>OutgoingMailServerAuthentication</key>
                <string>EmailAuthPassword</string>
                <key>OutgoingMailServerHostName</key>
                <string>%s</string>
                <key>OutgoingMailServerPortNumber</key>
                <integer>%d</integer>
                <key>OutgoingMailServerUseSSL</key>
                %s
                <key>OutgoingMailServerUsername</key>
                <string>%s</string>
                <key>OutgoingPasswordSameAsIncomingPassword</key>
                <true />
                <key>PayloadDescription</key>
                <string>%s</string>
                <key>PayloadDisplayName</key>
                <string>%s</string>
                <key>PayloadIdentifier</key>
                <string>%s.%s</string>
                <key>PayloadOrganization</key>
                <string>%s</string>
                <key>PayloadType</key>
                <string>com.apple.mail.managed</string>
                <key>PayloadUUID</key>
                <string>%s</string>
                <key>PayloadVersion</key>
                <integer>1</integer>
                <key>IncomingMailServerIMAPPathPrefix</key>
                <string>INBOX</string>
            </dict>
        </array>
        <key>PayloadDescription</key>
        <string>%s</string>
        <key>PayloadDisplayName</key>
        <string>%s</string>
        <key>PayloadIdentifier</key>
        <string>%s.%s</string>
        <key>PayloadOrganization</key>
        <string>%s</string>
        <key>PayloadType</key>
        <string>Configuration</string>
        <key>PayloadUUID</key>
        <string>%s</string>
        <key>PayloadVersion</key>
        <integer>1</integer>
    </dict>
</plist>`,
			account, username, account, mailHost, incomingPort, sslTrueFalse, account,
			mailHost, outgoingPort, sslTrueFalse, account,
			description, description, payloadIdentifierBase, innerUUID, domain, innerUUID,
			description, description, payloadIdentifierBase, outerUUID, domain, payloadUUID,
		)
		filename = username + "_apple_mail.mobileconfig"
		mimetype = "application/x-apple-aspen-config"
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded "+configType+" configuration for email "+username, ipAddress)

	w.Header().Set("Content-Type", mimetype)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte(content))
}
