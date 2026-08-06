package emails

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// ImportRow is one parsed row of an uploaded email-import file.
type ImportRow struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Quota       string `json:"quota"`
	Domain      string `json:"domain"`
	DomainValid bool   `json:"domain_valid"`
	EmailExists bool   `json:"email_exists"`
	IsValid     bool   `json:"is_valid"`
}

// csvSeparatorRE splits each row on either a comma or a semicolon, since
// uploaded files use either as a field delimiter.
var csvSeparatorRE = regexp.MustCompile(`[;,]`)

// parseImportCSV reads an uploaded email-import file, discarding the
// header row and keeping at most the first 3 columns (email, password,
// quota) of each remaining row.
func parseImportCSV(r io.Reader) ([][3]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var rows [][3]string
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if first {
			first = false
			continue // header row
		}
		fields := csvSeparatorRE.Split(line, -1)
		var row [3]string
		for i := 0; i < 3 && i < len(fields); i++ {
			row[i] = strings.TrimSpace(fields[i])
		}
		rows = append(rows, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

// handleImportEmails validates an uploaded CSV of mailboxes and stages the
// valid/invalid rows for confirmation before any accounts are created.
func handleImportEmails(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderImportPage(a, w, r)
		return
	}

	ctx := r.Context()
	userID, _ := auth.UserID(r)

	file, header, err := r.FormFile("file")
	if err != nil {
		flashAndRedirect(a, w, r, "error", "Error: No file uploaded.", "/emails/import")
		return
	}
	defer file.Close()

	filename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(filename, ".csv") {
		// .xls/.xlsx are intentionally unsupported (no Excel parsing
		// dependency pulled in for one upload path) - reported the same
		// way as any other unrecognized extension.
		flashAndRedirect(a, w, r, "error", `Error: Unsupported file format, please upload a .csv or .xls file.`, "/emails/import")
		return
	}

	rawRows, err := parseImportCSV(file)
	if err != nil {
		flashAndRedirect(a, w, r, "error", "Error reading file: "+err.Error(), "/emails/import")
		return
	}

	if len(rawRows) == 0 {
		flashAndRedirect(a, w, r, "error", "Error: File must have at least 2 columns: email, password", "/emails/import")
		return
	}

	for _, row := range rawRows {
		if !isValidEmail(row[0]) {
			flashAndRedirect(a, w, r, "error", "Error: File is invalid - the first column must contain valid email addresses.", "/emails/import")
			return
		}
	}

	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)
	// existingEmails is the set of already-registered bare addresses,
	// extracted from the formatted "* addr ( used / quota ) [pct%]" list
	// lines so EmailExists actually matches against a plain address
	// rather than the whole formatted line.
	existingEmails := make(map[string]bool)
	for _, addr := range addressesOf(GetEmailList(ctx, a, userID, currentUsername, userDomains)) {
		existingEmails[addr] = true
	}

	var validRows, invalidRows []ImportRow
	for _, raw := range rawRows {
		_, domain, _ := strings.Cut(raw[0], "@")
		row := ImportRow{
			Email: raw[0], Password: raw[1], Quota: raw[2], Domain: domain,
			DomainValid: a.CheckDomainBelongsToUser(ctx, userID, domain),
			EmailExists: existingEmails[raw[0]],
		}
		row.IsValid = row.DomainValid && !row.EmailExists
		if row.IsValid {
			validRows = append(validRows, row)
		} else {
			invalidRows = append(invalidRows, row)
		}
	}

	importToken := randomURLToken(24)
	payload, _ := json.Marshal(validRows)
	_ = a.Cache.Raw().Set(ctx, fmt.Sprintf("email_import:%d:%s", userID, importToken), payload, 10*time.Minute).Err()

	sess, _ := a.Sessions.Get(r, session.CookieName)
	sess.Values["import_token"] = importToken
	_ = a.Sessions.Save(r, w, sess)

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"valid_users": validRows, "invalid_users": invalidRows})
		return
	}

	renderConfirmImportPage(a, w, r, validRows, invalidRows)
}

// handleConfirmEmailImport creates the previously staged mailboxes one by
// one, streaming progress to the client as plain text.
func handleConfirmEmailImport(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	sess, _ := a.Sessions.Get(r, session.CookieName)
	importToken, _ := sess.Values["import_token"].(string)
	delete(sess.Values, "import_token")
	_ = a.Sessions.Save(r, w, sess)

	if importToken == "" {
		flashAndRedirect(a, w, r, "error", "Error: No import data found, please try again.", "/emails/import")
		return
	}

	cacheKey := fmt.Sprintf("email_import:%d:%s", userID, importToken)
	payload, getErr := a.Cache.Raw().Get(ctx, cacheKey).Bytes()
	_ = a.Cache.Raw().Del(ctx, cacheKey).Err()
	if getErr != nil {
		flashAndRedirect(a, w, r, "error", "Error: No import data found, please try again.", "/emails/import")
		return
	}
	var users []ImportRow
	if err := json.Unmarshal(payload, &users); err != nil || len(users) == 0 {
		flashAndRedirect(a, w, r, "error", "Error: No import data found, please try again.", "/emails/import")
		return
	}

	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	_, maxEmailQuota, _ := a.QueryPlanEmailMailboxLimits(ctx, planID)
	ipAddress := reqip.ClientIP(r)

	w.Header().Set("Content-Type", "text/plain")
	flusher, canFlush := w.(http.Flusher)

	writeLine := func(format string, args ...any) {
		fmt.Fprintf(w, format, args...)
		if canFlush {
			flusher.Flush()
		}
	}

	total := len(users)
	validCount := 0

	for idx, user := range users {
		email := user.Email
		password := user.Password
		quota := user.Quota
		_, domain, _ := strings.Cut(email, "@")

		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			writeLine("[%d/%d] Skipping %s: domain '%s' does not exist.\n", idx+1, total, email, domain)
			continue
		}

		validCount++
		writeLine("[%d/%d] Creating: %s...\n", idx+1, total, email)

		out, cmdErr := exec.CommandContext(ctx, "opencli", "email-setup", "email", "add", email, password, "--wait").CombinedOutput()
		if cmdErr != nil {
			errorMessage := strings.TrimSpace(string(out))
			if strings.Contains(errorMessage, "Supplied non-number argument") || strings.Contains(errorMessage, "User doesn't exist") {
				writeLine("%s\n", errorMessage)
			} else {
				writeLine("Failed to add email %s: %s\n", email, errorMessage)
			}
		}

		if quota != "" {
			quotaStr := strings.TrimSpace(quota)
			m := importQuotaRE.FindStringSubmatch(quotaStr)
			if m == nil {
				writeLine("Invalid quota format for %s: '%s'\n", email, quotaStr)
			} else {
				amount, unit := m[1], m[2]
				if amount == "0" {
					writeLine("[%d/%d] Skipping quota for %s (0 specified)\n", idx+1, total, email)
					writeLine("\n")
					continue
				}
				if unit == "" {
					unit = "M"
				} else {
					unit = strings.ToUpper(unit)
				}
				requestedQuota := amount + unit

				if strings.ToUpper(unit) != "B" && unit != "K" && unit != "M" && unit != "G" && unit != "T" {
					writeLine("Invalid quota unit for %s: '%s' (valid: B, K, M, G, T)\n", email, unit)
				} else {
					fullQuota := requestedQuota
					if maxEmailQuota != "0" {
						reqBytes, reqErr := quotaToBytes(requestedQuota)
						maxBytes, maxErr := quotaToBytes(maxEmailQuota)
						if reqErr != nil || maxErr != nil {
							return
						}
						if reqBytes > maxBytes {
							writeLine("Warning: Specified quota for %s is over the max_email_quota for the plan. Plan limit %s will be used instead.\n", email, maxEmailQuota)
							fullQuota = maxEmailQuota
						}
					}

					writeLine("[%d/%d] Setting quota for %s to %s...\n", idx+1, total, email, fullQuota)
					qOut, qErr := exec.CommandContext(ctx, "opencli", "email-setup", "quota", "set", email, fullQuota).CombinedOutput()
					if qErr != nil {
						writeLine("Quota set failed for %s: %s\n", email, strings.TrimSpace(string(qOut)))
					} else if len(qOut) > 0 {
						writeLine("%s", string(qOut))
					}
				}
			}
		}

		writeLine("\n")
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+strconv.Itoa(validCount)+" email accounts", ipAddress)
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	writeLine("Import complete. %d accounts created.\n", validCount)
}

var importQuotaRE = regexp.MustCompile(`(?i)^(\d+)([BKMGT])?$`)
