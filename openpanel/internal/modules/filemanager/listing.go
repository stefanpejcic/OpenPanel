package filemanager

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
)

// FolderInfo describes one folder returned by the folder picker.
type FolderInfo struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	HasSubfolders bool   `json:"has_subfolders"`
}

// handleFolders serves the copy/move modal's folder picker.
func handleFolders(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	safeDir, perr := paths.SecureUserPath("HOME", user.Context, pathParam, true)
	if perr != nil {
		if pe, ok := perr.(*paths.Error); ok && pe.Code == http.StatusNotFound {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "The specified directory does not exist."})
			return
		}
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"error": msg})
		return
	}

	entries, err := os.ReadDir(safeDir)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "The specified directory does not exist."})
		return
	}

	var folderInfo []FolderInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folderPath := filepath.Join(safeDir, e.Name())
		hasSubfolders := false
		if subEntries, subErr := os.ReadDir(folderPath); subErr == nil {
			for _, sub := range subEntries {
				if sub.IsDir() {
					hasSubfolders = true
					break
				}
			}
		}
		folderInfo = append(folderInfo, FolderInfo{
			Name: e.Name(), Path: filepath.Join(pathParam, e.Name()), HasSubfolders: hasSubfolders,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"folders": folderInfo})
}

// handleFiles serves the main file-manager listing page.
func handleFiles(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string) {
	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	extensions := stripQuotes(a.Config.Get("filemanager_edit_extensions",
		".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf"))
	images := stripQuotes(a.Config.Get("filemanager_image_extensions", ".jpg .jpeg .png .gif .webp .avif"))
	archives := stripQuotes(a.Config.Get("filemanager_archives_extensions", ".zip .tar .gz .tar.gz"))

	directory, perr := paths.SecureUserPath("HOME", user.Context, pathParam, true)
	if perr != nil {
		flashAndRedirect(a, w, r, "error", "Directory does not exist.", "/files")
		return
	}

	title := pathParam
	if title == "" {
		title = "File Manager"
	}

	lsFlag := "-la"
	if r.URL.Query().Get("hidden_files") == "false" {
		lsFlag = "-l"
	}

	filemanagerEditSize := atoiDefault(a.Config.Get("filemanager_edit_size", "5"), 5)
	filemanagerViewSize := atoiDefault(a.Config.Get("filemanager_view_size", "5"), 5)
	filemanagerDownloadSize := atoiDefault(a.Config.Get("filemanager_download_size", "2000"), 2000)
	filemanagerUploadSize := atoiDefault(a.Config.Get("filemanager_upload_size", "2000"), 2000)
	filemanagerFilesPerPage := atoiDefault(a.Config.Get("filemanager_files_per_page", "500"), 500)

	var filesInfo []FileEntry
	var readmeContent string
	var hasReadme bool

	out, lsErr := exec.CommandContext(ctx, "ls", lsFlag, directory).Output()
	if lsErr != nil {
		if pathParam != "" {
			flashOnlySession(a, r, w, "error", "Error accessing the specified directory.")
		}
	} else {
		filesInfo = parseLsOutput(string(out))
		readmePath := filepath.Join(directory, "README.md")
		if data, readErr := os.ReadFile(readmePath); readErr == nil {
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
		writeJSON(w, http.StatusOK, map[string]any{
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

	view := r.URL.Query().Get("view")
	if view != "classic" && view != "modern" {
		view = a.Config.Get("filemanager_buttons_style", "classic")
	}

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
