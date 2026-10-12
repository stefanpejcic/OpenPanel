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

// handleExtractFiles extracts an uploaded archive (zip/tar/tar.gz/tgz/gz) into a destination directory, validating member paths first
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
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error occurred before starting archive extraction: %(error)s", "error", mkErr.Error()), filesRedirectPath(pathParam))
		return
	}
	archivePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, selectedFile), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if verr := validateArchiveMembers(archivePath, selectedFile, destinationPath); verr != nil {
		web.FlashRedirect(a, w, r, "error", verr.Error(), filesRedirectPath(pathParam))
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
		extractErr = runCommand(timeoutCtx, "tar", "--no-same-owner", "--no-same-permissions", "-xzf", archivePath, "-C", destinationPath)
	case strings.HasSuffix(selectedFile, ".tar"):
		extractErr = runCommand(timeoutCtx, "tar", "--no-same-owner", "--no-same-permissions", "-xf", archivePath, "-C", destinationPath)
	case strings.HasSuffix(selectedFile, ".gz"):
		extractErr = extractSingleGzip(archivePath, selectedFile, destinationPath)
	default:
		web.FlashRedirect(a, w, r, "error", "Unsupported file format!", filesRedirectPath(pathParam))
		return
	}

	if extractErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Extraction timed out after %(minutes)s minutes!", "minutes", maxTime), filesRedirectPath(pathParam))
			return
		}
		web.FlashRedirect(a, w, r, "error", "Extraction failed.", filesRedirectPath(pathParam))
		return
	}

	chownRecursive(ctx, a, destinationPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "Extracted "+selectedFile+" into "+destinationPath+" using File Manager", reqip.ClientIP(r))
	web.FlashRedirect(a, w, r, "success", "File extracted successfully", filesRedirectPath(pathParam))
}

// isSafeMember reports whether an extracted member resolves to somewhere inside destinationPath and doesn't pass through a symlink that's already there
func isSafeMember(destinationPath, memberName string) bool {
	target := filepath.Join(destinationPath, memberName)
	rel, err := filepath.Rel(destinationPath, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	current := destinationPath
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if fi, err := os.Lstat(current); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return false
		}
	}
	return true
}

// isSafeLink reports whether a symlink member's target stays inside destinationPath
func isSafeLink(destinationPath, memberName, linkTarget string) bool {
	if linkTarget == "" || filepath.IsAbs(linkTarget) {
		return false
	}
	resolved := filepath.Join(destinationPath, filepath.Dir(memberName), linkTarget)
	rel, err := filepath.Rel(destinationPath, resolved)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// validateArchiveMembers pre-checks every entry (paths and link targets) before shelling out to unzip/tar, and fails closed if the archive can't be read
func validateArchiveMembers(archivePath, selectedFile, destinationPath string) error {
	switch {
	case strings.HasSuffix(selectedFile, ".zip"):
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return fmt.Errorf("Could not read archive: %v", err)
		}
		defer zr.Close()
		for _, f := range zr.File {
			if !isSafeMember(destinationPath, f.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", f.Name)
			}
			if f.Mode()&os.ModeSymlink != 0 {
				target, err := readZipLink(f)
				if err != nil || !isSafeLink(destinationPath, f.Name, target) {
					return fmt.Errorf("Unsafe symlink in archive: %s", f.Name)
				}
			}
		}
	case strings.HasSuffix(selectedFile, ".tar.gz"), strings.HasSuffix(selectedFile, ".tgz"), strings.HasSuffix(selectedFile, ".tar"):
		f, err := os.Open(archivePath)
		if err != nil {
			return fmt.Errorf("Could not read archive: %v", err)
		}
		defer f.Close()
		var reader io.Reader = f
		if strings.HasSuffix(selectedFile, ".tar.gz") || strings.HasSuffix(selectedFile, ".tgz") {
			gz, gzErr := gzip.NewReader(f)
			if gzErr != nil {
				return fmt.Errorf("Could not read archive: %v", gzErr)
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
				return fmt.Errorf("Could not read archive: %v", err)
			}
			if !isSafeMember(destinationPath, hdr.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", hdr.Name)
			}
			switch hdr.Typeflag {
			case tar.TypeSymlink:
				if !isSafeLink(destinationPath, hdr.Name, hdr.Linkname) {
					return fmt.Errorf("Unsafe symlink in archive: %s", hdr.Name)
				}
			case tar.TypeLink:
				if !isSafeMember(destinationPath, hdr.Linkname) {
					return fmt.Errorf("Unsafe hard link in archive: %s", hdr.Name)
				}
			case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
				return fmt.Errorf("Device files are not allowed in archive: %s", hdr.Name)
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

// readZipLink returns a zip symlink entry's target, which is stored as the entry's content
func readZipLink(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	b, err := io.ReadAll(io.LimitReader(rc, 4096))
	return string(b), err
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

func supportedArchiveExt(ext string) bool {
	return ext == "zip" || ext == "tar" || ext == "tar.gz"
}

// createArchive runs zip/tar without a shell, the ./ prefix stops a file named like --checkpoint-action from being read as an option and zip -y stores symlinks instead of following them
func createArchive(ctx context.Context, ext, dir, archivePath string, files []string, maxMinutes int) ([]byte, error) {
	var name string
	var args []string
	switch ext {
	case "zip":
		name, args = "zip", []string{"-r", "-y", archivePath}
	case "tar":
		name, args = "tar", []string{"-cf", archivePath}
	case "tar.gz":
		name, args = "tar", []string{"-czf", archivePath}
	default:
		return nil, fmt.Errorf("unsupported archive format %q", ext)
	}
	for _, f := range files {
		args = append(args, "./"+f)
	}
	if maxMinutes <= 0 {
		maxMinutes = 5
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(maxMinutes)*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// handleCompressFiles archives the selected files into a new zip/tar/tar.gz by shelling out to the corresponding command-line tool
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
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Missing required form data for compression: %(missing)s", "missing", strings.Join(missing, ", ")), filesRedirectPath(pathParam))
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

	if !supportedArchiveExt(ext) {
		web.FlashRedirect(a, w, r, "error", "Unsupported archive format.", filesRedirectPath(pathParam))
		return
	}
	maxTime := atoiDefault(a.Config.Get("filemanager_compress_max_time", "5"), 5)
	out, cmdErr := createArchive(ctx, ext, parentPath, archivePath, validatedFiles, maxTime)
	if cmdErr != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = "Unknown error"
		}
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Archive creation failed: %(err_msg)s", "err_msg", errMsg), filesRedirectPath(pathParam))
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
