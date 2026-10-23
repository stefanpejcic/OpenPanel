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

	root, destinationPath, perr := openUserPath(user, extractionPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()
	if mkErr := root.MkdirAll(destinationPath, 0o755); mkErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error occurred before starting archive extraction: %(error)s", "error", mkErr.Error()), filesRedirectPath(pathParam))
		return
	}
	archivePath, perr := root.Rel(filepath.Join(pathParam, selectedFile), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	if archiveKind(selectedFile) == "" {
		web.FlashRedirect(a, w, r, "error", "Unsupported file format!", filesRedirectPath(pathParam))
		return
	}

	if verr := validateArchiveMembers(root, archivePath, selectedFile, destinationPath); verr != nil {
		web.FlashRedirect(a, w, r, "error", verr.Error(), filesRedirectPath(pathParam))
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
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Extraction timed out after %(minutes)s minutes!", "minutes", maxTime), filesRedirectPath(pathParam))
			return
		}
		web.FlashRedirect(a, w, r, "error", "Extraction failed.", filesRedirectPath(pathParam))
		return
	}

	chownRecursive(ctx, a, root, destinationPath, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username, "Extracted "+selectedFile+" into "+root.Abs(destinationPath)+" using File Manager", reqip.ClientIP(r))
	web.FlashRedirect(a, w, r, "success", "File extracted successfully", filesRedirectPath(pathParam))
}

// archiveKind maps a filename to the extractor that handles it, "" when it isn't supported
func archiveKind(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.HasSuffix(name, ".zip"):
		return "zip"
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(name, ".tar"):
		return "tar"
	case strings.HasSuffix(name, ".gz"):
		return "gz"
	}
	return ""
}

// gzOutputName is the file a .gz extracts to, its name without the .gz in any case
func gzOutputName(selectedFile string) string {
	name := path.Base(selectedFile)
	if ext := path.Ext(name); strings.EqualFold(ext, ".gz") {
		name = strings.TrimSuffix(name, ext)
	}
	return path.Base(name)
}

// memberRel maps an archive member name to a path under destRel, false if it would climb out of it
func memberRel(destRel, name string) (string, bool) {
	clean, ok := paths.MemberName(name)
	if !ok {
		return "", false
	}
	return path.Join(destRel, clean), true
}

// isSafeMember reports whether a member stays inside destRel and doesn't pass through a symlink that's already there
func isSafeMember(root *paths.UserRoot, destRel, name string) bool {
	clean, ok := paths.MemberName(name)
	if !ok {
		return false
	}
	current := destRel
	for _, part := range strings.Split(clean, "/") {
		if part == "" {
			continue
		}
		current = path.Join(current, part)
		if fi, err := root.Lstat(current); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return false
		}
	}
	return true
}

// isSafeLink reports whether a symlink member's target stays inside destRel
func isSafeLink(destRel, name, linkTarget string) bool {
	if linkTarget == "" || path.IsAbs(linkTarget) {
		return false
	}
	clean, ok := paths.MemberName(name)
	if !ok {
		return false
	}
	// resolve against a fake root so climbing past destRel shows up as a leading ..
	resolved := path.Join(path.Dir(clean), linkTarget)
	return resolved != ".." && !strings.HasPrefix(resolved, "../")
}

// validateArchiveMembers pre-checks every entry (paths and link targets) before extracting anything, and fails closed if the archive can't be read
func validateArchiveMembers(root *paths.UserRoot, archiveRel, selectedFile, destRel string) error {
	switch archiveKind(selectedFile) {
	case "zip":
		zr, closeFn, err := openZip(root, archiveRel)
		if err != nil {
			return fmt.Errorf("Could not read archive: %v", err)
		}
		defer closeFn()
		for _, f := range zr.File {
			if !isSafeMember(root, destRel, f.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", f.Name)
			}
			if f.Mode()&os.ModeSymlink != 0 {
				target, err := readZipLink(f)
				if err != nil || !isSafeLink(destRel, f.Name, target) {
					return fmt.Errorf("Unsafe symlink in archive: %s", f.Name)
				}
			}
		}
	case "tar", "tar.gz":
		tr, closeFn, err := openTar(root, archiveRel, selectedFile)
		if err != nil {
			return fmt.Errorf("Could not read archive: %v", err)
		}
		defer closeFn()
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("Could not read archive: %v", err)
			}
			if !isSafeMember(root, destRel, hdr.Name) {
				return fmt.Errorf("Unsafe file path in archive: %s", hdr.Name)
			}
			switch hdr.Typeflag {
			case tar.TypeSymlink:
				if !isSafeLink(destRel, hdr.Name, hdr.Linkname) {
					return fmt.Errorf("Unsafe symlink in archive: %s", hdr.Name)
				}
			case tar.TypeLink:
				if !isSafeMember(root, destRel, hdr.Linkname) {
					return fmt.Errorf("Unsafe hard link in archive: %s", hdr.Name)
				}
			case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
				return fmt.Errorf("Device files are not allowed in archive: %s", hdr.Name)
			}
		}
	case "gz":
		outName := gzOutputName(selectedFile)
		if !isSafeMember(root, destRel, outName) {
			return fmt.Errorf("Unsafe file path in gzip: %s", outName)
		}
	}
	return nil
}

func openZip(root *paths.UserRoot, archiveRel string) (*zip.Reader, func(), error) {
	f, err := root.Open(archiveRel)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return zr, func() { _ = f.Close() }, nil
}

func openTar(root *paths.UserRoot, archiveRel, selectedFile string) (*tar.Reader, func(), error) {
	f, err := root.Open(archiveRel)
	if err != nil {
		return nil, nil, err
	}
	if archiveKind(selectedFile) != "tar.gz" {
		return tar.NewReader(f), func() { _ = f.Close() }, nil
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return tar.NewReader(gz), func() { _ = gz.Close(); _ = f.Close() }, nil
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

// extractArchive unpacks archiveRel into destRel, every write goes through root so a member or a swapped dir can't land outside it
func extractArchive(ctx context.Context, root *paths.UserRoot, archiveRel, selectedFile, destRel string) error {
	switch archiveKind(selectedFile) {
	case "zip":
		zr, closeFn, err := openZip(root, archiveRel)
		if err != nil {
			return err
		}
		defer closeFn()
		return paths.ExtractZip(ctx, root.Root, destRel, zr, 0)
	case "tar", "tar.gz":
		tr, closeFn, err := openTar(root, archiveRel, selectedFile)
		if err != nil {
			return err
		}
		defer closeFn()
		return paths.ExtractTar(ctx, root.Root, destRel, tr, 0)
	case "gz":
		f, err := root.Open(archiveRel)
		if err != nil {
			return err
		}
		defer f.Close()
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		outName := gzOutputName(selectedFile)
		return paths.WriteFileFrom(root.Root, path.Join(destRel, outName), paths.CtxReader(ctx, gz), 0o644)
	}
	return fmt.Errorf("unsupported archive %q", selectedFile)
}

func supportedArchiveExt(ext string) bool {
	return ext == "zip" || ext == "tar" || ext == "tar.gz"
}

// createArchive writes dirRel's selected entries into archiveRel without shelling out, symlinks are stored as links and never followed
func createArchive(ctx context.Context, root *paths.UserRoot, ext, dirRel, archiveRel string, names []string, maxMinutes int) (err error) {
	if maxMinutes <= 0 {
		maxMinutes = 5
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(maxMinutes)*time.Minute)
	defer cancel()

	out, err := root.OpenFile(archiveRel, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = root.Remove(archiveRel)
		}
	}()
	return paths.WriteArchive(ctx, root.Root, ext, dirRel, names, out)
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

	root, parentPath, perr := openUserPath(user, pathParam, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
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
		http.Error(w, msg, status)
		return
	}

	validatedFiles, perr := selectedArchiveEntries(root, pathParam, parentPath, selectedFiles)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if !supportedArchiveExt(ext) {
		web.FlashRedirect(a, w, r, "error", "Unsupported archive format.", filesRedirectPath(pathParam))
		return
	}
	maxTime := atoiDefault(a.Config.Get("filemanager_compress_max_time", "5"), 5)
	if cmdErr := createArchive(ctx, root, ext, parentPath, archivePath, validatedFiles, maxTime); cmdErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Archive creation failed: %(err_msg)s", "err_msg", cmdErr.Error()), filesRedirectPath(pathParam))
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "Created archive "+filepath.Base(archiveRelPath)+" with File Manager", reqip.ClientIP(r))

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, root, archivePath, user.Context); chownErr != nil {
		flash.Add(sess, "error", "Failed to set archive ownership")
	}
	flash.Add(sess, "success", "FILEMANAGER - Archive created successfully.")
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// selectedArchiveEntries validates each selected name under pathParam and returns it relative to parentRel
func selectedArchiveEntries(root *paths.UserRoot, pathParam, parentRel string, selected []string) ([]string, error) {
	names := make([]string, 0, len(selected))
	for _, f := range selected {
		rel, err := root.Rel(filepath.Join(pathParam, f), true)
		if err != nil {
			return nil, err
		}
		if rel == parentRel {
			return nil, &paths.Error{Code: http.StatusForbidden, Message: "Invalid selection"}
		}
		names = append(names, relTo(parentRel, rel))
	}
	return names, nil
}
