package filemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// apiUploadFiles is handleUploadFiles's POST path (upload.go) with a JSON
// response instead of a flash + re-rendered form.
func apiUploadFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fileLimitMB := atoiDefault(a.Config.Get("filemanager_upload_size", "5"), 1000)
	fileLimitBytes := int64(fileLimitMB) * 1024 * 1024

	if parseErr := r.ParseMultipartForm(fileLimitBytes + 1<<20); parseErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Error uploading files: " + parseErr.Error()})
		return
	}
	pathParam := r.Form.Get("path_param")

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No files were uploaded"})
		return
	}

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
		destPath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, safeName), false)
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
		dst, createErr := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
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

		if chownErr := chownToUser(ctx, a, destPath, user.Context); chownErr != nil {
			errs = append(errs, "Failed to set ownership on "+fh.Filename)
		}

		_ = logger.RecordUserAction(a.Config, user.Username, "uploaded a new file via File Manager API", reqip.ClientIP(r))
		uploaded = append(uploaded, safeName)
	}

	status := http.StatusOK
	if len(uploaded) == 0 {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, map[string]any{"uploaded": uploaded, "errors": errs})
}

// apiExtractFiles is handleExtractFiles (archive.go) with a JSON body/response.
func apiExtractFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ArchiveName        string `json:"archive_name"`
		Path               string `json:"path"`
		ExtractDestination string `json:"extract_destination"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.ArchiveName = r.Form.Get("archiveName")
		body.Path = r.Form.Get("path_param")
		body.ExtractDestination = r.Form.Get("extractDestination")
	}
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

	destinationPath, perr := paths.SecureUserPath("HOME", user.Context, extractionPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}
	if mkErr := os.MkdirAll(destinationPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error occurred before starting archive extraction: " + mkErr.Error()})
		return
	}
	archivePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, selectedFile), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	if verr := validateArchiveMembers(archivePath, selectedFile, destinationPath); verr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": verr.Error()})
		return
	}

	maxTime, tErr := strconv.ParseFloat(a.Config.Get("filemanager_extract_max_time", "5"), 64)
	if tErr != nil || maxTime <= 0 {
		maxTime = 5.0
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(maxTime*float64(time.Minute)))
	defer cancel()

	var extractErr error
	switch {
	case strings.HasSuffix(selectedFile, ".zip"):
		extractErr = runCommand(timeoutCtx, "unzip", "-o", archivePath, "-d", destinationPath)
	case strings.HasSuffix(selectedFile, ".tar.gz"), strings.HasSuffix(selectedFile, ".tgz"):
		extractErr = runCommand(timeoutCtx, "tar", "-xzf", archivePath, "-C", destinationPath)
	case strings.HasSuffix(selectedFile, ".tar"):
		extractErr = runCommand(timeoutCtx, "tar", "-xf", archivePath, "-C", destinationPath)
	case strings.HasSuffix(selectedFile, ".gz"):
		extractErr = extractSingleGzip(archivePath, selectedFile, destinationPath)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported file format!"})
		return
	}

	if extractErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": fmt.Sprintf("Extraction timed out after %g minutes!", maxTime)})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Extraction failed."})
		return
	}

	chownRecursive(ctx, a, destinationPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "Extracted "+selectedFile+" into "+destinationPath+" using File Manager API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "File extracted successfully"})
}

// apiCompressFiles is handleCompressFiles (archive.go) with a JSON body/response.
func apiCompressFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ArchiveName   string   `json:"archive_name"`
		Extension     string   `json:"extension"`
		Path          string   `json:"path"`
		SelectedFiles []string `json:"selected_files"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.ArchiveName = r.Form.Get("archiveName")
		body.Extension = r.Form.Get("extension")
		body.Path = r.Form.Get("pathParam")
		body.SelectedFiles = r.Form["selectedFiles[]"]
	}
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required fields: " + strings.Join(missing, ", ")})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	parentPath, perr := paths.SecureUserPath("HOME", user.Context, pathParam, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	archiveRelPath := archiveNameRaw
	if !strings.HasSuffix(archiveRelPath, "."+ext) {
		archiveRelPath += "." + ext
	}
	archivePath, perr := paths.SecureUserPath("HOME", user.Context, archiveRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	validatedFiles := make([]string, 0, len(body.SelectedFiles))
	for _, f := range body.SelectedFiles {
		filePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, f), true)
		if perr != nil {
			status, msg := pathErrorStatus(perr)
			writeJSON(w, status, map[string]string{"error": msg})
			return
		}
		validatedFiles = append(validatedFiles, filepath.Base(filePath))
	}

	maxTime := a.Config.Get("filemanager_compress_max_time", "5")
	quotedFiles := make([]string, len(validatedFiles))
	for i, f := range validatedFiles {
		quotedFiles[i] = shellQuote(f)
	}
	fileList := strings.Join(quotedFiles, " ")

	var archiveCmd string
	switch ext {
	case "zip":
		archiveCmd = fmt.Sprintf("cd %s && timeout %sm zip -r %s %s", shellQuote(parentPath), maxTime, shellQuote(archivePath), fileList)
	case "tar":
		archiveCmd = fmt.Sprintf("cd %s && timeout %sm tar -cf %s %s", shellQuote(parentPath), maxTime, shellQuote(archivePath), fileList)
	case "tar.gz":
		archiveCmd = fmt.Sprintf("cd %s && timeout %sm tar -czf %s %s", shellQuote(parentPath), maxTime, shellQuote(archivePath), fileList)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported archive format."})
		return
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", archiveCmd)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = "Unknown error"
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Archive creation failed: " + errMsg})
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "Created archive "+filepath.Base(archiveRelPath)+" with File Manager API", reqip.ClientIP(r))

	warning := ""
	if chownErr := chownToUser(ctx, a, archivePath, user.Context); chownErr != nil {
		warning = "Failed to set archive ownership"
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Archive created successfully.", "path": archiveRelPath, "warning": warning})
}
