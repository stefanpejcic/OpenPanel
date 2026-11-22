package crons

import (
	"context"
	"testing"
)

func TestAddJobRejectsInjectedFields(t *testing.T) {
	bad := []struct{ comment, container, command string }{
		{"flarum-x", "php-fpm-8.3", "php /var/www/html/x/flarum schedule:run\n[job-exec \"evil\"]\ncontainer = docker-proxy"},
		{"flarum-x", "docker-proxy", "id"},
		{"flarum-x", "cron", "id"},
		{"a\"]\n[job-exec \"b", "php-fpm-8.3", "id"},
	}
	for _, c := range bad {
		if err := AddJob(context.Background(), "nobody-test", c.comment, "0 * * * * *", c.container, c.command, true); err == nil {
			t.Fatalf("AddJob(%q, %q, %q) should have been rejected", c.comment, c.container, c.command)
		}
	}
}
