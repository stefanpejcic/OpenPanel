package account

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"

	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func bearer(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const (
	roToken = "op_mcp_readonly000000000000000000000000000000000"
	rwToken = "op_mcp_writable000000000000000000000000000000000"
)

func TestReadOnlyMCPTokenCantManageTokens(t *testing.T) {
	ctxDir := featureContext(t, "mcp", "api", "favorites")
	a, store := newFakeApp(t, nil, map[int]*fakeUser{7: {username: "alice", email: "alice@example.test", context: ctxDir}})
	store.addToken(7, roToken, true)
	ownerID := store.addToken(7, rwToken, false)

	mux := http.NewServeMux()
	RegisterMCPAPI(mux, a)
	RegisterMCPEndpoint(mux, a)
	apiregistry.Handle(mux, a, "favorites", "POST /api/testwrite", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := auth.LoadUser(a)(mux)

	if rec := bearer(h, "POST", "/api/account/mcp", roToken, `{"name":"x","read_only":false}`); rec.Code != http.StatusForbidden {
		t.Fatalf("read-only token minting a token: got %d %s, want 403", rec.Code, rec.Body)
	}
	if rec := bearer(h, "DELETE", "/api/account/mcp/"+itoa(ownerID), roToken, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("read-only token revoking a token: got %d, want 403", rec.Code)
	}
	rec := bearer(h, "POST", "/mcp", roToken, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"account_mcp_post","arguments":{"body":{"name":"x","read_only":false}}}}`)
	if !strings.Contains(rec.Body.String(), "read-only") || strings.Contains(rec.Body.String(), "op_mcp_") {
		t.Fatalf("tools/call write tool with read-only token should be refused, got %s", rec.Body)
	}
	if store.tokenCount(7) != 2 {
		t.Fatalf("read-only token changed the token list, now %d tokens", store.tokenCount(7))
	}

	// read-only still reads, and the transport itself still works
	if rec := bearer(h, "GET", "/api/account/mcp", roToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("read-only GET: got %d", rec.Code)
	}
	if rec := bearer(h, "POST", "/mcp", roToken, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "tools") {
		t.Fatalf("read-only tools/list: got %d %s", rec.Code, rec.Body)
	}
	// a writable token can still manage tokens
	if rec := bearer(h, "POST", "/api/account/mcp", rwToken, `{"name":"y","read_only":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("writable token minting: got %d %s, want 201", rec.Code, rec.Body)
	}
}

func TestPasswordChangeRevokesMCPTokensAndResetDropsPasskeys(t *testing.T) {
	ctxDir := featureContext(t, "mcp", "api", "favorites")
	a, store := newFakeApp(t, nil, map[int]*fakeUser{2: {username: "victim", email: "victim@example.test", context: ctxDir}})
	store.addToken(2, rwToken, false)
	store.passkeys[2] = 1

	whoami := auth.LoadUser(a)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, _ := auth.UserID(r)
		w.Write([]byte(itoa(uid)))
	}))
	if got := bearer(whoami, "GET", "/", rwToken, "").Body.String(); got != "2" {
		t.Fatalf("token should resolve before the change, got %q", got)
	}

	sess := &sessions.Session{Values: map[interface{}]interface{}{}}
	if reason := updatePasswordByID(context.Background(), a, sess, 2, "Tq8!vLm#Zr4@Wx9$Pb2^"); reason != "" {
		t.Fatalf("updatePasswordByID: %s", reason)
	}
	if store.tokenCount(2) != 0 {
		t.Fatalf("mcp tokens survived the password change")
	}
	if got := bearer(whoami, "GET", "/", rwToken, "").Body.String(); got == "2" {
		t.Fatalf("old mcp token still authenticates after password change")
	}

	revokePasskeysOnRecovery(context.Background(), a, 2)
	if store.passkeys[2] != 0 {
		t.Fatalf("passkeys survived the reset")
	}
}

func postReset(h http.Handler, ip, email string) *httptest.ResponseRecorder {
	form := url.Values{"email": {email}}
	req := httptest.NewRequest("POST", "/reset_password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = ip + ":40000"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestResetRequestThrottledAndUniform(t *testing.T) {
	port, rcpts := smtpSink(t)
	cfg := config.Config{"mail_server": "127.0.0.1", "mail_port": port, "mail_default_sender": "panel@example.test"}
	a, _ := newFakeApp(t, cfg, map[int]*fakeUser{2: {username: "victim", email: "victim@example.com"}})
	mux := http.NewServeMux()
	RegisterPasswordReset(mux, a)

	// typed address only matches case-insensitively, the mail goes to the stored one
	known := postReset(mux, "198.51.100.1", "VICTIM@example.com")
	select {
	case got := <-rcpts:
		if got != "victim@example.com" {
			t.Fatalf("reset mail went to %q, want the stored address", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("no reset mail sent")
	}

	// a second request inside the cooldown doesn't send again
	postReset(mux, "198.51.100.2", "victim@example.com")
	select {
	case got := <-rcpts:
		t.Fatalf("second reset mail sent inside the cooldown to %q", got)
	case <-time.After(300 * time.Millisecond):
	}

	unknown := postReset(mux, "198.51.100.3", "nobody@example.com")
	if unknown.Code != known.Code || unknown.Body.String() != known.Body.String() {
		t.Fatalf("unknown email response differs from known one, the form leaks which emails exist")
	}

	for i := 0; i < 5; i++ {
		if rec := postReset(mux, "198.51.100.9", "nobody@example.com"); rec.Code != http.StatusOK {
			t.Fatalf("request %d from one IP: got %d", i+1, rec.Code)
		}
	}
	if rec := postReset(mux, "198.51.100.9", "nobody@example.com"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th request from one IP within a minute: got %d, want 429", rec.Code)
	}
}

func TestPasskeyLoginChallengeIsSingleUse(t *testing.T) {
	a, _ := newFakeApp(t, nil, map[int]*fakeUser{})

	begin := httptest.NewRequest("POST", "/login/passkey/begin", nil)
	begin.Host = "panel.example.test"
	brec := httptest.NewRecorder()
	handlePasskeysLoginBegin(a, brec, begin)
	if brec.Code != http.StatusOK {
		t.Fatalf("begin: got %d %s", brec.Code, brec.Body)
	}
	cookies := brec.Result().Cookies()

	complete := func() string {
		req := httptest.NewRequest("POST", "/login/passkey/complete", strings.NewReader(`{}`))
		req.Host = "panel.example.test"
		for _, c := range cookies {
			req.AddCookie(c)
		}
		rec := httptest.NewRecorder()
		handlePasskeysLoginComplete(a, rec, req)
		return rec.Body.String()
	}
	// first use gets past the challenge lookup and fails on the bogus assertion instead
	if first := complete(); strings.Contains(first, "Login session expired") {
		t.Fatalf("first complete didn't find the challenge: %s", first)
	}
	// replaying the same begin cookie finds nothing
	if second := complete(); !strings.Contains(second, "Login session expired") {
		t.Fatalf("replayed begin cookie still carried a usable challenge: %s", second)
	}

	if got := takePasskeyCeremony(context.Background(), a, "auth", ""); got != "" {
		t.Fatalf("empty ceremony id returned %q", got)
	}
}

func TestBearerTokensFollowTwofaEnforcement(t *testing.T) {
	ctxDir := featureContext(t, "mcp", "api", "favorites", "twofa")
	a, store := newFakeApp(t, nil, map[int]*fakeUser{
		7: {username: "noenroll", context: ctxDir},
		8: {username: "enrolled", context: ctxDir, twofa: true},
	})
	a.TwofaEnforce = true
	a.SetEnabledModulesForTest("twofa", "mcp", "api", "favorites")
	store.addToken(7, rwToken, false)
	enrolledToken := "op_mcp_enrolled000000000000000000000000000000000"
	store.addToken(8, enrolledToken, false)

	mux := http.NewServeMux()
	apiregistry.Handle(mux, a, "favorites", "POST /api/testwrite", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	apiregistry.Handle(mux, a, "twofa", "GET /api/testtwofa", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := auth.LoadUser(a)(mux)

	if rec := bearer(h, "POST", "/api/testwrite", rwToken, `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("un-enrolled token under twofa_enforce: got %d, want 403", rec.Code)
	}
	if rec := bearer(h, "GET", "/api/testtwofa", rwToken, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("un-enrolled token should still reach 2fa enrollment: got %d", rec.Code)
	}
	if rec := bearer(h, "POST", "/api/testwrite", enrolledToken, `{}`); rec.Code != http.StatusNoContent {
		t.Fatalf("enrolled token: got %d, want 204", rec.Code)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestAPILoginRejectsNonJSONBodies(t *testing.T) {
	a, _ := newFakeApp(t, nil, map[int]*fakeUser{})
	mux := http.NewServeMux()
	RegisterAPILogin(mux, a)

	// what a cross-site <form enctype="text/plain"> sends
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"username":"u","password":"p","x":"="}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("text/plain /api/login: got %d, want 415", rec.Code)
	}
	if strings.Contains(rec.Header().Get("Set-Cookie"), "OPENPANEL=") {
		t.Fatalf("/api/login set the browser session cookie")
	}
}

func TestEmailChangeEndsOtherSessions(t *testing.T) {
	a, _ := newFakeApp(t, nil, map[int]*fakeUser{})
	ctx := context.Background()
	for _, k := range []string{"session:2:here", "session:2:other", "session:3:someoneelse"} {
		a.Cache.Raw().HSet(ctx, k, "user_id", "x")
	}

	if n := clearOtherUserSessions(ctx, a, 2, "here"); n != 1 {
		t.Fatalf("cleared %d sessions, want 1", n)
	}
	for k, want := range map[string]bool{"session:2:here": true, "session:2:other": false, "session:3:someoneelse": true} {
		n, _ := a.Cache.Raw().Exists(ctx, k).Result()
		if (n == 1) != want {
			t.Fatalf("%s exists=%v, want %v", k, n == 1, want)
		}
	}
}
