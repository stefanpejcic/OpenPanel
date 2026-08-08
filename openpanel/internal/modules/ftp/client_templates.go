package ftp

import (
	"encoding/xml"
	"io"
	"os"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// ftpServerConfPath optionally overrides the FTP server host/port shown on
// the /ftp page and baked into downloaded client configs. A missing file
// (the common case) just means "use the auto-detected defaults". A var
// (not const) so tests can point it at a temp file.
var ftpServerConfPath = "/etc/openpanel/ftp/server.conf"

const (
	// ftpFileZillaTemplatePath / ftpCyberduckTemplatePath optionally hold an
	// admin-provided client-config template. When present and well-formed,
	// its content (with placeholders substituted) is served instead of the
	// built-in generator.
	ftpFileZillaTemplatePath = "/etc/openpanel/ftp/filezilla.conf"
	ftpCyberduckTemplatePath = "/etc/openpanel/ftp/cyberduck.conf"

	defaultFTPPort = "21"
)

// resolveFTPHostPort returns the host/port to display and to bake into
// client config downloads. Both default to defaultHost/21 unless
// overridden by an admin in /etc/openpanel/ftp/server.conf ("hostname"
// and/or "port" keys). An invalid port override is ignored.
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

// renderFTPClientTemplate reads an optional admin-provided template file at
// path, substitutes {host}/{port}/{username}/{path} placeholders, and
// returns (content, true) when the file exists, is non-empty, and the
// result is well-formed XML. Returns ("", false) otherwise (missing file,
// read error, or invalid XML), so callers fall back to their own
// hardcoded generator.
func renderFTPClientTemplate(path, host, port, username, ftpPath string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return "", false
	}

	content := strings.NewReplacer(
		"{host}", host,
		"{port}", port,
		"{username}", username,
		"{path}", ftpPath,
	).Replace(string(data))

	if !isWellFormedXML(content) {
		return "", false
	}
	return content, true
}

// isWellFormedXML reports whether s parses as a complete XML document,
// without requiring it to match any particular schema/struct.
func isWellFormedXML(s string) bool {
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		if _, err := dec.Token(); err != nil {
			return err == io.EOF
		}
	}
}
