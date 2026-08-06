package web

import (
	"encoding/json"
	"html/template"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// FlashDisplay is a pre-processed flash message ready for the flash-message
// stack: reversed order (newest on top), a 1-based Index, and ZClass
// already resolved from the z-index cycle, with any leading "<br>" stripped
// from the message text.
type FlashDisplay struct {
	Index  int
	ZClass string
	Text   template.HTML
}

var flashZClasses = []string{"z-10", "z-20", "z-30", "z-40", "z-50"}

// BuildFlashDisplay reverses messages (newest first), cycles z-index
// classes across them, and strips a leading "<br>" from each message's text.
func BuildFlashDisplay(messages []flash.Message) []FlashDisplay {
	result := make([]FlashDisplay, len(messages))
	for i := range messages {
		src := messages[len(messages)-1-i] // reverse
		zClass := "z-10"
		if i < len(flashZClasses) {
			zClass = flashZClasses[i]
		}
		cleaned := strings.Replace(src.Text, "<br>", "", 1)
		result[i] = FlashDisplay{Index: i + 1, ZClass: zClass, Text: template.HTML(cleaned)} //nolint:gosec // flash text is server-generated, not raw user input
	}
	return result
}

// LayoutData is everything the shared page layout and its partials need -
// nav, branding, flashes, translator, and the rest of the per-request
// context common to every authenticated page.
type LayoutData struct {
	Title         string
	BrandName     string
	Favicon       string
	Logo          string
	CSRFToken     string
	PanelDir      string
	FoundABugLink string
	PanelVersion  string
	CustomPlugins bool
	CustomCSS     bool
	CustomJS      bool

	NavGroups       []NavGroup
	UserAllowed     map[string]bool
	UserAllowedJSON template.JS
	IsEnterprise    bool

	CurrentUsername string
	HostingPlanName string
	AvatarType      string
	GravatarURL     string

	RequestPath   string
	Flashes       []FlashDisplay
	Impersonating bool
	AdminPort     string

	// PasswordStrength is the clamped password_strength config value every
	// page's passwordStrength() Alpine component reads as its minimum-score
	// threshold.
	PasswordStrength int

	// Service is set by pages that manage a single service; the dashboard
	// doesn't set it, so it's always "" there and any partials guarded on
	// {{if .Service}} render nothing.
	Service string

	T i18n.Translator
}

// UserAllowedList renders m's keys as a JSON array, for embedding the
// user's permission flags into the page as a client-side JS value.
func UserAllowedList(allowed map[string]bool) template.JS {
	keys := make([]string, 0, len(allowed))
	for k, ok := range allowed {
		if ok {
			keys = append(keys, k)
		}
	}
	b, _ := json.Marshal(keys)
	return template.JS(b) //nolint:gosec // keys are feature-flag names from our own config, not user input
}
