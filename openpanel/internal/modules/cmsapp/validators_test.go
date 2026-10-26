package cmsapp

import "testing"

func TestValidVersion(t *testing.T) {
	valid := []string{"5.0.2", "31.0.1", "3.10.11", "1.2", "1.2.3.4"}
	for _, v := range valid {
		if !ValidVersion(v) {
			t.Errorf("ValidVersion(%q) = false, want true", v)
		}
	}
	// empty, traversal, path and URL metacharacters must all be rejected
	invalid := []string{
		"", "latest", "v5.0.2",
		"../../../../etc/openpanel/x",
		"5.0.2/../../../etc",
		"5.0.2 && reboot",
		"5.0.2.tar.gz",
		"1.2.3.4.5",
	}
	for _, v := range invalid {
		if ValidVersion(v) {
			t.Errorf("ValidVersion(%q) = true, want false", v)
		}
	}
}
