package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

const (
	modifyToken = "op_mcp_modifytoken0000000000000000000000000000"
	dnsToken    = "op_mcp_dnsonlytoken000000000000000000000000000"
	fullToken   = "op_mcp_fullscopedtoken0000000000000000000000000"
)

// scopedMux is a panel with a few fake routes behind the real RequireAPI
func scopedMux(t *testing.T) (http.Handler, *fakeStore) {
	t.Helper()
	h, store, _ := scopedMuxApp(t)
	return h, store
}

// scopedMuxApp is scopedMux plus the app, extra adds features to the account
func scopedMuxApp(t *testing.T, extra ...string) (http.Handler, *fakeStore, *appctx.App) {
	t.Helper()
	ctxDir := featureContext(t, append([]string{"mcp", "api", "favorites", "dns", "websites"}, extra...)...)
	a, store := newFakeApp(t, nil, map[int]*fakeUser{7: {username: "alice", email: "alice@example.test", context: ctxDir}})
	store.scoped = true
	mux := http.NewServeMux()
	RegisterMCPAPI(mux, a)
	RegisterMCPEndpoint(mux, a)
	ok := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }
	apiregistry.Handle(mux, a, "favorites", "POST /api/testwrite", ok)
	apiregistry.Handle(mux, a, "favorites", "DELETE /api/testwrite/{id}", ok)
	apiregistry.Handle(mux, a, "favorites", "POST /api/testwrite/{id}/restore", ok)
	apiregistry.Handle(mux, a, "dns", "POST /api/testdns", ok)
	apiregistry.Handle(mux, a, "websites", "GET /api/sites/wp-cli/{action}", ok)
	apiregistry.Handle(mux, a, "websites", "GET /api/sites/{rest...}", ok)
	return auth.LoadUser(a)(mux), store, a
}

func TestModifyTokenBlocksDestructiveRoutes(t *testing.T) {
	h, store := scopedMux(t)
	store.addScopedToken(7, modifyToken, "agent", "modify", "")

	if rec := bearer(h, "POST", "/api/testwrite", modifyToken, `{}`); rec.Code != http.StatusNoContent {
		t.Fatalf("modify token create: got %d %s", rec.Code, rec.Body)
	}
	if rec := bearer(h, "DELETE", "/api/testwrite/3", modifyToken, ""); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "delete") {
		t.Fatalf("modify token delete: got %d %s, want 403", rec.Code, rec.Body)
	}
	// destructive by path even though it's a POST
	if rec := bearer(h, "POST", "/api/testwrite/3/restore", modifyToken, `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("modify token restore: got %d, want 403", rec.Code)
	}
	if rec := bearer(h, "POST", "/api/account/mcp", modifyToken, `{"name":"wider"}`); rec.Code != http.StatusForbidden {
		t.Fatalf("modify token minting a token: got %d, want 403", rec.Code)
	}
	if _, ok := store.tokenByName("wider"); ok {
		t.Fatal("modify token managed to create a token")
	}
}

func TestFeatureLimitedTokenStaysInItsFeatures(t *testing.T) {
	h, store := scopedMux(t)
	store.addScopedToken(7, dnsToken, "dns bot", "full", "dns")

	if rec := bearer(h, "POST", "/api/testdns", dnsToken, `{}`); rec.Code != http.StatusNoContent {
		t.Fatalf("dns token on dns route: got %d %s", rec.Code, rec.Body)
	}
	if rec := bearer(h, "POST", "/api/testwrite", dnsToken, `{}`); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "favorites") {
		t.Fatalf("dns token on favorites route: got %d %s, want 403", rec.Code, rec.Body)
	}
	// the transport itself still works
	if rec := bearer(h, "POST", "/mcp", dnsToken, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`); rec.Code != http.StatusOK {
		t.Fatalf("dns token tools/list: got %d %s", rec.Code, rec.Body)
	}
}

// apiJWT is the classic API login token, the way scripts make MCP tokens now that tokens can't
func apiJWT(t *testing.T, a *appctx.App, userID int) string {
	t.Helper()
	jwt, err := mintAPIToken(context.Background(), a, userID)
	if err != nil {
		t.Fatalf("mint jwt: %v", err)
	}
	return jwt
}

func TestCreateScopedTokenViaAPI(t *testing.T) {
	h, store, a := scopedMuxApp(t)
	jwt := apiJWT(t, a, 7)

	rec := bearer(h, "POST", "/api/account/mcp", jwt, `{"name":"deployer","mode":"modify","features":["dns"," dns","favorites"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d %s", rec.Code, rec.Body)
	}
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out["mode"] != "modify" || out["read_only"] != false {
		t.Errorf("response = %v", out)
	}
	tok, ok := store.tokenByName("deployer")
	if !ok || tok.mode != "modify" || tok.feats != "dns,favorites" || tok.ro {
		t.Errorf("stored token = %+v", tok)
	}

	// the new token is limited the way it was asked for
	raw, _ := out["token"].(string)
	if rec := bearer(h, "DELETE", "/api/testwrite/1", raw, ""); rec.Code != http.StatusForbidden {
		t.Errorf("new modify token deleting: got %d", rec.Code)
	}

	if rec := bearer(h, "POST", "/api/account/mcp", jwt, `{"name":"x","mode":"admin"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad mode: got %d", rec.Code)
	}
	// only features the account has, a made-up name or one off the plan is refused
	for _, f := range []string{"nope", "docker"} {
		rec := bearer(h, "POST", "/api/account/mcp", jwt, `{"name":"x","features":["dns","`+f+`"]}`)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), f) {
			t.Errorf("feature %s: got %d %s", f, rec.Code, rec.Body)
		}
	}
	// old clients still get a read-only token from read_only alone
	rec = bearer(h, "POST", "/api/account/mcp", jwt, `{"name":"legacy-ro","read_only":true}`)
	if tok, _ := store.tokenByName("legacy-ro"); rec.Code != http.StatusCreated || tok.mode != "readonly" || !tok.ro {
		t.Errorf("legacy read_only: %d %+v", rec.Code, tok)
	}
}

func TestNewTokenWithoutFeaturesGetsTodaysList(t *testing.T) {
	h, store, a := scopedMuxApp(t, "account", "twofa", "passkeys", "phpmyadmin")
	rec := bearer(h, "POST", "/api/account/mcp", apiJWT(t, a, 7), `{"name":"everything"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d %s", rec.Code, rec.Body)
	}
	tok, _ := store.tokenByName("everything")
	feats := "," + fmt.Sprint(tok.feats) + ","
	for _, want := range []string{"dns", "favorites", "websites"} {
		if !strings.Contains(feats, ","+want+",") {
			t.Errorf("%s missing from %s", want, feats)
		}
	}
	// the token runs on api and mcp anyway, and the sensitive ones need ticking by hand
	for _, never := range []string{"api", "mcp", "account", "twofa", "passkeys", "phpmyadmin"} {
		if strings.Contains(feats, ","+never+",") {
			t.Errorf("%s got into the default list %s", never, feats)
		}
	}
	// asking for a sensitive one by name works
	rec = bearer(h, "POST", "/api/account/mcp", apiJWT(t, a, 7), `{"name":"2fa","features":["twofa"]}`)
	if tok, _ := store.tokenByName("2fa"); rec.Code != http.StatusCreated || tok.feats != "twofa" {
		t.Errorf("twofa token: %d %+v", rec.Code, tok)
	}
}

func TestScopedTokenNeedsColumns(t *testing.T) {
	h, store, a := scopedMuxApp(t)
	store.scoped = false
	// without the columns there's nowhere to keep the list, so no token at all
	for _, body := range []string{`{"name":"x","mode":"modify"}`, `{"name":"plain"}`} {
		if rec := bearer(h, "POST", "/api/account/mcp", apiJWT(t, a, 7), body); rec.Code != http.StatusConflict {
			t.Fatalf("%s without columns: got %d %s, want 409", body, rec.Code, rec.Body)
		}
	}
}

func TestTokensCantManageTokens(t *testing.T) {
	h, store, a := scopedMuxApp(t)
	ownerID := store.addScopedToken(7, fullToken, "owner", "full", "")
	if rec := bearer(h, "POST", "/api/account/mcp", fullToken, `{"name":"wider"}`); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "can't create or revoke tokens") {
		t.Fatalf("full token minting: got %d %s", rec.Code, rec.Body)
	}
	if rec := bearer(h, "DELETE", "/api/account/mcp/"+itoa(ownerID), fullToken, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("full token revoking: got %d", rec.Code)
	}
	if rec := bearer(h, "GET", "/api/account/mcp", fullToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("listing is still fine: got %d", rec.Code)
	}
	// the user's own login still can
	if rec := bearer(h, "POST", "/api/account/mcp", apiJWT(t, a, 7), `{"name":"y"}`); rec.Code != http.StatusCreated {
		t.Fatalf("jwt minting: got %d %s", rec.Code, rec.Body)
	}
}

func TestOldTokensGetFrozenOnFirstUse(t *testing.T) {
	h, store, _ := scopedMuxApp(t)
	store.addScopedToken(7, fullToken, "old", "full", "")
	if rec := bearer(h, "POST", "/api/testdns", fullToken, `{}`); rec.Code != http.StatusNoContent {
		t.Fatalf("old token on dns: got %d %s", rec.Code, rec.Body)
	}
	tok, _ := store.tokenByName("old")
	feats := "," + fmt.Sprint(tok.feats) + ","
	if tok.feats == nil || !strings.Contains(feats, ",dns,") || !strings.Contains(feats, ",favorites,") || strings.Contains(feats, ",api,") {
		t.Fatalf("old token wasn't pinned to the account's features: %v", tok.feats)
	}
}

func TestEmptyListAllowsNothing(t *testing.T) {
	h, store, _ := scopedMuxApp(t)
	id := store.addScopedToken(7, dnsToken, "empty", "full", "dns")
	store.setFeats(id, "")
	if rec := bearer(h, "POST", "/api/testdns", dnsToken, `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("empty list token on dns: got %d %s, want 403", rec.Code, rec.Body)
	}
}

func TestReadOnlyTokenCantRunWPCLIWritesOverGET(t *testing.T) {
	h, store := scopedMux(t)
	store.addScopedToken(7, roToken, "viewer", "readonly", "")
	store.addScopedToken(7, fullToken, "owner", "full", "")
	q := "?website=example.com&docroot=/var/www/html/example.com"
	if rec := bearer(h, "GET", "/api/sites/wp-cli/update_now"+q, roToken, ""); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "read-only") {
		t.Fatalf("read-only token running update_now over GET: got %d %s, want 403", rec.Code, rec.Body)
	}
	if rec := bearer(h, "GET", "/api/sites/wp-cli/site_info"+q, roToken, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("read-only token reading site_info: got %d %s", rec.Code, rec.Body)
	}
	if rec := bearer(h, "GET", "/api/sites/wp-cli/update_now"+q, fullToken, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("full token running update_now: got %d %s", rec.Code, rec.Body)
	}
}

func TestOnlyFullTokensReadCredentials(t *testing.T) {
	h, store := scopedMux(t)
	store.addScopedToken(7, roToken, "viewer", "readonly", "")
	store.addScopedToken(7, modifyToken, "agent", "modify", "")
	store.addScopedToken(7, fullToken, "owner", "full", "")
	for _, tok := range []string{roToken, modifyToken} {
		if rec := bearer(h, "GET", "/api/sites/example.com/wp-info", tok, ""); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "database passwords") {
			t.Errorf("limited token reading wp-info: got %d %s, want 403", rec.Code, rec.Body)
		}
	}
	if rec := bearer(h, "GET", "/api/sites/example.com/wp-info", fullToken, ""); rec.Code != http.StatusNoContent {
		t.Errorf("full token reading wp-info: got %d %s", rec.Code, rec.Body)
	}
	// full mode isn't enough without the feature in the list
	store.addScopedToken(7, dnsToken, "dns bot", "full", "dns")
	if rec := bearer(h, "GET", "/api/sites/example.com/wp-info", dnsToken, ""); rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "websites") {
		t.Errorf("dns-only full token reading wp-info: got %d %s, want 403", rec.Code, rec.Body)
	}
}
