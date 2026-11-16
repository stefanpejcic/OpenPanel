package botcontrol

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRateLimitReadsTheAdminValue(t *testing.T) {
	orig := botControlConfPath
	t.Cleanup(func() { botControlConfPath = orig })
	botControlConfPath = filepath.Join(t.TempDir(), "botcontrol.conf")

	if got := RateLimit(); got != 10 {
		t.Fatalf("missing file gives %d, want 10", got)
	}
	conf := "(botcontrol_block) {\n    respond 403\n}\n\n(botcontrol_ratelimit) {\n    handle @ai_crawlers {\n        rate_limit {\n            zone botcontrol_{args[0]} {\n                events 60\n                window 1m\n            }\n        }\n    }\n}\n"
	if err := os.WriteFile(botControlConfPath, []byte(conf), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := RateLimit(); got != 60 {
		t.Fatalf("got %d, want 60 from the snippet", got)
	}
	if err := os.WriteFile(botControlConfPath, []byte("(botcontrol_ratelimit) {\n    events 0\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := RateLimit(); got != 10 {
		t.Fatalf("zero falls back to %d, want 10", got)
	}
}
