package domains

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// CertCheck is what the install sheet shows about a pasted certificate and key before it's installed
type CertCheck struct {
	Problem  string   `json:"problem,omitempty"` // translated, "" when it can be installed
	Subject  string   `json:"subject,omitempty"`
	Issuer   string   `json:"issuer,omitempty"`
	Names    []string `json:"names,omitempty"`
	Expires  string   `json:"expires,omitempty"`
	DaysLeft int      `json:"days_left"`
	Chain    int      `json:"chain"` // certificates in the pasted text, leaf included
	KeyType  string   `json:"key_type,omitempty"`
	KeyMatch bool     `json:"key_match"`
	Covers   bool     `json:"covers"`
}

// certProblem is one blocking check, msg may use %(date)s, %(domain)s and %(names)s
type certProblem struct{ msg, date, names string }

// checkCustomCert parses the pasted PEM the way the install will use it and reports the first thing that would break the site
func checkCustomCert(certPEM, keyPEM, domain string, now time.Time) (CertCheck, *certProblem) {
	var c CertCheck
	var leaf *x509.Certificate
	rest := []byte(strings.TrimSpace(certPEM))
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return c, &certProblem{msg: "Invalid certificate."}
		}
		c.Chain++
		if leaf == nil {
			leaf = cert
		}
	}
	if leaf == nil {
		return c, &certProblem{msg: "Invalid certificate."}
	}
	c.Subject, c.Issuer = leaf.Subject.CommonName, leaf.Issuer.CommonName
	if c.Issuer == "" && len(leaf.Issuer.Organization) > 0 {
		c.Issuer = leaf.Issuer.Organization[0]
	}
	c.Names = leaf.DNSNames
	if len(c.Names) == 0 && c.Subject != "" {
		c.Names = []string{c.Subject}
	}
	c.Expires = leaf.NotAfter.UTC().Format("2 Jan 2006")
	c.DaysLeft = int(leaf.NotAfter.Sub(now).Hours() / 24)
	c.Covers = leaf.VerifyHostname(domain) == nil

	key, keyType, keyErr := parsePrivateKeyPEM(keyPEM)
	c.KeyType = keyType
	if keyErr == "" {
		if pub, ok := key.(interface{ Public() crypto.PublicKey }); ok {
			if eq, ok := pub.Public().(interface{ Equal(crypto.PublicKey) bool }); ok {
				c.KeyMatch = eq.Equal(leaf.PublicKey)
			}
		}
	}

	switch {
	case now.After(leaf.NotAfter):
		return c, &certProblem{msg: "The certificate expired on %(date)s.", date: c.Expires}
	case now.Before(leaf.NotBefore):
		return c, &certProblem{msg: "The certificate isn't valid until %(date)s.", date: leaf.NotBefore.UTC().Format("2 Jan 2006")}
	case !c.Covers:
		return c, &certProblem{msg: "The certificate doesn't cover %(domain)s. It is for %(names)s.", names: strings.Join(c.Names, ", ")}
	case keyErr != "":
		return c, &certProblem{msg: keyErr}
	case !c.KeyMatch:
		return c, &certProblem{msg: "The private key doesn't belong to this certificate."}
	}
	return c, nil
}

// parsePrivateKeyPEM reads an unencrypted PKCS#8, PKCS#1 or SEC 1 key; errMsg is "" on success
func parsePrivateKeyPEM(keyPEM string) (key any, keyType, errMsg string) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(keyPEM)))
	if block == nil || !strings.Contains(block.Type, "PRIVATE KEY") {
		return nil, "", "Invalid private key."
	}
	//nolint:staticcheck // only detecting the legacy encrypted header, not decrypting it
	if block.Type == "ENCRYPTED PRIVATE KEY" || x509.IsEncryptedPEMBlock(block) {
		return nil, "", "The private key is encrypted. Paste it without a passphrase."
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return k, keyTypeName(k), ""
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, "RSA", ""
	}
	if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return k, "ECDSA", ""
	}
	return nil, "", "Invalid private key."
}

func keyTypeName(k any) string {
	switch k.(type) {
	case *rsa.PrivateKey:
		return "RSA"
	case *ecdsa.PrivateKey:
		return "ECDSA"
	case ed25519.PrivateKey:
		return "Ed25519"
	}
	return ""
}

// text fills a certProblem in the request's language
func (p *certProblem) text(a *appctx.App, r *http.Request, domain string) string {
	if p == nil {
		return ""
	}
	return web.Tr(a, r, p.msg, "date", p.date, "domain", domain, "names", p.names)
}

// handleCustomSSLCheck answers the install sheet's live check, nothing is written
func handleCustomSSLCheck(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	_ = r.ParseForm()
	domain := strings.TrimSpace(r.Form.Get("domain_name"))
	if domain == "" || !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"problem": web.Tr(a, r, "You do not own this domain.")})
		return
	}
	c, p := checkCustomCert(r.Form.Get("certificate"), r.Form.Get("private_key"), domain, time.Now())
	c.Problem = p.text(a, r, domain)
	web.WriteJSON(w, http.StatusOK, c)
}
