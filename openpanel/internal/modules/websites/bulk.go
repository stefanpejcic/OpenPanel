package websites

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// bulkItem is one selected row from the /sites table, as posted by the
// bulk-actions bottom bar.
type bulkItem struct {
	ID       int    `json:"id"`
	SiteName string `json:"site_name"`
	Type     string `json:"type"`
	Docroot  string `json:"docroot"`
}

type bulkRequest struct {
	Action string     `json:"action"`
	Sites  []bulkItem `json:"sites"`
}

type bulkResult struct {
	SiteName string `json:"site_name"`
	OK       bool   `json:"ok"`
	Message  string `json:"message"`
}

// cmsRemoveTypes covers every type with its own per-type uninstall route
// (POST /<type>/remove, form field "id") - the same set $isCMS uses in
// sites.html, since every CMS there already has a working Delete tab.
var cmsRemoveTypes = map[string]bool{
	"wordpress": true, "joomla": true, "opencart": true, "nextcloud": true,
	"prestashop": true, "drupal": true, "matomo": true, "moodle": true,
	"mediawiki": true, "flarum": true, "sofawiki": true,
}

// cmsBackupTypes covers every type with a working GET /<type>/backup/run
// route - same set as cmsRemoveTypes, all 11 modules have backups.go.
var cmsBackupTypes = cmsRemoveTypes

// cmsUpdateTypes covers only the types with a real one-click "Update now"
// CLI flow (POST /<type>/update?domain=&docroot=). WordPress/Joomla/
// OpenCart/PrestaShop are browser-link-only by design, and PM2 apps
// (nodejs/python/ruby) require a version/requirements form, not a simple
// bulk update.
var cmsUpdateTypes = map[string]bool{
	"drupal": true, "nextcloud": true, "matomo": true, "moodle": true, "mediawiki": true, "flarum": true,
}

func isPM2Type(typeLower string) bool {
	return strings.Contains(typeLower, "nodejs") || strings.Contains(typeLower, "python") || strings.Contains(typeLower, "ruby")
}

// handleSitesBulk runs Update/Backup/Detach/Delete across multiple selected
// sites at once. It dispatches to the exact same per-type routes each
// manager page's own Update/Backup/Delete buttons already call, reusing
// the current request's authenticated context (session cookie + CSRF
// already verified once for this request) - one internal, in-process HTTP
// call per selected site, no new per-type logic duplicated here.
func handleSitesBulk(a *appctx.App, mux *http.ServeMux, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req bulkRequest
	if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if len(req.Sites) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No sites selected"})
		return
	}

	results := make([]bulkResult, 0, len(req.Sites))
	for _, item := range req.Sites {
		domainRoot, _ := splitDomainAndFolder(item.SiteName)
		if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
			results = append(results, bulkResult{SiteName: item.SiteName, OK: false, Message: "You do not own this domain."})
			continue
		}
		results = append(results, dispatchBulkItem(a, mux, r, req.Action, item))
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "ran bulk action '"+req.Action+"' on "+itoa(len(req.Sites))+" site(s)", reqip.ClientIP(r))
	flashSess(a, w, r, bulkFlashCategory(results), bulkFlashMessage(req.Action, results))
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

// dispatchBulkItem routes one selected site's action to the existing
// per-type handler already registered on mux, by replaying the current
// (already-authenticated) request's context into a synthetic in-process
// request - it never leaves the process and never touches the network.
func dispatchBulkItem(a *appctx.App, mux *http.ServeMux, r *http.Request, action string, item bulkItem) bulkResult {
	typeLower := strings.ToLower(item.Type)

	switch action {
	case "detach":
		return internalDispatch(mux, r, item.SiteName, "POST", "/sites/detach", url.Values{"id": {itoa(item.ID)}}, nil)

	case "delete":
		switch {
		case isPM2Type(typeLower):
			return internalDispatch(mux, r, item.SiteName, "POST", "/pm2/delete/"+item.SiteName, nil, nil)
		case typeLower == "websitebuilder":
			return internalDispatch(mux, r, item.SiteName, "POST", "/website-builder/remove", url.Values{"id": {itoa(item.ID)}}, nil)
		case cmsRemoveTypes[typeLower]:
			return internalDispatch(mux, r, item.SiteName, "POST", "/"+typeLower+"/remove", url.Values{"id": {itoa(item.ID)}}, nil)
		default:
			return bulkResult{SiteName: item.SiteName, OK: false, Message: "Delete is not supported for this site type via bulk actions; use Detach instead."}
		}

	case "update":
		if !cmsUpdateTypes[typeLower] {
			return bulkResult{SiteName: item.SiteName, OK: false, Message: "One-click update is not available for this site type."}
		}
		q := url.Values{"domain": {item.SiteName}, "docroot": {item.Docroot}}
		return internalDispatch(mux, r, item.SiteName, "POST", "/"+typeLower+"/update", nil, q)

	case "backup":
		if !cmsBackupTypes[typeLower] {
			return bulkResult{SiteName: item.SiteName, OK: false, Message: "Backups are not available for this site type."}
		}
		q := url.Values{"docroot": {item.Docroot}, "backup_database": {"true"}, "backup_files": {"true"}}
		return internalDispatch(mux, r, item.SiteName, "GET", "/"+typeLower+"/backup/run/"+item.SiteName, nil, q)

	default:
		return bulkResult{SiteName: item.SiteName, OK: false, Message: "Unknown bulk action."}
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

// bulkFlashCategory/bulkFlashMessage turn a batch of per-site results into
// the one flash banner shown after the page reload that follows a bulk
// action - "success"/"warning"/"danger" drive the same flash styling
// every other redirect-based action in this app already uses.
func bulkFlashCategory(results []bulkResult) string {
	failed := 0
	for _, res := range results {
		if !res.OK {
			failed++
		}
	}
	switch {
	case failed == 0:
		return "success"
	case failed == len(results):
		return "danger"
	default:
		return "warning"
	}
}

func bulkFlashMessage(action string, results []bulkResult) string {
	var failed []bulkResult
	for _, res := range results {
		if !res.OK {
			failed = append(failed, res)
		}
	}
	actionTitle := strings.ToUpper(action[:1]) + action[1:]
	if len(failed) == 0 {
		return actionTitle + ": completed successfully for all " + itoa(len(results)) + " selected site(s)."
	}

	msg := actionTitle + ": " + itoa(len(failed)) + " of " + itoa(len(results)) + " selected site(s) failed."
	for _, res := range failed {
		msg += " " + res.SiteName + " (" + res.Message + ")."
	}
	return msg
}

// ndjsonErrorRE pulls the value out of a `{"error":"..."}` line from one of
// the streamed-ndjson update handlers (drupal/nextcloud/matomo/moodle/
// mediawiki/flarum's update.go) - good enough for the single-line JSON
// objects those handlers emit, without pulling in a JSON decoder per line.
var ndjsonErrorRE = regexp.MustCompile(`"error"\s*:\s*"((?:[^"\\]|\\.)*)"`)

// summarizeResult extracts the one line worth showing a human for a
// dispatched action's raw response body: the actual error message from an
// ndjson status/error stream if one was reported, otherwise just the first
// line (e.g. a redirect target or a plain-text success message).
func summarizeResult(body string) string {
	if m := ndjsonErrorRE.FindStringSubmatch(body); m != nil {
		msg := strings.ReplaceAll(m[1], `\"`, `"`)
		msg = strings.ReplaceAll(msg, `\n`, " ")
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return msg
	}
	line := body
	if idx := strings.IndexByte(line, '\n'); idx != -1 {
		line = line[:idx]
	}
	if len(line) > 150 {
		line = line[:150]
	}
	return line
}

// internalDispatch replays r's already-authenticated session (cookie,
// context, remote address) into a brand-new in-process request against
// path, served directly by mux - the same mux every route in this app is
// already registered on; it never leaves the process and never touches
// the network. form is sent as an application/x-www-form-urlencoded body
// (for routes reading r.FormValue), query is appended to the URL (for
// routes reading r.URL.Query()); either may be nil.
func internalDispatch(mux *http.ServeMux, r *http.Request, siteName, method, path string, form, query url.Values) bulkResult {
	target := path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	var body *strings.Reader
	if len(form) > 0 {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}

	req := httptest.NewRequest(method, target, body)
	req = req.WithContext(r.Context())
	req.RemoteAddr = r.RemoteAddr
	for _, h := range []string{"Cookie", "X-Forwarded-For", "User-Agent", "Referer"} {
		if v := r.Header.Get(h); v != "" {
			req.Header.Set(h, v)
		}
	}
	if len(form) > 0 {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if loc := rec.Header().Get("Location"); strings.Contains(loc, "/login") {
		return bulkResult{SiteName: siteName, OK: false, Message: "Internal dispatch was not authenticated."}
	}

	raw := strings.TrimSpace(rec.Body.String())
	// Routes that succeed via flashAndRedirect (302 to /sites or /website)
	// carry the real status in the flash cookie, not the body - treat any
	// non-login redirect or 2xx as success unless the body is an ndjson
	// stream that reported an "error" key itself.
	ok := rec.Code < 400 && !strings.Contains(raw, `"error"`)
	msg := summarizeResult(raw)
	if ok && msg == "" {
		msg = "Done."
	}
	return bulkResult{SiteName: siteName, OK: ok, Message: msg}
}

// handleSitesBulkAPI is the /api/sites/bulk entry point. Update/Backup/
// Delete reuse per-type routes that are gated by session login (not API
// bearer tokens), so only Detach - already session-free, a plain DB-row
// deletion - works here; the others report a clear per-item error instead
// of silently failing.
func handleSitesBulkAPI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	var req bulkRequest
	if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if len(req.Sites) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No sites selected"})
		return
	}

	if req.Action != "detach" {
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "Only the 'detach' bulk action is currently available via the API; use the web UI for update/backup/delete.",
		})
		return
	}

	results := make([]bulkResult, 0, len(req.Sites))
	for _, item := range req.Sites {
		domainRoot, _ := splitDomainAndFolder(item.SiteName)
		if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
			results = append(results, bulkResult{SiteName: item.SiteName, OK: false, Message: "You do not own this domain."})
			continue
		}
		if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", item.ID); delErr != nil {
			results = append(results, bulkResult{SiteName: item.SiteName, OK: false, Message: "An error occurred during detachment."})
			continue
		}
		_ = a.Cache.Delete(ctx, "get_user_websites:"+itoa(userID))
		results = append(results, bulkResult{SiteName: item.SiteName, OK: true, Message: "Detached successfully!"})
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}
