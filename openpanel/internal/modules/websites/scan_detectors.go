package websites

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// This file holds the eight walkers handleSitesScan (scan.go) calls, one
// per CMS type. Each mirrors wordpress/manage.go's walkForWPConfig +
// handleScanWordPress shape: walk the user's html_data volume for that
// type's marker config file, skip anything already tracked, extract the DB
// name/host, repair the host if it's empty/localhost/127.0.0.1, verify
// connectivity, then insert. See scan.go's package doc comment for why the
// host-repair step exists here (it's not needed by WordPress's original
// wp-cli-mediated scan, but is needed for all eight here since this
// implementation reads every config file directly instead).

// ---------------------- WordPress ---------------------- //

var (
	scanWPDBHostRE = regexp.MustCompile(`define\(\s*'DB_HOST'\s*,\s*'([^']*)'\s*\)`)
	scanWPDBNameRE = regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'([^']*)'\s*\)`)
)

func scanWordPress(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "wp-config.php" {
			return nil
		}
		root := filepath.Dir(path)
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		configFile := path
		content, readErr := os.ReadFile(configFile)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanWPDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(configFile, text, scanWPDBHostRE, "define('DB_HOST', '"+mysqlVersion+"')"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in wp-config.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanReadVersionSimple(filepath.Join(root, "wp-includes", "version.php"), regexp.MustCompile(`\$wp_version\s*=\s*'([^']*)'`))
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "wordpress", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "wordpress", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// ---------------------- Joomla ---------------------- //

var (
	scanJoomlaHostRE   = regexp.MustCompile(`\$host\s*=\s*'([^']*)'`)
	scanJoomlaDBNameRE = regexp.MustCompile(`\$db\s*=\s*'([^']*)'`)
)

func scanJoomla(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "configuration.php" {
			return nil
		}
		root := filepath.Dir(path)
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanJoomlaDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanJoomlaHostRE, "$$host = '"+mysqlVersion+"'"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in configuration.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanJoomlaVersion(root)
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "joomla", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "joomla", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

func scanJoomlaVersion(root string) string {
	content, err := os.ReadFile(filepath.Join(root, "libraries", "src", "Version.php"))
	if err != nil {
		return "Unknown"
	}
	text := string(content)
	major := regexp.MustCompile(`MAJOR_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	minor := regexp.MustCompile(`MINOR_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	patch := regexp.MustCompile(`PATCH_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	if major == nil || minor == nil || patch == nil {
		return "Unknown"
	}
	return major[1] + "." + minor[1] + "." + patch[1]
}

// ---------------------- OpenCart ---------------------- //

var (
	scanOCDBHostnameRE = regexp.MustCompile(`define\('DB_HOSTNAME',\s*'([^']*)'\)`)
	scanOCDBDatabaseRE = regexp.MustCompile(`define\('DB_DATABASE',\s*'([^']*)'\)`)
)

func scanOpenCart(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "config.php" {
			return nil
		}
		root := filepath.Dir(path)
		content, readErr := os.ReadFile(path)
		if readErr != nil || !scanOCDBHostnameRE.Match(content) {
			return nil
		}
		if info, statErr := os.Stat(filepath.Join(root, "admin", "config.php")); statErr != nil || info.IsDir() {
			return nil
		}

		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		text := string(content)
		nameMatch := scanOCDBDatabaseRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanOCDBHostnameRE, "define('DB_HOSTNAME', '"+mysqlVersion+"')"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in config.php)")
			return nil
		}
		adminConfigPath := filepath.Join(root, "admin", "config.php")
		adminContent, _ := os.ReadFile(adminConfigPath)
		if _, ok := scanRepairHostInFile(adminConfigPath, string(adminContent), scanOCDBHostnameRE, "define('DB_HOSTNAME', '"+mysqlVersion+"')"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in admin/config.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanReadVersionSimple(filepath.Join(root, "index.php"), regexp.MustCompile(`VERSION'\s*,\s*'([^']*)'`))
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "opencart", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "opencart", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// ---------------------- Nextcloud ---------------------- //

var (
	scanNCDBHostRE = regexp.MustCompile(`'dbhost'\s*=>\s*'([^']*)'`)
	scanNCDBNameRE = regexp.MustCompile(`'dbname'\s*=>\s*'([^']*)'`)
)

func scanNextcloud(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "config.php" || filepath.Base(filepath.Dir(path)) != "config" {
			return nil
		}
		root := filepath.Dir(filepath.Dir(path))
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanNCDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanNCDBHostRE, "'dbhost' => '"+mysqlVersion+"'"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in config/config.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanReadVersionSimple(filepath.Join(root, "version.php"), regexp.MustCompile(`OC_VersionString\s*=\s*'([^']*)'`))
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "nextcloud", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "nextcloud", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// ---------------------- PrestaShop ---------------------- //

var (
	scanPSDBHostRE = regexp.MustCompile(`'database_host'\s*=>\s*'([^']*)'`)
	scanPSDBNameRE = regexp.MustCompile(`'database_name'\s*=>\s*'([^']*)'`)
)

func scanPrestashop(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "parameters.php" || filepath.Base(filepath.Dir(path)) != "config" {
			return nil
		}
		appDir := filepath.Dir(filepath.Dir(path))
		if filepath.Base(appDir) != "app" {
			return nil
		}
		root := filepath.Dir(appDir)
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanPSDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanPSDBHostRE, "'database_host' => '"+mysqlVersion+"'"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in app/config/parameters.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		// PrestaShop's compiled Symfony cache bakes in DB credentials -
		// clear it after any repair so a stale container config doesn't
		// 500 the front-end even though parameters.php is now correct
		// (same fix applied by prestashop/clone.go this session).
		_ = os.RemoveAll(filepath.Join(root, "var", "cache", "prod"))
		_ = os.RemoveAll(filepath.Join(root, "var", "cache", "dev"))

		version := scanReadVersionSimple(filepath.Join(root, "src", "Core", "Version.php"), regexp.MustCompile(`const VERSION\s*=\s*'([^']*)'`))
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "prestashop", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "prestashop", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// ---------------------- Drupal ---------------------- //

var (
	scanDrupalDBHostRE = regexp.MustCompile(`'host'\s*=>\s*'([^']*)'`)
	scanDrupalDBNameRE = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)
	scanDrupalVerRE    = regexp.MustCompile(`"version":\s*"([^"]*)"`)
)

// stripPHPCommentLinesForScan drops lines whose first non-whitespace
// character is "*" - settings.php's leading documentation block has
// placeholder 'database' => 'database_name' style example lines that
// would otherwise match before the real, appended $databases array (same
// technique drupal/backups.go's extractDrupalDatabaseInfoForBackup uses).
func stripPHPCommentLinesForScan(content string) string {
	var codeLines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	return strings.Join(codeLines, "\n")
}

func scanDrupal(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "settings.php" || filepath.Base(filepath.Dir(path)) != "default" {
			return nil
		}
		sitesDir := filepath.Dir(filepath.Dir(path))
		if filepath.Base(sitesDir) != "sites" {
			return nil
		}
		root := filepath.Dir(sitesDir)
		// drupal/recommended-project (this panel's own install.go, and the
		// standard modern Composer-based Drupal layout generally) puts the
		// real document root in a "web" subdirectory one level below the
		// actual site root, with the top-level docroot entry symlinked
		// into it - so a site root ending in "web" needs its parent used
		// for the site name/version lookup, matching how install.go
		// registers the domain's docroot (without "/web") in the first
		// place. Without this, a found site here would get an extra
		// "/web" appended to its site name, mismatching the already-
		// tracked row and importing a bogus duplicate (caught live during
		// this feature's own verification).
		if filepath.Base(root) == "web" {
			root = filepath.Dir(root)
		}
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fullText := string(content)
		codeOnly := stripPHPCommentLinesForScan(fullText)
		nameMatch := scanDrupalDBNameRE.FindStringSubmatch(codeOnly)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		hostMatch := scanDrupalDBHostRE.FindStringSubmatch(codeOnly)
		host := ""
		if hostMatch != nil {
			host = hostMatch[1]
		}
		if host == "" || host == "localhost" || host == "127.0.0.1" {
			newContent := scanDrupalDBHostRE.ReplaceAllString(fullText, "'host' => '"+mysqlVersion+"'")
			if newContent != fullText {
				if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
					outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in sites/default/settings.php)")
					return nil
				}
			}
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanDrupalVersion(root)
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "drupal", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "drupal", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

func scanDrupalVersion(root string) string {
	content, err := os.ReadFile(filepath.Join(root, "composer.lock"))
	if err != nil {
		return "Unknown"
	}
	text := string(content)
	idx := strings.Index(text, `"name": "drupal/core-recommended"`)
	if idx == -1 {
		idx = strings.Index(text, `"name": "drupal/core"`)
	}
	if idx == -1 {
		return "Unknown"
	}
	m := scanDrupalVerRE.FindStringSubmatch(text[idx:])
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// ---------------------- Matomo ---------------------- //

var (
	scanMatomoDBHostRE = regexp.MustCompile(`(?m)^host\s*=\s*"([^"]*)"`)
	scanMatomoDBNameRE = regexp.MustCompile(`(?m)^dbname\s*=\s*"([^"]*)"`)
	scanMatomoVerRE    = regexp.MustCompile(`VERSION\s*=\s*'([^']*)'`)
)

func scanMatomo(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "config.ini.php" || filepath.Base(filepath.Dir(path)) != "config" {
			return nil
		}
		root := filepath.Dir(filepath.Dir(path))
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanMatomoDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanMatomoDBHostRE, `host = "`+mysqlVersion+`"`); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in config/config.ini.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanReadVersionSimple(filepath.Join(root, "core", "Version.php"), scanMatomoVerRE)
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "matomo", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "matomo", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// ---------------------- Moodle ---------------------- //

var (
	scanMoodleDBHostRE  = regexp.MustCompile(`CFG->dbhost\s*=\s*'([^']*)'`)
	scanMoodleDBNameRE  = regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)
	scanMoodleWwwrootRE = regexp.MustCompile(`CFG->wwwroot\s*=\s*'https?://([^']*)'`)
	// Captures only the leading numeric portion (e.g. "5.2.1+") of
	// version.php's $release, not its full human-friendly string (e.g.
	// "5.2.1+ (Build: 20260807)") - the full string overflows the sites
	// table's version column (confirmed live: "Data too long for column
	// 'version'"), and install.go itself only ever stores the short form
	// anyway (the requested/latest version string, not $release).
	scanMoodleReleaseRE = regexp.MustCompile(`\$release\s*=\s*'([\d.]+\+?)`)
)

// scanMoodle differs from every other detector: Moodle's config.php does
// NOT live under any domain's docroot (docroot is a symlink to
// <approot>/public - see moodle package's doc comment for why), so instead
// of walking under each domain's directory, this walks the whole
// html_data root for any config.php containing the "$CFG->dbhost" marker,
// wherever it happens to live, and derives the domain from the config's
// own $CFG->wwwroot (which install.go always sets to
// "https://domain[/subdir]") rather than from the file's path - this
// generalizes to both this panel's own "<slug>_moodleapp" convention and a
// hypothetical foreign/migrated Moodle layout placed directly under a
// domain's docroot.
func scanMoodle(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "config.php" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil || !scanMoodleDBHostRE.Match(content) {
			return nil
		}
		text := string(content)

		wwwrootMatch := scanMoodleWwwrootRE.FindStringSubmatch(text)
		if wwwrootMatch == nil {
			return nil
		}
		siteName := strings.TrimSuffix(wwwrootMatch[1], "/")
		domainName := siteName
		if idx := strings.Index(siteName, "/"); idx != -1 {
			domainName = siteName[:idx]
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		nameMatch := scanMoodleDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanMoodleDBHostRE, "CFG->dbhost = '"+mysqlVersion+"'"); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in config.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		root := filepath.Dir(path)
		version := scanReadVersionSimple(filepath.Join(root, "public", "version.php"), scanMoodleReleaseRE)
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "moodle", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "moodle", path: siteName, version: version})
		}
		return nil
	})
}

// ---------------------- SHARED HELPERS ---------------------- //

// scanReadVersionSimple reads path and applies re, returning "Unknown" on
// any failure - the common shape of every getXVersion-style function this
// session already has, duplicated here per this package's local-helper
// convention.
func scanReadVersionSimple(path string, re *regexp.Regexp) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "Unknown"
	}
	m := re.FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// scanRepairHostInFile rewrites the DB-host field matched by hostRE to
// replacement (in place, host-side) only if the currently-matched value is
// empty/localhost/127.0.0.1. Returns (didRewrite, ok) - ok is false only on
// a write error, letting callers distinguish "nothing needed fixing" from
// "tried to fix and failed".
//
// CAUTION for callers: replacement goes through regexp.ReplaceAllString,
// where a literal "$" is NOT literal - Go's regexp package treats $name as
// a capture-group reference (expanding an unmatched one to "") and $$ as
// ---------------------- MediaWiki ---------------------- //

var (
	scanMediaWikiHostRE   = regexp.MustCompile(`\$wgDBserver\s*=\s*"([^"]*)"`)
	scanMediaWikiDBNameRE = regexp.MustCompile(`\$wgDBname\s*=\s*"([^"]*)"`)
)

func scanMediaWiki(ctx context.Context, a *appctx.App, userID int, userContext, baseDirectory, wwwBaseDirectory, mysqlVersion string, outcome *scanOutcome) {
	_ = filepath.WalkDir(baseDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if scanSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "LocalSettings.php" {
			return nil
		}
		root := filepath.Dir(path)
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		siteName, domainName := siteNameAndDomain(containerRoot, wwwBaseDirectory)

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) || scanCheckSiteExists(a, siteName) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(content)
		nameMatch := scanMediaWikiDBNameRE.FindStringSubmatch(text)
		if nameMatch == nil {
			return nil
		}
		dbName := nameMatch[1]

		if _, ok := scanRepairHostInFile(path, text, scanMediaWikiHostRE, `$$wgDBserver = "`+mysqlVersion+`"`); !ok {
			outcome.skipped = append(outcome.skipped, siteName+" (could not repair DB host in LocalSettings.php)")
			return nil
		}

		if !scanVerifyDB(ctx, userContext, dbName) {
			outcome.skipped = append(outcome.skipped, siteName+" (database '"+dbName+"' not reachable)")
			return nil
		}

		version := scanReadVersionSimple(filepath.Join(root, "includes", "Defines.php"), regexp.MustCompile(`define\(\s*'MW_VERSION',\s*'([^']*)'\s*\)`))
		if insertErr := insertScannedSite(ctx, a, siteName, domainName, "mediawiki", version); insertErr != nil {
			outcome.skipped = append(outcome.skipped, siteName+" (insert failed: "+insertErr.Error()+")")
		} else {
			outcome.found = append(outcome.found, scanFound{cmsType: "mediawiki", path: strings.TrimPrefix(containerRoot, wwwBaseDirectory), version: version})
		}
		return nil
	})
}

// an escaped literal "$". Two live bugs here already came from getting
// this wrong in opposite directions:
//   - Joomla's hostRE (`\$host\s*=...`) INCLUDES the "$" in the match, so
//     the replacement must supply its own escaped "$$host" - an
//     unescaped "$host" silently ate the whole "$host" token.
//   - Moodle's hostRE (`CFG->dbhost\s*=...`) does NOT include the "$"
//     (the source's leading "$" sits just outside the match and survives
//     untouched), so its replacement must NOT add another "$" prefix at
//     all - an over-cautious "$$CFG->dbhost" doubled up into a literal
//     "$$CFG->dbhost" in the file.
//
// Always check whether hostRE's own pattern starts with `\$` before
// deciding whether replacement needs one too.
func scanRepairHostInFile(path, currentText string, hostRE *regexp.Regexp, replacement string) (bool, bool) {
	hostMatch := hostRE.FindStringSubmatch(currentText)
	host := ""
	if hostMatch != nil {
		host = hostMatch[1]
	}
	if host != "" && host != "localhost" && host != "127.0.0.1" {
		return false, true
	}
	newContent := hostRE.ReplaceAllString(currentText, replacement)
	if newContent == currentText {
		return false, true
	}
	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		return false, false
	}
	return true, true
}
