package backupwizard

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var backupWizardPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/backup_wizard.html",
)

// BackupWizardPageData is backup_wizard.html's template context.
type BackupWizardPageData struct {
	web.LayoutData
	InProgress        bool
	InProgressStarted string
	InProgressSize    string
	Backups           []BackupFile
	BackupIncludes    []string
}

func renderBackupWizardPage(a *appctx.App, w http.ResponseWriter, r *http.Request, inProgress bool, started, size string, backups []BackupFile) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Backup Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	includes := []string{
		layout.T.Get("The home directory"), layout.T.Get("Databases"), layout.T.Get("Domains"),
		layout.T.Get("Websites"), layout.T.Get("Email accounts, filters, aliases"), layout.T.Get("FTP accounts"),
		layout.T.Get("DNS zones"), layout.T.Get("SSL certificates"), layout.T.Get("Cronjobs"),
		layout.T.Get("Containers and images"),
	}
	data := BackupWizardPageData{
		LayoutData: layout, InProgress: inProgress, InProgressStarted: started,
		InProgressSize: size, Backups: backups, BackupIncludes: includes,
	}
	if err := backupWizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BACKUP_WIZARD - template render error: %v", err)
	}
}
