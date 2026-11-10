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

var blockedRanges = mustCIDRs(
	"0.0.0.0/8",
	"100.64.0.0/10", // CGNAT
	"192.0.0.0/24",
	"198.18.0.0/15",
	"240.0.0.0/4",
	"::/96",          // IPv4-compatible, ::7f00:1 is 127.0.0.1
	"64:ff9b::/96",   // NAT64
	"64:ff9b:1::/48", // local-use NAT64
	"2001::/32",      // Teredo
	"2001:db8::/32",
	"2002::/16", // 6to4
	"fec0::/10", // old site-local
)

func mustCIDRs(cidrs ...string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic(err)
		}
		out = append(out, n)
	}
	return out
}

// IsPublicIP reports whether ip is fine to connect to on an account's behalf
func IsPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	for _, n := range blockedRanges {
		if n.Contains(ip) {
			return false
		}
	}
	return true
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
