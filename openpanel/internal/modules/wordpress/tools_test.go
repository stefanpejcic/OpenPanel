package wordpress

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// real `wp plugin list --format=json --fields=...` from wp.tests.openpanel.org, plus one plugin with an update
const realPluginList = `[{"name":"akismet","title":"Akismet Anti-spam: Spam Protection","status":"inactive","version":"5.7.2","update":"none","update_version":"","auto_update":"off"},{"name":"hello","title":"Hello Dolly","status":"inactive","version":"1.7.2","update":"none","update_version":"","auto_update":"off"},{"name":"mu-plugin","title":"Force SSL behind Varnish","status":"must-use","version":"1.0","update":false,"update_version":"","auto_update":"off"}]`

const realThemeList = `[{"name":"twentytwentyfive","title":"Twenty Twenty-Five","status":"active","version":"1.5","update":"none","update_version":"","auto_update":"off"},{"name":"twentytwentyfour","title":"Twenty Twenty-Four","status":"inactive","version":"1.6","update":"none","update_version":"","auto_update":"off"},{"name":"twentytwentythree","title":"Twenty Twenty-Three","status":"inactive","version":"1.7","update":"none","update_version":"","auto_update":"off"}]`

func TestParseWPItems(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		wantLen     int
		wantUpdates int
		check       func(t *testing.T, items []wpItem)
		wantErr     bool
	}{
		{name: "real plugins", in: realPluginList, wantLen: 3, wantUpdates: 0, check: func(t *testing.T, items []wpItem) {
			if items[0].Title != "Akismet Anti-spam: Spam Protection" || items[0].Version != "5.7.2" || !items[0].Manageable || items[0].AutoUpdate {
				t.Errorf("akismet = %+v", items[0])
			}
			if items[2].Manageable || items[2].HasUpdate {
				t.Errorf("must-use plugin should not be manageable: %+v", items[2])
			}
		}},
		{name: "real themes", in: realThemeList + "\n", wantLen: 3},
		{name: "update available and auto-update on", in: `[{"name":"woocommerce","title":"WooCommerce","status":"active","version":"9.3.2","update":"available","update_version":"9.4.0","auto_update":"on"}]`, wantLen: 1, wantUpdates: 1, check: func(t *testing.T, items []wpItem) {
			if items[0].UpdateVersion != "9.4.0" || !items[0].AutoUpdate {
				t.Errorf("woocommerce = %+v", items[0])
			}
		}},
		{name: "newer than wordpress.org is not an update", in: `[{"name":"x","status":"active","version":"2.0","update":"version higher than expected","update_version":"1.0","auto_update":"off"}]`, wantLen: 1, check: func(t *testing.T, items []wpItem) {
			if items[0].Title != "x" {
				t.Errorf("title should fall back to name, got %q", items[0].Title)
			}
		}},
		{name: "dropin with update stays unmanaged", in: `[{"name":"object-cache.php","status":"dropin","version":"","update":"available","update_version":"2","auto_update":"off"}]`, wantLen: 1},
		{name: "odd name is not manageable", in: `[{"name":"--evil","status":"active","version":"1","update":"available","update_version":"2","auto_update":"off"}]`, wantLen: 1},
		{name: "empty list", in: `[]`, wantLen: 0},
		{name: "php warning instead of json", in: "PHP Warning: something\n", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items, err := parseWPItems([]byte(tc.in))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(items) != tc.wantLen || countUpdates(items) != tc.wantUpdates {
				t.Fatalf("got %d items, %d updates; want %d, %d", len(items), countUpdates(items), tc.wantLen, tc.wantUpdates)
			}
			if tc.check != nil {
				tc.check(t, items)
			}
		})
	}
}

func TestParseSearchReplace(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantTotal int
		wantRows  int
		wantOK    bool
	}{
		// real dry runs on wp.tests.openpanel.org
		{name: "dry run, several tables", in: "Table\tColumn\tReplacements\tType\nwp_options\toption_value\t2\tPHP\nwp_posts\tpost_content\t2\tSQL\nwp_users\tuser_url\t1\tSQL\nSuccess: 5 replacements to be made.\n", wantTotal: 5, wantRows: 3, wantOK: true},
		{name: "dry run, one replacement", in: "Table\tColumn\tReplacements\tType\nwp_posts\tpost_title\t1\tSQL\nSuccess: 1 replacement to be made.\n", wantTotal: 1, wantRows: 1, wantOK: true},
		{name: "dry run, no match", in: "Success: 0 replacements to be made.\n", wantTotal: 0, wantRows: 0, wantOK: true},
		// wp-cli's message for a real run
		{name: "real run", in: "Table\tColumn\tReplacements\tType\nwp_posts\tpost_content\t2\tSQL\nSuccess: Made 2 replacements.\n", wantTotal: 2, wantRows: 1, wantOK: true},
		{name: "real run, one", in: "Success: Made 1 replacement.\r\n", wantTotal: 1, wantOK: true},
		{name: "garbage", in: "Warning: something odd\n", wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, total, ok := parseSearchReplace(tc.in)
			if ok != tc.wantOK || total != tc.wantTotal || len(rows) != tc.wantRows {
				t.Fatalf("got rows=%v total=%d ok=%v", rows, total, ok)
			}
		})
	}
	rows, _, _ := parseSearchReplace("Table\tColumn\tReplacements\tType\nwp_options\toption_value\t2\tPHP\n")
	if rows[0] != (srRow{Table: "wp_options", Column: "option_value", Replacements: 2}) {
		t.Errorf("row = %+v", rows[0])
	}
}

func TestParseConfigConstants(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want map[string]bool
	}{
		// real output, DISALLOW_FILE_EDIT isn't defined on the test site
		{name: "real", in: `[{"name":"WP_DEBUG","value":false,"type":"constant"}]`, want: map[string]bool{"WP_DEBUG": false}},
		{name: "both set", in: `[{"name":"WP_DEBUG","value":true},{"name":"DISALLOW_FILE_EDIT","value":true}]`, want: map[string]bool{"WP_DEBUG": true, "DISALLOW_FILE_EDIT": true}},
		{name: "string and number values", in: `[{"name":"WP_DEBUG","value":"1"},{"name":"DISALLOW_FILE_EDIT","value":"'false'"}]`, want: map[string]bool{"WP_DEBUG": true, "DISALLOW_FILE_EDIT": false}},
		{name: "numeric one", in: `[{"name":"WP_DEBUG","value":1}]`, want: map[string]bool{"WP_DEBUG": true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseConfigConstants([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("%s = %v, want %v", k, got[k], v)
				}
			}
		})
	}
	if _, err := parseConfigConstants([]byte("Error: nope")); err == nil {
		t.Error("expected an error for non-JSON")
	}
}

func TestWPCLIErrorText(t *testing.T) {
	cases := []struct{ stderr, stdout, want string }{
		{"Error: No matching entries found in 'wp-config.php'.\n", "", "No matching entries found in 'wp-config.php'."},
		{"Error: Parameter errors:\n unknown --foo parameter\nDid you mean '--log'?\n", "", "Parameter errors:"},
		{"", "Warning: thing\nlast words\n", "last words"},
		{"Error: The site you have requested is not installed.\n", "Warning: x\n", "The site you have requested is not installed."},
		{"", "", "wp-cli failed"},
	}
	for _, tc := range cases {
		if got := wpCLIErrorText(tc.stderr, tc.stdout); got != tc.want {
			t.Errorf("wpCLIErrorText(%q, %q) = %q, want %q", tc.stderr, tc.stdout, got, tc.want)
		}
	}
	if got := wpCLIErrorText(strings.Repeat("x", 400), ""); len(got) != 303 {
		t.Errorf("long errors should be cut, got %d chars", len(got))
	}
}

func TestParseIDs(t *testing.T) {
	cases := map[string]int{"": 0, "\n": 0, "1": 1, "12 34 56\n": 3, "12 abc 7": 2}
	for in, want := range cases {
		if got := parseIDs(in); len(got) != want {
			t.Errorf("parseIDs(%q) = %v, want %d ids", in, got, want)
		}
	}
}

func TestSummarizeTableMaint(t *testing.T) {
	// CHECK TABLE shape from the test site's MariaDB, plus what InnoDB says to OPTIMIZE and REPAIR
	rows := [][]any{
		{[]byte("testinguser_wp1.wp_posts"), []byte("optimize"), []byte("note"), []byte("Table does not support optimize, doing recreate + analyze instead")},
		{[]byte("testinguser_wp1.wp_posts"), []byte("optimize"), []byte("status"), []byte("OK")},
		{[]byte("testinguser_wp1.wp_options"), []byte("optimize"), []byte("note"), []byte("Table does not support optimize, doing recreate + analyze instead")},
		{[]byte("testinguser_wp1.wp_options"), []byte("optimize"), []byte("status"), []byte("OK")},
		{"testinguser_wp1.wp_links", "repair", "Error", "Table 'wp_links' is marked as crashed"},
		{"short row"},
	}
	sum := summarizeTableMaint(rows)
	if sum.Tables != 3 || sum.OK != 2 || len(sum.Notes) != 1 || len(sum.Errors) != 1 {
		t.Fatalf("summary = %+v", sum)
	}
	repair := summarizeTableMaint([][]any{{"db.wp_posts", "repair", "note", "The storage engine for the table doesn't support repair"}})
	if repair.Tables != 1 || repair.OK != 0 || len(repair.Notes) != 1 {
		t.Fatalf("repair summary = %+v", repair)
	}
}

func TestQuoteTablesAndLikePrefix(t *testing.T) {
	if got := quoteTables([]string{"wp_posts", "we`ird"}); got != "`wp_posts`, `we``ird`" {
		t.Errorf("quoteTables = %s", got)
	}
	if got := likePrefix("wp_"); got != `wp\_%` {
		t.Errorf("likePrefix = %s", got)
	}
}

func TestValidSRString(t *testing.T) {
	cases := map[string]bool{
		"https://example.com":    true,
		"Hello world!":           true,
		"a=b c":                  true,
		"":                       false,
		"--url=evil":             false,
		"-x":                     false,
		"two\nlines":             false,
		"nul\x00":                false,
		strings.Repeat("a", 501): false,
	}
	for in, want := range cases {
		if got := validSRString(in); got != want {
			t.Errorf("validSRString(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDryRunGate(t *testing.T) {
	s := &wpToolSite{UserContext: "u1", Site: "example.com"}
	key := srKey(s, "http://a", "https://a")
	now := time.Now()
	if takeDryRun(key, now) {
		t.Fatal("no dry run yet")
	}
	rememberDryRun(key, now)
	if takeDryRun(srKey(s, "http://a", "https://b"), now) {
		t.Fatal("dry run for other values must not count")
	}
	if !takeDryRun(key, now.Add(time.Minute)) {
		t.Fatal("dry run should allow one real run")
	}
	if takeDryRun(key, now.Add(time.Minute)) {
		t.Fatal("dry run is used up after one run")
	}
	rememberDryRun(key, now)
	if takeDryRun(key, now.Add(srDryRunTTL+time.Second)) {
		t.Fatal("stale dry run must not count")
	}
}

func TestItemKindAndNames(t *testing.T) {
	cases := []struct {
		form    url.Values
		want    []string
		wantErr bool
	}{
		{form: url.Values{"kind": {"plugin"}, "name": {"akismet"}}, want: []string{"akismet"}},
		{form: url.Values{"kind": {"theme"}, "name": {"twentytwentyfive,twentytwentyfour", "twentytwentyfive"}}, want: []string{"twentytwentyfive", "twentytwentyfour"}},
		{form: url.Values{"kind": {"core"}, "name": {"x"}}, wantErr: true},
		{form: url.Values{"kind": {"plugin"}, "name": {"--all"}}, wantErr: true},
		{form: url.Values{"kind": {"plugin"}, "name": {"a b"}}, wantErr: true},
		{form: url.Values{"kind": {"plugin"}}, wantErr: true},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodPost, "/wordpress/tools/update", strings.NewReader(tc.form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_, names, err := itemKindAndNames(r)
		if (err != nil) != tc.wantErr {
			t.Errorf("%v: err = %v", tc.form, err)
			continue
		}
		if strings.Join(names, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%v: names = %v, want %v", tc.form, names, tc.want)
		}
	}
}

func TestSecurityReviewCount(t *testing.T) {
	yes, no := true, false
	res := wpSecurityResult{FileEditOff: &no, DebugOn: &yes, ConfigMode: "0644", ConfigModeOK: configModeOK(0o644), XMLRPCFile: true}
	res.countReview()
	if res.Review != 4 {
		t.Errorf("review = %d, want 4", res.Review)
	}
	res = wpSecurityResult{FileEditOff: &yes, DebugOn: &no, ConfigMode: "0600", ConfigModeOK: configModeOK(0o600), XMLRPCFile: true, XMLRPCBlocked: true}
	res.countReview()
	if res.Review != 0 {
		t.Errorf("review = %d, want 0", res.Review)
	}
	// unknown constants (wp-cli failed) don't count either way
	res = wpSecurityResult{ConfigError: "boom"}
	res.countReview()
	if res.Review != 0 {
		t.Errorf("review = %d, want 0", res.Review)
	}
	for mode, want := range map[uint32]bool{0o600: true, 0o400: true, 0o640: false, 0o644: false, 0o660: false} {
		if configModeOK(mode) != want {
			t.Errorf("configModeOK(%o) != %v", mode, want)
		}
	}
}

func TestWPToolsMethodGate(t *testing.T) {
	a := &appctx.App{}
	cases := []struct {
		method, action string
		want           int
	}{
		{http.MethodGet, "update", http.StatusMethodNotAllowed},
		{http.MethodGet, "search-replace", http.StatusMethodNotAllowed},
		{http.MethodGet, "config-chmod", http.StatusMethodNotAllowed},
		{http.MethodPost, "updates", http.StatusMethodNotAllowed},
		{http.MethodGet, "nope", http.StatusNotFound},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(tc.method, "/wordpress/tools/"+tc.action+"?site=example.com", nil)
		r.SetPathValue("action", tc.action)
		w := httptest.NewRecorder()
		handleWPTools(a, w, r)
		if w.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.action, w.Code, tc.want)
		}
	}
}

func TestSiteBackupDir(t *testing.T) {
	cases := []struct {
		site, date string
		want       string
		ok         bool
	}{
		{"example.com", "2026-10-29_10-00-00", "/home/u/docker-data/volumes/u_html_data/_data/backups/example.com/2026-10-29_10-00-00", true},
		{"example.com/blog", "2026-10-29_10-00-00", "/home/u/docker-data/volumes/u_html_data/_data/backups/example.com/blog/2026-10-29_10-00-00", true},
		{"example.com", "../other.com", "", false},
		{"example.com/../other.com", "2026-10-29_10-00-00", "", false},
		{"example.com", "", "", false},
	}
	for _, tc := range cases {
		got, ok := siteBackupDir("u", tc.site, tc.date)
		if ok != tc.ok || got != tc.want {
			t.Errorf("siteBackupDir(%q, %q) = %q, %v", tc.site, tc.date, got, ok)
		}
	}
	if _, ok := backupParts["files"]; !ok {
		t.Error("files part missing")
	}
	if _, ok := backupParts["../x"]; ok {
		t.Error("unexpected part")
	}
}

func TestCopyDirExcept(t *testing.T) {
	src, dst := filepath.Join(t.TempDir(), "wp-content"), filepath.Join(t.TempDir(), "wp-content")
	for _, f := range []string{"plugins/hello.php", "themes/t/style.css", "uploads/2026/10/big.jpg", "index.php"} {
		p := filepath.Join(src, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := copyDirExcept(src, dst, "uploads"); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"plugins/hello.php", "themes/t/style.css", "index.php"} {
		if _, err := os.Stat(filepath.Join(dst, f)); err != nil {
			t.Errorf("%s not copied: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "uploads", "2026")); !os.IsNotExist(err) {
		t.Error("uploads content should be skipped")
	}
	if info, err := os.Stat(filepath.Join(dst, "uploads")); err != nil || !info.IsDir() {
		t.Error("an empty uploads folder should exist")
	}
}

func TestBoolFlag(t *testing.T) {
	yes, no := true, false
	if boolFlag(nil) != "1" || boolFlag(&yes) != "1" || boolFlag(&no) != "0" {
		t.Error("boolFlag mismatch")
	}
}

// every /wordpress/tools/<action> the manager page calls has to be a known action
func TestManagerPageToolActions(t *testing.T) {
	raw, err := os.ReadFile("../../web/templates/manager/wp/single.html")
	if err != nil {
		t.Skip("template not found:", err)
	}
	matches := regexp.MustCompile(`/wordpress/tools/([a-z-]+)`).FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatal("manager page calls no tool endpoints")
	}
	for _, m := range matches {
		if !wpToolReadActions[m[1]] && !wpToolWriteActions[m[1]] {
			t.Errorf("single.html calls unknown tool action %q", m[1])
		}
	}
}
