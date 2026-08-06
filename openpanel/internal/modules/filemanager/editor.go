package filemanager

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

var allowedEditors = map[string]bool{"monaco": true, "ace": true, "codemirror": true, "text": true}

// isEditableFile is a crude binary-file sniff: no NUL byte in the first
// 8KB.
func isEditableFile(realPath string) bool {
	f, err := os.Open(realPath)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return false
		}
	}
	return true
}

func editExtensions(a *appctx.App) []string {
	raw := stripQuotes(a.Config.Get("filemanager_edit_extensions",
		".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf"))
	return strings.Fields(raw)
}

func imageExtensions(a *appctx.App) []string {
	raw := stripQuotes(a.Config.Get("filemanager_image_extensions", ".jpg .jpeg .png .gif .webp .avif"))
	return strings.Fields(raw)
}

// splitExtNames splits a config extension list into (dotted extensions,
// bare filename substrings) depending on whether each entry starts with
// a dot.
func splitExtNames(exts []string) (dotted, bare []string) {
	for _, e := range exts {
		lower := strings.ToLower(e)
		if strings.HasPrefix(e, ".") {
			dotted = append(dotted, lower)
		} else {
			bare = append(bare, lower)
		}
	}
	return dotted, bare
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// unixLineEndingExts lists file types that get \r\n -> \n normalization
// and a stripped BOM on save.
var unixLineEndingExts = []string{
	".sh", ".bash", ".zsh", ".py", ".conf", ".service", ".env", ".ini", ".cfg", ".json", ".php",
	".html", ".htm", ".css", ".js", ".sql", ".xml", ".txt", ".md", ".htaccess", ".log", "error_log",
}

// handleEditFile serves the file editor page on GET and saves the
// submitted content on POST.
func handleEditFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filePath string) {
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
		http.Error(w, msg, status)
		return
	}

	if !containsString(dottedExts, fileExt) && !containsAny(strings.ToLower(filename), bareNames) {
		if !isEditableFile(realPath) {
			flashAndRedirect(a, w, r, "error", "Editing this file type is not allowed.", "/files")
			return
		}
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		editedContent := r.Form.Get("editor_content")

		contentToWrite := editedContent
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

		sess, _ := a.Sessions.Get(r, session.CookieName)
		if writeErr != nil {
			flash.Add(sess, "error", "Error saving file.")
		} else {
			_ = chownToUser(ctx, a, realPath, user.Context)
			flash.Add(sess, "success", "File saved successfully")
			_ = logger.RecordUserAction(a.Config, user.Username, "edited file "+filePath, reqip.ClientIP(r))
		}
		_ = a.Sessions.Save(r, w, sess)

		http.Redirect(w, r, "/file-manager/edit-file/"+filePath, http.StatusFound)
		return
	}

	// GET
	isNew := r.URL.Query().Get("new") == "true"
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_editor_size", "5"), 5)
	fileLimitBytes := int64(fileLimitMB) * 1024 * 1024
	var fileContent string

	if !isNew {
		info, statErr := os.Stat(realPath)
		if statErr != nil {
			flashAndRedirect(a, w, r, "error", "Error accessing file size.", "/files")
			return
		}
		if info.Size() > fileLimitBytes {
			flashAndRedirect(a, w, r, "error",
				"File is too large to open in the editor (limit is "+strconv.Itoa(fileLimitMB)+" MB)", "/files")
			return
		}
	}

	data, readErr := os.ReadFile(realPath)
	if readErr != nil {
		if !isNew {
			flashAndRedirect(a, w, r, "error", "Error reading file", "/files")
			return
		}
	} else {
		fileContent = string(data)
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, fileContent)
		return
	}

	editor := r.URL.Query().Get("editor")
	if editor == "" {
		editor = "monaco"
	}
	if !allowedEditors[editor] {
		editor = "monaco"
	}

	renderEditFilePage(a, w, r, filePath, fileContent, editor)
}

// handleDownloadFile streams the file to the client, aborting once a
// wall-clock time limit is exceeded so a stalled connection can't hold
// the handler open indefinitely.
func handleDownloadFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
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
		http.Error(w, msg, status)
		return
	}

	info, statErr := os.Stat(realPath)
	if statErr != nil {
		flashAndRedirect(a, w, r, "error", "Error retrieving file size.", filesRedirectPath(pathParam))
		return
	}
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_download_size", "500"), 500)
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		flashAndRedirect(a, w, r, "error", "File size exceeds "+strconv.Itoa(fileLimitMB)+"MB limit. Download aborted.", filesRedirectPath(pathParam))
		return
	}

	f, openErr := os.Open(realPath)
	if openErr != nil {
		flashAndRedirect(a, w, r, "error", "Error retrieving file size.", filesRedirectPath(pathParam))
		return
	}
	defer f.Close()

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
		if readErr == io.EOF {
			return
		}
		if readErr != nil {
			return
		}
	}
}

// handleViewFile serves a file's raw content for the in-browser viewer,
// enforcing the same extension allowlist and binary-file check as the
// editor.
func handleViewFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
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
			flashAndRedirect(a, w, r, "error", "Editing this file type is not allowed.", "/files")
			return
		}
	}

	realPath, perr := paths.SecureUserPath("HOME", user.Context, filename, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	isNew := r.URL.Query().Get("new") == "true"
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_view_size", "500"), 500)

	if !isNew {
		info, statErr := os.Stat(realPath)
		if statErr != nil {
			flashAndRedirect(a, w, r, "error", "Error accessing file size.", "/files")
			return
		}
		if info.Size() > int64(fileLimitMB)*1024*1024 {
			flashAndRedirect(a, w, r, "error", "File is too large to view in the browser (limit is "+strconv.Itoa(fileLimitMB)+" MB)", "/files")
			return
		}
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if strings.HasPrefix(mimeType, "image/") {
		http.ServeFile(w, r, realPath)
		return
	}

	var content []byte
	if isNew {
		if _, statErr := os.Stat(realPath); statErr != nil {
			content = nil
		} else {
			content, err = os.ReadFile(realPath)
		}
	} else {
		content, err = os.ReadFile(realPath)
	}
	if err != nil {
		flashAndRedirect(a, w, r, "error", "Error opening file.", "/files")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}
