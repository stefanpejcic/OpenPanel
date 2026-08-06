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

// buildBreadcrumbs builds the breadcrumb trail for the current path: each
// crumb links to the cumulative path through itself.
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

// parseDuOutput splits each "<size> <path>" line of `du` output into a
// table row.
func parseDuOutput(output, urlPath string) []Row {
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
		directory := strings.TrimPrefix(line, count)
		directory = strings.TrimLeft(directory, " \t")

		row := Row{Directory: directory, Count: count, Href: urlPath + url.PathEscape(directory) + "/"}
		if strings.HasPrefix(urlPath, "/disk-usage/files/") {
			row.FileManagerHref = strings.Replace(urlPath, "disk-usage/files", "files", 1) + url.PathEscape(directory)
		}
		rows = append(rows, row)
	}
	return rows
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
		trimmed := strings.TrimSuffix(urlPath, "/")
		if idx := strings.LastIndex(trimmed, "/"); idx != -1 {
			data.UpLink = trimmed[:idx]
		}
	}

	if err := diskUsagePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DISK_USAGE - template render error: %v", err)
	}
}
