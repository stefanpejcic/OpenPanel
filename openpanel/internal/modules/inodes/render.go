package inodes

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
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
		escaped := make([]string, i+1)
		for j, p := range parts[:i+1] {
			escaped[j] = url.PathEscape(p)
		}
		crumbs[i] = Breadcrumb{
			Name: part,
			Path: "/" + root + "/" + strings.Join(escaped, "/"),
			Last: i == len(parts)-1,
		}
	}
	return crumbs
}

// docrootVolume is the account's html_data volume under the home dir, the only folder the File Manager shows
var docrootVolume = regexp.MustCompile(`^/inodes-explorer/docker-data/volumes/[^/]+_html_data/_data/`)

// fileManagerBase is the /files/ link for the folder at base, the old /inodes-explorer/files/ check pointed at /home/<ctx>/files which the File Manager never shows
func fileManagerBase(base string) (string, bool) {
	prefix := docrootVolume.FindString(base)
	if prefix == "" {
		return "", false
	}
	return "/files/" + strings.TrimPrefix(base, prefix), true
}

// escapePath escapes each segment of the decoded URL path, html/template leaves # ? and %XX in hrefs as they are
func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// Row is one folders_to_navigate table row.
type Row struct {
	Directory, Count, Href, FileManagerHref string
}

// parseInodesOutput splits each "<count> <path>" line of inodesOutput into a table row, skipping the "." directory row
func parseInodesOutput(output, urlPath string) []Row {
	base := escapePath(urlPath)
	var rows []Row
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// inodesOutput writes "<count> <name>", trimming more than that one space cut leading spaces off folder names
		count, directory, ok := strings.Cut(line, " ")
		if !ok || directory == "" || directory == "." {
			continue
		}

		row := Row{Directory: directory, Count: count, Href: base + url.PathEscape(directory) + "/"}
		if fmBase, ok := fileManagerBase(base); ok {
			row.FileManagerHref = fmBase + url.PathEscape(directory)
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
	layout, _, err := web.BuildLayoutData(a, w, r, "Inode Usage")
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
		trimmed := strings.TrimSuffix(escapePath(urlPath), "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			data.UpLink = trimmed[:idx]
		}
	}

	if err := inodesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("INODES - template render error: %v", err)
	}
}
