package prestashop

import "testing"

func TestLoginDBInfoDocrootValidation(t *testing.T) {
	for _, docroot := range []string{"/var/www/html/../etc", "/var/www/html/a/../../b", "/etc", ""} {
		if got := extractPrestashopDatabaseInfoForLogin("u", docroot)["error"]; got != "invalid docroot" {
			t.Errorf("extractPrestashopDatabaseInfoForLogin(%q) error = %q, want invalid docroot", docroot, got)
		}
	}
	// a normal docroot gets past the guard and only fails because there's no config file here
	if got := extractPrestashopDatabaseInfoForLogin("u", "/var/www/html/example.com")["error"]; got == "invalid docroot" {
		t.Errorf("extractPrestashopDatabaseInfoForLogin rejected a valid docroot")
	}
}
