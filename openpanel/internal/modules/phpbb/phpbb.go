// Package phpbb installs and manages phpBB (phpbb.com) inside an existing domain's docroot and php-fpm container - same shape as internal/modules/flarum, the other MySQL-backed forum module
// unlike Flarum, phpBB ships no CLI update mechanism, so update.go doesn't exist here - the Update tab is browser-link-only, the same pattern Joomla/OpenCart/PrestaShop use
// install drives install/phpbbcli.php's "install" command (a full non-interactive equivalent of the browser setup wizard), not the regular bin/phpbbcli.php which refuses to run until phpBB is already installed - install/ is deleted afterward per phpBB's documented post-install security step
package phpbb

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var cms = cmsapp.New(cmsapp.App{
	Install:       handleInstallStream,
	InstallFields: []string{"domain_id", "subdirectory", "board_name", "board_description", "admin_username", "admin_password", "admin_email"},
	CloneDBPrefix: "phpbb_clone_",
	CloneConfig:   cloneConfig,
	CloneVersion: func(c *cmsapp.Clone) string {
		return web.FormOr(c.R, "version", phpbbVersion)
	},
	CanClone:     true,
	APICloneForm: apiCloneForm,
	Slug:         "phpbb",
	Name:         "phpBB",
	RemoveDB:     removePhpbbDB,
	DBMode:       cmsapp.DBOptionalPrefix,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractPhpbbDatabaseInfoForBackup(userContext, docroot)
	},
	ConfigFile:        "config.php",
	ChownAfterRestore: true,
})

// extractPhpbbDatabaseInfoForBackup reads config.php straight off the host filesystem, reusing phpbbDBNameRE (defined in manage.go)
func extractPhpbbDatabaseInfoForBackup(userContext, docroot string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(docroot, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config.php"))
	if err != nil {
		return map[string]string{"error": "config.php not found"}
	}
	text := string(content)

	nameMatch := phpbbDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config.php"}
	}
	return map[string]string{"database_name": nameMatch[1]}
}

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
