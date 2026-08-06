package app

import (
	"reflect"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestParseEnabledModules(t *testing.T) {
	cfg := config.Config{"enabled_modules": `'mysql', "docker" ,websites`}
	got := parseEnabledModules(cfg)
	want := []string{"mysql", "docker", "websites", "dashboard"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseEnabledModules() = %v, want %v", got, want)
	}
}

func TestParseEnabledModulesAlwaysIncludesMainModules(t *testing.T) {
	cfg := config.Config{"enabled_modules": "mysql"}
	got := parseEnabledModules(cfg)
	for _, want := range []string{"dashboard", "websites"} {
		found := false
		for _, m := range got {
			if m == want {
				found = true
			}
		}
		if !found {
			t.Errorf("expected %q in %v", want, got)
		}
	}
}

func TestParseEnabledModulesEmptyConfig(t *testing.T) {
	got := parseEnabledModules(config.Config{})
	want := []string{"dashboard", "websites"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseEnabledModules(empty) = %v, want %v", got, want)
	}
}

func TestAtoiDefault(t *testing.T) {
	if got := atoiDefault("42", 10); got != 42 {
		t.Errorf("atoiDefault(42) = %d, want 42", got)
	}
	if got := atoiDefault("", 10); got != 10 {
		t.Errorf("atoiDefault(\"\") = %d, want 10", got)
	}
	if got := atoiDefault("nope", 10); got != 10 {
		t.Errorf("atoiDefault(nope) = %d, want 10", got)
	}
}
