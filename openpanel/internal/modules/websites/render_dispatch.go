package websites

import (
	"log"
	"net/http"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// pageData is embedded by every /website dispatcher page's data struct so
// the shared partials (temp_link, screenshot, visitors, waf_panel,
// pagespeed_panel) can rely on a consistent field set regardless of which
// CMS type is being rendered.
type pageData struct {
	web.LayoutData
	CurrentDomain        string
	Docroot              string
	PagespeedAPIKeyValue string
}

// PagespeedAPIKey satisfies the field name the pagespeed_panel partial
// reads, without colliding with the PagespeedAPIKeyValue field name used
// by every dispatch branch's literal struct above.
func (p pageData) PagespeedAPIKey() string { return p.PagespeedAPIKeyValue }

var phpAppPage = loadPage("manager/php_app.html")

// PHPAppPageData is manager/php_app.html's template context.
type PHPAppPageData struct {
	pageData
	Container                  ContainerInfo
	PHPVersion                 string
	InitialProject             string
	AutorunComposerInstall     bool
	ComposerOptimizeAutoloader bool
}

func renderPHPAppPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data PHPAppPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.CurrentDomain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	if err := phpAppPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - php_app template render error: %v", err)
	}
}

var websiteBuilderPage = loadPage("manager/websitebuilder.html")

// WebsiteBuilderPageData is manager/websitebuilder.html's template context.
type WebsiteBuilderPageData struct {
	pageData
	Container ContainerInfo
}

func renderWebsiteBuilderPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data WebsiteBuilderPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.CurrentDomain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	if err := websiteBuilderPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - websitebuilder template render error: %v", err)
	}
}

// SecurityToggle is one row of the Security tab's hardening-rule list.
// The rule IDs themselves come from `opencli websites-secure`, but their
// labels/tech details are only defined here.
type SecurityToggle struct {
	ID, Label, TechDetails string
}

var wpSecurityToggles = []SecurityToggle{
	{"wp_manager_wp_config", "Block access to wp-config.php",
		"Directly blocks any HTTP requests to the wp-config.php file, preventing exposure of database credentials even if PHP processing fails."},
	{"wp_manager_xmlrpc", "Block access to xmlrpc.php",
		"Returns a 403 error for all xmlrpc.php requests. Disables XML-based remote publishing and Pingbacks to prevent DDoS and brute-force amplification."},
	{"wp_manager_uploads_php", "Disable PHP in uploads",
		"Scans for any .php files within /wp-content/uploads/ and blocks them. Essential for preventing malicious shells from running after an unauthorized file upload."},
	{"wp_manager_wp_includes_php", "Restrict wp-includes PHP",
		"Blocks all PHP execution in the wp-includes folder, with a specific exclusion for wp-tinymce.php to ensure the editor keeps working."},
	{"wp_manager_admin_script_concat", "Disable Script Concatenation",
		"Blocks access to load-scripts.php and load-styles.php. This prevents a common ReDoS vulnerability used to spike server CPU via the admin dashboard."},
	{"wp_manager_author_scan", "Block Author Enumeration",
		"Identifies and blocks requests containing the \"author=\" query parameter, stopping bots from discovering valid administrative usernames."},
	{"wp_manager_sensitive_files", "Protect Sensitive Files",
		"Uses Regex to block access to backup files (.bak, .swp), readme/license files, and dangerous extensions like .log, .sh, or .exe."},
	{"wp_manager_cache_php", "Disable PHP in Cache",
		"Prevents the execution of PHP scripts stored within cache directories, stopping \"cache poisoning\" or \"file inclusion\" exploits."},
	{"wp_manager_env_files", "Protect Environment Files",
		"Blocks public access to .env, .htaccess, and .htpasswd files which often contain high-value secrets and server configurations."},
	{"wp_manager_bad_bots", "Block Malicious Bots",
		"Uses a User-Agent filter to block aggressive scrapers and security scanners like AhrefsBot, SemrushBot, MJ12bot, and Nikto."},
}

// WPSinglePageData is manager/wp/single.html's template context.
type WPSinglePageData struct {
	pageData
	Domains              []appctx.Domain
	Container            ContainerInfo
	BackupFilesAvailable bool
	IsSubdirectory       bool
	MainDomain           string
	CurrentPHPVersion    string
	AvailablePHPVersions []string
	SecurityToggles      []SecurityToggle
}

func renderWPSinglePage(a *appctx.App, w http.ResponseWriter, r *http.Request, data WPSinglePageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.CurrentDomain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	data.SecurityToggles = wpSecurityToggles
	sort.Sort(sort.Reverse(sort.StringSlice(data.AvailablePHPVersions)))
	if err := wpSinglePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - wp single template render error: %v", err)
	}
}

var wpSinglePage = loadPage("manager/wp/single.html", "manager/wp/_shared.html")

var pythonNodeAppsPage = loadPage("manager/python_node_apps.html")

// PythonNodeAppsPageData is manager/python_node_apps.html's template
// context.
type PythonNodeAppsPageData struct {
	pageData
	Container ContainerInfo
	// Service is container.container.split('_')[0] verbatim (no case
	// change) - the pm2/docker process id used throughout the page's JS.
	Service string
	PM2Data map[string]string
	// Type is Container.Type lowercased ("python" or "nodejs").
	Type string
	// PM2Status is pm2_data.status stringified ("true"/"false"/"unknown").
	PM2Status                                                                   string
	CPU, RAM, PIDs, StartupFile, CustomCmd, Workdir, CurrentVersion, GitRepoURL string
	RequirementsSelected                                                        bool
}

func renderPythonNodeAppsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data PythonNodeAppsPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.CurrentDomain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout

	prefix := data.PM2Data["prefix"]
	pm2val := func(key string) string {
		return strings.Trim(data.PM2Data[prefix+key], `"`)
	}

	data.Type = strings.ToLower(data.Container.Type)
	switch data.PM2Data["status"] {
	case "true":
		data.PM2Status = "true"
	case "false":
		data.PM2Status = "false"
	default:
		data.PM2Status = "unknown"
	}
	data.CPU = pm2val("CPU")
	data.RAM = strings.TrimSuffix(strings.TrimSuffix(pm2val("RAM"), "g"), "G")
	data.PIDs = pm2val("PIDS")
	if data.PIDs == "" {
		data.PIDs = "100"
	}
	data.StartupFile = pm2val("STARTUP_FILE")
	data.CustomCmd = pm2val("CUSTOM_CMD")
	data.Workdir = pm2val("WORKDIR")
	data.CurrentVersion = pm2val("TAG")
	data.GitRepoURL = pm2val("GIT_URL")
	data.RequirementsSelected = pm2val("REQUIREMENTS") == "1"

	if idx := strings.Index(data.Container.Container, "_"); idx != -1 {
		data.Service = data.Container.Container[:idx]
	} else {
		data.Service = data.Container.Container
	}

	if err := pythonNodeAppsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - python_node_apps template render error: %v", err)
	}
}
