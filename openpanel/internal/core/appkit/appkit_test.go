package appkit

import "testing"

func TestSafeDocroot(t *testing.T) {
	for docroot, want := range map[string]bool{
		"/var/www/html/example.com":            true,
		"/var/www/html/example.com/blog":       true,
		"/var/www/html/xn--80ak6aa92e.com/a_b": true,
		"/var/www/html/":                       false,
		"/var/www/html/../../etc":              false,
		"/var/www/html/a/../b":                 false,
		"/var/www/html/a; rm -rf /":            false,
		"/var/www/html/a$(id)":                 false,
		"example.com":                          false,
		"/etc/passwd":                          false,
	} {
		if got := SafeDocroot(docroot); got != want {
			t.Errorf("SafeDocroot(%q) = %v, want %v", docroot, got, want)
		}
	}
}
