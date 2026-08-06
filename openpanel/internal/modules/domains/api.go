package domains

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
)

// RegisterAPI wires the domains REST endpoints onto mux. Every sub-resource
// shares a {domain} prefix with a literal suffix (e.g.
// /api/domains/{domain}/status) - Go's http.ServeMux requires a "{...}"
// wildcard to be the final segment, so each verb gets one "{rest...}"
// catch-all and the dispatch funcs below strip the known suffix by hand to
// route to the right handler. apiregistry.Add still records each logical
// route separately for /api/endpoints.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "domains", "GET /api/domains", func(w http.ResponseWriter, r *http.Request) { apiDomainsList(a, w, r) })
	apiregistry.Handle(mux, a, "domains", "POST /api/domains", func(w http.ResponseWriter, r *http.Request) { apiDomainsCreate(a, w, r) })

	apiregistry.Add("POST /api/domains/{domain}/suspend")
	apiregistry.Add("POST /api/domains/{domain}/unsuspend")
	apiregistry.Add("POST /api/domains/{domain}/ssl")
	apiregistry.Add("POST /api/domains/{domain}/dns/records")
	apiregistry.Add("POST /api/domains/{domain}/dns/restart")
	mux.Handle("POST /api/domains/{rest...}", auth.RequireAPI(a, "domains")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDomainsPostDispatch(a, w, r) })))

	apiregistry.Add("GET /api/domains/{domain}/status")
	apiregistry.Add("GET /api/domains/{domain}/docroot")
	apiregistry.Add("GET /api/domains/{domain}/redirect")
	apiregistry.Add("GET /api/domains/{domain}/ssl")
	apiregistry.Add("GET /api/domains/{domain}/vhost")
	apiregistry.Add("GET /api/domains/{domain}/dns")
	apiregistry.Add("GET /api/domains/{domain}/dns/export")
	apiregistry.Add("GET /api/domains/{domain}/logs")
	mux.Handle("GET /api/domains/{rest...}", auth.RequireAPI(a, "domains")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDomainsGetDispatch(a, w, r) })))

	apiregistry.Add("PUT /api/domains/{domain}/docroot")
	apiregistry.Add("PUT /api/domains/{domain}/redirect")
	apiregistry.Add("PUT /api/domains/{domain}/vhost")
	apiregistry.Add("PUT /api/domains/{domain}/dns")
	apiregistry.Add("PUT /api/domains/{domain}/dns/records/{row_id}")
	mux.Handle("PUT /api/domains/{rest...}", auth.RequireAPI(a, "domains")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDomainsPutDispatch(a, w, r) })))

	apiregistry.Add("DELETE /api/domains/{domain}")
	apiregistry.Add("DELETE /api/domains/{domain}/redirect")
	apiregistry.Add("DELETE /api/domains/{domain}/dns/records/{row_id}")
	mux.Handle("DELETE /api/domains/{rest...}", auth.RequireAPI(a, "domains")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiDomainsDeleteDispatch(a, w, r) })))
}

func writeAPIDomainsJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiOwnDomainOr403Domains(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, domain string) bool {
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeAPIDomainsJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain."})
		return false
	}
	return true
}

// splitTrailingRowID splits "<domain>/dns/records/<row_id>" into its parts.
func splitTrailingRowID(rest string) (domain, rowID string, ok bool) {
	const marker = "/dns/records/"
	idx := strings.LastIndex(rest, marker)
	if idx == -1 {
		return "", "", false
	}
	tail := rest[idx+len(marker):]
	if tail == "" || !isAllDigitsDomains(tail) {
		return "", "", false
	}
	return rest[:idx], tail, true
}

func isAllDigitsDomains(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// apiDomainsGetDispatch dispatches GET /api/domains/{domain}[/status|/docroot|/redirect|/ssl|/vhost|/dns|/dns/export|/logs].
func apiDomainsGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/status"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/status"))
		apiDomainsStatus(a, w, r)
	case strings.HasSuffix(rest, "/docroot"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/docroot"))
		apiDomainsGetDocroot(a, w, r)
	case strings.HasSuffix(rest, "/redirect"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/redirect"))
		apiDomainsGetRedirect(a, w, r)
	case strings.HasSuffix(rest, "/ssl"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/ssl"))
		apiDomainsGetSSL(a, w, r)
	case strings.HasSuffix(rest, "/vhost"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/vhost"))
		apiDomainsGetVhost(a, w, r)
	case strings.HasSuffix(rest, "/dns/export"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/dns/export"))
		apiDomainsExportDNS(a, w, r)
	case strings.HasSuffix(rest, "/dns"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/dns"))
		apiDomainsGetDNS(a, w, r)
	case strings.HasSuffix(rest, "/logs"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/logs"))
		apiDomainsLogs(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

// apiDomainsPostDispatch dispatches POST /api/domains/{domain}[/suspend|/unsuspend|/ssl|/dns/records|/dns/restart].
func apiDomainsPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/suspend"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/suspend"))
		apiDomainsSuspend(a, w, r)
	case strings.HasSuffix(rest, "/unsuspend"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/unsuspend"))
		apiDomainsUnsuspend(a, w, r)
	case strings.HasSuffix(rest, "/ssl"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/ssl"))
		apiDomainsConfigureSSL(a, w, r)
	case strings.HasSuffix(rest, "/dns/records"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/dns/records"))
		apiDomainsAddDNSRecord(a, w, r)
	case strings.HasSuffix(rest, "/dns/restart"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/dns/restart"))
		apiDomainsRestartDNS(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

// apiDomainsPutDispatch dispatches PUT /api/domains/{domain}[/docroot|/redirect|/vhost|/dns|/dns/records/{row_id}].
func apiDomainsPutDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if domain, rowID, ok := splitTrailingRowID(rest); ok {
		r.SetPathValue("domain", domain)
		r.SetPathValue("row_id", rowID)
		apiDomainsUpdateDNSRecord(a, w, r)
		return
	}
	switch {
	case strings.HasSuffix(rest, "/docroot"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/docroot"))
		apiDomainsSetDocroot(a, w, r)
	case strings.HasSuffix(rest, "/redirect"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/redirect"))
		apiDomainsSetRedirect(a, w, r)
	case strings.HasSuffix(rest, "/vhost"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/vhost"))
		apiDomainsSetVhost(a, w, r)
	case strings.HasSuffix(rest, "/dns"):
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/dns"))
		apiDomainsSaveDNS(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

// apiDomainsDeleteDispatch dispatches DELETE /api/domains/{domain}[/redirect|/dns/records/{row_id}].
func apiDomainsDeleteDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if domain, rowID, ok := splitTrailingRowID(rest); ok {
		r.SetPathValue("domain", domain)
		r.SetPathValue("row_id", rowID)
		apiDomainsDeleteDNSRecord(a, w, r)
		return
	}
	if strings.HasSuffix(rest, "/redirect") {
		r.SetPathValue("domain", strings.TrimSuffix(rest, "/redirect"))
		apiDomainsDeleteRedirect(a, w, r)
		return
	}
	r.SetPathValue("domain", rest)
	apiDomainsDelete(a, w, r)
}

// ── List & create ────────────────────────────────────────────────────────

// apiDomainsList lists all domains for this account.
func apiDomainsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainData, err := domainsWithSites(ctx, a, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	type apiDomainEntry struct {
		Domain         string `json:"domain"`
		Docroot        string `json:"docroot"`
		PHPVersion     string `json:"php_version"`
		SiteCount      int    `json:"site_count"`
		RedirectURL    string `json:"redirect_url"`
		SSL            string `json:"ssl"`
		Status         string `json:"status"`
		SuspendComment string `json:"suspend_comment"`
		HasDNSZone     bool   `json:"has_dns_zone"`
	}

	result := make([]apiDomainEntry, 0, len(domainData))
	for _, d := range domainData {
		status := isRewriteCondEnabled(ctx, a, d.DomainURL)
		result = append(result, apiDomainEntry{
			Domain: d.DomainURL, Docroot: d.Docroot, PHPVersion: d.PHPVersion, SiteCount: d.SiteCount,
			RedirectURL: getRedirectURL(d.DomainURL), SSL: status.HTTPS, Status: status.Suspended,
			SuspendComment: status.SuspendComment, HasDNSZone: apiFileExists("/etc/bind/zones/" + d.DomainURL + ".zone"),
		})
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domains": result, "total": len(result)})
}

// apiDomainsCreate adds a new domain.
func apiDomainsCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Domain             string `json:"domain"`
		Docroot            string `json:"docroot"`
		SkipContainers     string `json:"skip_containers"`
		HsEd25519PublicKey string `json:"hs_ed25519_public_key"`
		HsEd25519SecretKey string `json:"hs_ed25519_secret_key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	domainURL := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(body.Domain), "."))
	docroot := strings.TrimSpace(body.Docroot)
	if docroot == "" {
		docroot = "/var/www/html/"
	}

	if domainURL == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Domain name is required."})
		return
	}

	resolved, ok := resolveUnderVarWWWHTML(docroot)
	if !ok {
		resolved = "/var/www/html/"
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	domainsLimit := 0
	if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
		domainsLimit = atoiDefault(plan.DomainsLimit, 0)
	}
	if domainsLimit != 0 {
		existing, _ := a.AllDomainsForUser(ctx, userID)
		urls := make([]appctx.Domain, len(existing))
		for i, d := range existing {
			urls[i] = appctx.Domain{DomainURL: d.DomainURL}
		}
		mains, _ := appctx.Categorize(urls)
		if len(mains) >= domainsLimit {
			writeAPIDomainsJSON(w, http.StatusConflict, map[string]string{"error": "Domain limit reached for your hosting plan."})
			return
		}
	}

	args := []string{"domains-add", domainURL, currentUsername, "--docroot", resolved}
	if body.SkipContainers != "" {
		args = append(args, body.SkipContainers)
	}
	if strings.HasSuffix(domainURL, ".onion") {
		if body.HsEd25519PublicKey == "" || body.HsEd25519SecretKey == "" {
			writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Both hs_ed25519_public_key and hs_ed25519_secret_key are required for .onion domains."})
			return
		}
		args = append(args, "--hs_ed25519_public_key", body.HsEd25519PublicKey, "--hs_ed25519_secret_key", body.HsEd25519SecretKey)
	}

	cctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	out, cmdErr := exec.CommandContext(cctx, "opencli", args...).CombinedOutput()
	if cctx.Err() == context.DeadlineExceeded {
		writeAPIDomainsJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "Domain creation timed out. The domain may still be provisioning."})
		return
	}
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add domain.", "output": string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "added domain "+domainURL+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusCreated, map[string]string{"domain": domainURL, "docroot": resolved, "output": strings.TrimSpace(string(out))})
}

// apiDomainsDelete removes a domain.
func apiDomainsDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	out, cmdErr := exec.CommandContext(cctx, "opencli", "domains-delete", domain).CombinedOutput()
	if cctx.Err() == context.DeadlineExceeded {
		writeAPIDomainsJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "Deletion timed out."})
		return
	}
	_ = cmdErr
	if !strings.Contains(strings.ToLower(string(out)), "deleted successfully") {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete domain.", "output": string(out)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted domain "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "deleted": true})
}

// ── Status ───────────────────────────────────────────────────────────────

// apiDomainsStatus returns a domain's suspend/active status.
func apiDomainsStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	status := isRewriteCondEnabled(ctx, a, domain)
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "ssl": status.HTTPS, "status": status.Suspended,
		"suspend_comment": status.SuspendComment, "redirect_url": getRedirectURL(domain),
	})
}

// ── Suspend / unsuspend ─────────────────────────────────────────────────

// apiDomainsSuspend suspends a domain.
func apiDomainsSuspend(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-suspend", domain).CombinedOutput()
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	invalidateRewriteCondCache(ctx, a, domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "suspended domain "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "suspended": true, "output": strings.TrimSpace(string(out))})
}

// apiDomainsUnsuspend reactivates a suspended domain.
func apiDomainsUnsuspend(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-unsuspend", domain).CombinedOutput()
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	invalidateRewriteCondCache(ctx, a, domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "unsuspended domain "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "suspended": false, "output": strings.TrimSpace(string(out))})
}

// ── Docroot ──────────────────────────────────────────────────────────────

// apiDomainsGetDocroot returns a domain's document root.
func apiDomainsGetDocroot(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-docroot", domain).CombinedOutput()
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "docroot": strings.TrimSpace(string(out))})
}

// apiDomainsSetDocroot changes a domain's document root.
func apiDomainsSetDocroot(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Docroot string `json:"docroot"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newDocroot := strings.TrimSpace(body.Docroot)
	if newDocroot == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "docroot is required."})
		return
	}
	resolved, ok := resolveUnderVarWWWHTML(newDocroot)
	if !ok {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "docroot must be inside '/var/www/html/'."})
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-docroot", domain, "update", resolved).CombinedOutput()
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed docroot for "+domain+" to "+resolved+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "docroot": resolved})
}

// ── Redirects ────────────────────────────────────────────────────────────

// apiDomainsGetRedirect returns a domain's redirect rule, if any.
func apiDomainsGetRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "redirect_url": getRedirectURL(domain)})
}

var apiHTTPURLRE = regexp.MustCompile(`^(https?://)`)

// apiDomainsSetRedirect sets or replaces a domain's redirect rule.
func apiDomainsSetRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		RedirectURL string `json:"redirect_url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	redirectTo := strings.TrimSpace(body.RedirectURL)
	if !apiHTTPURLRE.MatchString(redirectTo) {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_url must start with http:// or https://."})
		return
	}

	confPath := domainConfPath(domain)
	content, readErr := readTextFile(confPath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}
	updated := insertOrReplaceRedirect(content, redirectTo)
	if writeErr := writeTextFile(confPath, updated); writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}
	reloadCaddyWebserver(r)

	_ = logger.RecordUserAction(a.Config, currentUsername, "set redirect for "+domain+" to "+redirectTo+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "redirect_url": redirectTo})
}

// apiDomainsDeleteRedirect removes a domain's redirect rule.
func apiDomainsDeleteRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	redirectURL := getRedirectURL(domain)
	if redirectURL == "" {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Domain " + domain + " has no redirect configured."})
		return
	}

	confPath := domainConfPath(domain)
	content, readErr := readTextFile(confPath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}
	newContent := strings.ReplaceAll(content, "redir "+redirectURL, "")
	if writeErr := writeTextFile(confPath, newContent); writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}
	reloadCaddyWebserver(r)

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted redirect for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "deleted": true})
}

// ── SSL ──────────────────────────────────────────────────────────────────

// apiDomainsGetSSL returns a domain's SSL certificate status.
func apiDomainsGetSSL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}

	currentSetting := ""
	if out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domain, "status").CombinedOutput(); cmdErr == nil {
		currentSetting = strings.ToLower(strings.TrimSpace(string(out)))
	}
	keys := ""
	if out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domain, "info").CombinedOutput(); cmdErr == nil {
		keys = strings.TrimSpace(string(out))
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "ssl_mode": currentSetting, "keys": keys})
}

// apiDomainsConfigureSSL requests a new certificate or installs a custom one.
func apiDomainsConfigureSSL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Action      string `json:"action"`
		PublicPath  string `json:"public_path"`
		PrivatePath string `json:"private_path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := strings.TrimSpace(body.Action)

	switch action {
	case "autossl":
		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domain, "auto").CombinedOutput()
		if cmdErr != nil {
			writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "enabled AutoSSL for "+domain+" via API", reqip.ClientIP(r))
		writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "action": action, "output": strings.TrimSpace(string(out))})

	case "generate":
		success, message := triggerSSLGeneration(ctx, domain)
		if success {
			_ = logger.RecordUserAction(a.Config, currentUsername, "generated SSL certificate for "+domain+" via API", reqip.ClientIP(r))
		}
		status := http.StatusOK
		if !success {
			status = http.StatusAccepted
		}
		writeAPIDomainsJSON(w, status, map[string]any{"domain": domain, "action": action, "success": success, "message": message})

	case "switch_and_generate":
		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domain, "auto").CombinedOutput()
		if cmdErr != nil {
			writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
			return
		}
		success, message := triggerSSLGeneration(ctx, domain)
		_ = logger.RecordUserAction(a.Config, currentUsername, "switched "+domain+" to AutoSSL and triggered certificate generation via API", reqip.ClientIP(r))
		writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "action": action, "success": success, "message": message})

	case "custom":
		publicPath, publicOK := resolveUnderVarWWWHTML(body.PublicPath)
		if !publicOK {
			writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "public_path must be inside /var/www/html/."})
			return
		}
		privatePath, privateOK := resolveUnderVarWWWHTML(body.PrivatePath)
		if !privateOK {
			writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "private_path must be inside /var/www/html/."})
			return
		}
		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domain, "custom", publicPath, privatePath).CombinedOutput()
		if cmdErr != nil {
			writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": strings.TrimSpace(string(out))})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "configured custom SSL for "+domain+" via API", reqip.ClientIP(r))
		writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "action": action, "output": strings.TrimSpace(string(out))})

	default:
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be one of: autossl, generate, switch_and_generate, custom"})
	}
}

// ── VHost ────────────────────────────────────────────────────────────────

// apiDomainsGetVhost returns a domain's raw vhost config.
func apiDomainsGetVhost(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	_, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	path := vhostFilePath(userContext, domain)
	if _, statErr := os.Stat(path); statErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "VHost file not found for " + domain + "."})
		return
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "vhost": readVhostContent(userContext, domain)})
}

// apiDomainsSetVhost overwrites a domain's raw vhost config.
func apiDomainsSetVhost(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Vhost string `json:"vhost"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Vhost == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "vhost content is required."})
		return
	}

	webServerPreference := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	success, message := writeVhostContent(ctx, domain, userContext, webServerPreference, body.Vhost)
	if !success {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": message})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited VHost for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "updated": true, "message": message})
}

// ── DNS ──────────────────────────────────────────────────────────────────

const apiZoneFileDir = "/etc/bind/zones/"

func apiZoneFilePath(domain string) string {
	return apiZoneFileDir + domain + ".zone"
}

type apiDomainDNSRecord struct {
	LineNumber    int    `json:"line_number"`
	EndLineNumber int    `json:"end_line_number"`
	Record        string `json:"record"`
}

// apiDomainsGetDNS returns a domain's DNS zone records.
func apiDomainsGetDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}

	zonePath := apiZoneFilePath(domain)
	content, readErr := os.ReadFile(zonePath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found for " + domain + "."})
		return
	}

	rawLines := strings.Split(string(content), "\n")
	var serial *string
	for _, line := range rawLines {
		if strings.Contains(line, "; Serial number") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				serial = &fields[0]
			}
			break
		}
	}

	var records []apiDomainDNSRecord
	total := len(rawLines)
	for i := 0; i < total; i++ {
		line := rawLines[i]
		lineNum := i + 1
		skip := len(line) > 0 && strings.ContainsRune(" \t#$;", rune(line[0]))
		if !skip && strings.HasPrefix(line, "@") && (strings.Contains(line, "SOA") || strings.HasPrefix(line, "@ IN NS")) {
			skip = true
		}
		if skip {
			continue
		}
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		full := stripped
		openC := strings.Count(full, "(")
		closeC := strings.Count(full, ")")
		end := lineNum
		j := i
		for openC > closeC && j+1 < total {
			j++
			cont := strings.TrimSpace(rawLines[j])
			full += " " + cont
			openC += strings.Count(cont, "(")
			closeC += strings.Count(cont, ")")
			end = j + 1
		}
		records = append(records, apiDomainDNSRecord{LineNumber: lineNum, EndLineNumber: end, Record: full})
		i = j
	}
	if records == nil {
		records = []apiDomainDNSRecord{}
	}

	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "serial": serial, "records": records})
}

// apiDomainsSaveDNS overwrites a domain's DNS zone file.
func apiDomainsSaveDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		ZoneContent string `json:"zone_content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newContent := strings.TrimSpace(body.ZoneContent)
	if newContent == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "zone_content is required."})
		return
	}
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	tmpFile, tmpErr := os.CreateTemp("", "domains-dns-*.zone")
	if tmpErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": tmpErr.Error()})
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, writeErr := tmpFile.WriteString(newContent); writeErr != nil {
		_ = tmpFile.Close()
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}
	_ = tmpFile.Close()

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if cpErr := exec.CommandContext(cctx, "podman", "cp", tmpPath, "openpanel_dns:/tmp/"+domain+".zone").Run(); cpErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": cpErr.Error()})
		return
	}
	checkOut, checkErr := exec.CommandContext(cctx, "podman", "exec", "openpanel_dns", "named-checkzone", domain, "/tmp/"+domain+".zone").CombinedOutput()
	if checkErr != nil {
		writeAPIDomainsJSON(w, 422, map[string]string{"error": "Zone validation failed.", "details": strings.TrimSpace(string(checkOut))})
		return
	}

	if writeErr := os.WriteFile(apiZoneFilePath(domain), []byte(newContent), 0o644); writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	dns.RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "saved DNS zone for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "saved": true})
}

// apiDomainsAddDNSRecord appends one DNS record to a domain's zone.
func apiDomainsAddDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Name     string `json:"name"`
		TTL      any    `json:"ttl"`
		Type     string `json:"type"`
		Value    string `json:"value"`
		Priority string `json:"priority"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	name := strings.TrimSpace(body.Name)
	ttl := "3600"
	switch v := body.TTL.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			ttl = strings.TrimSpace(v)
		}
	case float64:
		ttl = strconv.Itoa(int(v))
	}
	recordType := strings.ToUpper(strings.TrimSpace(body.Type))
	record := strings.TrimSpace(body.Value)
	priority := strings.TrimSpace(body.Priority)

	if name == "" || recordType == "" || record == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "name, type, and value are required."})
		return
	}

	zonePath := apiZoneFilePath(domain)
	if _, statErr := os.Stat(zonePath); statErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found."})
		return
	}

	if strings.HasSuffix(name, domain) {
		name += "."
	}
	if strings.HasSuffix(record, domain) {
		record += "."
	}
	if recordType == "TXT" && !(strings.HasPrefix(record, `"`) && strings.HasSuffix(record, `"`)) {
		record = `"` + record + `"`
	}

	var newRecord string
	if priority != "" {
		newRecord = name + " " + ttl + " IN " + recordType + " " + priority + " " + record
	} else {
		newRecord = name + " " + ttl + " IN " + recordType + " " + record
	}

	if recordType == "CNAME" && cnameRecordExistsDomains(ctx, a, zonePath, name) {
		writeAPIDomainsJSON(w, http.StatusConflict, map[string]string{"error": "A CNAME record with this name already exists."})
		return
	}

	f, openErr := os.OpenFile(zonePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
		return
	}
	_, writeErr := f.WriteString(newRecord + "\n")
	_ = f.Close()
	if writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	dns.RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "added DNS record "+newRecord+" for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusCreated, map[string]string{"domain": domain, "record": newRecord})
}

func apiFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// cnameRecordExistsDomains checks whether a zone file already has a CNAME
// record for name, memoized 10s. Duplicated here rather than reusing the
// dns package's version, which is unexported.
func cnameRecordExistsDomains(ctx context.Context, a *appctx.App, zoneFilePath, name string) bool {
	exists, _ := cache.Memoize(ctx, a.Cache, "cname_record_exists:"+zoneFilePath+":"+name, 10*time.Second, func() (bool, error) {
		content, err := os.ReadFile(zoneFilePath)
		if err != nil {
			return false, nil
		}
		re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(name) + `\s+\d+\s+IN\s+CNAME`)
		for _, line := range strings.Split(string(content), "\n") {
			if re.MatchString(line) {
				return true, nil
			}
		}
		return false, nil
	})
	return exists
}

// apiDomainsUpdateDNSRecord edits one DNS record by row ID.
func apiDomainsUpdateDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rowID, convErr := strconv.Atoi(r.PathValue("row_id"))
	if convErr != nil {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row_id."})
		return
	}

	var body struct {
		Content  string `json:"content"`
		EndRowID any    `json:"end_row_id"`
		Serial   string `json:"serial"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newContent := strings.TrimSpace(body.Content)
	if newContent == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required."})
		return
	}
	endRowID := rowID
	if v, ok := body.EndRowID.(float64); ok {
		endRowID = int(v)
	}
	serial := strings.TrimSpace(body.Serial)

	zonePath := apiZoneFilePath(domain)
	content, readErr := os.ReadFile(zonePath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found."})
		return
	}
	lines := readLinesKeepEndsDomains(string(content))

	var zoneSerial string
	for _, line := range lines {
		if strings.Contains(line, "; Serial number") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				zoneSerial = fields[0]
			}
			break
		}
	}
	if serial != "" && zoneSerial != "" && serial != zoneSerial {
		writeAPIDomainsJSON(w, http.StatusConflict, map[string]string{"error": "Serial number mismatch — zone was modified concurrently. Refresh and try again."})
		return
	}

	if !(rowID >= 1 && rowID <= len(lines) && rowID <= endRowID && endRowID <= len(lines)) {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row_id " + strconv.Itoa(rowID) + "."})
		return
	}

	block := newContent
	if !strings.HasSuffix(block, "\n") {
		block += "\n"
	}
	newLines := make([]string, 0, len(lines))
	newLines = append(newLines, lines[:rowID-1]...)
	newLines = append(newLines, block)
	newLines = append(newLines, lines[endRowID:]...)

	if writeErr := os.WriteFile(zonePath, []byte(strings.Join(newLines, "")), 0o644); writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	dns.RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated DNS record (row "+strconv.Itoa(rowID)+") for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "row_id": rowID, "updated": true})
}

// apiDomainsDeleteDNSRecord removes one DNS record by row ID.
func apiDomainsDeleteDNSRecord(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rowID, convErr := strconv.Atoi(r.PathValue("row_id"))
	if convErr != nil {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row_id."})
		return
	}

	var body struct {
		EndRowID any `json:"end_row_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	endRowID, hasEndRowID := 0, false
	if v, ok := body.EndRowID.(float64); ok {
		endRowID = int(v)
		hasEndRowID = true
	}

	zonePath := apiZoneFilePath(domain)
	content, readErr := os.ReadFile(zonePath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found."})
		return
	}
	lines := readLinesKeepEndsDomains(string(content))

	zeroBased := rowID - 1
	if !(zeroBased >= 0 && zeroBased < len(lines)) {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid row_id " + strconv.Itoa(rowID) + "."})
		return
	}

	var deleted string
	if hasEndRowID && endRowID > rowID {
		if endRowID > len(lines) {
			endRowID = len(lines)
		}
		deleted = strings.Join(lines[zeroBased:endRowID], "")
		lines = append(lines[:zeroBased], lines[endRowID:]...)
	} else {
		deleted = lines[zeroBased]
		lines = append(lines[:zeroBased], lines[zeroBased+1:]...)
	}

	if writeErr := os.WriteFile(zonePath, []byte(strings.Join(lines, "")), 0o644); writeErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	dns.RestartDNSService(domain)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted DNS record (row "+strconv.Itoa(rowID)+") for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "row_id": rowID, "deleted": true, "deleted_record": strings.TrimSpace(deleted),
	})
}

func readLinesKeepEndsDomains(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// apiDomainsRestartDNS restarts the DNS service for a domain's zone.
func apiDomainsRestartDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-dns", "default", domain, "-y").CombinedOutput()
	if cmdErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to reset DNS zone: " + strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "reset DNS zone for "+domain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "restarted": true})
}

// apiDomainsExportDNS downloads a domain's zone file.
func apiDomainsExportDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	zonePath := apiZoneFilePath(domain)
	content, readErr := os.ReadFile(zonePath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Zone file not found."})
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := ";;\n;; Domain:     " + domain + "\n;; Exported:   " + timestamp + "\n;;\n"

	_ = logger.RecordUserAction(a.Config, currentUsername, "exported DNS zone for "+domain+" via API", reqip.ClientIP(r))

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", `attachment; filename="`+domain+`.zone"`)
	_, _ = w.Write([]byte(header))
	_, _ = w.Write(content)
}

// ── Access logs ──────────────────────────────────────────────────────────

// apiDomainsLogs returns recent access-log entries for a domain.
func apiDomainsLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}

	logPath := "/var/log/caddy/domlogs/" + domain + "/access.log"
	info, statErr := os.Stat(logPath)
	if statErr != nil {
		writeAPIDomainsJSON(w, http.StatusNotFound, map[string]string{"error": "Log file not found for " + domain + "."})
		return
	}
	if info.Size() == 0 {
		writeAPIDomainsJSON(w, http.StatusOK, map[string]any{"domain": domain, "logs": []AccessLogEntry{}, "total": 0})
		return
	}

	page := atoiDefault(r.URL.Query().Get("page"), 1)
	itemsPerPage := atoiDefault(a.Config.Get("domain_log_per_page", "1000"), 1000)
	showAll := r.URL.Query().Get("show_all") == "true"

	content, readErr := os.ReadFile(logPath)
	if readErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	var allLogs []AccessLogEntry
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry AccessLogEntry
		if json.Unmarshal([]byte(line), &entry) == nil {
			allLogs = append(allLogs, entry)
		}
	}
	for i, j := 0, len(allLogs)-1; i < j; i, j = i+1, j-1 {
		allLogs[i], allLogs[j] = allLogs[j], allLogs[i]
	}
	total := len(allLogs)

	var logs []AccessLogEntry
	var totalPages int
	if showAll {
		logs = allLogs
		totalPages = 1
	} else {
		start := (page - 1) * itemsPerPage
		end := start + itemsPerPage
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		logs = allLogs[start:end]
		totalPages = (total + itemsPerPage - 1) / itemsPerPage
	}
	if logs == nil {
		logs = []AccessLogEntry{}
	}

	writeAPIDomainsJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "logs": logs, "total": total,
		"page": page, "total_pages": totalPages, "items_per_page": itemsPerPage,
	})
}
