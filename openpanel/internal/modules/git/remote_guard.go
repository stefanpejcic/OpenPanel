package git

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/safenet"
)

var errRemoteNotPublic = errors.New("the remote repository address is not a public address")

// remoteAddrOK is swappable in tests that clone from a loopback server
var remoteAddrOK = safenet.IsPublicIP

// lookupRemoteIP is swappable in tests
var lookupRemoteIP = func(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

// remoteHostPort pulls host and port out of an http(s), ssh:// or scp-style remote
func remoteHostPort(remote string) (host, port string, ssh bool, err error) {
	if scpURLRE.MatchString(remote) {
		_, rest, _ := strings.Cut(remote, "@")
		host, _, _ = strings.Cut(rest, ":")
		return host, "22", true, nil
	}
	u, err := url.Parse(remote)
	if err != nil || u.Hostname() == "" {
		return "", "", false, errors.New("invalid remote URL")
	}
	port = u.Port()
	switch {
	case port != "":
	case u.Scheme == "https":
		port = "443"
	case u.Scheme == "http":
		port = "80"
	default:
		port = "22"
	}
	return u.Hostname(), port, u.Scheme == "ssh", nil
}

// pinRemote resolves the remote once and refuses private, loopback and link-local addresses, git is then told to use exactly this address
func pinRemote(ctx context.Context, remote string) (host, port, ip string, err error) {
	host, port, _, err = remoteHostPort(remote)
	if err != nil {
		return "", "", "", err
	}
	lctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var ips []net.IP
	if parsed := net.ParseIP(host); parsed != nil {
		ips = []net.IP{parsed}
	} else if ips, err = lookupRemoteIP(lctx, host); err != nil || len(ips) == 0 {
		return "", "", "", errors.New("could not resolve the remote host " + host)
	}
	for _, a := range ips {
		if !remoteAddrOK(a) {
			return "", "", "", errRemoteNotPublic
		}
	}
	return host, port, ips[0].String(), nil
}
