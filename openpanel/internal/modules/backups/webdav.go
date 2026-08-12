package backups

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// webdavStore is the remoteStore implementation for the "webdav"
// backup.env section. There's no client library involved - just plain
// PROPFIND/GET requests over net/http, which is all a WebDAV client needs
// for listing a folder and downloading a file.
type webdavStore struct {
	config map[string]string
}

// targetURL is the folder backups live in: WEBDAV_URL (the server root)
// joined with WEBDAV_PATH (the folder within it), always ending in "/" so
// resolving a filename against it with url.Parse/ResolveReference never
// accidentally drops the last path segment.
func (s *webdavStore) targetURL() (*url.URL, error) {
	base := strings.TrimSpace(s.config["WEBDAV_URL"])
	if base == "" {
		return nil, errors.New("WEBDAV_URL is not configured")
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if p := strings.Trim(s.config["WEBDAV_PATH"], "/"); p != "" {
		u, err = u.Parse(p + "/")
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (s *webdavStore) httpClient() *http.Client {
	client := &http.Client{}
	if s.config["WEBDAV_URL_INSECURE"] == "true" || s.config["WEBDAV_URL_INSECURE"] == "1" {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec // opt-in via WEBDAV_URL_INSECURE, mirrors AWS_ENDPOINT_INSECURE/etc.
	}
	return client
}

func (s *webdavStore) newRequest(ctx context.Context, method string, u *url.URL, extraHeaders map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if user := s.config["WEBDAV_USERNAME"]; user != "" {
		req.SetBasicAuth(user, s.config["WEBDAV_PASSWORD"])
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	return req, nil
}

type davMultistatus struct {
	Responses []davResponse `xml:"response"`
}

type davResponse struct {
	Href     string        `xml:"href"`
	Propstat []davPropstat `xml:"propstat"`
}

type davPropstat struct {
	Prop davProp `xml:"prop"`
}

type davProp struct {
	ResourceType struct {
		Collection *struct{} `xml:"collection"`
	} `xml:"resourcetype"`
}

func (r davResponse) isCollection() bool {
	for _, ps := range r.Propstat {
		if ps.Prop.ResourceType.Collection != nil {
			return true
		}
	}
	return false
}

// List issues a Depth:1 PROPFIND against the target folder and returns the
// non-directory entries' filenames.
func (s *webdavStore) List(ctx context.Context) ([]string, error) {
	target, err := s.targetURL()
	if err != nil {
		return nil, err
	}

	req, err := s.newRequest(ctx, "PROPFIND", target, map[string]string{
		"Depth":        "1",
		"Content-Type": "application/xml",
	})
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("PROPFIND %s: unexpected status %s", target, resp.Status)
	}

	var ms davMultistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, err
	}

	targetPath := strings.TrimRight(target.Path, "/")

	var names []string
	for _, r := range ms.Responses {
		if r.isCollection() {
			continue
		}
		hrefURL, err := url.Parse(r.Href)
		if err != nil {
			continue
		}
		entryPath := strings.TrimRight(hrefURL.Path, "/")
		if entryPath == "" || entryPath == targetPath {
			continue
		}
		names = append(names, path.Base(entryPath))
	}
	return names, nil
}

// Fetch downloads one file from the target folder into a local temp file.
func (s *webdavStore) Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error) {
	target, err := s.targetURL()
	if err != nil {
		return "", nil, err
	}
	fileURL, err := target.Parse(filename)
	if err != nil {
		return "", nil, err
	}

	req, err := s.newRequest(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("GET %s: unexpected status %s", fileURL, resp.Status)
	}

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	localPath = filepath.Join(tmpDir, filename)
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, resp.Body); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

func (s *webdavStore) Classify(ctx context.Context, filename string) (BackupInfo, error) {
	return classifyViaDownload(ctx, s, filename)
}
