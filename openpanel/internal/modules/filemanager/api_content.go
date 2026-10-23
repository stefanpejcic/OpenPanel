package filemanager

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// apiGetFileContent is handleEditFile's GET branch (editor.go), reused via forceJSONOutput so it already returns the raw content as a bare JSON string (matching the web editor's own "?output=json" AJAX contract) - same allowlist/binary-sniff/size-limit checks apply
func apiGetFileContent(a *appctx.App, w http.ResponseWriter, r *http.Request, filePath string) {
	handleEditFile(a, w, forceJSONOutput(r), filePath)
}

// apiSaveFileContent is handleEditFile's POST branch (editor.go) with a JSON body/response instead of a form post + flash/redirect
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

	root, rel, perr := openUserPath(user, filePath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	if !containsString(dottedExts, fileExt) && !hasBareName(strings.ToLower(filename), bareNames) {
		if !isEditableFile(root, rel) {
			web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Editing this file type is not allowed."})
			return
		}
	}

	content, ok := apiContentBody(r)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing content field."})
		return
	}

	contentToWrite := content
	if containsString(unixLineEndingExts, fileExt) || containsString(unixLineEndingExts, strings.ToLower(filename)) {
		contentToWrite = strings.ReplaceAll(contentToWrite, "\r\n", "\n")
		contentToWrite = strings.ReplaceAll(contentToWrite, "\r", "\n")
		contentToWrite = strings.TrimPrefix(contentToWrite, "\ufeff")
	}

	if writeErr := root.WriteFileAtomic(rel, []byte(contentToWrite), 0o644); writeErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving file."})
		return
	}

	_ = chownToUser(ctx, a, root, rel, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "edited file "+filePath+" via API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "File saved successfully"})
}

// apiContentBody reads content from JSON or a form post, ok is false when no content field was sent at all
func apiContentBody(r *http.Request) (string, bool) {
	var body struct {
		Content *string `json:"content"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		for _, key := range []string{"editor_content", "content"} {
			if form.Has(key) {
				v := form.Get(key)
				body.Content = &v
				return
			}
		}
	})
	if body.Content == nil {
		return "", false
	}
	return *body.Content, true
}

// apiDownloadFile is handleDownloadFile (editor.go) with JSON error responses instead of flash+redirect - the success path (streaming the file) is unchanged
func apiDownloadFile(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
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
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	f, openErr := root.Open(rel)
	if openErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Error retrieving file size."})
		return
	}
	defer f.Close()
	info, statErr := f.Stat()
	if statErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Error retrieving file size."})
		return
	}
	fileLimitMB := atoiDefault(a.Config.Get("filemanager_download_size", "500"), 500)
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		web.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "File size exceeds " + strconv.Itoa(fileLimitMB) + "MB limit. Download aborted."})
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "downloaded "+filepath.Join(pathParam, filename)+" via File Manager API", reqip.ClientIP(r))

	maxDuration := time.Duration(atoiDefault(a.Config.Get("filemanager_download_max_time", "60"), 60)) * time.Minute
	deadline := time.Now().Add(maxDuration)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+path.Base(rel)+`"`)

	sendFile(w, f, info.Size(), func() bool { return time.Now().After(deadline) })
}

// apiViewFile is handleViewFile (editor.go) with JSON error responses instead of flash+redirect - the success path (serving raw content) is unchanged
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

	root, rel, perr := openUserPath(user, filename, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	if !containsString(dottedExts, fileExt) && !hasBareName(strings.ToLower(baseName), editBareNames) && !isEditableFile(root, rel) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Viewing this file type is not allowed."})
		return
	}

	fileLimitMB := atoiDefault(a.Config.Get("filemanager_view_size", "500"), 500)
	f, openErr := root.Open(rel)
	if openErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Error accessing file size."})
		return
	}
	defer f.Close()
	info, statErr := f.Stat()
	if statErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Error accessing file size."})
		return
	}
	if info.Size() > int64(fileLimitMB)*1024*1024 {
		web.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "File is too large to view (limit is " + strconv.Itoa(fileLimitMB) + " MB)"})
		return
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if strings.HasPrefix(mimeType, "image/") {
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}

	content, readErr := io.ReadAll(f)
	if readErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error opening file."})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}
