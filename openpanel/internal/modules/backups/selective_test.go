package backups

import (
	"context"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// a domain that failed used to vanish from the result as soon as another one was restored
func TestRestoreSelectedItemsReportsFailedDomains(t *testing.T) {
	orig := addMissingDomain
	addMissingDomain = func(_ context.Context, _ *appctx.App, _ int, _, domain, _, _ string, plan *restorePlan) error {
		if domain == "b.com" {
			return errors.New("b.com belongs to another account on this server")
		}
		plan.domains[domain] = true
		return nil
	}
	t.Cleanup(func() { addMissingDomain = orig })

	plan, _ := testPlan(t)
	store := &fakeStore{archives: map[string][]byte{"b.tar.gz": makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/vhosts/a.com.conf", body: []byte("a.com {\n}\n")},
		{name: "/backup/vhosts/b.com.conf", body: []byte("b.com {\n}\n")},
	})}}
	a := &appctx.App{Cache: cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))}
	req := restoreRequest{Username: "u", UserContext: "u-test-nonexistent", Form: url.Values{"items": {"a.com", "b.com"}}}

	status, out := restoreSelectedItems(context.Background(), a, req, store, "b.tar.gz", "domains", plan)
	msg, _ := out["message"].(string)
	failed, _ := out["failed"].([]string)
	if status != 200 || !strings.Contains(msg, "Restored (1 items).") || !strings.Contains(msg, "b.com belongs to another account") || len(failed) != 1 {
		t.Fatalf("got %d %v", status, out)
	}
}
