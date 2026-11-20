package web

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// AllowUserTheme is allow_user_theme=yes in openpanel.config, users can then pick any installed theme for themselves
var AllowUserTheme bool

// SetAllowUserTheme reads allow_user_theme, anything but yes keeps everyone on template=
func SetAllowUserTheme(v string) {
	AllowUserTheme = strings.EqualFold(strings.TrimSpace(strings.Trim(strings.TrimSpace(v), `"'`)), "yes")
}

// themeCookie tells /static which theme a browser is on, for files a theme's CSS loads by relative path
const themeCookie = "op_theme"

// ThemeOption is one choice in the user's theme picker
type ThemeOption struct {
	Folder string // "default" for the built-in templates
	Name   string
}

// InstalledThemes is Default plus every folder in ThemesDir with a readme.txt, named by its Theme Name header
func InstalledThemes() []ThemeOption {
	out := []ThemeOption{{Folder: "default", Name: "Default"}}
	entries, _ := os.ReadDir(ThemesDir)
	var themes []ThemeOption
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "default" || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		readme := filepath.Join(ThemesDir, e.Name(), "readme.txt")
		if _, err := os.Stat(readme); err != nil {
			continue
		}
		name := readmeHeader(readme, "Theme Name")
		if name == "" {
			name = e.Name()
		}
		themes = append(themes, ThemeOption{Folder: e.Name(), Name: name})
	}
	sort.Slice(themes, func(i, j int) bool { return strings.ToLower(themes[i].Name) < strings.ToLower(themes[j].Name) })
	return append(out, themes...)
}

// readmeHeader is one "Key: value" header line of a theme's readme.txt
func readmeHeader(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for i := 0; sc.Scan() && i < 30; i++ {
		if k, v, ok := strings.Cut(sc.Text(), ":"); ok && strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ThemeScreenshot is the screenshot.png of an installed theme, "" for Default and anything that isn't a theme
func ThemeScreenshot(folder string) string {
	d := themeByName(folder)
	if d == nil || d.dir == "" {
		return ""
	}
	p := filepath.Join(d.dir, "screenshot.png")
	if info, err := os.Stat(p); err != nil || info.IsDir() {
		return ""
	}
	return p
}

// userThemeFile holds the theme folder a user picked, next to their locale file
func userThemeFile(userContext string) string {
	return filepath.Join("/home", userContext, "theme")
}

// UserThemeChoice is what the user saved, "" when they never picked one
func UserThemeChoice(userContext string) string {
	if userContext == "" || strings.ContainsAny(userContext, `/\`) {
		return ""
	}
	b, err := os.ReadFile(userThemeFile(userContext))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// SaveUserTheme stores folder as the user's theme, it has to be "default" or an installed theme
func SaveUserTheme(userContext, folder string) bool {
	if userContext == "" || strings.ContainsAny(userContext, `/\`) || themeByName(folder) == nil {
		return false
	}
	return tenantfs.WriteFile(userThemeFile(userContext), []byte(folder+"\n"), 0o644) == nil
}

// userTheme is the theme a user sees, template= unless switching is on and they picked one that still exists
func userTheme(userContext string) *themeDef {
	if !AllowUserTheme {
		return defaultTheme
	}
	if def := themeByName(UserThemeChoice(userContext)); def != nil {
		return def
	}
	return defaultTheme
}

// RequestTheme is the theme r's page renders with
func RequestTheme(r *http.Request) *themeDef {
	if r == nil {
		return defaultTheme
	}
	return userTheme(auth.UserContext(r))
}

// ApplyTheme fills in the theme fields of a layout, for pages that build their LayoutData by hand
func ApplyTheme(l *LayoutData, r *http.Request) {
	d := RequestTheme(r)
	l.theme, l.ThemeKit, l.ThemeFolder, l.ThemeSwitch = d, d.kit(), d.folder(), AllowUserTheme
	if AllowUserTheme {
		l.ThemeOptions = InstalledThemes()
	}
}

// themeFolder is what the picker and the cookie call a theme
func (d *themeDef) folder() string {
	if d == nil || d.dir == "" {
		return "default"
	}
	return d.name
}

// staticTheme is the theme a /static request is for: ?t= from {{static}}, then the cookie set with the last page
func staticTheme(r *http.Request) *themeDef {
	if !AllowUserTheme {
		return defaultTheme
	}
	name := r.URL.Query().Get("t")
	if name == "" {
		if c, err := r.Cookie(themeCookie); err == nil {
			name = c.Value
		}
	}
	if def := themeByName(name); def != nil {
		return def
	}
	return defaultTheme
}

// setThemeCookie remembers the page's theme for the static files it pulls in without ?t=
func setThemeCookie(w http.ResponseWriter, d *themeDef) {
	if !AllowUserTheme {
		return
	}
	http.SetCookie(w, &http.Cookie{Name: themeCookie, Value: d.folder(), Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 31536000})
}
