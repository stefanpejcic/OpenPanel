package filemanager

import "testing"

func TestPermissionsFormat(t *testing.T) {
	for value, want := range map[string]bool{
		"755": true, "644": true, "0755": true, "000": true,
		// these passed before, and os.FileMode silently dropped the special digit
		"1777": false, "2755": false, "4755": false, "7777": false,
		"75": false, "888": false, "abc": false, "": false, "00755": false,
	} {
		if got := permissionsRE.MatchString(value); got != want {
			t.Errorf("permissionsRE.MatchString(%q) = %v, want %v", value, got, want)
		}
	}
}
