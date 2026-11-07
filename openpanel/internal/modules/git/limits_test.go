package git

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func TestPushOverStorageLimitIsRefused(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	ctx := context.Background()
	oldAccount, oldLimit := gitAccount, gitStorageLimit
	gitAccount = func(_ context.Context, _ *appctx.App, user string) (string, bool) { return userContext, user == "bob" }
	t.Cleanup(func() { gitAccount, gitStorageLimit = oldAccount, oldLimit })

	r := Repo{Name: "site", Source: SourceLocal, Branch: "main", TargetDir: "/var/www/html/example.com", DeployMode: DeployManual}
	pw := "s3cretpass"
	r.PushPassHash, _ = hashPassword(pw)
	if err := initRepo(ctx, userContext, &r); err != nil {
		t.Fatal(err)
	}
	updateRepos(userContext, func(rs []Repo) ([]Repo, error) { return append(rs, r), nil })
	gitStorageLimit = gitUsage(userContext) + 256<<10

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { handleSmartHTTP(nil, w, req) }))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	work := t.TempDir()
	initMain(t, work)
	blob := make([]byte, 2<<20)
	rand.Read(blob)
	os.WriteFile(filepath.Join(work, "big.bin"), blob, 0o644)
	gitIn(t, work, "add", ".")
	gitIn(t, work, "commit", "-q", "-m", "big")

	push := exec.Command("git", "push", "-q", "http://bob:"+pw+"@"+host+"/git-http/bob/site.git", "main")
	push.Dir = work
	push.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "HOME="+work)
	if out, err := push.CombinedOutput(); err == nil {
		t.Fatalf("a 2 MiB push over a 256 KiB allowance went through: %s", out)
	}
	if gitUsage(userContext) > gitStorageLimit {
		t.Fatalf("usage %d is over the limit %d", gitUsage(userContext), gitStorageLimit)
	}
}
