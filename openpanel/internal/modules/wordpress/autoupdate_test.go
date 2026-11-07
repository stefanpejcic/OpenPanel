package wordpress

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseUpdatePolicy(t *testing.T) {
	cases := []struct {
		name, in string
		want     wpUpdatePolicy
	}{
		{"none set", `[]`, wpUpdatePolicy{Core: "minor"}},
		{"panel form defaults", `[{"name":"WP_AUTO_UPDATE_PLUGINS","value":false,"type":"constant"},{"name":"WP_AUTO_UPDATE_THEMES","value":false,"type":"constant"},{"name":"WP_AUTO_UPDATE_CORE","value":"minor","type":"constant"}]`, wpUpdatePolicy{Core: "minor"}},
		{"all on", `[{"name":"WP_AUTO_UPDATE_PLUGINS","value":true,"type":"constant"},{"name":"WP_AUTO_UPDATE_THEMES","value":true,"type":"constant"},{"name":"WP_AUTO_UPDATE_CORE","value":true,"type":"constant"}]`, wpUpdatePolicy{Core: "true", Plugins: true, Themes: true}},
		{"core off", `[{"name":"WP_AUTO_UPDATE_CORE","value":false,"type":"constant"}]`, wpUpdatePolicy{Core: "false"}},
		{"quoted strings", `[{"name":"WP_AUTO_UPDATE_CORE","value":"'true'","type":"constant"},{"name":"WP_AUTO_UPDATE_PLUGINS","value":"1","type":"constant"}]`, wpUpdatePolicy{Core: "true", Plugins: true}},
		{"beta counts as minor", `[{"name":"WP_AUTO_UPDATE_CORE","value":"beta","type":"constant"}]`, wpUpdatePolicy{Core: "minor"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseUpdatePolicy([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %+v want %+v", got, tc.want)
			}
		})
	}
	if _, err := parseUpdatePolicy([]byte("Error: No matching entries found in 'wp-config.php'.")); err == nil {
		t.Error("expected an error for non-JSON")
	}
}

func TestPolicyCoreArgs(t *testing.T) {
	cases := []struct {
		core       string
		check, upd []string
	}{
		{"false", nil, nil},
		{"minor", []string{"core", "check-update", "--minor", "--format=json", "--force-check"}, []string{"core", "update", "--minor"}},
		{"true", []string{"core", "check-update", "--format=json", "--force-check"}, []string{"core", "update"}},
	}
	for _, tc := range cases {
		p := wpUpdatePolicy{Core: tc.core}
		if got := coreCheckArgs(p); !reflect.DeepEqual(got, tc.check) {
			t.Errorf("%s: check args %v, want %v", tc.core, got, tc.check)
		}
		if got := coreUpdateArgs(p); !reflect.DeepEqual(got, tc.upd) {
			t.Errorf("%s: update args %v, want %v", tc.core, got, tc.upd)
		}
	}
}

func TestParseCoreCheck(t *testing.T) {
	cases := []struct{ in, want string }{
		{"[]", ""},
		{"", ""},
		{"Success: WordPress is at the latest version.", ""},
		{`[{"version":"7.1.5","update_type":"minor","package_url":"https://downloads.wordpress.org/release/wordpress-7.1.5-partial-4.zip"}]`, "7.1.5"},
		{`[{"version":"7.2","update_type":"major","package_url":"x"},{"version":"7.1.5","update_type":"minor","package_url":"y"}]`, "7.2"},
	}
	for _, tc := range cases {
		got, err := parseCoreCheck([]byte(tc.in))
		if err != nil || got != tc.want {
			t.Errorf("parseCoreCheck(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
}

func TestPlanCommands(t *testing.T) {
	plugins := []wpItem{
		{Name: "akismet", Title: "Akismet Anti-spam", HasUpdate: true, AutoUpdate: true},
		{Name: "hello", Title: "Hello Dolly", HasUpdate: true},
		{Name: "classic-editor", Title: "Classic Editor", AutoUpdate: true},
	}
	names, titles := pickAutoItems(plugins, false)
	if !reflect.DeepEqual(names, []string{"akismet"}) || !reflect.DeepEqual(titles, []string{"Akismet Anti-spam"}) {
		t.Errorf("per-item pick = %v %v", names, titles)
	}
	names, _ = pickAutoItems(plugins, true)
	if !reflect.DeepEqual(names, []string{"akismet", "hello"}) {
		t.Errorf("all pick = %v", names)
	}

	plan := autoUpdatePlan{CoreArgs: []string{"core", "update", "--minor"}, Plugins: []string{"akismet", "hello"}, Themes: []string{"twentytwentyfive"}}
	want := [][]string{
		{"core", "update", "--minor", "--skip-plugins", "--skip-themes"},
		{"plugin", "update", "akismet", "hello", "--skip-plugins", "--skip-themes"},
		{"theme", "update", "twentytwentyfive", "--skip-plugins", "--skip-themes"},
	}
	if got := plan.commands(); !reflect.DeepEqual(got, want) {
		t.Errorf("commands = %v", got)
	}
	if !(autoUpdatePlan{}).empty() || plan.empty() {
		t.Error("empty() is wrong")
	}
}

func TestAutoUpdateState(t *testing.T) {
	root := filepath.Join(t.TempDir(), "backups", "example.com")
	st, err := readAutoUpdateState(root)
	if err != nil || st.Enabled || st.LastRun != 0 {
		t.Fatalf("missing file should read as off: %+v %v", st, err)
	}
	res := &autoUpdateResult{When: 1761750000, Backup: "2026-10-29_16-39-37", Updated: []string{"Akismet Anti-spam"}}
	if err := writeAutoUpdateState(root, autoUpdateState{Enabled: true, LastRun: 1761750000, LastResult: res}); err != nil {
		t.Fatal(err)
	}
	st, err = readAutoUpdateState(root)
	if err != nil || !st.Enabled || st.LastRun != 1761750000 || st.LastResult == nil || st.LastResult.Backup != res.Backup || st.LastResult.Updated[0] != "Akismet Anti-spam" {
		t.Fatalf("round trip lost data: %+v %v", st, err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "auto-update.json"))
	for _, key := range []string{`"enabled": true`, `"last_run": 1761750000`, `"last_result"`, `"when"`, `"backup"`, `"updated"`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("auto-update.json is missing %s:\n%s", key, raw)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "auto-update.json.tmp")); err == nil {
		t.Error("temp file left behind")
	}

	st, err = changeAutoUpdateState(root, func(st *autoUpdateState) { st.Enabled = false })
	if err != nil || st.Enabled || st.LastResult == nil {
		t.Fatalf("change should keep the rest: %+v %v", st, err)
	}

	_ = os.WriteFile(filepath.Join(root, "auto-update.json"), []byte("{nope"), 0o644)
	if _, err := readAutoUpdateState(root); err == nil {
		t.Error("broken file should error on read")
	}
	if st, err := changeAutoUpdateState(root, func(st *autoUpdateState) { st.Enabled = true }); err != nil || !st.Enabled {
		t.Errorf("a broken file should be rewritten: %+v %v", st, err)
	}
}

func TestAutoUpdateDue(t *testing.T) {
	now := time.Unix(1761750000, 0)
	cases := []struct {
		st   autoUpdateState
		want bool
	}{
		{autoUpdateState{}, false},
		{autoUpdateState{Enabled: true}, true},
		{autoUpdateState{Enabled: true, LastRun: now.Add(-23 * time.Hour).Unix()}, false},
		{autoUpdateState{Enabled: true, LastRun: now.Add(-25 * time.Hour).Unix()}, true},
		{autoUpdateState{Enabled: false, LastRun: now.Add(-48 * time.Hour).Unix()}, false},
	}
	for i, tc := range cases {
		if got := autoUpdateDue(tc.st, now); got != tc.want {
			t.Errorf("case %d: due = %v, want %v", i, got, tc.want)
		}
	}
}

func TestPruneAutoBackups(t *testing.T) {
	root := t.TempDir()
	mk := func(name string, tagged bool) {
		dir := filepath.Join(root, name)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "database.sql"), []byte("--"), 0o644)
		if tagged {
			_ = os.WriteFile(filepath.Join(dir, autoUpdateMarker), []byte("x"), 0o644)
		}
	}
	tagged := []string{"2026-10-20_03-00-00", "2026-10-21_03-00-00", "2026-10-22_03-00-00", "2026-10-23_03-00-00", "2026-10-24_03-00-00"}
	for _, n := range tagged {
		mk(n, true)
	}
	mk("2026-10-01_12-00-00", false)
	mk("2026-10-25_12-00-00", false)
	mk("not-a-backup", true)
	_ = os.WriteFile(filepath.Join(root, autoUpdateStateName), []byte("{}"), 0o644)

	removed := pruneAutoBackups(root, 3)
	if !reflect.DeepEqual(removed, tagged[:2]) {
		t.Errorf("removed %v, want %v", removed, tagged[:2])
	}
	entries, _ := os.ReadDir(root)
	var left []string
	for _, e := range entries {
		left = append(left, e.Name())
	}
	want := []string{"2026-10-01_12-00-00", "2026-10-22_03-00-00", "2026-10-23_03-00-00", "2026-10-24_03-00-00", "2026-10-25_12-00-00", autoUpdateStateName, "not-a-backup"}
	slices.Sort(left)
	slices.Sort(want)
	if !reflect.DeepEqual(left, want) {
		t.Errorf("left %v, want %v", left, want)
	}
	if again := pruneAutoBackups(root, 3); len(again) != 0 {
		t.Errorf("second prune removed %v", again)
	}
}

// fakeRunner answers wp-cli calls by their first two args and records every call
type fakeRunner struct {
	replies   map[string]string
	errs      map[string]error
	backupErr error
	calls     []string
	backedUp  bool
}

func (f *fakeRunner) wp(_ context.Context, _ time.Duration, args ...string) (string, error) {
	f.calls = append(f.calls, strings.Join(args, " "))
	key := args[0] + " " + args[1]
	return f.replies[key], f.errs[key]
}

func (f *fakeRunner) backup(context.Context) (string, error) {
	f.backedUp = true
	if f.backupErr != nil {
		return "", f.backupErr
	}
	return "/home/u/docker-data/volumes/u_html_data/_data/backups/example.com/2026-10-29_03-00-00", nil
}

func (f *fakeRunner) ran(prefix string) bool {
	return slices.ContainsFunc(f.calls, func(c string) bool { return strings.HasPrefix(c, prefix) })
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		replies: map[string]string{
			"config list":       `[{"name":"WP_AUTO_UPDATE_PLUGINS","value":false,"type":"constant"},{"name":"WP_AUTO_UPDATE_CORE","value":"minor","type":"constant"}]`,
			"core check-update": `[]`,
			"plugin list":       `[{"name":"akismet","title":"Akismet Anti-spam","status":"inactive","version":"5.7.1","update":"available","update_version":"5.7.2","auto_update":"on"},{"name":"hello","title":"Hello Dolly","status":"inactive","version":"1.7.1","update":"available","update_version":"1.7.2","auto_update":"off"}]`,
			"theme list":        `[{"name":"twentytwentyfive","title":"Twenty Twenty-Five","status":"active","version":"1.5","update":"none","update_version":"","auto_update":"off"}]`,
			"plugin update":     "Success: Updated 1 of 1 plugins.",
		},
		errs: map[string]error{},
	}
}

func TestRunAutoUpdateBackupGate(t *testing.T) {
	f := newFakeRunner()
	f.backupErr = errors.New("mariadb-dump failed: exit status 2")
	res := runAutoUpdate(context.Background(), f, time.Unix(1761750000, 0))
	if !f.backedUp {
		t.Fatal("backup wasn't attempted")
	}
	if f.ran("plugin update") || f.ran("theme update") || f.ran("core update") {
		t.Errorf("updates ran after a failed backup: %v", f.calls)
	}
	if !strings.HasPrefix(res.Error, "Backup failed, updates skipped: mariadb-dump failed") || len(res.Updated) != 0 || res.Backup != "" {
		t.Errorf("result = %+v", res)
	}
}

func TestRunAutoUpdateFollowsPolicy(t *testing.T) {
	f := newFakeRunner()
	res := runAutoUpdate(context.Background(), f, time.Unix(1761750000, 0))
	if res.Stamp != time.Unix(1761750000, 0).Format("2006-01-02_15-04-05") {
		t.Errorf("stamp = %q", res.Stamp)
	}
	if res.Error != "" || res.Backup != "2026-10-29_03-00-00" || !reflect.DeepEqual(res.Updated, []string{"Akismet Anti-spam"}) {
		t.Errorf("result = %+v", res)
	}
	if !slices.Contains(f.calls, "plugin update akismet --skip-plugins --skip-themes") {
		t.Errorf("expected only akismet to be updated: %v", f.calls)
	}
	if f.ran("core update") || f.ran("theme update") {
		t.Errorf("nothing else was due: %v", f.calls)
	}
	backupAt := slices.IndexFunc(f.calls, func(c string) bool { return strings.HasPrefix(c, "plugin update") })
	if backupAt < 0 || !f.backedUp {
		t.Error("backup should run before the update")
	}

	f = newFakeRunner()
	f.replies["core check-update"] = `[{"version":"7.1.5","update_type":"minor","package_url":"x"}]`
	f.replies["config list"] = `[{"name":"WP_AUTO_UPDATE_PLUGINS","value":true,"type":"constant"}]`
	f.errs["plugin update"] = errors.New("Could not create directory.")
	res = runAutoUpdate(context.Background(), f, time.Unix(1761750000, 0))
	if !slices.Contains(f.calls, "core update --minor --skip-plugins --skip-themes") || !f.ran("core update-db") {
		t.Errorf("expected a minor core update: %v", f.calls)
	}
	if !slices.Contains(f.calls, "plugin update akismet hello --skip-plugins --skip-themes") {
		t.Errorf("all plugins should update when WP_AUTO_UPDATE_PLUGINS is true: %v", f.calls)
	}
	if !reflect.DeepEqual(res.Updated, []string{"WordPress 7.1.5"}) || res.Error != "Plugins: Could not create directory." {
		t.Errorf("result = %+v", res)
	}
}

func TestRunAutoUpdateNothingDue(t *testing.T) {
	f := newFakeRunner()
	f.replies["plugin list"] = `[]`
	f.errs["config list"] = errors.New("No matching entries found in 'wp-config.php'.")
	res := runAutoUpdate(context.Background(), f, time.Unix(1761750000, 0))
	if f.backedUp || res.Note == "" || res.Error != "" {
		t.Errorf("no backup should be taken when nothing is due: %+v %v", res, f.calls)
	}

	f = newFakeRunner()
	f.errs["plugin list"] = errors.New("Error establishing a database connection.")
	res = runAutoUpdate(context.Background(), f, time.Unix(1761750000, 0))
	if f.backedUp || !strings.Contains(res.Error, "database connection") {
		t.Errorf("a failed list should stop before the backup: %+v", res)
	}
}

func TestBackupThenUpdate(t *testing.T) {
	var phases []string
	phase := func(p string) { phases = append(phases, p) }
	updated := false
	update := func(context.Context) (string, error) { updated = true; return "Success: Updated 1 of 1 plugins.", nil }

	_, _, err := backupThenUpdate(context.Background(), true, func(context.Context) (string, error) { return "", errors.New("Failed to retrieve table prefix.") }, update, phase)
	var gate *backupGateError
	if !errors.As(err, &gate) || updated {
		t.Fatalf("update ran after a failed backup (err %v)", err)
	}
	if err.Error() != "Backup failed, nothing was updated: Failed to retrieve table prefix." {
		t.Errorf("error = %q", err)
	}

	phases = nil
	name, out, err := backupThenUpdate(context.Background(), true, func(context.Context) (string, error) {
		return "/home/u/docker-data/volumes/u_html_data/_data/backups/example.com/2026-10-29_16-39-37", nil
	}, update, phase)
	if err != nil || !updated || name != "2026-10-29_16-39-37" || out == "" || !reflect.DeepEqual(phases, []string{"backup", "update"}) {
		t.Errorf("got %q %q %v %v", name, out, err, phases)
	}

	phases, updated = nil, false
	backedUp := false
	name, _, err = backupThenUpdate(context.Background(), false, func(context.Context) (string, error) { backedUp = true; return "", nil }, update, phase)
	if err != nil || backedUp || !updated || name != "" || !reflect.DeepEqual(phases, []string{"update"}) {
		t.Errorf("without backup: %q %v backedUp=%v phases=%v", name, err, backedUp, phases)
	}
}

func TestTruthy(t *testing.T) {
	for _, v := range []any{true, 1.0, "1", "true", "TRUE"} {
		if !truthy(v) {
			t.Errorf("truthy(%v) = false", v)
		}
	}
	for _, v := range []any{nil, false, 0.0, "0", "", "yes"} {
		if truthy(v) {
			t.Errorf("truthy(%v) = true", v)
		}
	}
}
