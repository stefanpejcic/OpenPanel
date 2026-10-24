package filemanager

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// FolderInfo describes one folder returned by the folder picker.
type FolderInfo struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	HasSubfolders bool   `json:"has_subfolders"`
}

// hasSubfolder reads dir in batches and stops at the first folder, reading and sorting every entry was slow on big upload dirs
func hasSubfolder(root *paths.UserRoot, dir string) bool {
	f, err := root.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()
	for {
		entries, err := f.ReadDir(256)
		for _, e := range entries {
			if e.IsDir() {
				return true
			}
		}
		if err != nil {
			return false
		}
	}
}

// handleFolders serves the copy/move modal's folder picker.
func handleFolders(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, safeDir, perr := openUserPath(user, pathParam, true)
	if perr != nil {
		if pe, ok := perr.(*paths.Error); ok && pe.Code == http.StatusNotFound {
			web.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "The specified directory does not exist."})
			return
		}
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"error": msg})
		return
	}

	defer root.Close()

	entries, err := fs.ReadDir(root.FS(), safeDir)
	if err != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "The specified directory does not exist."})
		return
	}

	var folderInfo []FolderInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folderInfo = append(folderInfo, FolderInfo{
			Name: e.Name(), Path: filepath.Join(pathParam, e.Name()), HasSubfolders: hasSubfolder(root, path.Join(safeDir, e.Name())),
		})
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{"folders": folderInfo})
}

// filesPerPage is the listing page size, falling back to 500 since OpenAdmin accepts 0 and the paging divides by it
func filesPerPage(a *appctx.App) int {
	if n := atoiDefault(a.Config.Get("filemanager_files_per_page", "500"), 500); n > 0 {
		return n
	}
	return 500
}

// handleFiles serves the main file-manager listing page.
func handleFiles(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string) {
	// the malware scanner has its own page for quarantined files
	if p := strings.Trim(pathParam, "/"); (p == ".quarantine" || strings.HasPrefix(p, ".quarantine/")) && userAllows(a, r, "malware_scan") {
		http.Redirect(w, r, "/malware-scanner/quarantine", http.StatusFound)
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	extensions := stripQuotes(a.Config.Get("filemanager_edit_extensions",
		".txt .md error_log .log env gitconfig cfg htaccess Gemfile .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf .java"))
	images := stripQuotes(a.Config.Get("filemanager_image_extensions", ".jpg .jpeg .png .gif .webp .avif"))
	archives := stripQuotes(a.Config.Get("filemanager_archives_extensions", ".zip .tar .gz .tar.gz .tgz"))

	root, directory, perr := openUserPath(user, pathParam, true)
	if perr != nil {
		web.FlashRedirect(a, w, r, "error", "Directory does not exist.", "/files")
		return
	}
	defer root.Close()

	title := pathParam
	if title == "" {
		title = "File Manager"
	}

	lsFlag := "-la"
	if r.URL.Query().Get("hidden_files") == "false" {
		lsFlag = "-l"
	}

	filemanagerEditSize := editSizeMB(a)
	filemanagerViewSize := atoiDefault(a.Config.Get("filemanager_view_size", "5"), 5)
	filemanagerDownloadSize := atoiDefault(a.Config.Get("filemanager_download_size", "2000"), 2000)
	filemanagerUploadSize := uploadSizeMB(a)
	filemanagerFilesPerPage := filesPerPage(a)

	var filesInfo []FileEntry
	var readmeContent string
	var hasReadme bool

	out, lsErr := runOnDir(ctx, root, directory, "ls", lsFlag)
	if lsErr != nil {
		if pathParam != "" {
			flashOnlySession(a, r, w, "error", "Error accessing the specified directory.")
		}
	} else {
		filesInfo = parseLsOutput(string(out))
		if data, ok := readReadme(root, path.Join(directory, "README.md")); ok {
			readmeContent = string(data)
			hasReadme = true
		}
	}

	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	totalFiles := len(filesInfo)
	totalPages := (totalFiles + filemanagerFilesPerPage - 1) / filemanagerFilesPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	startIndex := (page - 1) * filemanagerFilesPerPage
	endIndex := startIndex + filemanagerFilesPerPage
	if endIndex > totalFiles {
		endIndex = totalFiles
	}
	if startIndex > totalFiles {
		startIndex = totalFiles
	}
	paginatedFiles := filesInfo[startIndex:endIndex]

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"files_info": paginatedFiles,
			"pagination": map[string]any{
				"current_page": page, "total_pages": totalPages,
				"per_page": filemanagerFilesPerPage, "total_files": totalFiles,
			},
			"limits": map[string]any{
				"edit_size_mb": filemanagerEditSize, "view_size_mb": filemanagerViewSize,
				"download_size_mb": filemanagerDownloadSize, "upload_size_mb": filemanagerUploadSize,
			},
			"extensions": map[string]any{"extensions": extensions, "images": images, "archives": archives},
		})
		return
	}

	view := web.ReadFilemanagerView(a, r)

	startLineNumber := 0
	if totalFiles > 0 {
		startLineNumber = startIndex + 1
	}

	renderFilesPage(a, w, r, filesPageParams{
		Title: title, PathParam: pathParam, FilesInfo: paginatedFiles,
		HasReadme: hasReadme, ReadmeContent: readmeContent,
		FilemanagerEditSize: filemanagerEditSize, FilemanagerViewSize: filemanagerViewSize,
		FilemanagerDownloadSize: filemanagerDownloadSize, FilemanagerUploadSize: filemanagerUploadSize,
		Extensions: extensions, Images: images, Archives: archives,
		View: view, CurrentPage: page, TotalPages: totalPages, TotalFiles: totalFiles,
		StartLineNumber: startLineNumber, EndLineNumber: endIndex,
	})
}

// readmeMaxBytes is the biggest README.md shown under a listing
const readmeMaxBytes = 1 << 20

// readReadme reads README.md for the listing, a bigger one or one that isn't a regular file is skipped, it used to be read whole on every visit
func readReadme(root *paths.UserRoot, rel string) ([]byte, bool) {
	// non-blocking so a fifo named README.md can't hang the listing before the type check
	f, err := root.OpenFile(rel, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > readmeMaxBytes {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(f, readmeMaxBytes))
	return data, err == nil
}

// runOnDir runs name args... /proc/self/fd/3/ with rel already opened as fd 3, so the command sees that exact dir even if its path gets swapped
func runOnDir(ctx context.Context, root *paths.UserRoot, rel, name string, args ...string) ([]byte, error) {
	dir, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	if info, err := dir.Stat(); err != nil || !info.IsDir() {
		return nil, os.ErrNotExist
	}
	cmd := exec.CommandContext(ctx, name, append(args, "/proc/self/fd/3/")...)
	cmd.ExtraFiles = []*os.File{dir}
	return cmd.Output()
}
