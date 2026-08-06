package filemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

const wgetStateDir = "/tmp/filemanager_wget_states"

// downloadState is the shared progress/status record for a wget download,
// persisted to disk so status polling survives across requests/goroutines.
type downloadState struct {
	Progress           int    `json:"progress"`
	Status             string `json:"status"`
	Message            string `json:"message"`
	URL                string `json:"url"`
	ParentPath         string `json:"parent_path"`
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

var extensionSuffixRE = regexp.MustCompile(`\.[A-Za-z0-9]{1,10}$`)

// filenameFromURL derives a safe local filename from a download URL,
// falling back to a generated name when the URL path or query has none.
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

// validateWgetURL only allows http/https, and requires the resolved host
// to not be a private/loopback/link-local/reserved/multicast/unspecified
// address - an SSRF guard against http://localhost:6379/ or
// http://169.254.169.254/latest/meta-data/. It must actually be called
// from the download path, not just defined, or the guard is worthless.
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
	ip := ips[0]
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("Internal IP ranges are not permitted")
	}
	return nil
}

// handleWgetFiles validates the submitted URL and destination, then kicks
// off an asynchronous download and returns its download ID for polling.
func handleWgetFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	// The form is submitted as multipart/form-data (a JS FormData body),
	// not application/x-www-form-urlencoded - ParseForm() alone doesn't
	// read multipart bodies, which left every field empty.
	_ = r.ParseMultipartForm(1 << 20)
	rawURL := strings.TrimSpace(r.Form.Get("url"))
	pathParam := strings.TrimSpace(r.Form.Get("path_param"))

	if rawURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "No URL provided"})
		return
	}
	if verr := validateWgetURL(rawURL); verr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": verr.Error()})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	parentPath, perr := paths.SecureUserPath("HOME", user.Context, pathParam, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"error": msg})
		return
	}
	displayPath := filepath.Join("/var/www/html/", pathParam)

	filename := filenameFromURL(rawURL)
	if _, statErr := os.Stat(filepath.Join(parentPath, filename)); statErr == nil {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "File already exists: " + filename})
		return
	}

	info := downloadState{
		Progress: 0, Status: "queued", Message: "Download queued...",
		URL: rawURL, ParentPath: parentPath, DisplayPath: displayPath,
		Context: user.Context, CurrentUsername: user.Username,
	}
	downloadID := uuid.NewString()
	if err := saveDownloadState(downloadID, info); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error during download."})
		return
	}

	go runWgetWithProgress(a, downloadID)

	writeJSON(w, http.StatusOK, map[string]any{"download_id": downloadID})
}

// runWgetWithProgress runs the actual `wget` download and updates the
// persisted state as it progresses. It runs detached from the triggering
// request (context.Background()) so the download continues even after the
// HTTP response that started it has been sent.
func runWgetWithProgress(a *appctx.App, downloadID string) {
	info, ok := loadDownloadState(downloadID)
	if !ok {
		return
	}

	if err := os.MkdirAll(info.ParentPath, 0o755); err != nil {
		info.Status, info.Message = "error", "Internal server error during download."
		_ = saveDownloadState(downloadID, info)
		return
	}

	filename := filenameFromURL(info.URL)
	downloadedFilePath := filepath.Join(info.ParentPath, filename)
	info.DownloadedFilePath = downloadedFilePath
	info.Status, info.Message, info.Progress = "downloading", "Starting download...", 0
	_ = saveDownloadState(downloadID, info)

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "wget", "--progress=dot:mega", "-O", downloadedFilePath, info.URL)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		info.Status, info.Message = "error", "Internal server error during download."
		_ = saveDownloadState(downloadID, info)
		return
	}
	if startErr := cmd.Start(); startErr != nil {
		info.Status, info.Message = "error", "Internal server error during download."
		_ = saveDownloadState(downloadID, info)
		return
	}

	buf := make([]byte, 4096)
	var lineBuf strings.Builder
	for {
		n, readErr := stderr.Read(buf)
		if n > 0 {
			lineBuf.Write(buf[:n])
			for {
				chunk := lineBuf.String()
				idx := strings.IndexAny(chunk, "\r\n")
				if idx == -1 {
					break
				}
				line := chunk[:idx]
				lineBuf.Reset()
				lineBuf.WriteString(chunk[idx+1:])
				if m := wgetPercentRE.FindStringSubmatch(line); m != nil {
					if pct := atoiDefault(m[1], -1); pct >= 0 {
						cur, _ := loadDownloadState(downloadID)
						cur.Progress, cur.Status, cur.Message = pct, "downloading", "Download in progress..."
						_ = saveDownloadState(downloadID, cur)
					}
				}
			}
		}
		if readErr != nil {
			break
		}
	}

	waitErr := cmd.Wait()
	final, _ := loadDownloadState(downloadID)
	if waitErr == nil {
		final.Progress, final.Status = 100, "done"
		final.Message = "File downloaded from URL successfully to " + final.DisplayPath
		_ = saveDownloadState(downloadID, final)

		if chownErr := chownToUser(ctx, a, downloadedFilePath, final.Context); chownErr != nil {
			cur, _ := loadDownloadState(downloadID)
			cur.Message += " (Failed to set ownership for the downloaded file.)"
			_ = saveDownloadState(downloadID, cur)
		}
	} else {
		final.Status, final.Message = "error", "Failed to download the file"
		_ = saveDownloadState(downloadID, final)
	}
}

var wgetPercentRE = regexp.MustCompile(`(\d+)%`)

// handleWgetStatus reports the current progress/status of a download,
// logging the completed action exactly once the first time it's seen done.
func handleWgetStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	downloadID := r.PathValue("download_id")
	info, ok := loadDownloadState(downloadID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Download not found"})
		return
	}

	if info.Status == "done" && !info.Logged {
		_ = logger.RecordUserAction(a.Config, info.CurrentUsername, "downloaded file from URL "+info.URL+" via File Manager", reqip.ClientIP(r))
		info.Logged = true
		_ = saveDownloadState(downloadID, info)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"progress": info.Progress, "status": info.Status, "message": info.Message,
	})
}
