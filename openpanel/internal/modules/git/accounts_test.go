package git

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/githubapi"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const goodToken = "github_pat_0123456789abcdefghij"

// fakeGitHub answers the handful of API calls the module makes and remembers what it was asked
type fakeGitHub struct {
	mu       sync.Mutex
	calls    map[string]int
	hookURLs []string
	deleted  []string
}

func (f *fakeGitHub) count(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[key]
}

func withFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()
	f := &fakeGitHub{calls: map[string]int{}}
	mux := http.NewServeMux()
	authed := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f.mu.Lock()
			f.calls[r.Method+" "+r.URL.Path]++
			f.mu.Unlock()
			if r.Header.Get("Authorization") != "Bearer "+goodToken {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("GET /user", authed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("github-authentication-token-expiration", "2099-01-02 03:04:05 UTC")
		_, _ = w.Write([]byte(`{"login":"octo"}`))
	}))
	mux.HandleFunc("GET /user/repos", authed(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`[{"full_name":"acme/site","private":true,"default_branch":"trunk"},{"full_name":"octo/blog"},{"full_name":"bad name/../x"}]`))
	}))
	mux.HandleFunc("GET /repos/acme/site", authed(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"full_name":"acme/site","private":true,"default_branch":"trunk"}`))
	}))
	mux.HandleFunc("GET /repos/acme/site/branches", authed(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"name":"trunk"},{"name":"dev"},{"name":"bad..name"}]`))
	}))
	mux.HandleFunc("POST /repos/acme/site/hooks", authed(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Events []string          `json:"events"`
			Config map[string]string `json:"config"`
		}
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		f.mu.Lock()
		f.hookURLs = append(f.hookURLs, body.Config["url"])
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":77}`))
	}))
	mux.HandleFunc("DELETE /repos/acme/site/hooks/{id}", authed(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.deleted = append(f.deleted, r.PathValue("id"))
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("/", authed(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }))
	srv := httptest.NewServer(mux)
	old := githubapi.APIBase
	githubapi.APIBase = srv.URL
	t.Cleanup(func() { githubapi.APIBase = old; srv.Close(); hostedCache.Clear() })
	return f
}

func TestConnectAccount(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	withFakeGitHub(t)
	ctx := context.Background()

	if _, err := connectAccount(ctx, userContext, "not a token", ""); err == nil {
		t.Fatal("a malformed token should be refused before any API call")
	}
	if _, err := connectAccount(ctx, userContext, "github_pat_wrongwrongwrongwrong", ""); err == nil || !strings.Contains(err.Error(), "did not accept") {
		t.Fatalf("a token GitHub rejects should say so, got %v", err)
	}
	acct, err := connectAccount(ctx, userContext, "  "+goodToken+"\n", " Work ")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if acct.Login != "octo" || acct.Title() != "Work" || acct.Expires != "2099-01-02 03:04:05 UTC" {
		t.Fatalf("stored account = %+v", acct)
	}
	if fi, err := os.Stat(accountsPath(userContext)); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("accounts.json should be 0600, got %v %v", fi, err)
	}
	if _, err := connectAccount(ctx, userContext, goodToken, ""); err == nil || !strings.Contains(err.Error(), "already connected") {
		t.Fatalf("the same token twice should be refused, got %v", err)
	}

	accounts, _ := loadAccounts(userContext)
	views, _ := json.Marshal(accountViews(accounts, []Repo{{Name: "a", AccountID: acct.ID}, {Name: "b"}}))
	if strings.Contains(string(views), goodToken) {
		t.Fatal("the token must never reach the UI")
	}
	if !strings.Contains(string(views), `"repos":1`) || !strings.Contains(string(views), `"token_expires":"2099-01-02"`) {
		t.Fatalf("views = %s", views)
	}

	oldMax := gitMaxAccounts
	gitMaxAccounts = 1
	t.Cleanup(func() { gitMaxAccounts = oldMax })
	updateAccounts(userContext, func([]Account) ([]Account, error) {
		return []Account{{ID: "other0000001", Provider: ProviderGitHub, Login: "someone", Token: "github_pat_anotheranotheranother"}}, nil
	})
	if _, err := connectAccount(ctx, userContext, goodToken, ""); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("connecting past git_max_accounts should be refused, got %v", err)
	}
}

func TestHostedListing(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	gh := withFakeGitHub(t)
	ctx := context.Background()
	acct, err := connectAccount(ctx, userContext, goodToken, "")
	if err != nil {
		t.Fatal(err)
	}

	repos, truncated, err := listHostedRepos(ctx, userContext, acct, false)
	if err != nil || truncated || len(repos) != 2 || repos[0].FullName != "acme/site" || !repos[0].Private {
		t.Fatalf("repos = %+v truncated=%v err=%v", repos, truncated, err)
	}
	if _, _, err := listHostedRepos(ctx, userContext, acct, false); err != nil || gh.count("GET /user/repos") != 1 {
		t.Fatalf("second listing should come from the cache, API hit %d times", gh.count("GET /user/repos"))
	}
	if _, _, err := listHostedRepos(ctx, userContext, acct, true); err != nil || gh.count("GET /user/repos") != 2 {
		t.Fatal("refresh should ask GitHub again")
	}

	branches, err := listHostedBranches(ctx, acct, "acme/site")
	if err != nil || strings.Join(branches, ",") != "trunk,dev" {
		t.Fatalf("branches = %v err=%v", branches, err)
	}
	if _, err := listHostedBranches(ctx, acct, "../etc/passwd"); err == nil {
		t.Fatal("a made up repository name should not reach the API")
	}
	if _, err := hostedRepo(ctx, acct, "octo/missing"); err == nil || !strings.Contains(err.Error(), "could not find") {
		t.Fatalf("a repo the token can't see should be reported, got %v", err)
	}
}

// tokenGitServer is dumbGitServer behind the Basic auth git sends for an account repo
func tokenGitServer(bare string, seen *[]string) http.Handler {
	files := http.StripPrefix("/acme/site.git", http.FileServer(http.Dir(bare)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		*seen = append(*seen, user+":"+pass)
		if user != "x-access-token" || pass != goodToken {
			w.Header().Set("WWW-Authenticate", `Basic realm="x"`)
			http.Error(w, "auth", http.StatusUnauthorized)
			return
		}
		files.ServeHTTP(w, r)
	})
}

func TestAccountRepoClonesWithToken(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	withFakeGitHub(t)
	ctx := context.Background()
	oldOK := remoteAddrOK
	remoteAddrOK = func(net.IP) bool { return true }
	t.Cleanup(func() { remoteAddrOK = oldOK })

	upstream := t.TempDir()
	gitIn(t, upstream, "init", "-q")
	gitIn(t, upstream, "symbolic-ref", "HEAD", "refs/heads/trunk")
	os.WriteFile(filepath.Join(upstream, "a.txt"), []byte("1"), 0o644)
	gitIn(t, upstream, "add", ".")
	gitIn(t, upstream, "commit", "-q", "-m", "one")
	bare := filepath.Join(t.TempDir(), "site.git")
	gitIn(t, upstream, "clone", "-q", "--bare", upstream, bare)
	gitIn(t, bare, "update-server-info")
	var seen []string
	srv := httptest.NewServer(tokenGitServer(bare, &seen))
	defer srv.Close()
	oldBase := githubCloneBase
	githubCloneBase = srv.URL + "/"
	t.Cleanup(func() { githubCloneBase = oldBase })

	acct, err := connectAccount(ctx, userContext, goodToken, "")
	if err != nil {
		t.Fatal(err)
	}
	r := Repo{Name: "site", Source: SourceRemote, AccountID: acct.ID, RemoteRepo: "acme/site", RemoteURL: githubCloneBase + "acme/site.git", TargetDir: "/var/www/html/example.com", DeployMode: DeployManual}
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatalf("clone with the account token: %v (server saw %v)", err, seen)
	}
	if r.Branch != "trunk" {
		t.Fatalf("branch = %q", r.Branch)
	}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })

	// the token is looked up on every fetch and never written into the repo
	cfg, _ := os.ReadFile(filepath.Join(repoDir(userContext, "site"), "config"))
	stored, _ := os.ReadFile(storePath(userContext))
	if strings.Contains(string(cfg), goodToken) || strings.Contains(string(stored), goodToken) {
		t.Fatal("the token leaked into the repo config or repos.json")
	}

	elsewhere := r
	elsewhere.RemoteURL = "https://example.com/acme/site.git"
	if _, err := remoteEnv(ctx, userContext, elsewhere); err == nil {
		t.Fatal("an account token must not be sent to another host")
	}

	if _, used, err := disconnectAccount(ctx, userContext, acct.ID); err != nil || used != 1 {
		t.Fatalf("disconnect: used=%d err=%v", used, err)
	}
	if _, err := fetchRepo(ctx, userContext, r); !errors.Is(err, errAccountGone) {
		t.Fatalf("pull after disconnect should name the missing account, got %v", err)
	}
	if accounts, _ := loadAccounts(userContext); len(accounts) != 0 {
		t.Fatalf("account still stored: %+v", accounts)
	}
}

func TestWebhookIsAddedAndRemoved(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	gh := withFakeGitHub(t)
	ctx := context.Background()
	acct, err := connectAccount(ctx, userContext, goodToken, "")
	if err != nil {
		t.Fatal(err)
	}
	r := Repo{Name: "site", Source: SourceRemote, AccountID: acct.ID, RemoteRepo: "acme/site", RemoteURL: githubCloneBase + "acme/site.git", DeployMode: DeployAutomatic, WebhookToken: "tok123"}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })

	for _, base := range []string{"http://panel.example.com:2083/", "https://203.0.113.5:2083/", "https://localhost:2083/"} {
		if added, err := ensureHook(ctx, userContext, "bob", base, r); added || !errors.Is(err, errHookUnreachable) {
			t.Fatalf("%s: added=%v err=%v, GitHub can't deliver there", base, added, err)
		}
	}
	if gh.count("POST /repos/acme/site/hooks") != 0 {
		t.Fatal("no hook should be requested for an unreachable panel")
	}

	added, err := ensureHook(ctx, userContext, "bob", "https://panel.example.com:2083/", r)
	if err != nil || !added {
		t.Fatalf("ensureHook: added=%v err=%v", added, err)
	}
	if len(gh.hookURLs) != 1 || gh.hookURLs[0] != "https://panel.example.com:2083/git-webhook/bob/tok123" {
		t.Fatalf("hook url = %v", gh.hookURLs)
	}
	saved, _ := getRepo(userContext, "site")
	if saved.HookID != 77 {
		t.Fatalf("hook id not stored: %+v", saved)
	}
	if added, err := ensureHook(ctx, userContext, "bob", "https://panel.example.com:2083/", saved); added || err != nil || gh.count("POST /repos/acme/site/hooks") != 1 {
		t.Fatal("a repo that already has its hook should be left alone")
	}

	dropHook(ctx, userContext, saved)
	if len(gh.deleted) != 1 || gh.deleted[0] != "77" {
		t.Fatalf("deleted = %v", gh.deleted)
	}
	if saved, _ = getRepo(userContext, "site"); saved.HookID != 0 {
		t.Fatal("hook id should be cleared")
	}

	// a repo added by URL never gets a hook
	plain := Repo{Name: "plain", Source: SourceRemote, RemoteURL: "https://github.com/acme/site.git"}
	if added, err := ensureHook(ctx, userContext, "bob", "https://panel.example.com:2083/", plain); added || err != nil {
		t.Fatalf("url repo: added=%v err=%v", added, err)
	}
}

func TestSettingsOnAccountRepo(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	withFakeGitHub(t)
	ctx := context.Background()
	first, err := connectAccount(ctx, userContext, goodToken, "one")
	if err != nil {
		t.Fatal(err)
	}
	// a second entry with the same working token stands in for another account
	second := Account{ID: "second000001", Provider: ProviderGitHub, Login: "octo", Label: "two", Token: goodToken, AddedAt: nowStr()}
	dead := Account{ID: "dead00000001", Provider: ProviderGitHub, Login: "ghost", Token: "github_pat_revokedrevokedrevoked", AddedAt: nowStr()}
	updateAccounts(userContext, func(a []Account) ([]Account, error) { return append(a, second, dead), nil })

	r := Repo{Name: "site", Source: SourceRemote, AccountID: first.ID, RemoteRepo: "acme/site", RemoteURL: githubCloneBase + "acme/site.git", Branch: "trunk", TargetDir: "/var/www/html/example.com", DeployMode: DeployManual, HookID: 77}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })

	// the settings form of a url repo sends these, on an account repo they must not stick
	if err := updateSettings(ctx, userContext, "site", url.Values{"remote_user": {"mallory"}, "remote_pass": {"x"}, "deploy_mode": {DeployNone}}); err != nil {
		t.Fatalf("settings: %v", err)
	}
	saved, _ := getRepo(userContext, "site")
	if saved.RemoteUser != "" || saved.RemotePass != "" || saved.AccountID != first.ID || saved.DeployMode != DeployNone {
		t.Fatalf("after settings: %+v", saved)
	}

	if err := updateSettings(ctx, userContext, "site", url.Values{"account": {dead.ID}}); err == nil {
		t.Fatal("an account that can't read the repo should be refused")
	}
	if err := updateSettings(ctx, userContext, "site", url.Values{"account": {second.ID}}); err != nil {
		t.Fatalf("relink: %v", err)
	}
	saved, _ = getRepo(userContext, "site")
	if saved.AccountID != second.ID || saved.HookID != 0 {
		t.Fatalf("after relink: %+v, the old account's hook should be gone", saved)
	}

	accounts, _ := loadAccounts(userContext)
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) {
		return append(rs, Repo{Name: "lost", RemoteRepo: "acme/lost", AccountID: "nobody"}, Repo{Name: "plain", RemoteURL: "https://gitlab.com/u/p.git"}), nil
	})
	repos, _ := loadRepos(userContext)
	links := repoLinks(accounts, repos)
	if len(links) != 2 || links["site"].Account != "two" || links["site"].AccountGone || !links["lost"].AccountGone {
		t.Fatalf("repo links = %+v", links)
	}
}

func TestRenderGitPageAccounts(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "git": true}
	layout := web.LayoutData{
		Title: "Git", BrandName: "Test Panel", CSRFToken: "tok", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(allowed, nil, "/git"), UserAllowed: allowed,
		UserAllowedJSON: web.UserAllowedList(allowed), CurrentUsername: "bob",
		RequestPath: "/git", AdminPort: "2087", T: mgr.Translator("en"),
	}
	data := GitPageData{
		LayoutData: layout, Username: "bob",
		Accounts: []AccountView{{ID: "acct00000001", Provider: ProviderGitHub, Login: "octo", Title: "Work", AddedAt: "2026-11-28", Expires: "2026-12-01", ExpiresSoon: true, Repos: 1}},
		Repos: []RepoView{
			{Name: "site", Source: SourceRemote, RemoteURL: "https://github.com/acme/site.git", RemoteRepo: "acme/site", AccountID: "acct00000001", Account: "Work", HookAdded: true, Branch: "trunk", DeployMode: DeployAutomatic, WebhookURL: "https://x/git-webhook/bob/t"},
			{Name: "orphan", Source: SourceRemote, RemoteURL: "https://github.com/acme/old.git", RemoteRepo: "acme/old", AccountID: "gone", AccountGone: true, Branch: "main", DeployMode: DeployManual, WebhookURL: "https://x/git-webhook/bob/u"},
			{Name: "plain", Source: SourceRemote, RemoteURL: "https://gitlab.com/u/plain.git", Branch: "main", DeployMode: DeployManual, WebhookURL: "https://x/git-webhook/bob/v"},
		},
		BulkActions: gitBulkActions(mgr.Translator("en"), web.ThemeClassic),
	}
	w := httptest.NewRecorder()
	if err := gitPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		`id="github-accounts"`, `id="git-accounts"`, `id="account-acct00000001"`, `name="action" value="account_add"`, `name="action" value="account_remove"`,
		`id="via-github"`, `id="github_repo"`, `id="remote_url"`, "Token expires soon", "Account disconnected", "Disconnected, pick an account",
		"This webhook was added to the repository on GitHub for you", "Add this URL as a push webhook",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}
	// only the url repo keeps the username and password fields
	if n := strings.Count(body, `name="remote_user" value=`); n != 1 {
		t.Errorf("remote_user settings field rendered %d times, want 1", n)
	}
	if n := strings.Count(body, `<select name="account"`); n != 2 {
		t.Errorf("account select rendered %d times in settings, want 2", n)
	}
}
