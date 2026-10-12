package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestCert(t *testing.T, certFile, keyFile, cn string) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: cn}, NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	_ = os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600)
	_ = os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600)
}

func TestCertReloaderPicksUpRenewal(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "a.crt"), filepath.Join(dir, "a.key")
	writeTestCert(t, certFile, keyFile, "old")
	c, err := newCertReloader(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := c.GetCertificate(nil)

	writeTestCert(t, certFile, keyFile, "new")
	future := time.Now().Add(time.Hour)
	_ = os.Chtimes(certFile, future, future)
	c.checkedAt = time.Time{}

	second, _ := c.GetCertificate(nil)
	if first == second {
		t.Fatal("certificate was not reloaded after the file changed")
	}
}
