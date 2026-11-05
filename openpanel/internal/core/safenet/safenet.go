// Package safenet dials only public addresses, for connections to hosts an account picked itself.
package safenet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// ErrBlockedAddress is returned when the address being connected to is loopback, private, link-local or similar
var ErrBlockedAddress = errors.New("destination address is not public")

var cgnatRange = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// IsPublicIP reports whether ip is fine to connect to on an account's behalf
func IsPublicIP(ip net.IP) bool {
	return !(ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() || cgnatRange.Contains(ip))
}

// Dialer checks the IP actually being connected to, so redirects and DNS rebinding can't reach internal services
var Dialer = &net.Dialer{
	Timeout: 15 * time.Second,
	Control: func(network, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil || !IsPublicIP(ip) {
			return fmt.Errorf("%w: %s", ErrBlockedAddress, host)
		}
		return nil
	},
}

// Transport is an http.Transport that only reaches public addresses, insecure skips TLS verification
func Transport(insecure bool) *http.Transport {
	t := &http.Transport{
		Proxy:                 nil,
		DialContext:           Dialer.DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	if insecure {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in by the account for its own endpoint
	}
	return t
}
