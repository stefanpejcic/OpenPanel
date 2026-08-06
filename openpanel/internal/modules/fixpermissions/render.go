package fixpermissions

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var fixPermissionsPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/fix_permissions.html",
)

// FixPermissionsPageData is fix_permissions.html's template context.
type FixPermissionsPageData struct {
	web.LayoutData
	Directories []string
}

func renderFixPermissionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, directories []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Fix Permissions")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FixPermissionsPageData{LayoutData: layout, Directories: directories}
	if err := fixPermissionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FIX_PERMISSIONS - template render error: %v", err)
	}
}
