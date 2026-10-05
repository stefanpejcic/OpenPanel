package cmsapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// InstallPageData is manager/<slug>_install.html's template context
type InstallPageData struct {
	web.LayoutData
	Domains []appctx.Domain
	App     web.AppNames
}

// HandleInstallPage renders the install form and checks the plan's site limit on GET, POST goes to the app's Install stream
func (app *App) HandleInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := web.AtoiDefault(plan.WebsitesLimit, 0)
	websiteCount, _ := appkit.CountUserWebsites(a, userID)

	if websitesLimit != 0 && websiteCount >= websitesLimit {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/x-ndjson")
			flusher, canFlush := w.(http.Flusher)
			web.WriteNDJSON(w, flusher, canFlush, map[string]any{"error": "You have reached the maximum number of sites allowed." + plan.UpgradeMessage()})
			return
		}
		web.Flash(a, w, r, "warning", web.Tr(a, r, "You have reached the maximum number of sites allowed.%(upgrade_message)s", "upgrade_message", plan.UpgradeMessage()))
	} else if r.Method == http.MethodPost {
		app.Install(a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	layout, _, err := web.BuildLayoutData(a, w, r, "Install "+app.Name)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InstallPageData{LayoutData: layout, Domains: domains, App: web.NewAppNames(app.Slug, app.Name)}
	if err := app.installPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("%s - install template render error: %v", strings.ToUpper(app.Slug), err)
	}
}

// DecodeJSONForm reads the API's JSON body into form values for the listed string fields, missing ones become ""
func DecodeJSONForm(r *http.Request, fields ...string) (url.Values, error) {
	var body map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	form := url.Values{}
	for _, f := range fields {
		var v string
		if raw, ok := body[f]; ok {
			if err := json.Unmarshal(raw, &v); err != nil {
				return nil, err
			}
		}
		form.Set(f, v)
	}
	return form, nil
}

// HandleAPIInstall feeds the API's JSON body into the same install flow as the UI form
func (app *App) HandleAPIInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	form, err := DecodeJSONForm(r, app.InstallFields...)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if form.Get("domain_id") == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain_id is required"})
		return
	}
	app.HandleInstallPage(a, w, web.WithForm(r, form))
}

// HandleAPIRemove runs the UI remove handler with {site_id} as its "id" field and output=json
func (app *App) HandleAPIRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	cloned := web.WithForm(r, url.Values{"id": {r.PathValue("site_id")}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	app.remover()(a, w, cloned)
}

// ResolveSite looks up a site of type slug by id, returning its name and real install path (docroot plus subdirectory)
func ResolveSite(ctx context.Context, a *appctx.App, slug, siteID string) (siteName, docroot string, ok bool) {
	var rootDocroot sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = ?`, siteID, slug)
	if err := row.Scan(&siteName, &rootDocroot); err != nil || !rootDocroot.Valid || rootDocroot.String == "" {
		return "", "", false
	}
	docroot = rootDocroot.String
	if idx := strings.Index(siteName, "/"); idx != -1 {
		docroot = strings.TrimSuffix(rootDocroot.String, "/") + "/" + siteName[idx+1:]
	}
	return siteName, docroot, true
}

// ResolveSiteDomainDocroot is ResolveSite without the subdirectory, it returns the domain's own docroot
func ResolveSiteDomainDocroot(ctx context.Context, a *appctx.App, slug, siteID string) (siteName, docroot string, ok bool) {
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = ?`, siteID, slug)
	if err := row.Scan(&siteName, &docroot); err != nil {
		return "", "", false
	}
	return siteName, docroot, true
}

// SiteQuery wraps a domain/docroot query-param handler for the API, filling both from the {site_id} path value
func SiteQuery(slug string, h Handler) Handler {
	return func(a *appctx.App, w http.ResponseWriter, r *http.Request) {
		domain, docroot, ok := ResolveSite(r.Context(), a, slug, r.PathValue("site_id"))
		if !ok {
			web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
			return
		}
		q := r.URL.Query()
		q.Set("domain", domain)
		q.Set("docroot", docroot)
		r.URL.RawQuery = q.Encode()
		h(a, w, r)
	}
}

// SiteQueryPost is SiteQuery for handlers that want the domain's own docroot and a POST
func SiteQueryPost(slug string, h Handler) Handler {
	return func(a *appctx.App, w http.ResponseWriter, r *http.Request) {
		siteName, docroot, ok := ResolveSiteDomainDocroot(r.Context(), a, slug, r.PathValue("site_id"))
		if !ok {
			web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
			return
		}
		cloned := r.Clone(r.Context())
		cloned.Method = http.MethodPost
		q := cloned.URL.Query()
		q.Set("domain", siteName)
		q.Set("docroot", docroot)
		cloned.URL.RawQuery = q.Encode()
		h(a, w, cloned)
	}
}

// Install is the request state every install stream starts from
type Install struct {
	Ctx          context.Context
	UserID       int
	Username     string
	UserContext  string
	IP           string
	Emit         func(map[string]any)
	DomainID     string
	Domain       appkit.DomainRow
	Subdirectory string
}

// StartInstall runs the checks every install stream begins with and takes the per-user install lock, when ok the caller must defer appkit.RemoveLockFile(in.Username)
func StartInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) (*Install, bool) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { web.WriteNDJSON(w, flusher, canFlush, v) }

	in := &Install{Ctx: ctx, UserID: userID, Username: currentUsername, UserContext: userContext, IP: reqip.ClientIP(r), Emit: emit}
	domainID := r.FormValue("domain_id")
	if domainID == "" {
		emit(map[string]any{"error": "Missing required field: domain"})
		return nil, false
	}
	in.DomainID = domainID

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	if err := appkit.CreateLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return nil, false
	}
	ok := false
	defer func() {
		if !ok {
			appkit.RemoveLockFile(currentUsername)
		}
	}()

	dom, found, dbErr := appkit.LookupDomainByID(ctx, a, domainID)
	if dbErr != nil {
		emit(map[string]any{"error": "An error occurred fetching docroot for domain from database."})
		return nil, false
	}
	if !found {
		emit(map[string]any{"error": "Domain not found"})
		return nil, false
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, dom.DomainURL) {
		return nil, false
	}
	in.Domain = dom

	emit(map[string]any{"status": "Validating provided data"})
	in.Subdirectory = strings.ToLower(strings.ReplaceAll(r.FormValue("subdirectory"), " ", ""))
	if !IsValidSubdirectory(in.Subdirectory) {
		emit(map[string]any{"error": "Invalid subdirectory."})
		return nil, false
	}
	ok = true
	return in, true
}
