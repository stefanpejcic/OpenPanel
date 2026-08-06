package auth

import "testing"

func TestIsLicenseExempt(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/login", true},
		{"/reset_password", true},
		{"/reset_password/abc123token", true},
		{"/static/js/app.js", true},
		{"/dashboard", false},
		{"/websites", false},
		{"/login_autologin", false}, // distinct path, not a prefix match of "/login"
	}
	for _, c := range cases {
		if got := isLicenseExempt(c.path); got != c.want {
			t.Errorf("isLicenseExempt(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
