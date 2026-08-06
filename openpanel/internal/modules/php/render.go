package php

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"php/_shared.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var defaultVersionPage = loadPage("php/default.html")
var settingsPage = loadPage("php/settings.html")

// infoPage is a standalone document (its own {{define "layout"}}, no panel
// chrome), since a raw `php -i` dump is meant to be viewed on its own, not
// inside the sidebar/header shell. It must NOT be combined with pageFiles:
// both it and base.html define "layout", and html/template panics on a
// duplicate {{define}} within one template set.
var infoPage = web.MustLoadPage("php/info.html")

var iniEditorPage = loadPage("php/ini_editor.html")
var optionsPage = loadPage("php/options.html")
var extensionsPage = loadPage("php/extensions.html")
var phpMyAdminUnavailablePage = loadPage("mysql/phpmyadmin_unavailable.html")

// DefaultVersionPageData is php/default.html's template context.
type DefaultVersionPageData struct {
	web.LayoutData
	PHPDefaultVersion string
	InstalledVersions []string
	Service           string
}

func renderDefaultPage(a *appctx.App, w http.ResponseWriter, r *http.Request, phpDefaultVersion, service string, installedVersions []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Default PHP version")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DefaultVersionPageData{LayoutData: layout, PHPDefaultVersion: phpDefaultVersion, InstalledVersions: installedVersions, Service: service}
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
}

func renderPHPSettingsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rows []PHPDomainRow, counts []PHPVersionCount, outdatedDomains int, availableVersions []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Settings")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sortedAvailable := append([]string(nil), availableVersions...)
	sortVersionsDesc(sortedAvailable)
	data := PHPSettingsPageData{
		LayoutData: layout, Domains: rows, VersionCounts: counts,
		OutdatedDomains: outdatedDomains, AvailablePHPVersions: sortedAvailable,
	}
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
}

func renderPHPIniEditorPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version, service, title, filePath, fileContent string, installedVersions []string, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPIniEditorPageData{
		LayoutData: layout, Version: version, Service: service, FilePath: filePath,
		FileContent: fileContent, InstalledVersions: installedVersions, Issues: issues,
	}
	if err := iniEditorPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - ini editor template render error: %v", err)
	}
}

// PHPOptionsPageData is php/options.html's template context (both the
// version-picker state and the per-version options-table state).
type PHPOptionsPageData struct {
	web.LayoutData
	Version           string
	InstalledVersions []string
	Fields            []OptionField
	Issues            []HealthIssue
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

func renderPHPOptionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, version, title string, fields []OptionField, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PHPOptionsPageData{LayoutData: layout, Version: version, Fields: fields, Issues: issues}
	if err := optionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - options template render error: %v", err)
	}
}

// PHPExtensionsPageData is php/extensions.html's template context (both the
// version-picker state and the per-version extensions-table state).
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

// PHPMyAdminUnavailablePageData is mysql/phpmyadmin_unavailable.html's
// template context.
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
