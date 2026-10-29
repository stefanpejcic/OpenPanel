package mediawiki

import "testing"

func TestLoginDBInfoDocrootValidation(t *testing.T) {
	for _, docroot := range []string{"/var/www/html/../etc", "/var/www/html/a/../../b", "/etc", ""} {
		if got := extractMediaWikiDatabaseInfoForLogin("u", docroot)["error"]; got != "invalid docroot" {
			t.Errorf("extractMediaWikiDatabaseInfoForLogin(%q) error = %q, want invalid docroot", docroot, got)
		}
	}
	// a normal docroot gets past the guard and only fails because there's no config file here
	if got := extractMediaWikiDatabaseInfoForLogin("u", "/var/www/html/example.com")["error"]; got == "invalid docroot" {
		t.Errorf("extractMediaWikiDatabaseInfoForLogin rejected a valid docroot")
	}
}
