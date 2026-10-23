package filemanager

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func withStateDir(t *testing.T) {
	t.Helper()
	old := wgetStateDir
	wgetStateDir = t.TempDir()
	t.Cleanup(func() { wgetStateDir = old })
}

func TestPruneDownloadStates(t *testing.T) {
	withStateDir(t)
	if err := saveDownloadState("old", downloadState{Status: "done"}); err != nil {
		t.Fatal(err)
	}
	if err := saveDownloadState("fresh", downloadState{Status: "downloading"}); err != nil {
		t.Fatal(err)
	}
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(filepath.Join(wgetStateDir, "old.json"), twoDaysAgo, twoDaysAgo); err != nil {
		t.Fatal(err)
	}

	// nothing used to remove these, one file per download piled up in /tmp
	pruneDownloadStates(24 * time.Hour)
	if _, ok := loadDownloadState("old"); ok {
		t.Error("old state file was not pruned")
	}
	if _, ok := loadDownloadState("fresh"); !ok {
		t.Error("fresh state file was pruned")
	}
}

func TestDownloadStateOnlyForItsOwner(t *testing.T) {
	withStateDir(t)
	if err := saveDownloadState("dl", downloadState{Context: "alice", URL: "https://example.com/a.zip"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadDownloadStateFor("dl", "alice"); !ok {
		t.Error("owner can't read their own download")
	}
	// any logged-in account holding the id used to see the URL and progress
	if _, ok := loadDownloadStateFor("dl", "bob"); ok {
		t.Error("another account can read the download")
	}
	if _, ok := loadDownloadStateFor("dl", ""); ok {
		t.Error("empty context can read the download")
	}
}

func TestClaimDoneLogOnce(t *testing.T) {
	withStateDir(t)
	if err := saveDownloadState("dl", downloadState{Status: "done", Context: "alice"}); err != nil {
		t.Fatal(err)
	}
	// concurrent polls used to read logged=false together and log the download several times
	var wg sync.WaitGroup
	var claims atomic.Int32
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, first := claimDoneLog("dl"); first {
				claims.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := claims.Load(); n != 1 {
		t.Fatalf("claimed %d times, want 1", n)
	}
}
