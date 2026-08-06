package app

import "testing"

func TestParseLicenseResponse(t *testing.T) {
	body := `<status>Active</status><expires>2099-01-01</expires>`
	got := parseLicenseResponse(body)
	if got["status"] != "Active" {
		t.Errorf("status = %q, want Active", got["status"])
	}
	if got["expires"] != "2099-01-01" {
		t.Errorf("expires = %q, want 2099-01-01", got["expires"])
	}
}

func TestParseLicenseResponseMismatchedTagsIgnored(t *testing.T) {
	// Malformed/mismatched tags shouldn't produce a spurious entry.
	got := parseLicenseResponse(`<status>Active</notstatus>`)
	if _, ok := got["status"]; ok {
		t.Errorf("expected mismatched tag pair to be ignored, got %v", got)
	}
}

func TestLicenseCheckPasses(t *testing.T) {
	cases := []struct {
		name         string
		licenseKey   string
		licenseValid bool
		want         bool
	}{
		{"no license key configured", "", false, true},
		{"enterprise key always passes", "enterprise-abc", false, true},
		{"noc key always passes", "noc-abc", false, true},
		{"lifetime key always passes", "lifetime-abc", false, true},
		{"regular key, valid", "abc123", true, true},
		{"regular key, invalid", "abc123", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := &App{LicenseKey: c.licenseKey, LicenseValid: c.licenseValid}
			if got := a.LicenseCheckPasses(); got != c.want {
				t.Errorf("LicenseCheckPasses() = %v, want %v", got, c.want)
			}
		})
	}
}
