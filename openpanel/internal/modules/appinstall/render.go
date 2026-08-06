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

const pythonIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#3776AB" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-brand-python"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 9h-7a2 2 0 0 0 -2 2v4a2 2 0 0 0 2 2h3" /><path d="M12 15h7a2 2 0 0 0 2 -2v-4a2 2 0 0 0 -2 -2h-3" /><path d="M8 9v-4a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v5a2 2 0 0 1 -2 2h-4a2 2 0 0 0 -2 2v5a2 2 0 0 0 2 2h4a2 2 0 0 0 2 -2v-4" /><path d="M11 6l0 .01" /><path fill="#FFA500" d="M13 18l0 .01" /></svg>`

const nodejsIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#339933" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-brand-nodejs"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 9v8.044a2 2 0 0 1 -2.996 1.734l-1.568 -.9a3 3 0 0 1 -1.436 -2.561v-6.635a3 3 0 0 1 1.436 -2.56l6 -3.667a3 3 0 0 1 3.128 0l6 3.667a3 3 0 0 1 1.436 2.561v6.634a3 3 0 0 1 -1.436 2.56l-6 3.667a3 3 0 0 1 -3.128 0" /><path d="M17 9h-3.5a1.5 1.5 0 0 0 0 3h2a1.5 1.5 0 0 1 0 3h-3.5" /></svg>`

func displayFor(kind Kind) kindDisplay {
	if kind.PyOrNode == "NODE" {
		return kindDisplay{
			Icon:  template.HTML(nodejsIconSVG), //nolint:gosec // static, server-defined markup, not user input
			Label: "NodeJS", RunCommand: "node", RequiredExtension: ".js",
			RequirementsLabel:   "Run NPM install before starting the app",
			RequirementsTooltip: "When enabled, this option will first run npm install using the package.json file, then launch the application. If the application is already built, you can skip this option.",
		}
	}
	return kindDisplay{
		Icon:  template.HTML(pythonIconSVG), //nolint:gosec // static, server-defined markup, not user input
		Label: "Python", RunCommand: "python", RequiredExtension: ".py",
		RequirementsLabel:   "Run PIP install before starting the app",
		RequirementsTooltip: "When enabled, this option will first run pip install using the requirements.txt file, then launch the application. If the application is already built, you can skip this option.",
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
