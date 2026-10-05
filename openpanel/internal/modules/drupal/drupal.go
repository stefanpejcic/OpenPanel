// Package drupal installs and manages a Composer-based Drupal site (`composer create-project drupal/recommended-project` + `drush site:install`) inside an existing domain's docroot, same shape as phpapp not wordpress - deliberately minimal, just install, a small read-only manage page, a one-time admin login link via Drush's `user:login`, and uninstall, MySQL/MariaDB only
package drupal

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:              handleInstallStream,
	InstallFields:        []string{"domain_id", "subdirectory", "site_name", "drupal_version", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:                handleDrupalLogin,
	Cache:                handleDrupalCacheRebuild,
	Logs:                 handleDrupalLogs,
	Maintenance:          handleDrupalMaintenance,
	CloneDBPrefix:        "drupal_clone_",
	CloneConfig:          cloneConfig,
	CloneNoSearchReplace: true,
	Update:               handleDrupalUpdate,
	CanClone:             true,
	APICloneForm:         apiCloneForm,
	APIUpdate:            cmsapp.SiteQuery("drupal", handleDrupalUpdate),
	APICache:             cmsapp.SiteQuery("drupal", handleDrupalCacheRebuild),
	Slug:                 "drupal",
	Name:                 "Drupal",
	RemoveConfig:         "settings.php",
	RemoveDBNameRE:       removeDBNameRE,
	RemoveDBUserRE:       removeDBUserRE,
	RemoveConfigPath: func(s *cmsapp.Site) string {
		return filepath.Join(s.HostPath, "web", "sites", "default", "settings.php")
	},
	RemoveConfigText: stripPHPCommentLines,
	DBMode:           cmsapp.DBOptionalPrefix,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractDrupalDatabaseInfoForBackup(userContext, docroot)
	},
	ConfigFile: "settings.php",
})

var (
	drupalBackupDBNameRE   = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)
	drupalBackupDBPrefixRE = regexp.MustCompile(`'prefix'\s*=>\s*'([^']*)'`)
)

// extractDrupalDatabaseInfoForBackup reads settings.php straight off the host filesystem, skipping the doc comment block the same way websites.extractDrupalDatabaseInfo does so the placeholder example lines don't match first
func extractDrupalDatabaseInfoForBackup(userContext, docroot string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(docroot, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "sites", "default", "settings.php"))
	if err != nil {
		return map[string]string{"error": "sites/default/settings.php not found"}
	}

	var codeLines []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	text := strings.Join(codeLines, "\n")

	nameMatch := drupalBackupDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in settings.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_prefix": ""}
	if prefixMatch := drupalBackupDBPrefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
