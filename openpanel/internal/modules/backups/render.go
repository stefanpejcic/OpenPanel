package backups

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

var backupsPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/backups.html")...)
var backupSettingsPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/backup_settings.html")...)
var backupDestinationsPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/backup_destinations.html")...)
var backupRestorePage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/backup_restore.html")...)

// BackupsPageData is backups.html's template context.
type BackupsPageData struct {
	web.LayoutData
	Target         string
	HasCredentials bool
	ServiceActive  bool
}

func renderBackupsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, target string, hasCredentials, serviceActive bool) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Backups")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := BackupsPageData{LayoutData: layout, Target: target, HasCredentials: hasCredentials, ServiceActive: serviceActive}
	if err := backupsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BACKUPS - backups template render error: %v", err)
	}
}

// BackupSettingsPageData is backup_settings.html's template context.
type BackupSettingsPageData struct {
	web.LayoutData
	Error    string
	Target   string
	Values   []KV
	Settings []KV
}

func renderBackupSettingsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, errMsg, target string, values, settings []KV) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Settings")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := BackupSettingsPageData{LayoutData: layout, Error: errMsg, Target: target, Values: values, Settings: settings}
	if err := backupSettingsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BACKUPS - settings template render error: %v", err)
	}
}

// DestinationCard is one backup_destinations.html target card.
type DestinationCard struct {
	Target, Title, Description, DocLink string
	Icon                                template.HTML
	Active                              bool
}

var azureIcon = template.HTML(`<svg width="150" height="150" class="size-5" viewBox="0 0 96 96" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="a" x1="-1032.172" x2="-1059.213" y1="145.312" y2="65.426" gradientTransform="matrix(1 0 0 -1 1075 158)" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="#114a8b"/><stop offset="1" stop-color="#0669bc"/></linearGradient><linearGradient id="b" x1="-1027.165" x2="-997.482" y1="147.642" y2="68.561" gradientTransform="matrix(1 0 0 -1 1075 158)" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="#3ccbf4"/><stop offset="1" stop-color="#2892df"/></linearGradient></defs><path fill="url(#a)" d="M33.338 6.544h26.038l-27.03 80.087a4.152 4.152 0 0 1-3.933 2.824H8.149a4.145 4.145 0 0 1-3.928-5.47L29.404 9.368a4.152 4.152 0 0 1 3.934-2.825z"/><path fill="url(#b)" d="M66.595 9.364a4.145 4.145 0 0 0-3.928-2.82H33.648a4.146 4.146 0 0 1 3.928 2.82l25.184 74.62a4.146 4.146 0 0 1-3.928 5.472h29.02a4.146 4.146 0 0 0 3.927-5.472z"/></svg>`) //nolint:gosec // static inline SVG, not user input

var dropboxIcon = template.HTML(`<svg fill="none" viewBox="0 0 165 140" class="size-5" aria-hidden="true"><path fill="#0061FF" d="M82.256 26.215 41.133 52.43 82.256 78.646 41.133 104.861 0 78.498l41.133-26.215L0 26.215 41.133 0l41.123 26.215ZM40.912 113.286l41.133-26.215 41.123 26.215-41.123 26.215L40.912 113.286ZM82.246 78.498l41.133-26.215L82.246 26.215 123.168 0l41.133 26.215-41.133 26.215 41.133 26.215-41.133 26.215L82.246 78.498Z"></path></svg>`) //nolint:gosec // static inline SVG, not user input

type destinationMeta struct {
	Title, Description, DocLink string
	Icon                        template.HTML
}

// BackupDestinationsPageData is backup_destinations.html's template
// context.
type BackupDestinationsPageData struct {
	web.LayoutData
	Active string
	Cards  []DestinationCard
}

func renderBackupDestinationsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, active string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Destination")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	meta := map[string]destinationMeta{
		"s3": {
			Title: layout.T.Get("S3-compatible"), Description: layout.T.Get("AWS S3, Filebase, Google Cloud Storage, MinIO, etc."),
			DocLink: "https://openpanel.com/docs/panel/files/backups/#s3",
		},
		"ssh": {
			Title: layout.T.Get("SSH"), Description: layout.T.Get("Remote backups to another server using via SSH or SFTP."),
			DocLink: "https://openpanel.com/docs/panel/files/backups/#sshsftp",
		},
		"webdav": {
			Title: layout.T.Get("WebDAV"), Description: layout.T.Get("Remote backups using WebDAV protocol."),
			DocLink: "https://openpanel.com/docs/panel/files/backups/#webdav",
		},
		"azure": {
			Title: layout.T.Get("Azure"), Description: layout.T.Get("Remote backups to Azure Blob Storage."),
			DocLink: "https://openpanel.com/docs/panel/files/backups/#azure", Icon: azureIcon,
		},
		"dropbox": {
			Title: layout.T.Get("Dropbox"), Description: layout.T.Get("Remote backups to Dropbox cloud storage."),
			DocLink: "https://openpanel.com/docs/panel/files/backups/#dropbox", Icon: dropboxIcon,
		},
	}

	cards := make([]DestinationCard, 0, len(sectionOrder))
	for _, target := range sectionOrder {
		m := meta[target]
		cards = append(cards, DestinationCard{
			Target: target, Title: m.Title, Description: m.Description, DocLink: m.DocLink,
			Icon: m.Icon, Active: target == active,
		})
	}

	data := BackupDestinationsPageData{LayoutData: layout, Active: active, Cards: cards}
	if err := backupDestinationsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BACKUPS - destinations template render error: %v", err)
	}
}

// RestoreBackupRow is one backup_restore.html table row.
type RestoreBackupRow struct {
	BackupInfo
	JSON template.JS
}

// BackupRestorePageData is backup_restore.html's template context.
type BackupRestorePageData struct {
	web.LayoutData
	Backups      []RestoreBackupRow
	Reindexing   bool
	ReindexError string
}

func renderBackupRestorePage(a *appctx.App, w http.ResponseWriter, r *http.Request, backups []BackupInfo, reindexing bool, reindexError string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Restore")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows := make([]RestoreBackupRow, len(backups))
	for i, b := range backups {
		rows[i] = RestoreBackupRow{BackupInfo: b, JSON: jsonJS(b)}
	}

	data := BackupRestorePageData{LayoutData: layout, Backups: rows, Reindexing: reindexing, ReindexError: reindexError}
	if err := backupRestorePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("BACKUPS - restore template render error: %v", err)
	}
}

func jsonJS(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return template.JS(b) //nolint:gosec // JSON-encoded struct embedded in a <script>/attribute context, not raw HTML
}
