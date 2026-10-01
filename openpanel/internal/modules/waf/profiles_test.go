package waf

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// both http and https blocks, like the real template
const testProfilesConf = `http://example.com {
    coraza_waf {
        directives ` + "`" + `
            Include /etc/openpanel/caddy/coraza_rules.conf
            Include /etc/openpanel/caddy/coreruleset/crs-setup.conf.example
            Include /etc/openpanel/caddy/coreruleset/rules/*.conf
            SecRuleEngine On
            SecRuleRemoveById 007 920350
            SecRuleRemoveByTag example
        ` + "`" + `
    }
}
https://example.com {
    coraza_waf {
        directives ` + "`" + `
            Include /etc/openpanel/caddy/coraza_rules.conf
            Include /etc/openpanel/caddy/coreruleset/crs-setup.conf.example
            Include /etc/openpanel/caddy/coreruleset/rules/*.conf
            SecRuleEngine On
            SecRuleRemoveById 007 920350
            SecRuleRemoveByTag example
        ` + "`" + `
    }
}
`

// withTestPlugins fakes an installed plugins dir: wordpress (config+before), phpmyadmin (config+before+after)
func withTestPlugins(t *testing.T) string {
	t.Helper()
	dir := t.TempDir() + "/"
	orig := crsPluginsDir
	crsPluginsDir = dir
	t.Cleanup(func() { crsPluginsDir = orig })
	files := map[string][]string{"wordpress": {"config", "before"}, "phpmyadmin": {"config", "before", "after"}}
	for key, stages := range files {
		for _, stage := range stages {
			p := pluginFile(key, stage)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("# test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

func writeProfilesConf(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(testProfilesConf), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAvailableProfilesOnlyInstalled(t *testing.T) {
	withTestPlugins(t)
	var keys []string
	for _, p := range availableProfiles() {
		keys = append(keys, p.Key)
	}
	if strings.Join(keys, ",") != "wordpress,phpmyadmin" {
		t.Fatalf("got %v", keys)
	}
}

func TestRewriteProfilesOrderAndBothBlocks(t *testing.T) {
	dir := withTestPlugins(t)
	out, err := rewriteProfiles(testProfilesConf, []string{"wordpress", "phpmyadmin"}, "")
	if err != nil {
		t.Fatal(err)
	}
	block := strings.Split(out, "https://example.com")[1]
	setup := strings.Index(block, "crs-setup.conf.example")
	wpConfig := strings.Index(block, dir+"wordpress-rule-exclusions-plugin/plugins/wordpress-rule-exclusions-config.conf")
	wpBefore := strings.Index(block, "wordpress-rule-exclusions-before.conf")
	rules := strings.Index(block, "/coreruleset/rules/*.conf")
	pmaAfter := strings.Index(block, "phpmyadmin-rule-exclusions-after.conf")
	if !(setup < wpConfig && wpConfig < wpBefore && wpBefore < rules && rules < pmaAfter) {
		t.Fatalf("wrong include order:\n%s", block)
	}
	if strings.Contains(out, "wordpress-rule-exclusions-after.conf") {
		t.Fatal("included a plugin file that doesn't exist")
	}
	if n := strings.Count(out, "wordpress-rule-exclusions-before.conf"); n != 2 {
		t.Fatalf("expected plugin in both http and https blocks, got %d", n)
	}
	if !strings.Contains(out, "            Include "+dir+"wordpress") {
		t.Fatal("include not indented like the rules include")
	}
	if got := strings.Join(parseWAFProfiles(out), ","); got != "wordpress,phpmyadmin" {
		t.Fatalf("parse after rewrite: %s", got)
	}
	if !strings.Contains(out, "SecRuleRemoveById 007 920350") {
		t.Fatal("rule removals lost")
	}
}

func TestRewriteProfilesIsIdempotentAndRemoves(t *testing.T) {
	withTestPlugins(t)
	once, _ := rewriteProfiles(testProfilesConf, []string{"wordpress"}, "")
	twice, _ := rewriteProfiles(once, []string{"wordpress"}, "")
	if once != twice {
		t.Fatal("rewrite with the same profiles changed the file")
	}
	cleared, _ := rewriteProfiles(twice, nil, "")
	if cleared != testProfilesConf {
		t.Fatalf("clearing profiles didn't restore the original:\n%s", cleared)
	}
}

func TestRewriteProfilesNoRulesInclude(t *testing.T) {
	withTestPlugins(t)
	if _, err := rewriteProfiles(testDomainConf, []string{"wordpress"}, ""); !errors.Is(err, errNoRulesInclude) {
		t.Fatalf("expected errNoRulesInclude, got %v", err)
	}
}

func TestSetDomainProfiles(t *testing.T) {
	withTestPlugins(t)
	path, reloads := withTestDomain(t)
	writeProfilesConf(t, path)

	active, _, err := setDomainProfiles(context.Background(), "example.com", []string{"WordPress", "wordpress", ""}, "")
	if err != nil || strings.Join(active, ",") != "wordpress" {
		t.Fatalf("active=%v err=%v", active, err)
	}
	if *reloads != 1 {
		t.Fatalf("expected 1 reload, got %d", *reloads)
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", []string{"wordpress"}, ""); err != nil || *reloads != 1 {
		t.Fatalf("no-op should not reload, reloads=%d err=%v", *reloads, err)
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", []string{"drupal"}, ""); err == nil {
		t.Fatal("expected error for a plugin that isn't installed")
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", []string{"../../etc"}, ""); err == nil {
		t.Fatal("expected error for a junk key")
	}
}

func TestSetDomainProfilesRollsBackOnReloadFailure(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	calls := 0
	reloadCaddy = func(context.Context) error {
		calls++
		if calls == 1 {
			return errors.New("bad config")
		}
		return nil
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", []string{"wordpress"}, ""); !errors.Is(err, errProfileReload) {
		t.Fatalf("expected errProfileReload, got %v", err)
	}
	content, _ := os.ReadFile(path)
	if string(content) != testProfilesConf {
		t.Fatalf("conf not restored:\n%s", content)
	}
}

func TestProfilesForSites(t *testing.T) {
	got := profilesForSites([]detectedSite{
		{SiteName: "example.com", Type: "wordpress"},
		{SiteName: "example.com/blog", Type: "WordPress"},
		{SiteName: "example.com/cloud", Type: "nextcloud"},
		{SiteName: "example.com/shop", Type: "opencart"},
	})
	if strings.Join(got["wordpress"], ",") != "example.com,example.com/blog" || len(got["nextcloud"]) != 1 || len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestBuildProfileViewsSuggestsOnlyInactive(t *testing.T) {
	withTestPlugins(t)
	all, suggested := buildProfileViews(nil, map[string][]string{"wordpress": {"example.com"}, "nextcloud": {"example.com/cloud"}})
	if len(all) != 2 || len(suggested) != 1 || suggested[0].Key != "wordpress" {
		t.Fatalf("all=%v suggested=%v", all, suggested)
	}
	_, suggested = buildProfileViews([]string{"wordpress"}, map[string][]string{"wordpress": {"example.com"}})
	if len(suggested) != 0 {
		t.Fatal("active profile still suggested")
	}
}

func TestRenderWAFDomainPageWithProfiles(t *testing.T) {
	withTestPlugins(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	profiles, suggested := buildProfileViews(nil, map[string][]string{"wordpress": {"example.com"}})
	data := WAFDomainPageData{
		LayoutData: baseLayout(mgr, "/server/waf/example.com"), Domain: "example.com", Status: "On",
		Profiles: profiles, Suggested: suggested, Stats: wafLogStats{Checks: 12, Blocks: 3}, Level: "strict",
		Levels: []LevelView{{Key: "compatibility", Name: "Compatibility"}, {Key: "standard", Name: "Standard"}, {Key: "strict", Name: "Strict"}},
	}
	w := httptest.NewRecorder()
	if err := wafDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"wafSettings(", `&#34;Key&#34;:&#34;wordpress&#34;`, `&#34;Recommended&#34;:true`, "/server/waf/profiles/", "3 blocked", "bi bi-shield-fill-check", "width:240px", "left:-70px", `name="return_to" value="domain"`, `value="strict"`, "Protection level", "Default"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "not available on this server") {
		t.Error("empty state shown although profiles are installed")
	}
}

func TestRenderWAFDomainPageNoPluginsInstalled(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WAFDomainPageData{LayoutData: baseLayout(mgr, "/server/waf/example.com"), Domain: "example.com", Status: "Off"}
	w := httptest.NewRecorder()
	if err := wafDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "not available on this server") || !strings.Contains(body, "Firewall is off") {
		t.Error("expected empty profiles state and off status")
	}
}

func TestProfilesFromRequestMultipartAndURLEncoded(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("csrf_token", "x")
	_ = mw.WriteField("profiles", "wordpress")
	_ = mw.WriteField("profiles", "phpmyadmin")
	_ = mw.Close()
	r := httptest.NewRequest("POST", "/server/waf/profiles/example.com", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	if got := strings.Join(profilesFromRequest(r), ","); got != "wordpress,phpmyadmin" {
		t.Fatalf("multipart: got %q", got)
	}

	r = httptest.NewRequest("POST", "/server/waf/profiles/example.com", strings.NewReader("profiles=drupal"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := strings.Join(profilesFromRequest(r), ","); got != "drupal" {
		t.Fatalf("urlencoded: got %q", got)
	}

	r = httptest.NewRequest("POST", "/server/waf/profiles/example.com", strings.NewReader("csrf_token=x"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := profilesFromRequest(r); len(got) != 0 {
		t.Fatalf("empty: got %v", got)
	}
}

func TestLevels(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	ctx := context.Background()

	if got := parseWAFLevel(testProfilesConf); got != "standard" {
		t.Fatalf("no line should be standard, got %s", got)
	}
	active, level, err := setDomainProfiles(ctx, "example.com", []string{"wordpress"}, "strict")
	if err != nil || level != "strict" || len(active) != 1 {
		t.Fatalf("active=%v level=%s err=%v", active, level, err)
	}
	content, _ := os.ReadFile(path)
	block := strings.Split(string(content), "https://example.com")[1]
	setup := strings.Index(block, "crs-setup.conf.example")
	lvl := strings.Index(block, "id:"+levelRuleID+",")
	wp := strings.Index(block, "wordpress-rule-exclusions-config.conf")
	rules := strings.Index(block, "/coreruleset/rules/*.conf")
	if !(setup < lvl && lvl < wp && wp < rules) {
		t.Fatalf("level line must sit after crs-setup and before plugins/rules:\n%s", block)
	}
	if !strings.Contains(block, "blocking_paranoia_level=2") || strings.Count(string(content), "id:"+levelRuleID+",") != 2 {
		t.Fatalf("expected strict line in both blocks:\n%s", content)
	}

	// profile change with no level keeps strict
	if _, level, _ = setDomainProfiles(ctx, "example.com", nil, ""); level != "strict" {
		t.Fatalf("level lost on profile change: %s", level)
	}
	if _, level, _ = setDomainProfiles(ctx, "example.com", nil, "compatibility"); level != "compatibility" {
		t.Fatalf("got %s", level)
	}
	if _, level, _ = setDomainProfiles(ctx, "example.com", nil, "standard"); level != "standard" {
		t.Fatalf("got %s", level)
	}
	content, _ = os.ReadFile(path)
	if string(content) != testProfilesConf {
		t.Fatalf("standard with no profiles should restore the original:\n%s", content)
	}
	if _, _, err := setDomainProfiles(ctx, "example.com", nil, "paranoid"); !errors.Is(err, errUnknownLevel) {
		t.Fatalf("expected errUnknownLevel, got %v", err)
	}
}

func TestCustomLevelKeptOnProfileChange(t *testing.T) {
	withTestPlugins(t)
	custom := strings.Replace(testProfilesConf, "            Include /etc/openpanel/caddy/coreruleset/rules/*.conf",
		"            SecAction \"id:"+levelRuleID+",phase:1,pass,nolog,setvar:tx.blocking_paranoia_level=3,setvar:tx.inbound_anomaly_score_threshold=7\"\n            Include /etc/openpanel/caddy/coreruleset/rules/*.conf", -1)
	if got := parseWAFLevel(custom); got != "custom" {
		t.Fatalf("got %s", got)
	}
	out, err := rewriteProfiles(custom, []string{"wordpress"}, "")
	if err != nil || parseWAFLevel(out) != "custom" || strings.Count(out, "blocking_paranoia_level=3") != 2 {
		t.Fatalf("custom level not kept err=%v:\n%s", err, out)
	}
}

func TestSetWAFForDomains(t *testing.T) {
	path, reloads := withTestDomain(t)
	writeProfilesConf(t, path)
	changed, failed, err := setWAFForDomains(context.Background(), []string{"example.com", "missing.com"}, false)
	if err != nil || changed != 1 || len(failed) != 1 || *reloads != 1 {
		t.Fatalf("changed=%d failed=%v err=%v reloads=%d", changed, failed, err, *reloads)
	}
	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "SecRuleEngine On") || strings.Count(string(content), "SecRuleEngine Off") != 2 {
		t.Fatalf("not turned off in both blocks:\n%s", content)
	}
	if changed, _, _ = setWAFForDomains(context.Background(), []string{"example.com"}, false); changed != 0 || *reloads != 1 {
		t.Fatal("no-op should not write or reload")
	}
}

func TestDomainWAFRow(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	if _, _, err := setDomainProfiles(context.Background(), "example.com", []string{"wordpress"}, "strict"); err != nil {
		t.Fatal(err)
	}
	row := domainWAFRow("example.com", map[string][]string{"wordpress": {"example.com"}, "phpmyadmin": {"example.com/pma"}})
	if row.Status != "On" || row.Level != "strict" || strings.Join(row.Profiles, ",") != "WordPress" {
		t.Fatalf("got %+v", row)
	}
	if strings.Join(row.Recommended, ",") != "phpMyAdmin" {
		t.Fatalf("active profile recommended or detected one missing: %v", row.Recommended)
	}
}

func TestRenderWAFListPageDefaultCardAndColumns(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WAFListPageData{
		LayoutData:   baseLayout(mgr, "/server/waf"),
		Domains:      []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "off.example.com"}},
		ModsecStatus: map[string]string{"example.com": "On", "off.example.com": "Off"},
		Rows: map[string]DomainWAFRow{
			"example.com":     {Status: "On", Level: "Strict", Profiles: []string{"WordPress"}, Blocks: 7},
			"off.example.com": {Status: "Off", Level: "Standard", Recommended: []string{"Nextcloud"}},
		},
		AccountEnabled:  true,
		Recommendations: []WAFIssue{{ID: "waf-profile:off.example.com:Nextcloud", Severity: "info", Message: "Nextcloud found on off.example.com.", Link: "/server/waf/off.example.com"}},
	}
	w := httptest.NewRecorder()
	if err := wafListPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Firewall for all domains", `action="/server/waf/all"`, `name="action" value="disable"`, "Apply to all domains", "1 / 2", "Blocked (24h)", ">7</a>", "Strict", "WordPress", "reportHealthIssues('waf-profiles'", "Nextcloud found on off.example.com.", ", 24)"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "Recommended:") {
		t.Error("recommendations should be toasts, not in the profiles column")
	}
	if strings.Contains(body, `type="hidden" name="action"`) {
		t.Error("hidden action field would override the clicked button")
	}
}

func TestRenderWAFListPageFilters(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WAFListPageData{
		LayoutData:     baseLayout(mgr, "/server/waf"),
		Domains:        []appctx.Domain{{DomainURL: "example.com"}},
		ModsecStatus:   map[string]string{"example.com": "On"},
		Rows:           map[string]DomainWAFRow{"example.com": {Status: "On", Level: "Strict", LevelKey: "strict", ProfileKeys: []string{"wordpress"}, LevelStrength: 3, LevelDots: []bool{true, true, true}}},
		LevelOptions:   []LevelView{{Key: "strict", Name: "Strict"}},
		ProfileOptions: []ProfileView{{Key: "wordpress", Name: "WordPress"}},
	}
	w := httptest.NewRecorder()
	if err := wafListPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{`name="waf-status-filter"`, `name="waf-level-filter" value="strict"`, `name="waf-profile-filter" value="wordpress"`, `name="waf-profile-filter" value="none"`, "profileFilter === 'none' ? 1 === 0 : ([&#34;wordpress&#34;] || [])", `data-sort-value="3"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestCurrentProfilesForBulkModes(t *testing.T) {
	withTestPlugins(t)
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	if len(currentProfiles("example.com")) != 0 || currentProfiles("missing.com") != nil {
		t.Fatal("expected no profiles")
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", append(currentProfiles("example.com"), "wordpress"), ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := setDomainProfiles(context.Background(), "example.com", append(currentProfiles("example.com"), "phpmyadmin"), "strict"); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(currentProfiles("example.com"), ","); got != "wordpress,phpmyadmin" {
		t.Fatalf("add mode should keep existing profiles, got %s", got)
	}
}
