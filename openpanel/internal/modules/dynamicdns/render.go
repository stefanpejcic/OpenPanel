package dynamicdns

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var dynamicDNSPage = web.LoadPage("domains/_shared.html", "domains/dynamic_dns.html")

// DynHost is one row of the hosts table and one entry of the page's Alpine data
type DynHost struct {
	Domain  string `json:"domain"`
	Sub     string `json:"sub"`
	FQDN    string `json:"fqdn"`
	IP      string `json:"ip"`
	Token   string `json:"token"`
	Line    int    `json:"line"`
	Updated string `json:"updated"`
	// Waiting means the record still holds the 0.0.0.0 placeholder, no device has called yet
	Waiting bool `json:"waiting"`
}

func newDynHost(domain string, e DynDNSEntry) DynHost {
	return DynHost{
		Domain: domain, Sub: e.Subdomain, FQDN: e.Subdomain + "." + domain, IP: e.Record,
		Token: e.Token, Line: e.LineNumber, Updated: e.LastUpdated, Waiting: e.Record == "0.0.0.0",
	}
}

// TokenHint is the first four token characters the table shows before the dots
func (h DynHost) TokenHint() string {
	if len(h.Token) < 4 {
		return h.Token
	}
	return h.Token[:4]
}

// UpdatedUTC is the stored timestamp as "2026-10-26 13:28:42 UTC"
func (h DynHost) UpdatedUTC() string {
	if h.Updated == "" {
		return ""
	}
	return strings.Replace(strings.Replace(h.Updated, "T", " ", 1), "Z", " UTC", 1)
}

// DynamicDNSPageData is domains/dynamic_dns.html's template context.
type DynamicDNSPageData struct {
	web.LayoutData
	Hosts []DynHost
	// DomainNames are the domains a host can go on, .onion ones have no zone
	DomainNames []string
	FirstDomain string
	// Ago holds each host's "5 minutes ago" text, same order as Hosts
	Ago []string
	// DomainCount is how many domains have at least one host
	DomainCount int
	// Selected is the host the "Connect a device" card opens on
	Selected int
	// BaseURL is the panel's public scheme://host:port/ the update URLs hang off
	BaseURL string
}

func renderDynamicDNSPage(a *appctx.App, w http.ResponseWriter, r *http.Request, hosts []DynHost, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Dynamic DNS")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DynamicDNSPageData{LayoutData: layout, Hosts: hosts, BaseURL: web.PublicBaseURL(r.Context(), a, r)}
	for _, d := range domains {
		if !strings.HasSuffix(d.DomainURL, ".onion") {
			data.DomainNames = append(data.DomainNames, d.DomainURL)
		}
	}
	fillHostSummary(&data, r.URL.Query().Get("host"), time.Now())
	if err := dynamicDNSPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DYNAMICDNS - template render error: %v", err)
	}
}

// fillHostSummary works out the relative times, the domain count and which host the connect card starts on
func fillHostSummary(data *DynamicDNSPageData, selectHost string, now time.Time) {
	seen := map[string]bool{}
	data.Ago = make([]string, len(data.Hosts))
	for i, h := range data.Hosts {
		seen[h.Domain] = true
		data.Ago[i] = timeAgo(data.T, h.Updated, now)
		if selectHost != "" && strings.EqualFold(h.FQDN, selectHost) {
			data.Selected = i
		}
	}
	data.DomainCount = len(seen)
	if len(data.DomainNames) > 0 {
		data.FirstDomain = data.DomainNames[0]
	}
}

// timeAgo turns a stored "2006-01-02T15:04:05Z" stamp into "5 minutes ago" with the base.html relative-date strings
func timeAgo(t i18n.Translator, stamp string, now time.Time) string {
	then, err := time.Parse("2006-01-02T15:04:05Z", strings.TrimSpace(stamp))
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
		return t.Get("just now")
	case d < time.Hour:
		return unit(int(d.Minutes()), "1 minute ago", "%%(n)s minutes ago")
	case d < 24*time.Hour:
		return unit(int(d.Hours()), "1 hour ago", "%%(n)s hours ago")
	case d < 30*24*time.Hour:
		return unit(int(d.Hours()/24), "1 day ago", "%%(n)s days ago")
	case d < 365*24*time.Hour:
		return unit(int(d.Hours()/24/30), "1 month ago", "%%(n)s months ago")
	}
	return unit(int(d.Hours()/24/365), "1 year ago", "%%(n)s years ago")
}
