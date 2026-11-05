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
	if err := saveDownloadState("11111111-1111-4111-8111-111111111111", downloadState{Status: "done"}); err != nil {
		t.Fatal(err)
	}
	if err := saveDownloadState("22222222-2222-4222-8222-222222222222", downloadState{Status: "downloading"}); err != nil {
		t.Fatal(err)
	}
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(filepath.Join(wgetStateDir, "11111111-1111-4111-8111-111111111111.json"), twoDaysAgo, twoDaysAgo); err != nil {
		t.Fatal(err)
	}

	// nothing used to remove these, one file per download piled up in /tmp
	pruneDownloadStates(24 * time.Hour)
	if _, ok := loadDownloadState("11111111-1111-4111-8111-111111111111"); ok {
		t.Error("old state file was not pruned")
	}
	if _, ok := loadDownloadState("22222222-2222-4222-8222-222222222222"); !ok {
		t.Error("fresh state file was pruned")
	}
}

func TestDownloadStateOnlyForItsOwner(t *testing.T) {
	withStateDir(t)
	if err := saveDownloadState("33333333-3333-4333-8333-333333333333", downloadState{Context: "alice", URL: "https://example.com/a.zip"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadDownloadStateFor("33333333-3333-4333-8333-333333333333", "alice"); !ok {
		t.Error("owner can't read their own download")
	}
	// any logged-in account holding the id used to see the URL and progress
	if _, ok := loadDownloadStateFor("33333333-3333-4333-8333-333333333333", "bob"); ok {
		t.Error("another account can read the download")
	}
	if _, ok := loadDownloadStateFor("33333333-3333-4333-8333-333333333333", ""); ok {
		t.Error("empty context can read the download")
	}
}

func TestClaimDoneLogOnce(t *testing.T) {
	withStateDir(t)
	if err := saveDownloadState("33333333-3333-4333-8333-333333333333", downloadState{Status: "done", Context: "alice"}); err != nil {
		t.Fatal(err)
	}
	// concurrent polls used to read logged=false together and log the download several times
	var wg sync.WaitGroup
	var claims atomic.Int32
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, first := claimDoneLog("33333333-3333-4333-8333-333333333333"); first {
				claims.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := claims.Load(); n != 1 {
		t.Fatalf("claimed %d times, want 1", n)
	}
}

func TestDownloadIDMustBeAUUID(t *testing.T) {
	withStateDir(t)
	// the id comes from the url path, a ../ here used to reach files outside the state dir
	for _, id := range []string{"../../home/alice/evil", "..%2Fevil", "dl", "", "33333333-3333-4333-8333-333333333333/../x"} {
		if _, err := stateFile(id); err == nil {
			t.Errorf("stateFile(%q) accepted a non-uuid id", id)
		}
		if err := saveDownloadState(id, downloadState{Status: "done"}); err == nil {
			t.Errorf("saveDownloadState(%q) wrote a state file", id)
		}
	}
	if _, err := stateFile("33333333-3333-4333-8333-333333333333"); err != nil {
		t.Errorf("valid uuid refused: %v", err)
	}
}
