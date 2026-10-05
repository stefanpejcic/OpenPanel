// Package dokuwiki installs and manages DokuWiki inside an existing domain's docroot, same shape as sofawiki (the other flat-file module), but it ships real dated releases (so Update is feasible), needs PHP 7.4+, and gets fully configured at install time - install.go writes conf/local.php, users.auth.php and acl.auth.php directly instead of running install.php, then deletes install.php same as DokuWiki's own docs recommend
package dokuwiki

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "admin_email", "admin_user", "admin_password", "admin_full_name", "site_title"},
	CloneAfterCopy: cloneSetTitle,
	CloneVersion:   cloneVersion,
	Update:         handleDokuwikiUpdate,
	CanClone:       true,
	APICloneForm:   apiCloneForm,
	APIUpdate:      cmsapp.SiteQueryPost("dokuwiki", handleDokuwikiUpdate),
	Slug:           "dokuwiki",
	Name:           "DokuWiki",
	DBMode:         cmsapp.DBNone,
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
