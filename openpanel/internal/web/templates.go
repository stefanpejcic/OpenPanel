// Package web renders the panel's pages. Templates live under
// web/templates/ (Go html/template syntax) and are embedded into the
// binary via go:embed, independent of the top-level templates/ directory.
//
// Each "page" is its own *template.Template combining a shared layout
// (e.g. user/_login.html's "layout" definition) with that page's content
// block, parsed once at startup and cached - parsing per request would
// both be wasteful and, since template.Funcs()/re-parsing mutates a
// template in place, unsafe to share across concurrent requests.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"unicode"
)

// all: is required, not just stylistic - Go's embed skips files/dirs whose
// name starts with "_" or "." by default, and partials in this tree follow
// exactly that convention (e.g. user/_login.html).
//
//go:embed all:templates
var templateFS embed.FS

// funcMap holds only stateless helpers. Anything request-scoped (the
// translator, CSRF token, flashes) is passed as template data instead - see
// internal/core/i18n's Translator doc comment for why.
var funcMap = template.FuncMap{
	"static": func(path string) string { return "/static/" + path },
	// firstUpper returns the first character of s, uppercased - used for
	// the letter-avatar initial.
	"firstUpper": func(s string) string {
		for _, r := range s {
			return string(unicode.ToUpper(r))
		}
		return ""
	},
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
	// contains reports whether the first argument contains the second as a
	// substring.
	"contains": strings.Contains,
	// title uppercases the first letter of s and lowercases the rest (not
	// Go's strings.Title, which title-cases every word).
	"title": func(s string) string {
		if s == "" {
			return s
		}
		lower := strings.ToLower(s)
		r := []rune(lower)
		return strings.ToUpper(string(r[0])) + string(r[1:])
	},
	// replace replaces every occurrence of old with newStr in s, not just
	// the first.
	"replace":   func(s, old, newStr string) string { return strings.ReplaceAll(s, old, newStr) },
	"split":     strings.Split,
	"trimSpace": strings.TrimSpace,
	"join":      strings.Join,
	"lower":     strings.ToLower,
	"upper":     strings.ToUpper,
	// truncate returns the first n bytes of s, or the whole string if it's
	// already shorter.
	"truncate": func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n]
	},
	// isSystemMySQLUser mirrors the hardcoded system-account list in
	// mysql/users.html: these usernames get a "System User" badge instead
	// of password-change/delete actions.
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
	// dict builds a map[string]any from alternating key/value arguments,
	// for passing multiple values into a {{template}} call (which only
	// accepts a single pipeline argument).
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
	// toJSON marshals a value to a JSON literal for embedding in Alpine.js
	// x-data attributes built from server-side values.
	"toJSON": func(v any) template.JS {
		b, err := json.Marshal(v)
		if err != nil {
			return "null"
		}
		return template.JS(b) //nolint:gosec // values passed through this are our own server-computed data (feature flags, section keys), not raw user input
	},
}

// Page is a parsed, ready-to-execute template set for one route.
type Page struct {
	tmpl *template.Template
}

// MustLoadPage parses files (relative to web/templates/) into a single
// template set. By convention the layout file among them defines a
// top-level {{define "layout"}} that Render executes. Panics on a parse
// error, since a broken template is a startup-time bug, not a runtime
// condition to recover from.
func MustLoadPage(files ...string) *Page {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = "templates/" + f
	}
	t := template.Must(template.New("layout").Funcs(funcMap).ParseFS(templateFS, paths...))
	return &Page{tmpl: t}
}

func (p *Page) Render(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return p.tmpl.ExecuteTemplate(w, "layout", data)
}

// MustLoadFragment parses files the same way MustLoadPage does (same
// FuncMap, same embedded FS), but without a "layout" wrapper to execute -
// for standalone fragments like an email body that get rendered to a
// buffer via ExecuteTemplate(name, ...) rather than served as a page.
func MustLoadFragment(files ...string) *template.Template {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = "templates/" + f
	}
	return template.Must(template.New("fragment").Funcs(funcMap).ParseFS(templateFS, paths...))
}
