package filemanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var wgetStateDir = "/tmp/filemanager_wget_states"

// state files left this long are from finished or abandoned downloads, running ones rewrite theirs on every progress update
const wgetStateMaxAge = 24 * time.Hour

// wgetLogMu makes "log the finished download once" a single step across concurrent status polls
var wgetLogMu sync.Mutex

// downloadState is the shared progress/status record for a wget download, persisted to disk so status polling survives across requests/goroutines
type downloadState struct {
	Progress           int    `json:"progress"`
	Status             string `json:"status"`
	Message            string `json:"message"`
	URL                string `json:"url"`
	ParentRel          string `json:"parent_rel"`
	DisplayPath        string `json:"display_path"`
	Context            string `json:"context"`
	CurrentUsername    string `json:"current_username"`
	DownloadedFilePath string `json:"downloaded_file_path"`
	Logged             bool   `json:"logged"`
}

func stateFile(downloadID string) (string, error) {
	if err := os.MkdirAll(wgetStateDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(wgetStateDir, downloadID+".json"), nil
}

func saveDownloadState(downloadID string, s downloadState) error {
	path, err := stateFile(downloadID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func loadDownloadState(downloadID string) (downloadState, bool) {
	path, err := stateFile(downloadID)
	if err != nil {
		return downloadState{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return downloadState{}, false
	}
	var s downloadState
	if json.Unmarshal(data, &s) != nil {
		return downloadState{}, false
	}
	return s, true
}

// loadDownloadStateFor is loadDownloadState for the account that started the download, anyone else gets not found
func loadDownloadStateFor(downloadID, userContext string) (downloadState, bool) {
	s, ok := loadDownloadState(downloadID)
	if !ok || s.Context == "" || s.Context != userContext {
		return downloadState{}, false
	}
	return s, true
}

// claimDoneLog marks a finished download as logged and reports true only to the first caller
func claimDoneLog(downloadID string) (downloadState, bool) {
	wgetLogMu.Lock()
	defer wgetLogMu.Unlock()
	s, ok := loadDownloadState(downloadID)
	if !ok || s.Status != "done" || s.Logged {
		return s, false
	}
	s.Logged = true
	_ = saveDownloadState(downloadID, s)
	return s, true
}

// pruneDownloadStates deletes state files older than maxAge, nothing else ever removed them
func pruneDownloadStates(maxAge time.Duration) {
	entries, err := os.ReadDir(wgetStateDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.Mode().IsRegular() && time.Since(info.ModTime()) > maxAge {
			_ = os.Remove(filepath.Join(wgetStateDir, e.Name()))
		}
	}
}

var extensionSuffixRE = regexp.MustCompile(`\.[A-Za-z0-9]{1,10}$`)

// filenameFromURL derives a safe local filename from a download URL, falling back to a generated name when the URL path or query has none
func filenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("downloaded_%s", uuid.NewString()[:8])
	}
	pathName := filepath.Base(strings.TrimRight(parsed.Path, "/"))

	var name string
	if extensionSuffixRE.MatchString(pathName) {
		name = pathName
	} else if qf := parsed.Query().Get("filename"); qf != "" {
		name = qf
	} else {
		name = fmt.Sprintf("downloaded_%s", uuid.NewString()[:8])
	}

	name = secureFilename(name)
	if name == "" {
		name = fmt.Sprintf("downloaded_%s", uuid.NewString()[:8])
	}
	return name
}

// validateWgetURL only allows http/https, and requires the resolved host to not be a private/loopback/link-local/reserved/multicast/unspecified address - an SSRF guard against http://localhost:6379/ or the AWS metadata endpoint, worthless unless actually called from the download path
func validateWgetURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("Only http/https URLs are allowed")
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("No host in URL")
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return fmt.Errorf("Could not resolve host")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return fmt.Errorf("Internal IP ranges are not permitted")
		}
	}
	return nil
}

// handleWgetFiles validates the submitted URL and destination, then kicks off an asynchronous download and returns its download ID for polling
func handleWgetFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	// the form is submitted as multipart/form-data (a JS FormData body), not urlencoded - ParseForm() alone doesn't read multipart bodies, which left every field empty
	_ = r.ParseMultipartForm(1 << 20)
	rawURL := strings.TrimSpace(r.Form.Get("url"))
	pathParam := strings.TrimSpace(r.Form.Get("path_param"))

	if rawURL == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "No URL provided"})
		return
	}
	if verr := validateWgetURL(rawURL); verr != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": verr.Error()})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, parentRel, perr := openUserPath(user, pathParam, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"error": msg})
		return
	}
	defer root.Close()
	displayPath := filepath.Join("/var/www/html/", pathParam)

	filename := filenameFromURL(rawURL)
	if _, statErr := root.Lstat(path.Join(parentRel, filename)); statErr == nil {
		web.WriteJSON(w, http.StatusConflict, map[string]any{"error": "File already exists: " + filename})
		return
	}

	info := downloadState{
		Progress: 0, Status: "queued", Message: "Download queued...",
		URL: rawURL, ParentRel: parentRel, DisplayPath: displayPath,
		Context: user.Context, CurrentUsername: user.Username,
	}
	pruneDownloadStates(wgetStateMaxAge)
	downloadID := uuid.NewString()
	if err := saveDownloadState(downloadID, info); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error during download."})
		return
	}

	go runWgetWithProgress(a, downloadID)

	web.WriteJSON(w, http.StatusOK, map[string]any{"download_id": downloadID})
}

// runWgetWithProgress downloads the URL in the background with the SSRF-safe client, it runs detached from the request so the download continues after the response is sent
func runWgetWithProgress(a *appctx.App, downloadID string) {
	info, ok := loadDownloadState(downloadID)
	if !ok {
		return
	}
	fail := func(msg string) {
		cur, _ := loadDownloadState(downloadID)
		cur.Status, cur.Message = "error", msg
		_ = saveDownloadState(downloadID, cur)
	}

	// reopen and recheck, the dir could have been swapped since the request was validated
	root, err := paths.OpenUserRoot("HOME", info.Context)
	if err != nil {
		fail("Internal server error during download.")
		return
	}
	defer root.Close()
	parentRel, err := root.Rel(info.ParentRel, false)
	if err != nil {
		fail("Internal server error during download.")
		return
	}
	if err := root.MkdirAll(parentRel, 0o755); err != nil {
		fail("Internal server error during download.")
		return
	}

	filename := filenameFromURL(info.URL)
	downloadedRel := path.Join(parentRel, filename)
	info.DownloadedFilePath = root.Abs(downloadedRel)
	info.Status, info.Message, info.Progress = "downloading", "Starting download...", 0
	_ = saveDownloadState(downloadID, info)

	maxTime := time.Duration(atoiDefault(a.Config.Get("filemanager_download_max_time", "60"), 60)) * time.Minute
	maxBytes := int64(uploadSizeMB(a)) << 20
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()

	lastPct := -1
	err = downloadToFile(ctx, root.Root, info.URL, downloadedRel, maxBytes, func(pct int) {
		if pct == lastPct {
			return
		}
		lastPct = pct
		cur, _ := loadDownloadState(downloadID)
		cur.Progress, cur.Status, cur.Message = pct, "downloading", "Download in progress..."
		_ = saveDownloadState(downloadID, cur)
	})
	if err != nil {
		log.Printf("FILEMANAGER - download of %s failed: %v", info.URL, err)
		msg := "Failed to download the file"
		switch {
		case errors.Is(err, errDownloadTooLarge):
			msg = fmt.Sprintf("Failed to download the file: it is larger than %d MB", maxBytes>>20)
		case errors.Is(err, errBlockedAddress):
			msg = "Failed to download the file: internal IP ranges are not permitted"
		case errors.Is(err, context.DeadlineExceeded):
			msg = "Failed to download the file: it took too long"
		}
		fail(msg)
		return
	}

	final, _ := loadDownloadState(downloadID)
	final.Progress, final.Status = 100, "done"
	final.Message = "File downloaded from URL successfully to " + final.DisplayPath
	if chownErr := chownToUser(ctx, a, root, downloadedRel, final.Context); chownErr != nil {
		final.Message += " (Failed to set ownership for the downloaded file.)"
	}
	_ = saveDownloadState(downloadID, final)
}

// handleWgetStatus reports the current progress/status of a download, logging the completed action exactly once the first time it's seen done
func handleWgetStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	downloadID := r.PathValue("download_id")
	user, err := currentUser(r.Context(), a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	info, ok := loadDownloadStateFor(downloadID, user.Context)
	if !ok {
		web.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "Download not found"})
		return
	}

	if info.Status == "done" && !info.Logged {
		if claimed, first := claimDoneLog(downloadID); first {
			_ = logger.RecordUserAction(a.Config, claimed.CurrentUsername, "downloaded file from URL "+claimed.URL+" via File Manager", reqip.ClientIP(r))
		}
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{
		"progress": info.Progress, "status": info.Status, "message": info.Message,
	})
}
