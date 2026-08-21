package dns

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleUpdateDNSRecord replaces one or more zone-file lines (rowID
// through endRowID) with new content, after checking the serial number
// hasn't changed since the client last loaded the zone.
func handleUpdateDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.PathValue("domain")
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	rowID, convErr := strconv.Atoi(r.PathValue("row_id"))
	if convErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Invalid row ID"})
		return
	}

	_ = r.ParseForm()
	newContent := r.Form.Get("newContent")
	endRowID := rowID
	if v := r.Form.Get("endRowId"); v != "" {
		if parsed, perr := strconv.Atoi(v); perr == nil {
			endRowID = parsed
		}
	}

	path := zoneFilePath(domain)
	if !fileExists(path) {
		flashAndRedirect(a, w, r, "error", "Error:Zone file not found for domain: "+domain, "/domains/edit-dns-zone")
		return
	}

	serialFromPost := r.Form.Get("serial")
	if serialFromPost == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Serial number not provided. Please try again."})
		return
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Error updating DNS record: " + readErr.Error()})
		return
	}
	lines := readLinesKeepEnds(string(content))
	serialFromZone := readSerialNumber(lines)
	if serialFromPost != serialFromZone {
		writeJSON(w, http.StatusConflict, map[string]string{"status": "error", "message": "Serial number mismatch. Zone was edited by another user/program. Please try again."})
		return
	}

	if rowID < 1 || rowID > len(lines) || rowID > endRowID || endRowID > len(lines) {
		writeJSON(w, http.StatusOK, map[string]string{"error": fmt.Sprintf("Invalid row ID: %d", rowID)})
		return
	}

	newBlock := newContent
	if !strings.HasSuffix(newBlock, "\n") {
		newBlock += "\n"
	}
	newLines := make([]string, 0, len(lines))
	newLines = append(newLines, lines[:rowID-1]...)
	newLines = append(newLines, newBlock)
	newLines = append(newLines, lines[endRowID:]...)

	if writeErr := os.WriteFile(path, []byte(strings.Join(newLines, "")), 0o644); writeErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Error updating DNS record: " + writeErr.Error()})
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated DNS record to "+newContent+" for domain "+domain, ipAddress)
	RestartDNSService(domain)

	writeJSON(w, http.StatusOK, map[string]string{"updated_row": newContent, "message": fmt.Sprintf("Row with ID %d updated successfully", rowID)})
}

// handleDeleteDNSRecord deletes a zone record by line number. rowId in
// the URL is 0-indexed (the JS caller sends item.line_number - 1), unlike
// handleUpdateDNSRecord's 1-indexed row_id - an inconsistency between the
// two routes kept for compatibility with the existing frontend.
func handleDeleteDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rowID, convErr := strconv.Atoi(r.PathValue("rowId"))
	if convErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Invalid row ID"})
		return
	}

	_ = r.ParseForm()
	domain := r.Form.Get("domain")
	endRowID, hasEndRowID := 0, false
	if v := r.Form.Get("endRowId"); v != "" {
		if parsed, perr := strconv.Atoi(v); perr == nil {
			endRowID = parsed
			hasEndRowID = true
		}
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	path := zoneFilePath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Error deleting DNS record: " + readErr.Error()})
		return
	}
	lines := readLinesKeepEnds(string(content))

	if rowID < 0 || rowID >= len(lines) {
		writeJSON(w, http.StatusOK, map[string]string{"error": fmt.Sprintf("Invalid row ID: %d", rowID)})
		return
	}

	var deletedRow string
	if hasEndRowID && endRowID > rowID+1 {
		if endRowID > len(lines) {
			endRowID = len(lines)
		}
		deletedRow = strings.Join(lines[rowID:endRowID], "")
		lines = append(lines[:rowID], lines[endRowID:]...)
	} else {
		deletedRow = lines[rowID]
		lines = append(lines[:rowID], lines[rowID+1:]...)
	}

	if writeErr := os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644); writeErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Error deleting DNS record: " + writeErr.Error()})
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted DNS record "+deletedRow+" for domain "+domain, ipAddress)
	RestartDNSService(domain)

	writeJSON(w, http.StatusOK, map[string]string{"message": "Row deleted successfully", "deleted_row": strings.TrimSpace(deletedRow)})
}

// handleAddDNSRecord appends a new record to the domain's zone file.
func handleAddDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	name := r.Form.Get("Name")
	ttl := r.Form.Get("TTL")
	recordType := r.Form.Get("Type")
	record := r.Form.Get("Record")
	priority := r.Form.Get("Priority")
	domain := r.Form.Get("Domain")

	redirectTarget := "/domains/edit-dns-zone"
	if domain != "" {
		redirectTarget = "/domains/edit-dns-zone/" + domain
	}

	if name == "" || ttl == "" || recordType == "" || record == "" || domain == "" {
		flashAndRedirect(a, w, r, "error", "Please provide all required fields.", redirectTarget)
		return
	}

	for _, v := range []string{name, ttl, recordType, record, priority, domain} {
		if strings.ContainsAny(v, "\n\r") {
			flashAndRedirect(a, w, r, "error", "Invalid characters in submitted data.", redirectTarget)
			return
		}
	}

	path := zoneFilePath(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if !fileExists(path) {
		flashAndRedirect(a, w, r, "error", "Zone file not found.", "/domains/edit-dns-zone")
		return
	}

	if strings.HasSuffix(name, domain) {
		name += "."
	}
	if strings.HasSuffix(record, domain) {
		record += "."
	}
	if recordType == "TXT" && !(strings.HasPrefix(record, `"`) && strings.HasSuffix(record, `"`)) {
		record = `"` + record + `"`
	}

	var newRecord string
	if priority != "" {
		newRecord = fmt.Sprintf("%s %s IN %s %s %s", name, ttl, recordType, priority, record)
	} else {
		newRecord = fmt.Sprintf("%s %s IN %s %s", name, ttl, recordType, record)
	}

	if recordType == "CNAME" && CnameRecordExists(ctx, a, path, name) {
		flashAndRedirect(a, w, r, "error", "CNAME record with this name already exists.", redirectTarget)
		return
	}

	f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		flashAndRedirect(a, w, r, "error", "Error adding DNS record: "+openErr.Error(), redirectTarget)
		return
	}
	_, writeErr := f.WriteString(newRecord + "\n")
	_ = f.Close()
	if writeErr != nil {
		flashAndRedirect(a, w, r, "error", "Error adding DNS record: "+writeErr.Error(), redirectTarget)
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "added DNS record "+newRecord+" for domain "+domain, ipAddress)
	RestartDNSService(domain)
	flashAndRedirect(a, w, r, "success", "DNS record added successfully.", redirectTarget)
}

// handleSaveDNSZone validates the whole new zone content by copying it
// into the shared DNS container and running named-checkzone before ever
// touching the real on-disk file.
func handleSaveDNSZone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	newContent := r.Form.Get("zone_content")
	if newContent == "" {
		writeText(w, "New content not provided.")
		return
	}
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	tmpFile, tmpErr := os.CreateTemp("", "dns-zone-*")
	if tmpErr != nil {
		writeText(w, "Failed - zone validation error")
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	_, writeErr := tmpFile.WriteString(newContent)
	_ = tmpFile.Close()
	if writeErr != nil {
		writeText(w, "Failed - zone validation error")
		return
	}

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := exec.CommandContext(cctx, "podman", "cp", tmpPath, "openpanel_dns:/tmp/"+domain+".zone").Run(); err != nil {
		writeText(w, "Failed - zone validation error")
		return
	}
	if err := exec.CommandContext(cctx, "podman", "exec", "openpanel_dns", "named-checkzone", domain, "/tmp/"+domain+".zone").Run(); err != nil {
		writeText(w, "Failed - zone validation error")
		return
	}

	path := zoneFilePath(domain)
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		writeText(w, "Failed - zone validation error")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited DNS zone for domain "+domain, ipAddress)
	RestartDNSService(domain)

	writeText(w, "DNS zone saved and DNS service restarted.")
}

// handleRestartDNSZone resets a domain's zone file back to the default
// template via opencli.
func handleRestartDNSZone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")

	if domain == "" {
		flashAndRedirect(a, w, r, "error", "Please provide a domain name.", "/domains/edit-dns-zone")
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := exec.CommandContext(cctx, "opencli", "domains-dns", "default", domain, "-y").Run(); err != nil {
		flashAndRedirect(a, w, r, "error", "Failed to restart DNS zone: "+err.Error(), "/domains/edit-dns-zone/"+domain)
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "reset DNS zone for domain "+domain, ipAddress)
	flashAndRedirect(a, w, r, "success", "DNS zone restarted successfully.", "/domains/edit-dns-zone/"+domain)
}

// zoneExportHeaderTemplate is the informational header block prepended to
// an exported zone file.
const zoneExportHeaderTemplate = `
;;
;; Domain:     %s
;; Exported:   %s
;;
;; This file is intended for use for informational and archival
;; purposes ONLY and MUST be edited before use on a production
;; DNS server.  In particular, you must:
;;   -- update the SOA record with the correct authoritative name server
;;   -- update the SOA record with the contact e-mail address information
;;   -- update the NS record(s) with the authoritative name servers for this domain.
;;
;; For further information, please consult the BIND documentation
;; located on the following website:
;;
;; http://www.isc.org/
;;
;; And RFC 1035:
;;
;; http://www.ietf.org/rfc/rfc1035.txt
;;
;; Please note that we do NOT offer technical support for any use
;; of this zone data, the BIND name server, or any other third-party
;; DNS software.
;;
;; Use at your own risk.
;;
`

// handleExportDNSZone streams a domain's zone file for download, prefixed
// with an informational header.
func handleExportDNSZone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")

	if domain == "" {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Please provide a domain name."})
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	content, readErr := os.ReadFile(zoneFilePath(domain))
	if readErr != nil {
		writeText(w, "Zone file not found.")
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf(zoneExportHeaderTemplate, domain, timestamp)

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "exported DNS zone for domain "+domain, ipAddress)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+domain+`.zone"`)
	_, _ = w.Write([]byte(header))
	_, _ = w.Write(content)
}
