package cmsapp

import "testing"

func TestIsValidSubdirectory(t *testing.T) {
	for subdir, want := range map[string]bool{
		"":                 true, // no subdirectory is a valid root install
		"blog":             true,
		"shop/eu":          true,
		"My_Site-1":        true,
		"../etc":           false,
		"a/../../etc":      false,
		"/etc":             false,
		"../../../../root": false,
	} {
		if got := IsValidSubdirectory(subdir); got != want {
			t.Errorf("IsValidSubdirectory(%q) = %v, want %v", subdir, got, want)
		}
	}
}
