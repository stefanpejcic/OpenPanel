package php

import "testing"

func TestPhpVersionFromSegment(t *testing.T) {
	cases := map[string]string{"php8.2": "8.2", "php7.4": "7.4", "": ""}
	for in, want := range cases {
		if got := phpVersionFromSegment(in); got != want {
			t.Errorf("phpVersionFromSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPhpVersionFromIniSegment(t *testing.T) {
	cases := map[string]string{"php8.2.ini": "8.2", "php7.4.ini": "7.4", "": ""}
	for in, want := range cases {
		if got := phpVersionFromIniSegment(in); got != want {
			t.Errorf("phpVersionFromIniSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSortVersionsDesc(t *testing.T) {
	versions := []string{"7.4", "8.10", "8.2", "8.1"}
	sortVersionsDesc(versions)
	want := []string{"8.10", "8.2", "8.1", "7.4"}
	for i, w := range want {
		if versions[i] != w {
			t.Fatalf("sortVersionsDesc = %v, want %v", versions, want)
		}
	}
}

func TestParseConfigContent(t *testing.T) {
	content := "; comment\n# also comment\n[section]\nmemory_limit = 256M\nshort_open_tag\n\ndate.timezone=UTC\n"
	entries := parseConfigContent(content)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %+v", len(entries), entries)
	}
	if entries[0].Key != "memory_limit" || entries[0].Value != "256M" || !entries[0].HasValue {
		t.Errorf("entries[0] = %+v", entries[0])
	}
	if entries[1].Key != "short_open_tag" || entries[1].HasValue {
		t.Errorf("entries[1] = %+v", entries[1])
	}
	if entries[2].Key != "date.timezone" || entries[2].Value != "UTC" {
		t.Errorf("entries[2] = %+v", entries[2])
	}
}

func TestConfigEntriesToMap(t *testing.T) {
	entries := []ConfigEntry{{Key: "a", Value: "1", HasValue: true}, {Key: "a", Value: "2", HasValue: true}}
	m := configEntriesToMap(entries)
	if m["a"] != "2" {
		t.Errorf("expected later entry to win, got %q", m["a"])
	}
}

func TestBuildOptionField(t *testing.T) {
	tz := []string{"UTC", "Europe/Belgrade"}

	if f := buildOptionField("display_errors", "1", tz); f.Kind != "checkbox_binary" || !f.Checked {
		t.Errorf("display_errors=1: %+v", f)
	}
	if f := buildOptionField("display_errors", "0", tz); f.Checked {
		t.Errorf("display_errors=0 should be unchecked: %+v", f)
	}
	if f := buildOptionField("date.timezone", "Europe/Belgrade", tz); f.Kind != "timezone" || len(f.Timezones) != 2 {
		t.Errorf("date.timezone: %+v", f)
	}
	if f := buildOptionField("date.timezone", "Custom/Zone", tz); f.Kind != "timezone" || len(f.Timezones) != 3 {
		t.Errorf("date.timezone with unknown value should prepend it: %+v", f)
	}
	if f := buildOptionField("zlib.output_compression", "On", tz); f.Kind != "checkbox_binary" || !f.Checked {
		t.Errorf("zlib.output_compression=On: %+v", f)
	}
	if f := buildOptionField("output_buffering", "On", tz); f.Kind != "checkbox_onoff" || !f.Checked {
		t.Errorf("output_buffering=On: %+v", f)
	}
	if f := buildOptionField("post_max_size", "256M", tz); f.Kind != "unit" || f.NumberPart != "256" || f.UnitPart != "M" {
		t.Errorf("post_max_size=256M: %+v", f)
	}
	if f := buildOptionField("max_input_vars", "1000", tz); f.Kind != "number" {
		t.Errorf("max_input_vars=1000: %+v", f)
	}
	if f := buildOptionField("open_basedir", "/var/www:/tmp", tz); f.Kind != "text" {
		t.Errorf("open_basedir: %+v", f)
	}
}

func TestClassifyPHPVersionLevel(t *testing.T) {
	data := map[string]VersionInfo{
		"8.3": {StatusLabel: "Latest", IsLatestVersion: true},
		"8.1": {StatusLabel: "Secure", IsSecureVersion: true},
		"7.4": {StatusLabel: "EOL", IsEOLVersion: true},
	}
	if level, _ := classifyPHPVersionLevel("8.3", data); level != "good" {
		t.Errorf("8.3 should be good, got %s", level)
	}
	if level, _ := classifyPHPVersionLevel("8.1", data); level != "secure" {
		t.Errorf("8.1 should be secure, got %s", level)
	}
	if level, _ := classifyPHPVersionLevel("7.4", data); level != "unsupported" {
		t.Errorf("7.4 (EOL only) should be unsupported, got %s", level)
	}
	if level, _ := classifyPHPVersionLevel("5.6", data); level != "unsupported" {
		t.Errorf("5.6 (unknown) should be unsupported, got %s", level)
	}
}

func TestParseExtensionsTable(t *testing.T) {
	body := "" +
		"intro text\n" +
		"| Extension | PHP 8.2 | PHP 8.3 |\n" +
		"|:---:|:---:|:---:|\n" +
		"| bcmath | &check; | &check; |\n" +
		"| xdebug[^1] | &check; |  |\n" +
		"trailing text\n"
	table := parseExtensionsTable(body)
	if !table["bcmath"]["8.2"] || !table["bcmath"]["8.3"] {
		t.Errorf("bcmath: %+v", table["bcmath"])
	}
	if !table["xdebug"]["8.2"] || table["xdebug"]["8.3"] {
		t.Errorf("xdebug (footnote-stripped name, 8.3 unsupported): %+v", table["xdebug"])
	}
}

func TestContainsString(t *testing.T) {
	if !containsString([]string{"8.1", "8.2"}, "8.2") {
		t.Error("expected true")
	}
	if containsString([]string{"8.1"}, "8.2") {
		t.Error("expected false")
	}
}
