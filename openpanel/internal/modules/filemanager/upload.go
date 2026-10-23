package filemanager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var secureFilenameRE = regexp.MustCompile(`[^A-Za-z0-9_.-]`)

// secureFilename strips directory components and anything but ASCII letters/digits/dot/dash/underscore, so an uploaded filename can't escape the target directory or inject odd characters into a path
func secureFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = secureFilenameRE.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	return name
}

// uploadSucceeded is true only when files were sent and none of them failed, rejected names count as failures too
func uploadSucceeded(files int, allSuccess bool, errs []string) bool {
	return files > 0 && allSuccess && len(errs) == 0
}

// handleUploadFiles handles both the initial GET (render the upload form) and the POST that actually saves the submitted files
func handleUploadFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fileLimitMB := atoiDefault(a.Config.Get("filemanager_upload_size", "5"), 1000)
	fileLimitBytes := int64(fileLimitMB) * 1024 * 1024

	var pathParam string
	if r.Method == http.MethodGet {
		pathParam = r.URL.Query().Get("path_param")
	}

	if r.Method == http.MethodPost {
		// the /files drop zone posts via XHR and wants errors back inline instead of a flash
		wantsJSON := strings.Contains(r.Header.Get("Accept"), "application/json")
		pathParam = r.URL.Query().Get("path_param")
		if pp := r.FormValue("path_param"); pp != "" {
			pathParam = pp
		}

		if parseErr := r.ParseMultipartForm(fileLimitBytes + 1<<20); parseErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "Error uploading files: " + parseErr.Error()})
			return
		}
		pathParam = r.Form.Get("path_param")

		files := r.MultipartForm.File["files"]
		if len(files) == 0 && !wantsJSON {
			flashOnlySession(a, r, w, "error", "No files were uploaded")
		}

		root, perr := paths.OpenUserRoot("HOME", user.Context)
		if perr != nil {
			status, msg := pathErrorStatus(perr)
			http.Error(w, msg, status)
			return
		}
		defer root.Close()

		allSuccess := true
		var errs []string

		for _, fh := range files {
			if fh.Filename == "" {
				continue
			}
			if fh.Size > fileLimitBytes {
				allSuccess = false
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
				allSuccess = false
				status, msg := pathErrorStatus(perr)
				http.Error(w, msg, status)
				return
			}

			src, openErr := fh.Open()
			if openErr != nil {
				allSuccess = false
				errs = append(errs, "Failed to save "+fh.Filename)
				continue
			}
			dst, createErr := root.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if createErr != nil {
				_ = src.Close()
				allSuccess = false
				errs = append(errs, "Failed to save "+fh.Filename)
				continue
			}
			_, copyErr := io.Copy(dst, src)
			_ = src.Close()
			_ = dst.Close()
			if copyErr != nil {
				allSuccess = false
				errs = append(errs, "Failed to save "+fh.Filename)
				continue
			}

			if chownErr := chownToUser(ctx, a, root, destPath, user.Context); chownErr != nil {
				allSuccess = false
				errs = append(errs, "Failed to set ownership on "+fh.Filename)
			}

			_ = logger.RecordUserAction(a.Config, user.Username, "uploaded a new file via File Manager", reqip.ClientIP(r))
		}

		if wantsJSON {
			if uploadSucceeded(len(files), allSuccess, errs) {
				flashOnlySession(a, r, w, "success", "Files uploaded successfully.")
				web.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
			} else {
				if len(files) == 0 {
					errs = append(errs, "No files were uploaded")
				}
				web.WriteJSON(w, http.StatusOK, map[string]any{"success": false, "errors": errs})
			}
			return
		}

		switch {
		case len(files) == 0:
			// "No files were uploaded" is already flashed above
		case uploadSucceeded(len(files), allSuccess, errs):
			flashOnlySession(a, r, w, "success", "Files uploaded successfully.")
		default:
			flashOnlySession(a, r, w, "error", web.Tr(a, r, "Some files failed: %(errs)s", "errs", strings.Join(errs, "; ")))
		}
	}

	renderUploadPage(a, w, r, pathParam, fileLimitMB)
}

// flashOnlySession queues a flash and saves the session immediately - unlike most POST handlers here, upload_files() always falls through to re-rendering the upload form instead of redirecting, so the flash has to be persisted right away instead of piggybacking on a later redirect's Set-Cookie write
func flashOnlySession(a *appctx.App, r *http.Request, w http.ResponseWriter, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}
