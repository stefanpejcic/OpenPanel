package inodes

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var inodesPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/inodes.html",
)

// Breadcrumb is one segment of the /inodes-explorer/... path trail.
type Breadcrumb struct {
	Name, Path string
	Last       bool
}

// buildBreadcrumbs mirrors diskusage's buildBreadcrumbs.
func buildBreadcrumbs(root, urlPath string) []Breadcrumb {
	trimmed := strings.Trim(urlPath, "/")
	var parts []string
	for _, p := range strings.Split(trimmed, "/") {
		if p != "" && p != root {
			parts = append(parts, p)
		}
	}
	crumbs := make([]Breadcrumb, len(parts))
	for i, part := range parts {
		crumbs[i] = Breadcrumb{
			Name: part,
			Path: "/" + root + "/" + strings.Join(parts[:i+1], "/"),
			Last: i == len(parts)-1,
		}
	}
	return crumbs
}

// Row is one folders_to_navigate table row.
type Row struct {
	Directory, Count, Href, FileManagerHref string
}

// parseInodesOutput splits each "<count> <path>" line of inodesOutput into
// a table row, skipping the "." directory row.
func parseInodesOutput(output, urlPath string) []Row {
	var rows []Row
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		count := fields[0]
		directory := strings.TrimLeft(strings.TrimPrefix(line, count), " \t")
		if directory == "." {
			continue
		}

		row := Row{Directory: directory, Count: count, Href: urlPath + url.PathEscape(directory) + "/"}
		if strings.HasPrefix(urlPath, "/inodes-explorer/files/") {
			row.FileManagerHref = strings.Replace(urlPath, "inodes-explorer/files", "files", 1) + url.PathEscape(directory)
		}
		rows = append(rows, row)
	}
	return rows
}

// InodesPageData is inodes.html's template context.
type InodesPageData struct {
	web.LayoutData
	Breadcrumbs []Breadcrumb
	Rows        []Row
	ShowUpLink  bool
	UpLink      string
}

func renderInodesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, inodesOutputText string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Inodes Explorer")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	urlPath := r.URL.Path
	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}

	data := InodesPageData{
		LayoutData:  layout,
		Breadcrumbs: buildBreadcrumbs("inodes-explorer", urlPath),
		Rows:        parseInodesOutput(inodesOutputText, urlPath),
	}
	if urlPath != "/inodes-explorer/" {
		data.ShowUpLink = true
		trimmed := strings.TrimSuffix(urlPath, "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			data.UpLink = trimmed[:idx]
		}
	}

	if err := inodesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("INODES - template render error: %v", err)
	}
}
