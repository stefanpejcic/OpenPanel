package postgresql

import (
	"strings"
	"testing"
)

func TestPostgresContainerStatusDetail(t *testing.T) {
	cases := []struct {
		state, health string
		wantEmpty     bool
		wantSubstr    string
	}{
		{"running", "healthy", true, ""},
		{"running", "unhealthy", false, "unhealthy"},
		{"running", "starting", false, "still initializing"},
		{"running", "", false, "Unable to retrieve"},
		{"not_found", "", false, "installation is underway"},
		{"created", "", false, "not yet started"},
		{"restarting", "", false, "restarting"},
		{"paused", "", false, "paused"},
		{"exited", "", false, "stopped"},
		{"removing", "", false, "being deleted"},
		{"dead", "", false, "crashed"},
		{"bogus", "", false, "Unable to retrieve"},
	}
	for _, c := range cases {
		got := postgresContainerStatusDetail(c.state, c.health)
		if c.wantEmpty && got != "" {
			t.Errorf("state=%q health=%q: got %q, want empty", c.state, c.health, got)
		}
		if !c.wantEmpty && !strings.Contains(got, c.wantSubstr) {
			t.Errorf("state=%q health=%q: got %q, want substring %q", c.state, c.health, got, c.wantSubstr)
		}
	}
}
