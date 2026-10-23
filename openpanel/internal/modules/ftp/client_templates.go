package ftp

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// ftpServerConfPath optionally overrides the FTP server host/port shown on the /ftp page and baked into downloaded client configs - a missing file (the common case) just means "use the auto-detected defaults", a var (not const) so tests can point it at a temp file
var ftpServerConfPath = "/etc/openpanel/ftp/server.conf"

const (
	// ftpFileZillaTemplatePath / ftpCyberduckTemplatePath optionally hold an admin-provided client-config template - when present and well-formed, its content (with placeholders substituted) is served instead of the built-in generator
	ftpFileZillaTemplatePath = "/etc/openpanel/ftp/filezilla.conf"
	ftpCyberduckTemplatePath = "/etc/openpanel/ftp/cyberduck.conf"

	defaultFTPPort = "21"
)

// resolveFTPHostPort returns the host/port to display and to bake into client config downloads - both default to defaultHost/21 unless overridden by an admin in /etc/openpanel/ftp/server.conf ("hostname" and/or "port" keys), an invalid port override is ignored
func resolveFTPHostPort(defaultHost string) (host, port string) {
	cfg, _ := config.Load(ftpServerConfPath)

	host = defaultHost
	if v := strings.TrimSpace(cfg.Get("hostname", "")); v != "" {
		host = v
	}

	port = defaultFTPPort
	if v := strings.TrimSpace(cfg.Get("port", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 65535 {
			port = v
		}
	}
	return host, port
}

// renderFTPClientTemplate reads an optional admin-provided template file at path, substitutes {host}/{port}/{username}/{path} placeholders, and returns (content, true) when the file exists, is non-empty, and the result is well-formed XML - returns ("", false) otherwise so callers fall back to their own hardcoded generator
func renderFTPClientTemplate(path, host, port, username, ftpPath string) (string, bool) {
	data, err := tenantfs.ReadFile(path)
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return "", false
	}

	content := strings.NewReplacer(
		"{host}", xmlText(host),
		"{port}", xmlText(port),
		"{username}", xmlText(username),
		"{path}", xmlText(ftpPath),
	).Replace(string(data))

	if !isWellFormedXML(content) {
		return "", false
	}
	return content, true
}

// xmlText escapes s for XML text or attribute values, folder names can have & < > or quotes in them
func xmlText(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// ftpClientConfig is the cyberduck or filezilla file for one account, from the admin template when there's a usable one
func ftpClientConfig(clientType, host, port, username, ftpPath string) string {
	if clientType == "cyberduck" {
		if rendered, ok := renderFTPClientTemplate(ftpCyberduckTemplatePath, host, port, username, ftpPath); ok {
			return rendered
		}
		return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<bookmark>\n    <hostname>" + xmlText(host) + "</hostname>\n    <username>" + xmlText(username) + "</username>\n    <protocol>ftp</protocol>\n    <path>" + xmlText(ftpPath) + "</path>\n</bookmark>"
	}
	if rendered, ok := renderFTPClientTemplate(ftpFileZillaTemplatePath, host, port, username, ftpPath); ok {
		return rendered
	}
	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<FileZilla3>\n    <Servers>\n        <Server>\n            <Host>" + xmlText(host) + "</Host>\n            <Port>" + xmlText(port) + "</Port>\n            <Protocol>0</Protocol>\n            <Type>0</Type>\n            <User>" + xmlText(username) + "</User>\n            <Logontype>2</Logontype>\n            <EncodingType>Auto</EncodingType>\n            <Name>" + xmlText(username) + "</Name>\n            <RemoteDir>" + xmlText(ftpPath) + "</RemoteDir>\n            <UsePassive>1</UsePassive>\n            <BypassProxy>0</BypassProxy>\n            <Encryption>0</Encryption>\n        </Server>\n    </Servers>\n</FileZilla3>"
}

// isWellFormedXML reports whether s parses as a complete XML document, without requiring it to match any particular schema/struct
func isWellFormedXML(s string) bool {
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		if _, err := dec.Token(); err != nil {
			return err == io.EOF
		}
	}
}
