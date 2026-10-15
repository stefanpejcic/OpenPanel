package backups

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var backupsPage = web.LoadPage("files/backups.html")
var backupSettingsPage = web.LoadPage("files/backup_settings.html")
var backupDestinationsPage = web.LoadPage("files/backup_destinations.html")
var backupRestorePage = web.LoadPage("files/backup_restore.html")

// BackupsPageData is backups.html's template context.
type BackupsPageData struct {
	web.LayoutData
	Target         string
	HasCredentials bool
	ServiceActive  bool
	AdminManaged   bool
}

func renderBackupsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, target string, hasCredentials, serviceActive, adminManaged bool) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Backups")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := BackupsPageData{LayoutData: layout, Target: target, HasCredentials: hasCredentials, ServiceActive: serviceActive, AdminManaged: adminManaged}
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
	// errMsg doubles as a sentinel in backups.go, so only translate it for display
	displayErr := map[string]string{
		"no backup target configured":        layout.T.Get("No backup target is configured."),
		"multiple backup targets configured": layout.T.Get("Multiple backup targets are configured."),
	}[errMsg]
	if displayErr == "" {
		displayErr = errMsg
	}
	data := BackupSettingsPageData{LayoutData: layout, Error: displayErr, Target: target, Values: values, Settings: settings}
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

// BackupDestinationsPageData is backup_destinations.html's template context.
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
			DocLink: "https://openpanel.com/docs/panel/backups/#s3",
		},
		"ssh": {
			Title: layout.T.Get("SSH"), Description: layout.T.Get("Remote backups to another server using via SSH or SFTP."),
			DocLink: "https://openpanel.com/docs/panel/backups/#sshsftp",
		},
		"webdav": {
			Title: layout.T.Get("WebDAV"), Description: layout.T.Get("Remote backups using WebDAV protocol."),
			DocLink: "https://openpanel.com/docs/panel/backups/#webdav",
		},
		"azure": {
			Title: layout.T.Get("Azure"), Description: layout.T.Get("Remote backups to Azure Blob Storage."),
			DocLink: "https://openpanel.com/docs/panel/backups/#azure", Icon: azureIcon,
		},
		"dropbox": {
			Title: layout.T.Get("Dropbox"), Description: layout.T.Get("Remote backups to Dropbox cloud storage."),
			DocLink: "https://openpanel.com/docs/panel/backups/#dropbox", Icon: dropboxIcon,
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

// restoreItem is one backup as the restore page's Alpine code sees it
type restoreItem struct {
	File      string      `json:"file"`
	Created   string      `json:"created,omitempty"`
	Size      int64       `json:"size"`
	SizeText  string      `json:"size_text,omitempty"`
	Types     []string    `json:"types"`
	Databases []restoreDB `json:"databases"`
	Lists     []string    `json:"lists"`
	// small per-backup lists read from the reindex cache, the website file list stays behind /backups/files
	Mailboxes []backupItem `json:"mailboxes"`
	Domains   []backupItem `json:"domains"`
	Zones     []backupItem `json:"zones"`
	Error     string       `json:"error,omitempty"`
}

type restoreDB struct {
	Name     string `json:"name"`
	Engine   string `json:"engine,omitempty"`
	SizeText string `json:"size_text,omitempty"`
}

// docker-volume-backup names archives <user>-%Y-%m-%dT%H-%M-%S, in the container's UTC
var archiveTimeRE = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2})-(\d{2})-(\d{2})`)

func archiveCreated(name string) string {
	m := archiveTimeRE.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	t, err := time.Parse("2006-01-02T15:04:05", m[1]+"T"+m[2]+":"+m[3]+":"+m[4])
	if err != nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// restoreItems turns the index into the page's data, newest backup first, userHome "" skips the mailbox/domain/zone lists
func restoreItems(backups []BackupInfo, userHome string, domains map[string]bool) []restoreItem {
	items := make([]restoreItem, 0, len(backups))
	for _, b := range backups {
		item := restoreItem{File: b.BackupFile, Created: archiveCreated(b.BackupFile), Size: b.Size, Types: b.Types, Databases: []restoreDB{}, Lists: b.Lists, Error: b.Error}
		if item.Lists == nil {
			item.Lists = []string{}
		}
		item.Mailboxes, item.Domains, item.Zones = []backupItem{}, []backupItem{}, []backupItem{}
		for _, kind := range item.Lists {
			if userHome == "" || kind == "files" {
				continue
			}
			list, err := readItems(listPath(userHome, b.BackupFile, kind), kind, domains)
			if err != nil {
				continue
			}
			switch kind {
			case "mail":
				item.Mailboxes = list
			case "vhosts":
				item.Domains = list
			case "dns":
				item.Zones = list
			}
		}
		if item.Types == nil {
			item.Types = []string{}
		}
		if b.Size > 0 {
			item.SizeText = formatSize(b.Size)
		}
		for _, db := range b.Databases {
			row := restoreDB{Name: db, Engine: b.DatabaseEngines[db]}
			if size, ok := b.DatabaseSizes[db]; ok {
				row.SizeText = formatSize(size)
			}
			item.Databases = append(item.Databases, row)
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Created != items[j].Created {
			return items[i].Created > items[j].Created
		}
		return items[i].File > items[j].File
	})
	return items
}

func formatSize(numBytes int64) string {
	size := float64(numBytes)
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if size < 1024.0 {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
		size /= 1024.0
	}
	return fmt.Sprintf("%.1f TB", size)
}

// BackupRestorePageData is backup_restore.html's template context.
type BackupRestorePageData struct {
	web.LayoutData
	Backups      []restoreItem
	PageJSON     template.JS
	Reindexing   bool
	ReindexError string
}

func renderBackupRestorePage(a *appctx.App, w http.ResponseWriter, r *http.Request, backups []BackupInfo, reindexing bool, reindexError string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Restore")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userHome, domains := "", map[string]bool{}
	if userID, _, userContext, injErr := auth.Injected(a, r); injErr == nil {
		userHome = "/home/" + userContext
		if plan, planErr := newRestorePlan(r.Context(), a, userID, userContext); planErr == nil {
			domains = plan.domains
		}
	}
	items := restoreItems(backups, userHome, domains)
	data := BackupRestorePageData{LayoutData: layout, Backups: items, PageJSON: jsonJS(items), Reindexing: reindexing, ReindexError: reindexError}
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
