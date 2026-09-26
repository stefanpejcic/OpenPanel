package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// BulkAction is one button in the shared bulk-actions bar (partials/_bulk.html)
type BulkAction struct {
	Key     string
	Label   string
	Confirm string // question shown before running it
	Danger  bool
	Input   *BulkInput // asks for a value on confirm, e.g. a new CPU limit
	Feature string     // only offered when the user has this feature
}

// BulkInput describes the value field shown on confirm, Options turns it into a select
type BulkInput struct {
	Type        string // "number", "text", "password" or "select"
	Placeholder string
	Hint        string
	Min         string
	Max         string
	Step        string
	Default     string
	Options     []BulkOption
}

// Usable is false for a select with nothing to pick, like "Assign user" when there are no users
func (b BulkAction) Usable() bool {
	return b.Input == nil || b.Input.Type != "select" || len(b.Input.Options) > 0
}

// Initial is the value the field starts with: Default, or the first option of a select
func (i *BulkInput) Initial() string {
	if i.Default == "" && len(i.Options) > 0 {
		return i.Options[0].Value
	}
	return i.Default
}

type BulkOption struct {
	Value string
	Label string
}

// BulkRequest is what bulk-actions.js posts: the action key, the optional input value and the selected row keys
type BulkRequest struct {
	Action string   `json:"action"`
	Value  string   `json:"value"`
	Items  []string `json:"items"`
}

type BulkResult struct {
	Item    string `json:"item"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// DecodeBulkRequest reads the JSON body and writes a 400 itself when it's unusable
func DecodeBulkRequest(a *appctx.App, w http.ResponseWriter, r *http.Request) (BulkRequest, bool) {
	var req BulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BulkError(a, w, r, "Invalid request body.")
		return req, false
	}
	if len(req.Items) == 0 {
		BulkError(a, w, r, "Nothing selected.")
		return req, false
	}
	return req, true
}

// RunBulk calls fn for every item and collects the results in order
func RunBulk(items []string, fn func(item string) BulkResult) []BulkResult {
	results := make([]BulkResult, 0, len(items))
	for _, item := range items {
		res := fn(item)
		if res.Item == "" {
			res.Item = item
		}
		results = append(results, res)
	}
	return results
}

// FinishBulk flashes one summary banner for the page reload and returns the per-item results as JSON
func FinishBulk(a *appctx.App, w http.ResponseWriter, r *http.Request, actionLabel string, results []BulkResult) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, BulkFlashCategory(results), BulkFlashMessage(a, r, actionLabel, results))
	_ = a.Sessions.Save(r, w, sess)
	writeBulkJSON(w, http.StatusOK, map[string]any{"results": results})
}

func BulkFlashCategory(results []BulkResult) string {
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

func BulkFlashMessage(a *appctx.App, r *http.Request, actionLabel string, results []BulkResult) string {
	var failed []BulkResult
	for _, res := range results {
		if !res.OK {
			failed = append(failed, res)
		}
	}
	total := strconv.Itoa(len(results))
	if len(failed) == 0 {
		return Tr(a, r, "%(action_title)s: completed successfully for all %(total)s selected item(s).", "action_title", actionLabel, "total", total) + " " + bulkItemList(results)
	}
	msg := Tr(a, r, "%(action_title)s: %(failed)s of %(total)s selected item(s) failed.", "action_title", actionLabel, "failed", strconv.Itoa(len(failed)), "total", total)
	for _, res := range failed {
		// flashes render as HTML and items can be user-named
		msg += " " + html.EscapeString(res.Item) + " (" + html.EscapeString(Tr(a, r, res.Message)) + ")."
	}
	return msg
}

// bulkItemList names the items in a success summary, capped so a huge selection stays readable
func bulkItemList(results []BulkResult) string {
	const maxNames = 20
	names := make([]string, 0, maxNames)
	for i, res := range results {
		if i == maxNames {
			names = append(names, "+"+strconv.Itoa(len(results)-maxNames))
			break
		}
		names = append(names, html.EscapeString(res.Item))
	}
	return strings.Join(names, ", ") + "."
}

// BulkError rejects the whole request with a translated message the bulk bar shows as a toast, kv fills its %(name)s placeholders
func BulkError(a *appctx.App, w http.ResponseWriter, r *http.Request, msg string, kv ...any) {
	writeBulkJSON(w, http.StatusBadRequest, map[string]string{"error": Tr(a, r, msg, kv...)})
}

func writeBulkJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// BulkCall is one in-process request to an existing route, JSON is sent as the body when set, otherwise Form
type BulkCall struct {
	Label  string // shown in the summary instead of the item key
	Method string
	Path   string
	Form   url.Values
	Query  url.Values
	JSON   any
}

var tagRE = regexp.MustCompile(`<[^>]*>`)

// Dispatch runs call against h with the current request's session, so a bulk action goes through the same handler, checks and messages as the single-item button
func Dispatch(a *appctx.App, h http.Handler, r *http.Request, call BulkCall) BulkResult {
	target := call.Path
	if len(call.Query) > 0 {
		target += "?" + call.Query.Encode()
	}
	var body io.Reader = strings.NewReader("")
	contentType := ""
	switch {
	case call.JSON != nil:
		b, _ := json.Marshal(call.JSON)
		body, contentType = bytes.NewReader(b), "application/json"
	case len(call.Form) > 0:
		body, contentType = strings.NewReader(call.Form.Encode()), "application/x-www-form-urlencoded"
	}

	req := httptest.NewRequest(call.Method, target, body)
	req = req.WithContext(withoutSessionCache{r.Context()})
	req.RemoteAddr = r.RemoteAddr
	req.Host = r.Host
	for _, k := range []string{"Cookie", "X-Forwarded-For", "User-Agent", "Referer", "Accept-Language"} {
		if v := r.Header.Get(k); v != "" {
			req.Header.Set(k, v)
		}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return dispatchResult(a, rec, len(sessionFlashes(a, r.Cookies())))
}

// withoutSessionCache hides gorilla/sessions' per-request session cache, so a replayed handler loads its own copy of the session instead of adding its flashes to the real request's one
type withoutSessionCache struct{ context.Context }

func (c withoutSessionCache) Value(key any) any {
	if t := reflect.TypeOf(key); t != nil && t.PkgPath() == "github.com/gorilla/sessions" {
		return nil
	}
	return c.Context.Value(key)
}

// sessionFlashes decodes the flashes stored in the session cookie among cookies
func sessionFlashes(a *appctx.App, cookies []*http.Cookie) []flash.Message {
	var out []flash.Message
	for _, c := range cookies {
		if c.Name != session.CookieName {
			continue
		}
		fake := httptest.NewRequest(http.MethodGet, "/", nil)
		fake.AddCookie(c)
		if sess, err := a.Sessions.New(fake, session.CookieName); err == nil {
			for _, f := range sess.Flashes() {
				if m, ok := f.(flash.Message); ok {
					out = append(out, m)
				}
			}
		}
	}
	return out
}

// dispatchResult reads the outcome the way a browser would see it: the flash the handler set, its JSON reply, or the status code, skipping the first oldFlashes that were already in the user's cookie
func dispatchResult(a *appctx.App, rec *httptest.ResponseRecorder, oldFlashes int) BulkResult {
	if strings.Contains(rec.Header().Get("Location"), "/login") {
		return BulkResult{Message: "Not authenticated."}
	}

	// flashAndRedirect handlers report errors only in the flash, the status is a plain 302
	var flashes []flash.Message
	for _, c := range rec.Result().Cookies() {
		if c.Name != session.CookieName {
			continue
		}
		// each Save rewrites the whole cookie, so the first one holds the new flash before any render pops it
		if got := sessionFlashes(a, []*http.Cookie{c}); len(got) > oldFlashes {
			flashes = append(flashes, got[oldFlashes:]...)
		}
	}
	var warning string
	hasSuccess := false
	for _, f := range flashes {
		switch f.Category {
		case "error", "danger":
			return BulkResult{Message: cleanMessage(f.Text)}
		case "warning":
			warning = f.Text
		case "success":
			hasSuccess = true
		}
	}
	// some handlers flash a warning instead of an error for things like a missing config file
	if warning != "" && !hasSuccess {
		return BulkResult{Message: cleanMessage(warning)}
	}

	raw := strings.TrimSpace(rec.Body.String())
	var obj map[string]any
	if json.Unmarshal([]byte(raw), &obj) == nil {
		if e, _ := obj["error"].(string); e != "" {
			return BulkResult{Message: cleanMessage(e)}
		}
		if ok, isBool := obj["success"].(bool); isBool && !ok {
			msg, _ := obj["error_message"].(string)
			if msg == "" {
				msg = fmt.Sprint(obj["message"])
			}
			return BulkResult{Message: cleanMessage(msg)}
		}
		if s, _ := obj["status"].(string); s == "error" {
			return BulkResult{Message: cleanMessage(fmt.Sprint(obj["message"]))}
		}
		// per-part results, like OPTIMIZE on every table of a database
		if list, _ := obj["results"].([]any); len(list) > 0 {
			for _, item := range list {
				if m, _ := item.(map[string]any); m != nil && m["status"] == "error" {
					return BulkResult{Message: cleanMessage(fmt.Sprint(m["error"]))}
				}
			}
		}
	}
	if rec.Code >= 400 {
		msg := raw
		if obj != nil {
			msg = fmt.Sprint(obj["message"])
		}
		if msg == "" || msg == "<nil>" {
			msg = http.StatusText(rec.Code)
		}
		return BulkResult{Message: cleanMessage(msg)}
	}

	msg := "Done."
	if len(flashes) > 0 {
		msg = cleanMessage(flashes[len(flashes)-1].Text)
	} else if m, _ := obj["message"].(string); m != "" {
		msg = cleanMessage(m)
	}
	return BulkResult{OK: true, Message: msg}
}

// cleanMessage turns handler output into one short plain-text line
func cleanMessage(s string) string {
	s = strings.TrimSpace(html.UnescapeString(tagRE.ReplaceAllString(s, " ")))
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 200 {
		s = s[:200] + "..."
	}
	return s
}

// BulkRoute maps one selected item to the existing route that handles it, or returns nil and the item's result when nothing needs to run
type BulkRoute func(action, value, item string) (*BulkCall, BulkResult)

// Skip is a BulkRoute result for an item that can't run the action
func Skip(msg string) (*BulkCall, BulkResult) { return nil, BulkResult{Message: msg} }

// Call is a BulkRoute result that replays call
func Call(call BulkCall) (*BulkCall, BulkResult) { return &call, BulkResult{} }

// ServeBulkDispatch is a complete POST /<page>/bulk handler: validate the action, then replay each item through its single-item route on h
func ServeBulkDispatch(a *appctx.App, h http.Handler, w http.ResponseWriter, r *http.Request, actions []BulkAction, route BulkRoute) {
	req, ok := DecodeBulkRequest(a, w, r)
	if !ok {
		return
	}
	act, found := FindBulkAction(actions, req.Action)
	if !found {
		BulkError(a, w, r, "Unknown bulk action.")
		return
	}
	// passwords keep their spaces
	if act.Input == nil || act.Input.Type != "password" {
		req.Value = strings.TrimSpace(req.Value)
	}
	if act.Input != nil && req.Value == "" {
		BulkError(a, w, r, "A value is required for this action.")
		return
	}
	results := RunBulk(req.Items, func(item string) BulkResult {
		call, res := route(req.Action, req.Value, item)
		if call == nil {
			return res
		}
		res = Dispatch(a, h, r, *call)
		res.Item = call.Label
		return res
	})
	FinishBulk(a, w, r, act.Label, results)
}

// AllowedBulkActions drops actions whose Feature the user doesn't have
func AllowedBulkActions(a *appctx.App, r *http.Request, actions []BulkAction) []BulkAction {
	allowed := map[string]bool{}
	if userID, ok := auth.UserID(r); ok {
		if injected, err := a.InjectData(r.Context(), userID); err == nil {
			list, _ := injected["user_allowed"].([]string)
			for _, f := range list {
				allowed[f] = true
			}
		}
	}
	out := actions[:0:0]
	for _, act := range actions {
		if act.Feature == "" || allowed[act.Feature] {
			out = append(out, act)
		}
	}
	return out
}

func FindBulkAction(actions []BulkAction, key string) (BulkAction, bool) {
	for _, act := range actions {
		if act.Key == key {
			return act, true
		}
	}
	return BulkAction{}, false
}
