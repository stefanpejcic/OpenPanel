package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func testBulkApp() *appctx.App {
	return &appctx.App{Sessions: session.NewStore([]byte("test-secret-key-0123456789abcdef"))}
}

// flashHandler mimics flashAndRedirect: the outcome only lives in the flash, the status is always 302
func flashHandler(a *appctx.App, category, text string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, category, text)
		_ = a.Sessions.Save(r, w, sess)
		http.Redirect(w, r, "/somewhere", http.StatusFound)
	}
}

func jsonHandler(status int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
}

func TestDispatchResult(t *testing.T) {
	a := testBulkApp()
	cases := []struct {
		name    string
		h       http.Handler
		ok      bool
		message string
	}{
		{"success flash", flashHandler(a, "success", "Deleted <b>x</b>."), true, "Deleted x ."},
		{"error flash on 302", flashHandler(a, "error", "Nope"), false, "Nope"},
		{"danger flash", flashHandler(a, "danger", "Bad"), false, "Bad"},
		{"lone warning", flashHandler(a, "warning", "Config file not found"), false, "Config file not found"},
		{"json error", jsonHandler(200, map[string]string{"error": "broken"}), false, "broken"},
		{"json success false", jsonHandler(200, map[string]any{"success": false, "message": "no"}), false, "no"},
		{"json error_message", jsonHandler(200, map[string]any{"success": false, "error_message": "PID not found"}), false, "PID not found"},
		{"json ok", jsonHandler(200, map[string]any{"message": "fine"}), true, "fine"},
		{"nested result error", jsonHandler(200, map[string]any{"results": []any{map[string]any{"status": "ok"}, map[string]any{"status": "error", "error": "table crashed"}}}), false, "table crashed"},
		{"403 text", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "You do not own this domain.", 403) }), false, "You do not own this domain."},
		{"login redirect", http.RedirectHandler("/login", http.StatusFound), false, "Not authenticated."},
		{"plain 200", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), true, "Done."},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/page/bulk", nil)
		res := Dispatch(a, c.h, req, BulkCall{Method: http.MethodPost, Path: "/x"})
		if res.OK != c.ok || res.Message != c.message {
			t.Errorf("%s: got ok=%v %q, want ok=%v %q", c.name, res.OK, res.Message, c.ok, c.message)
		}
	}
}

func TestDispatchSendsFormQueryAndJSON(t *testing.T) {
	a := testBulkApp()
	var got *http.Request
	var body map[string]string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		if r.Header.Get("Content-Type") == "application/json" {
			_ = json.NewDecoder(r.Body).Decode(&body)
		} else {
			_ = r.ParseForm()
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/page/bulk", nil)
	req.Header.Set("Cookie", "OPENPANEL=abc")

	Dispatch(a, h, req, BulkCall{Method: http.MethodPost, Path: "/x", Form: map[string][]string{"a": {"1"}}, Query: map[string][]string{"q": {"2"}}})
	if got.Form.Get("a") != "1" || got.URL.Query().Get("q") != "2" || got.Header.Get("Cookie") != "OPENPANEL=abc" {
		t.Errorf("form/query/cookie not passed: %v %v %q", got.Form, got.URL.RawQuery, got.Header.Get("Cookie"))
	}
	Dispatch(a, h, req, BulkCall{Method: http.MethodDelete, Path: "/x", JSON: map[string]string{"link": "/y"}})
	if got.Method != http.MethodDelete || body["link"] != "/y" {
		t.Errorf("json call not passed: %s %v", got.Method, body)
	}
}

func TestBulkFlashMessageEscapes(t *testing.T) {
	if cleanMessage("line one\nline two") != "line one" {
		t.Error("cleanMessage should keep the first line")
	}
	if got := BulkFlashCategory([]BulkResult{{OK: true}, {OK: false}}); got != "warning" {
		t.Errorf("mixed results category = %q", got)
	}
	if got := BulkFlashCategory([]BulkResult{{OK: false}}); got != "danger" {
		t.Errorf("all failed category = %q", got)
	}
}

func TestBulkInputInitial(t *testing.T) {
	if (&BulkInput{Options: []BulkOption{{Value: "8.3"}, {Value: "8.2"}}}).Initial() != "8.3" {
		t.Error("select should start on its first option")
	}
	if (&BulkInput{Default: "/var/www/html"}).Initial() != "/var/www/html" {
		t.Error("default should win")
	}
}

// flashes left in the user's cookie by an earlier request must not count as this item's result
func TestDispatchIgnoresOldFlashes(t *testing.T) {
	a := testBulkApp()
	old := httptest.NewRecorder()
	flashHandler(a, "danger", "old failure").ServeHTTP(old, httptest.NewRequest(http.MethodGet, "/", nil))
	req := httptest.NewRequest(http.MethodPost, "/page/bulk", nil)
	for _, c := range old.Result().Cookies() {
		req.AddCookie(c)
	}
	res := Dispatch(a, flashHandler(a, "success", "fresh"), req, BulkCall{Method: http.MethodPost, Path: "/x"})
	if !res.OK || res.Message != "fresh" {
		t.Errorf("got ok=%v %q, want the new success flash", res.OK, res.Message)
	}
	res = Dispatch(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), req, BulkCall{Method: http.MethodPost, Path: "/x"})
	if !res.OK {
		t.Errorf("handler that set no flash should succeed, got %q", res.Message)
	}
}

func TestBulkActionUsable(t *testing.T) {
	if (BulkAction{Input: &BulkInput{Type: "select"}}).Usable() {
		t.Error("empty select should not be usable")
	}
	if !(BulkAction{}).Usable() || !(BulkAction{Input: &BulkInput{Type: "text"}}).Usable() {
		t.Error("plain and text actions are usable")
	}
}

// the real request's session must not collect the replayed handlers' flashes, only FinishBulk's summary
func TestDispatchKeepsFlashesOutOfRealSession(t *testing.T) {
	a := testBulkApp()
	req := httptest.NewRequest(http.MethodPost, "/page/bulk", nil)
	// middleware usually loads the session first, which caches it in the request context
	real, _ := a.Sessions.Get(req, session.CookieName)
	for _, id := range []string{"1", "2"} {
		Dispatch(a, flashHandler(a, "success", "Query "+id+" killed."), req, BulkCall{Method: http.MethodPost, Path: "/x"})
	}
	if n := len(real.Flashes()); n != 0 {
		t.Errorf("real session got %d per-item flashes, want 0", n)
	}
}

func TestBulkItemList(t *testing.T) {
	if got := bulkItemList([]BulkResult{{Item: "3038"}, {Item: "<b>x</b>"}}); got != "3038, &lt;b&gt;x&lt;/b&gt;." {
		t.Errorf("got %q", got)
	}
	many := make([]BulkResult, 25)
	for i := range many {
		many[i].Item = strconv.Itoa(i)
	}
	if got := bulkItemList(many); !strings.HasSuffix(got, "19, +5.") {
		t.Errorf("cap not applied: %q", got)
	}
}
