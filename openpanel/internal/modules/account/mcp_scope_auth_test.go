package account

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

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
	ctxDir := featureContext(t, "mcp", "api", "favorites", "dns")
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
	return auth.LoadUser(a)(mux), store
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

func TestCreateScopedTokenViaAPI(t *testing.T) {
	h, store := scopedMux(t)
	store.addScopedToken(7, fullToken, "owner", "full", "")

	rec := bearer(h, "POST", "/api/account/mcp", fullToken, `{"name":"deployer","mode":"modify","features":["dns"," dns","favorites"]}`)
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

	if rec := bearer(h, "POST", "/api/account/mcp", fullToken, `{"name":"x","mode":"admin"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad mode: got %d", rec.Code)
	}
	if rec := bearer(h, "POST", "/api/account/mcp", fullToken, `{"name":"x","features":["nope"]}`); rec.Code != http.StatusBadRequest {
		t.Errorf("unknown feature: got %d", rec.Code)
	}
	// old clients still get a read-only token from read_only alone
	rec = bearer(h, "POST", "/api/account/mcp", fullToken, `{"name":"legacy-ro","read_only":true}`)
	if tok, _ := store.tokenByName("legacy-ro"); rec.Code != http.StatusCreated || tok.mode != "readonly" || !tok.ro {
		t.Errorf("legacy read_only: %d %+v", rec.Code, tok)
	}
}

func TestScopedTokenNeedsColumns(t *testing.T) {
	h, store := scopedMux(t)
	store.scoped = false
	store.addToken(7, rwToken, false)
	if rec := bearer(h, "POST", "/api/account/mcp", rwToken, `{"name":"x","mode":"modify"}`); rec.Code != http.StatusConflict {
		t.Fatalf("modify without columns: got %d %s, want 409", rec.Code, rec.Body)
	}
	if rec := bearer(h, "POST", "/api/account/mcp", rwToken, `{"name":"plain"}`); rec.Code != http.StatusCreated {
		t.Fatalf("plain token without columns: got %d %s", rec.Code, rec.Body)
	}
}
