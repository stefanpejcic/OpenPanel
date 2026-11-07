package dashboard

import (
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestTimeAgo(t *testing.T) {
	tr := i18n.NewManager(t.TempDir(), nil).Translator("en")
	now := time.Date(2026, 10, 29, 14, 45, 2, 0, time.Local)
	for _, tc := range []struct{ stamp, want string }{
		{"2026-10-29 14:45:02", "1 second ago"},
		{"2026-10-29 14:44:30", "32 seconds ago"},
		{"2026-10-29 14:44:00", "1 minute ago"},
		{"2026-10-29 14:24:05", "20 minutes ago"},
		{"2026-10-29 12:40:00", "2 hours ago"},
		{"2026-10-28 14:00:00", "1 day ago"},
		{"2026-10-20 14:00:00", "1 week ago"},
		{"2026-08-01 14:00:00", "2 months ago"},
		{"2024-10-01 14:00:00", "2 years ago"},
		{"2026-10-29 15:00:00", "1 second ago"},
		{"", ""},
		{"yesterday", ""},
	} {
		if got := timeAgo(tr, tc.stamp, now); got != tc.want {
			t.Errorf("timeAgo(%q) = %q, want %q", tc.stamp, got, tc.want)
		}
	}
}

func TestParseContainerStates(t *testing.T) {
	// real `podman ps -a --format "{{.Names}} {{.State}}"` output from the test VM
	out := "memcached running\napache running\nphp-fpm-8.5 running\nn8n running\nflaskapp exited\nnodeapp running\n\n"
	got := parseContainerStates(out)
	for name, want := range map[string]string{"nodeapp": "running", "flaskapp": "exited", "php-fpm-8.5": "running", "missing": ""} {
		if got[name] != want {
			t.Errorf("state of %s = %q, want %q", name, got[name], want)
		}
	}
	if len(parseContainerStates("")) != 0 {
		t.Error("empty output should give no states")
	}
}

func TestDashboardSiteRunsOnAndState(t *testing.T) {
	for _, tc := range []struct {
		site       DashboardSite
		runs, want string
	}{
		{DashboardSite{Type: "WordPress", PHP: "8.3"}, "PHP 8.3", ""},
		{DashboardSite{Type: "Node.js", Service: "nodeapp", Port: "3000", State: "running"}, "nodeapp:3000", "Running"},
		{DashboardSite{Type: "Docker", Service: "uptimekuma", Port: "3001", State: "exited"}, "uptimekuma:3001", "Stopped"},
		{DashboardSite{Type: "n8n", Service: "n8n"}, "n8n", ""},
		{DashboardSite{Type: "Python"}, "Python", ""},
	} {
		if got := tc.site.RunsOn(); got != tc.runs {
			t.Errorf("RunsOn() = %q, want %q", got, tc.runs)
		}
		if got := tc.site.StateLabel(); got != tc.want {
			t.Errorf("StateLabel() = %q, want %q", got, tc.want)
		}
	}
}
