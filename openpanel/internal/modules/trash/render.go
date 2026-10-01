package trash

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var trashPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/trash.html",
	"files/trash_partials.html",
)

// Row is one table_trash.html row, with the original_path display value (the '_data/'-relative shortening) precomputed since html/template has no string-split primitive
type Row struct {
	Entry
	DisplayOriginalPath string
}

// displayOriginalPath splits the original path on the first "_data/" substring, which for the usual "<context>_html_data/_data/<name>" shape actually falls inside "..._html_data/" rather than the later trash-volume "_data" folder - keeps a redundant "_data/" prefix, preserving existing display behavior exactly rather than "fixing" it
func displayOriginalPath(original string) string {
	if original == "" {
		return ""
	}
	if idx := strings.Index(original, "_data/"); idx != -1 {
		return "/var/www/html/" + original[idx+len("_data/"):]
	}
	return original
}

func toRow(e Entry) Row {
	return Row{Entry: e, DisplayOriginalPath: displayOriginalPath(e.OriginalPath)}
}

// TrashPageData is trash.html's template context.
type TrashPageData struct {
	web.LayoutData
	Title       string
	PathParam   string
	Directories []Row
	Files       []Row
	Rows        []Row
	// PurgeNote is shown under the heading when autopurge_trash is set to a number of days
	PurgeNote string
}

func renderTrashPage(a *appctx.App, w http.ResponseWriter, r *http.Request, title, pathParam string, filesInfo []Entry) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := TrashPageData{LayoutData: layout, Title: title, PathParam: pathParam}
	if days, convErr := strconv.Atoi(strings.TrimSpace(a.Config.Get("autopurge_trash", ""))); convErr == nil && days > 0 {
		data.PurgeNote = web.Tr(a, r, "Items in the trash are automatically deleted after %(days)s days.", "days", strconv.Itoa(days))
	}
	for _, info := range filesInfo {
		// .trash_restore is the trash's own bookkeeping, not something the user put there
		if info.Name == ".trash_restore" {
			continue
		}
		row := toRow(info)
		if info.Type == "directory" {
			data.Directories = append(data.Directories, row)
		} else {
			data.Files = append(data.Files, row)
		}
	}
	data.Rows = append(append([]Row{}, data.Directories...), data.Files...)

	if err := trashPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("TRASH - template render error: %v", err)
	}
}
