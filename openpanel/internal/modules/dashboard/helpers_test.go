package dashboard

import (
	"context"
	"path/filepath"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

func testApp(t *testing.T) *appctx.App {
	t.Helper()
	return &appctx.App{Cache: cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))}
}

func TestCountFTPAccounts(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()

	// countFTPAccounts hardcodes a system path this sandbox can't write to - verify the "missing file -> 0" path, real behavior for a non-installed FTP feature
	got := countFTPAccounts(a, ctx, "definitely-does-not-exist-user")
	if got != 0 {
		t.Errorf("countFTPAccounts() for missing file = %d, want 0", got)
	}
}

func TestAtoiDefaultRAMParsing(t *testing.T) {
	// Mirrors dashboard()'s `int(str(ram_limit).rstrip('gG'))`.
	cases := []struct {
		raw  string
		want int
	}{
		{"4g", 0}, // atoiDefault doesn't strip suffixes itself - caller does via strings.TrimRight first
		{"4", 4},
		{"", 0},
		{"bad", 0},
	}
	for _, c := range cases {
		if got := atoiDefault(c.raw, 0); got != c.want {
			t.Errorf("atoiDefault(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}
