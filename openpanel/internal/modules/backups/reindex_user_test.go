package backups

import "testing"

func TestReindexUserRejectsBadContext(t *testing.T) {
	for _, ctx := range []string{"", "../etc", "a/b", "a b"} {
		if err := ReindexUser(ctx); err == nil {
			t.Errorf("ReindexUser(%q) should fail", ctx)
		}
	}
}
