package tinyphotogallery

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// This file mirrors drupal/backups.go and flarum/backups.go's directory
// layout and naming (same backups/<domain>/<timestamp>/files.tar.gz
// structure), minus everything database-related - TinyPhotoGallery has no
// database, so there's no database.sql to ever look for here.

func itoa(n int) string { return strconv.Itoa(n) }

var tinyphotogalleryBackupFolderRE = regexp.MustCompile(`^20\d{2}-`)

type tinyphotogalleryBackupDateInfo struct {
	Date           string `json:"date"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

// handleTinyPhotoGalleryGetBackupDates mirrors drupal/backups.go's
// handleDrupalGetBackupDates, minus the database-backup flag.
func handleTinyPhotoGalleryGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	selectedDomain := r.PathValue("selected_domain")

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	backupsPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain
	if _, statErr := os.Stat(backupsPath); statErr != nil {
		if mkErr := os.MkdirAll(backupsPath, 0o755); mkErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
			return
		}
	}

	entries, readErr := os.ReadDir(backupsPath)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	var dates []tinyphotogalleryBackupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !tinyphotogalleryBackupFolderRE.MatchString(entry.Name()) {
			continue
		}
		folderPath := filepath.Join(backupsPath, entry.Name())
		files, _ := os.ReadDir(folderPath)
		info := tinyphotogalleryBackupDateInfo{Date: entry.Name()}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".tar.gz") {
				info.HasFilesBackup = true
			}
		}
		dates = append(dates, info)
	}

	writeJSON(w, http.StatusOK, dates)
}

var tinyphotogalleryBackupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// handleTinyPhotoGalleryRestoreBackup mirrors drupal/backups.go's
// handleDrupalRestoreBackup, files-only.
func handleTinyPhotoGalleryRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	backupDate := r.URL.Query().Get("backup_date")
	docroot := r.URL.Query().Get("docroot")

	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if docroot == "" {
		docroot = "/var/www/html/" + selectedDomain
	}
	if !strings.HasPrefix(docroot, "/var/www/html/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid docroot"})
		return
	}

	if !tinyphotogalleryBackupDateRE.MatchString(backupDate) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
		return
	}

	docrootOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/") + "/"
	backupsPathOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain + "/"
	backupsPathInContainer := "/var/www/html/backups/" + selectedDomain + "/"

	backupDatePathOnHostOS := filepath.Join(backupsPathOnHostOS, backupDate)
	backupDatePathInContainer := filepath.Join(backupsPathInContainer, backupDate)
	targzPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "files.tar.gz")

	if _, statErr := os.Stat(targzPathOnHostOS); statErr != nil {
		_, _ = w.Write([]byte("No files to restore, expected file: " + backupDatePathInContainer + "/files.tar.gz."))
		return
	}

	if runErr := exec.CommandContext(ctx, "tar", "-xzf", targzPathOnHostOS, "-C", docrootOnHostOS).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "restored TinyPhotoGallery files backup from "+backupDatePathInContainer+" on "+selectedDomain, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup restored successfully: files."))
}

// handleTinyPhotoGalleryRunBackup mirrors drupal/backups.go's handleDrupalRunBackup,
// files-only.
func handleTinyPhotoGalleryRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	docroot := r.URL.Query().Get("docroot")
	if docroot == "" {
		http.Error(w, "Document root is not provided or invalid.", http.StatusInternalServerError)
		return
	}

	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.URL.Query().Get("backup_files") != "true" {
		http.Error(w, "No backup options selected.", http.StatusBadRequest)
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	_ = os.MkdirAll(htmlVolume, 0o755)

	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr == nil {
		_ = os.Chown(htmlVolume, uid, uid)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDirectory := filepath.Join(htmlVolume, "backups", selectedDomain, timestamp)
	inPHPBackupDirectory := filepath.Join("/var/www/html/backups", selectedDomain, timestamp)
	if mkErr := os.MkdirAll(backupDirectory, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}
	if uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", itoa(uid)+":"+itoa(uid), backupDirectory).Run()
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpContainer := webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		phpContainer = "php-fpm-" + phpVersion
	}

	tarArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "bash", "-c", "cd "+docroot+" && tar -czf "+inPHPBackupDirectory+"/files.tar.gz .")
	if runErr := podmanmanager.Command(ctx, userContext, tarArgv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "generated files backup for TinyPhotoGallery website "+selectedDomain, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup completed successfully!"))
}
