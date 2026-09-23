package dashboard

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestParseLimit(t *testing.T) {
	cases := []struct {
		raw     string
		size    bool
		display string
		mb      float64
	}{
		{"0", false, "∞", 0},
		{"", false, "∞", 0},
		{"0 GB", true, "∞", 0},
		{"25", false, "25", 25},
		{"5 GB", true, "5 GB", 5 * 1024},
		{"2g", true, "2 GB", 2 * 1024},
		{"10G", true, "10 GB", 10 * 1024},
		{"512 MB", true, "512 MB", 512},
	}
	for _, c := range cases {
		got := parseLimit(c.raw, c.size)
		if got.display != c.display || got.value != c.mb || !got.ok {
			t.Errorf("parseLimit(%q) = %+v, want display %q value %v", c.raw, got, c.display, c.mb)
		}
	}
	if got := parseLimit("lots", false); got.ok || got.display != "lots" {
		t.Errorf("expected an unparseable limit shown as-is and never compared, got %+v", got)
	}
}

func TestCompareLimits(t *testing.T) {
	// mirrors the test server's Standard plan -> Developer plus
	current := appctx.PlanDetails{WebsitesLimit: "10", DomainsLimit: "0", DiskLimit: "5 GB", RAM: "2g", MaxEmailQuota: "10G", InodesLimit: "1000000"}
	upsell := appctx.PlanDetails{WebsitesLimit: "25", DomainsLimit: "25", DiskLimit: "50 GB", RAM: "6g", MaxEmailQuota: "2G", InodesLimit: "1000000"}

	all := func(...string) bool { return true }
	changes := map[string]string{}
	for _, row := range compareLimits(current, upsell, all) {
		changes[row.Label] = row.Change
	}
	want := map[string]string{
		"Websites Limit": "up",
		"Domains Limit":  "down", // unlimited -> 25 is shown honestly as a decrease
		"Disk Limit":     "up",
		"Memory Limit":   "up",
		"Storage quota":  "down",
		"Inodes Limit":   "",
		"Database Limit": "", // both unlimited
	}
	for label, change := range want {
		if changes[label] != change {
			t.Errorf("%s: change %q, want %q", label, changes[label], change)
		}
	}
}

func TestRenderUpgradePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "mysql": true}
	upsell := map[string]bool{"postgresql": true, "phpmyadmin": true}

	data := UpgradePageData{
		LayoutData: web.LayoutData{
			Title:          "Upgrade",
			CSRFToken:      "tok",
			PanelDir:       "ltr",
			NavGroups:      web.BuildClassicSidebarNav(allowed, upsell, "/dashboard/upgrade"),
			PageTabs:       web.DashboardTabs("/dashboard/upgrade", true),
			UserAllowed:    allowed,
			UpsellAllowed:  upsell,
			UpsellPlanName: "Business",
			UpsellURL:      "https://example.com/upgrade",
			RequestPath:    "/dashboard/upgrade",
			T:              mgr.Translator("en"),
		},
		CurrentPlanName: "Starter",
		Limits:          compareLimits(appctx.PlanDetails{WebsitesLimit: "10"}, appctx.PlanDetails{WebsitesLimit: "25"}, func(...string) bool { return true }),
		Features:        web.UpgradeFeatures(allowed, upsell),
	}

	w := httptest.NewRecorder()
	if err := upgradePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Upgrade to Business",
		"Starter",
		"PostgreSQL",
		"phpMyAdmin",
		`href="https://example.com/upgrade" target="_blank" rel="noopener"`,
		`href="/dashboard/upgrade"`,
		"↑ ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered upgrade page missing %q", want)
		}
	}
}

func TestCompareLimitsSkipsInactiveFeatures(t *testing.T) {
	has := featureCheck(map[string]bool{"mysql": true}, map[string]bool{"ftp": true})
	var labels []string
	for _, row := range compareLimits(appctx.PlanDetails{}, appctx.PlanDetails{}, has) {
		labels = append(labels, row.Label)
	}
	if got := strings.Join(labels, ","); got != "Database Limit,FTP Limit,Disk Limit,Inodes Limit,CPU Limit,Memory Limit,Bandwidth Limit" {
		t.Errorf("unexpected limit rows: %s", got)
	}
}
