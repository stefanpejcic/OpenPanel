// Package pluginpage renders a page for any installed plugin that declares
// a "link" in its readme.txt - the plugin's own binary supplies the page's
// content via its "page" subcommand, and this package wraps that content in
// the standard app chrome (sidebar/topbar/flashes). This is what lets a
// plugin add a full page to the UI (e.g. "/new-plugin") without OpenPanel
// knowing anything about that plugin ahead of time - see the plugin
// boilerplate repo for the exact "page" subcommand contract.
package pluginpage

import (
	"html/template"
	"log"
	"net/http"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	corePlugins "gist.github.com/stefanpejcic/openpanel/internal/core/plugins"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

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

var pluginPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "plugin/page.html")...)

// PageData is plugin/page.html's template context.
type PageData struct {
	web.LayoutData
	PluginContent template.HTML
}

// TryServe renders the page for the installed plugin whose readme.txt "link" matches r.URL.Path, if any, and reports whether it did. A false return means no plugin claims that path and nothing was written to w, so the caller should fall through to its own 404 - a true return means a response was written (successfully, or a login redirect/403/error), so the caller must not write anything more.
func TryServe(a *appctx.App, w http.ResponseWriter, r *http.Request) bool {
	plugin, ok := corePlugins.FindByLink(corePlugins.BaseDir, r.URL.Path)
	if !ok {
		return false
	}
	name := plugin["folder"]
	auth.RequireLogin(a, name)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serve(a, w, r, plugin)
	})).ServeHTTP(w, r)
	return true
}

func serve(a *appctx.App, w http.ResponseWriter, r *http.Request, plugin corePlugins.Plugin) {
	name := plugin["folder"]
	title := plugin["title"]
	if title == "" {
		title = name
	}

	out, err := corePlugins.Exec(r.Context(), corePlugins.BaseDir, name, 5*time.Second, "page")
	if err != nil {
		log.Printf("plugin %q: page call failed: %v", name, err)
		http.Error(w, "This plugin's page could not be loaded.", http.StatusBadGateway)
		return
	}

	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PageData{
		LayoutData: layout,
		// plugin output is server-installed executable code the admin already explicitly chose to trust - same trust level as any other exec'd plugin, not user input
		PluginContent: template.HTML(out), //nolint:gosec
	}
	if err := pluginPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("plugin %q: page template render error: %v", name, err)
	}
}
