package diskusage

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var diskUsagePage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/disk_usage.html",
)

// Breadcrumb is one segment of the /disk-usage/... path trail.
type Breadcrumb struct {
	Name, Path string
	Last       bool
}

// buildBreadcrumbs builds the breadcrumb trail for the current path: each crumb links to the cumulative path through itself
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

// parseDuOutput splits each "<size> <path>" line of `du` output into a table row
func parseDuOutput(output, urlPath string) []Row {
	base := escapePath(urlPath)
	var rows []Row
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count, directory, ok := splitDuLine(line)
		if !ok {
			continue
		}

		row := Row{Directory: directory, Count: count, Href: base + url.PathEscape(directory) + "/"}
		if strings.HasPrefix(base, "/disk-usage/files/") {
			row.FileManagerHref = strings.Replace(base, "disk-usage/files", "files", 1) + url.PathEscape(directory)
		}
		rows = append(rows, row)
	}
	return rows
}

// splitDuLine splits a du line at the tab, trimming spaces after the size cut leading spaces off folder names
func splitDuLine(line string) (size, name string, ok bool) {
	if size, name, ok = strings.Cut(line, "\t"); ok {
		return size, name, name != ""
	}
	// not du's own format, split at the first run of whitespace
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}
	return fields[0], strings.TrimLeft(strings.TrimPrefix(line, fields[0]), " "), true
}

// DiskUsagePageData is disk_usage.html's template context.
type DiskUsagePageData struct {
	web.LayoutData
	Breadcrumbs []Breadcrumb
	Rows        []Row
	ShowUpLink  bool
	UpLink      string
}

func renderDiskUsagePage(a *appctx.App, w http.ResponseWriter, r *http.Request, duOutput string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Disk Usage")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	urlPath := r.URL.Path
	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}

	data := DiskUsagePageData{
		LayoutData:  layout,
		Breadcrumbs: buildBreadcrumbs("disk-usage", urlPath),
		Rows:        parseDuOutput(duOutput, urlPath),
	}
	if urlPath != "/disk-usage/" {
		data.ShowUpLink = true
		trimmed := strings.TrimSuffix(escapePath(urlPath), "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			data.UpLink = trimmed[:idx]
		}
	}

	if err := diskUsagePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DISK_USAGE - template render error: %v", err)
	}
}
