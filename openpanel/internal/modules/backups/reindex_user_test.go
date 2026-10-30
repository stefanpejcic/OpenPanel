package backups

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReindexUserRejectsBadContext(t *testing.T) {
	for _, ctx := range []string{"", "../etc", "a/b", "a b"} {
		if err := ReindexUser(ctx); err == nil {
			t.Errorf("ReindexUser(%q) should fail", ctx)
		}
	}
}

func TestTakeReindexLock(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "reindex.lock")
	if took, err := takeReindexLock(lock); !took || err != nil {
		t.Fatalf("first take: %v %v", took, err)
	}
	if took, err := takeReindexLock(lock); took || err != nil {
		t.Fatalf("second take while held: %v %v", took, err)
	}
	if !reindexRunning(lock) {
		t.Fatal("fresh lock not reported as running")
	}

	// a lock left behind by a panel restart used to block reindexing for good
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	if reindexRunning(lock) {
		t.Fatal("stale lock reported as running")
	}
	if took, err := takeReindexLock(lock); !took || err != nil {
		t.Fatalf("take over a stale lock: %v %v", took, err)
	}

	if _, err := takeReindexLock(filepath.Join(t.TempDir(), "missing", "reindex.lock")); err == nil {
		t.Fatal("expected an error when the lock can't be created")
	}
}

func TestKeepLockFresh(t *testing.T) {
	orig := reindexLockTouch
	reindexLockTouch = 10 * time.Millisecond
	t.Cleanup(func() { reindexLockTouch = orig })

	lock := filepath.Join(t.TempDir(), "reindex.lock")
	old := time.Now().Add(-time.Hour)
	if err := os.WriteFile(lock, []byte("running"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	stop := keepLockFresh(lock)
	// waits for the touch instead of a fixed sleep, a busy test machine can be slow to run the ticker
	deadline := time.Now().Add(5 * time.Second)
	for !reindexRunning(lock) {
		if time.Now().After(deadline) {
			t.Fatal("a running reindex let its lock go stale")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// once stop returns nothing writes the lock again, the reindex removes it right after
	stop()
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * reindexLockTouch)
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("the lock came back after stop: %v", err)
	}
}
