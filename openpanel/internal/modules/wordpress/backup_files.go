package wordpress

import (
	"net/http"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// backupParts maps the download part to the file run_wordpress_backup writes
var backupParts = map[string]struct{ file, contentType string }{
	"files":    {"files.tar.gz", "application/gzip"},
	"database": {"database.sql", "application/sql"},
}

// siteBackupDir is the dated backup folder for site, refused if the cleaned path leaves that site's own backups folder
func siteBackupDir(userContext, site, date string) (string, bool) {
	if !validSiteName(site) || !backupDateRE.MatchString(date) {
		return "", false
	}
	base := filepath.Clean("/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + site)
	dir := filepath.Clean(filepath.Join(base, date))
	if filepath.Dir(dir) != base {
		return "", false
	}
	return dir, true
}

// backupRequest does the ownership and path checks shared by download and delete
func backupRequest(a *appctx.App, w http.ResponseWriter, r *http.Request) (dir, site, username string, ok bool) {
	site = r.PathValue("selected_domain")
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return "", "", "", false
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, strings.Split(site, "/")[0]) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain."})
		return "", "", "", false
	}
	dir, valid := siteBackupDir(userContext, site, r.FormValue("backup_date"))
	if !valid {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid site or backup date."})
		return "", "", "", false
	}
	if info, statErr := tenantfs.Lstat(dir); statErr != nil || !info.IsDir() {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Backup not found."})
		return "", "", "", false
	}
	return dir, site, username, true
}

// handleDownloadBackup streams one file of a WP Manager backup
func handleDownloadBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	part, known := backupParts[r.FormValue("part")]
	if !known {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "part must be files or database"})
		return
	}
	dir, site, username, ok := backupRequest(a, w, r)
	if !ok {
		return
	}
	f, err := tenantfs.Open(filepath.Join(dir, part.file))
	if err != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "This backup has no " + part.file + "."})
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "This backup has no " + part.file + "."})
		return
	}
	name := strings.ReplaceAll(site, "/", "_") + "-" + filepath.Base(dir) + "-" + part.file
	w.Header().Set("Content-Type", part.contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_ = logger.RecordUserAction(a.Config, username, "downloaded WordPress backup "+filepath.Base(dir)+" ("+part.file+") of "+site, reqip.ClientIP(r))
	http.ServeContent(w, r, name, info.ModTime(), f)
}

// handleDeleteBackup removes one dated WP Manager backup folder
func handleDeleteBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	dir, site, username, ok := backupRequest(a, w, r)
	if !ok {
		return
	}
	if err := tenantfs.RemoveAll(dir); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "deleted WordPress backup "+filepath.Base(dir)+" of "+site, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Backup " + filepath.Base(dir) + " deleted."})
}
