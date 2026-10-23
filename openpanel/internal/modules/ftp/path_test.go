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

func TestFTPPathInput(t *testing.T) {
	cases := map[string]string{
		// no trailing slash used to become /var/www/html/var/www/html
		"/var/www/html":        "/var/www/html/",
		"/var/www/html/":       "/var/www/html/",
		"/var/www/html/site":   "/var/www/html/site",
		" /var/www/html/site ": "/var/www/html/site",
		"site":                 "/var/www/html/site",
		"/site":                "/var/www/html/site",
		"/var/www/html/site/":  "/var/www/html/site",
		"\t/var/www/html\n":    "/var/www/html/",
	}
	for in, want := range cases {
		if got, ok := cleanFTPPath(ftpPathInput(in)); !ok || got != want {
			t.Errorf("cleanFTPPath(ftpPathInput(%q)) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if got, ok := cleanFTPPath(ftpPathInput("/var/www/html/../etc")); ok {
		t.Errorf("traversal accepted as %q", got)
	}
}
