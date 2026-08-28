package websites

import (
	"database/sql"
	"net/http"
	"net/url"
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

// explorerHref builds a disk-usage/inodes-explorer link (base is
// "/disk-usage/" or "/inodes-explorer/") for a site's docroot. Both
// explorers browse the account's home directory
// (/home/<userContext>/...), not the docroot's container-side path
// (/var/www/html/<...>) - the docroot volume is bind-mounted from
// docker-data/volumes/<userContext>_html_data/_data/ under that home
// directory, so that's the prefix used to translate one into the other.
func explorerHref(base, userContext, docroot string) string {
	rel := strings.TrimPrefix(docroot, "/var/www/html/")
	full := "docker-data/volumes/" + userContext + "_html_data/_data/" + rel
	segments := strings.Split(full, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return base + strings.Join(segments, "/")
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

	basePageData := pageData{
		CurrentDomain:        websiteParam,
		Docroot:              docroot,
		PagespeedAPIKeyValue: pagespeedAPIKey,
		DiskUsageHref:        explorerHref("/disk-usage/", userContext, docroot),
		InodesExplorerHref:   explorerHref("/inodes-explorer/", userContext, docroot),
	}

	switch cmsType {
	case "wordpress":
		backupFilesAvailable := checkBackupFilesExist(a, r, websiteParam)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		renderWPSinglePage(a, w, r, WPSinglePageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			BackupFilesAvailable: backupFilesAvailable,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "python", "nodejs", "ruby":
		pm2Type := "NODE"
		if cmsType == "python" {
			pm2Type = "PY"
		} else if cmsType == "ruby" {
			pm2Type = "RUBY"
		}
		pm2Data := getPM2ForApplication(a, r, userContext, container.Container, pm2Type)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		renderPythonNodeAppsPage(a, w, r, PythonNodeAppsPageData{
			pageData:  basePageData,
			Container: container,
			PM2Data:   pm2Data,
			EnvVars:   getCurrentEnvVars(userContext, container.Container),
			Domains:   domains,
		})

	case "php":
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		settings := getPHPAppSettings(userContext, websiteParam)
		renderPHPAppPage(a, w, r, PHPAppPageData{
			pageData:                   basePageData,
			Container:                  container,
			PHPVersion:                 currentPHPVersion,
			InitialProject:             settings.initialProject,
			AutorunComposerInstall:     settings.autorunComposerInstall,
			ComposerOptimizeAutoloader: settings.composerOptimizeAutoloader,
		})

	case "websitebuilder", "sitebuilder":
		renderWebsiteBuilderPage(a, w, r, WebsiteBuilderPageData{
			pageData:  basePageData,
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
			pageData:             basePageData,
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

	case "flarum":
		dbInfo := extractFlarumDatabaseInfo(userContext, docroot)
		flarumVersion := getFlarumVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderFlarumAppPage(a, w, r, FlarumAppPageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			FlarumVersion:        flarumVersion,
			PHPVersion:           currentPHPVersion,
			MySQLVersion:         mysqlVersion,
			DBInfo:               dbInfo,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "sofawiki":
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderSofawikiAppPage(a, w, r, SofawikiAppPageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
		})

	case "dokuwiki":
		dokuwikiVersion := getDokuwikiVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderDokuwikiAppPage(a, w, r, DokuwikiAppPageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			DokuwikiVersion:      dokuwikiVersion,
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
			pageData:             basePageData,
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
			pageData:             basePageData,
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
			pageData:             basePageData,
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
			pageData:             basePageData,
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
			pageData:             basePageData,
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
			pageData:             basePageData,
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

	case "mediawiki":
		dbInfo := extractMediaWikiDatabaseInfo(userContext, docroot)
		mediawikiVersion := getMediaWikiVersion(userContext, docroot)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		mysqlVersion := getMySQLVersion(a, r, userContext)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		renderMediaWikiAppPage(a, w, r, MediaWikiAppPageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			MediaWikiVersion:     mediawikiVersion,
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

// getCurrentEnvVars reads a Python/NodeJS/Ruby app's current
// docker-compose.yml `environment:` list for the Env Vars tab's textarea -
// empty string if the service has none set yet (the common case: these
// apps ship with no environment: block at all until the tab's own save
// handler adds one). containerName mirrors getPM2ForApplication's own
// "strip prefix at the first underscore" normalization, since the DB's
// sites.container value (e.g. "RUBYTEST") can carry a user-id suffix the
// compose file's lowercase service key never has.
func getCurrentEnvVars(userContext, containerName string) string {
	normalized := containerName
	if idx := strings.Index(normalized, "_"); idx != -1 {
		normalized = normalized[:idx]
	}
	serviceName := strings.ToLower(normalized)

	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		return ""
	}
	services, ok := composeData["services"].(map[string]any)
	if !ok {
		return ""
	}
	svc, ok := services[serviceName].(map[string]any)
	if !ok {
		return ""
	}
	rawEnv, ok := svc["environment"].([]any)
	if !ok {
		return ""
	}
	lines := make([]string, 0, len(rawEnv))
	for _, v := range rawEnv {
		if s, ok := v.(string); ok {
			lines = append(lines, s)
		}
	}
	return strings.Join(lines, "\n")
}
