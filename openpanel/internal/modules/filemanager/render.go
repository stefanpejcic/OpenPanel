package filemanager

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var filesPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/filemanager.html",
	"files/filemanager_partials.html",
)

var uploadPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/upload_file.html",
)

var editFilePage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/edit_file.html",
)

// Breadcrumb mirrors one segment of the filemanager.html / edit_file.html
// path breadcrumb trail.
type Breadcrumb struct {
	Name string
	Path string
	Last bool
}

// buildBreadcrumbs splits a path into breadcrumb segments after trimming
// leading/trailing slashes.
func buildBreadcrumbs(pathParam string) []Breadcrumb {
	trimmed := strings.Trim(pathParam, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	crumbs := make([]Breadcrumb, len(parts))
	for i, part := range parts {
		crumbs[i] = Breadcrumb{
			Name: part,
			Path: strings.Join(parts[:i+1], "/"),
			Last: i == len(parts)-1,
		}
	}
	return crumbs
}

// PageEntry is one rendered pagination control: either a page number link
// or an ellipsis. Pages that match neither condition in the pagination
// rules below render nothing at all, so they simply don't appear here.
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries mirrors filemanager.html's pagination loop exactly:
// current page (active), first/last page, and current±2 render as links;
// page 2 and total_pages-1 render as an ellipsis when they don't already
// qualify above; every other page renders nothing.
func buildPageEntries(current, total int) []PageEntry {
	var entries []PageEntry
	for p := 1; p <= total; p++ {
		switch {
		case p == current:
			entries = append(entries, PageEntry{Number: p})
		case p == 1 || p == total || (p >= current-2 && p <= current+2):
			entries = append(entries, PageEntry{Number: p})
		case p == 2 || p == total-1:
			entries = append(entries, PageEntry{IsEllipsis: true})
		}
	}
	return entries
}

// ViewRow is one files_table row.
type ViewRow struct {
	Index                                                                int
	Name, Type, Size, Date, Owner, Group, Links, LinkTarget, Permissions string
	Href, DownloadURL                                                    string
}

func urlEncodedJoin(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('/')
		}
		segments := strings.Split(p, "/")
		for j, seg := range segments {
			if j > 0 {
				b.WriteByte('/')
			}
			b.WriteString(url.PathEscape(seg))
		}
	}
	return b.String()
}

// filesPageParams is what handleFiles passes to renderFilesPage - the
// already-paginated slice of files() plus the surrounding page context.
type filesPageParams struct {
	Title                               string
	PathParam                           string
	FilesInfo                           []FileEntry
	HasReadme                           bool
	ReadmeContent                       string
	FilemanagerEditSize                 int
	FilemanagerViewSize                 int
	FilemanagerDownloadSize             int
	FilemanagerUploadSize               int
	Extensions, Images, Archives        string
	View                                string
	CurrentPage, TotalPages, TotalFiles int
	StartLineNumber, EndLineNumber      int
}

// FilesPageData is filemanager.html's full template context.
type FilesPageData struct {
	web.LayoutData
	PathParam   string
	View        string
	Breadcrumbs []Breadcrumb

	Directories   []ViewRow
	Files         []ViewRow
	Rows          []ViewRow
	ShowUpLink    bool
	UpLink        string
	HasReadme     bool
	ReadmeContent string

	FilemanagerEditSize, FilemanagerViewSize, FilemanagerDownloadSize, FilemanagerUploadSize int
	Extensions, Images, Archives                                                             string

	CurrentPage, TotalPages, TotalFiles int
	StartLineNumber, EndLineNumber      int
	PrevPage, NextPage                  int
	PageEntries                         []PageEntry
}

func renderFilesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, p filesPageParams) {
	layout, _, err := web.BuildLayoutData(a, w, r, p.Title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := FilesPageData{
		LayoutData: layout, PathParam: p.PathParam, View: p.View,
		Breadcrumbs: buildBreadcrumbs(p.PathParam),
		HasReadme:   p.HasReadme, ReadmeContent: p.ReadmeContent,
		FilemanagerEditSize: p.FilemanagerEditSize, FilemanagerViewSize: p.FilemanagerViewSize,
		FilemanagerDownloadSize: p.FilemanagerDownloadSize, FilemanagerUploadSize: p.FilemanagerUploadSize,
		Extensions: p.Extensions, Images: p.Images, Archives: p.Archives,
		CurrentPage: p.CurrentPage, TotalPages: p.TotalPages, TotalFiles: p.TotalFiles,
		StartLineNumber: p.StartLineNumber, EndLineNumber: p.EndLineNumber,
		PrevPage: p.CurrentPage - 1, NextPage: p.CurrentPage + 1,
		PageEntries: buildPageEntries(p.CurrentPage, p.TotalPages),
	}

	if p.PathParam != "" && p.PathParam != "/" {
		data.ShowUpLink = true
		trimmed := strings.TrimRight(p.PathParam, "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			data.UpLink = "/files/" + urlEncodedJoin(trimmed[:idx]) + "?view=" + url.QueryEscape(p.View)
		} else {
			data.UpLink = "/files?view=" + url.QueryEscape(p.View)
		}
	}

	idx := 0
	for _, info := range p.FilesInfo {
		row := ViewRow{
			Index: idx, Name: info.Name, Type: info.Type, Size: info.Size, Date: info.Date,
			Owner: info.Owner, Group: info.Group, Links: info.Links, LinkTarget: info.LinkTarget,
			Permissions: info.Permissions,
		}
		idx++
		switch info.Type {
		case "directory":
			row.Href = "/files/" + urlEncodedJoin(p.PathParam, info.Name) + "?view=" + url.QueryEscape(p.View)
			data.Directories = append(data.Directories, row)
		default:
			row.DownloadURL = "/file-manager/download-file/" + urlEncodedJoin(info.Name) + "?path_param=" + url.QueryEscape(p.PathParam)
			data.Files = append(data.Files, row)
		}
	}
	data.Rows = append(append([]ViewRow{}, data.Directories...), data.Files...)
	for i := range data.Rows {
		data.Rows[i].Index = i
	}

	if err := filesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FILEMANAGER - files template render error: %v", err)
	}
}

// UploadPageData is upload_file.html's template context.
type UploadPageData struct {
	web.LayoutData
	PathParam             string
	FilemanagerUploadSize int
}

func renderUploadPage(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string, uploadSizeMB int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Upload Files")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := UploadPageData{LayoutData: layout, PathParam: pathParam, FilemanagerUploadSize: uploadSizeMB}
	if err := uploadPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FILEMANAGER - upload template render error: %v", err)
	}
}

// EditorOption is one entry of edit_file.html's editor-switcher dropdown.
type EditorOption struct {
	Value, Label string
	Selected     bool
}

var editorLabels = map[string]string{
	"monaco": "Monaco", "ace": "Ace", "codemirror": "CodeMirror", "text": "Plain Text",
}

// EditFilePageData is edit_file.html's template context.
type EditFilePageData struct {
	web.LayoutData
	FilePath        string
	FilePathJSON    template.JS
	FileContent     string
	FileContentJSON template.JS
	Editor          string
	EditorLabel     string
	EditorOptions   []EditorOption
	Breadcrumbs     []Breadcrumb
	FullscreenQS    string
}

func renderEditFilePage(a *appctx.App, w http.ResponseWriter, r *http.Request, filePath, fileContent, editor string) {
	layout, _, err := web.BuildLayoutData(a, w, r, path.Base(filePath))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	label := editorLabels[editor]
	if label == "" {
		label = editorLabels["monaco"]
	}

	opts := make([]EditorOption, 0, 4)
	for _, v := range []string{"monaco", "ace", "codemirror", "text"} {
		opts = append(opts, EditorOption{Value: v, Label: editorLabels[v], Selected: v == editor})
	}

	fullscreenQS := ""
	if r.URL.Query().Get("fullscreen") == "true" {
		fullscreenQS = "&fullscreen=true"
	}

	data := EditFilePageData{
		LayoutData: layout, FilePath: filePath, FilePathJSON: jsonString(filePath),
		FileContent: fileContent, FileContentJSON: jsonString(fileContent),
		Editor: editor, EditorLabel: label, EditorOptions: opts,
		Breadcrumbs: buildBreadcrumbs(filePath), FullscreenQS: fullscreenQS,
	}
	if err := editFilePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FILEMANAGER - edit_file template render error: %v", err)
	}
}

func jsonString(s string) template.JS {
	b, err := json.Marshal(s)
	if err != nil {
		return "\"\""
	}
	return template.JS(b) //nolint:gosec // JSON-encoded string literal embedded in a <script> block, not raw HTML
}
