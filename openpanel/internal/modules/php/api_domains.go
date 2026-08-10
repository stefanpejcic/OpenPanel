package php

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterDomainsAPI wires the per-domain PHP version API routes onto mux.
// PUT /api/php/domains/{domain} would collide with PUT /api/php/{version}/options
// - both are 3-segment patterns with the wildcard in a different position,
// which Go's http.ServeMux rejects as ambiguous (e.g. "/api/php/domains/options"
// would match either) - so the domain is passed in the body instead of the
// path here.
func RegisterDomainsAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php", "GET /api/php/domains", func(w http.ResponseWriter, r *http.Request) { apiPHPDomainsList(a, w, r) })
	apiregistry.Handle(mux, a, "php", "PUT /api/php/domains", func(w http.ResponseWriter, r *http.Request) { apiPHPDomainSet(a, w, r) })
}

// apiPHPDomainsList returns every domain owned by the caller together with
// its currently assigned PHP version.
func apiPHPDomainsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpDefaultVersion := webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION")
	installedVersions := FetchPHPVersions(ctx, a, userContext)
	phpVersionsData := fetchPHPVersionsAPI(ctx)

	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	rows, counts, outdated := buildPHPDomainRows(userContext, domainsList, phpVersionsData)

	writeJSON(w, http.StatusOK, map[string]any{
		"domains":                rows,
		"counts":                 counts,
		"outdated_domains":       outdated,
		"available_php_versions": installedVersions,
		"php_default_version":    phpDefaultVersion,
		"is_litespeed":           isLitespeed,
	})
}

// apiPHPDomainSet switches one domain to a different (already installed)
// PHP version, rewriting its vhost upstream and restarting services as
// needed. Not available under LiteSpeed, which has no per-domain PHP - use
// PUT /api/php/default there instead.
func apiPHPDomainSet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Domain  string `json:"domain"`
		Version string `json:"version"`
	}
	if decErr := json.NewDecoder(r.Body).Decode(&body); decErr != nil {
		_ = r.ParseForm()
		body.Domain = r.Form.Get("domain")
		body.Version = r.Form.Get("version")
	}
	domainURL := strings.TrimSpace(body.Domain)
	if domainURL == "" || !validDomainRE.MatchString(domainURL) {
		writeJSONError(w, http.StatusBadRequest, "Invalid domain")
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		writeJSONError(w, http.StatusBadRequest, "Only one PHP version can be used on LiteSpeed; set the server-wide default with PUT /api/php/default instead")
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
		writeJSONError(w, http.StatusForbidden, "You do not own this domain")
		return
	}

	newVersion := strings.TrimSpace(body.Version)

	installedVersions := FetchPHPVersions(ctx, a, userContext)
	if newVersion == "" || !phpVersionFormRE.MatchString(newVersion) || !containsString(installedVersions, newVersion) {
		writeJSONError(w, http.StatusBadRequest, "Invalid or unavailable PHP version selected")
		return
	}

	oldVersion := GetPHPVForDomain(ctx, a, userContext, domainURL)
	if oldVersion == "/" || !phpVersionFormRE.MatchString(oldVersion) {
		writeJSONError(w, http.StatusUnprocessableEntity, "Could not determine current PHP version for domain "+domainURL)
		return
	}

	if oldVersion == newVersion {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "Domain " + domainURL + " already uses PHP " + newVersion, "domain": domainURL, "version": newVersion,
		})
		return
	}

	newPHPContainer := "php-fpm-" + newVersion
	if !docker.IsServiceRunning(ctx, userContext, newPHPContainer) {
		result := docker.StartOrStopContainer(ctx, userContext, newPHPContainer, "activate", "run")
		if !result.Success {
			writeJSONError(w, http.StatusInternalServerError, "Failed to start PHP "+newVersion+": "+result.Message)
			return
		}
	}

	confFile := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
	content, readErr := os.ReadFile(confFile)
	if readErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read vhost configuration for "+domainURL+": "+readErr.Error())
		return
	}

	updated := strings.ReplaceAll(string(content), "php-fpm-"+oldVersion, "php-fpm-"+newVersion)
	if writeErr := os.WriteFile(confFile, []byte(updated), 0o644); writeErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to write vhost configuration for "+domainURL+": "+writeErr.Error())
		return
	}
	if strings.Contains(updated, "php-fpm-"+oldVersion) {
		writeJSONError(w, http.StatusInternalServerError, "Error occurred while updating PHP version for domain "+domainURL)
		return
	}

	stopPHPServiceIfRunningAndUnused(ctx, userContext, oldVersion)

	_, _ = a.DB.ExecContext(ctx, "UPDATE domains SET php_version = ? WHERE domain_url = ?", newVersion, domainURL)

	reloadArgv := podmanmanager.PodmanArgv(userContext, "restart", webServer)
	_ = podmanmanager.Command(ctx, userContext, reloadArgv).Run()

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, fmt.Sprintf("changed PHP version for domain %s from %s to %s", domainURL, oldVersion, newVersion), ipAddress)

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     fmt.Sprintf("PHP version for domain %s updated from %s to %s", domainURL, oldVersion, newVersion),
		"domain":      domainURL,
		"old_version": oldVersion,
		"new_version": newVersion,
	})
}
