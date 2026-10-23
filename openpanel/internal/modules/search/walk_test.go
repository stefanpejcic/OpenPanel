package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWalkLimitedDepthStopsEverywhere(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"a", "b", "c", "d", "e", "f"} {
		if err := os.MkdirAll(filepath.Join(root, d, "deeper"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	visits := 0
	// stop after 3 folders, this used to only end the current folder while its siblings kept being visited
	_ = walkLimitedDepth(root, strings.Count(root, string(os.PathSeparator)), maxSearchDepth, func(string, []os.DirEntry) bool {
		visits++
		return visits < 3
	})
	if visits != 3 {
		t.Fatalf("visit called %d times after asking to stop at 3", visits)
	}
}
