package websites

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"golang.org/x/net/idna"
)

// ContainerInfo is a site's row from the sites table plus its container.
type ContainerInfo struct {
	ID          int
	SiteName    string
	DomainID    int
	AdminEmail  string
	Version     string
	CreatedDate string
	Type        string
	Container   string
	Path        string
	Port        string
}

// getContainerFromDatabase looks up a site's container info by site name.
func getContainerFromDatabase(a *appctx.App, r *http.Request, siteName string) (ContainerInfo, bool) {
	var (
		c                                                            ContainerInfo
		adminEmail, version, createdDate, typ, container, path, port sql.NullString
		domainID                                                     sql.NullInt64
	)
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT id, site_name, domain_id, admin_email, version, created_date, type, container, path, ports
		FROM sites WHERE site_name = ? LIMIT 1`, siteName)
	if err := row.Scan(&c.ID, &c.SiteName, &domainID, &adminEmail, &version, &createdDate, &typ, &container, &path, &port); err != nil {
		return ContainerInfo{}, false
	}
	c.DomainID = int(domainID.Int64)
	c.AdminEmail, c.Version, c.CreatedDate = adminEmail.String, version.String, createdDate.String
	c.Type, c.Container, c.Path, c.Port = typ.String, container.String, path.String, port.String
	return c, true
}

// checkBackupFilesExist reports whether a domain has any backup files,
// memoized for 5 minutes since it's a directory listing checked on every
// page load.
func checkBackupFilesExist(a *appctx.App, r *http.Request, selectedDomain string) bool {
	result, _ := cache.Memoize(r.Context(), a.Cache, "check_backup_files_exist:"+selectedDomain, 5*time.Minute, func() (bool, error) {
		backupDir := "/var/www/html/backups/" + selectedDomain
		entries, err := os.ReadDir(backupDir)
		if err != nil {
			return false, nil
		}
		return len(entries) > 0, nil
	})
	return result
}

// getPagespeedInsightsAPIKey reads the user's PageSpeed Insights API key
// from disk, memoized for 60s since it's read on every page load.
func getPagespeedInsightsAPIKey(a *appctx.App, r *http.Request, userContext string) string {
	key, _ := cache.Memoize(r.Context(), a.Cache, "get_pagespeed_insights_api_key:"+userContext, 60*time.Second, func() (string, error) {
		content, err := os.ReadFile("/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/pagespeed_api_key.txt")
		if err != nil {
			return "", nil
		}
		return strings.TrimSpace(string(content)), nil
	})
	return key
}

// handleWebsiteDispatch resolves a domain to its site's container and
// renders the type-specific management page (WordPress, Python/Node app,
// site builder, ...).
func handleWebsiteDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domainParam := r.URL.Query().Get("domain")
	if domainParam == "" {
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}

	websiteParam := domainParam
	domain, folderParam := splitDomainAndFolder(domainParam)

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	lookupKey := domainParam
	if folderParam == "" {
		lookupKey = domainParam
	} else {
		lookupKey = websiteParam
	}
	container, found := getContainerFromDatabase(a, r, lookupKey)
	if !found {
		http.NotFound(w, r)
		return
	}

	domainNameUsed, idnaErr := idna.ToASCII(domain)
	if idnaErr != nil {
		flashAndRedirect(a, w, r, "danger", "Invalid domain name format.", "/sites")
		return
	}

	var docroot string
	row := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domainNameUsed)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "danger", "Unable to detect docroot for the domain.", "/sites")
		return
	}
	if folderParam != "" {
		docroot = docroot + "/" + folderParam
	}

	cmsType := strings.ToLower(container.Type)
	pagespeedAPIKey := getPagespeedInsightsAPIKey(a, r, userContext)

	switch cmsType {
	case "wordpress":
		backupFilesAvailable := checkBackupFilesExist(a, r, websiteParam)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		renderWPSinglePage(a, w, r, WPSinglePageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			BackupFilesAvailable: backupFilesAvailable,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "python", "nodejs":
		pm2Type := "NODE"
		if cmsType == "python" {
			pm2Type = "PY"
		}
		pm2Data := getPM2ForApplication(a, r, userContext, container.Container, pm2Type)
		renderPythonNodeAppsPage(a, w, r, PythonNodeAppsPageData{
			pageData:  pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Container: container,
			PM2Data:   pm2Data,
		})

	case "php":
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		settings := getPHPAppSettings(userContext, websiteParam)
		renderPHPAppPage(a, w, r, PHPAppPageData{
			pageData:                   pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Container:                  container,
			PHPVersion:                 currentPHPVersion,
			InitialProject:             settings.initialProject,
			AutorunComposerInstall:     settings.autorunComposerInstall,
			ComposerOptimizeAutoloader: settings.composerOptimizeAutoloader,
		})

	case "websitebuilder", "sitebuilder":
		renderWebsiteBuilderPage(a, w, r, WebsiteBuilderPageData{
			pageData:  pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Container: container,
		})

	case "drupal":
		dbInfo := extractDrupalDatabaseInfo(userContext, docroot)
		drupalVersion := getDrupalVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderDrupalAppPage(a, w, r, DrupalAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			DrupalVersion:        drupalVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "joomla":
		dbInfo := extractJoomlaDatabaseInfo(userContext, docroot)
		joomlaVersion := getJoomlaVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderJoomlaAppPage(a, w, r, JoomlaAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			JoomlaVersion:        joomlaVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "opencart":
		dbInfo := extractOpenCartDatabaseInfo(userContext, docroot)
		openCartVersion := getOpenCartVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderOpenCartAppPage(a, w, r, OpenCartAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			OpenCartVersion:      openCartVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "prestashop":
		dbInfo := extractPrestashopDatabaseInfo(userContext, docroot)
		prestashopVersion := getPrestashopVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderPrestashopAppPage(a, w, r, PrestashopAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			PrestashopVersion:    prestashopVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "nextcloud":
		dbInfo := extractNextcloudDatabaseInfo(userContext, docroot)
		nextcloudVersion := getNextcloudVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderNextcloudAppPage(a, w, r, NextcloudAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			NextcloudVersion:     nextcloudVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "matomo":
		dbInfo := extractMatomoDatabaseInfo(userContext, docroot)
		matomoVersion := getMatomoVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderMatomoAppPage(a, w, r, MatomoAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			MatomoVersion:        matomoVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "moodle":
		dbInfo := extractMoodleDatabaseInfo(userContext, docroot)
		moodleVersion := getMoodleVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderMoodleAppPage(a, w, r, MoodleAppPageData{
			pageData:             pageData{CurrentDomain: websiteParam, Docroot: docroot, PagespeedAPIKeyValue: pagespeedAPIKey},
			Domains:              domains,
			Container:            container,
			MoodleVersion:        moodleVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	default:
		// mautic/anything else: not a supported CMS type here.
		writeJSON(w, http.StatusOK, map[string]string{"error": "Unknown CMS type"})
	}
}

// phpAppSettings is a PHP app's .env-stored settings, written by
// internal/modules/phpapp's install/manage handlers.
type phpAppSettings struct {
	initialProject             string
	autorunComposerInstall     bool
	composerOptimizeAutoloader bool
}

// getPHPAppSettings reads a PHP app's settings from .env. There's no
// dedicated container to key them off of (unlike NodeJS/Python's
// getPM2ForApplication, keyed by container name) - php apps use a synthetic
// prefix derived from the site name instead (see
// internal/modules/phpapp/install.go's phpAppEnvPrefix, duplicated here).
func getPHPAppSettings(userContext, siteName string) phpAppSettings {
	prefix := docker.ServiceKeyPrefix(strings.ReplaceAll(siteName, "/", "_")) + "_PHP_"
	initialProject, _ := docker.GetEnvValue(userContext, prefix+"INITIAL_PROJECT")
	autorun, _ := docker.GetEnvValue(userContext, prefix+"AUTORUN_COMPOSER_INSTALL")
	optimize, _ := docker.GetEnvValue(userContext, prefix+"COMPOSER_OPTIMIZE_AUTOLOADER")
	return phpAppSettings{
		initialProject:             initialProject,
		autorunComposerInstall:     autorun == "1",
		composerOptimizeAutoloader: optimize == "1",
	}
}

// getPM2ForApplication reads the process-manager (PM2) env vars for a
// Python/Node app's container from the user's .env file.
func getPM2ForApplication(a *appctx.App, r *http.Request, userContext, prefix, appType string) map[string]string {
	normalized := prefix
	if idx := strings.Index(prefix, "_"); idx != -1 {
		normalized = prefix[:idx]
	}
	pm2Prefix := strings.ToUpper(normalized) + "_" + appType + "_"

	data := map[string]string{"prefix": pm2Prefix}
	data["status"] = "false"
	if docker.IsServiceRunning(r.Context(), userContext, strings.ToLower(normalized)) {
		data["status"] = "true"
	}

	envFile := "/home/" + userContext + "/.env"
	contentBytes, err := os.ReadFile(envFile)
	if err == nil {
		content := string(contentBytes)
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			if strings.HasPrefix(key, pm2Prefix) {
				data[key] = strings.Trim(strings.TrimSpace(parts[1]), `"`)
			}
		}
	}
	return data
}
