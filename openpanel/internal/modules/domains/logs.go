package domains

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// caddyDomlogsDir is a var so tests can point it at a temp dir
var caddyDomlogsDir = "/var/log/caddy/domlogs"

// AccessLogTLS is the part of Caddy's "tls" object the page shows
type AccessLogTLS struct {
	Version    int    `json:"version"`
	Proto      string `json:"proto,omitempty"`
	ServerName string `json:"server_name,omitempty"`
}

// AccessLogRequest mirrors one Caddy JSON access-log line's nested "request" object
type AccessLogRequest struct {
	ClientIP string              `json:"client_ip"`
	Method   string              `json:"method"`
	URI      string              `json:"uri"`
	Host     string              `json:"host,omitempty"`
	Proto    string              `json:"proto,omitempty"`
	Headers  map[string][]string `json:"headers,omitempty"`
	TLS      *AccessLogTLS       `json:"tls,omitempty"`
}

// AccessLogEntry mirrors one Caddy JSON access-log line.
type AccessLogEntry struct {
	TS          float64             `json:"ts"`
	Level       string              `json:"level"`
	Logger      string              `json:"logger"`
	Msg         string              `json:"msg"`
	Request     AccessLogRequest    `json:"request"`
	Status      int                 `json:"status"`
	Size        int                 `json:"size"`
	Duration    float64             `json:"duration"`
	BytesRead   int                 `json:"bytes_read"`
	RespHeaders map[string][]string `json:"resp_headers,omitempty"`
	Raw         string              `json:"-"`
}

// headerValue is the first value of a header, Caddy logs canonical names but a proxy may not
func headerValue(h map[string][]string, name string) string {
	if v := h[name]; len(v) > 0 {
		return v[0]
	}
	for k, v := range h {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func (e AccessLogEntry) UserAgent() string   { return headerValue(e.Request.Headers, "User-Agent") }
func (e AccessLogEntry) Referer() string     { return headerValue(e.Request.Headers, "Referer") }
func (e AccessLogEntry) ContentType() string { return headerValue(e.RespHeaders, "Content-Type") }
func (e AccessLogEntry) Server() string      { return headerValue(e.RespHeaders, "Server") }

// TLSLabel reads like "TLS 1.3 · h2", "none" for plain http
func (e AccessLogEntry) TLSLabel() string {
	t := e.Request.TLS
	if t == nil || t.Version == 0 {
		return "none"
	}
	label := map[int]string{769: "TLS 1.0", 770: "TLS 1.1", 771: "TLS 1.2", 772: "TLS 1.3"}[t.Version]
	if label == "" {
		label = "TLS"
	}
	if t.Proto != "" {
		label += " · " + t.Proto
	}
	return label
}

// StatusClass is the first digit of the status, what the 2xx-5xx chips filter on
func (e AccessLogEntry) StatusClass() string {
	if e.Status < 100 || e.Status > 599 {
		return ""
	}
	return strconv.Itoa(e.Status / 100)
}

// matches is the search box: client IP, request, status, user agent, referrer, level, protocol and host
func (e AccessLogEntry) matches(q string) bool {
	if q == "" {
		return true
	}
	hay := strings.ToLower(strings.Join([]string{e.Request.ClientIP, e.Request.Method, e.Request.URI, strconv.Itoa(e.Status), e.UserAgent(), e.Referer(), e.Level, e.Request.Proto, e.Request.Host}, " "))
	return strings.Contains(hay, q)
}

// parseAccessLog reads a Caddy JSON log, newest first, skipping lines that aren't JSON
func parseAccessLog(content string) []AccessLogEntry {
	var entries []AccessLogEntry
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry AccessLogEntry
		if json.Unmarshal([]byte(line), &entry) == nil {
			entry.Raw = line
			entries = append(entries, entry)
		}
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

// accessLogPath is the domain's access.log, false when the name could step outside the log dir
func accessLogPath(domain string) (string, bool) {
	if domain == "" || domain == "." || domain == ".." || strings.ContainsAny(domain, `/\`) || strings.ContainsRune(domain, 0) {
		return "", false
	}
	dir := filepath.Clean(caddyDomlogsDir)
	p := filepath.Join(dir, domain, "access.log")
	if filepath.Dir(filepath.Dir(p)) != dir {
		return "", false
	}
	return filepath.ToSlash(p), true
}

func humanBytes(n int) string {
	switch {
	case n < 1000:
		return strconv.Itoa(n) + " B"
	case n < 1e6:
		return fmt.Sprintf("%.1f KB", float64(n)/1e3)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/1e6)
}

func humanDuration(s float64) string {
	switch {
	case s < 1:
		return strconv.Itoa(int(math.Round(s*1000))) + " ms"
	case s < 10:
		return fmt.Sprintf("%.2f s", s)
	}
	return fmt.Sprintf("%.1f s", s)
}

// AccessLogRow is one table row with its display strings worked out
type AccessLogRow struct {
	AccessLogEntry
	Index             int
	SortTS            string
	Day, Clock, Stamp string
	Path              string
	SizeText, DurText string
	Slow              bool
	Tone              string
}

// AccessLogDetail is what the expanded row and the raw view show, kept out of the table html
type AccessLogDetail struct {
	Raw      string `json:"raw"`
	URL      string `json:"url"`
	Shown    string `json:"shown"`
	UA       string `json:"ua"`
	Referer  string `json:"ref"`
	IP       string `json:"ip"`
	Stamp    string `json:"stamp"`
	Proto    string `json:"proto"`
	TLS      string `json:"tls"`
	Response string `json:"resp"`
	CType    string `json:"ctype"`
	Server   string `json:"server"`
	Level    string `json:"level"`
}

func statusTone(s int) string {
	switch {
	case s >= 500:
		return "is-danger"
	case s >= 400:
		return "is-warning"
	case s >= 300:
		return "is-accent"
	case s >= 200:
		return "is-success"
	}
	return ""
}

func buildAccessLogRows(domain string, entries []AccessLogEntry) ([]AccessLogRow, []AccessLogDetail) {
	rows := make([]AccessLogRow, len(entries))
	details := make([]AccessLogDetail, len(entries))
	for i, e := range entries {
		sec, frac := math.Modf(e.TS)
		t := time.Unix(int64(sec), int64(frac*1e9)).UTC()
		path, err := url.PathUnescape(e.Request.URI)
		if err != nil {
			path = e.Request.URI
		}
		rows[i] = AccessLogRow{
			AccessLogEntry: e, Index: i, SortTS: strconv.FormatFloat(e.TS, 'f', 6, 64), Day: t.Format("Jan 2"), Clock: t.Format("15:04:05"),
			Stamp: t.Format("2006-01-02 15:04:05.000") + " UTC", Path: path,
			SizeText: humanBytes(e.Size), DurText: humanDuration(e.Duration), Slow: e.Duration >= 5, Tone: statusTone(e.Status),
		}
		scheme := "http://"
		if e.Request.TLS != nil {
			scheme = "https://"
		}
		details[i] = AccessLogDetail{
			Raw: e.Raw, URL: scheme + domain + e.Request.URI, Shown: scheme + domain + path,
			UA: e.UserAgent(), Referer: e.Referer(), IP: e.Request.ClientIP, Stamp: rows[i].Stamp,
			Proto: e.Request.Proto, TLS: e.TLSLabel(), Response: rows[i].SizeText + " in " + rows[i].DurText,
			CType: e.ContentType(), Server: e.Server(), Level: e.Level,
		}
	}
	return rows, details
}

// DomainLogFile is one row of the domain picker
type DomainLogFile struct {
	Domain string
	Size   int64
	Exists bool
}

func domainLogFiles(domains []appctx.Domain) []DomainLogFile {
	out := make([]DomainLogFile, 0, len(domains))
	for _, d := range domains {
		f := DomainLogFile{Domain: d.DomainURL}
		if p, ok := accessLogPath(d.DomainURL); ok {
			if info, err := tenantfs.Stat(p); err == nil {
				f.Size, f.Exists = info.Size(), true
			}
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Domain < out[j].Domain })
	return out
}

// handleViewDomainAccessLog shows recent access-log entries for a domain.
func handleViewDomainAccessLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainName := r.PathValue("domain_name")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderDomainLogsSelectPage(a, w, r, domainLogFiles(domainsList), "")
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	logFilePath, ok := accessLogPath(domainName)
	if !ok {
		http.Error(w, "Invalid domain.", http.StatusBadRequest)
		return
	}
	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	files := domainLogFiles(domainsList)
	content, err := tenantfs.ReadFile(logFilePath)
	if err != nil {
		renderDomainLogsSelectPage(a, w, r, files, domainName)
		return
	}

	data := buildDomainLogsData(domainName, parseAccessLog(string(content)), r.URL.Query(),
		web.AtoiDefault(a.Config.Get("domain_log_per_page", "1000"), 1000),
		web.AtoiDefault(a.Config.Get("domain_log_max_for_show_all", "10000"), 10000))
	renderDomainLogsPage(a, w, r, data, files)
}

// buildDomainLogsData applies the search and status chip to the whole log, then pages it
func buildDomainLogsData(domain string, entries []AccessLogEntry, params url.Values, perPage, maxShowAll int) DomainLogsPageData {
	d := DomainLogsPageData{
		DomainName: domain, TotalLines: len(entries), TotalAllowedLinesForShowAll: maxShowAll,
		Search: strings.TrimSpace(params.Get("q")), Counts: map[string]int{},
	}
	if s := params.Get("status"); len(s) == 1 && s >= "2" && s <= "5" {
		d.Status = s
	}
	if params.Get("view") == "raw" {
		d.View = "raw"
	}
	q := strings.ToLower(d.Search)
	var matched []AccessLogEntry
	for _, e := range entries {
		if !e.matches(q) {
			continue
		}
		d.Counts[""]++
		if c := e.StatusClass(); c != "" {
			d.Counts[c]++
		}
		if d.Status == "" || e.StatusClass() == d.Status {
			matched = append(matched, e)
		}
	}
	d.Matched = len(matched)

	if perPage < 1 {
		perPage = 1000
	}
	d.ShowAll = params.Get("show_all") == "true" && d.Matched <= maxShowAll
	if d.ShowAll {
		perPage = max(d.Matched, 1)
	}
	d.ItemsPerPage = perPage
	d.TotalPages = max(1, (d.Matched+perPage-1)/perPage)
	d.CurrentPage = min(max(web.AtoiDefault(params.Get("page"), 1), 1), d.TotalPages)
	start := (d.CurrentPage - 1) * perPage
	end := min(start+perPage, d.Matched)
	d.Rows, d.Details = buildAccessLogRows(domain, matched[start:end])
	d.LogPageEntries = buildLogPageEntries(d.CurrentPage, d.TotalPages)
	return d
}

// handleDownloadDomainAccessLog sends the domain's raw access.log as it is on disk
func handleDownloadDomainAccessLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domainName := r.PathValue("domain_name")
	p, ok := accessLogPath(domainName)
	if !ok {
		http.Error(w, "Invalid domain.", http.StatusBadRequest)
		return
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	f, err := tenantfs.Open(p)
	if err != nil {
		http.Error(w, "Log file not found.", http.StatusNotFound)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+domainName+`-access.log"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if info, err := f.Stat(); err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	}
	_, _ = io.Copy(w, f)
}
