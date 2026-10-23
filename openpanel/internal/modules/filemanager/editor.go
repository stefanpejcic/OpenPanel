package filemanager

import (
	"io"
	"mime"
	"net/http"
	"path"
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
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var allowedEditors = map[string]bool{"monaco": true, "ace": true, "codemirror": true, "text": true}

// isEditableFile is a crude binary-file sniff: no NUL byte in the first 8KB
func isEditableFile(root *paths.UserRoot, rel string) bool {
	f, err := root.Open(rel)
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

// editSizeMB is the editor's size limit, from the same key openpanel.config and OpenAdmin use
func editSizeMB(a *appctx.App) int {
	return atoiDefault(a.Config.Get("filemanager_edit_size", "5"), 5)
}

// uploadSizeMB is the per-file upload limit, the upload handlers used to default to 5 while the page said 2000
func uploadSizeMB(a *appctx.App) int {
	return atoiDefault(a.Config.Get("filemanager_upload_size", "2000"), 2000)
}

func imageExtensions(a *appctx.App) []string {
	raw := stripQuotes(a.Config.Get("filemanager_image_extensions", ".jpg .jpeg .png .gif .webp .avif"))
	return strings.Fields(raw)
}

// splitExtNames splits a config extension list into (dotted extensions, bare filename substrings) depending on whether each entry starts with a dot
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

// hasBareName reports whether one of the bare entries is a whole part of name, so "env" matches .env.local but not envelope.pdf
func hasBareName(name string, bare []string) bool {
	for _, b := range bare {
		if b == "" {
			continue
		}
		for from := 0; ; {
			i := strings.Index(name[from:], b)
			if i == -1 {
				break
			}
			start, end := from+i, from+i+len(b)
			if (start == 0 || isNameSep(name[start-1])) && (end == len(name) || isNameSep(name[end])) {
				return true
			}
			from = start + 1
		}
	}
	return false
}

func isNameSep(c byte) bool {
	return c == '.' || c == '_' || c == '-'
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// unixLineEndingExts lists file types that get \r\n -> \n normalization and a stripped BOM on save
var unixLineEndingExts = []string{
	".sh", ".bash", ".zsh", ".py", ".conf", ".service", ".env", ".ini", ".cfg", ".json", ".php",
	".html", ".htm", ".css", ".js", ".sql", ".xml", ".txt", ".md", ".htaccess", ".log", "error_log",
}

// handleEditFile serves the file editor page on GET and saves the submitted content on POST
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

	// output=json callers (the API) get a JSON error instead of a redirect to /files
	wantsJSON := r.URL.Query().Get("output") == "json"
	fail := func(status int, msg string) {
		if wantsJSON {
			web.WriteJSON(w, status, map[string]string{"error": msg})
			return
		}
		web.FlashRedirect(a, w, r, "error", msg, "/files")
	}

	root, rel, perr := openUserPath(user, filePath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		if wantsJSON {
			web.WriteJSON(w, status, map[string]string{"error": msg})
		} else {
			http.Error(w, msg, status)
		}
		return
	}
	defer root.Close()

	if !containsString(dottedExts, fileExt) && !hasBareName(strings.ToLower(filename), bareNames) {
		if !isEditableFile(root, rel) {
			fail(http.StatusForbidden, "Editing this file type is not allowed.")
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

		// the body was read after the path check, so the write has to go through the root handle, not the path string
		writeErr := root.WriteFileAtomic(rel, []byte(contentToWrite), 0o644)

		sess, _ := a.Sessions.Get(r, session.CookieName)
		if writeErr != nil {
			flash.Add(sess, "error", "Error saving file.")
		} else {
			_ = chownToUser(ctx, a, root, rel, user.Context)
			flash.Add(sess, "success", "File saved successfully")
			_ = logger.RecordUserAction(a.Config, user.Username, "edited file "+filePath, reqip.ClientIP(r))
		}
		_ = a.Sessions.Save(r, w, sess)

		http.Redirect(w, r, "/file-manager/edit-file/"+escapeURLPath(filePath), http.StatusFound)
		return
	}

	// GET
	isNew := r.URL.Query().Get("new") == "true"
	fileLimitMB := editSizeMB(a)
	fileLimitBytes := int64(fileLimitMB) * 1024 * 1024
	var fileContent string

	f, openErr := root.Open(rel)
	if openErr != nil {
		if !isNew {
			fail(http.StatusNotFound, "Error accessing file size.")
			return
		}
	} else {
		defer f.Close()
		info, statErr := f.Stat()
		if statErr != nil {
			fail(http.StatusNotFound, "Error accessing file size.")
			return
		}
		// new=true only means the file may not exist yet, it must not skip the limit or the editor loads a cut-off copy
		if info.Size() > fileLimitBytes {
			fail(http.StatusRequestEntityTooLarge,
				web.Tr(a, r, "File is too large to open in the editor (limit is %(file_limit_mb)s MB)", "file_limit_mb", strconv.Itoa(fileLimitMB)))
			return
		}
		data, readErr := io.ReadAll(io.LimitReader(f, fileLimitBytes+1))
		if readErr != nil {
			if !isNew {
				fail(http.StatusInternalServerError, "Error reading file")
				return
			}
		} else {
			fileContent = string(data)
		}
	}

	if wantsJSON {
		web.WriteJSON(w, http.StatusOK, fileContent)
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

// handleDownloadFile streams the file to the client, aborting once a wall-clock time limit is exceeded so a stalled connection can't hold the handler open indefinitely
func handleDownloadFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	pathParam := r.URL.Query().Get("path_param")

	root, rel, perr := openUserPath(user, filepath.Join(pathParam, filename), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	f, openErr := root.Open(rel)
	if openErr != nil {
		web.FlashRedirect(a, w, r, "error", "Error retrieving file size.", filesRedirectPath(pathParam))
		return
	}
	defer f.Close()
	info, statErr := f.Stat()
	if statErr != nil {
		web.FlashRedirect(a, w, r, "error", "Error retrieving file size.", filesRedirectPath(pathParam))
		return
	}
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_download_size", "500"), 500)
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "File size exceeds %(file_limit_mb)sMB limit. Download aborted.", "file_limit_mb", strconv.Itoa(fileLimitMB)), filesRedirectPath(pathParam))
		return
	}

	maxDuration := time.Duration(atoiDefault(a.Config.Get("filemanager_download_max_time", "60"), 60)) * time.Minute
	deadline := time.Now().Add(maxDuration)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", attachmentHeader(path.Base(rel)))

	sendFile(w, f, info.Size(), func() bool { return time.Now().After(deadline) })
}

// attachmentHeader is the Content-Disposition for a download, mime puts non-ASCII names in filename*= instead of raw UTF-8
func attachmentHeader(name string) string {
	if v := mime.FormatMediaType("attachment", map[string]string{"filename": name}); v != "" {
		return v
	}
	return "attachment"
}

// sendFile streams f with its size declared, a download cut off by the time limit or a read error drops the connection so the browser doesn't keep a short file as complete
func sendFile(w http.ResponseWriter, f io.Reader, size int64, expired func() bool) {
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	buf := make([]byte, 8192)
	var sent int64
	for sent < size {
		if expired() {
			panic(http.ErrAbortHandler)
		}
		n, readErr := f.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
			sent += int64(n)
		}
		if readErr != nil {
			if readErr == io.EOF && sent >= size {
				return
			}
			panic(http.ErrAbortHandler)
		}
	}
}

// handleViewFile serves a file's raw content for the in-browser viewer, enforcing the same extension allowlist and binary-file check as the editor
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

	root, rel, perr := openUserPath(user, filename, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	if !containsString(dottedExts, fileExt) && !hasBareName(strings.ToLower(baseName), editBareNames) && !isEditableFile(root, rel) {
		web.FlashRedirect(a, w, r, "error", "Editing this file type is not allowed.", "/files")
		return
	}

	fileLimitMB := atoiDefault(a.Config.Get("filemanager_view_size", "500"), 500)

	f, openErr := root.Open(rel)
	if openErr != nil {
		web.FlashRedirect(a, w, r, "error", "Error opening file.", "/files")
		return
	}
	defer f.Close()
	info, statErr := f.Stat()
	if statErr != nil {
		web.FlashRedirect(a, w, r, "error", "Error accessing file size.", "/files")
		return
	}
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "File is too large to view in the browser (limit is %(file_limit_mb)s MB)", "file_limit_mb", strconv.Itoa(fileLimitMB)), "/files")
		return
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if strings.HasPrefix(mimeType, "image/") {
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}

	content, err := io.ReadAll(f)
	if err != nil {
		web.FlashRedirect(a, w, r, "error", "Error opening file.", "/files")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}
