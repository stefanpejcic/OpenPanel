package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/safenet"
)

var (
	errBlockedAddress   = safenet.ErrBlockedAddress
	errDownloadTooLarge = errors.New("download exceeds size limit")
)

func isPublicIP(ip net.IP) bool { return safenet.IsPublicIP(ip) }

var safeDialer = safenet.Dialer

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
