package nextcloud

import "testing"

func TestLoginDBInfoDocrootValidation(t *testing.T) {
	for _, docroot := range []string{"/var/www/html/../etc", "/var/www/html/a/../../b", "/etc", ""} {
		if got := extractNextcloudDatabaseInfoForLogin("u", docroot)["error"]; got != "invalid docroot" {
			t.Errorf("extractNextcloudDatabaseInfoForLogin(%q) error = %q, want invalid docroot", docroot, got)
		}
	}
	// a normal docroot gets past the guard and only fails because there's no config file here
	if got := extractNextcloudDatabaseInfoForLogin("u", "/var/www/html/example.com")["error"]; got == "invalid docroot" {
		t.Errorf("extractNextcloudDatabaseInfoForLogin rejected a valid docroot")
	}
}
