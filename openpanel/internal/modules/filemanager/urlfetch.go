package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

var (
	errBlockedAddress   = errors.New("destination address is not public")
	errDownloadTooLarge = errors.New("download exceeds size limit")
)

func isPublicIP(ip net.IP) bool {
	return !(ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() || cgnatRange.Contains(ip))
}

var cgnatRange = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// safeDialer checks the IP actually being connected to, so redirects and DNS rebinding can't reach internal services
var safeDialer = &net.Dialer{
	Timeout: 15 * time.Second,
	Control: func(network, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil || !isPublicIP(ip) {
			return fmt.Errorf("%w: %s", errBlockedAddress, host)
		}
		return nil
	},
}

var safeHTTPClient = &http.Client{
	Transport: &http.Transport{
		Proxy:                 nil,
		DialContext:           safeDialer.DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return errors.New("redirect to unsupported scheme")
		}
		return nil
	},
}

// downloadToFile streams rawURL into dest under root (which must not exist yet) and removes the partial file on any error
func downloadToFile(ctx context.Context, root *os.Root, rawURL, dest string, maxBytes int64, progress func(pct int)) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "OpenPanel-FileManager")
	resp, err := safeHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	if resp.ContentLength > maxBytes {
		return errDownloadTooLarge
	}

	// O_EXCL so a symlink planted at the destination is never followed
	f, err := root.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = root.Remove(dest)
		}
	}()

	var written int64
	buf := make([]byte, 64<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			written += int64(n)
			if written > maxBytes {
				return errDownloadTooLarge
			}
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			if resp.ContentLength > 0 && progress != nil {
				progress(int(written * 100 / resp.ContentLength))
			}
		}
		if rerr == io.EOF {
			return nil
		}
		if rerr != nil {
			return rerr
		}
	}
}
