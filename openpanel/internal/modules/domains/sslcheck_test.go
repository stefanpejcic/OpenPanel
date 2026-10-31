package domains

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

func testCertPEM(t *testing.T, key any, pub any, names []string, from, to time.Time) string {
	t.Helper()
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: names[0]}, Issuer: pkix.Name{CommonName: "Test CA"}, DNSNames: names, NotBefore: from, NotAfter: to}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestCheckCustomCert(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rk, _ := rsa.GenerateKey(rand.Reader, 2048)
	der, _ := x509.MarshalPKCS8PrivateKey(ec)
	ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	der, _ = x509.MarshalPKCS8PrivateKey(other)
	otherPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	rsaPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rk)}))

	good := testCertPEM(t, ec, &ec.PublicKey, []string{"example.com", "*.example.com"}, now.AddDate(0, -1, 0), now.AddDate(0, 2, 0))
	c, p := checkCustomCert(good+good, ecPEM, "shop.example.com", now)
	if p != nil || !c.KeyMatch || !c.Covers || c.Chain != 2 || c.KeyType != "ECDSA" || c.Issuer != "example.com" || c.DaysLeft < 55 {
		t.Fatalf("good: %+v %+v", c, p)
	}
	rsaCert := testCertPEM(t, rk, &rk.PublicKey, []string{"example.com"}, now.AddDate(0, -1, 0), now.AddDate(1, 0, 0))
	if c, p := checkCustomCert(rsaCert, rsaPEM, "example.com", now); p != nil || c.KeyType != "RSA" || !c.KeyMatch {
		t.Fatalf("rsa pkcs1: %+v %+v", c, p)
	}

	cases := []struct {
		name, cert, key, domain, want string
	}{
		{"garbage cert", "hello", ecPEM, "example.com", "Invalid certificate."},
		{"bad key", good, "-----BEGIN PRIVATE KEY-----\nAAAA\n-----END PRIVATE KEY-----", "example.com", "Invalid private key."},
		{"encrypted key", good, "-----BEGIN ENCRYPTED PRIVATE KEY-----\nAAAA\n-----END ENCRYPTED PRIVATE KEY-----", "example.com", "encrypted"},
		{"other key", good, otherPEM, "example.com", "doesn't belong"},
		{"wrong domain", good, ecPEM, "example.org", "doesn't cover"},
		{"expired", testCertPEM(t, ec, &ec.PublicKey, []string{"example.com"}, now.AddDate(-1, 0, 0), now.AddDate(0, 0, -1)), ecPEM, "example.com", "expired"},
		{"not yet", testCertPEM(t, ec, &ec.PublicKey, []string{"example.com"}, now.AddDate(0, 0, 2), now.AddDate(1, 0, 0)), ecPEM, "example.com", "isn't valid until"},
	}
	for _, tc := range cases {
		_, p := checkCustomCert(tc.cert, tc.key, tc.domain, now)
		if p == nil || !strings.Contains(p.msg, tc.want) {
			t.Errorf("%s: got %+v, want %q", tc.name, p, tc.want)
		}
	}
}
