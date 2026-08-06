package i18n

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

func testCache(t *testing.T) *cache.Cache {
	t.Helper()
	c := cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))
	t.Cleanup(func() { c.Close() })
	return c
}

func TestScanLocales(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "de", "LC_MESSAGES", "messages.po"), "")
	mustWrite(t, filepath.Join(dir, "sr", "LC_MESSAGES", "messages.po"), "")
	// "fr" has no LC_MESSAGES/messages.po yet - shouldn't be listed.
	if err := os.MkdirAll(filepath.Join(dir, "fr"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := NewManager(dir, testCache(t))
	got := m.scanLocales()

	want := map[string]bool{"en": true, "de": true, "sr": true}
	if len(got) != len(want) {
		t.Fatalf("scanLocales() = %v, want keys %v", got, want)
	}
	for _, l := range got {
		if !want[l] {
			t.Errorf("unexpected locale %q in %v", l, got)
		}
	}
}

func TestAvailableLocalesFallsBackToEnglish(t *testing.T) {
	m := NewManager(filepath.Join(t.TempDir(), "does-not-exist"), testCache(t))
	got := m.AvailableLocales(context.Background())
	if len(got) != 1 || got[0] != "en" {
		t.Errorf("AvailableLocales() = %v, want [en]", got)
	}
}

func TestSystemDefaultLocale(t *testing.T) {
	dir := t.TempDir()
	c := testCache(t)

	m := NewManager(dir, c)
	if got := m.SystemDefaultLocale(context.Background()); got != FallbackLocale {
		t.Errorf("with no default_locale file, got %q, want %q", got, FallbackLocale)
	}
}

func TestUserLocale(t *testing.T) {
	if got := UserLocale(""); got != "" {
		t.Errorf("UserLocale(\"\") = %q, want \"\"", got)
	}
	if got := UserLocale("nonexistent-user-xyz"); got != "" {
		t.Errorf("UserLocale for missing file = %q, want \"\"", got)
	}
}

func TestBestMatch(t *testing.T) {
	available := []string{"en", "de", "sr"}

	cases := []struct {
		header string
		want   string
	}{
		{"de-DE,de;q=0.9,en;q=0.8", "de"},
		{"sr-RS@latin,sr;q=0.9", "sr"},
		{"fr-FR,fr;q=0.9,en;q=0.5", "en"},
		{"", ""},
		{"xx-XX", ""},
	}

	for _, c := range cases {
		if got := bestMatch(c.header, available); got != c.want {
			t.Errorf("bestMatch(%q) = %q, want %q", c.header, got, c.want)
		}
	}
}

func TestResolveLocalePriority(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "de", "LC_MESSAGES", "messages.po"), "")
	m := NewManager(dir, testCache(t))
	ctx := context.Background()

	if got := m.ResolveLocale(ctx, "fr", "de", "en"); got != "fr" {
		t.Errorf("session locale should win, got %q", got)
	}
	if got := m.ResolveLocale(ctx, "", "de", "en"); got != "de" {
		t.Errorf("user locale should win over accept-language, got %q", got)
	}
	if got := m.ResolveLocale(ctx, "", "", "de-DE,de;q=0.9"); got != "de" {
		t.Errorf("accept-language should win over system default, got %q", got)
	}
	if got := m.ResolveLocale(ctx, "", "", ""); got != FallbackLocale {
		t.Errorf("should fall back to system default (en), got %q", got)
	}
}

func TestSubstitute(t *testing.T) {
	got := substitute("Firewall for %(domain)s is now %(status)s", []string{"domain", "example.com", "status", "On"})
	want := "Firewall for example.com is now On"
	if got != want {
		t.Errorf("substitute() = %q, want %q", got, want)
	}
}

// TestGetAgainstRealCatalog is a smoke test against the project's actual
// Serbian translation file (sibling ../OpenPanel/translations repo) to
// verify gotext parses real .po files from this codebase, not just
// synthetic fixtures. It's skipped if that repo isn't checked out.
func TestGetAgainstRealCatalog(t *testing.T) {
	src := "/home/stefan/OpenPanel/translations/sr-rs/messages.po"
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("real translations repo not available: %v", err)
	}

	dir := t.TempDir()
	dst := filepath.Join(dir, "sr", "LC_MESSAGES", "messages.po")
	mustWrite(t, dst, string(data))

	m := NewManager(dir, testCache(t))
	tr := m.Translator("sr")

	got := tr.Get("IP address mismatch. Please login again.")
	want := "IP adresa se ne poklapa. Molimo prijavite se ponovo."
	if got != want {
		t.Errorf("Get() = %q, want %q", got, want)
	}

	// A string with no translation entry should fall back to itself.
	if got := tr.Get("Some string that definitely isn't in the catalog"); got != "Some string that definitely isn't in the catalog" {
		t.Errorf("untranslated fallback = %q, want the original string", got)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
