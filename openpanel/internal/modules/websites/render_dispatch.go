package websites

import (
	"log"
	"net/http"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// pageData is embedded by every /website dispatcher page's data struct so the shared partials (temp_link, screenshot, visitors, waf_panel, pagespeed_panel) can rely on a consistent field set regardless of which CMS type is being rendered.
type pageData struct {
	web.LayoutData
	CurrentDomain        string
	Docroot              string
	PagespeedAPIKeyValue string
	DiskUsageHref        string // points into the disk-usage/inodes-explorer browsers at this site's docroot, built from the account's home directory not the container-side Docroot path, computed once in handleWebsiteDispatch (see explorerHref)
	InodesExplorerHref   string
}

// PagespeedAPIKey satisfies the field name the pagespeed_panel partial reads, without colliding with the PagespeedAPIKeyValue field name used by every dispatch branch's literal struct above.
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

// SecurityToggle is one row of the Security tab's hardening-rule list - the rule IDs come from `opencli websites-secure`, but their labels/tech details are only defined here.
type SecurityToggle struct {
	ID, Label, TechDetails string
}

var wpSecurityToggles = []SecurityToggle{
	{"wp_manager_disable_wp_admin", "Disable wp-admin",
		"Returns a 401 error for requests to /wp-admin, completely blocking access to the WordPress dashboard. admin-ajax.php and admin-post.php stay reachable so forms and plugins on the site keep working. Turn it off again to log in to wp-admin."},
	{"wp_manager_mitigate_spam_logins", "Mitigate Spam Logins and Comments",
		"Drops POST requests to wp-login.php and wp-comments-post.php that don't come with a Referer from the same site, stopping bots that submit logins and comments directly."},
	{"wp_manager_wp_config", "Block access to wp-config.php",
		"Directly blocks any HTTP requests to the wp-config.php file, preventing exposure of database credentials even if PHP processing fails."},
	{"wp_manager_uploads_php", "Disable PHP in uploads",
		"Scans for any .php files within /wp-content/uploads/ and blocks them. Essential for preventing malicious shells from running after an unauthorized file upload."},
	{"wp_manager_xmlrpc", "Block access to xmlrpc.php",
		"Returns a 403 error for all xmlrpc.php requests. Disables XML-based remote publishing and Pingbacks to prevent DDoS and brute-force amplification."},
	{"wp_manager_env_files", "Protect Environment Files",
		"Blocks public access to .env (including .env.local, .env.production and similar), .htaccess, and .htpasswd files in any folder, which often contain high-value secrets and server configurations."},
	{"wp_manager_sensitive_files", "Protect Sensitive Files",
		"Uses Regex to block access to backup files (.bak, .swp), readme/license files, and dangerous extensions like .log, .sh, or .exe."},
	{"wp_manager_author_scan", "Block Author Enumeration",
		"Blocks requests containing the \"author=\" query parameter and the REST API users list (/wp-json/wp/v2/users), stopping bots from discovering valid administrative usernames. Logged in users can still use the users endpoint, so the block editor keeps working."},
	{"wp_manager_bad_bots", "Block Malicious Bots",
		"Uses a User-Agent filter to block aggressive scrapers and security scanners like AhrefsBot, SemrushBot, MJ12bot, and Nikto."},
	{"wp_manager_wp_includes_php", "Restrict wp-includes PHP",
		"Blocks all PHP execution in the wp-includes folder, with a specific exclusion for wp-tinymce.php to ensure the editor keeps working."},
	{"wp_manager_cache_php", "Disable PHP in Cache",
		"Prevents the execution of PHP scripts stored within cache directories, stopping \"cache poisoning\" or \"file inclusion\" exploits."},
	{"wp_manager_admin_script_concat", "Disable Script Concatenation",
		"Blocks access to load-scripts.php and load-styles.php. This prevents a common ReDoS vulnerability used to spike server CPU via the admin dashboard."},
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

var wpSinglePage = loadPage("manager/wp/single.html", "manager/wp/_shared.html", "manager/wp/_caching.html")

var pythonNodeAppsPage = loadPage("manager/python_node_apps.html")

// PythonNodeAppsPageData is manager/python_node_apps.html's template context.
type PythonNodeAppsPageData struct {
	pageData
	Container                                                                   ContainerInfo
	Service                                                                     string // container.container.split('_')[0] verbatim, the pm2/docker process id used throughout the page's JS
	PM2Data                                                                     map[string]string
	Type                                                                        string // Container.Type lowercased ("python" or "nodejs")
	PM2Status                                                                   string // pm2_data.status stringified ("true"/"false"/"unknown")
	CPU, RAM, PIDs, StartupFile, CustomCmd, Workdir, CurrentVersion, GitRepoURL string
	RequirementsSelected                                                        bool
	EnvVars                                                                     string          // the service's current `environment:` list from docker-compose.yml, one "KEY=VALUE" per line, for the Env Vars tab's textarea - empty if none set yet
	Domains                                                                     []appctx.Domain // every domain the user owns, for the Clone tab's target domain dropdown - same list appinstall's own install page uses
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
	if data.StartupFile == "" {
		// an empty STARTUP_FILE is a valid, common install-time choice (see appinstall/install.go's defaultStartupFile) - the docker-compose template substitutes the same per-type default at runtime, so show that default here instead of a blank field
		switch data.Type {
		case "nodejs":
			data.StartupFile = "index.js"
		case "ruby":
			data.StartupFile = "app.rb"
		default:
			data.StartupFile = "app.py"
		}
	}
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

var n8nAppPage = loadPage("manager/n8n_app.html")

// N8nAppPageData is manager/n8n_app.html's template context - a much smaller counterpart of PythonNodeAppsPageData since n8n has no startup file/package manager/git deploy concept (see appinstall.Kind.Simple).
type N8nAppPageData struct {
	pageData
	Container      ContainerInfo
	Service        string // container.container.split('_')[0] verbatim, the pm2/docker process id used throughout the page's JS
	PM2Data        map[string]string
	PM2Status      string // pm2_data.status stringified ("true"/"false"/"unknown")
	CPU, RAM, PIDs string
	CurrentVersion string
	EnvVars        string // the service's current `environment:` list from docker-compose.yml, one "KEY=VALUE" per line, for the Env Vars tab's textarea
}

func renderN8nAppPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data N8nAppPageData) {
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
	data.CurrentVersion = pm2val("TAG")

	if idx := strings.Index(data.Container.Container, "_"); idx != -1 {
		data.Service = data.Container.Container[:idx]
	} else {
		data.Service = data.Container.Container
	}

	if err := n8nAppPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - n8n_app template render error: %v", err)
	}
}

// CMSAppPageData is manager/<type>_app.html's template context for every one-click CMS site
type CMSAppPageData struct {
	pageData
	Domains              []appctx.Domain
	Container            ContainerInfo
	Version              string
	PHPVersion           string
	MySQLVersion         string
	DBInfo               map[string]string
	IsSubdirectory       bool
	MainDomain           string
	CurrentPHPVersion    string
	AvailablePHPVersions []string
	HasPhotos            bool
	App                  web.AppNames
}

// cmsPage is how the /website dispatcher reads one CMS type, dbInfo is nil for apps without a database
type cmsPage struct {
	name    string
	page    *web.Page
	dbInfo  func(userContext, docroot string) map[string]string
	version func(userContext, docroot string) string
}

var cmsPages = map[string]cmsPage{
	"drupal":           {name: "Drupal", page: loadPage("manager/_cms_app_shared.html", "manager/drupal_app.html"), dbInfo: extractDrupalDatabaseInfo, version: getDrupalVersion},
	"flarum":           {name: "Flarum", page: loadPage("manager/_cms_app_shared.html", "manager/flarum_app.html"), dbInfo: extractFlarumDatabaseInfo, version: getFlarumVersion},
	"sofawiki":         {name: "SofaWiki", page: loadPage("manager/_cms_app_shared.html", "manager/sofawiki_app.html"), dbInfo: nil, version: nil},
	"tinyphotogallery": {name: "TinyPhotoGallery", page: loadPage("manager/_cms_app_shared.html", "manager/tinyphotogallery_app.html"), dbInfo: nil, version: nil},
	"tinyfilemanager":  {name: "TinyFileManager", page: loadPage("manager/_cms_app_shared.html", "manager/tinyfilemanager_app.html"), dbInfo: nil, version: nil},
	"phpbb":            {name: "phpBB", page: loadPage("manager/_cms_app_shared.html", "manager/phpbb_app.html"), dbInfo: extractPhpbbDatabaseInfo, version: getPhpbbVersion},
	"dokuwiki":         {name: "DokuWiki", page: loadPage("manager/_cms_app_shared.html", "manager/dokuwiki_app.html"), dbInfo: nil, version: getDokuwikiVersion},
	"joomla":           {name: "Joomla", page: loadPage("manager/_cms_app_shared.html", "manager/joomla_app.html"), dbInfo: extractJoomlaDatabaseInfo, version: getJoomlaVersion},
	"opencart":         {name: "OpenCart", page: loadPage("manager/_cms_app_shared.html", "manager/opencart_app.html"), dbInfo: extractOpenCartDatabaseInfo, version: getOpenCartVersion},
	"prestashop":       {name: "PrestaShop", page: loadPage("manager/_cms_app_shared.html", "manager/prestashop_app.html"), dbInfo: extractPrestashopDatabaseInfo, version: getPrestashopVersion},
	"nextcloud":        {name: "Nextcloud", page: loadPage("manager/_cms_app_shared.html", "manager/nextcloud_app.html"), dbInfo: extractNextcloudDatabaseInfo, version: getNextcloudVersion},
	"matomo":           {name: "Matomo", page: loadPage("manager/_cms_app_shared.html", "manager/matomo_app.html"), dbInfo: extractMatomoDatabaseInfo, version: getMatomoVersion},
	"moodle":           {name: "Moodle", page: loadPage("manager/_cms_app_shared.html", "manager/moodle_app.html"), dbInfo: extractMoodleDatabaseInfo, version: getMoodleVersion},
	"ojs":              {name: "OJS", page: loadPage("manager/_cms_app_shared.html", "manager/ojs_app.html"), dbInfo: extractOJSDatabaseInfo, version: nil},
	"mediawiki":        {name: "MediaWiki", page: loadPage("manager/_cms_app_shared.html", "manager/mediawiki_app.html"), dbInfo: extractMediaWikiDatabaseInfo, version: getMediaWikiVersion},
}

func renderCMSAppPage(a *appctx.App, w http.ResponseWriter, r *http.Request, cmsType string, data CMSAppPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.CurrentDomain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	data.App = web.NewAppNames(cmsType, cmsPages[cmsType].name)
	sort.Sort(sort.Reverse(sort.StringSlice(data.AvailablePHPVersions)))
	if err := cmsPages[cmsType].page.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - %s_app template render error: %v", cmsType, err)
	}
}
