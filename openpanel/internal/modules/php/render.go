package php

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func loadPage(files ...string) *web.Page {
	return web.LoadPage(append([]string{"php/_shared.html"}, files...)...)
}

var defaultVersionPage = loadPage("php/default.html")
var settingsPage = loadPage("php/settings.html")

// infoPage is a standalone document (own {{define "layout"}}, no panel chrome) since a raw `php -i` dump is meant to be viewed on its own - must NOT go through web.LoadPage since both it and base.html define "layout" and html/template panics on a duplicate
var infoPage = web.MustLoadPage("php/info.html")

var iniEditorPage = loadPage("php/ini_editor.html")
var optionsPage = loadPage("php/options.html", "partials/_db_tuning.html")
var extensionsPage = loadPage("php/extensions.html")
var phpMyAdminUnavailablePage = loadPage("mysql/phpmyadmin_unavailable.html")

// DefaultVersionPageData is php/default.html's template context.
type DefaultVersionPageData struct {
	web.LayoutData
	PHPDefaultVersion string
	Service           string
	IsLitespeed       bool
	Choices           []PHPVersionChoice
}

// PHPVersionChoice is one version card on the default PHP version page
type PHPVersionChoice struct {
	Version     string
	Description string
	Current     bool
	Level       string // good, secure or unsupported, for the badge colour
	Label       string
}

// phpVersionDescription is the same support hint the onboarding wizard shows, from the version's status label
func phpVersionDescription(info VersionInfo, known bool) string {
	if !known {
		return ""
	}
	label := strings.ToLower(info.StatusLabel)
	switch {
	// "Unsupported" contains "supported", so end of life has to be checked first
	case info.IsEOLVersion || strings.Contains(label, "unsupported"):
		return "No longer receives any patches - you should definitely plan an upgrade."
	case info.IsLatestVersion || strings.Contains(label, "latest"):
		return "The latest version - recommended for the newest features and performance."
	case strings.Contains(label, "security"):
		return "Security fixes only - fine for older apps, but you should plan an upgrade."
	case strings.Contains(label, "supported"):
		return "Fully supported - use this if your apps aren't yet tested on the latest version."
	}
	return ""
}

// phpVersionChoices lists installed versions newest first, on LiteSpeed only the versions it has images for
func phpVersionChoices(installed []string, current string, isLitespeed bool, api map[string]VersionInfo) []PHPVersionChoice {
	versions := append([]string{}, installed...)
	sort.Slice(versions, func(i, j int) bool { return versionLess(versions[j], versions[i]) })
	var out []PHPVersionChoice
	for _, v := range versions {
		if isLitespeed && !litespeedDefaultVersions[v] {
			continue
		}
		info, known := api[v]
		// the API only lists recent versions, anything older than all of them is long past end of life
		if !known && olderThanAll(v, api) {
			info, known = VersionInfo{IsEOLVersion: true}, true
		}
		level, label := "", ""
		if known {
			level, label = classifyPHPVersionLevel(v, api)
		}
		out = append(out, PHPVersionChoice{
			Version: v, Description: phpVersionDescription(info, known),
			Current: v == current || strings.HasPrefix(current, v+"."),
			Level:   level, Label: label,
		})
	}
	return out
}

func olderThanAll(v string, api map[string]VersionInfo) bool {
	if len(api) == 0 {
		return false
	}
	for known := range api {
		if !versionLess(v, known) {
			return false
		}
	}
	return true
}

// versionLess compares versions like 8.4 and 8.10 by number, not as text
func versionLess(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) && i < len(pb); i++ {
		na, _ := strconv.Atoi(pa[i])
		nb, _ := strconv.Atoi(pb[i])
		if na != nb {
			return na < nb
		}
	}
	return len(pa) < len(pb)
}

func renderDefaultPage(a *appctx.App, w http.ResponseWriter, r *http.Request, phpDefaultVersion, service string, installedVersions []string, isLitespeed bool) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Default PHP version")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DefaultVersionPageData{
		LayoutData: layout, PHPDefaultVersion: phpDefaultVersion, Service: service, IsLitespeed: isLitespeed,
		Choices: phpVersionChoices(installedVersions, phpDefaultVersion, isLitespeed, fetchPHPVersionsAPI(r.Context())),
	}
	if err := defaultVersionPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - default version template render error: %v", err)
	}
}

// PHPSettingsPageData is php/settings.html's template context.
type PHPSettingsPageData struct {
	web.LayoutData
	Domains              []PHPDomainRow
	VersionCounts        []PHPVersionCount
	OutdatedDomains      int
	AvailablePHPVersions []string
	DefaultVersion       string // DEFAULT_PHP_VERSION from the user's .env
	WebServer            string // label like Apache, named in the busy dialog's steps
}

// webServerLabel is how a WEB_SERVER value reads in page headings
func webServerLabel(ws string) string {
	switch ws {
	case "apache":
		return "Apache"
	case "nginx":
		return "Nginx"
	case "openresty":
		return "OpenResty"
	case "openlitespeed":
		return "OpenLiteSpeed"
	case "litespeed":
		return "LiteSpeed"
	}
	return ws
}

// envLabels reads the default PHP version and the web server label for the page eyebrows
func envLabels(injectedData map[string]any) (defaultVersion, webServer string) {
	userContext, _ := injectedData["context"].(string)
	if userContext == "" {
		return "", ""
	}
	return webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION"), webServerLabel(webserver.GetEnvFileValue(userContext, "WEB_SERVER"))
}

func renderPHPSettingsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rows []PHPDomainRow, counts []PHPVersionCount, outdatedDomains int, availableVersions []string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Settings")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sortedAvailable := append([]string(nil), availableVersions...)
	sortVersionsDesc(sortedAvailable)
	layout.BulkActions = phpDomainsBulkActions(layout.T, sortedAvailable)
	data := PHPSettingsPageData{
		LayoutData: layout, Domains: rows, VersionCounts: counts,
		OutdatedDomains: outdatedDomains, AvailablePHPVersions: sortedAvailable,
	}
	data.DefaultVersion, data.WebServer = envLabels(injectedData)
	if err := settingsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - settings template render error: %v", err)
	}
}

// PHPInfoPageData is php/info.html's template context.
type PHPInfoPageData struct {
	web.LayoutData
	Version           string
	Service           string
	FileContent       string
	InstalledVersions []string
}

func renderPHPInfoPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version, service, title, fileContent string, installedVersions []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPInfoPageData{LayoutData: layout, Version: version, Service: service, FileContent: fileContent, InstalledVersions: installedVersions}
	if err := infoPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - info template render error: %v", err)
	}
}

// PHPIniEditorPageData is php/ini_editor.html's template context.
type PHPIniEditorPageData struct {
	web.LayoutData
	Version           string
	Service           string
	FilePath          string
	FileContent       string
	InstalledVersions []string
	Issues            []HealthIssue
	WebServer         string // label like Apache
	SavedPath         string // where the file lives in the user's home
}

// LineCount is how many lines the php.ini has, for the editor footer
func (d PHPIniEditorPageData) LineCount() int {
	if d.FileContent == "" {
		return 0
	}
	return strings.Count(strings.TrimSuffix(d.FileContent, "\n"), "\n") + 1
}

func renderPHPIniEditorPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version, service, title, filePath, fileContent string, installedVersions []string, issues []HealthIssue) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPIniEditorPageData{
		LayoutData: layout, Version: version, Service: service, FilePath: filePath,
		FileContent: fileContent, InstalledVersions: installedVersions, Issues: issues,
	}
	_, data.WebServer = envLabels(injectedData)
	if userContext, _ := injectedData["context"].(string); userContext != "" && version != "" {
		data.SavedPath = "/home/" + userContext + "/php.ini/" + version + ".ini"
	}
	if err := iniEditorPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - ini editor template render error: %v", err)
	}
}

// PHPOptionsPageData is php/options.html's template context (both the version-picker state and the per-version options-table state)
type PHPOptionsPageData struct {
	web.LayoutData
	Version           string
	InstalledVersions []string
	Fields            []OptionField
	Issues            []HealthIssue
	StartupFunctions  []string // disable_functions entries that stop PHP-FPM from starting, warned about live on the field
	WebServer         string   // label like Apache
}

func renderPHPOptionsSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, title string, installedVersions []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPOptionsPageData{LayoutData: layout, InstalledVersions: installedVersions}
	if err := optionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - options select template render error: %v", err)
	}
}

func renderPHPOptionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version, title string, fields []OptionField, issues []HealthIssue, startupFuncs []string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPOptionsPageData{LayoutData: layout, Version: version, Fields: fields, Issues: issues, StartupFunctions: startupFuncs}
	_, data.WebServer = envLabels(injectedData)
	if err := optionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - options template render error: %v", err)
	}
}

// PHPExtensionsPageData is php/extensions.html's template context (both the version-picker state and the per-version extensions-table state)
type PHPExtensionsPageData struct {
	web.LayoutData
	Version           string
	InstalledVersions []string
	Extensions        []ExtensionRow
	ExtensionsHistory []string
	RecentlyRemoved   []string
}

func renderPHPExtensionsSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, installedVersions []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "PHP Extensions")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPExtensionsPageData{LayoutData: layout, InstalledVersions: installedVersions}
	if err := extensionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - extensions select template render error: %v", err)
	}
}

func renderPHPExtensionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version string, extensions []ExtensionRow, history, recentlyRemoved []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "PHP "+version+" Extensions")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPExtensionsPageData{
		LayoutData: layout, Version: version, Extensions: extensions,
		ExtensionsHistory: history, RecentlyRemoved: recentlyRemoved,
	}
	if err := extensionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - extensions template render error: %v", err)
	}
}

// PHPMyAdminUnavailablePageData is mysql/phpmyadmin_unavailable.html's template context
type PHPMyAdminUnavailablePageData struct {
	web.LayoutData
	ErrorMessage string
}

func renderPHPMyAdminUnavailablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, errorMessage string, status int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "phpMyAdmin")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPMyAdminUnavailablePageData{LayoutData: layout, ErrorMessage: errorMessage}
	if err := phpMyAdminUnavailablePage.Render(w, status, data); err != nil {
		log.Printf("PHP - phpmyadmin unavailable template render error: %v", err)
	}
}
