package joomla

import "testing"

func TestLoginDBInfoDocrootValidation(t *testing.T) {
	for _, docroot := range []string{"/var/www/html/../etc", "/var/www/html/a/../../b", "/etc", ""} {
		if got := extractJoomlaDatabaseInfoForLogin("u", docroot)["error"]; got != "invalid docroot" {
			t.Errorf("extractJoomlaDatabaseInfoForLogin(%q) error = %q, want invalid docroot", docroot, got)
		}
	}
	// a normal docroot gets past the guard and only fails because there's no config file here
	if got := extractJoomlaDatabaseInfoForLogin("u", "/var/www/html/example.com")["error"]; got == "invalid docroot" {
		t.Errorf("extractJoomlaDatabaseInfoForLogin rejected a valid docroot")
	}
}
