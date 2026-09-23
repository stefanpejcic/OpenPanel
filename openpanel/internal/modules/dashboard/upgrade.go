package dashboard

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var upgradePage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"dashboard/upgrade.html",
)

// LimitRow is one plan limit shown side by side, Change is "up", "down" or "" when the new plan doesn't change it
type LimitRow struct {
	Label   string
	Current string
	New     string
	Change  string
}

// UpgradePageData drives dashboard/upgrade.html
type UpgradePageData struct {
	web.LayoutData

	CurrentPlanName string
	Limits          []LimitRow
	Features        []web.FeatureGroup
}

func handleUpgrade(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, injected, err := web.BuildLayoutData(a, w, r, "Upgrade now")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// upsell is Enterprise-only and optional per plan, without one there's nothing to compare
	if layout.UpsellPlanName == "" || layout.UpsellURL == "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	ctx := r.Context()
	planID, _ := injected["hosting_plan"].(int)
	current, _ := a.QueryPlanDetailsByID(ctx, planID)
	upsellID, _ := strconv.Atoi(current.UpsellPlanID)
	upsell, _ := a.QueryPlanDetailsByID(ctx, upsellID)

	data := UpgradePageData{
		LayoutData:      layout,
		CurrentPlanName: layout.HostingPlanName,
		Limits:          compareLimits(current, upsell, featureCheck(layout.UserAllowed, layout.UpsellAllowed)),
		Features:        web.UpgradeFeatures(layout.UserAllowed, layout.UpsellAllowed),
	}
	if err := upgradePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DASHBOARD - upgrade template render error: %v", err)
	}
}

// featureCheck reports whether any of the keys is enabled in either plan
func featureCheck(allowed, upsellAllowed map[string]bool) func(keys ...string) bool {
	return func(keys ...string) bool {
		for _, k := range keys {
			if allowed[k] || upsellAllowed[k] {
				return true
			}
		}
		return false
	}
}

// compareLimits lists the limits side by side, skipping ones for features neither plan has (no email limit without email and so on), labels reuse the plan editor's already translated strings
func compareLimits(current, upsell appctx.PlanDetails, has func(keys ...string) bool) []LimitRow {
	rows := []struct {
		label, cur, new string
		size            bool
		keys            []string // nil means always shown
	}{
		{"Websites Limit", current.WebsitesLimit, upsell.WebsitesLimit, false, []string{"wordpress", "drupal", "joomla", "opencart", "nextcloud", "prestashop", "matomo", "moodle", "mediawiki", "website_builder", "nodejs", "python"}},
		{"Domains Limit", current.DomainsLimit, upsell.DomainsLimit, false, []string{"domains"}},
		{"Database Limit", current.DBLimit, upsell.DBLimit, false, []string{"mysql", "postgresql", "mongodb"}},
		{"Email Limit", current.EmailLimit, upsell.EmailLimit, false, []string{"emails"}},
		{"FTP Limit", current.FTPLimit, upsell.FTPLimit, false, []string{"ftp"}},
		{"Disk Limit", current.DiskLimit, upsell.DiskLimit, true, nil},
		{"Inodes Limit", current.InodesLimit, upsell.InodesLimit, false, nil},
		{"CPU Limit", current.CPU, upsell.CPU, false, nil},
		{"Memory Limit", current.RAM, upsell.RAM, true, nil},
		{"Bandwidth Limit", current.Bandwidth, upsell.Bandwidth, false, nil},
		{"Storage quota", current.MaxEmailQuota, upsell.MaxEmailQuota, true, []string{"emails"}},
	}
	out := make([]LimitRow, 0, len(rows))
	for _, row := range rows {
		if row.keys != nil && !has(row.keys...) {
			continue
		}
		c, n := parseLimit(row.cur, row.size), parseLimit(row.new, row.size)
		out = append(out, LimitRow{Label: row.label, Current: c.display, New: n.display, Change: limitChange(c, n)})
	}
	return out
}

type planLimit struct {
	value     float64 // in MB for sizes
	unlimited bool
	ok        bool // false when the value couldn't be parsed, so it's shown but never compared
	display   string
}

var limitRE = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Z]*)$`)

var sizeUnitsMB = map[string]float64{"": 1024, "k": 1.0 / 1024, "kb": 1.0 / 1024, "m": 1, "mb": 1, "g": 1024, "gb": 1024, "t": 1024 * 1024, "tb": 1024 * 1024}

// parseLimit reads a plans-table limit, where 0 or empty means unlimited (shown as ∞ like the dashboard usage widget) and sizes come as "5 GB", "2g" or "10G"
func parseLimit(raw string, size bool) planLimit {
	raw = strings.TrimSpace(raw)
	m := limitRE.FindStringSubmatch(raw)
	if raw == "" || (m != nil && m[1] == "0") {
		return planLimit{unlimited: true, ok: true, display: "∞"}
	}
	if m == nil {
		return planLimit{display: raw}
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	if !size {
		return planLimit{value: n, ok: m[2] == "", display: raw}
	}
	unit := strings.ToLower(m[2])
	mult, known := sizeUnitsMB[unit]
	if !known {
		return planLimit{display: raw}
	}
	// a bare size number is taken as GB, same as the plan editor's default unit
	shown := strings.ToUpper(strings.TrimSuffix(unit, "b"))
	if shown == "" {
		shown = "G"
	}
	return planLimit{value: n * mult, ok: true, display: m[1] + " " + shown + "B"}
}

func limitChange(c, n planLimit) string {
	switch {
	case !c.ok || !n.ok, c.unlimited && n.unlimited:
		return ""
	case n.unlimited:
		return "up"
	case c.unlimited:
		return "down"
	case n.value > c.value:
		return "up"
	case n.value < c.value:
		return "down"
	}
	return ""
}
