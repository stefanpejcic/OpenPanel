package domains

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"
)

// sslBaseDir is where opencli domains-ssl keeps both acme and custom certs
var sslBaseDir = "/etc/openpanel/caddy/ssl"

const sslExpiringSoonWindow = 14 * 24 * time.Hour

// SSLCertInfo is what the domains list shows about a domain's actual certificate
type SSLCertInfo struct {
	Problem string // "" | "expiring" | "expired" | "missing" | "unreadable"
	Expires string // 2006-01-02, empty when unknown
	Issuer  string
}

// sslCertPath uses the same layout opencli domains-ssl writes to
func sslCertPath(domain, httpsType string) string {
	switch httpsType {
	case "Custom":
		return filepath.Join(sslBaseDir, "custom", domain, "fullchain.pem")
	case "Automatic":
		return filepath.Join(sslBaseDir, "acme-v02.api.letsencrypt.org-directory", domain, domain+".crt")
	}
	return ""
}

// readSSLCertInfo parses the leaf cert to get expiry, same approach as OpenAdmin's domains list
func readSSLCertInfo(domain, httpsType string) SSLCertInfo {
	path := sslCertPath(domain, httpsType)
	if path == "" {
		return SSLCertInfo{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return SSLCertInfo{Problem: "missing"}
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return SSLCertInfo{Problem: "unreadable"}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return SSLCertInfo{Problem: "unreadable"}
	}

	info := SSLCertInfo{Expires: cert.NotAfter.Format("2006-01-02"), Issuer: cert.Issuer.CommonName}
	if len(cert.Issuer.Organization) > 0 {
		info.Issuer = cert.Issuer.Organization[0]
	}
	switch {
	case time.Now().After(cert.NotAfter):
		info.Problem = "expired"
	case time.Until(cert.NotAfter) <= sslExpiringSoonWindow:
		info.Problem = "expiring"
	}
	return info
}
