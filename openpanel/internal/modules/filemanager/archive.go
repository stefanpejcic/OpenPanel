package filemanager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// handleExtractFiles extracts an uploaded archive (zip/tar/tar.gz/tgz/gz)
// into a destination directory, validating member paths first.
func handleExtractFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	selectedFile := strings.TrimSpace(r.Form.Get("archiveName"))
	pathParam := strings.TrimPrefix(r.Form.Get("path_param"), "/")
	extractionPathInput := strings.TrimSpace(r.Form.Get("extractDestination"))

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
		http.Error(w, msg, status)
		return
	}
	if mkErr := os.MkdirAll(destinationPath, 0o755); mkErr != nil {
		flashAndRedirect(a, w, r, "error", "Error occurred before starting archive extraction: "+mkErr.Error(), filesRedirectPath(pathParam))
		return
	}
	archivePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, selectedFile), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if verr := validateArchiveMembers(archivePath, selectedFile, destinationPath); verr != nil {
		flashAndRedirect(a, w, r, "error", verr.Error(), filesRedirectPath(pathParam))
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
		flashAndRedirect(a, w, r, "error", "Unsupported file format!", filesRedirectPath(pathParam))
		return
	}

	if extractErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			flashAndRedirect(a, w, r, "error", fmt.Sprintf("Extraction timed out after %g minutes!", maxTime), filesRedirectPath(pathParam))
			return
		}
		flashAndRedirect(a, w, r, "error", "Extraction failed.", filesRedirectPath(pathParam))
		return
	}

	chownRecursive(ctx, a, destinationPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "Extracted "+selectedFile+" into "+destinationPath+" using File Manager", reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", "File extracted successfully", filesRedirectPath(pathParam))
}

// isSafeMember reports whether an extracted member resolves to somewhere
// inside destinationPath - a zip-slip / tar-slip guard.
func isSafeMember(destinationPath, memberName string) bool {
	target := filepath.Join(destinationPath, memberName)
	rel, err := filepath.Rel(destinationPath, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// validateArchiveMembers pre-checks every entry in the archive before
// shelling out to unzip/tar, so a malicious archive can't write outside
// destinationPath.
func validateArchiveMembers(archivePath, selectedFile, destinationPath string) error {
	switch {
	case strings.HasSuffix(selectedFile, ".zip"):
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return nil //nolint:nilerr // only member-path safety is validated here, not archive integrity (the extract step below reports real failures)
		}
		defer zr.Close()
		for _, f := range zr.File {
			if !isSafeMember(destinationPath, f.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", f.Name)
			}
		}
	case strings.HasSuffix(selectedFile, ".tar.gz"), strings.HasSuffix(selectedFile, ".tgz"), strings.HasSuffix(selectedFile, ".tar"):
		f, err := os.Open(archivePath)
		if err != nil {
			return nil //nolint:nilerr
		}
		defer f.Close()
		var reader io.Reader = f
		if strings.HasSuffix(selectedFile, ".tar.gz") || strings.HasSuffix(selectedFile, ".tgz") {
			gz, gzErr := gzip.NewReader(f)
			if gzErr != nil {
				return nil //nolint:nilerr
			}
			defer gz.Close()
			reader = gz
		}
		tr := tar.NewReader(reader)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil //nolint:nilerr
			}
			if !isSafeMember(destinationPath, hdr.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", hdr.Name)
			}
		}
	case strings.HasSuffix(selectedFile, ".gz"):
		outName := filepath.Base(strings.TrimSuffix(selectedFile, ".gz"))
		if !isSafeMember(destinationPath, outName) {
			return fmt.Errorf("Unsafe file path in gzip: %s", outName)
		}
	}
	return nil
}

func extractSingleGzip(archivePath, selectedFile, destinationPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	outName := filepath.Base(strings.TrimSuffix(selectedFile, ".gz"))
	outPath := filepath.Join(destinationPath, outName)
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, gz)
	return err
}

var shellArgQuoteRE = regexp.MustCompile(`[^\w@%+=:,./-]`)

// shellQuote is the minimal POSIX-shell single-quoting needed to safely
// embed an argument in the "sh -c" command string this route builds -
// compression shells out to zip/tar rather than reimplementing archive
// creation in Go.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !shellArgQuoteRE.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// handleCompressFiles archives the selected files into a new zip/tar/
// tar.gz by shelling out to the corresponding command-line tool.
func handleCompressFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	archiveNameRaw := strings.TrimPrefix(strings.TrimSpace(r.Form.Get("archiveName")), "/")
	ext := strings.TrimSpace(r.Form.Get("extension"))
	pathParam := strings.TrimPrefix(r.Form.Get("pathParam"), "/")
	selectedFiles := r.Form["selectedFiles[]"]

	var missing []string
	if archiveNameRaw == "" {
		missing = append(missing, "archiveName")
	}
	if ext == "" {
		missing = append(missing, "extension")
	}
	if len(selectedFiles) == 0 {
		missing = append(missing, "selectedFiles")
	}
	if len(missing) > 0 {
		flashAndRedirect(a, w, r, "error", "Missing required form data for compression: "+strings.Join(missing, ", "), filesRedirectPath(pathParam))
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
		http.Error(w, msg, status)
		return
	}

	archiveRelPath := archiveNameRaw
	if !strings.HasSuffix(archiveRelPath, "."+ext) {
		archiveRelPath += "." + ext
	}
	archivePath, perr := paths.SecureUserPath("HOME", user.Context, archiveRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	validatedFiles := make([]string, 0, len(selectedFiles))
	for _, f := range selectedFiles {
		filePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, f), true)
		if perr != nil {
			status, msg := pathErrorStatus(perr)
			http.Error(w, msg, status)
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
		flashAndRedirect(a, w, r, "error", "Unsupported archive format.", filesRedirectPath(pathParam))
		return
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", archiveCmd)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = "Unknown error"
		}
		flashAndRedirect(a, w, r, "error", "Archive creation failed: "+errMsg, filesRedirectPath(pathParam))
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "Created archive "+filepath.Base(archiveRelPath)+" with File Manager", reqip.ClientIP(r))

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, archivePath, user.Context); chownErr != nil {
		flash.Add(sess, "error", "Failed to set archive ownership")
	}
	flash.Add(sess, "success", "FILEMANAGER - Archive created successfully.")
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}
