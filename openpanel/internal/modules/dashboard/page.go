package dashboard

import (
	"html/template"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// HowToTopic mirrors one entry of how_to_topics.
type HowToTopic struct {
	Link  string
	Title string
}

// CustomSectionItem mirrors one entry of custom_section_items.
type CustomSectionItem struct {
	URL   string
	Icon  string
	Label string
}

// DashboardPageData is everything dashboard.html and its includes need,
// combining the shared app-shell data (web.LayoutData) with dashboard()'s
// own render_template kwargs.
type DashboardPageData struct {
	web.LayoutData

	Sections []Section
	TourShow bool

	CustomMessage template.HTML

	CustomSectionTitle    string
	CustomSectionItems    []CustomSectionItem
	CustomSectionPosition string

	TwofaEnabled       bool
	TwofaNag           string
	TwofaStatusMessage template.HTML

	IPAddress          string
	LastIP             string
	IPCountyFlag       string
	NS1, NS2, NS3, NS4 string

	HowToGuides       string
	HowToTopics       []HowToTopic
	KnowledgeBaseLink string

	UserWebsitesCount int
	MainDomainsCount  int
	DBUsage           int
	EmailCount        int
	FTPCount          int

	WebsitesLimit int
	DomainsLimit  int
	DBLimit       int
	EmailLimit    int
	FTPLimit      int
}

// twofaStatusMessage mirrors twofa.html's
// `_('2FA is <b>{status}</b> for your account.').format(status=status_text)` -
// status_text itself ("enabled"/"disabled") is plain English in the
// Python source, not passed through _(), so it's never translated either.
func twofaStatusMessage(t i18n.Translator, enabled bool) template.HTML {
	status := "disabled"
	if enabled {
		status = "enabled"
	}
	msg := t.Get("2FA is <b>{status}</b> for your account.")
	return template.HTML(strings.Replace(msg, "{status}", status, 1)) //nolint:gosec // msg is a translation-catalog string plus a fixed English word, not user input
}
