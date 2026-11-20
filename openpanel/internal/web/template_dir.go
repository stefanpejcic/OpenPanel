package web

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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
	defaultTheme = builtinTheme
	themeCache.Range(func(k, _ any) bool { themeCache.Delete(k); return true })
	name = strings.TrimSpace(strings.Trim(strings.TrimSpace(name), `"'`))
	if name == "" {
		return
	}
	def, why := loadThemeDef(name)
	if def == nil {
		log.Printf("BOOTSTRAP - %s, using the built-in templates", why)
		return
	}
	defaultTheme = def
	pageFS, customTemplateDir, themePartials = def.fsys, def.dir, def.partials
	themeLog.Printf("BOOTSTRAP - using theme %s from %s, missing files fall back to the built-in ones", name, def.dir)
}

// themeDef is one user panel theme, the built-in templates are the one with no name
type themeDef struct {
	name     string
	dir      string
	fsys     fs.FS
	partials []string
	// menuStyle is the theme's "Menu Style:" readme header, classic or modern, it then wins over the user's and admin's choice
	menuStyle string
	once      sync.Once
	funcs     template.FuncMap
}

var builtinTheme = &themeDef{fsys: templateFS}

// defaultTheme is template= from openpanel.config, what everyone sees unless they picked their own theme
var defaultTheme = builtinTheme

// themeCache keeps the themes users picked, filled on first use
var themeCache sync.Map

// funcMap is the shared one with {{static}} pointing at this theme's files
func (d *themeDef) funcMap() template.FuncMap {
	d.once.Do(func() {
		d.funcs = template.FuncMap{}
		for k, v := range funcMap {
			d.funcs[k] = v
		}
		d.funcs["static"] = func(rel string) string { return themeStaticURL(d, rel) }
	})
	return d.funcs
}

// kit is ThemeClassic on the built-in templates and ThemeModern on any theme folder
func (d *themeDef) kit() string {
	if d == nil || d.dir == "" {
		return ThemeClassic
	}
	return ThemeModern
}

// loadThemeDef checks name is a theme folder (has readme.txt) in ThemesDir, the string says why when it isn't
func loadThemeDef(name string) (*themeDef, string) {
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return nil, "template=" + name + " is not a theme folder name"
	}
	dir := filepath.Join(ThemesDir, name)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, "theme " + name + " not found in " + ThemesDir
	}
	if _, err := os.Stat(filepath.Join(dir, "readme.txt")); err != nil {
		return nil, "theme " + name + " has no readme.txt"
	}
	def := &themeDef{name: name, dir: dir, fsys: overlayFS{custom: os.DirFS(dir)}}
	if ms := strings.ToLower(readmeHeader(filepath.Join(dir, "readme.txt"), "Menu Style")); ms == "classic" || ms == "modern" {
		def.menuStyle = ms
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "partials"))
	for _, e := range entries {
		rel := "templates/partials/" + e.Name()
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		if _, err := fs.Stat(templateFS, rel); err != nil {
			def.partials = append(def.partials, rel)
		}
	}
	return def, ""
}

// themeByName is the installed theme with that folder name, "default" is the built-in templates, anything else that isn't a theme is nil
func themeByName(name string) *themeDef {
	switch name {
	case "":
		return nil
	case "default":
		return builtinTheme
	}
	if defaultTheme.name == name {
		return defaultTheme
	}
	if _, err := os.Stat(filepath.Join(ThemesDir, name, "readme.txt")); err != nil {
		return nil
	}
	if v, ok := themeCache.Load(name); ok {
		return v.(*themeDef)
	}
	def, _ := loadThemeDef(name)
	if def == nil {
		return nil
	}
	v, _ := themeCache.LoadOrStore(name, def)
	return v.(*themeDef)
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
func (d *themeDef) pagePaths(paths []string) []string {
	if len(d.partials) == 0 || !slices.Contains(paths, "templates/base.html") {
		return paths
	}
	out := append([]string{}, paths...)
	for _, p := range d.partials {
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// pagePaths is pagePaths for the default theme
func pagePaths(paths []string) []string { return defaultTheme.pagePaths(paths) }

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

// staticPath is the theme's static/<rel> on disk, "" when it's the built-in theme or it doesn't have that file
func (d *themeDef) staticPath(rel string) string {
	if d == nil || d.dir == "" || !fs.ValidPath(rel) {
		return ""
	}
	p := filepath.Join(d.dir, "static", filepath.FromSlash(rel))
	if info, err := os.Stat(p); err != nil || info.IsDir() {
		return ""
	}
	return p
}

// themeStaticPath is staticPath for the default theme
func themeStaticPath(rel string) string { return defaultTheme.staticPath(rel) }

// ThemeClassic is the markup of the built-in templates, ThemeModern the op- kit a template= theme like modern uses
const (
	ThemeModern  = "modern"
	ThemeClassic = "classic"
)

// Theme is classic when r's page renders from the built-in templates and modern on a theme folder, handlers pass the extra data modern pages read on that
func Theme(r *http.Request) string { return RequestTheme(r).kit() }

// ByTheme picks the classic value when r's page uses the built-in templates, for icons and other markup the two draw differently
func ByTheme(r *http.Request, modern, classic string) string { return ByKit(Theme(r), modern, classic) }

// ByKit is ByTheme for code that already has the kit, like LayoutData.ThemeKit
func ByKit(kit, modern, classic string) string {
	if kit == ThemeClassic {
		return classic
	}
	return modern
}
