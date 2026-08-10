package websites

import (
	"database/sql"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterSitesAPI wires the /api/sites routes onto mux. Several of these
// routes share a <domain> prefix with a literal suffix (e.g.
// /api/sites/{domain}/safebrowsing, /temporary-link, /visitors, /wp-info),
// and a domain itself may contain
// slashes (subfolder installs) - so the most specific literal suffix has
// to win over treating the whole tail as the domain. Go's http.ServeMux
// requires a "{...}" wildcard to be the final segment, so a single
// "{rest...}" catch-all per method is registered instead, and
// apiSitesGetDispatch/apiSitesPostDispatch manually strip the known
// literal suffixes off the tail to resolve the real route. apiregistry.Add
// still records each logical route separately so /api/endpoints lists them
// individually.
func RegisterSitesAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "websites", "GET /api/sites", func(w http.ResponseWriter, r *http.Request) { apiSitesList(a, w, r) })

	apiregistry.Add("GET /api/sites/{domain}")
	apiregistry.Add("GET /api/sites/{domain}/safebrowsing")
	apiregistry.Add("GET /api/sites/{domain}/pagespeed")
	apiregistry.Add("GET /api/sites/{domain}/wp-vulnerability")
	apiregistry.Add("GET /api/sites/{domain}/temporary-link")
	apiregistry.Add("GET /api/sites/{domain}/visitors")
	apiregistry.Add("GET /api/sites/{domain}/wp-info")
	mux.Handle("GET /api/sites/{rest...}", auth.RequireAPI(a, "websites")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiSitesGetDispatch(a, w, r) })))

	apiregistry.Add("POST /api/sites/{domain}/pagespeed")
	apiregistry.Add("POST /api/sites/{domain}/wp-vulnerability")
	mux.Handle("POST /api/sites/{rest...}", auth.RequireAPI(a, "websites")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiSitesPostDispatch(a, w, r) })))
}

// apiSitesGetDispatch resolves the literal-suffix-wins routing for
// GET /api/sites/{domain}[/safebrowsing|/pagespeed|/wp-vulnerability|/temporary-link|/visitors|/wp-info].
func apiSitesGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/safebrowsing"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/safebrowsing"))
		apiSafeBrowsing(a, w, r)
	case strings.HasSuffix(rest, "/pagespeed"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/pagespeed"))
		apiPagespeedGet(a, w, r)
	case strings.HasSuffix(rest, "/wp-vulnerability"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/wp-vulnerability"))
		apiWPVulnerabilityGet(a, w, r)
	case strings.HasSuffix(rest, "/temporary-link"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/temporary-link"))
		apiTemporaryLink(a, w, r)
	case strings.HasSuffix(rest, "/visitors"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/visitors"))
		apiVisitors(a, w, r)
	case strings.HasSuffix(rest, "/wp-info"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/wp-info"))
		apiWPInfo(a, w, r)
	default:
		r.SetPathValue("domain", rest)
		apiSiteDetail(a, w, r)
	}
}

// apiSitesPostDispatch replicates the same routing for the two POST-only
// suffix routes; nothing else is registered for POST /api/sites/*.
func apiSitesPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/pagespeed"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/pagespeed"))
		apiPagespeedRefresh(a, w, r)
	case strings.HasSuffix(rest, "/wp-vulnerability"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/wp-vulnerability"))
		apiWPVulnerabilityScan(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

// apiSitesList returns every site owned by the current user.
func apiSitesList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT site_name, domain_id, admin_email, version, created_date, type, container, ports
		FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)`, userID)
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	defer rows.Close()

	type siteEntry struct {
		SiteName    *string `json:"site_name"`
		DomainID    *int    `json:"domain_id"`
		AdminEmail  *string `json:"admin_email"`
		Version     *string `json:"version"`
		CreatedDate *string `json:"created_date"`
		Type        *string `json:"type"`
		Container   *string `json:"container"`
		Ports       *string `json:"ports"`
	}
	sites := []siteEntry{}
	for rows.Next() {
		var (
			siteName, adminEmail, version, createdDate, typ, container, ports sql.NullString
			domainID                                                          sql.NullInt64
		)
		if scanErr := rows.Scan(&siteName, &domainID, &adminEmail, &version, &createdDate, &typ, &container, &ports); scanErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		sites = append(sites, siteEntry{
			SiteName: nullableStr(siteName), DomainID: nullableInt(domainID), AdminEmail: nullableStr(adminEmail),
			Version: nullableStr(version), CreatedDate: nullableStr(createdDate), Type: nullableStr(typ),
			Container: nullableStr(container), Ports: nullableStr(ports),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"sites": sites, "count": len(sites)})
}

func nullableStr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func nullableInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

// apiSiteDetail returns the stored row for a single site the caller owns.
func apiSiteDetail(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	row := a.DB.QueryRowContext(ctx, `
		SELECT id, site_name, domain_id, admin_email, version, created_date, type, container, path, ports
		FROM sites WHERE site_name = ? LIMIT 1`, domain)

	var (
		id                                                                      sql.NullInt64
		siteName, adminEmail, version, createdDate, typ, container, path, ports sql.NullString
		domainID                                                                sql.NullInt64
	)
	if scanErr := row.Scan(&id, &siteName, &domainID, &adminEmail, &version, &createdDate, &typ, &container, &path, &ports); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": scanErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id": nullableInt(id), "site_name": nullableStr(siteName), "domain_id": nullableInt(domainID),
		"admin_email": nullableStr(adminEmail), "version": nullableStr(version), "created_date": nullableStr(createdDate),
		"type": nullableStr(typ), "container": nullableStr(container), "path": nullableStr(path), "ports": nullableStr(ports),
	})
}

// apiSafeBrowsing exposes the same cached lookup and response shape as the
// UI's handleGoogleSafeBrowsing, just without the redirect/HTML wrapping
// around ownership failures.
func apiSafeBrowsing(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	result, err := safeBrowsingData(ctx, a, domain)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to contact Google Safe Browsing API", "details": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// apiPagespeedGet returns the cached PageSpeed report for a domain, or a
// "no data yet" message if a scan hasn't been run.
func apiPagespeedGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if !websiteParamRE.MatchString(domain) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid website parameter"})
		return
	}

	filename := strings.ReplaceAll(strings.ReplaceAll(domain, "://", "_"), "/", "_") + ".json"
	filePath := filepath.Join("/etc/openpanel/openpanel/websites", filename)

	if content, readErr := os.ReadFile(filePath); readErr == nil {
		writeRawJSON(w, http.StatusOK, content)
		return
	}

	out, runErr := exec.CommandContext(ctx, "opencli", "websites-pagespeed", domain).CombinedOutput()
	message := strings.TrimSpace(string(out))
	if message == "" {
		message = "No data yet"
	}
	status := http.StatusOK
	if runErr != nil {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, map[string]string{"message": message})
}

// apiPagespeedRefresh kicks off a PageSpeed scan in the background and
// returns immediately.
func apiPagespeedRefresh(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if !websiteParamRE.MatchString(domain) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid website parameter"})
		return
	}

	cmd := exec.Command("opencli", "websites-pagespeed", domain)
	if startErr := cmd.Start(); startErr == nil {
		go func() { _ = cmd.Wait() }()
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "initiated PageSpeed refresh for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "PageSpeed data gathering started"})
}

// apiWPVulnerabilityGet returns the cached WordPress vulnerability report
// for a domain, triggering a scan and waiting briefly for it if no cached
// result exists yet.
func apiWPVulnerabilityGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	filename := strings.ReplaceAll(strings.ReplaceAll(domain, "://", "_"), "/", "_") + ".json"
	filePath := filepath.Join("/etc/openpanel/wordpress/vulnerability", filename)

	if content, readErr := os.ReadFile(filePath); readErr == nil {
		writeRawJSON(w, http.StatusOK, content)
		return
	}

	_ = exec.CommandContext(ctx, "opencli", "websites-vulnerability", domain).Run()
	time.Sleep(2 * time.Second)

	data := map[string]any{"core_version_": map[string]any{}, "plugin_": map[string]any{}, "theme_": map[string]any{}}
	if content, readErr := os.ReadFile(filePath); readErr == nil {
		writeRawJSON(w, http.StatusOK, content)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// apiWPVulnerabilityScan runs a WordPress vulnerability scan synchronously
// and returns its result.
func apiWPVulnerabilityScan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	cmd := exec.CommandContext(ctx, "opencli", "websites-vulnerability", domain)
	runErr := cmd.Run()
	returnCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			returnCode = exitErr.ExitCode()
		} else {
			returnCode = -1
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "scanned WP vulnerabilities for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"message": "Scan completed for " + domain, "returncode": returnCode})
}

func writeRawJSON(w http.ResponseWriter, status int, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}
