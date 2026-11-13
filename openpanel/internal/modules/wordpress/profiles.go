package wordpress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// admins manage these in OpenAdmin > Settings > Custom code
var wpProfilesDir = "/etc/openpanel/wordpress/sets/profiles"

var wpProfileSlugRE = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

// Set picked on the install form, "" keeps the global or user set, "none" skips them
const wpSetNone = "none"

type WPProfile struct {
	Slug    string   `json:"slug"`
	Name    string   `json:"name"`
	Plugins []string `json:"plugins"`
	Themes  []string `json:"themes"`
}

func loadWPProfile(slug string) (WPProfile, bool) {
	if !wpProfileSlugRE.MatchString(slug) {
		return WPProfile{}, false
	}
	raw, err := os.ReadFile(filepath.Join(wpProfilesDir, slug+".json"))
	if err != nil {
		return WPProfile{}, false
	}
	var p WPProfile
	if json.Unmarshal(raw, &p) != nil {
		return WPProfile{}, false
	}
	p.Slug = slug
	if p.Name == "" {
		p.Name = slug
	}
	return p, true
}

func listWPProfiles() []WPProfile {
	entries, _ := os.ReadDir(wpProfilesDir)
	var out []WPProfile
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".json")
		if e.IsDir() || !ok {
			continue
		}
		if p, ok := loadWPProfile(slug); ok {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}
