package websites

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/domains"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func pngOf(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBlankScreenshot(t *testing.T) {
	white := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	for i := range white.Pix {
		white.Pix[i] = 0xff
	}
	if !blankScreenshot(pngOf(t, white)) {
		t.Error("an all-white capture should count as blank")
	}

	// a stray dot doesn't make a page
	white.Set(10, 10, color.Black)
	if !blankScreenshot(pngOf(t, white)) {
		t.Error("one dark pixel should still count as blank")
	}

	page := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	for y := 0; y < 720; y++ {
		for x := 0; x < 1280; x++ {
			c := color.RGBA{0xff, 0xff, 0xff, 0xff}
			if y < 80 || (y > 200 && y < 230 && x > 100 && x < 900) {
				c = color.RGBA{0x21, 0x75, 0x9b, 0xff}
			}
			page.Set(x, y, c)
		}
	}
	if blankScreenshot(pngOf(t, page)) {
		t.Error("a page with a header and text isn't blank")
	}
	if blankScreenshot([]byte("not an image")) {
		t.Error("undecodable data is left to the browser, not called blank")
	}
}

// the Options tab saves from the POST body, the API's query string form still works
func TestSiteInfoOptionsFromBody(t *testing.T) {
	body := url.Values{"website": {"example.com"}, "blogname": {"My site"}, "admin_email": {"a@example.com"}, "blog_public": {"0"}}
	post := httptest.NewRequest(http.MethodPost, "/wordpress/wp-cli/update_site_information", strings.NewReader(body.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	got := map[string]string{}
	for _, kv := range siteInfoOptionsFrom(post) {
		got[kv[0]] = kv[1]
	}
	if len(got) != 3 || got["blogname"] != "My site" || got["admin_email"] != "a@example.com" || got["blog_public"] != "0" {
		t.Errorf("options from body = %v", got)
	}

	get := httptest.NewRequest(http.MethodPost, "/api/sites/wp-cli/update_site_information?blogdescription=Hi&website=example.com", nil)
	opts := siteInfoOptionsFrom(get)
	if len(opts) != 1 || opts[0] != [2]string{"blogdescription", "Hi"} {
		t.Errorf("options from query = %v", opts)
	}
}

func TestSSLState(t *testing.T) {
	cases := []struct {
		https string
		cert  domains.SSLCertInfo
		want  string
	}{
		{"Unknown", domains.SSLCertInfo{}, ""},
		{"Automatic", domains.SSLCertInfo{Expires: "2027-01-01"}, "valid"},
		{"Custom", domains.SSLCertInfo{Problem: "expiring"}, "expiring"},
		{"Automatic", domains.SSLCertInfo{Problem: "expired"}, "expired"},
		{"Automatic", domains.SSLCertInfo{Problem: "missing"}, "missing"},
		{"Automatic", domains.SSLCertInfo{Problem: "unreadable"}, "missing"},
	}
	for _, c := range cases {
		if got := sslState(c.https, c.cert); got != c.want {
			t.Errorf("sslState(%q, %+v) = %q, want %q", c.https, c.cert, got, c.want)
		}
	}
}

func renderWP(t *testing.T, data WPSinglePageData, allow ...string) string {
	t.Helper()
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/website")
	for _, k := range allow {
		layout.UserAllowed[k] = true
	}
	layout.NavTrail = web.BuildNavTrail(layout.UserAllowed, nil, "/website", "/website", "example.com", layout.PageTabs)
	layout.ActivateTab("/wordpress")
	data.pageData = pageData{LayoutData: layout, CurrentDomain: "example.com", Docroot: "/var/www/html/example.com"}
	if data.Container.SiteName == "" {
		data.Container = ContainerInfo{SiteName: "example.com", Version: "7.1.4", CreatedDate: "2026-03-14"}
	}
	w := httptest.NewRecorder()
	if err := wpSinglePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return w.Body.String()
}

func TestWPManagerHeader(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{CurrentPHPVersion: "8.3", SSLState: "valid", SSLExpires: "2027-01-01", RedisOn: true})
	for _, want := range []string{`<span class="op-status is-success">Online</span>`, "PHP 8.3", "SSL valid", `title="Expires 2027-01-01"`, "Redis object cache", `id="wp-hero-updates"`} {
		if !strings.Contains(body, want) {
			t.Errorf("header: missing %q", want)
		}
	}
	// the login menu is a kit menu now, not the old Tailwind dropdown
	if !strings.Contains(body, `class="op-menu op-wp-users"`) || strings.Contains(body, "wpUserDropdownOpen") {
		t.Error("login as admin menu should be an op-menu")
	}

	off := renderWP(t, WPSinglePageData{Suspended: true, SSLState: "missing"})
	for _, want := range []string{`<span class="op-status is-danger">Suspended</span>`, "No SSL certificate yet"} {
		if !strings.Contains(off, want) {
			t.Errorf("suspended header: missing %q", want)
		}
	}
	for _, bad := range []string{"Redis object cache</span>", ">Online<", "SSL valid"} {
		if strings.Contains(off, bad) {
			t.Errorf("suspended header shouldn't show %q", bad)
		}
	}
	if strings.Contains(renderWP(t, WPSinglePageData{}), ">SSL ") {
		t.Error("no SSL badge without a certificate set up")
	}
}

// a WordPress site under /website highlights the WordPress tab and crumb, not Sites
func TestWPManagerNavTab(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{})
	if !strings.Contains(body, `<a href="/wordpress"  aria-current="page">WordPress</a>`) {
		t.Error("the WordPress tab should be the current one")
	}
	if strings.Contains(body, `<a href="/sites"  aria-current="page">`) {
		t.Error("Sites shouldn't be highlighted on a WordPress site")
	}
}

func TestWPManagerOverview(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{CurrentPHPVersion: "8.3", AvailablePHPVersions: []string{"8.3", "8.2"}}, "php", "containers")
	for _, want := range []string{
		`id="wp-speed-score"`, `id="wp-plugin-count"`, `id="wp-theme-count"`, `id="wp-disk-total"`,
		"Site details", `id="wp-admin-user"`, `id="wp-auto-update-mode"`, `id="created_date" data-relative-date="2026-03-14"`,
		`id="php-version"`, `id="wp-version"`, `id="wp-version-release-nag"`, `id="mysql-version"`, `id="db-type-text"`,
		`<select id="wp-new-php" x-model="selected" class="op-select"`, "'/wordpress/tools/overview'", "wpParseSize(data.size)",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("overview: missing %q", want)
		}
	}
	if strings.Contains(body, `id="partials_pagespeed"`) {
		t.Error("the Speed card became a tile, the old panel shouldn't render")
	}
}

func TestWPManagerClone(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{
		Domains:    []appctx.Domain{{DomainURL: "example.com", Docroot: "/var/www/html/example.com"}, {DomainURL: "blog.example.com", Docroot: "/var/www/html/blog.example.com"}, {DomainURL: "new.example.com", Docroot: "/var/www/html/new.example.com"}},
		TakenSites: map[string]string{"example.com": "wordpress", "blog.example.com": "joomla"},
	})
	i := strings.Index(body, `id="target_domain"`)
	j := strings.Index(body[i:], "</select>")
	sel := body[i : i+j]
	if strings.Contains(sel, `value="example.com"`) {
		t.Error("clone targets shouldn't offer the site itself")
	}
	if !strings.Contains(sel, `value="blog.example.com" data-taken="joomla"`) || !strings.Contains(sel, "has a site") {
		t.Error("a domain that hosts a site should be flagged")
	}
	if !strings.Contains(sel, `value="new.example.com">new.example.com`) {
		t.Error("a free domain should be offered as is")
	}
	for _, want := range []string{`var taken = {"blog.example.com":"joomla","example.com":"wordpress"}`, `:disabled="!target || !!takenBy()"`, "estimate()", "Pick the domain the copy goes to."} {
		if !strings.Contains(body, want) {
			t.Errorf("clone: missing %q", want)
		}
	}
	if strings.Contains(body, "Please choose the URL to install the software") {
		t.Error("clone still has the installer's help text")
	}
}

func TestWPManagerTabs(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{}, "error_logs")
	for _, want := range []string{
		// Options save in the POST body
		`fetch("/wordpress/wp-cli/update_site_information", {`, `id="cancelGeneralBtn"`, "Discourage search engines",
		// Debugging loads when the tab shows, however it was opened, and saves each switch
		`x-data="wpDebug()" x-effect="if (activeTab === `, `:disabled="!loaded || saving" @change="save()"`, "Debug settings",
		// Security rows read as the protected state
		"File editor in wp-admin is off</b>", "Debug mode is off</b>", "wp-config.php is private</b>", `aria-label="Tools"`,
		// error toasts show the error
		"showToast(`${response.error}`, 'error');",
		// one way to start a backup
		"backupNow(true, false)", `id="db-size-for-backup"`, `id="files-size-for-backup"`,
		// Updates
		`<span class="op-hint" x-show="!it.has_update">Up to date</span>`, `class="op-wp-foot-lines"`,
		// tab loaders run once for ?tab=, base.html's replayed click is skipped
		"if (skipReplay && !e.isTrusted)",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tabs: missing %q", want)
		}
	}
	for _, bad := range []string{
		"update_site_information?website=", "showToast(`${response.message}`, 'error');", `id="runBackup"`, "Custom backup",
		"badgeText = 'Latest'", "text-blue-500", "XML-RPC is reachable", `id="saveSettingsBtn"`,
	} {
		if strings.Contains(body, bad) {
			t.Errorf("tabs: still has %q", bad)
		}
	}
	if strings.Count(body, `class="op-check-row op-wp-plain op-wp-tool"`) != 3 {
		t.Error("salts, reinstall and integrity should be rows of one Tools card")
	}
}

// Debugging: settings card, then the debug.log tail, then the site's error logs
func TestWPManagerDebugging(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{}, "error_logs")
	start, end := strings.Index(body, `id="debugging" role="tabpanel"`), strings.Index(body, `id="clone" role="tabpanel"`)
	if start < 0 || end < start {
		t.Fatal("debugging panel not found")
	}
	tab := body[start:end]
	settings, logCard, errCard := strings.Index(tab, `id="updateDebugForm"`), strings.Index(tab, `x-data="wpDebugLog()"`), strings.Index(tab, "<h2>Error Logs</h2>")
	if settings < 0 || logCard < 0 || errCard < 0 || settings > logCard || logCard > errCard {
		t.Fatalf("cards out of order: settings %d, debug.log %d, error logs %d", settings, logCard, errCard)
	}
	for _, want := range []string{
		`<h2>debug.log</h2>`, `x-effect="if (activeTab === 'debugging' && !loaded && !loading) load()"`,
		"wpToolGet('/wordpress/tools/debug-log')", "wpToolPost('/wordpress/tools/debug-log-clear')",
		`id="SAVEQUERIES"`, `class="op-editor op-wp-debuglog-box"`, `<div class="op-gutter" aria-hidden="true">`,
		"No debug.log yet. Turn on WP_DEBUG and WP_DEBUG_LOG to start one.", "debug.log is empty.", `x-show="error"`,
		"Click again to confirm", "Clear Log", "Refresh",
		`class="op-wp-doclink" href="https://wordpress.org/support/article/debugging-in-wordpress/"`,
		// error logs card: rows, refresh, View Logs and the "Showing N out of M" footer
		"Latest PHP, web server and HTTP 5xx errors for this website.", `class="op-wp-errlog"`, "exact: 1, limit: 10, format: 'json'",
		`href="/error-logs?domain=example.com&exact=1">`, "View Logs", `class="op-hint op-wp-errlog-more"`, "No errors found for these filters.",
	} {
		if !strings.Contains(tab, want) {
			t.Errorf("debugging: missing %q", want)
		}
	}
	for _, bad := range []string{"All error logs for this site", "Each switch changes wp-config.php right away."} {
		if strings.Contains(tab, bad) {
			t.Errorf("debugging: still has %q", bad)
		}
	}
	if strings.Count(tab, `class="op-check-row op-switch-row"`) != 5 {
		t.Error("five debug switches expected")
	}
	if strings.Count(tab, `x-effect="if (activeTab === 'debugging' && !loaded && !loading) load()"`) != 3 {
		t.Error("all three cards should load when the tab opens")
	}

	// no error logs module, no error logs card
	if off := renderWP(t, WPSinglePageData{}); strings.Contains(off, "<h2>Error Logs</h2>") || !strings.Contains(off, `x-data="wpDebugLog()"`) {
		t.Error("the error logs card needs the error_logs feature, debug.log doesn't")
	}
}

// the shared screenshot partial uses kit icons in every manager
func TestScreenshotPartialIcons(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderWP(t, WPSinglePageData{})
	for _, bad := range []string{"bi-arrow-counterclockwise", "bi-box-arrow-up-right"} {
		if strings.Contains(body, bad) {
			t.Errorf("screenshot still uses %s", bad)
		}
	}
	for _, want := range []string{"op-shot-open", "op-shot-refresh", "op-shot-retry"} {
		if !strings.Contains(body, want) {
			t.Errorf("screenshot: missing %s", want)
		}
	}
}
