package dashboard

import (
	"context"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/domains"
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

// DashboardPageData is everything dashboard.html and its includes need, combining the shared app-shell data (web.LayoutData) with dashboard()'s own render_template kwargs
type DashboardPageData struct {
	web.LayoutData

	Sections       []Section
	TourShow       bool
	OnboardingShow bool
	OnboardingSkip OnboardingSkip

	BackupsAdminManaged bool

	CustomMessage template.HTML

	CustomSectionTitle    string
	CustomSectionItems    []CustomSectionItem
	CustomSectionPosition string

	TwofaEnabled       bool
	TwofaNag           string
	TwofaStatusMessage template.HTML

	IPAddress    string
	LastIP       string
	IPCountyFlag string
	// LastLoginAgo is LastLoginTime as "2 hours ago"
	LastLoginAgo, LastLoginTime string
	NS1, NS2, NS3, NS4          string

	HowToGuides       string
	HowToTopics       []HowToTopic
	KnowledgeBaseLink string

	UserWebsitesCount int
	MainDomainsCount  int
	// Websites is the newest few sites for the Overview's Websites card
	Websites   []DashboardSite
	DBUsage    int
	EmailCount int
	FTPCount   int

	WebsitesLimit int
	DomainsLimit  int
	DBLimit       int
	EmailLimit    int
	FTPLimit      int

	// LimitReached is set when any resource above is at/over its plan limit, so the dashboard can surface an upgrade prompt. UpgradePlanName/UpgradeURL are only non-empty when the current plan has an upsell plan configured in openadmin.
	LimitReached    bool
	UpgradePlanName string
	UpgradeURL      string
}

// twofaStatusMessage builds the "2FA is <b>{status}</b> for your account." message, translating status through the same "enabled"/"disabled" catalog keys twofa_settings.html already uses
func twofaStatusMessage(t i18n.Translator, enabled bool) template.HTML {
	status := t.Get("disabled")
	if enabled {
		status = t.Get("enabled")
	}
	msg := t.Get("2FA is <b>{status}</b> for your account.")
	return template.HTML(strings.Replace(msg, "{status}", status, 1)) //nolint:gosec // msg is a translation-catalog string plus a translated status word, not user input
}

// DashboardSite is one row of the Overview's Websites card
type DashboardSite struct {
	Name, Type, PHP, Href string
	Suspended             bool
	// Service and Port are set for apps in their own container, State is its podman state when known
	Service, Port, State string
}

// Tile is the op_icon name and hue of the site's type tile
func (s DashboardSite) Tile() (icon, hue string) {
	switch s.Type {
	case "WordPress":
		return "wordpress", "green"
	case "Node.js":
		return "terminal", "green"
	case "Python":
		return "terminal", "yellow"
	case "Ruby":
		return "terminal", "red"
	case "Java":
		return "terminal", "orange"
	case "PHP":
		return "php", "purple"
	case "n8n", "Docker":
		return "box", "cyan"
	case "Website Builder":
		return "layout", "indigo"
	}
	return "apps", "blue"
}

// TileIcon and TileHue split Tile for templates
func (s DashboardSite) TileIcon() string { i, _ := s.Tile(); return i }
func (s DashboardSite) TileHue() string  { _, h := s.Tile(); return h }

// RunsOn is the PHP version, or the runtime for apps in their own container
func (s DashboardSite) RunsOn() string {
	if s.PHP != "" {
		return "PHP " + s.PHP
	}
	if s.Service != "" {
		if s.Port != "" {
			return s.Service + ":" + s.Port
		}
		return s.Service
	}
	return s.Type
}

// dashboardSites keeps the newest six sites and marks the ones whose domain is suspended
func dashboardSites(ctx context.Context, a *appctx.App, userContext string, sites []UserWebsite) []DashboardSite {
	if len(sites) > 6 {
		sites = sites[:6]
	}
	rows := make([]DashboardSite, 0, len(sites))
	var states map[string]string
	for _, s := range sites {
		php := s.PHPVersion
		row := DashboardSite{
			Name: s.SiteName, Type: siteTypeLabel(s.Type),
			Href:      "/website?domain=" + url.QueryEscape(s.SiteName),
			Suspended: s.DomainURL != "" && domains.IsSuspended(ctx, a, s.DomainURL),
		}
		// apps in their own runtime container don't use the domain's PHP version
		switch t := strings.ToLower(s.Type); {
		case strings.Contains(t, "node"), strings.Contains(t, "python"), strings.Contains(t, "ruby"), strings.Contains(t, "java"), strings.Contains(t, "n8n"), strings.Contains(t, "docker"):
			php = ""
			if s.Container != "" {
				row.Service, row.Port = strings.ToLower(s.Container), s.Ports
				if states == nil {
					states = containerStates(ctx, a, userContext)
				}
				row.State = states[row.Service]
			}
		}
		row.PHP = php
		rows = append(rows, row)
	}
	return rows
}

// containerStates is one podman ps for all of the account's containers, name to state, cached 30s
func containerStates(ctx context.Context, a *appctx.App, userContext string) map[string]string {
	if userContext == "" || a.Cache == nil {
		return map[string]string{}
	}
	states, _ := cache.Memoize(ctx, a.Cache, "dashboard_container_states:"+userContext, 30*time.Second, func() (map[string]string, error) {
		pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		argv := podmanmanager.PodmanArgv(userContext, "ps", "-a", "--format", "{{.Names}} {{.State}}")
		out, err := podmanmanager.Command(pctx, userContext, argv).Output()
		if err != nil {
			return map[string]string{}, nil //nolint:nilerr // no state is shown as plain Active
		}
		return parseContainerStates(string(out)), nil
	})
	if states == nil {
		states = map[string]string{}
	}
	return states
}

// parseContainerStates reads "name state" lines from podman ps
func parseContainerStates(out string) map[string]string {
	states := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) >= 2 {
			states[f[0]] = strings.ToLower(f[1])
		}
	}
	return states
}

// StateLabel is Running or Stopped for a container app, empty when the state is unknown
func (s DashboardSite) StateLabel() string {
	switch s.State {
	case "":
		return ""
	case "running":
		return "Running"
	}
	return "Stopped"
}

// timeAgo turns a "2006-01-02 15:04:05" login time into "2 hours ago" using the base.html relative-date strings
func timeAgo(t i18n.Translator, stamp string, now time.Time) string {
	then, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(stamp), time.Local)
	if err != nil {
		return ""
	}
	d := max(now.Sub(then), 0)
	unit := func(n int, one, many string) string {
		if n == 1 {
			return t.Get(one)
		}
		return strings.Replace(t.Get(many), "%%(n)s", strconv.Itoa(n), 1)
	}
	switch {
	case d < time.Minute:
		return unit(max(int(d.Seconds()), 1), "1 second ago", "%%(n)s seconds ago")
	case d < time.Hour:
		return unit(int(d.Minutes()), "1 minute ago", "%%(n)s minutes ago")
	case d < 24*time.Hour:
		return unit(int(d.Hours()), "1 hour ago", "%%(n)s hours ago")
	case d < 7*24*time.Hour:
		return unit(int(d.Hours()/24), "1 day ago", "%%(n)s days ago")
	case d < 30*24*time.Hour:
		return unit(int(d.Hours()/24/7), "1 week ago", "%%(n)s weeks ago")
	case d < 365*24*time.Hour:
		return unit(int(d.Hours()/24/30), "1 month ago", "%%(n)s months ago")
	}
	return unit(int(d.Hours()/24/365), "1 year ago", "%%(n)s years ago")
}

// siteTypeLabel turns the stored site type into the product name people know
func siteTypeLabel(t string) string {
	names := map[string]string{"wordpress": "WordPress", "nodejs": "Node.js", "python": "Python", "ruby": "Ruby", "java": "Java", "n8n": "n8n",
		"php": "PHP", "drupal": "Drupal", "joomla": "Joomla", "websitebuilder": "Website Builder", "website builder": "Website Builder", "docker": "Docker"}
	if n, ok := names[strings.ToLower(strings.TrimSpace(t))]; ok {
		return n
	}
	if t == "" {
		return t
	}
	return strings.ToUpper(t[:1]) + t[1:]
}

// PlanTile is one usage tile on the Overview view
type PlanTile struct {
	Label, Href, Icon string
	Used, Limit, Pct  int
}

// PlanTiles feeds the Overview tiles, a limit of 0 means unlimited
func (d DashboardPageData) PlanTiles() []PlanTile {
	allowed := d.UserAllowed
	tile := func(label, href, icon string, used, limit int) PlanTile {
		t := PlanTile{Label: d.T.Get(label), Href: href, Icon: icon, Used: used, Limit: limit}
		if limit > 0 {
			t.Pct = min(100, used*100/limit)
		}
		return t
	}
	var tiles []PlanTile
	if allowed["wordpress"] || allowed["website_builder"] || allowed["nodejs"] || allowed["python"] {
		tiles = append(tiles, tile("Websites", "/sites", "browser", d.UserWebsitesCount, d.WebsitesLimit))
	}
	if allowed["domains"] {
		tiles = append(tiles, tile("Domains", "/domains", "globe", d.MainDomainsCount, d.DomainsLimit))
	}
	if allowed["mysql"] || allowed["postgresql"] {
		href := "/postgresql"
		if allowed["mysql"] {
			href = "/mysql"
		}
		tiles = append(tiles, tile("Databases", href, "db", d.DBUsage, d.DBLimit))
	}
	if allowed["emails"] {
		tiles = append(tiles, tile("Email accounts", "/emails", "mail", d.EmailCount, d.EmailLimit))
	}
	return tiles
}

// QuickActions is a fixed pick of common tools for the Overview chips, or the first usable item of each section when the plan has few of them
func (d DashboardPageData) QuickActions() []SectionItem {
	var out []SectionItem
	// the Modern theme's picks, when the plan has them
	for _, href := range []string{"/files", "/phpmyadmin", "/domains/edit-dns-zone", "/domains/ssl", "/cronjobs", "/backups"} {
	find:
		for _, s := range d.Sections {
			for _, it := range s.Items {
				if it.Href == href && !it.Disabled {
					out = append(out, it)
					break find
				}
			}
		}
	}
	if len(out) >= 3 {
		return out
	}
	out = nil
	for _, s := range d.Sections {
		for _, it := range s.Items {
			if !it.Disabled {
				out = append(out, it)
				break
			}
		}
		if len(out) == 8 {
			break
		}
	}
	return out
}
