// Package flarum installs and manages a Composer-based Flarum forum (`composer create-project flarum/flarum`) inside an existing domain's docroot, same shape as internal/modules/drupal - the non-interactive installer is `php <installPath>/flarum install --file=<json> --config=<path>` (see install.go's doc comment for details that don't match docs.flarum.org's browser-wizard-only docs), and two Drupal-parity features are deliberately NOT implemented since Flarum has no equivalent: maintenance mode (Flarum core has no "site offline" concept) and admin auto-login (Flarum's console has no session/login command, and forging a session against access_tokens directly would be working around undocumented internals)
package flarum

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

var cms = cmsapp.New(cmsapp.App{
	Install:          handleInstallStream,
	InstallFields:    []string{"domain_id", "subdirectory", "site_name", "flarum_version", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Cache:            handleFlarumCacheClear,
	Logs:             handleFlarumLogs,
	CloneDBPrefix:    "flarum_clone_",
	CloneConfig:      cloneConfig,
	CloneAfterConfig: cloneAddCron,
	Update:           handleFlarumUpdate,
	CanClone:         true,
	APICloneForm:     apiCloneForm,
	APIUpdate:        cmsapp.SiteQuery("flarum", handleFlarumUpdate),
	APICache:         cmsapp.SiteQuery("flarum", handleFlarumCacheClear),
	Slug:             "flarum",
	Name:             "Flarum",
	RemoveConfig:     "config.php",
	RemoveDBNameRE:   removeDBNameRE,
	RemoveDBUserRE:   removeDBUserRE,
	OnRemove: func(ctx context.Context, s *cmsapp.Site) {
		_ = crons.RemoveJobByComment(ctx, s.UserContext, flarumCronComment(s.Name))
	},
	DBMode: cmsapp.DBOptionalPrefix,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractFlarumDatabaseInfoForBackup(userContext, docroot)
	},
	ConfigFile: "config.php",
})

var flarumBackupDBNameRE = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)

// extractFlarumDatabaseInfoForBackup reads config.php straight off the host filesystem
func extractFlarumDatabaseInfoForBackup(userContext, docroot string) map[string]string {
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

	nameMatch := flarumBackupDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config.php"}
	}
	return map[string]string{"database_name": nameMatch[1], "database_prefix": ""}
}

func Register(mux *http.ServeMux, a *appctx.App) {
	cms.Register(mux, a)
	scheduler := auth.RequireLogin(a, "flarum")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFlarumScheduler(a, w, r)
	}))
	mux.Handle("GET /flarum/scheduler", scheduler)
	mux.Handle("POST /flarum/scheduler", scheduler)
}

func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
