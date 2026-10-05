package cmsapp

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Clone is one in-progress clone, handed to the app's clone hooks
type Clone struct {
	Ctx                 context.Context
	R                   *http.Request
	UserContext         string
	MySQLVersion        string
	ProvidedDomain      string // source site name, "domain.tld/sub"
	DstDomain           string
	DstFolder           string // destination subdirectory, "" for the domain root
	DstDomainWithSubdir string
	Docroot             string // destination install path in the container
	PHPVersion          string // destination domain's php version
	DstPath             string // destination install path on the host
	DstDB               string
	DstDBUser           string
	DstDBUserPassword   string
}

// RewriteConfig applies edit to a config file under the clone, then checks it now mentions the new db, failStep is the error step reported otherwise
func (c *Clone) RewriteConfig(relPath string, edit func(string) string, failStep string) map[string]any {
	configFile := filepath.Join(c.DstPath, relPath)
	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		return map[string]any{"status": "error", "details": readErr.Error()}
	}
	strContent := edit(string(content))
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		return map[string]any{"status": "error", "details": writeErr.Error()}
	}
	if !strings.Contains(strContent, c.DstDB) {
		return map[string]any{"status": "error", "step": failStep}
	}
	return nil
}

// HandleClone copies a site's files (and database, unless the app has none) to another domain or subdirectory
func (app *App) HandleClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	websiteCount, _ := appkit.CountUserWebsites(a, userID)
	if !cmsclone.WithinSiteLimit(ctx, a, userID, websiteCount) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed" + a.UpgradeMessageForUser(ctx, userID)})
		return
	}

	hasDB := app.DBMode != DBNone
	providedDomain := r.FormValue("source_domain")
	dstDomain := r.FormValue("target_domain")
	srcFolder := r.FormValue("source_folder")
	dstFolder := r.FormValue("subdirectory")

	c := &Clone{Ctx: ctx, R: r, UserContext: userContext, ProvidedDomain: providedDomain, DstDomain: dstDomain, DstFolder: dstFolder}
	var srcDB string
	if hasDB {
		srcDB = r.FormValue("source_db")
		c.DstDB = strings.ToLower(web.FormOr(r, "target_db", app.CloneDBPrefix+appkit.RandomString(6)))
		c.DstDBUser = strings.ToLower(web.FormOr(r, "target_db_user", c.DstDB))
		c.DstDBUserPassword = web.FormOr(r, "target_db_user_password", appkit.RandomString(16))
	}

	if providedDomain == "" || dstDomain == "" || (hasDB && srcDB == "") || srcFolder == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required form fields"})
		return
	}

	domainID, docroot, phpVersion, dstDomainWithSubdir, ok := cmsclone.ResolveDestination(ctx, a, dstDomain, dstFolder)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination domain not found in database"})
		return
	}
	c.Docroot, c.PHPVersion, c.DstDomainWithSubdir = docroot, phpVersion, dstDomainWithSubdir

	srcDomain := strings.Split(providedDomain, "/")[0]

	valid := cmsclone.ValidDomain(srcDomain) && cmsclone.ValidDomain(dstDomain) && cmsclone.ValidDocroot(srcFolder) && cmsclone.ValidDocroot(docroot)
	if hasDB {
		valid = valid && cmsclone.ValidDB(srcDB) && cmsclone.ValidDB(c.DstDB) && cmsclone.ValidDB(c.DstDBUser)
	}
	if !valid {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input or unsafe docroot"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, srcDomain) || !a.CheckDomainBelongsToUser(ctx, userID, dstDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	var dumpCmd string
	if hasDB {
		var dumpCmdErr error
		dumpCmd, c.MySQLVersion, dumpCmdErr = cmsclone.SelectDumpCommand(userContext)
		if dumpCmdErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": dumpCmdErr.Error()})
			return
		}
	}

	const wwwBaseDirectory = "/var/www/html/"
	baseDirectory := htmlVolume(userContext)
	srcPath := strings.Replace(filepath.Clean(srcFolder), wwwBaseDirectory, baseDirectory, 1)
	c.DstPath = strings.Replace(filepath.Clean(docroot), wwwBaseDirectory, baseDirectory, 1)

	if info, statErr := os.Stat(srcPath); statErr != nil || !info.IsDir() {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Source folder not found: " + srcFolder})
		return
	}
	if mkErr := os.MkdirAll(c.DstPath, 0o755); mkErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy " + app.Name + " files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", c.DstPath+"/").Run(); cpErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy " + app.Name + " files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, c.DstPath)
	if app.CloneAfterCopy != nil {
		app.CloneAfterCopy(c)
	}

	if hasDB {
		_, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, c.MySQLVersion, dumpCmd, srcDB, c.DstDB, c.DstDBUser, c.DstDBUserPassword)
		if dbErr != nil {
			if cmsclone.DumpStageFailed(dbErr) {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
			} else {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
			}
			return
		}
	}

	if app.CloneConfig != nil {
		if errBody := app.CloneConfig(c); errBody != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errBody)
			return
		}
	}
	if app.CloneAfterConfig != nil {
		app.CloneAfterConfig(c)
	}

	adminEmail := web.FormOr(r, "admin_email", "admin@"+dstDomain)
	version := web.FormOr(r, app.Slug+"_version", "latest")
	if app.CloneVersion != nil {
		version = app.CloneVersion(c)
	}
	if hasDB && !app.CloneNoSearchReplace {
		// rewrites hardcoded source-domain URLs left in content, the generic equivalent of wp-cli's search-replace
		cmsclone.SearchReplaceDatabase(ctx, userContext, c.DstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)
	}

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: web.WriteJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: app.Name, CMSType: app.Slug,
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: version,
		SrcPath: srcPath, DstPath: c.DstPath, DstDB: c.DstDB,
	})
}

// HandleAPIClone runs the clone with the form the app's APICloneForm builds from the API request
func (app *App) HandleAPIClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	form, ok := app.APICloneForm(a, w, r)
	if !ok {
		return
	}
	app.cloner()(a, w, web.WithForm(r, form))
}
