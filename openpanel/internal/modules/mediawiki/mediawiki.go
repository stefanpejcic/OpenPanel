// Package mediawiki installs and manages a MediaWiki site (downloaded from releases.wikimedia.org's packaged tarballs + MediaWiki's own genuine non-interactive `maintenance/install.php` CLI installer) inside an existing domain's docroot, same shape as internal/modules/joomla (flat docroot, no public/ split) - since MediaWiki ships no CLI equivalent of Drupal's `drush uli` for a one-time admin login, this mirrors joomla/wordpress's approach instead: a small token table (created lazily, isolated from MediaWiki's own schema) plus a login helper PHP file deployed into the docroot at install time (see login_php.go) that verifies the token, then binds an admin User to the request's session through MediaWiki's own User::setCookies() API; and since MediaWiki's job queue (maintenance/runJobs.php) needs periodic execution for async work, install.go registers a per-minute job via internal/modules/crons.AddJob, and manage.go's uninstall handler removes it via crons.RemoveJobByComment
package mediawiki

import (
	"context"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// mediawikiCronComment is the crons.ini comment used both by install.go's crons.AddJob call and manage.go's crons.RemoveJobByComment call - it must be derived identically in both places to actually find/remove the same job later
func mediawikiCronComment(selectedDomain string) string {
	return "mediawiki-" + appkit.SiteSlug(selectedDomain)
}

var cms = cmsapp.New(cmsapp.App{
	Install:          handleInstallStream,
	InstallFields:    []string{"domain_id", "subdirectory", "mediawiki_version", "site_name", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:            handleMediaWikiLogin,
	Logs:             cmsapp.ShellLogs(`f=$(ls -t "$1"/logs/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "No log entries found. MediaWiki does not write a debug log unless explicitly configured."),
	Versions:         handleMediaWikiVersions,
	CloneDBPrefix:    "mediawiki_clone_",
	CloneConfig:      cloneConfig,
	CloneAfterConfig: cloneAddCron,
	Update:           handleMediaWikiUpdate,
	CanClone:         true,
	APICloneForm:     apiCloneForm,
	APIUpdate:        cmsapp.SiteQueryPost("mediawiki", handleMediaWikiUpdate),
	Slug:             "mediawiki",
	Name:             "MediaWiki",
	RemoveConfig:     "LocalSettings.php",
	RemoveDBNameRE:   removeDBNameRE,
	RemoveDBUserRE:   removeDBUserRE,
	OnRemove: func(ctx context.Context, s *cmsapp.Site) {
		_ = crons.RemoveJobByComment(ctx, s.UserContext, mediawikiCronComment(s.Name))
	},
	DBMode: cmsapp.DBOptionalPrefix,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractMediaWikiDatabaseInfoForLogin(userContext, docroot)
	},
	ConfigFile: "LocalSettings.php",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
