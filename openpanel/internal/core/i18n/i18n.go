// Package i18n reuses the panel's existing gettext .po/.mo catalogs (no
// retranslation needed) and resolves the active locale via a priority
// chain: session -> per-account locale file -> Accept-Language -> system
// default -> "en".
package i18n

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/leonelquinteros/gotext"

	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

const (
	TranslationsDir   = "/etc/openpanel/openpanel/translations/"
	DefaultLocaleFile = "/etc/openpanel/openpanel/default_locale"
	FallbackLocale    = "en"
	domain            = "messages"
)

type Manager struct {
	dir   string
	cache *cache.Cache

	mu      sync.RWMutex
	locales map[string]*gotext.Locale
}

func NewManager(dir string, c *cache.Cache) *Manager {
	return &Manager{dir: dir, cache: c, locales: make(map[string]*gotext.Locale)}
}

// AvailableLocales returns every subdirectory of dir containing
// LC_MESSAGES/messages.po, plus "en", cached for 1h.
//
// Note: `opencli locale` busts the cached locale list after installing a
// new locale, but that cache-busting doesn't reach this Go cache entry, so
// a freshly installed locale won't show up here until this entry's own 1h
// TTL expires. Revisit when opencli is updated for the Go binary.
func (m *Manager) AvailableLocales(ctx context.Context) []string {
	locales, _ := cache.Memoize(ctx, m.cache, "app.get_available_locales", time.Hour, func() ([]string, error) {
		return m.scanLocales(), nil
	})
	if len(locales) == 0 {
		return []string{FallbackLocale}
	}
	return locales
}

func (m *Manager) scanLocales() []string {
	locales := []string{FallbackLocale}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return locales
	}

	for _, e := range entries {
		if !e.IsDir() || e.Name() == FallbackLocale {
			continue
		}
		poPath := filepath.Join(m.dir, e.Name(), "LC_MESSAGES", "messages.po")
		if _, err := os.Stat(poPath); err == nil {
			locales = append(locales, e.Name())
		}
	}

	return locales
}

// SystemDefaultLocale reads the system-wide default locale, cached for 24h.
func (m *Manager) SystemDefaultLocale(ctx context.Context) string {
	locale, _ := cache.Memoize(ctx, m.cache, "app.get_system_default_locale", 24*time.Hour, func() (string, error) {
		data, err := os.ReadFile(DefaultLocaleFile)
		if err != nil {
			return FallbackLocale, nil
		}
		if v := strings.TrimSpace(string(data)); v != "" {
			return v, nil
		}
		return FallbackLocale, nil
	})
	if locale == "" {
		return FallbackLocale
	}
	return locale
}

// UserLocale reads /home/<systemUsername>/locale, the per-account override
// checked before falling back to Accept-Language. Returns "" if unset, so
// the caller can keep checking the rest of the priority chain.
func UserLocale(systemUsername string) string {
	if systemUsername == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join("/home", systemUsername, "locale"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ResolveLocale runs the locale priority chain. sessionLocale and
// userLocale are pre-extracted by the caller (session cookie value, and
// UserLocale() above) since this package doesn't know about sessions or the
// user database.
func (m *Manager) ResolveLocale(ctx context.Context, sessionLocale, userLocale, acceptLanguageHeader string) string {
	if sessionLocale != "" {
		return sessionLocale
	}
	if userLocale != "" {
		return userLocale
	}
	if best := bestMatch(acceptLanguageHeader, m.AvailableLocales(ctx)); best != "" {
		return best
	}
	return m.SystemDefaultLocale(ctx)
}

func (m *Manager) locale(lang string) *gotext.Locale {
	m.mu.RLock()
	loc, ok := m.locales[lang]
	m.mu.RUnlock()
	if ok {
		return loc
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if loc, ok := m.locales[lang]; ok { // re-check after acquiring write lock
		return loc
	}
	loc = gotext.NewLocale(m.dir, lang)
	loc.AddDomain(domain)
	m.locales[lang] = loc
	return loc
}

// Get translates str for lang. kv is an alternating key/value list
// substituted into %(key)s placeholders in the translated string, e.g.
// Get("en", "Hello %(name)s", "name", "Stefan"). A missing translation
// falls back to str itself, with substitution still applied.
func (m *Manager) Get(lang, str string, kv ...string) string {
	// Bound method value, not a direct m.locale(lang).Get(str) call: go
	// vet's printf checker sees gotext's Get(str string, vars ...any)
	// internally Sprintf-ing and flags str as a "non-constant format
	// string" - str is a translation key, not a format string, and no
	// vars are ever passed here, so there's nothing to format. Routing
	// the call through a variable is enough to stop vet tracing it as a
	// printf wrapper.
	get := m.locale(lang).Get
	translated := get(str)
	if len(kv) == 0 {
		return translated
	}
	return substitute(translated, kv)
}

func substitute(s string, kv []string) string {
	result := s
	for i := 0; i+1 < len(kv); i += 2 {
		result = strings.ReplaceAll(result, "%("+kv[i]+")s", kv[i+1])
	}
	return result
}

// Translator is bound to one request's resolved locale and passed as
// template data (e.g. `{{.T.Get "text"}}`) rather than registered in the
// shared *template.Template's FuncMap, since Funcs() mutates that shared,
// cached template and isn't safe to call per-request under concurrent
// handlers.
type Translator struct {
	manager *Manager
	locale  string
}

func (m *Manager) Translator(locale string) Translator {
	return Translator{manager: m, locale: locale}
}

func (t Translator) Get(str string, kv ...string) string {
	return t.manager.Get(t.locale, str, kv...)
}

func (t Translator) GetN(str, plural string, n int) string {
	return t.manager.locale(t.locale).GetN(str, plural, n)
}

func (t Translator) Locale() string {
	return t.locale
}

type acceptEntry struct {
	tag string
	q   float64
}

func parseAcceptLanguage(header string) []acceptEntry {
	var entries []acceptEntry

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tag := part
		q := 1.0

		if i := strings.Index(part, ";"); i >= 0 {
			tag = strings.TrimSpace(part[:i])
			if qPart := strings.TrimSpace(part[i+1:]); strings.HasPrefix(qPart, "q=") {
				if parsed, err := strconv.ParseFloat(qPart[2:], 64); err == nil {
					q = parsed
				}
			}
		}

		if tag == "" || tag == "*" {
			continue
		}
		entries = append(entries, acceptEntry{tag: strings.ToLower(tag), q: q})
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].q > entries[j].q })
	return entries
}

// bestMatch picks the best Accept-Language match against our catalog,
// which only ever contains primary-subtag locale codes (en, de, sr, ...):
// try an exact tag match first, then a primary-subtag match (so "en-US"
// matches available "en"), in header preference order.
func bestMatch(header string, available []string) string {
	entries := parseAcceptLanguage(header)
	if len(entries) == 0 {
		return ""
	}

	avail := make(map[string]string, len(available))
	for _, l := range available {
		avail[strings.ToLower(l)] = l
	}

	for _, e := range entries {
		if orig, ok := avail[e.tag]; ok {
			return orig
		}
	}
	for _, e := range entries {
		primary := e.tag
		if i := strings.IndexAny(primary, "-_"); i >= 0 {
			primary = primary[:i]
		}
		if orig, ok := avail[primary]; ok {
			return orig
		}
	}

	return ""
}
