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
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// isSafeMember reports whether an extracted member resolves to somewhere
// inside destinationPath - a zip-slip / tar-slip guard, duplicated from
// internal/modules/filemanager/archive.go's isSafeMember since it's
// unexported there.
func isSafeMember(destinationPath, memberName string) bool {
	target := filepath.Join(destinationPath, memberName)
	rel, err := filepath.Rel(destinationPath, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// validateArchiveMembers pre-checks every entry in the archive before
// shelling out to unzip/tar, so a malicious archive can't write outside
// destinationPath. Duplicated (zip/tar.gz/tgz/tar branches only, no bare
// .gz - initial_project archives are always multi-file projects) from
// filemanager's validateArchiveMembers.
func validateArchiveMembers(archivePath, ext, destinationPath string) error {
	switch ext {
	case "zip":
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return fmt.Errorf("invalid zip archive: %w", err)
		}
		defer zr.Close()
		for _, f := range zr.File {
			if !isSafeMember(destinationPath, f.Name) {
				return fmt.Errorf("unsafe file path in archive: %s", f.Name)
			}
		}
	case "tar.gz", "tgz", "tar":
		f, err := os.Open(archivePath)
		if err != nil {
			return fmt.Errorf("invalid archive: %w", err)
		}
		defer f.Close()
		var reader io.Reader = f
		if ext == "tar.gz" || ext == "tgz" {
			gz, gzErr := gzip.NewReader(f)
			if gzErr != nil {
				return fmt.Errorf("invalid gzip archive: %w", gzErr)
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
				return fmt.Errorf("invalid tar archive: %w", err)
			}
			if !isSafeMember(destinationPath, hdr.Name) {
				return fmt.Errorf("unsafe file path in archive: %s", hdr.Name)
			}
		}
	}
	return nil
}

// downloadFile fetches url into a new temp file under dir, returning its
// path. Capped at 200MB to keep a misbehaving/huge remote file from filling
// disk.
func downloadFile(ctx context.Context, url, dir, ext string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 2 * time.Minute}
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
		os.Remove(out.Name())
		return "", err
	}
	if info, statErr := out.Stat(); statErr == nil && info.Size() > maxBytes {
		os.Remove(out.Name())
		return "", fmt.Errorf("archive too large (over 200MB)")
	}
	return out.Name(), nil
}

// archiveExt returns the recognized archive extension for url ("zip",
// "tar.gz", "tgz", or "tar"), matching isArchiveURL's regex.
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

// downloadAndExtractInitialProject downloads an archive URL and extracts it
// into hostDestPath (a host-filesystem path, e.g. resolved via
// paths.SecureUserPath("HOME", ...) - the html_data volume is bind-mounted
// there, so no podman exec is needed for this part; only Composer itself
// needs to run inside the php-fpm container, since only it has PHP).
func downloadAndExtractInitialProject(ctx context.Context, url, hostDestPath string) error {
	ext := archiveExt(url)
	tmpDir := os.TempDir()
	archivePath, err := downloadFile(ctx, url, tmpDir, ext)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	if err := validateArchiveMembers(archivePath, ext, hostDestPath); err != nil {
		return err
	}

	extractCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	switch ext {
	case "zip":
		cmd = exec.CommandContext(extractCtx, "unzip", "-o", archivePath, "-d", hostDestPath)
	case "tar.gz", "tgz":
		cmd = exec.CommandContext(extractCtx, "tar", "-xzf", archivePath, "-C", hostDestPath)
	case "tar":
		cmd = exec.CommandContext(extractCtx, "tar", "-xf", archivePath, "-C", hostDestPath)
	}
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("extraction failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
