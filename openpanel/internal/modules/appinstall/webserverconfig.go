package appinstall

import (
	"context"
	"net"
	"regexp"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// vhostConfPath is the shared vhost config path used by all of this file's edit_*Config() functions
func vhostConfPath(userContext, domainURL string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
}

// editLswsConfig inserts an extprocessor+context proxy block into the OpenLiteSpeed vhost conf, skipping if a block for this service already exists, and restoring from a .bak on any write failure
func editLswsConfig(userContext, domainURL, subdirectory, serviceName string, port int) string {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := tenantfs.ReadFile(confPath)
	if err != nil {
		return "Error: vhost configuration file does not exist."
	}
	vhostData := string(content)

	marker := "# PM2-PROXY-START:" + serviceName
	if strings.Contains(vhostData, marker) {
		return ""
	}

	address := "http://" + serviceName
	if port != 0 {
		address += ":" + strconv.Itoa(port)
	}
	location := "/"
	if subdirectory != "" {
		location = "/" + subdirectory + "/"
	}
	extprocName := strings.NewReplacer(".", "_", "/", "_").Replace(strings.ToLower(serviceName))

	configBlock := "\n# PM2-PROXY-START:" + serviceName + "\n" +
		"extprocessor " + extprocName + " {\n" +
		"  type                    proxy\n" +
		"  address                 " + address + "\n" +
		"  maxConns                100\n" +
		"  initTimeout             60\n" +
		"  retryTimeout            0\n" +
		"  respBuffer              0\n" +
		"}\n\n" +
		"context " + location + " {\n" +
		"  type                    proxy\n" +
		"  handler                 " + extprocName + "\n" +
		"  allowBrowse             1\n" +
		"  proxyWebSocket          1\n" +
		"}\n" +
		"# PM2-PROXY-END:" + serviceName + "\n"

	backupPath := confPath + ".bak"
	if copyErr := copyFile(confPath, backupPath); copyErr != nil {
		return "Error: " + copyErr.Error()
	}

	updated := strings.TrimRight(vhostData, " \t\n\r") + "\n" + configBlock
	if writeErr := tenantfs.WriteFile(confPath, []byte(updated), 0o644); writeErr != nil {
		if fileExists(backupPath) {
			_ = copyFile(backupPath, confPath)
		}
		return "Error: " + writeErr.Error()
	}
	return ""
}

// editApacheConfig inserts a ProxyPass/ProxyPassReverse pair before every occurrence of a marker line (</VirtualHost>, or the DirectoryIndex line for a root install). Unlike editLswsConfig, failures are just logged and swallowed - the install flow continues regardless.
func editApacheConfig(userContext, domainURL, subdirectory, serviceName string, port int, wsUpgrade bool) {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := tenantfs.ReadFile(confPath)
	if err != nil {
		return
	}
	vhostData := string(content)

	var marker, configLines string
	target := "http://" + serviceName
	if port != 0 {
		target += ":" + strconv.Itoa(port)
	}
	// "upgrade=websocket" hands off any Upgrade-requested connection on this ProxyPass to mod_proxy_wstunnel automatically - needed by apps like n8n whose UI depends on a WebSocket connection back through the proxy (its real-time push channel), not needed by the plain HTTP API apps (nodejs/python/ruby/java) this function otherwise serves. Requires mod_proxy_wstunnel loaded in httpd.conf.
	proxyPassSuffix := ""
	if wsUpgrade {
		proxyPassSuffix = " upgrade=websocket"
	}
	if subdirectory != "" {
		marker = "</VirtualHost>"
		configLines = "\tProxyPass /" + subdirectory + "/ " + target + "/" + proxyPassSuffix + "\n\tProxyPassReverse /" + subdirectory + "/ " + target + "/"
	} else {
		marker = "DirectoryIndex index.php index.html default_page.html"
		configLines = "\tProxyPass / " + target + "/" + proxyPassSuffix + "\n\tProxyPassReverse / " + target + "/"
	}

	backupPath := confPath + ".bak"
	if copyErr := copyFile(confPath, backupPath); copyErr != nil {
		return
	}

	var markerPositions []int
	searchFrom := 0
	for {
		idx := strings.Index(vhostData[searchFrom:], marker)
		if idx == -1 {
			break
		}
		pos := searchFrom + idx
		markerPositions = append(markerPositions, pos)
		searchFrom = pos + len(marker)
	}

	if len(markerPositions) == 0 {
		return
	}

	var b strings.Builder
	last := 0
	for _, pos := range markerPositions {
		b.WriteString(vhostData[last:pos])
		b.WriteString(configLines)
		b.WriteString("\n")
		b.WriteString(vhostData[pos : pos+len(marker)])
		last = pos + len(marker)
	}
	b.WriteString(vhostData[last:])

	if writeErr := tenantfs.WriteFile(confPath, []byte(b.String()), 0o644); writeErr != nil {
		_ = copyFile(backupPath, confPath)
	}
}

var nginxTryFilesRE = regexp.MustCompile(`(?m)^([ \t]*)(try_files[ \t])`)

// nginxRootAppHeader replaces a root app's "location / {", stripServiceProxies turns it back
func nginxRootAppHeader(service string) string {
	return "location ^~ / { # OPENPANEL-APP:" + service
}

// nginxAppTag comments out a line a root app disables, the trailing space keeps kuma from matching kuma2
func nginxAppTag(service string) string {
	return "# OPENPANEL-APP:" + service + " "
}

// insertNginxProxy adds configLines to every server block, a subdirectory app gets its own location next to location /, a root app takes location / over
func insertNginxProxy(conf, subdirectory, serviceName, configLines string) (string, string) {
	serverBlocks := strings.Split(conf, "server {")
	if len(serverBlocks) < 2 {
		return "", "Error: no server block found in vhost configuration."
	}
	updatedBlocks := make([]string, len(serverBlocks))
	updatedBlocks[0] = serverBlocks[0]
	for i := 1; i < len(serverBlocks); i++ {
		block := "server {" + serverBlocks[i]
		markerPos := strings.Index(strings.ToLower(block), "location / {")
		if markerPos == -1 {
			return "", "Error: no location / block found in vhost configuration."
		}
		if subdirectory != "" {
			block = block[:markerPos] + configLines + "\n" + block[markerPos:]
		} else {
			locationBlockEnd := strings.Index(block[markerPos:], "}")
			if locationBlockEnd == -1 {
				return "", "Error: location / block is not closed in vhost configuration."
			}
			locationBlockEnd += markerPos
			// try_files would hand every path but / to PHP and the regex locations would serve the app's files from disk, both changes are tagged so delete can undo them
			loc := block[markerPos:locationBlockEnd]
			loc = nginxRootAppHeader(serviceName) + loc[len("location / {"):]
			loc = nginxTryFilesRE.ReplaceAllString(loc, "${1}"+nginxAppTag(serviceName)+"${2}")
			block = block[:markerPos] + loc + "\n" + configLines + block[locationBlockEnd:]
		}
		updatedBlocks[i] = block
	}

	return strings.Join(updatedBlocks, ""), ""
}

// the subdirectory goes into a rewrite regex for the per-request form, anything else keeps the plain proxy_pass
var nginxPlainSubdirRE = regexp.MustCompile(`^[A-Za-z0-9._~-]+(/[A-Za-z0-9._~-]+)*$`)

// nginxProxyLines is the proxy part of an app's nginx block, with a resolver the upstream is looked up per request so a stopped app only 502s its own URLs instead of keeping nginx from starting
func nginxProxyLines(upstream, subdirectory, resolver string) string {
	if resolver == "" || (subdirectory != "" && !nginxPlainSubdirRE.MatchString(subdirectory)) {
		return "proxy_pass " + upstream + "/;"
	}
	// the set line comes first, stripServiceProxies finds the app by it
	lines := []string{"set $openpanel_upstream " + upstream + ";", "resolver " + resolver + " valid=10s ipv6=off;"}
	if subdirectory != "" {
		// a proxy_pass with variables can't swap the prefix like proxy_pass http://app/; does
		lines = append(lines, "rewrite ^/"+regexp.QuoteMeta(subdirectory)+"/(.*)$ /$1 break;")
	}
	lines = append(lines, "proxy_pass $openpanel_upstream;")
	return strings.Join(lines, "\n            ")
}

// nginxConfigLines is what insertNginxProxy adds, its own location for a subdirectory app or the lines that go inside location / for a root one
func nginxConfigLines(subdirectory, proxyLines, proxyHeaders string) string {
	if subdirectory != "" {
		// ^~ so the vhost's .php and static file regex locations can't grab the app's URLs
		return "\n        location ^~ /" + subdirectory + "/ {\n            " + proxyLines + "\n            " + proxyHeaders + "\n        }\n        "
	}
	return "\n        " + proxyLines + "\n            " + proxyHeaders + "\n        "
}

// nginxResolver is the gateway of the account's www network, where aardvark-dns answers for its containers, or "" if podman can't tell
func nginxResolver(ctx context.Context, userContext string) string {
	argv := podmanmanager.PodmanArgv(userContext, "network", "inspect", userContext+"_www", "--format", "{{range .Subnets}}{{.Gateway}} {{end}}")
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return ""
	}
	for _, f := range strings.Fields(string(out)) {
		if ip := net.ParseIP(f); ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

// editNginxConfig inserts a location/proxy_pass block into every "server {" block's "location / {" section, then runs nginx -t and restores the old file if it fails, returning the error or ""
func editNginxConfig(ctx context.Context, userContext, domainURL, subdirectory, serviceName, webServer string, port int, wsUpgrade bool) string {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := tenantfs.ReadFile(confPath)
	if err != nil {
		return "Error: vhost configuration file does not exist."
	}
	nginxConfigData := string(content)

	proxyHeadersRaw, phErr := tenantfs.ReadFile("/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt")
	if phErr != nil {
		return "Error: " + phErr.Error()
	}
	proxyHeaders := string(proxyHeadersRaw)
	// no LoadModule needed for nginx (unlike Apache's mod_proxy_wstunnel) - ngx_http_proxy_module handles the Upgrade handshake natively once these headers are forwarded and HTTP/1.1 is used towards the upstream. Needed by apps like n8n whose UI depends on a WebSocket connection back through the proxy (its real-time push channel).
	if wsUpgrade {
		// default headers file already has proxy_http_version, nginx fails on a duplicate
		if !strings.Contains(proxyHeaders, "proxy_http_version") {
			proxyHeaders += "\n            proxy_http_version 1.1;"
		}
		proxyHeaders += "\n            proxy_set_header Upgrade $http_upgrade;\n            proxy_set_header Connection \"upgrade\";"
	}

	upstream := "http://" + serviceName
	if port != 0 {
		upstream += ":" + strconv.Itoa(port)
	}
	configLines := nginxConfigLines(subdirectory, nginxProxyLines(upstream, subdirectory, nginxResolver(ctx, userContext)), proxyHeaders)

	updated, errMsg := insertNginxProxy(nginxConfigData, subdirectory, serviceName, configLines)
	if errMsg != "" {
		return errMsg
	}
	if err := tenantfs.WriteFile(confPath, []byte(updated), 0o644); err != nil {
		_ = tenantfs.WriteFile(confPath, content, 0o644)
		return "Error: " + err.Error()
	}

	// nginx -t only works inside a running container
	if docker.IsServiceRunning(ctx, userContext, webServer) {
		if ok, out := webserver.TestWebserverConfig(ctx, userContext, webServer); !ok {
			_ = tenantfs.WriteFile(confPath, content, 0o644)
			return "Error: proxy config failed " + webServer + " syntax check and was reverted: " + out
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	data, err := tenantfs.ReadFile(src)
	if err != nil {
		return err
	}
	return tenantfs.WriteFile(dst, data, 0o644)
}
