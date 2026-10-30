package crons

import (
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestNewJobProblem(t *testing.T) {
	containers := []string{"php-fpm-8.5", "mariadb"}
	existing := []CronJob{{Comment: "Nightly"}}
	cases := []struct {
		name, container, comment string
		auto                     bool
		want                     string
	}{
		{"ok named", "php-fpm-8.5", "Weekly report", false, ""},
		{"ok auto", "mariadb", "mariadb", true, ""},
		{"unknown container", "other-users-db", "x", false, "isn't one of your containers"},
		{"empty container list", "php-fpm-8.5", "", true, ""},
		{"too long", "mariadb", strings.Repeat("a", 31), false, "30 characters"},
		{"quote", "mariadb", `a"b`, false, "double quotes"},
		{"backslash", "mariadb", `a\b`, false, "double quotes"},
		{"taken", "mariadb", "Nightly", false, "already exists"},
		{"auto name may repeat", "mariadb", "Nightly", true, ""},
	}
	for _, c := range cases {
		list := containers
		if c.name == "empty container list" {
			list = nil
			c.want = "isn't one of your containers"
		}
		got := newJobProblem(c.container, c.comment, c.auto, list, existing)
		if c.want == "" && got != "" || c.want != "" && !strings.Contains(got, c.want) {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSchedulePreview(t *testing.T) {
	tr := i18n.NewManager(t.TempDir(), nil).Translator("en")
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	p := schedulePreview(tr, "0 30 2 * * *", now, time.UTC)
	if p.Error != "" || len(p.Next) != 3 || p.Next[0] != "2026-10-31T02:30:00Z" || p.Text != "Every day at 02:30" {
		t.Errorf("daily: %+v", p)
	}
	if p := schedulePreview(tr, "30 2 * * *", now, time.UTC); !strings.Contains(p.Error, "only 5 fields") || len(p.Next) != 0 {
		t.Errorf("5 fields: %+v", p)
	}
	if p := schedulePreview(tr, "0 99 * * * *", now, time.UTC); !strings.HasPrefix(p.Error, "Invalid schedule:") {
		t.Errorf("out of range: %+v", p)
	}
	loc, _ := time.LoadLocation("Europe/Belgrade")
	if p := schedulePreview(tr, "@daily", now, loc); p.Error != "" || p.Next[0] != "2026-10-31T00:00:00+01:00" {
		t.Errorf("zone: %+v", p)
	}
}
