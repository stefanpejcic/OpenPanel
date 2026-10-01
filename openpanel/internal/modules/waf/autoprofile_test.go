package waf

import (
	"context"
	"os"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func withOwnerAllows(t *testing.T, allowed bool) {
	t.Helper()
	orig := ownerAllowsWAF
	ownerAllowsWAF = func(context.Context, *appctx.App, string) bool { return allowed }
	t.Cleanup(func() { ownerAllowsWAF = orig })
}

func TestEnableProfileForNewSite(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	withOwnerAllows(t, true)

	// subfolder install, level kept
	if _, _, err := setDomainProfiles(context.Background(), "example.com", nil, "strict"); err != nil {
		t.Fatal(err)
	}
	EnableProfileForNewSite(nil, "example.com/blog", "wordpress")
	content, _ := os.ReadFile(path)
	if got := strings.Join(parseWAFProfiles(string(content)), ","); got != "wordpress" || parseWAFLevel(string(content)) != "strict" {
		t.Fatalf("profiles=%s level=%s", got, parseWAFLevel(string(content)))
	}
	// second install of the same app doesn't touch the file
	before := string(content)
	EnableProfileForNewSite(nil, "example.com", "WordPress")
	if content, _ = os.ReadFile(path); string(content) != before {
		t.Fatal("conf changed for an already active profile")
	}
	// apps without a profile and plugins that aren't installed are ignored
	EnableProfileForNewSite(nil, "example.com", "joomla")
	EnableProfileForNewSite(nil, "example.com", "nextcloud")
	if content, _ = os.ReadFile(path); string(content) != before {
		t.Fatal("conf changed for an app without an installed profile")
	}
}

func TestEnableProfileForNewSiteNeedsWAFOnPlan(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	withOwnerAllows(t, false)
	EnableProfileForNewSite(nil, "example.com", "wordpress")
	content, _ := os.ReadFile(path)
	if string(content) != testProfilesConf {
		t.Fatal("profile turned on although waf isn't on the owner's plan")
	}
}
