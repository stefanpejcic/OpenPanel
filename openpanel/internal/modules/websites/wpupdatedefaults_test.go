package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestWPUpdateDefaultArgs(t *testing.T) {
	join := func(args [][]string) string {
		var lines []string
		for _, a := range args {
			lines = append(lines, strings.Join(a, " "))
		}
		return strings.Join(lines, "\n")
	}
	for _, tc := range []struct {
		name string
		cfg  config.Config
		want string
	}{
		{"nothing set", config.Config{}, ""},
		{"same as WordPress", config.Config{"wp_auto_update_core": "minor", "wp_auto_update_plugins": "false", "wp_auto_update_themes": "false"}, ""},
		{"unknown core value", config.Config{"wp_auto_update_core": "beta"}, ""},
		{"everything on", config.Config{"wp_auto_update_core": "true", "wp_auto_update_plugins": "true", "wp_auto_update_themes": "true"},
			"config set WP_AUTO_UPDATE_PLUGINS true --raw --type=constant\nconfig set WP_AUTO_UPDATE_THEMES true --raw --type=constant\nconfig set WP_AUTO_UPDATE_CORE true --raw --type=constant"},
		{"manual only", config.Config{"wp_auto_update_core": "false"},
			"config set WP_AUTO_UPDATE_PLUGINS false --raw --type=constant\nconfig set WP_AUTO_UPDATE_THEMES false --raw --type=constant\nconfig set WP_AUTO_UPDATE_CORE false --raw --type=constant"},
		{"plugins only, quoted", config.Config{"wp_auto_update_plugins": `"TRUE"`},
			"config set WP_AUTO_UPDATE_PLUGINS true --raw --type=constant\nconfig set WP_AUTO_UPDATE_THEMES false --raw --type=constant\nconfig set WP_AUTO_UPDATE_CORE minor --type=constant"},
	} {
		if got := join(WPUpdateDefaultArgs(tc.cfg)); got != tc.want {
			t.Errorf("%s:\n got %q\nwant %q", tc.name, got, tc.want)
		}
	}
}

func TestWPUpdatePrefsLocked(t *testing.T) {
	for value, want := range map[string]bool{"yes": true, " YES ": true, "no": false, "": false, "true": false} {
		if got := WPUpdatePrefsLocked(config.Config{"wp_auto_update_lock": value}); got != want {
			t.Errorf("lock %q = %v, want %v", value, got, want)
		}
	}
	if WPUpdatePrefsLocked(config.Config{}) {
		t.Error("the lock should be off when the key is missing")
	}
}

func TestUpdatesTabShowsTheLock(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(locked bool) string {
		pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "example.com", Docroot: "/var/www/html/example.com"}
		w := httptest.NewRecorder()
		if err := wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com"}, UpdatePrefsLocked: locked}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}
	locked, open := render(true), render(false)
	for _, want := range []string{`id="updatesLockedNote"`, `id="saveUpdatesBtn" disabled`, "Turn on auto-updates"} {
		if strings.Contains(locked, want) != (want != "Turn on auto-updates") {
			t.Errorf("locked page: unexpected state for %s", want)
		}
	}
	for _, bad := range []string{`id="updatesLockedNote"`, `id="saveUpdatesBtn" disabled`} {
		if strings.Contains(open, bad) {
			t.Errorf("unlocked page should not have %s", bad)
		}
	}
	if !strings.Contains(open, "Turn on auto-updates") {
		t.Error("unlocked page lost the per-item auto-update menu entry")
	}
}
