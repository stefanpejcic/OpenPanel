package websites

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var websiteBuilderInstallPage = loadPage("manager/websitebuilder_install.html", "domains/_shared.html")
var grapesJSEditorFragment = web.MustLoadFragment("manager/grapejs_editor.html")

func atoiDefaultWB(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// WebsiteBuilderInstallPageData is manager/websitebuilder_install.html's
// template context.
type WebsiteBuilderInstallPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderWebsiteBuilderInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Website")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WebsiteBuilderInstallPageData{LayoutData: layout, Domains: domains}
	if err := websiteBuilderInstallPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WEBSITES - websitebuilder_install template render error: %v", err)
	}
}

// renderGrapesJSEditor renders a standalone page (no base.html layout)
// embedding the GrapesJS builder with the site's current HTML/CSS.
func renderGrapesJSEditor(a *appctx.App, w http.ResponseWriter, r *http.Request, currentDomain, html, css string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "GrapesJS Website Editor")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Favicon": layout.Favicon, "BrandName": layout.BrandName, "CSRFToken": layout.CSRFToken,
		"HTML": html, "CSS": css, "T": layout.T,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := grapesJSEditorFragment.ExecuteTemplate(w, "grapejs_editor", data); err != nil {
		log.Printf("WEBSITES - grapejs_editor template render error: %v", err)
	}
}

func writeNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

// handleWebsiteBuilderInstall serves the "create website" form on GET and
// dispatches to createHTMLSiteStream on POST, after checking the plan's
// site limit.
func handleWebsiteBuilderInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	injectedData, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefaultWB(plan.WebsitesLimit, 0)
	userWebsites, _ := dashboard.GetUserWebsites(a, ctx, userID)

	if websitesLimit != 0 && len(userWebsites) >= websitesLimit {
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed.")
	} else if r.Method == http.MethodPost {
		createHTMLSiteStream(a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderWebsiteBuilderInstallPage(a, w, r, domains)
}

// createHTMLSiteStream creates a new website-builder site, streaming
// newline-delimited JSON status updates as each step completes.
func createHTMLSiteStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

	ipAddress := reqip.ClientIP(r)
	_ = r.ParseForm()
	domainID := r.Form.Get("domain_id")

	krompirPath := "/etc/openpanel/openpanel/core/users/" + currentUsername + "/krompir.lock"
	if wErr := os.WriteFile(krompirPath, nil, 0o644); wErr != nil {
		emit(map[string]any{"error": "Error creating " + krompirPath + ": " + wErr.Error()})
		return
	}

	var selectedDomain, docroot, phpVersion string
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_id = ?", domainID)
	if scanErr := row.Scan(&selectedDomain, &docroot, &phpVersion); scanErr != nil {
		emit(map[string]any{"error": "Domain not found"})
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	subdirectory := strings.ToLower(r.Form.Get("subdirectory"))
	installPath := docroot
	if subdirectory != "" {
		subdirectory = strings.ReplaceAll(subdirectory, " ", "")
		if strings.Contains(subdirectory, "..") || strings.HasPrefix(subdirectory, "/") {
			emit(map[string]any{"error": "Invalid subdirectory."})
			return
		}
		installPath = docroot + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	cleanedInstallPath := strings.TrimPrefix(installPath, "/var/www/html/")
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + cleanedInstallPath

	filesToCheck := []struct{ name, errMsg string }{
		{".htaccess", "Error: .htaccess file already exists. Website creation cannot proceed!"},
		{"index.html", "Error: index.html file already exists in " + installPath + " - Website creation cannot proceed!"},
		{"index.php", "Error: index.php file already exists in " + installPath + " - Website creation cannot proceed!"},
	}
	for _, fc := range filesToCheck {
		if _, statErr := os.Stat(path.Join(volume, fc.name)); statErr == nil {
			emit(map[string]any{"error": fc.errMsg})
			return
		}
	}

	emit(map[string]any{"status": "Creating files.."})
	if mkErr := os.MkdirAll(volume, 0o755); mkErr != nil {
		emit(map[string]any{"error": mkErr.Error()})
		return
	}

	const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
</head>
<body>
<body></body>
<style>* { box-sizing: border-box; } body {margin: 0;}</style>
</body>
</html>`
	if wErr := os.WriteFile(path.Join(volume, "index.html"), []byte(htmlContent), 0o644); wErr != nil {
		emit(map[string]any{"error": wErr.Error()})
		return
	}

	const cssContent = `* { box-sizing: border-box; } body {margin: 0;}`
	if wErr := os.WriteFile(path.Join(volume, "style.css"), []byte(cssContent), 0o644); wErr != nil {
		emit(map[string]any{"error": wErr.Error()})
		return
	}

	emit(map[string]any{"status": "Setting files permissions to '755'"})
	_ = exec.CommandContext(ctx, "chmod", "-R", "755", volume).Run()

	emit(map[string]any{"status": "Setting files owner to '" + userContext + "'"})
	if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = exec.CommandContext(ctx, "chown", "-R", uidStr+":"+uidStr, volume).Run()
	}

	emit(map[string]any{"status": "Saving website information to SiteManager"})
	if _, insertErr := a.DB.ExecContext(ctx, "INSERT INTO sites (site_name, domain_id, type) VALUES (?, ?, ?)", selectedDomain, domainID, "websitebuilder"); insertErr != nil {
		emit(map[string]any{"error": "An error occurred while saving data: " + insertErr.Error()})
		return
	}

	_ = a.Cache.Delete(ctx, cacheKeyUserWebsites(userID))

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Website Builder on "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "Website created successfully on "+selectedDomain)

	emit(map[string]any{"status": "Website creation completed!"})

	_ = os.Remove(krompirPath)
}

func cacheKeyUserWebsites(userID int) string {
	return "get_user_websites:" + strconv.Itoa(userID)
}
