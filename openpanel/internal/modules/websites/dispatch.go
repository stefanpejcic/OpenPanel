package websites

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/domains"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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

// checkBackupFilesExist reports whether a domain has any backup files, memoized for 5 minutes since it's a directory listing checked on every page load.
func checkBackupFilesExist(a *appctx.App, r *http.Request, selectedDomain string) bool {
	result, _ := cache.Memoize(r.Context(), a.Cache, "check_backup_files_exist:"+selectedDomain, 5*time.Minute, func() (bool, error) {
		backupDir := "/var/www/html/backups/" + selectedDomain
		entries, err := tenantfs.ReadDir(backupDir)
		if err != nil {
			return false, nil
		}
		return len(entries) > 0, nil
	})
	return result
}

// explorerHref builds a disk-usage/inodes-explorer link (base "/disk-usage/" or "/inodes-explorer/") for a site's docroot - both explorers browse the account's home directory, not the docroot's container-side path, so the bind-mounted docker-data/volumes/<userContext>_html_data/_data/ prefix translates one into the other.
// dirHasEntries reports whether the given container-facing path (e.g. docroot+"/photos") exists and is non-empty on the host, translating it via the same html_data volume prefix explorerHref uses.
func dirHasEntries(userContext, containerPath string) bool {
	rel := strings.TrimPrefix(containerPath, "/var/www/html/")
	hostPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + rel
	entries, err := tenantfs.ReadDir(hostPath)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func explorerHref(base, userContext, docroot string) string {
	rel := strings.TrimPrefix(docroot, "/var/www/html/")
	full := "docker-data/volumes/" + userContext + "_html_data/_data/" + rel
	segments := strings.Split(full, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return base + strings.Join(segments, "/")
}

// getPagespeedInsightsAPIKey reads the user's PageSpeed Insights API key from disk, memoized for 60s since it's read on every page load.
func getPagespeedInsightsAPIKey(a *appctx.App, r *http.Request, userContext string) string {
	key, _ := cache.Memoize(r.Context(), a.Cache, "get_pagespeed_insights_api_key:"+userContext, 60*time.Second, func() (string, error) {
		content, err := tenantfs.ReadFile("/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/pagespeed_api_key.txt")
		if err != nil {
			return "", nil
		}
		return strings.TrimSpace(string(content)), nil
	})
	return key
}

// handleWebsiteDispatch resolves a domain to its site's container and renders the type-specific management page (WordPress, Python/Node app, site builder, ...).
func handleWebsiteDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := auth.Injected(a, r)
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
	// a domain proxied to a custom container is managed on the containers page
	if strings.EqualFold(container.Type, "docker") {
		target := "/containers/edit/"
		if r.URL.Query().Get("tab") == "remove" {
			target = "/containers/delete/"
		}
		http.Redirect(w, r, target+url.PathEscape(strings.ToLower(container.Container)), http.StatusFound)
		return
	}

	domainNameUsed, idnaErr := idna.ToASCII(domain)
	if idnaErr != nil {
		web.FlashRedirect(a, w, r, "danger", "Invalid domain name format.", "/sites")
		return
	}

	var docroot string
	row := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domainNameUsed)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		web.FlashRedirect(a, w, r, "danger", "Unable to detect docroot for the domain.", "/sites")
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

	if cp, ok := cmsPages[cmsType]; ok {
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		domains, _ := a.AllDomainsForUser(ctx, userID)
		data := CMSAppPageData{
			pageData:             basePageData,
			Domains:              domains,
			Container:            container,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: php.FetchPHPVersions(ctx, a, userContext),
		}
		if cp.dbInfo != nil {
			data.DBInfo = cp.dbInfo(userContext, docroot)
			data.PHPVersion = currentPHPVersion
			// a failed lookup is cached as "Version not found", show nothing rather than that
			if v := getMySQLVersion(a, r, userContext); v != "Version not found" {
				data.MySQLVersion = v
			}
			data.MySQLType = webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		}
		data.Switch = cmsSiteSwitch(r, a, userID)
		if cp.version != nil {
			data.Version = cp.version(userContext, docroot)
		}
		switch cmsType {
		case "ojs":
			data.Version = container.Version
		case "tinyphotogallery":
			data.HasPhotos = dirHasEntries(userContext, docroot+"/photos")
		}
		renderCMSAppPage(a, w, r, cmsType, data)
		return
	}

	switch cmsType {
	case "wordpress":
		backupFilesAvailable := checkBackupFilesExist(a, r, websiteParam)
		availablePHPVersions := php.FetchPHPVersions(ctx, a, userContext)
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		userDomains, _ := a.AllDomainsForUser(ctx, userID)
		suspended, https, cert := domains.SiteStatus(ctx, a, domainNameUsed)
		renderWPSinglePage(a, w, r, WPSinglePageData{
			pageData:             basePageData,
			Domains:              userDomains,
			Container:            container,
			BackupFilesAvailable: backupFilesAvailable,
			IsSubdirectory:       folderParam != "",
			MainDomain:           domain,
			CurrentPHPVersion:    currentPHPVersion,
			AvailablePHPVersions: availablePHPVersions,
			Suspended:            suspended,
			SSLState:             sslState(https, cert),
			SSLExpires:           cert.Expires,
			RedisOn:              wpRedisDropinOn(userContext, docroot),
			TakenSites:           accountSites(r, a, userID),
		})

	case "python", "nodejs", "ruby", "java":
		pm2Type := "NODE"
		if cmsType == "python" {
			pm2Type = "PY"
		} else if cmsType == "ruby" {
			pm2Type = "RUBY"
		} else if cmsType == "java" {
			pm2Type = "JAVA"
		}
		pm2Data := getPM2ForApplication(a, r, userContext, container.Container, pm2Type)
		svc := composeServiceFor(userContext, container.Container)
		manifestDir := strings.Trim(pm2Data[pm2Data["prefix"]+"WORKDIR"], `"`)
		if !strings.HasPrefix(manifestDir, "/var/www/html/") {
			manifestDir = docroot
		}
		_, https, cert := domains.SiteStatus(ctx, a, domainNameUsed)
		files, hidden := listSiteDir(userContext, docroot, 8)
		renderPythonNodeAppsPage(a, w, r, PythonNodeAppsPageData{
			pageData:     basePageData,
			Container:    container,
			PM2Data:      pm2Data,
			EnvVars:      envVarsOf(svc),
			Domains:      userDomains(ctx, a, userID),
			StartCommand: composeStartCommand(svc),
			Manifest:     loadManifest(userContext, cmsType, manifestDir),
			SSLState:     sslState(https, cert),
			SSLExpires:   cert.Expires,
			WebServer:    webserver.GetEnvFileValue(userContext, "WEB_SERVER"),
			Image:        svcImage(userContext, svc),
			Files:        files,
			HiddenFiles:  hidden,
		})

	case "n8n":
		pm2Data := getPM2ForApplication(a, r, userContext, container.Container, "N8N")
		svc := composeServiceFor(userContext, container.Container)
		renderN8nAppPage(a, w, r, N8nAppPageData{
			pageData:  basePageData,
			Container: container,
			PM2Data:   pm2Data,
			EnvVars:   envVarsOf(svc),
			Mounts:    composeMounts(svc),
			Networks:  composeNetworks(svc),
			Image:     svcImage(userContext, svc),
			Restart:   svcString(svc, "restart"),
		})

	case "php":
		currentPHPVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		settings := getPHPAppSettings(userContext, websiteParam)
		phpFiles, phpHidden := listSiteDir(userContext, docroot, 20)
		renderPHPAppPage(a, w, r, PHPAppPageData{
			pageData:                   basePageData,
			Container:                  container,
			PHPVersion:                 currentPHPVersion,
			InitialProject:             settings.initialProject,
			AutorunComposerInstall:     settings.autorunComposerInstall,
			ComposerOptimizeAutoloader: settings.composerOptimizeAutoloader,
			ComposerRows:               loadComposerRequirements(userContext, docroot),
			Packages:                   loadComposerPackages(userContext, docroot),
			Files:                      phpFiles,
			HiddenFiles:                phpHidden,
			WebServer:                  webserver.GetEnvFileValue(userContext, "WEB_SERVER"),
		})

	case "websitebuilder", "sitebuilder":
		source, hasSource := loadSiteSource(userContext, docroot)
		sbFiles, _ := listSiteDir(userContext, docroot, 12)
		renderWebsiteBuilderPage(a, w, r, WebsiteBuilderPageData{
			pageData:    basePageData,
			Container:   container,
			IndexSource: source,
			HasSource:   hasSource,
			BlankPage:   hasSource && blankPage(source),
			Files:       sbFiles,
		})

	default:
		// mautic/anything else: not a supported CMS type here
		web.WriteJSON(w, http.StatusOK, map[string]string{"error": "Unknown CMS type"})
	}
}

// phpAppSettings is a PHP app's .env-stored settings, written by internal/modules/phpapp's install/manage handlers.
type phpAppSettings struct {
	initialProject             string
	autorunComposerInstall     bool
	composerOptimizeAutoloader bool
}

// cmsSiteSwitch lists the account's sites that open in a CMS manager, for the header switcher
func cmsSiteSwitch(r *http.Request, a *appctx.App, userID int) []SiteOption {
	rows, err := a.DB.QueryContext(r.Context(), "SELECT site_name, type FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?) ORDER BY site_name", userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []SiteOption
	for rows.Next() {
		var name, typ sql.NullString
		if rows.Scan(&name, &typ) != nil || name.String == "" {
			continue
		}
		if cp, ok := cmsPages[strings.ToLower(typ.String)]; ok {
			out = append(out, SiteOption{Name: name.String, App: cp.name})
		}
	}
	return out
}

// getPHPAppSettings reads a PHP app's settings from .env - there's no dedicated container to key them off (unlike NodeJS/Python's getPM2ForApplication keyed by container name), so php apps use a synthetic prefix derived from the site name instead (see internal/modules/phpapp/install.go's phpAppEnvPrefix, duplicated here).
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

// getPM2ForApplication reads the process-manager (PM2) env vars for a Python/Node app's container from the user's .env file.
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
	contentBytes, err := tenantfs.ReadFile(envFile)
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

// getCurrentEnvVars reads a runtime app's compose `environment:` list for the Env Vars textarea, empty if none set yet
func getCurrentEnvVars(userContext, containerName string) string {
	return envVarsOf(composeServiceFor(userContext, containerName))
}

// userDomains is every domain on the account, nil on a lookup error
func userDomains(ctx context.Context, a *appctx.App, userID int) []appctx.Domain {
	d, _ := a.AllDomainsForUser(ctx, userID)
	return d
}
