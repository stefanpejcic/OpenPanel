package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGravatarURL(t *testing.T) {
	if got := gravatarURL("letter", "user@example.com"); got != "" {
		t.Errorf("avatar_type=letter should return empty URL, got %q", got)
	}
	if got := gravatarURL("gravatar", ""); got != "" {
		t.Errorf("empty email should return empty URL, got %q", got)
	}
	got := gravatarURL("gravatar", "User@Example.com")
	want := "https://www.gravatar.com/avatar/b58996c504c5638798eb6b511e6f49af?s=150&d=identicon"
	if got != want {
		t.Errorf("gravatarURL() = %q, want %q", got, want)
	}
}

func TestLoadFeaturesFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "features.txt")
	if err := os.WriteFile(path, []byte("websites\nmysql\n\n  docker  \n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := loadFeaturesFromFile(path)
	if err != nil {
		t.Fatalf("loadFeaturesFromFile: %v", err)
	}
	want := []string{"websites", "mysql", "docker"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoadFeaturesFromFileMissing(t *testing.T) {
	_, err := loadFeaturesFromFile(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}
