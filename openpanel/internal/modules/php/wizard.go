package php

import (
	"log"
	"net/http"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var wizardPage = loadPage("php/wizard.html")

// WizardVersion is one entry of the wizard's version select
type WizardVersion struct {
	Version   string `json:"v"`
	Label     string `json:"label"`
	Level     string `json:"level"`
	Installed bool   `json:"installed"`
	InUse     bool   `json:"in_use"` // default or used by a domain, so its php-fpm is most likely running
}

// WizardDomain is one domain and the PHP version its vhost points at
type WizardDomain struct {
	Domain  string `json:"d"`
	Version string `json:"v"`
	Level   string `json:"level"`
}

// PHPWizardPageData is php/wizard.html's template context, the page loads extensions and options itself from the existing JSON routes
type PHPWizardPageData struct {
	web.LayoutData
	Default     string
	WebServer   string
	IsLitespeed bool
	Versions    []WizardVersion
	Domains     []WizardDomain
	DomainCount int
	Open        string
	Tab         string
}

// Installed counts the versions the account can use
func (d PHPWizardPageData) Installed() []string {
	var out []string
	for _, v := range d.Versions {
		if v.Installed {
			out = append(out, v.Version)
		}
	}
	return out
}

// NotInstalled counts the known versions the provider hasn't set up for this account
func (d PHPWizardPageData) NotInstalled() []string {
	var out []string
	for _, v := range d.Versions {
		if !v.Installed {
			out = append(out, v.Version)
		}
	}
	return out
}

// wizardVersions lists installed versions first, then the API's other versions as not installed, each group newest first
func wizardVersions(installed []string, def string, domains []WizardDomain, isLitespeed bool, api map[string]VersionInfo) []WizardVersion {
	used := map[string]bool{def: true}
	for _, d := range domains {
		used[d.Version] = true
	}
	have := map[string]bool{}
	var out []WizardVersion
	add := func(v string, inst bool) {
		level, label := classifyPHPVersionLevel(v, api)
		if len(api) == 0 {
			level, label = "", ""
		}
		out = append(out, WizardVersion{Version: v, Label: label, Level: level, Installed: inst, InUse: inst && used[v]})
	}
	sorted := append([]string(nil), installed...)
	sortVersionsDesc(sorted)
	for _, v := range sorted {
		have[v] = true
		add(v, true)
	}
	var missing []string
	if isLitespeed {
		for v := range litespeedDefaultVersions {
			if !have[v] {
				missing = append(missing, v)
			}
		}
	} else {
		for v := range api {
			if !have[v] && phpVersionFormRE.MatchString(v) {
				missing = append(missing, v)
			}
		}
	}
	sortVersionsDesc(missing)
	for _, v := range missing {
		add(v, false)
	}
	return out
}

// wizardOpen picks the version the page opens on: ?v= when it's installed, else the default, else the newest
func wizardOpen(requested, def string, versions []WizardVersion) string {
	var first string
	for _, v := range versions {
		if !v.Installed {
			continue
		}
		if first == "" {
			first = v.Version
		}
		if v.Version == requested {
			return requested
		}
	}
	for _, v := range versions {
		if v.Installed && (v.Version == def || strings.HasPrefix(def, v.Version+".")) {
			return v.Version
		}
	}
	return first
}

// handlePHPWizard only renders, every save goes through the existing PHP routes
func handlePHPWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	def, _ := computeDefaultPHPVersionAndService(ctx, userContext, webServer, isLitespeed)
	api := fetchPHPVersionsAPI(ctx)

	var domains []WizardDomain
	list, _ := a.AllDomainsForUser(ctx, userID)
	// LiteSpeed runs one PHP for every domain, so there's nothing per domain to show
	if !isLitespeed {
		if len(list) > 0 {
			rows, _, _ := buildPHPDomainRows(userContext, list, api)
			for _, row := range rows {
				if row.PHPVersion != "/" {
					domains = append(domains, WizardDomain{Domain: row.DomainURL, Version: row.PHPVersion, Level: row.Level})
				}
			}
		}
		sort.SliceStable(domains, func(i, j int) bool { return domains[i].Domain < domains[j].Domain })
	}

	versions := wizardVersions(FetchPHPVersions(ctx, a, userContext), def, domains, isLitespeed, api)

	layout, _, err := web.BuildLayoutData(a, w, r, "PHP Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tab := r.URL.Query().Get("tab")
	if tab != "options" && tab != "domains" {
		tab = "extensions"
	}
	data := PHPWizardPageData{
		LayoutData: layout, Default: def, WebServer: webServer, IsLitespeed: isLitespeed,
		Versions: versions, Domains: domains, DomainCount: len(list), Open: wizardOpen(r.URL.Query().Get("v"), def, versions), Tab: tab,
	}
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PHP - wizard template render error: %v", err)
	}
}
