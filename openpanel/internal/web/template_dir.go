package web

import (
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// themeLog goes to stderr even when dev_mode is off, a page falling back to the built-in one is something the admin should see
var themeLog = log.New(os.Stderr, "", log.LstdFlags)

// pageFS is what pages parse from, the embedded templates unless template= points at a custom dir
var pageFS fs.FS = templateFS

// customTemplateDir is the dir from template=, "" when the built-in templates are used
var customTemplateDir string

// themePartials are partials/*.html the theme adds that the binary doesn't have, every base.html page gets them
var themePartials []string

// ThemesDir holds one folder per user panel theme, template= names one of them
var ThemesDir = "/etc/openpanel/openpanel/templates"

// SetTheme switches pages to the theme folder named by openpanel.config's template=, empty or invalid keeps the built-in templates
func SetTheme(name string) {
	pageFS, customTemplateDir, themePartials = templateFS, "", nil
	name = strings.TrimSpace(strings.Trim(strings.TrimSpace(name), `"'`))
	if name == "" {
		return
	}
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		log.Printf("BOOTSTRAP - template=%s is not a theme folder name, using the built-in templates", name)
		return
	}
	dir := filepath.Join(ThemesDir, name)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		log.Printf("BOOTSTRAP - theme %s not found in %s, using the built-in templates", name, ThemesDir)
		return
	}
	// readme.txt is what makes a folder a theme
	if _, err := os.Stat(filepath.Join(dir, "readme.txt")); err != nil {
		log.Printf("BOOTSTRAP - theme %s has no readme.txt, using the built-in templates", name)
		return
	}
	pageFS = overlayFS{custom: os.DirFS(dir)}
	customTemplateDir = dir
	entries, _ := os.ReadDir(filepath.Join(dir, "partials"))
	for _, e := range entries {
		rel := "templates/partials/" + e.Name()
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		if _, err := fs.Stat(templateFS, rel); err != nil {
			themePartials = append(themePartials, rel)
		}
	}
	themeLog.Printf("BOOTSTRAP - using theme %s from %s, missing files fall back to the built-in ones", name, dir)
}

// overlayFS serves templates/<name> from the custom dir when it has that file, else from the embedded copy
type overlayFS struct {
	custom fs.FS
}

func (o overlayFS) Open(name string) (fs.File, error) {
	if rel, ok := strings.CutPrefix(name, "templates/"); ok {
		if f, err := o.custom.Open(rel); err == nil {
			if info, err := f.Stat(); err == nil && !info.IsDir() {
				return f, nil
			}
			f.Close()
		}
	}
	return templateFS.Open(name)
}

// pagePaths adds the theme's own partials to a page built on base.html
func pagePaths(paths []string) []string {
	if len(themePartials) == 0 || !slices.Contains(paths, "templates/base.html") {
		return paths
	}
	out := append([]string{}, paths...)
	for _, p := range themePartials {
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// existingPaths drops _partials fsys doesn't have, a page can list a partial only the modern theme ships and still parse on the built-in templates
func existingPaths(fsys fs.FS, paths []string) []string {
	out := paths[:0:0]
	for _, p := range paths {
		if strings.HasPrefix(path.Base(p), "_") {
			if _, err := fs.Stat(fsys, p); err != nil {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

// themeStaticPath is the theme's static/<rel> on disk, "" when there's no theme or it doesn't have that file
func themeStaticPath(rel string) string {
	if customTemplateDir == "" || !fs.ValidPath(rel) {
		return ""
	}
	p := filepath.Join(customTemplateDir, "static", filepath.FromSlash(rel))
	if info, err := os.Stat(p); err != nil || info.IsDir() {
		return ""
	}
	return p
}

// ThemeClassic is the markup of the built-in templates, ThemeModern the op- kit a template= theme like modern uses
const (
	ThemeModern  = "modern"
	ThemeClassic = "classic"
)

// Theme is classic on the built-in templates and modern once a template= theme is active, handlers pass the extra data modern pages read on that
func Theme() string {
	if customTemplateDir == "" {
		return ThemeClassic
	}
	return ThemeModern
}

// ByTheme picks the classic value on the built-in templates, for icons and other markup the two draw differently
func ByTheme(modern, classic string) string {
	if Theme() == ThemeClassic {
		return classic
	}
	return modern
}
