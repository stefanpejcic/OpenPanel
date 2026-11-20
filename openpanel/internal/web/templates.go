// Package web renders the panel's pages. Templates live under web/templates/ (Go html/template syntax) and are embedded into the binary via go:embed, independent of the top-level templates/ directory.
// Each "page" is its own *template.Template combining a shared layout with that page's content block, parsed on first use and cached - parsing per request would be wasteful and, since template.Funcs()/re-parsing mutates a template in place, unsafe to share across concurrent requests.
package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

// all: is required, not just stylistic - Go's embed skips files/dirs whose name starts with "_" or "." by default, and partials in this tree follow exactly that convention (e.g. user/_login.html).
//
//go:embed all:templates
var templateFS embed.FS

// funcMap holds only stateless helpers - anything request-scoped (the translator, CSRF token, flashes) is passed as template data instead, see internal/core/i18n's Translator doc comment for why.
var funcMap = template.FuncMap{
	"static": StaticURL,
	// admin custom code from in_header.html/in_footer.html, loaded at startup
	"customHeader": func() template.HTML { return customHeaderHTML },
	"customFooter": func() template.HTML { return customFooterHTML },
	// firstUpper returns the first character of s, uppercased - used for the letter-avatar initial
	"firstUpper": func(s string) string {
		for _, r := range s {
			return string(unicode.ToUpper(r))
		}
		return ""
	},
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
	// contains reports whether the first argument contains the second as a substring
	"contains": strings.Contains,
	// title uppercases the first letter of s and lowercases the rest (not Go's strings.Title, which title-cases every word)
	"title": func(s string) string {
		if s == "" {
			return s
		}
		lower := strings.ToLower(s)
		r := []rune(lower)
		return strings.ToUpper(string(r[0])) + string(r[1:])
	},
	// replace replaces every occurrence of old with newStr in s, not just the first
	"replace":   func(s, old, newStr string) string { return strings.ReplaceAll(s, old, newStr) },
	"split":     strings.Split,
	"trimSpace": strings.TrimSpace,
	"join":      strings.Join,
	"lower":     strings.ToLower,
	"upper":     strings.ToUpper,
	// truncate returns the first n bytes of s, or the whole string if it's already shorter
	"truncate": func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n]
	},
	// isSystemMySQLUser mirrors the hardcoded system-account list in mysql/users.html: these usernames get a "System User" badge instead of password-change/delete actions
	"isSystemMySQLUser": func(user string) bool {
		switch user {
		case "mysql.sys", "mysql", "sys", "mariadb.sys", "phpmyadmin", "mysql.session", "mysql.infoschema", "root", "healthcheck", "debian-sys-maint":
			return true
		default:
			return false
		}
	},
	"add":      func(a, b int) int { return a + b },
	"sub":      func(a, b int) int { return a - b },
	"mul":      func(a, b int) int { return a * b },
	"toString": func(v any) string { return fmt.Sprint(v) },
	// dict builds a map[string]any from alternating key/value arguments, for passing multiple values into a {{template}} call (which only accepts a single pipeline argument)
	"dict": func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, fmt.Errorf("dict: odd number of arguments")
		}
		m := make(map[string]any, len(pairs)/2)
		for i := 0; i < len(pairs); i += 2 {
			key, ok := pairs[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict: key %v is not a string", pairs[i])
			}
			m[key] = pairs[i+1]
		}
		return m, nil
	},
	// toJSON marshals a value to a JSON literal for embedding in Alpine.js x-data attributes built from server-side values
	"toJSON": func(v any) template.JS {
		b, err := json.Marshal(v)
		if err != nil {
			return "null"
		}
		return template.JS(b) //nolint:gosec // values passed through this are our own server-computed data (feature flags, section keys), not raw user input
	},
}

const bulkPartial = "partials/_bulk.html"

const dateRangePartial = "partials/_date_range.html"

const aiPartial = "partials/_ai_assistant.html"

// Page is the template set for one route, parsed on first use since parsing all of them up front cost ~115MB
type Page struct {
	paths []string
	mu    sync.Mutex
	// one parsed set per theme folder, "default" is the built-in templates
	tmpl map[string]*template.Template
	used atomic.Int64
}

// pageCacheSize is how many parsed pages stay in memory, the least recently used one gets parsed again on its next visit
const pageCacheSize = 32

var (
	pagesMu     sync.Mutex
	allPages    []*Page
	loadedPages []*Page
	pageClock   atomic.Int64
)

// MustLoadPage registers files (relative to web/templates/) as one template set - by convention the layout file among them defines a top-level {{define "layout"}} that Render executes. Parsing waits for the first Render, ParseAllPages covers broken templates in tests.
func MustLoadPage(files ...string) *Page {
	paths := make([]string, 0, len(files)+1)
	hasBase, hasBulk, hasDateRange, hasHeader, hasAI := false, false, false, false, false
	for _, f := range files {
		paths = append(paths, "templates/"+f)
		hasBase = hasBase || f == "base.html"
		hasBulk = hasBulk || f == bulkPartial
		hasDateRange = hasDateRange || f == dateRangePartial
		hasHeader = hasHeader || f == "partials/_header.html"
		hasAI = hasAI || f == aiPartial
	}
	// every panel page can use the bulk selection partial without listing it
	if hasBase && !hasBulk {
		paths = append(paths, "templates/"+bulkPartial)
	}
	if hasBase && !hasDateRange {
		paths = append(paths, "templates/"+dateRangePartial)
	}
	// the header calls the AI drawer, so it comes along with it
	if hasHeader && !hasAI {
		paths = append(paths, "templates/"+aiPartial)
	}
	p := &Page{paths: paths}
	pagesMu.Lock()
	allPages = append(allPages, p)
	pagesMu.Unlock()
	return p
}

func (p *Page) parse(d *themeDef) (*template.Template, error) {
	t, err := template.New("layout").Funcs(d.funcMap()).ParseFS(d.fsys, existingPaths(d.fsys, d.pagePaths(p.paths))...)
	if err != nil && d.dir != "" {
		// a broken custom template shouldn't take the page down, use the built-in one instead
		themeLog.Printf("theme %s: %s failed to parse, using the built-in page: %v", d.name, p.name(), err)
		return template.New("layout").Funcs(builtinTheme.funcMap()).ParseFS(templateFS, existingPaths(templateFS, p.paths)...)
	}
	return t, err
}

// template returns the set parsed for theme d, parsing it now if it isn't cached
func (p *Page) template(d *themeDef) (*template.Template, error) {
	p.used.Store(pageClock.Add(1))
	key := d.folder()
	p.mu.Lock()
	t := p.tmpl[key]
	fresh := t == nil
	if fresh {
		var err error
		if t, err = p.parse(d); err != nil {
			// leave the cache empty and unlock, a panic here used to keep the page locked forever
			p.mu.Unlock()
			return nil, err
		}
		if p.tmpl == nil {
			p.tmpl = map[string]*template.Template{}
		}
		p.tmpl[key] = t
	}
	p.mu.Unlock()
	if fresh {
		trackLoadedPage(p)
	}
	return t, nil
}

// trackLoadedPage remembers p as parsed and drops the least recently used page once there are too many
func trackLoadedPage(p *Page) {
	pagesMu.Lock()
	defer pagesMu.Unlock()
	for _, q := range loadedPages {
		if q == p {
			return
		}
	}
	loadedPages = append(loadedPages, p)
	if len(loadedPages) <= pageCacheSize {
		return
	}
	oldest := 0
	for i, q := range loadedPages {
		if q.used.Load() < loadedPages[oldest].used.Load() {
			oldest = i
		}
	}
	victim := loadedPages[oldest]
	loadedPages = append(loadedPages[:oldest], loadedPages[oldest+1:]...)
	victim.mu.Lock()
	victim.tmpl = nil
	victim.mu.Unlock()
}

// ParseAllPages parses every page without caching it, so tests still catch a broken template
func ParseAllPages() error {
	pagesMu.Lock()
	pages := append([]*Page{}, allPages...)
	pagesMu.Unlock()
	var errs []error
	for _, p := range pages {
		if _, err := p.parse(defaultTheme); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// layoutFiles is the standard authenticated-layout template set every panel page is built on
var layoutFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"partials/install_success.html",
	"partials/folder_picker.html",
	"partials/filter_chips.html",
}

// LoadPage is MustLoadPage with the standard layout files prepended
func LoadPage(files ...string) *Page {
	return MustLoadPage(append(append([]string{}, layoutFiles...), files...)...)
}

// themed is LayoutData, any page data built on it says which theme it's for
type themed interface{ pageTheme() *themeDef }

func (p *Page) Render(w http.ResponseWriter, status int, data any) error {
	d := defaultTheme
	if td, ok := data.(themed); ok {
		d = td.pageTheme()
	}
	t, err := p.template(d)
	if err != nil {
		http.Error(w, "page template failed to load", http.StatusInternalServerError)
		return err
	}
	setThemeCookie(w, d)
	if d.dir == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		return t.ExecuteTemplate(w, "layout", data)
	}
	// themed pages render to a buffer first so one that needs data this build doesn't pass can fall back to the built-in page
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		themeLog.Printf("theme %s: %s failed to render, using the built-in page: %v", d.name, p.name(), err)
		if bt, perr := p.template(builtinTheme); perr == nil {
			buf.Reset()
			err = bt.ExecuteTemplate(&buf, "layout", data)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
	return err
}

// name is the page's own template file, the last one that isn't a shared partial or layout
func (p *Page) name() string {
	for i := len(p.paths) - 1; i >= 0; i-- {
		if !strings.HasPrefix(p.paths[i], "templates/partials/") && p.paths[i] != "templates/base.html" {
			return p.paths[i]
		}
	}
	return p.paths[len(p.paths)-1]
}

// Fragment is a standalone template set like an email body, rendered with ExecuteTemplate(name, ...) instead of a "layout" wrapper
type Fragment struct {
	paths []string
	once  sync.Once
	tmpl  *template.Template
}

// MustLoadFragment checks files parse from the embedded FS right away, the set actually used is parsed on first use so template= is already applied
func MustLoadFragment(files ...string) *Fragment {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = "templates/" + f
	}
	template.Must(template.New("fragment").Funcs(funcMap).ParseFS(templateFS, paths...))
	return &Fragment{paths: paths}
}

func (f *Fragment) ExecuteTemplate(w io.Writer, name string, data any) error {
	f.once.Do(func() {
		t, err := template.New("fragment").Funcs(defaultTheme.funcMap()).ParseFS(defaultTheme.fsys, f.paths...)
		if err != nil {
			themeLog.Printf("theme %s: fragment failed to parse, using the built-in one: %v", defaultTheme.name, err)
			t = template.Must(template.New("fragment").Funcs(funcMap).ParseFS(templateFS, f.paths...))
		}
		f.tmpl = t
	})
	return f.tmpl.ExecuteTemplate(w, name, data)
}
