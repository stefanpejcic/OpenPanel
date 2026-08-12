package dns

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
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

// RegisterAPI wires the DNS REST endpoints onto mux. Several routes share
// a domain prefix with a literal suffix (e.g. /api/dns/<domain>/raw) - Go's
// http.ServeMux requires a "{...}" wildcard to be the final segment, so
// GET/PATCH/DELETE/POST each get one "{rest...}" catch-all per verb and
// the dispatch funcs below strip the known suffix by hand so that a
// literal suffix always wins over the wildcard. apiregistry.Add still
// records each logical route separately for /api/endpoints.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "dns", "GET /api/dns", func(w http.ResponseWriter, r *http.Request) { apiDNSList(a, w, r) })

	apiregistry.Add("GET /api/dns/{domain}")
	apiregistry.Add("GET /api/dns/{domain}/raw")
	apiregistry.Add("GET /api/dns/{domain}/export")
	mux.Handle("GET /api/dns/{rest...}", auth.RequireAPI(a, "dns")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDNSGetDispatch(a, w, r) })))

	apiregistry.Add("PUT /api/dns/{domain}/raw")
	mux.Handle("PUT /api/dns/{rest...}", auth.RequireAPI(a, "dns")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDNSPutDispatch(a, w, r) })))

	apiregistry.Add("POST /api/dns/{domain}/records")
	apiregistry.Add("POST /api/dns/{domain}/reset")
	mux.Handle("POST /api/dns/{rest...}", auth.RequireAPI(a, "dns")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDNSPostDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "dns", "PATCH /api/dns/{domain}/records/{row_id}", func(w http.ResponseWriter, r *http.Request) { apiDNSUpdateRecord(a, w, r) })
	apiregistry.Handle(mux, a, "dns", "DELETE /api/dns/{domain}/records/{row_id}", func(w http.ResponseWriter, r *http.Request) { apiDNSDeleteRecord(a, w, r) })
}

func writeAPIDNSJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiDNSGetDispatch dispatches GET /api/dns/{domain}[/raw|/export].
func apiDNSGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/raw"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/raw"))
		apiDNSRawGet(a, w, r)
	case strings.HasSuffix(rest, "/export"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/export"))
		apiDNSExport(a, w, r)
	default:
		r.SetPathValue("domain", rest)
		apiDNSGet(a, w, r)
	}
}

// apiDNSPutDispatch dispatches PUT /api/dns/{domain}/raw.
func apiDNSPutDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if strings.HasSuffix(rest, "/raw") {
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/raw"))
		apiDNSRawPut(a, w, r)
		return
	}
	http.NotFound(w, r)
}

// apiDNSPostDispatch dispatches POST /api/dns/{domain}/records|/reset.
func apiDNSPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/records"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/records"))
		apiDNSAddRecord(a, w, r)
	case strings.HasSuffix(rest, "/reset"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/reset"))
		apiDNSReset(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

// apiZoneRecord is one parsed record from a zone file's line-numbered
// breakdown.
type apiZoneRecord struct {
	LineNumber    int    `json:"line_number"`
	EndLineNumber int    `json:"end_line_number"`
	Line          string `json:"line"`
	Multiline     bool   `json:"multiline"`
}

// apiDNSList returns every domain owned by the user plus whether each has
// a zone file on disk.
func apiDNSList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domains, _ := a.AllDomainsForUser(r.Context(), userID)

	result := make([]map[string]any, 0, len(domains))
	for _, d := range domains {
		result = append(result, map[string]any{
			"domain_id":        d.DomainID,
			"domain_url":       d.DomainURL,
			"zone_file_exists": fileExists(zoneFilePath(d.DomainURL)),
		})
	}
	writeAPIDNSJSON(w, http.StatusOK, map[string]any{"domains": result})
}

// apiDNSGet returns a domain's zone parsed into line-numbered records,
// along with its serial number and any validation error.
func apiDNSGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	path := zoneFilePath(domain)
	if !fileExists(path) {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	entries := parseZoneWithLineNumbers(string(content))
	records := make([]apiZoneRecord, 0, len(entries))
	for _, e := range entries {
		records = append(records, apiZoneRecord{LineNumber: e.LineNumber, EndLineNumber: e.EndLineNumber, Line: e.Line, Multiline: e.Multiline})
	}
	serial := readSerialNumber(readLinesKeepEnds(string(content)))
	validationError := validateZoneFile(domain, path)

	writeAPIDNSJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "serial": serial, "records": records, "validation_error": validationError,
	})
}

// apiDNSRawGet returns a domain's zone file as raw text.
func apiDNSRawGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	path := zoneFilePath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"domain": domain, "content": string(content)})
}

// apiDNSRawPut replaces a domain's whole zone file with the submitted
// content, after validating it via named-checkzone.
func apiDNSRawPut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newContent := body.Content
	if newContent == "" {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	tmpFile, tmpErr := os.CreateTemp("", "dns-zone-*.zone")
	if tmpErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": tmpErr.Error()})
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, writeErr := tmpFile.WriteString(newContent); writeErr != nil {
		_ = tmpFile.Close()
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}
	_ = tmpFile.Close()

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if cpErr := exec.CommandContext(cctx, "podman", "cp", tmpPath, "openpanel_dns:/tmp/"+domain+".zone").Run(); cpErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": cpErr.Error()})
		return
	}
	checkOut, checkErr := exec.CommandContext(cctx, "podman", "exec", "openpanel_dns", "named-checkzone", domain, "/tmp/"+domain+".zone").CombinedOutput()
	if checkErr != nil {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "Zone validation failed", "details": strings.TrimSpace(string(checkOut))})
		return
	}

	path := zoneFilePath(domain)
	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited DNS zone for "+domain, reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"message": "DNS zone saved and reloaded"})
}

// apiDNSAddRecord appends a new record to a domain's zone file.
func apiDNSAddRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	path := zoneFilePath(domain)
	if !fileExists(path) {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}

	var body struct {
		Name     string `json:"name"`
		TTL      any    `json:"ttl"`
		Type     string `json:"type"`
		Record   string `json:"record"`
		Priority string `json:"priority"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	name := strings.TrimSpace(body.Name)
	ttl := "3600"
	switch v := body.TTL.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			ttl = strings.TrimSpace(v)
		}
	case float64:
		ttl = strconv.Itoa(int(v))
	}
	recordType := strings.ToUpper(strings.TrimSpace(body.Type))
	record := strings.TrimSpace(body.Record)
	priority := strings.TrimSpace(body.Priority)

	if name == "" || recordType == "" || record == "" {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "name, type, and record are required"})
		return
	}

	for _, v := range []string{name, ttl, recordType, record, priority, domain} {
		if strings.ContainsAny(v, "\n\r") {
			writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid characters in submitted data"})
			return
		}
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
	if recordType == "CNAME" && cnameRecordExists(ctx, a, path, name) {
		writeAPIDNSJSON(w, http.StatusConflict, map[string]string{"error": "CNAME record with this name already exists"})
		return
	}

	var newRecord string
	if priority != "" {
		newRecord = name + " " + ttl + " IN " + recordType + " " + priority + " " + record
	} else {
		newRecord = name + " " + ttl + " IN " + recordType + " " + record
	}

	f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
		return
	}
	_, writeErr := f.WriteString(newRecord + "\n")
	_ = f.Close()
	if writeErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "added DNS record "+newRecord+" for "+domain, reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusCreated, map[string]string{"message": "DNS record added", "record": newRecord})
}

// apiDNSUpdateRecord replaces one or more zone-file lines identified by
// row ID with new content.
func apiDNSUpdateRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	rowID, convErr := strconv.Atoi(r.PathValue("row_id"))
	if convErr != nil {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row ID"})
		return
	}

	path := zoneFilePath(domain)
	if !fileExists(path) {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}

	var body struct {
		Content  string `json:"content"`
		EndRowID any    `json:"end_row_id"`
		Serial   string `json:"serial"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newContent := strings.TrimSpace(body.Content)
	if newContent == "" {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}
	endRowID := rowID
	switch v := body.EndRowID.(type) {
	case float64:
		endRowID = int(v)
	}
	serial := strings.TrimSpace(body.Serial)

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}
	lines := readLinesKeepEnds(string(content))

	if serial != "" {
		zoneSerial := readSerialNumber(lines)
		if zoneSerial != "" && serial != zoneSerial {
			writeAPIDNSJSON(w, http.StatusConflict, map[string]string{"error": "Serial number mismatch — zone was modified externally"})
			return
		}
	}

	if !(rowID >= 1 && rowID <= len(lines) && rowID <= endRowID && endRowID <= len(lines)) {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row ID: " + strconv.Itoa(rowID)})
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
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated DNS record (line "+strconv.Itoa(rowID)+") for "+domain, reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"message": "Record at line " + strconv.Itoa(rowID) + " updated", "record": newContent})
}

// apiDNSDeleteRecord removes one or more zone-file lines identified by row
// ID.
func apiDNSDeleteRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	rowID, convErr := strconv.Atoi(r.PathValue("row_id"))
	if convErr != nil {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row ID"})
		return
	}

	path := zoneFilePath(domain)
	if !fileExists(path) {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}

	var body struct {
		EndRowID any `json:"end_row_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	endRowID, hasEndRowID := 0, false
	if v, ok := body.EndRowID.(float64); ok {
		endRowID = int(v)
		hasEndRowID = true
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}
	lines := readLinesKeepEnds(string(content))

	if !(rowID >= 1 && rowID <= len(lines)) {
		writeAPIDNSJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row ID: " + strconv.Itoa(rowID)})
		return
	}

	var deleted string
	if hasEndRowID && endRowID > rowID {
		if endRowID > len(lines) {
			endRowID = len(lines)
		}
		deleted = strings.Join(lines[rowID-1:endRowID], "")
		lines = append(lines[:rowID-1], lines[endRowID:]...)
	} else {
		deleted = lines[rowID-1]
		lines = append(lines[:rowID-1], lines[rowID:]...)
	}

	if writeErr := os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644); writeErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted DNS record from "+domain+": "+strings.TrimSpace(deleted), reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"message": "Record deleted", "deleted": strings.TrimSpace(deleted)})
}

// apiDNSReset resets a domain's zone file back to the default template.
func apiDNSReset(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, runErr := exec.CommandContext(cctx, "opencli", "domains-dns", "default", domain, "-y").CombinedOutput()
	if runErr != nil {
		writeAPIDNSJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "reset DNS zone for "+domain, reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"message": "DNS zone for " + domain + " reset to default"})
}

// apiDNSExport returns a domain's zone file content for download.
func apiDNSExport(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeAPIDNSJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	path := zoneFilePath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPIDNSJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "exported DNS zone for "+domain, reqip.ClientIP(r))
	writeAPIDNSJSON(w, http.StatusOK, map[string]string{"domain": domain, "content": string(content), "filename": domain + ".zone"})
}
