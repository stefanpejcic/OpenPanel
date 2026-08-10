package emails

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterEmailImportAPI wires the email import API routes onto mux, gated
// behind the same "email_import" feature as the web UI's /emails/import.
func RegisterEmailImportAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "email_import", "POST /api/emails/import", func(w http.ResponseWriter, r *http.Request) { apiEmailImportPreview(a, w, r) })
	apiregistry.Handle(mux, a, "email_import", "POST /api/emails/import/confirm", func(w http.ResponseWriter, r *http.Request) { apiEmailImportConfirm(a, w, r) })
}

// apiImportCacheKey mirrors the cache key format handleImportEmails /
// handleConfirmEmailImport use, so a staged import can be confirmed
// regardless of whether it was staged through the web UI or the API.
func apiImportCacheKey(userID int, token string) string {
	return fmt.Sprintf("email_import:%d:%s", userID, token)
}

// apiEmailImportPreview validates an uploaded CSV of mailboxes (email,
// password, quota columns) and stages the valid/invalid rows for
// confirmation before any accounts are created. Unlike the web UI, which
// keeps the staged import keyed off a token in the session cookie, the API
// is stateless (bearer-token auth, no session) so the token is returned in
// the response body instead - the caller passes it back to
// /api/emails/import/confirm.
func apiEmailImportPreview(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "No file uploaded (expected multipart field \"file\")."})
		return
	}
	defer file.Close()

	filename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(filename, ".csv") {
		// .xls/.xlsx are intentionally unsupported, same as the web UI.
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported file format, please upload a .csv file."})
		return
	}

	rawRows, parseErr := parseImportCSV(file)
	if parseErr != nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Error reading file: " + parseErr.Error()})
		return
	}
	if len(rawRows) == 0 {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "File must have at least 2 columns: email, password"})
		return
	}
	for _, row := range rawRows {
		if !isValidEmail(row[0]) {
			writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "File is invalid - the first column must contain valid email addresses."})
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
	_ = a.Cache.Raw().Set(ctx, apiImportCacheKey(userID, importToken), payload, 10*time.Minute).Err()

	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{
		"import_token":  importToken,
		"valid_users":   validRows,
		"invalid_users": invalidRows,
	})
}

// apiImportResult is the per-row outcome reported by apiEmailImportConfirm.
type apiImportResult struct {
	Email  string `json:"email"`
	Status string `json:"status"` // "created", "skipped", or "error"
	Detail string `json:"detail,omitempty"`
}

// apiEmailImportConfirm creates the mailboxes staged by a prior call to
// apiEmailImportPreview.
func apiEmailImportConfirm(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	var body struct {
		ImportToken string `json:"import_token"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.ImportToken = r.Form.Get("import_token")
	}
	importToken := strings.TrimSpace(body.ImportToken)
	if importToken == "" {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "import_token is required"})
		return
	}

	cacheKey := apiImportCacheKey(userID, importToken)
	payload, getErr := a.Cache.Raw().Get(ctx, cacheKey).Bytes()
	_ = a.Cache.Raw().Del(ctx, cacheKey).Err()
	if getErr != nil {
		writeAPIEmailsJSON(w, http.StatusNotFound, map[string]string{"error": "No import data found for this token, please try again."})
		return
	}
	var users []ImportRow
	if err := json.Unmarshal(payload, &users); err != nil || len(users) == 0 {
		writeAPIEmailsJSON(w, http.StatusNotFound, map[string]string{"error": "No import data found for this token, please try again."})
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

	results := make([]apiImportResult, 0, len(users))
	created := 0

	for _, user := range users {
		email := user.Email
		password := user.Password
		quota := user.Quota
		_, domain, _ := strings.Cut(email, "@")

		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			results = append(results, apiImportResult{Email: email, Status: "skipped", Detail: "domain '" + domain + "' does not exist"})
			continue
		}

		out, cmdErr := exec.CommandContext(ctx, "opencli", "email-setup", "email", "add", email, password, "--wait").CombinedOutput()
		if cmdErr != nil {
			errorMessage := strings.TrimSpace(string(out))
			if strings.Contains(errorMessage, "Supplied non-number argument") || strings.Contains(errorMessage, "User doesn't exist") {
				// Benign quirk of the underlying CLI; account was still created.
			} else {
				results = append(results, apiImportResult{Email: email, Status: "error", Detail: "Failed to add email: " + errorMessage})
				continue
			}
		}
		created++
		detail := ""

		if quota != "" {
			quotaStr := strings.TrimSpace(quota)
			m := importQuotaRE.FindStringSubmatch(quotaStr)
			if m == nil {
				detail = "invalid quota format '" + quotaStr + "'"
			} else {
				amount, unit := m[1], m[2]
				if amount == "0" {
					results = append(results, apiImportResult{Email: email, Status: "created", Detail: "quota skipped (0 specified)"})
					continue
				}
				if unit == "" {
					unit = "M"
				} else {
					unit = strings.ToUpper(unit)
				}
				requestedQuota := amount + unit

				if unit != "B" && unit != "K" && unit != "M" && unit != "G" && unit != "T" {
					detail = "invalid quota unit '" + unit + "' (valid: B, K, M, G, T)"
				} else {
					fullQuota := requestedQuota
					if maxEmailQuota != "0" {
						reqBytes, reqErr := quotaToBytes(requestedQuota)
						maxBytes, maxErr := quotaToBytes(maxEmailQuota)
						if reqErr != nil || maxErr != nil {
							results = append(results, apiImportResult{Email: email, Status: "error", Detail: "failed to evaluate quota limits"})
							continue
						}
						if reqBytes > maxBytes {
							fullQuota = maxEmailQuota
							detail = "quota capped to plan limit " + maxEmailQuota
						}
					}

					qOut, qErr := exec.CommandContext(ctx, "opencli", "email-setup", "quota", "set", email, fullQuota).CombinedOutput()
					if qErr != nil {
						detail = "quota set failed: " + strings.TrimSpace(string(qOut))
					}
				}
			}
		}

		results = append(results, apiImportResult{Email: email, Status: "created", Detail: detail})
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+strconv.Itoa(created)+" email accounts via API", reqip.ClientIP(r))
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{
		"message": "Import complete. " + strconv.Itoa(created) + " accounts created.",
		"created": created,
		"results": results,
	})
}
