package backups

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// lists saved at reindex, one per archive and kind, so browsing a backup never touches the destination
func listPath(userHome, backup, kind string) string {
	return filepath.Join(userHome, "backup_files", filepath.Base(backup)+"."+kind+".tsv")
}

func listsExist(userHome, backup string, kinds []string) bool {
	for _, kind := range kinds {
		if _, err := os.Stat(listPath(userHome, backup, kind)); err != nil {
			return false
		}
	}
	return true
}

func saveList(userHome, backup, kind string, r io.Reader) error {
	dest := listPath(userHome, backup, kind)
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dest)
}

// pruneFileLists drops lists of archives that are gone from the destination
func pruneFileLists(userHome string, backups []string) {
	keep := map[string]bool{}
	for _, b := range backups {
		keep[filepath.Base(b)] = true
	}
	entries, _ := os.ReadDir(filepath.Join(userHome, "backup_files"))
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".tsv")
		if dot := strings.LastIndex(name, "."); dot > 0 && keep[name[:dot]] {
			continue
		}
		_ = os.Remove(filepath.Join(userHome, "backup_files", e.Name()))
	}
}

type backupDirEntry struct {
	Name  string `json:"name"`
	Files int    `json:"files"`
	Size  int64  `json:"size"`
}

type backupFileEntry struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mtime int64  `json:"mtime"`
}

// listBackupDir returns the folders (with file count and total size) and files directly inside dir of a saved file list
func listBackupDir(listPath, dir string) ([]backupDirEntry, []backupFileEntry, error) {
	f, err := os.Open(listPath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	prefix := ""
	if dir != "" {
		prefix = dir + "/"
	}
	dirs := map[string]*backupDirEntry{}
	var files []backupFileEntry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		fields := strings.SplitN(sc.Text(), "\t", 3)
		if len(fields) != 3 || !strings.HasPrefix(fields[2], prefix) {
			continue
		}
		size, _ := strconv.ParseInt(fields[0], 10, 64)
		rest := fields[2][len(prefix):]
		if name, _, nested := strings.Cut(rest, "/"); nested {
			d := dirs[name]
			if d == nil {
				d = &backupDirEntry{Name: name}
				dirs[name] = d
			}
			d.Files++
			d.Size += size
			continue
		}
		mtime, _ := strconv.ParseInt(fields[1], 10, 64)
		files = append(files, backupFileEntry{Name: rest, Size: size, Mtime: mtime})
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}

	dirList := make([]backupDirEntry, 0, len(dirs))
	for _, d := range dirs {
		dirList = append(dirList, *d)
	}
	sort.Slice(dirList, func(i, j int) bool { return dirList[i].Name < dirList[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	if files == nil {
		files = []backupFileEntry{}
	}
	return dirList, files, nil
}

// cleanBackupPath turns a user supplied path under html/ into a clean relative one, false for anything trying to leave it
func cleanBackupPath(p string) (string, bool) {
	p = strings.Trim(p, "/")
	if p == "" {
		return "", true
	}
	clean := path.Clean(p)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "\x00") {
		return "", false
	}
	return clean, true
}

// walkCompare orders archive paths the way docker-volume-backup's filepath.WalkDir wrote them: name by name, a folder right before its contents
func walkCompare(a, b string) int {
	pa, pb := strings.Split(a, "/"), strings.Split(b, "/")
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] != pb[i] {
			return strings.Compare(pa[i], pb[i])
		}
	}
	return len(pa) - len(pb)
}

func insidePath(name, target string) bool {
	return name == target || strings.HasPrefix(name, target+"/")
}

// pastAll is true once name is beyond every target in walk order, nothing wanted can come after that
func pastAll(name string, targets []string) bool {
	for _, t := range targets {
		if insidePath(name, t) || walkCompare(name, t) < 0 {
			return false
		}
	}
	return true
}

var errFileNotInBackup = errors.New("file not found in this backup")

// streamHTMLFile finds one file under html/ in the archive and hands its content to send, without reading past it
func streamHTMLFile(ctx context.Context, store remoteStore, backup, rel string, send func(size int64, r io.Reader) error) error {
	target := "backup/html/" + rel
	rc, err := store.Open(ctx, backup)
	if err != nil {
		return err
	}
	defer rc.Close()
	gz, err := gzip.NewReader(rc)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return errFileNotInBackup
		}
		if err != nil {
			return err
		}
		name := strings.Trim(header.Name, "/")
		if name == target && header.Typeflag == tar.TypeReg {
			return send(header.Size, tr)
		}
		if pastAll(name, []string{target}) {
			return errFileNotInBackup
		}
	}
}

// backupFileParams reads and checks the backup_file/backup and path params shared by the file browser endpoints
func backupFileParams(r *http.Request, backupKey string) (string, string, bool) {
	backup := strings.TrimSpace(r.Form.Get(backupKey))
	if backup == "" || filepath.Base(backup) != backup {
		return "", "", false
	}
	rel, ok := cleanBackupPath(r.Form.Get("path"))
	return backup, rel, ok
}

// handleBackupFiles lists one folder of a backup's website files from the list saved at reindex
func handleBackupFiles(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	backup, dir, ok := backupFileParams(r, "backup")
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup or path."})
		return
	}
	dirs, files, err := listBackupDir(listPath("/home/"+userContext, backup, "files"), dir)
	if err != nil {
		if os.IsNotExist(err) {
			web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No file list for this backup, it was made before file browsing was available."})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"path": dir, "dirs": dirs, "files": files})
}

// handleDownloadBackupFile streams a single website file out of a backup archive to the browser
func handleDownloadBackupFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseMultipartForm(1 << 20)
	backup, rel, ok := backupFileParams(r, "backup_file")
	if !ok || rel == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup or path."})
		return
	}
	config, _ := readBackupEnv(userContext)
	store, err := newRemoteStore(config)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = streamHTMLFile(r.Context(), store, backup, rel, func(size int64, body io.Reader) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+strings.ReplaceAll(path.Base(rel), "\"", "")+"\"")
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		_, copyErr := io.Copy(w, body)
		return copyErr
	})
	switch {
	case errors.Is(err, errFileNotInBackup):
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "File not found in this backup."})
	case err != nil && w.Header().Get("Content-Length") == "":
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	case err == nil:
		_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded "+rel+" from backup "+backup, reqip.ClientIP(r))
	}
}
