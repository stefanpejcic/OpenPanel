package backups

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
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

// an entry that's in the backup but can't be written used to be reported as not found
func TestRestoreReportsWriteFailures(t *testing.T) {
	plan, tmp := testPlan(t)
	// a file sits where the backup has a folder
	if err := os.WriteFile(filepath.Join(tmp, "html", "site"), []byte("not a folder"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/html/ok.txt", body: []byte("ok")},
		{name: "/backup/html/site/index.php", body: []byte("<?php")},
	})

	extracted, err := restoreFilesFromTar(bytes.NewReader(archive), plan, 0, nil)
	var failures *writeFailures
	if len(extracted) != 1 || !errors.As(err, &failures) || failures.count != 1 || !strings.Contains(err.Error(), "html/site/index.php") {
		t.Fatalf("restoreFilesFromTar: extracted %v, err %v", extracted, err)
	}

	store := &fakeStore{archives: map[string][]byte{"b.tar.gz": archive}}
	a := &appctx.App{Cache: cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))}
	req := restoreRequest{Username: "u", UserContext: "u-test-nonexistent", Form: url.Values{"items": {"site"}}}
	status, out := restoreSelectedItems(context.Background(), a, req, store, "b.tar.gz", "paths", plan)
	if status != 500 || !strings.Contains(fmt.Sprint(out["error"]), "could not be written") {
		t.Fatalf("restoreSelectedItems: got %d %v, want a write error instead of not found", status, out)
	}
}
