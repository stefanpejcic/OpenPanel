package git

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestValidators(t *testing.T) {
	urls := map[string]bool{
		"https://github.com/user/repo.git":       true,
		"http://git.example.com:8080/a/b":        true,
		"git@github.com:user/repo.git":           true,
		"ssh://git@gitlab.com:2222/group/r.git":  true,
		"file:///etc":                            false,
		"/home/other/git/x.git":                  false,
		"ext::sh -c touch% /tmp/x":               false,
		"--upload-pack=touch /tmp/x":             false,
		"https://github.com/../etc":              false,
		"https://user:pass@github.com/u/r.git":   false,
		"git@github.com:user/repo.git; rm -rf /": false,
	}
	for u, want := range urls {
		if got := isValidRemoteURL(u); got != want {
			t.Errorf("isValidRemoteURL(%q) = %v, want %v", u, got, want)
		}
	}

	names := map[string]bool{"site": true, "example.com": true, "a_b-c": true, "": false, ".hidden": false, "x.git": false, "a/b": false, "a..b": false, "-x": false}
	for n, want := range names {
		if got := isValidRepoName(n); got != want {
			t.Errorf("isValidRepoName(%q) = %v, want %v", n, got, want)
		}
	}

	branches := map[string]bool{"main": true, "feature/x": true, "v1.2": true, "-x": false, "a..b": false, "a/": false, "x.lock": false, "a b": false}
	for b, want := range branches {
		if got := isValidBranch(b); got != want {
			t.Errorf("isValidBranch(%q) = %v, want %v", b, got, want)
		}
	}

	dirs := map[string]string{"/var/www/html/example.com": "/var/www/html/example.com", "example.com/public": "/var/www/html/example.com/public", "/var/www/html": "/var/www/html", "/etc": "", "/var/www/html/../x": "", "/var/www/html2": ""}
	for in, want := range dirs {
		got, ok := cleanTargetDir(in)
		if (want == "") == ok || got != want {
			t.Errorf("cleanTargetDir(%q) = %q,%v want %q", in, got, ok, want)
		}
	}
}

func TestRepoNameFromURL(t *testing.T) {
	for in, want := range map[string]string{
		"https://github.com/user/my-site.git": "my-site",
		"git@github.com:user/blog.git":        "blog",
		"https://gitlab.com/group/app/":       "app",
	} {
		if got := repoNameFromURL(in); got != want {
			t.Errorf("repoNameFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitHTTPPath(t *testing.T) {
	user, name, rest, ok := splitHTTPPath("/git-http/bob/site.git/info/refs")
	if !ok || user != "bob" || name != "site" || rest != "info/refs" {
		t.Fatalf("got %q %q %q %v", user, name, rest, ok)
	}
	for _, p := range []string{"/git-http/bob/site/info/refs", "/git-http/bob/..git/info/refs", "/git-http/bob"} {
		if _, _, _, ok := splitHTTPPath(p); ok {
			t.Errorf("expected %q to be rejected", p)
		}
	}
}

func TestRedact(t *testing.T) {
	if got := redact("fatal: unable to access 'https://bob:secret@github.com/x.git/'"); strings.Contains(got, "secret") {
		t.Errorf("redact left the password in: %q", got)
	}
}

// gitIn runs a plain git command in a work dir for test setup
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@x", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@x", "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// withFakeAccount points repos and the html volume at temp dirs and stubs out chown/podman
func withFakeAccount(t *testing.T) (userContext, htmlVolume string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	tmp := t.TempDir()
	rootOverride = filepath.Join(tmp, "git")
	userContext = "tester"
	htmlVolume = filepath.Join(tmp, "html")
	if err := os.MkdirAll(filepath.Join(htmlVolume, userContext), 0o755); err != nil {
		t.Fatal(err)
	}
	restoreHome := paths.SetHomeVolumeForTest(htmlVolume)
	oldChown, oldActions := chownTree, runActions
	chownTree = func(context.Context, string, string) error { return nil }
	t.Cleanup(func() { rootOverride = ""; chownTree, runActions = oldChown, oldActions; restoreHome() })
	return userContext, filepath.Join(htmlVolume, userContext)
}

func TestLocalRepoDeployFlow(t *testing.T) {
	userContext, volume := withFakeAccount(t)
	ctx := context.Background()

	r := Repo{Name: "site", Source: SourceLocal, Branch: "main", TargetDir: "/var/www/html/example.com", DeployMode: DeployAutomatic}
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatalf("initRepo: %v", err)
	}
	if err := updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil }); err != nil {
		t.Fatal(err)
	}
	if _, _, err := deployRepo(ctx, userContext, r); err == nil {
		t.Fatal("deploying an empty repo should fail")
	}

	// push like a client would, straight to the bare repo
	work := t.TempDir()
	gitIn(t, work, "init", "-q", "-b", "main")
	os.WriteFile(filepath.Join(work, "index.html"), []byte("v1"), 0o644)
	os.WriteFile(filepath.Join(work, "old.txt"), []byte("old"), 0o644)
	gitIn(t, work, "add", ".")
	gitIn(t, work, "commit", "-q", "-m", "first")
	gitIn(t, work, "push", "-q", repoDir(userContext, "site"), "main")

	var actionsRan string
	runActions = func(_ context.Context, _, container, workdir, script string) (string, error) {
		actionsRan = container + "|" + workdir + "|" + script
		return "ok", nil
	}
	r.Actions, r.ActionsContainer = "echo hi", "php-fpm-8.3"
	commit, _, err := deployRepo(ctx, userContext, r)
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "example.com", "index.html")); string(got) != "v1" {
		t.Fatalf("index.html = %q", got)
	}
	if _, err := os.Stat(filepath.Join(volume, "example.com", ".git")); err == nil {
		t.Fatal(".git must not be deployed into the docroot")
	}
	if actionsRan != "php-fpm-8.3|/var/www/html/example.com|echo hi" {
		t.Fatalf("actions got %q", actionsRan)
	}

	// second push deletes a file and changes another, deploy has to follow both
	os.Remove(filepath.Join(work, "old.txt"))
	os.WriteFile(filepath.Join(work, "index.html"), []byte("v2"), 0o644)
	gitIn(t, work, "add", "-A")
	gitIn(t, work, "commit", "-q", "-m", "second")
	gitIn(t, work, "push", "-q", repoDir(userContext, "site"), "main")
	// untracked user files in the docroot are left alone
	os.WriteFile(filepath.Join(volume, "example.com", "uploads.txt"), []byte("keep"), 0o644)

	commit2, _, err := deployRepo(ctx, userContext, r)
	if err != nil || commit2 == commit {
		t.Fatalf("second deploy: %v (%s)", err, commit2)
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "example.com", "index.html")); string(got) != "v2" {
		t.Fatalf("index.html = %q after second deploy", got)
	}
	if _, err := os.Stat(filepath.Join(volume, "example.com", "old.txt")); err == nil {
		t.Fatal("old.txt should have been removed")
	}
	if _, err := os.Stat(filepath.Join(volume, "example.com", "uploads.txt")); err != nil {
		t.Fatal("untracked file was removed")
	}

	commits, err := commitLog(ctx, userContext, r, 10)
	if err != nil || len(commits) != 2 || commits[0].Subject != "second" {
		t.Fatalf("log = %+v, %v", commits, err)
	}

	if err := deleteRepo(userContext, "site"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(repoDir(userContext, "site")); err == nil {
		t.Fatal("bare repo still on disk")
	}
	if _, err := os.Stat(filepath.Join(volume, "example.com", "index.html")); err != nil {
		t.Fatal("deployed files should stay after removing the repo")
	}
}

func TestRemoteRepoCloneAndPull(t *testing.T) {
	userContext, volume := withFakeAccount(t)
	ctx := context.Background()

	// the remote is served over http so isValidRemoteURL is happy and the real transport is used
	upstream := t.TempDir()
	gitIn(t, upstream, "init", "-q", "-b", "trunk")
	os.WriteFile(filepath.Join(upstream, "a.txt"), []byte("1"), 0o644)
	gitIn(t, upstream, "add", ".")
	gitIn(t, upstream, "commit", "-q", "-m", "one")
	bare := filepath.Join(t.TempDir(), "up.git")
	gitIn(t, upstream, "clone", "-q", "--bare", upstream, bare)
	gitIn(t, bare, "update-server-info")
	srv := httptest.NewServer(dumbGitServer(bare))
	defer srv.Close()

	r := Repo{Name: "up", Source: SourceRemote, RemoteURL: srv.URL + "/up.git", TargetDir: "/var/www/html/example.com", DeployMode: DeployAutomatic}
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if r.Branch != "trunk" {
		t.Fatalf("branch = %q, want the remote default", r.Branch)
	}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })

	gitIn(t, upstream, "checkout", "-q", "-b", "feature")
	os.WriteFile(filepath.Join(upstream, "a.txt"), []byte("2"), 0o644)
	gitIn(t, upstream, "commit", "-qam", "two")
	gitIn(t, upstream, "push", "-q", bare, "trunk", "feature")
	gitIn(t, bare, "update-server-info")

	if err := pullAndRecord(ctx, userContext, r); err != nil {
		t.Fatalf("pull: %v", err)
	}
	if got := listBranches(ctx, userContext, "up"); strings.Join(got, ",") != "feature,trunk" {
		t.Fatalf("branches = %v", got)
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "example.com", "a.txt")); string(got) != "1" {
		t.Fatalf("a.txt = %q, trunk should be deployed", got)
	}
	saved, _ := getRepo(userContext, "up")
	if saved.LastDeployed == "" || saved.LastError != "" {
		t.Fatalf("status not recorded: %+v", saved)
	}
}

func TestRenderGitPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "git": true}
	layout := web.LayoutData{
		Title: "Git", BrandName: "Test Panel", CSRFToken: "tok", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(allowed, nil, "/git"), UserAllowed: allowed,
		UserAllowedJSON: web.UserAllowedList(allowed), CurrentUsername: "bob",
		RequestPath: "/git", AdminPort: "2087", T: mgr.Translator("en"),
	}
	latest := &Commit{Hash: "abc", Short: "abc", Subject: "init"}
	data := GitPageData{
		LayoutData: layout, Username: "bob", Containers: []string{"php-fpm-8.3"}, PublicKey: "ssh-ed25519 AAAA",
		Repos: []RepoView{
			{Name: "site", Source: SourceLocal, Branch: "main", Domain: "example.com", TargetDir: "/var/www/html/example.com", DeployMode: DeployAutomatic, CloneURL: "https://x/git-http/bob/site.git", Latest: latest, BulkSkip: "pull"},
			{Name: "blog", Source: SourceRemote, RemoteURL: "https://github.com/u/blog.git", Branch: "trunk", Branches: []string{"trunk", "dev"}, DeployMode: DeployManual, WebhookURL: "https://x/git-webhook/bob/t", Actions: "echo", ActionsContainer: "php-fpm-8.3", LastError: "boom"},
		},
		pageExtras:  pageExtras{NewPassword: "pw123", NewPasswordRepo: "site"},
		BulkActions: gitBulkActions(mgr.Translator("en")),
	}
	w := httptest.NewRecorder()
	if err := gitPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{`id="add-repository"`, "pw123", "https://x/git-http/bob/site.git", "https://x/git-webhook/bob/t", `id="repo-blog"`, `id="git-table"`, `id="bulk-actions-bar"`, `data-bulk-action="mode"`, `data-bulk-skip="pull"`, "bi-git"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}
}

// dumbGitServer serves a bare repo over git's dumb http protocol, enough for clone and fetch
func dumbGitServer(bare string) http.Handler {
	return http.StripPrefix("/up.git", http.FileServer(http.Dir(bare)))
}

func TestSmartHTTPPushAndClone(t *testing.T) {
	userContext, volume := withFakeAccount(t)
	ctx := context.Background()
	oldAccount := gitAccount
	gitAccount = func(_ context.Context, _ *appctx.App, user string) (string, bool) { return userContext, user == "bob" }
	t.Cleanup(func() { gitAccount = oldAccount })

	r := Repo{Name: "site", Source: SourceLocal, Branch: "main", TargetDir: "/var/www/html/example.com", DeployMode: DeployAutomatic}
	pw := "s3cretpass"
	r.PushPassHash, _ = hashPassword(pw)
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatal(err)
	}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { handleSmartHTTP(nil, w, req) }))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	work := t.TempDir()
	gitIn(t, work, "init", "-q", "-b", "main")
	os.WriteFile(filepath.Join(work, "index.html"), []byte("pushed"), 0o644)
	gitIn(t, work, "add", ".")
	gitIn(t, work, "commit", "-q", "-m", "over http")

	bad := exec.Command("git", "push", "-q", "http://bob:wrong@"+host+"/git-http/bob/site.git", "main")
	bad.Dir = work
	bad.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "HOME="+work)
	if out, err := bad.CombinedOutput(); err == nil {
		t.Fatalf("push with a wrong password should fail: %s", out)
	}

	gitIn(t, work, "push", "-q", "http://bob:"+pw+"@"+host+"/git-http/bob/site.git", "main")

	// automatic mode deploys in the background after the push
	deployed := false
	for i := 0; i < 50 && !deployed; i++ {
		if got, _ := os.ReadFile(filepath.Join(volume, "example.com", "index.html")); string(got) == "pushed" {
			deployed = true
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !deployed {
		t.Fatal("push did not trigger a deployment")
	}

	clone := filepath.Join(t.TempDir(), "c")
	gitIn(t, work, "clone", "-q", "http://bob:"+pw+"@"+host+"/git-http/bob/site.git", clone)
	if got, _ := os.ReadFile(filepath.Join(clone, "index.html")); string(got) != "pushed" {
		t.Fatalf("clone has %q", got)
	}

	// dumb protocol paths must not leak objects
	resp, err := http.Get(srv.URL + "/git-http/bob/site.git/HEAD")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal("dumb HTTP path should be refused")
		}
	}
}

func TestRenderGitLog(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "git": true}
	layout := web.LayoutData{Title: "Git", BrandName: "Test", CSRFToken: "tok", PanelDir: "ltr", UserAllowed: allowed,
		UserAllowedJSON: web.UserAllowedList(allowed), RequestPath: "/git", T: mgr.Translator("en")}
	view := RepoView{Name: "site", Branch: "main", TargetDir: "/var/www/html/x", LastDeployed: "abc123", LastDeploy: "2026-10-12 19:37:00 UTC", LastDeployISO: "2026-10-12T19:37:00Z"}
	data := GitPageData{LayoutData: layout, Repos: []RepoView{view}, LogView: &view,
		Log: []Commit{{Hash: "abc123", Short: "abc123", Author: "dev", Date: "2026-10-12 19:30", ISO: "2026-10-12T19:30:00Z", Subject: "first"}}}
	w := httptest.NewRecorder()
	if err := gitPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Commit log", `data-relative-date="2026-10-12T19:30:00Z"`, `data-relative-date="2026-10-12T19:37:00Z"`, "deployed"} {
		if !strings.Contains(body, want) {
			t.Errorf("log page missing %q", want)
		}
	}
	if strings.Contains(body, `id="add-repository"`) {
		t.Error("log page should not show the repository list header")
	}
}

// the site tab posts FormData, which is multipart, and it has to reach the same action
func TestParseFormMultipart(t *testing.T) {
	var body strings.Builder
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("action", "deploy")
	_ = mw.WriteField("repo", "site")
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/git/action.json", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", mw.FormDataContentType())
	parseForm(req)
	if req.Form.Get("action") != "deploy" || req.Form.Get("repo") != "site" {
		t.Fatalf("multipart fields lost: %v", req.Form)
	}

	req = httptest.NewRequest(http.MethodPost, "/git/action.json", strings.NewReader("action=pull&repo=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	parseForm(req)
	if req.Form.Get("action") != "pull" {
		t.Fatalf("urlencoded fields lost: %v", req.Form)
	}
}

func TestRunCLI(t *testing.T) {
	userContext, volume := withFakeAccount(t)
	ctx := context.Background()
	r := Repo{Name: "cli", Source: SourceLocal, Branch: "main", TargetDir: "/var/www/html/cli.example.com", DeployMode: DeployManual}
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatal(err)
	}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })
	work := t.TempDir()
	gitIn(t, work, "init", "-q", "-b", "main")
	os.WriteFile(filepath.Join(work, "index.html"), []byte("cli"), 0o644)
	gitIn(t, work, "add", ".")
	gitIn(t, work, "commit", "-q", "-m", "one")
	gitIn(t, work, "push", "-q", repoDir(userContext, "cli"), "main")

	var out strings.Builder
	if code := RunCLI([]string{"deploy", userContext, "cli"}, &out); code != 0 {
		t.Fatalf("deploy exit %d: %s", code, out.String())
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "cli.example.com", "index.html")); string(got) != "cli" {
		t.Fatalf("not deployed: %q", got)
	}
	out.Reset()
	if code := RunCLI([]string{"pull", userContext, "cli"}, &out); code != 1 || !strings.Contains(out.String(), "local repository") {
		t.Fatalf("pull on a local repo should fail, got %d %q", code, out.String())
	}
	for _, bad := range [][]string{{"deploy", "../etc", "cli"}, {"deploy", userContext, "../x"}, {"nuke", userContext, "cli"}, {"deploy"}} {
		out.Reset()
		if code := RunCLI(bad, &out); code == 0 {
			t.Errorf("RunCLI(%v) should fail", bad)
		}
	}
}
