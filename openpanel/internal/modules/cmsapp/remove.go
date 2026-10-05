package cmsapp

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Site is an installed site being acted on
type Site struct {
	Name         string // "domain.tld" or "domain.tld/sub"
	Docroot      string // the domain's docroot
	InstallPath  string // container path of the install
	HostPath     string // same directory on the host
	UserContext  string
	Username     string
	PHPContainer string
}

// HandleRemove fully uninstalls a site: drops its database and user, removes its files and the sites row
func (app *App) HandleRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = ?`, id, app.Slug)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		web.FlashRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]
	subdirectory := parts[1:]

	if docroot == "" {
		web.FlashRedirect(a, w, r, "error", app.Name+" installation not found in the database", "/sites")
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	realInstallPath := strings.TrimPrefix(docroot, "/var/www/html/")
	if len(subdirectory) > 0 {
		realInstallPath = filepath.Join(append([]string{realInstallPath}, subdirectory...)...)
	}
	installPath := docroot
	if len(subdirectory) > 0 {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + strings.Join(subdirectory, "/")
	}
	s := &Site{
		Name:         siteName,
		Docroot:      docroot,
		InstallPath:  installPath,
		HostPath:     htmlVolume(userContext) + realInstallPath,
		UserContext:  userContext,
		Username:     currentUsername,
		PHPContainer: phpContainerFor(a, r, userContext, selectedDomain),
	}

	switch {
	case app.RemoveDB != nil:
		app.RemoveDB(a, w, r, s)
	case app.RemoveConfig != "":
		app.dropDBFromConfig(a, w, r, s)
	}
	if app.OnRemove != nil {
		app.OnRemove(ctx, s)
	}

	if app.RemoveFiles != nil {
		app.RemoveFiles(ctx, s)
	} else if entries := installmanifest.EntriesViaContainer(ctx, userContext, s.PHPContainer, installPath); entries != nil {
		// remove exactly what the install recorded creating, not the domain's docroot itself
		_ = installmanifest.RemoveViaContainer(ctx, userContext, s.PHPContainer, installPath, entries)
	} else {
		// installs made before the manifest existed
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", s.PHPContainer, "rm", "-rf", installPath)).Run()
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		web.Flash(a, w, r, "error", "An error occurred during "+app.Name+" uninstall.")
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled "+app.Name+" website "+siteName, reqip.ClientIP(r))

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]string{"message": app.Name + " uninstalled successfully"})
		return
	}
	web.FlashRedirect(a, w, r, "success", app.Name+" uninstalled successfully", "/sites")
}

// dropDBFromConfig reads the db name and user out of the site's config file and drops both
func (app *App) dropDBFromConfig(a *appctx.App, w http.ResponseWriter, r *http.Request, s *Site) {
	configFile := filepath.Join(s.HostPath, app.RemoveConfig)
	if app.RemoveConfigPath != nil {
		configFile = app.RemoveConfigPath(s)
	}
	content, _ := os.ReadFile(configFile)
	text := string(content)
	if app.RemoveConfigText != nil {
		text = app.RemoveConfigText(text)
	}

	dbNameMatch := app.RemoveDBNameRE.FindStringSubmatch(text)
	dbUserMatch := app.RemoveDBUserRE.FindStringSubmatch(text)
	if dbNameMatch == nil || dbUserMatch == nil {
		web.Flash(a, w, r, "warning", "Database name or user not found in "+app.RemoveConfig)
		return
	}
	ctx := r.Context()
	dbName, dbUser := dbNameMatch[1], dbUserMatch[1]
	_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
	_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
	appkit.InvalidateMySQLCaches(ctx, a, s.UserContext, s.Username)
}
