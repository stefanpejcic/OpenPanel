// Package dbexport delivers a database dump to the browser or into the user's /var/www/html, shared by the MySQL, PostgreSQL and MongoDB export routes.
package dbexport

import (
	"bytes"
	"compress/gzip"
	"errors"
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
)

// Send writes data as a file download, gzipping it first when asked
func Send(w http.ResponseWriter, filename, contentType string, data []byte, gz bool) {
	if gz {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, _ = zw.Write(data)
		_ = zw.Close()
		data = buf.Bytes()
		filename += ".gz"
		contentType = "application/gzip"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write(data)
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

// SaveToFiles writes data to <localPath>/<base>_<timestamp><ext>[.gz], owned by the user, and returns the path as the user sees it
func SaveToFiles(userContext, localPath, base, ext string, data []byte, gz bool) (string, error) {
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

	if err := writeFile(dest, data, gz); err != nil {
		_ = os.Remove(dest)
		return "", errors.New("Failed to write export to " + display + " - try export to browser instead.")
	}
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		id := strconv.Itoa(uid)
		_ = exec.Command("chown", id+":"+id, dest).Run()
	}
	return display, nil
}

func writeFile(path string, data []byte, gz bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if !gz {
		_, err = f.Write(data)
		return err
	}
	zw := gzip.NewWriter(f)
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}
