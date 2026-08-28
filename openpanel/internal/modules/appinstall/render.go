package appinstall

import (
	"html/template"
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// kindDisplay is the template-only presentation detail for a Kind (icon
// markup, label text) that appinstall's logic code has no need for -
// separated out so Kind itself stays a plain data type.
type kindDisplay struct {
	Icon                template.HTML
	Label               string
	RunCommand          string
	RequiredExtension   string
	RequirementsLabel   string
	RequirementsTooltip string
}

// displayFor looks up the pre-built per-type display config - the actual
// icon/label/tooltip data lives in nodejs.go/python.go/ruby.go, split by
// type; this is just the tiny bit of glue needed since kindDisplay is a
// template-only concern that Kind itself (a plain data type) doesn't carry.
func displayFor(kind Kind) kindDisplay {
	switch kind.AppType {
	case NodeJS.AppType:
		return nodeJSDisplay
	case Ruby.AppType:
		return rubyDisplay
	case Java.AppType:
		return javaDisplay
	default:
		return pythonDisplay
	}
}

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

var installPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "manager/app_install.html")...)

// InstallPageData is manager/app_install.html's template context.
type InstallPageData struct {
	web.LayoutData
	Kind    Kind
	Display kindDisplay
	Domains []appctx.Domain
}

func renderInstallPage(kind Kind, a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, kind.Title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InstallPageData{LayoutData: layout, Kind: kind, Display: displayFor(kind), Domains: domains}
	if err := installPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("APPINSTALL - install template render error: %v", err)
	}
}
