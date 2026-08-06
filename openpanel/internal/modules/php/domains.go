package php

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var phpVersionFormRE = regexp.MustCompile(`^\d+\.\d+$`)

// domainConfDeclaredVersionRE extracts the PHP version declared in a
// domain's vhost upstream line (`php-fpm-(\d+\.\d+):\d+`), distinct from
// domainConfVersionRE (php.go) which matches the looser `php-fpm-(\d+\.\d+)`
// against just the first line.
var domainConfDeclaredVersionRE = regexp.MustCompile(`php-fpm-(\d+\.\d+):\d+`)

// PHPDomainRow is one row of php/settings.html's table.
type PHPDomainRow struct {
	DomainID   int
	DomainURL  string
	PHPVersion string
	// Level is "unset" (php_version == "/"), "good", "secure", or
	// "unsupported" - which 3-bar badge color settings.html renders.
	Level string
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

// handlePHPDomains renders the per-domain PHP version table and handles
// the version-switch POST from it.
func handlePHPDomains(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		flashAndRedirect(a, w, r, "warning", "Only one PHP version can be used on Litespeed, to change version for all domains use this page.", "/php/default")
		return
	}

	phpVersionsData := fetchPHPVersionsAPI(ctx)
	installedVersions := FetchPHPVersions(ctx, a, userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainURL := r.Form.Get("domain_url")
		oldPHPVersion := r.Form.Get("old_php_version")
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
			flashAndRedirect(a, w, r, "error", "Invalid or unavailable PHP version selected.", target)
			return
		}
		if oldPHPVersion == "" || !phpVersionFormRE.MatchString(oldPHPVersion) {
			flashAndRedirect(a, w, r, "error", "Invalid current PHP version.", target)
			return
		}

		newPHPContainer := "php-fpm-" + newPHPVersion
		if !docker.IsServiceRunning(ctx, userContext, newPHPContainer) {
			result := docker.StartOrStopContainer(ctx, userContext, newPHPContainer, "activate", "run")
			if !result.Success {
				flashAndRedirect(a, w, r, "error", "Failed to start PHP "+newPHPVersion+": "+result.Message, target)
				return
			}
		}

		confFile := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
		content, readErr := os.ReadFile(confFile)
		if readErr != nil {
			flashAndRedirect(a, w, r, "error", "Failed to read vhost configuration for "+domainURL+": "+readErr.Error(), target)
			return
		}

		updated := strings.ReplaceAll(string(content), "php-fpm-"+oldPHPVersion, "php-fpm-"+newPHPVersion)
		_ = os.WriteFile(confFile, []byte(updated), 0o644)

		if strings.Contains(updated, "php-fpm-"+oldPHPVersion) {
			flashAndRedirect(a, w, r, "error", "Error occurred while updating PHP version for domain "+domainURL, target)
			return
		}

		stopPHPServiceIfRunningAndUnused(ctx, userContext, oldPHPVersion)

		_, _ = a.DB.ExecContext(ctx, "UPDATE domains SET php_version = ? WHERE domain_url = ?", newPHPVersion, domainURL)

		reloadArgv := podmanmanager.PodmanArgv(userContext, "restart", webServer)
		_ = podmanmanager.Command(ctx, userContext, reloadArgv).Run()

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, fmt.Sprintf("changed PHP version for domain %s from %s to %s", domainURL, oldPHPVersion, newPHPVersion), ipAddress)
		flashSess(a, w, r, "success", fmt.Sprintf("PHP version for domain %s updated from %s to %s", domainURL, oldPHPVersion, newPHPVersion))

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
		writeJSON(w, http.StatusOK, map[string]any{
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

// buildPHPDomainRows resolves each domain's PHP version by reading its
// vhost config and matching the php-fpm-X.Y:port upstream line, falling
// back to "/" when none is found, and tallies per-version counts plus the
// number of domains on an outdated (unsupported) version.
func buildPHPDomainRows(userContext string, domainsList []appctx.Domain, phpVersionsData map[string]VersionInfo) (rows []PHPDomainRow, counts []PHPVersionCount, outdatedDomains int) {
	countIndex := map[string]int{}

	for _, d := range domainsList {
		version := "/"
		configPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + d.DomainURL + ".conf"
		if content, err := os.ReadFile(configPath); err == nil {
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

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
