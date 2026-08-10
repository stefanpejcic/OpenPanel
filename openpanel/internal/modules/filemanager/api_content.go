package filemanager

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// apiGetFileContent is handleEditFile's GET branch (editor.go), reused via
// forceJSONOutput so it already returns the raw content as a bare JSON
// string (matching the web editor's own "?output=json" AJAX contract) -
// same allowlist/binary-sniff/size-limit checks apply.
func apiGetFileContent(a *appctx.App, w http.ResponseWriter, r *http.Request, filePath string) {
	handleEditFile(a, w, forceJSONOutput(r), filePath)
}

// apiSaveFileContent is handleEditFile's POST branch (editor.go) with a
// JSON body/response instead of a form post + flash/redirect.
func apiSaveFileContent(a *appctx.App, w http.ResponseWriter, r *http.Request, filePath string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if strings.HasSuffix(filePath, "wp_temp_openpanel-config.php") {
		filePath = strings.Replace(filePath, "wp_temp_openpanel-config.php", "wp-config.php", 1)
	}

	dottedExts, bareNames := splitExtNames(editExtensions(a))
	fileExt := strings.ToLower(filepath.Ext(filePath))
	filename := filepath.Base(filePath)

	realPath, perr := paths.SecureUserPath("HOME", user.Context, filePath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	if !containsString(dottedExts, fileExt) && !containsAny(strings.ToLower(filename), bareNames) {
		if !isEditableFile(realPath) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Editing this file type is not allowed."})
			return
		}
	}

	var body struct {
		Content string `json:"content"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Content = r.Form.Get("editor_content")
	}

	contentToWrite := body.Content
	if containsString(unixLineEndingExts, fileExt) || containsString(unixLineEndingExts, strings.ToLower(filename)) {
		contentToWrite = strings.ReplaceAll(contentToWrite, "\r\n", "\n")
		contentToWrite = strings.ReplaceAll(contentToWrite, "\r", "\n")
		contentToWrite = strings.TrimPrefix(contentToWrite, "\ufeff")
	}

	tmpPath := realPath + ".tmp"
	writeErr := os.WriteFile(tmpPath, []byte(contentToWrite), 0o644)
	if writeErr == nil {
		writeErr = os.Rename(tmpPath, realPath)
	}
	if writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving file."})
		return
	}

	_ = chownToUser(ctx, a, realPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "edited file "+filePath+" via API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "File saved successfully"})
}

// apiDownloadFile is handleDownloadFile (editor.go) with JSON error
// responses instead of flash+redirect. The success path (streaming the
// file) is unchanged.
func apiDownloadFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	pathParam := r.URL.Query().Get("path_param")

	realPath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, filename), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	info, statErr := os.Stat(realPath)
	if statErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Error retrieving file size."})
		return
	}
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_download_size", "500"), 500)
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "File size exceeds " + strconv.Itoa(fileLimitMB) + "MB limit. Download aborted."})
		return
	}

	f, openErr := os.Open(realPath)
	if openErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Error retrieving file size."})
		return
	}
	defer f.Close()

	_ = logger.RecordUserAction(a.Config, user.Username, "downloaded "+filepath.Join(pathParam, filename)+" via File Manager API", reqip.ClientIP(r))

	maxDuration := time.Duration(atoiDefault(a.Config.Get("filemanager_download_max_time", "60"), 60)) * time.Minute
	deadline := time.Now().Add(maxDuration)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(realPath)+`"`)

	buf := make([]byte, 8192)
	for {
		if time.Now().After(deadline) {
			return
		}
		n, readErr := f.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
		}
		if readErr == io.EOF || readErr != nil {
			return
		}
	}
}

// apiViewFile is handleViewFile (editor.go) with JSON error responses
// instead of flash+redirect. The success path (serving raw content) is
// unchanged.
func apiViewFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if strings.HasSuffix(filename, "wp_temp_openpanel-config.php") {
		filename = strings.Replace(filename, "wp_temp_openpanel-config.php", "wp-config.php", 1)
	}

	editExts := editExtensions(a)
	imageExts := imageExtensions(a)
	dottedExts, _ := splitExtNames(append(append([]string{}, editExts...), imageExts...))
	_, editBareNames := splitExtNames(editExts)
	fileExt := strings.ToLower(filepath.Ext(filename))
	baseName := filepath.Base(filename)

	if !containsString(dottedExts, fileExt) && !containsAny(strings.ToLower(baseName), editBareNames) {
		realPathCheck, perr := paths.SecureUserPath("HOME", user.Context, filename, false)
		if perr == nil && !isEditableFile(realPathCheck) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Viewing this file type is not allowed."})
			return
		}
	}

	realPath, perr := paths.SecureUserPath("HOME", user.Context, filename, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	fileLimitMB := atoiDefault(a.Config.Get("filemanager_view_size", "500"), 500)
	info, statErr := os.Stat(realPath)
	if statErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Error accessing file size."})
		return
	}
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "File is too large to view (limit is " + strconv.Itoa(fileLimitMB) + " MB)"})
		return
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if strings.HasPrefix(mimeType, "image/") {
		http.ServeFile(w, r, realPath)
		return
	}

	content, readErr := os.ReadFile(realPath)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error opening file."})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}
