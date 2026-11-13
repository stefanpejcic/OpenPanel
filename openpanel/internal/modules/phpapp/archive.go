package phpapp

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/safenet"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// isSafeMember reports whether an extracted member resolves to somewhere inside destinationPath - a zip-slip/tar-slip guard, duplicated from filemanager/archive.go's isSafeMember since it's unexported there
func isSafeMember(destinationPath, memberName string) bool {
	target := filepath.Join(destinationPath, memberName)
	rel, err := filepath.Rel(destinationPath, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// archiveMemberNames lists the entry names of the archive at archivePath, tar global headers aren't real files so they're left out
func archiveMemberNames(archivePath, ext string) ([]string, error) {
	var names []string
	switch ext {
	case "zip":
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return nil, fmt.Errorf("invalid zip archive: %w", err)
		}
		defer zr.Close()
		for _, f := range zr.File {
			names = append(names, f.Name)
		}
	case "tar.gz", "tgz", "tar":
		f, err := tenantfs.Open(archivePath)
		if err != nil {
			return nil, fmt.Errorf("invalid archive: %w", err)
		}
		defer f.Close()
		var reader io.Reader = f
		if ext == "tar.gz" || ext == "tgz" {
			gz, gzErr := gzip.NewReader(f)
			if gzErr != nil {
				return nil, fmt.Errorf("invalid gzip archive: %w", gzErr)
			}
			defer gz.Close()
			reader = gz
		}
		tr := tar.NewReader(reader)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("invalid tar archive: %w", err)
			}
			if hdr.Typeflag != tar.TypeXGlobalHeader {
				names = append(names, hdr.Name)
			}
		}
	}
	return names, nil
}

// validateArchiveMembers pre-checks every entry before extracting, so a malicious archive can't write outside destinationPath
func validateArchiveMembers(names []string, destinationPath string) error {
	for _, name := range names {
		if !isSafeMember(destinationPath, name) {
			return fmt.Errorf("unsafe file path in archive: %s", name)
		}
	}
	return nil
}

// wrapperStrip returns 1 when everything sits inside one top folder, like GitHub's project-1.0.0/ wrapper, so the project lands right in the install path
func wrapperStrip(names []string) int {
	top, nested := "", false
	for _, name := range names {
		clean, ok := paths.MemberName(name)
		if !ok {
			return 0
		}
		if clean == "" {
			continue
		}
		first, rest, _ := strings.Cut(clean, "/")
		if top != "" && first != top {
			return 0
		}
		top = first
		nested = nested || rest != ""
	}
	if nested {
		return 1
	}
	return 0
}

// firstMissingDir walks up from p to the topmost folder that doesn't exist yet, or p itself when it's already there
func firstMissingDir(p string) string {
	if _, err := tenantfs.Stat(p); err == nil {
		return p
	}
	for {
		parent := filepath.Dir(p)
		if parent == p {
			return p
		}
		if _, err := tenantfs.Stat(parent); err == nil {
			return p
		}
		p = parent
	}
}

// downloadFile fetches url into a new temp file under dir, returning its path - capped at 200MB to keep a misbehaving/huge remote file from filling disk
// downloadTransport dials only public addresses: the URL is user-supplied and this runs as root,
// and safenet re-checks the IP on every connection (including redirect hops), so neither a redirect
// nor DNS rebinding to 127.0.0.1 / 169.254.169.254 / RFC1918 lands. Swappable so tests can point at
// a loopback httptest server, matching the git module's remoteAddrOK.
var downloadTransport http.RoundTripper = safenet.Transport(false)

func downloadFile(ctx context.Context, url, dir, ext string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 2 * time.Minute, Transport: downloadTransport}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.CreateTemp(dir, "php-app-initial-*."+ext)
	if err != nil {
		return "", err
	}
	defer out.Close()

	const maxBytes = 200 << 20
	if _, err := io.Copy(out, io.LimitReader(resp.Body, maxBytes+1)); err != nil {
		tenantfs.Remove(out.Name())
		return "", err
	}
	if info, statErr := out.Stat(); statErr == nil && info.Size() > maxBytes {
		tenantfs.Remove(out.Name())
		return "", fmt.Errorf("archive too large (over 200MB)")
	}
	return out.Name(), nil
}

// archiveExt returns the recognized archive extension for url ("zip", "tar.gz", "tgz", or "tar"), matching isArchiveURL's regex
func archiveExt(url string) string {
	switch {
	case strings.HasSuffix(strings.ToLower(url), ".tar.gz"):
		return "tar.gz"
	case strings.HasSuffix(strings.ToLower(url), ".tgz"):
		return "tgz"
	case strings.HasSuffix(strings.ToLower(url), ".tar"):
		return "tar"
	default:
		return "zip"
	}
}

// downloadAndExtractInitialProject downloads an archive URL and extracts it into hostDestPath - a host-filesystem path where the html_data volume is bind-mounted, so no podman exec is needed here; only Composer needs to run inside the php-fpm container since only it has PHP
func downloadAndExtractInitialProject(ctx context.Context, url, hostDestPath string) error {
	ext := archiveExt(url)
	tmpDir := os.TempDir()
	archivePath, err := downloadFile(ctx, url, tmpDir, ext)
	if err != nil {
		return err
	}
	defer tenantfs.Remove(archivePath)

	names, err := archiveMemberNames(archivePath, ext)
	if err != nil {
		return err
	}
	if err := validateArchiveMembers(names, hostDestPath); err != nil {
		return err
	}
	strip := wrapperStrip(names)

	extractCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var extractErr error
	switch ext {
	case "zip":
		extractErr = tenantfs.ExtractZip(extractCtx, archivePath, hostDestPath, strip)
	case "tar.gz", "tgz", "tar":
		extractErr = tenantfs.ExtractTar(extractCtx, archivePath, hostDestPath, strip)
	}
	if extractErr != nil {
		return fmt.Errorf("extraction failed: %w", extractErr)
	}
	return nil
}
