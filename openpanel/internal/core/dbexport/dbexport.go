// Package dbexport delivers a database dump to the browser or into the user's /var/www/html, shared by the MySQL, PostgreSQL and MongoDB export routes.
package dbexport

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// WebRoot is the only folder users can export into, as they see it
const WebRoot = "/var/www/html"

var (
	ErrPathRequired = errors.New("Destination path is required for folder export.")
	ErrPathOutside  = errors.New("Invalid export path. Must be inside /var/www/html/")
	// ErrDumpFailed means the dump command itself failed, so callers can show their own "failed to export" message
	ErrDumpFailed = errors.New("dump command failed")
)

// Send runs cmd and streams its stdout to the browser as a download, gzipping on the fly; an error means nothing was sent yet
func Send(w http.ResponseWriter, cmd *exec.Cmd, filename, contentType string, gz bool) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// wait for the first byte so a dump that fails right away (bad db, engine down) is reported before headers go out
	br := bufio.NewReaderSize(out, 64<<10)
	if _, err := br.Peek(1); err != nil {
		if waitErr := cmd.Wait(); waitErr != nil {
			return waitErr
		}
		return ErrDumpFailed
	}

	if gz {
		filename += ".gz"
		contentType = "application/gzip"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	_, copyErr := copyDump(w, br, gz)
	if copyErr != nil {
		_ = cmd.Process.Kill()
	}
	if waitErr := cmd.Wait(); copyErr != nil || waitErr != nil {
		// headers are already out, so drop the connection to make the browser mark the download as failed instead of saving a truncated file
		panic(http.ErrAbortHandler)
	}
	return nil
}

// hostDir maps a /var/www/html path to the user's html volume on the host
func hostDir(userContext, localPath string) (string, error) {
	if localPath == "" {
		return "", ErrPathRequired
	}
	clean := filepath.Clean(localPath)
	if clean != WebRoot && !strings.HasPrefix(clean, WebRoot+string(filepath.Separator)) {
		return "", ErrPathOutside
	}
	rel, err := filepath.Rel(WebRoot, clean)
	if err != nil {
		return "", ErrPathOutside
	}
	return filepath.Join("/home/"+userContext+"/docker-data/volumes/"+userContext+"_html_data/_data/", rel), nil
}

// SaveToFiles runs cmd and streams its stdout to <localPath>/<base>_<timestamp><ext>[.gz], owned by the user, and returns the path as the user sees it
func SaveToFiles(cmd *exec.Cmd, userContext, localPath, base, ext string, gz bool) (string, error) {
	dir, err := hostDir(userContext, localPath)
	if err != nil {
		return "", err
	}
	if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
		return "", errors.New("Export path " + localPath + " does not exist or is not a directory.")
	}

	name := base + "_" + time.Now().Format("2006-01-02_15-04-05") + ext
	if gz {
		name += ".gz"
	}
	dest := filepath.Join(dir, name)
	display := filepath.Join(localPath, name)
	writeErr := errors.New("Failed to write export to " + display + " - try export to browser instead.")

	f, err := os.Create(dest)
	if err != nil {
		return "", writeErr
	}
	defer f.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		_ = os.Remove(dest)
		return "", ErrDumpFailed
	}
	if err := cmd.Start(); err != nil {
		_ = os.Remove(dest)
		return "", ErrDumpFailed
	}
	n, copyErr := copyDump(f, out, gz)
	if copyErr != nil {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	closeErr := f.Close()
	switch {
	case waitErr != nil || (copyErr == nil && n == 0):
		_ = os.Remove(dest)
		return "", ErrDumpFailed
	case copyErr != nil || closeErr != nil:
		_ = os.Remove(dest)
		return "", writeErr
	}

	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		id := strconv.Itoa(uid)
		_ = exec.Command("chown", id+":"+id, dest).Run()
	}
	return display, nil
}

// copyDump copies src into dst, through gzip when asked, and returns how many raw bytes it read
func copyDump(dst io.Writer, src io.Reader, gz bool) (int64, error) {
	if !gz {
		return io.Copy(dst, src)
	}
	zw := gzip.NewWriter(dst)
	n, err := io.Copy(zw, src)
	if closeErr := zw.Close(); err == nil {
		err = closeErr
	}
	return n, err
}
