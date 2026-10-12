package main

import (
	"crypto/tls"
	"log"
	"os"
	"sync"
	"time"
)

// certReloader serves the panel's TLS cert and picks up a renewed one from disk without a restart
type certReloader struct {
	certFile, keyFile string

	mu        sync.Mutex
	cert      *tls.Certificate
	modTime   time.Time
	checkedAt time.Time
}

const certCheckInterval = time.Minute

func newCertReloader(certFile, keyFile string) (*certReloader, error) {
	c := &certReloader{certFile: certFile, keyFile: keyFile}
	if err := c.load(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *certReloader) load() error {
	info, err := os.Stat(c.certFile)
	if err != nil {
		return err
	}
	cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return err
	}
	c.cert, c.modTime = &cert, info.ModTime()
	return nil
}

// GetCertificate stats the cert at most once a minute and reloads it when it changed, a broken new file keeps the old cert in place
func (c *certReloader) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.checkedAt) >= certCheckInterval {
		c.checkedAt = time.Now()
		if info, err := os.Stat(c.certFile); err == nil && !info.ModTime().Equal(c.modTime) {
			if err := c.load(); err != nil {
				log.Printf("BOOTSTRAP - renewed certificate %s could not be loaded, keeping the old one: %v", c.certFile, err)
			} else {
				log.Printf("BOOTSTRAP - reloaded renewed certificate %s", c.certFile)
			}
		}
	}
	return c.cert, nil
}
