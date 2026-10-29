package domains

import (
	"context"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleEditVhosts views or edits a domain's raw virtual-host config.
func handleEditVhosts(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	webServerPreference := webserver.GetEnvFileValue(userContext, "WEB_SERVER")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainName := r.Form.Get("domain_name")
		vhostContent := web.FormText(r, "vhost_content")

		if domainName == "" {
			web.FlashRedirect(a, w, r, "error", "Invalid request. Domain name and Vhost content must be provided.", "/domains/vhosts")
			return
		}
		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		ok, message := writeVhostContent(ctx, domainName, userContext, webServerPreference, vhostContent)
		if ok {
			web.FlashRedirect(a, w, r, "success", message, "/domains/vhosts?domain="+domainName)
			_ = logger.RecordUserAction(a.Config, currentUsername, "edited Virtual Host file for domain: "+domainName, reqip.ClientIP(r))
		} else {
			web.FlashRedirect(a, w, r, "error", message, "/domains/vhosts?domain="+domainName)
		}
		return
	}

	domainName := r.URL.Query().Get("domain")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderVhostSelectPage(a, w, r, webServerPreference, domainsList)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	vhostContent := readVhostContent(userContext, domainName)
	renderVhostEditPage(a, w, r, domainName, webServerPreference, vhostContent)
}

func vhostFilePath(userContext, domainName string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainName + ".conf"
}

// readVhostContent reads a domain's raw virtual-host config file.
func readVhostContent(userContext, domainName string) string {
	content, err := tenantfs.ReadFile(vhostFilePath(userContext, domainName))
	if err != nil {
		return ""
	}
	return string(content)
}

// writeVhostContent saves the file, then validates+reloads via webserver.ReloadWebserver
func writeVhostContent(ctx context.Context, domainName, userContext, webServerPreference, vhostContent string) (bool, string) {
	path := vhostFilePath(userContext, domainName)
	previous, prevErr := tenantfs.ReadFile(path)
	if err := tenantfs.WriteFile(path, []byte(vhostContent), 0o644); err != nil {
		return false, "Error: VirtualHost file could not be saved."
	}

	if strings.Contains(strings.ToLower(webServerPreference), "litespeed") {
		if err := webserver.ReloadWebserver(ctx, userContext, webServerPreference); err != nil {
			return false, "Saved file but reloading " + webServerPreference + " failed: " + err.Error()
		}
		return true, "Successfully saved file and reloaded " + webServerPreference + " service to apply changes."
	}

	if !webserver.HasConfigTest(webServerPreference) {
		return false, "Unsupported web server: " + webServerPreference
	}

	ok, _ := webserver.TestWebserverConfig(ctx, userContext, webServerPreference)
	if !ok {
		// put the old file back so the next webserver restart doesn't die on it
		if prevErr == nil {
			_ = tenantfs.WriteFile(path, previous, 0o644)
		}
		return false, "Error saving virtual host configuration. Configuration test failed."
	}

	if err := webserver.ReloadWebserver(ctx, userContext, webServerPreference); err != nil {
		return false, "Saved file but reloading " + webServerPreference + " failed: " + err.Error()
	}
	return true, "Successfully saved file and reloaded " + webServerPreference + " service to apply changes."
}
