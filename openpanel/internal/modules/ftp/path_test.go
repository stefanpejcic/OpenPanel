package ftp

import "testing"

func TestCleanFTPPath(t *testing.T) {
	ok := map[string]string{
		"/var/www/html/":            "/var/www/html/",
		"/var/www/html/example.com": "/var/www/html/example.com",
		"/var/www/html/a/./b/":      "/var/www/html/a/b",
		"/var/www/html/a//b":        "/var/www/html/a/b",
	}
	for in, want := range ok {
		if got, good := cleanFTPPath(in); !good || got != want {
			t.Errorf("cleanFTPPath(%q) = %q, %v; want %q", in, got, good, want)
		}
	}
	bad := []string{
		"/var/www/html/../../../../../../../home/victim/site",
		"/var/www/html/a/../b",
		"/var/www/html/..",
		"/var/www/htmlx",
		"/etc",
		"/var/www/html/a|b",
		"/var/www/html/*",
		"",
	}
	for _, in := range bad {
		if got, good := cleanFTPPath(in); good {
			t.Errorf("cleanFTPPath(%q) = %q, accepted", in, got)
		}
	}
}
