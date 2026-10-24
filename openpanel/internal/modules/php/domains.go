package php

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var phpVersionFormRE = regexp.MustCompile(`^\d+\.\d+$`)

// domainConfDeclaredVersionRE extracts the PHP version from a domain's vhost upstream line, distinct from domainConfVersionRE (php.go) which matches the looser pattern against just the first line
var domainConfDeclaredVersionRE = regexp.MustCompile(`php-fpm-(\d+\.\d+):\d+`)

// PHPDomainRow is one row of php/settings.html's table.
type PHPDomainRow struct {
	DomainID   int
	DomainURL  string
	PHPVersion string
	Level      string // "unset", "good", "secure", or "unsupported" - the 3-bar badge color settings.html renders
}

// PHPVersionCount is one entry of settings.html's summary counter row.
type PHPVersionCount struct {
	Version string
	Count   int
	Label   string
	Level   string // "good" | "secure" | "unsupported"
}

func classifyPHPVersionLevel(version string, data map[string]VersionInfo) (level, label string) {
	info, ok := data[version]
	if !ok {
		return "unsupported", "Unsupported"
	}
	if info.IsLatestVersion || info.IsFutureVersion || info.IsNextVersion {
		return "good", info.StatusLabel
	}
	if info.IsSecureVersion {
		return "secure", info.StatusLabel
	}
	return "unsupported", "Unsupported"
}

// handlePHPDomains renders the per-domain PHP version table and handles the version-switch POST from it
func handlePHPDomains(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		web.FlashRedirect(a, w, r, "warning", "Only one PHP version can be used on Litespeed, to change version for all domains use this page.", "/php/default")
		return
	}

	phpVersionsData := fetchPHPVersionsAPI(ctx)
	installedVersions := FetchPHPVersions(ctx, a, userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainURL := r.Form.Get("domain_url")
		newPHPVersion := r.Form.Get("new_php_version")
		redirectTo := r.Form.Get("redirect_to")
		if redirectTo == "" || !strings.HasPrefix(redirectTo, "/") || strings.HasPrefix(redirectTo, "//") || strings.Contains(redirectTo, "://") {
			redirectTo = ""
		}
		target := "/php/domains"
		if redirectTo != "" {
			target = redirectTo
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		validNew := newPHPVersion != "" && phpVersionFormRE.MatchString(newPHPVersion) && containsString(installedVersions, newPHPVersion)
		if !validNew {
			web.FlashRedirect(a, w, r, "error", "Invalid or unavailable PHP version selected.", target)
			return
		}
		confFile := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
		content, readErr := tenantfs.ReadFile(confFile)
		if readErr != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Failed to read vhost configuration for %(domain_url)s: %(error)s", "domain_url", domainURL, "error", readErr.Error()), target)
			return
		}
		// the page's old_php_version can be stale, what the vhost points at is what's really in use
		oldPHPVersion := currentVhostVersion(string(content))
		if oldPHPVersion == "" {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Could not determine the current PHP version for domain %(domain_url)s", "domain_url", domainURL), target)
			return
		}
		if oldPHPVersion == newPHPVersion {
			web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "Domain %(domain_url)s already uses PHP %(new_phpversion)s, nothing to change", "domain_url", domainURL, "new_phpversion", newPHPVersion), target)
			return
		}

		newPHPContainer := "php-fpm-" + newPHPVersion
		if !docker.IsServiceRunning(ctx, userContext, newPHPContainer) {
			result := docker.StartOrStopContainer(ctx, userContext, newPHPContainer, "activate", "run")
			if !result.Success {
				web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Failed to start PHP %(new_phpversion)s: %(message)s", "new_phpversion", newPHPVersion, "message", result.Message), target)
				return
			}
			webserver.WaitListening(ctx, userContext, newPHPContainer)
		}

		updateFailed := func() {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error occurred while updating PHP version for domain %(domain_url)s", "domain_url", domainURL), target)
		}
		updated, ok := switchVhostVersion(string(content), oldPHPVersion, newPHPVersion)
		if !ok {
			updateFailed()
			return
		}
		// a failed write used to go unnoticed since only the copy in memory was checked
		if writeErr := tenantfs.WriteFile(confFile, []byte(updated), 0o644); writeErr != nil {
			updateFailed()
			return
		}
		if written, err := tenantfs.ReadFile(confFile); err != nil || currentVhostVersion(string(written)) != newPHPVersion {
			updateFailed()
			return
		}

		_, _ = a.DB.ExecContext(ctx, "UPDATE domains SET php_version = ? WHERE domain_url = ?", newPHPVersion, domainURL)

		// old php only goes down once the webserver points elsewhere
		reloadErr := webserver.ReloadWebserver(ctx, userContext, webServer)
		if reloadErr == nil {
			stopPHPServiceIfRunningAndUnused(ctx, userContext, oldPHPVersion)
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, fmt.Sprintf("changed PHP version for domain %s from %s to %s", domainURL, oldPHPVersion, newPHPVersion), ipAddress)
		if reloadErr != nil {
			web.Flash(a, w, r, "error", web.Tr(a, r, "PHP version for domain %(domain_url)s was saved, but %(web_server)s could not apply it and still uses PHP %(old_phpversion)s: %(error)s", "domain_url", domainURL, "web_server", webServer, "old_phpversion", oldPHPVersion, "error", reloadErr.Error()))
		} else {
			web.Flash(a, w, r, "success", web.Tr(a, r, "PHP version for domain %(domain_url)s updated from %(old_phpversion)s to %(new_phpversion)s", "domain_url", domainURL, "old_phpversion", oldPHPVersion, "new_phpversion", newPHPVersion))
		}

		if redirectTo != "" {
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
	}

	phpDefaultVersion := webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION")

	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	rows, counts, outdated := buildPHPDomainRows(userContext, domainsList, phpVersionsData)

	if r.URL.Query().Get("output") == "json" {
		phpVersions := make(map[int]string, len(rows))
		for _, row := range rows {
			phpVersions[row.DomainID] = row.PHPVersion
		}
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"domains":                domainsList,
			"php_versions":           phpVersions,
			"available_php_versions": installedVersions,
			"php_default_version":    phpDefaultVersion,
			"php_versions_data":      phpVersionsData,
		})
		return
	}

	renderPHPSettingsPage(a, w, r, rows, counts, outdated, installedVersions)
}

// buildPHPDomainRows resolves each domain's PHP version from its vhost config's php-fpm-X.Y:port upstream line (falling back to "/"), and tallies per-version counts plus the number of domains on an outdated version
func buildPHPDomainRows(userContext string, domainsList []appctx.Domain, phpVersionsData map[string]VersionInfo) (rows []PHPDomainRow, counts []PHPVersionCount, outdatedDomains int) {
	countIndex := map[string]int{}

	for _, d := range domainsList {
		version := "/"
		configPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + d.DomainURL + ".conf"
		if content, err := tenantfs.ReadFile(configPath); err == nil {
			if m := domainConfDeclaredVersionRE.FindStringSubmatch(string(content)); m != nil {
				version = m[1]
			}
		}

		level := "unset"
		if version != "/" {
			level, _ = classifyPHPVersionLevel(version, phpVersionsData)
		}
		rows = append(rows, PHPDomainRow{DomainID: d.DomainID, DomainURL: d.DomainURL, PHPVersion: version, Level: level})

		if idx, ok := countIndex[version]; ok {
			counts[idx].Count++
		} else {
			countIndex[version] = len(counts)
			counts = append(counts, PHPVersionCount{Version: version, Count: 1})
		}
	}

	for i := range counts {
		level, label := classifyPHPVersionLevel(counts[i].Version, phpVersionsData)
		counts[i].Level = level
		counts[i].Label = label
		if level == "unsupported" {
			outdatedDomains += counts[i].Count
		}
	}

	return rows, counts, outdatedDomains
}

// currentVhostVersion is the php-fpm version a vhost points at, the same php-fpm-X.Y:port line the domains table reads, "" when there's none
func currentVhostVersion(content string) string {
	if m := domainConfDeclaredVersionRE.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	return ""
}

// switchVhostVersion points every php-fpm-<from> reference at <to>, ok is false when nothing changed or a reference was left behind
func switchVhostVersion(content, from, to string) (string, bool) {
	updated := strings.ReplaceAll(content, "php-fpm-"+from, "php-fpm-"+to)
	return updated, updated != content && !strings.Contains(updated, "php-fpm-"+from)
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
