// Package i18n reuses the panel's existing gettext .po/.mo catalogs and resolves the locale via session -> per-account file -> Accept-Language -> system default -> "en".
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

// AvailableLocales returns every subdirectory of dir with LC_MESSAGES/messages.po, plus "en", cached for 1h - `opencli locale` busts its own cache after installing a locale but not this one, so a new locale won't show up here until the TTL expires
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

// UserLocale reads /home/<systemUsername>/locale, the per-account override; returns "" if unset so the caller keeps checking the rest of the chain
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

// ResolveLocale runs the locale priority chain; sessionLocale/userLocale are pre-extracted by the caller since this package doesn't know about sessions or the user DB
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

// Get translates str for lang, substituting kv's alternating key/value pairs into %(key)s placeholders, e.g. Get("en", "Hello %(name)s", "name", "Stefan"). Falls back to str itself if untranslated.
func (m *Manager) Get(lang, str string, kv ...string) string {
	// bound as a variable, not called directly - otherwise go vet's printf checker flags str as a non-constant format string, even though it's just a translation key
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

// Translator is bound to one request's locale and passed as template data (`{{.T.Get "text"}}`) instead of living in the shared template's FuncMap, since Funcs() isn't safe to call per-request
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

// bestMatch picks the best Accept-Language match against our catalog (exact tag first, then primary subtag so "en-US" matches "en"), in header preference order
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
