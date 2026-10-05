// Package cmsapp holds what every one-click PHP CMS module (joomla, drupal, nextcloud, ...) shares, each module just describes itself with an App
package cmsapp

import (
	"context"
	"net/http"
	"net/url"
	"regexp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// DBMode says how backups pick the site's tables
type DBMode int

const (
	// DBPrefixed needs both db name and table prefix, only prefix% tables are dumped/dropped
	DBPrefixed DBMode = iota
	// DBOptionalPrefix needs only the db name, an empty prefix means every table in it
	DBOptionalPrefix
	// DBWhole dumps the whole database and imports over it without dropping tables first
	DBWhole
	// DBNone is a files-only app with no database
	DBNone
)

// App describes one CMS module, build it with New
type App struct {
	Slug string // url prefix, feature flag and sites.type
	Name string // display name used in messages and activity log

	// Install streams the NDJSON install for a POST to /<slug>/install
	Install Handler
	// InstallFields are the API's JSON body keys passed on as install form fields
	InstallFields []string

	// UI routes, nil ones aren't registered, Remove defaults to HandleRemove
	Remove, Login, Cache, Logs, Versions, Maintenance, Clone, Update Handler
	// API routes besides install/remove, nil ones aren't registered
	APIClone, APIUpdate, APICache Handler
	// APIClonePath overrides the default "POST /api/<slug>/sites/{site_id}/clone"
	APIClonePath string

	DBMode DBMode
	// DBInfo returns "database_name" and "database_prefix" read from the site's config
	DBInfo func(userContext, docroot, selectedDomain string) map[string]string
	// ConfigFile is named in the "Failed to determine database name" error
	ConfigFile string

	// FilesDir is the container dir backups tar instead of docroot, for apps whose content lives outside it (moodledata)
	FilesDir   func(selectedDomain string) string
	FilesLabel string // "files" unless set, e.g. "files (moodledata)"
	// ChownAfterRestore re-owns restored files to the user, host-side tar leaves them as nobody:nogroup
	ChownAfterRestore bool

	// RemoveConfig is the config file under the install holding the db name and user, also named in the "not found" warning
	RemoveConfig                   string
	RemoveDBNameRE, RemoveDBUserRE *regexp.Regexp
	// RemoveConfigPath overrides where RemoveConfig is read from
	RemoveConfigPath func(s *Site) string
	// RemoveConfigText filters the config before matching, drupal strips doc comments
	RemoveConfigText func(string) string
	// RemoveDB replaces the config-regex db drop
	RemoveDB func(a *appctx.App, w http.ResponseWriter, r *http.Request, s *Site)
	// OnRemove runs after the db is dropped, before the files go
	OnRemove func(ctx context.Context, s *Site)
	// RemoveFiles replaces the install-manifest file removal
	RemoveFiles func(ctx context.Context, s *Site)

	// CanClone turns on the shared HandleClone for apps that don't set their own Clone
	CanClone bool
	// APICloneForm turns an API clone request into HandleClone's form fields, writing the error itself when it returns false
	APICloneForm func(a *appctx.App, w http.ResponseWriter, r *http.Request) (url.Values, bool)
	// CloneDBPrefix starts the default target db name, e.g. "joomla_clone_"
	CloneDBPrefix string
	// CloneConfig points the copied site's config at its new db and domain, a non-nil result is sent back as a 500
	CloneConfig func(c *Clone) map[string]any
	// CloneAfterCopy runs right after the files are copied, CloneAfterConfig after the config is rewritten
	CloneAfterCopy, CloneAfterConfig func(c *Clone)
	// CloneVersion overrides the version recorded for the clone, default is the <slug>_version form field or "latest"
	CloneVersion func(c *Clone) string
	// CloneNoSearchReplace skips rewriting source-domain URLs in the cloned database
	CloneNoSearchReplace bool

	installPage *web.Page
}

func New(app App) *App {
	app.installPage = web.LoadPage("manager/_cms_install_shared.html", "manager/"+app.Slug+"_install.html")
	return &app
}

func (app *App) filesLabel() string {
	if app.FilesLabel != "" {
		return app.FilesLabel
	}
	return "files"
}
