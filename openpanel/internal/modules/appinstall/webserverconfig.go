package appinstall

import (
	"os"
	"strconv"
	"strings"
)

// vhostConfPath is the shared vhost config path used by all of this
// file's edit_*Config() functions.
func vhostConfPath(userContext, domainURL string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
}

// editLswsConfig inserts an extprocessor+context proxy block into the
// OpenLiteSpeed vhost conf, skipping if a block for this service already
// exists, and restoring from a .bak on any write failure.
func editLswsConfig(userContext, domainURL, subdirectory, serviceName string, port int) string {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := os.ReadFile(confPath)
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
	if writeErr := os.WriteFile(confPath, []byte(updated), 0o644); writeErr != nil {
		if fileExists(backupPath) {
			_ = copyFile(backupPath, confPath)
		}
		return "Error: " + writeErr.Error()
	}
	return ""
}

// editApacheConfig inserts a ProxyPass/ProxyPassReverse pair before every
// occurrence of a marker line (</VirtualHost>, or the DirectoryIndex line
// for the root-install case). Unlike editLswsConfig, this never returns
// an error the caller acts on (failures are logged and swallowed) - so it
// reports nothing back either; the install flow continues regardless.
func editApacheConfig(userContext, domainURL, subdirectory, serviceName string, port int) {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := os.ReadFile(confPath)
	if err != nil {
		return
	}
	vhostData := string(content)

	var marker, configLines string
	target := "http://" + serviceName
	if port != 0 {
		target += ":" + strconv.Itoa(port)
	}
	if subdirectory != "" {
		marker = "</VirtualHost>"
		configLines = "\tProxyPass /" + subdirectory + "/ " + target + "/\n\tProxyPassReverse /" + subdirectory + "/ " + target + "/"
	} else {
		marker = "DirectoryIndex index.php index.html default_page.html"
		configLines = "\tProxyPass / " + target + "/\n\tProxyPassReverse / " + target + "/"
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

	if writeErr := os.WriteFile(confPath, []byte(b.String()), 0o644); writeErr != nil {
		_ = copyFile(backupPath, confPath)
	}
}

// editNginxConfig inserts a location/proxy_pass block into every
// "server {" block's "location / {" section. Like editApacheConfig,
// every failure path here just logs, so this reports nothing back either.
func editNginxConfig(userContext, domainURL, subdirectory, serviceName string, port int) {
	confPath := vhostConfPath(userContext, domainURL)
	content, err := os.ReadFile(confPath)
	if err != nil {
		return
	}
	nginxConfigData := string(content)

	serverBlocks := strings.Split(nginxConfigData, "server {")
	if len(serverBlocks) < 2 {
		return
	}

	proxyHeadersRaw, phErr := os.ReadFile("/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt")
	if phErr != nil {
		return
	}
	proxyHeaders := string(proxyHeadersRaw)

	proxyURL := "http://" + serviceName
	if port != 0 {
		proxyURL += ":" + strconv.Itoa(port)
	}
	proxyURL += "/"

	var configLines string
	if subdirectory != "" {
		configLines = "\n        location /" + subdirectory + "/ {\n            proxy_pass " + proxyURL + ";\n            " + proxyHeaders + "\n        }\n        "
	} else {
		configLines = "\n        proxy_pass " + proxyURL + ";\n            " + proxyHeaders + "\n        "
	}

	updatedBlocks := make([]string, len(serverBlocks))
	updatedBlocks[0] = serverBlocks[0]
	for i := 1; i < len(serverBlocks); i++ {
		block := "server {" + serverBlocks[i]
		markerPos := strings.Index(strings.ToLower(block), "location / {")
		if markerPos == -1 {
			return
		}
		if subdirectory != "" {
			block = block[:markerPos] + configLines + "\n" + block[markerPos:]
		} else {
			locationBlockEnd := strings.Index(block[markerPos:], "}")
			if locationBlockEnd == -1 {
				return
			}
			locationBlockEnd += markerPos
			block = block[:locationBlockEnd] + "\n" + configLines + block[locationBlockEnd:]
		}
		updatedBlocks[i] = block
	}

	updated := strings.Join(updatedBlocks, "")
	_ = os.WriteFile(confPath, []byte(updated), 0o644)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
