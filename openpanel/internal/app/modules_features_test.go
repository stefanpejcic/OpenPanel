package app

import (
	"slices"
	"testing"
)

func TestWithoutDisabledModules(t *testing.T) {
	a := &App{enabledModuleSet: map[string]bool{"ssl": true, "ftp": true}}
	got := a.withoutDisabledModules([]string{"ssl", "redirects", "php_ini", "ftp", "mysql", "dashboard"})
	// redirects and php_ini are switched off, ftp and mysql are real modules that the router already gates
	want := []string{"ssl", "ftp", "mysql", "dashboard"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
