// Package webserverconf implements an in-browser editor for a user's main
// webserver configuration file, with a syntax-test-before-restart safety
// net and a restore-to-stock-defaults option.
package webserverconf

import (
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// defaultConfTemplates lists the stock configuration files shipped in
// openpanel-configuration, used to restore a user's webserver conf file
// back to the default.
var defaultConfTemplates = map[string]string{
	"nginx":         "/etc/openpanel/nginx/nginx.conf",
	"apache":        "/etc/openpanel/apache/httpd.conf",
	"openresty":     "/etc/openpanel/openresty/nginx.conf",
	"openlitespeed": "/etc/openpanel/openlitespeed/httpd_config.conf",
	// TODO: file for LSWS ENTERPRISE!
}

// webserverConfEntry is one row of the WEB_SERVER -> (filename, service,
// page title) lookup table.
type webserverConfEntry struct {
	ConfFile    string
	ServiceName string
	PageTitle   string
}

var webserverConfs = map[string]webserverConfEntry{
	"nginx":         {"nginx.conf", "nginx", "Nginx Configuration Editor"},
	"apache":        {"httpd.conf", "apache", "Apache Configuration Editor"},
	"openresty":     {"openresty.conf", "openresty", "OpenResty Configuration Editor"},
	"openlitespeed": {"openlitespeed.conf", "openlitespeed", "OpenLitespeed Configuration Editor"},
	"litespeed":     {"openlitespeed.conf", "litespeed", "Litespeed Configuration Editor"},
}

// lookupWebserverConf returns webServer's config entry, or a fallback with
// just a generic page title if webServer is unrecognized.
func lookupWebserverConf(webServer string) webserverConfEntry {
	if entry, ok := webserverConfs[webServer]; ok {
		return entry
	}
	return webserverConfEntry{PageTitle: "Web Server Configuration Editor"}
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
