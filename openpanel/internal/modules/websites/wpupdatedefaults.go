package websites

import (
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// WPUpdatePrefsLockedMessage is what a user gets when the admin locked the auto-update preferences
const WPUpdatePrefsLockedMessage = "Automatic update preferences are set by your hosting provider and can't be changed here."

func wpUpdateConfigValue(cfg config.Config, key, def string) string {
	v := strings.ToLower(strings.Trim(strings.TrimSpace(cfg.Get(key, def)), `"'`))
	if v == "" {
		return def
	}
	return v
}

// WPUpdatePrefsLocked is wp_auto_update_lock in openpanel.config, set from OpenAdmin > Settings > OpenPanel
func WPUpdatePrefsLocked(cfg config.Config) bool {
	return wpUpdateConfigValue(cfg, "wp_auto_update_lock", "no") == "yes"
}

// WPUpdateDefaultArgs is the `wp config set` arguments that give a new site the admin's auto-update preferences, nil when they match what WordPress does on its own
func WPUpdateDefaultArgs(cfg config.Config) [][]string {
	core := wpUpdateConfigValue(cfg, "wp_auto_update_core", "minor")
	if core != "true" && core != "false" {
		core = "minor"
	}
	plugins := wpUpdateConfigValue(cfg, "wp_auto_update_plugins", "false") == "true"
	themes := wpUpdateConfigValue(cfg, "wp_auto_update_themes", "false") == "true"
	if core == "minor" && !plugins && !themes {
		return nil
	}
	boolText := map[bool]string{true: "true", false: "false"}
	out := [][]string{
		{"config", "set", "WP_AUTO_UPDATE_PLUGINS", boolText[plugins], "--raw", "--type=constant"},
		{"config", "set", "WP_AUTO_UPDATE_THEMES", boolText[themes], "--raw", "--type=constant"},
	}
	// minor is a string in wp-config.php, true and false are booleans
	if core == "minor" {
		return append(out, []string{"config", "set", "WP_AUTO_UPDATE_CORE", "minor", "--type=constant"})
	}
	return append(out, []string{"config", "set", "WP_AUTO_UPDATE_CORE", core, "--raw", "--type=constant"})
}
