package web

import (
	"encoding/json"
	"html/template"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// FlashDisplay is a pre-processed flash message ready for the flash-message stack: reversed order (newest on top), a 1-based Index, ZClass already resolved from the z-index cycle, and any leading "<br>" stripped from the text.
type FlashDisplay struct {
	Index  int
	ZClass string
	Text   template.HTML
}

var flashZClasses = []string{"z-10", "z-20", "z-30", "z-40", "z-50"}

// BuildFlashDisplay reverses messages (newest first), cycles z-index classes across them, and strips a leading "<br>" from each message's text
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

// LayoutData is everything the shared page layout and its partials need - nav, branding, flashes, translator, and the rest of the per-request context common to every authenticated page.
type LayoutData struct {
	Title         string
	BrandName     string
	Favicon       string
	Logo          string
	CSRFToken     string
	PanelDir      string
	FoundABugLink string
	PanelVersion  string
	CustomCSS     bool
	CustomJS      bool

	MenuStyle       string     // "classic" or "modern", see ReadMenuStyle
	NavItems        []NavItem  // modern sidebar
	NavGroups       []NavGroup // classic sidebar
	PageTabs        []NavLink  // tab bar rendered above the page content, nil when the page's area has no siblings
	NavTrail        []NavLink  // header trail, the last one is the current page and has no Href
	UserAllowed     map[string]bool
	UserAllowedJSON template.JS
	IsEnterprise    bool

	// UpsellAllowed lists modules the current plan doesn't grant but the configured upsell plan does - the sidebar/dashboard show these greyed-out with an upgrade prompt instead of hiding them. UpsellPlanName/UpsellURL are blank when no upsell plan is configured.
	UpsellAllowed  map[string]bool
	UpsellPlanName string
	UpsellURL      string

	CurrentUsername string
	HostingPlanName string
	AvatarType      string
	GravatarURL     string

	RequestPath   string
	Flashes       []FlashDisplay
	Impersonating bool
	AdminPort     string

	PasswordStrength int // clamped password_strength config value every page's passwordStrength() Alpine component reads as its minimum-score threshold

	Service string // set by pages that manage a single service; "" on the dashboard, so {{if .Service}} partials render nothing there

	// DashboardIconView is the user's saved dashboard icon-section layout ("icon" or "list"), read server-side from the dashboard_icon_view cookie on every page (not just /dashboard) so the sidebar toggle and the dashboard's icon_section template both render the saved choice on first paint with no Alpine/localStorage hydration flicker
	DashboardIconView string

	// FilemanagerView is the resolved filemanager button style ("classic" or "modern"), read server-side on every page (not just /files) so the account menu's style toggle (only shown on /files) highlights the active choice with no flicker
	FilemanagerView string

	T i18n.Translator
}

// UserAllowedList renders m's keys as a JSON array, for embedding the user's permission flags into the page as a client-side JS value.
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
