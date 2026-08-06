package account

import (
	"os"
	"path/filepath"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestFavoritesMax(t *testing.T) {
	a := &appctx.App{Config: config.Config{"favorites_items": "5"}}
	if got := favoritesMax(a); got != 5 {
		t.Errorf("got %d, want 5", got)
	}
	a2 := &appctx.App{Config: config.Config{}}
	if got := favoritesMax(a2); got != 10 {
		t.Errorf("default: got %d, want 10", got)
	}
}

func TestEnsureFavoritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "favorites.json")
	if err := ensureFavoritesFile(path); err != nil {
		t.Fatalf("ensureFavoritesFile: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "[]" {
		t.Errorf("got %q, want []", content)
	}

	// Existing file with real content should not be overwritten.
	if err := os.WriteFile(path, []byte(`[{"link":"x","title":"y"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureFavoritesFile(path); err != nil {
		t.Fatalf("ensureFavoritesFile (existing): %v", err)
	}
	content, _ = os.ReadFile(path)
	if string(content) != `[{"link":"x","title":"y"}]` {
		t.Errorf("existing file was overwritten: %q", content)
	}
}

func TestReadFavoritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	if err := os.WriteFile(path, []byte(`[{"link":"dashboard","title":"Dashboard"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	favs, err := readFavoritesFile(path)
	if err != nil {
		t.Fatalf("readFavoritesFile: %v", err)
	}
	if len(favs) != 1 || favs[0].Link != "dashboard" || favs[0].Title != "Dashboard" {
		t.Errorf("got %+v", favs)
	}
}
