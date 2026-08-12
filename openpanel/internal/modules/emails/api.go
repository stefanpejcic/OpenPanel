package emails

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sieveparser"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// RegisterEmailsAPI wires the emails API routes onto mux. Literal
// sub-paths (aliases, configuration, default, deliverability, filters) are
// registered as their own patterns alongside the generic
// "{email...}"/"{...}" catch-alls - Go's http.ServeMux picks the pattern
// with more literal segments (a proper subset of matches) over the
// wildcard, so the more specific routes always win.
func RegisterEmailsAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "emails", "GET /api/emails", func(w http.ResponseWriter, r *http.Request) { apiEmailsList(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "POST /api/emails", func(w http.ResponseWriter, r *http.Request) { apiEmailsCreate(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/configuration/{config_type}/{account...}", func(w http.ResponseWriter, r *http.Request) { apiEmailConfiguration(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/aliases", func(w http.ResponseWriter, r *http.Request) { apiEmailAliasesList(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "POST /api/emails/aliases", func(w http.ResponseWriter, r *http.Request) { apiEmailAliasesCreate(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/aliases/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailAliasDetailGet(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "POST /api/emails/aliases/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailAliasDetailPost(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "DELETE /api/emails/aliases/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailAliasDetailDelete(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/default/{domain}", func(w http.ResponseWriter, r *http.Request) { apiEmailDefaultGet(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "PUT /api/emails/default/{domain}", func(w http.ResponseWriter, r *http.Request) { apiEmailDefaultPut(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/deliverability", func(w http.ResponseWriter, r *http.Request) { apiEmailDeliverabilityAll(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/deliverability/{domain}", func(w http.ResponseWriter, r *http.Request) { apiEmailDeliverabilityDomain(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/filters/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailFilterGet(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "PUT /api/emails/filters/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailFilterPut(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "GET /api/emails/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailDetailGet(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "PATCH /api/emails/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailDetailPatch(a, w, r) })
	apiregistry.Handle(mux, a, "emails", "DELETE /api/emails/{email...}", func(w http.ResponseWriter, r *http.Request) { apiEmailDetailDelete(a, w, r) })
}

func writeAPIEmailsJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiOwnDomainOr403(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, domain string) bool {
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeAPIEmailsJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain."})
		return false
	}
	return true
}

// apiOwnEmailOr403 validates email's format and writes a 403 if the
// current user doesn't own its domain.
func apiOwnEmailOr403(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, email string) bool {
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format."})
		return false
	}
	_, domain, _ := strings.Cut(email, "@")
	return apiOwnDomainOr403(a, w, r, userID, domain)
}

// ---------------------------------------------------------------------------
// Mailboxes
// ---------------------------------------------------------------------------

type apiMailboxEntry struct {
	Address    string `json:"address"`
	QuotaLimit string `json:"quota_limit,omitempty"`
	QuotaUsed  string `json:"quota_used,omitempty"`
}

// apiEmailsList returns the mailboxes owned by the current user.
func apiEmailsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	lines := GetEmailList(ctx, a, userID, currentUsername, userDomains)
	entries := make([]apiMailboxEntry, 0, len(lines))
	for _, line := range lines {
		parts := strings.Fields(strings.TrimLeft(line, "* "))
		if len(parts) == 0 {
			continue
		}
		e := apiMailboxEntry{Address: parts[0]}
		if len(parts) >= 2 {
			e.QuotaLimit = parts[1]
		}
		if len(parts) >= 3 {
			e.QuotaUsed = parts[2]
		}
		entries = append(entries, e)
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"emails": entries})
}

// apiEmailsCreate creates a new mailbox for the current user.
func apiEmailsCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	var body struct {
		Domain   string `json:"domain"`
		Username string `json:"username"`
		Password string `json:"password"`
		GB       string `json:"gb"`
		Format   string `json:"format"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Domain = r.Form.Get("domain")
		body.Username = r.Form.Get("username")
		body.Password = r.Form.Get("password")
		body.GB = r.Form.Get("gb")
		body.Format = r.Form.Get("format")
	}
	domain := strings.TrimSpace(body.Domain)
	username := strings.TrimSpace(body.Username)
	password := strings.TrimSpace(body.Password)
	gb := strings.TrimSpace(body.GB)
	format := strings.TrimSpace(body.Format)
	if format == "" {
		format = "G"
	}

	if domain == "" || username == "" || password == "" {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "domain, username, and password are required"})
		return
	}
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	if !validators.IsValidEmailUsername(username) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "username can only contain letters, numbers, and . _ % + - (no @)"})
		return
	}

	emailLimitStr, maxEmailQuota, _ := a.QueryPlanEmailMailboxLimits(ctx, planID)
	maxEmailQuotaNumeric, allocatedUnit := parseMaxQuota(maxEmailQuota)
	emailLimit, _ := strconv.Atoi(emailLimitStr)

	if emailLimit != 0 && GetEmailCount(ctx, a, userID, currentUsername, userDomains) >= emailLimit {
		writeAPIEmailsJSON(w, http.StatusForbidden, map[string]string{"error": "Email account limit reached for your plan"})
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
	if cmdErr != nil || strings.TrimSpace(string(stdout)) != "" {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if !strings.Contains(stderrStr, "Supplied non-number argument") && !strings.Contains(stderrStr, "User doesn't exist") {
			msg := stderrStr
			if msg == "" {
				msg = "command failed"
			}
			writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create email: " + msg})
			return
		}
	}

	if gb != "" {
		_, ok := emailSetQuota(email, gb, format, maxEmailQuota, maxEmailQuotaNumeric, allocatedUnit)
		if !ok {
			writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid quota format (B, k, M, G, T)"})
			return
		}
	} else if maxEmailQuota != "0" {
		_ = exec.CommandContext(ctx, "opencli", "email-setup", "quota", "set", email, fmt.Sprintf("%s%s", trimFloat(maxEmailQuotaNumeric), allocatedUnit)).Run()
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created email "+email, reqip.ClientIP(r))
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusCreated, map[string]string{"message": "Email " + email + " created successfully", "address": email})
}

// apiEmailDetailGet returns quota, incoming/outgoing restriction status, and server info for one email account.
func apiEmailDetailGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !apiOwnEmailOr403(a, w, r, userID, email) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	restriction := func(direction string) string {
		out, cmdErr := exec.Command("opencli", "email-setup", "email", "restrict", "list", direction).CombinedOutput()
		if cmdErr != nil {
			return "unknown"
		}
		if strings.Contains(string(out), email) {
			return "REJECT"
		}
		return "ACCEPT"
	}

	lines := GetEmailList(ctx, a, userID, currentUsername, userDomains)
	rawLine := ""
	for _, l := range lines {
		if strings.Contains(l, email) {
			rawLine = l
			break
		}
	}
	parts := strings.Fields(strings.TrimLeft(rawLine, "* "))
	var quotaLimit, quotaUsed *string
	if len(parts) >= 2 {
		quotaLimit = &parts[1]
	}
	if len(parts) >= 3 {
		quotaUsed = &parts[2]
	}

	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{
		"address": email, "quota_limit": quotaLimit, "quota_used": quotaUsed,
		"incoming": restriction("receive"), "outgoing": restriction("send"),
		"server_ip": a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername),
		"mail_host": getDedicatedOrSharedIP(ctx, currentUsername),
	})
}

// apiEmailDetailPatch applies quota, incoming/outgoing restriction, and password changes to one email account.
func apiEmailDetailPatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !apiOwnEmailOr403(a, w, r, userID, email) {
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
	maxEmailQuotaNumeric, allocatedUnit := parseMaxQuota(maxEmailQuota)

	var body struct {
		GB       string `json:"gb"`
		Format   string `json:"format"`
		Incoming string `json:"incoming"`
		Outgoing string `json:"outgoing"`
		Password string `json:"password"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.GB = r.Form.Get("gb")
		body.Format = r.Form.Get("format")
		body.Incoming = r.Form.Get("incoming")
		body.Outgoing = r.Form.Get("outgoing")
		body.Password = r.Form.Get("password")
	}
	gb := strings.TrimSpace(body.GB)
	format := strings.TrimSpace(body.Format)
	if format == "" {
		format = "G"
	}

	var actions []string

	if gb != "" {
		msg, ok := emailSetQuota(email, gb, format, maxEmailQuota, maxEmailQuotaNumeric, allocatedUnit)
		if !ok {
			writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid quota format (B, k, M, G, T)"})
			return
		}
		actions = append(actions, msg)
	}

	type restrictionCmd struct{ field, value, verb, direction, logMsg string }
	for _, c := range []restrictionCmd{
		{"incoming", "suspend", "add", "receive", "suspended incoming emails"},
		{"incoming", "allow", "del", "receive", "unsuspended incoming emails"},
		{"outgoing", "suspend", "add", "send", "suspended outgoing emails"},
		{"outgoing", "allow", "del", "send", "unsuspended outgoing emails"},
	} {
		val := body.Incoming
		if c.field == "outgoing" {
			val = body.Outgoing
		}
		if val == c.value {
			_ = exec.CommandContext(ctx, "opencli", "email-setup", "email", "restrict", c.verb, c.direction, email).Run()
			actions = append(actions, c.logMsg)
		}
	}

	newPassword := strings.TrimSpace(body.Password)
	if newPassword != "" {
		_ = exec.CommandContext(ctx, "opencli", "email-setup", "email", "update", email, newPassword).Run()
		actions = append(actions, "updated password")
	}

	if len(actions) == 0 {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "No changes provided (gb, incoming, outgoing, password)"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, strings.Join(actions, "; ")+" for "+email, reqip.ClientIP(r))
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"message": "Updated", "actions": actions})
}

// apiEmailDetailDelete deletes one email account.
func apiEmailDetailDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !apiOwnEmailOr403(a, w, r, userID, email) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, runErr := exec.CommandContext(ctx, "opencli", "email-setup", "email", "del", email).CombinedOutput()
	if runErr != nil {
		writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete " + email + ": " + strings.TrimSpace(string(out))})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted email "+email, reqip.ClientIP(r))
	InvalidateEmailCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusOK, map[string]string{"message": "Email " + email + " deleted"})
}

// ---------------------------------------------------------------------------
// Client configuration download
// ---------------------------------------------------------------------------

// apiEmailConfiguration generates a downloadable mail-client config file
// (Thunderbird, Outlook, or Apple mobileconfig) for one email account.
// Deliberately its own, simpler templates distinct from the richer ones
// handleEmailConfiguration (the UI's /emails/configuration route)
// generates.
func apiEmailConfiguration(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	account := r.PathValue("account")
	if !isValidEmail(account) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	configType := strings.ToLower(r.PathValue("config_type"))
	if configType != "thunderbird" && configType != "outlook" && configType != "apple" {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid config type (thunderbird, outlook, apple)"})
		return
	}

	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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
		pidBase := mailHost + "." + domain
		desc := account + " Secure Email Setup"
		content = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>PayloadContent</key><array><dict>
    <key>EmailAccountType</key><string>EmailTypeIMAP</string>
    <key>EmailAddress</key><string>%s</string>
    <key>EmailAccountName</key><string>%s</string>
    <key>IncomingMailServerHostName</key><string>%s</string>
    <key>IncomingMailServerPortNumber</key><integer>%d</integer>
    <key>IncomingMailServerUseSSL</key>%s
    <key>IncomingMailServerUsername</key><string>%s</string>
    <key>OutgoingMailServerHostName</key><string>%s</string>
    <key>OutgoingMailServerPortNumber</key><integer>%d</integer>
    <key>OutgoingMailServerUseSSL</key>%s
    <key>OutgoingMailServerUsername</key><string>%s</string>
    <key>PayloadType</key><string>com.apple.mail.managed</string>
    <key>PayloadUUID</key><string>%s</string>
    <key>PayloadVersion</key><integer>1</integer>
  </dict></array>
  <key>PayloadDescription</key><string>%s</string>
  <key>PayloadDisplayName</key><string>%s</string>
  <key>PayloadIdentifier</key><string>%s.%s</string>
  <key>PayloadType</key><string>Configuration</string>
  <key>PayloadUUID</key><string>%s</string>
  <key>PayloadVersion</key><integer>1</integer>
</dict></plist>`,
			account, username, mailHost, incomingPort, sslTrueFalse, account,
			mailHost, outgoingPort, sslTrueFalse, account,
			innerUUID, desc, desc, pidBase, outerUUID, payloadUUID,
		)
		filename = username + "_apple_profile.mobileconfig"
		mimetype = "application/x-apple-aspen-config"
	}

	w.Header().Set("Content-Type", mimetype)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte(content))
}

// ---------------------------------------------------------------------------
// Aliases
// ---------------------------------------------------------------------------

// apiEmailAliasesList returns every alias belonging to the current user's domains.
func apiEmailAliasesList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"aliases": GetAliasList(ctx, a, userID, currentUsername, domainSet(domains))})
}

// apiEmailAliasesCreate creates a new email alias.
func apiEmailAliasesCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Domain   string `json:"domain"`
		Target   string `json:"target"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Username = r.Form.Get("username")
		body.Domain = r.Form.Get("domain")
		body.Target = r.Form.Get("target")
	}
	username := strings.TrimSpace(body.Username)
	domain := strings.TrimSpace(body.Domain)
	target := strings.TrimSpace(body.Target)

	if username == "" || domain == "" || target == "" {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "username, domain, and target are required"})
		return
	}
	if !validators.IsValidEmailUsername(username) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "username can only contain letters, numbers, and . _ % + - (no @)"})
		return
	}
	if !isValidEmail(target) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "target must be a valid email address"})
		return
	}
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	source := username + "@" + domain

	out, cmdErr := exec.CommandContext(ctx, "opencli", "email-setup", "alias", "add", source, target).CombinedOutput()
	if cmdErr != nil {
		writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created alias "+source+" -> "+target, reqip.ClientIP(r))
	InvalidateAliasCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusCreated, map[string]string{
		"message": "Alias " + source + " → " + target + " created", "source": source, "target": target,
	})
}

// apiEmailAliasDetailGet returns the target addresses for one alias.
func apiEmailAliasDetailGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	_, domain, _ := strings.Cut(email, "@")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	aliasList := GetAliasList(ctx, a, userID, currentUsername, domainSet(domains))
	targets := []string{}
	for _, entry := range aliasList {
		if entry.Source == email {
			targets = entry.Targets
			break
		}
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"source": email, "targets": targets})
}

// apiEmailAliasDetailPost adds a target address to an existing alias.
func apiEmailAliasDetailPost(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	_, domain, _ := strings.Cut(email, "@")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Target string `json:"target"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Target = r.Form.Get("target")
	}
	target := strings.TrimSpace(body.Target)
	if target == "" || !isValidEmail(target) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "A valid target email address is required"})
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "email-setup", "alias", "add", email, target).CombinedOutput()
	if cmdErr != nil {
		writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "added alias "+email+" -> "+target, reqip.ClientIP(r))
	InvalidateAliasCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusCreated, map[string]string{"message": "Target " + target + " added to " + email})
}

// apiEmailAliasDetailDelete removes a single target from an alias, or the whole alias when delete_all is set.
func apiEmailAliasDetailDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	_, domain, _ := strings.Cut(email, "@")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Target    string `json:"target"`
		DeleteAll bool   `json:"delete_all"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	target := strings.TrimSpace(body.Target)

	if body.DeleteAll {
		domains, _ := a.AllDomainsForUser(ctx, userID)
		aliasList := GetAliasList(ctx, a, userID, currentUsername, domainSet(domains))
		for _, entry := range aliasList {
			if entry.Source == email {
				for _, t := range entry.Targets {
					_, _ = exec.CommandContext(ctx, "opencli", "email-setup", "alias", "del", email, t).CombinedOutput()
				}
				break
			}
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "deleted alias "+email, reqip.ClientIP(r))
		InvalidateAliasCache(ctx, a, userID, currentUsername)
		writeAPIEmailsJSON(w, http.StatusOK, map[string]string{"message": "Alias " + email + " deleted"})
		return
	}

	if target == "" || !isValidEmail(target) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Provide target (email) or delete_all: true"})
		return
	}

	_, _ = exec.CommandContext(ctx, "opencli", "email-setup", "alias", "del", email, target).CombinedOutput()
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted alias "+email+" -> "+target, reqip.ClientIP(r))
	InvalidateAliasCache(ctx, a, userID, currentUsername)
	writeAPIEmailsJSON(w, http.StatusOK, map[string]string{"message": "Target " + target + " removed from " + email})
}

// ---------------------------------------------------------------------------
// Default / catch-all alias
// ---------------------------------------------------------------------------

// apiEmailDefaultGet returns a domain's current default (catch-all) destination, if any.
func apiEmailDefaultGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	current := checkCurrentDefaultAliasForDomain(domain)
	var destination *string
	if current != "" {
		destination = &current
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"domain": domain, "destination": destination})
}

// apiEmailDefaultPut sets or clears a domain's default (catch-all) destination.
func apiEmailDefaultPut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Destination string `json:"destination"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Destination = r.Form.Get("destination")
	}
	destination := strings.TrimSpace(body.Destination)

	if destination != "" && !isValidEmail(destination) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "destination must be a valid email address"})
		return
	}

	if setErr := setDefaultAliasForDomain(domain, destination); setErr != nil {
		writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": setErr.Error()})
		return
	}

	if destination != "" {
		_ = logger.RecordUserAction(a.Config, currentUsername, "set default email for "+domain+" to "+destination, reqip.ClientIP(r))
	} else {
		_ = logger.RecordUserAction(a.Config, currentUsername, "removed default email for "+domain, reqip.ClientIP(r))
	}

	var destPtr *string
	if destination != "" {
		destPtr = &destination
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"domain": domain, "destination": destPtr})
}

// ---------------------------------------------------------------------------
// Deliverability
// ---------------------------------------------------------------------------

// apiEmailDeliverabilityAll checks deliverability (SPF/DKIM/DMARC/rDNS
// etc.) for every domain the current user owns, concurrently.
func apiEmailDeliverabilityAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	serverIP := getDedicatedOrSharedIP(ctx, currentUsername)

	results := make([]DeliverabilityCheck, len(domains))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, d := range domains {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, domainURL string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = checkDomainDeliverability(ctx, a, domainURL, serverIP)
		}(i, d.DomainURL)
	}
	wg.Wait()

	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"domains": results})
}

// apiEmailDeliverabilityDomain checks deliverability for a single domain.
func apiEmailDeliverabilityDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	serverIP := getDedicatedOrSharedIP(ctx, currentUsername)
	writeAPIEmailsJSON(w, http.StatusOK, checkDomainDeliverability(ctx, a, domain, serverIP))
}

// ---------------------------------------------------------------------------
// Sieve filters
// ---------------------------------------------------------------------------

// apiEmailFilterGet returns the raw and parsed Sieve filter for an email account.
func apiEmailFilterGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	resolvedPath, forbidden, resolveErr := resolveSievePath(email, userDomains)
	if forbidden {
		writeAPIEmailsJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain."})
		return
	}
	if resolveErr != nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": resolveErr.Error()})
		return
	}

	raw := ""
	if content, readErr := os.ReadFile(resolvedPath); readErr == nil {
		raw = string(content)
	}
	writeAPIEmailsJSON(w, http.StatusOK, map[string]any{"email": email, "raw": raw, "parsed": sieveparser.Parse(raw)})
}

// apiEmailFilterPut writes a new Sieve filter for an email account.
func apiEmailFilterPut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	email := r.PathValue("email")
	if !isValidEmail(email) {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
		return
	}
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	resolvedPath, forbidden, resolveErr := resolveSievePath(email, userDomains)
	if forbidden {
		writeAPIEmailsJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain."})
		return
	}
	if resolveErr != nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": resolveErr.Error()})
		return
	}

	var body struct {
		Content *string `json:"content"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		if r.Form.Has("content") {
			v := r.Form.Get("content")
			body.Content = &v
		}
	}
	if body.Content == nil {
		writeAPIEmailsJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	if writeErr := writeSieve(resolvedPath, *body.Content); writeErr != nil {
		writeAPIEmailsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving filter: " + writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited sieve filter for "+email, reqip.ClientIP(r))
	writeAPIEmailsJSON(w, http.StatusOK, map[string]string{"message": "Filter saved"})
}
