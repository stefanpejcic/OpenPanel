package filemanager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// apiUploadFiles is handleUploadFiles's POST path (upload.go) with a JSON response instead of a flash + re-rendered form
func apiUploadFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fileLimitMB := uploadSizeMB(a)
	fileLimitBytes := int64(fileLimitMB) * 1024 * 1024

	if parseErr := parseUploadForm(r); parseErr != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Error uploading files: " + parseErr.Error()})
		return
	}
	pathParam := r.Form.Get("path_param")

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "No files were uploaded"})
		return
	}

	root, perr := paths.OpenUserRoot("HOME", user.Context)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	uploaded := []string{}
	var errs []string

	for _, fh := range files {
		if fh.Filename == "" {
			continue
		}
		if fh.Size > fileLimitBytes {
			errs = append(errs, fmt.Sprintf("%s exceeds the upload limit of %d MB.", fh.Filename, fileLimitMB))
			continue
		}

		safeName := secureFilename(fh.Filename)
		if safeName == "" {
			errs = append(errs, "Invalid filename: "+fh.Filename)
			continue
		}
		destPath, perr := root.Rel(filepath.Join(pathParam, safeName), false)
		if perr != nil {
			_, msg := pathErrorStatus(perr)
			errs = append(errs, safeName+": "+msg)
			continue
		}

		src, openErr := fh.Open()
		if openErr != nil {
			errs = append(errs, "Failed to save "+fh.Filename)
			continue
		}
		dst, createErr := root.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if createErr != nil {
			_ = src.Close()
			errs = append(errs, "Failed to save "+fh.Filename)
			continue
		}
		_, copyErr := io.Copy(dst, src)
		_ = src.Close()
		_ = dst.Close()
		if copyErr != nil {
			errs = append(errs, "Failed to save "+fh.Filename)
			continue
		}

		if chownErr := chownToUser(ctx, a, root, destPath, user.Context); chownErr != nil {
			errs = append(errs, "Failed to set ownership on "+fh.Filename)
		}

		_ = logger.RecordUserAction(a.Config, user.Username, "uploaded a new file via File Manager API", reqip.ClientIP(r))
		uploaded = append(uploaded, safeName)
	}

	status := http.StatusOK
	if len(uploaded) == 0 {
		status = http.StatusInternalServerError
	}
	web.WriteJSON(w, status, map[string]any{"uploaded": uploaded, "errors": errs})
}

// apiExtractFiles is handleExtractFiles (archive.go) with a JSON body/response.
func apiExtractFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ArchiveName        string `json:"archive_name"`
		Path               string `json:"path"`
		ExtractDestination string `json:"extract_destination"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.ArchiveName = form.Get("archiveName")
		body.Path = form.Get("path_param")
		body.ExtractDestination = form.Get("extractDestination")
	})
	selectedFile := strings.TrimSpace(body.ArchiveName)
	pathParam := strings.TrimPrefix(body.Path, "/")
	extractionPathInput := strings.TrimSpace(body.ExtractDestination)

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var extractionPath string
	switch {
	case extractionPathInput == "/":
		extractionPath = ""
	case extractionPathInput == "":
		extractionPath = pathParam
	default:
		extractionPath = strings.TrimPrefix(extractionPathInput, "/")
	}

	root, destinationPath, perr := openUserPath(user, extractionPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()
	if mkErr := root.MkdirAll(destinationPath, 0o755); mkErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error occurred before starting archive extraction: " + mkErr.Error()})
		return
	}
	archivePath, perr := root.Rel(filepath.Join(pathParam, selectedFile), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	if archiveKind(selectedFile) == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported file format!"})
		return
	}

	if verr := validateArchiveMembers(root, archivePath, selectedFile, destinationPath); verr != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": verr.Error()})
		return
	}

	maxTime, tErr := strconv.ParseFloat(a.Config.Get("filemanager_extract_max_time", "5"), 64)
	if tErr != nil || maxTime <= 0 {
		maxTime = 5.0
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(maxTime*float64(time.Minute)))
	defer cancel()

	if extractErr := extractArchive(timeoutCtx, root, archivePath, selectedFile, destinationPath); extractErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			web.WriteJSON(w, http.StatusGatewayTimeout, map[string]string{"error": fmt.Sprintf("Extraction timed out after %g minutes!", maxTime)})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Extraction failed."})
		return
	}

	chownRecursive(ctx, a, root, destinationPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "Extracted "+selectedFile+" into "+root.Abs(destinationPath)+" using File Manager API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "File extracted successfully"})
}

// apiCompressFiles is handleCompressFiles (archive.go) with a JSON body/response.
func apiCompressFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ArchiveName   string   `json:"archive_name"`
		Extension     string   `json:"extension"`
		Path          string   `json:"path"`
		SelectedFiles []string `json:"selected_files"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.ArchiveName = form.Get("archiveName")
		body.Extension = form.Get("extension")
		body.Path = form.Get("pathParam")
		body.SelectedFiles = form["selectedFiles[]"]
	})
	archiveNameRaw := strings.TrimPrefix(strings.TrimSpace(body.ArchiveName), "/")
	ext := strings.TrimSpace(body.Extension)
	pathParam := strings.TrimPrefix(body.Path, "/")

	var missing []string
	if archiveNameRaw == "" {
		missing = append(missing, "archive_name")
	}
	if ext == "" {
		missing = append(missing, "extension")
	}
	if len(body.SelectedFiles) == 0 {
		missing = append(missing, "selected_files")
	}
	if len(missing) > 0 {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required fields: " + strings.Join(missing, ", ")})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, parentPath, perr := openUserPath(user, pathParam, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	archiveRelPath := archiveNameRaw
	if !strings.HasSuffix(archiveRelPath, "."+ext) {
		archiveRelPath += "." + ext
	}
	archivePath, perr := root.Rel(archiveRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	validatedFiles, perr := selectedArchiveEntries(root, pathParam, parentPath, body.SelectedFiles)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	if !supportedArchiveExt(ext) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported archive format."})
		return
	}
	maxTime := atoiDefault(a.Config.Get("filemanager_compress_max_time", "5"), 5)
	if cmdErr := createArchive(ctx, root, ext, parentPath, archivePath, validatedFiles, maxTime); cmdErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Archive creation failed: " + cmdErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "Created archive "+filepath.Base(archiveRelPath)+" with File Manager API", reqip.ClientIP(r))

	warning := ""
	if chownErr := chownToUser(ctx, a, root, archivePath, user.Context); chownErr != nil {
		warning = "Failed to set archive ownership"
	}
	web.WriteJSON(w, http.StatusCreated, map[string]string{"message": "Archive created successfully.", "path": archiveRelPath, "warning": warning})
}
