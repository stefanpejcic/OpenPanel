// Package webserverconf implements an in-browser editor for a user's main webserver configuration file, with a syntax-test-before-reload safety net and a restore-to-stock-defaults option.
package webserverconf

import "gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"

// defaultConfTemplates lists the stock configuration files shipped in openpanel-configuration, used to restore a user's webserver conf file back to the default
var defaultConfTemplates = map[string]string{
	"nginx":         "/etc/openpanel/nginx/nginx.conf",
	"apache":        "/etc/openpanel/apache/httpd.conf",
	"openresty":     "/etc/openpanel/openresty/nginx.conf",
	"openlitespeed": "/etc/openpanel/openlitespeed/httpd_config.conf",
	"litespeed":     "/etc/openpanel/litespeed/httpd.conf",
}

// webserverConfEntry is one row of the WEB_SERVER -> (filename, service, page title) lookup table
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
	"litespeed":     {"litespeed.conf", "litespeed", "LiteSpeed Configuration Editor"},
}

// lookupWebserverConf returns webServer's config entry, or a fallback with just a generic page title if webServer is unrecognized
func lookupWebserverConf(webServer string) webserverConfEntry {
	if entry, ok := webserverConfs[webServer]; ok {
		return entry
	}
	return webserverConfEntry{PageTitle: "Web Server Configuration Editor"}
}

func fileExists(path string) bool {
	info, err := tenantfs.Stat(path)
	return err == nil && !info.IsDir()
}
