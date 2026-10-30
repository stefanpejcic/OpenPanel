package wordpress

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// the Maintenance tab edits wp-content/maintenance.php, WordPress serves it while maintenance mode is on
const (
	maintenancePageShown = "wp-content/maintenance.php"
	maintenancePageMax   = 256 << 10
)

var (
	errMaintenanceTooBig = errors.New("wp-content/maintenance.php is over 256 KB, edit it in the File Manager.")
	errMaintenanceText   = errors.New("The page has to be plain UTF-8 text.")
	errNoWPContent       = errors.New("This site has no wp-content folder.")
)

type maintenancePage struct {
	Exists   bool   `json:"exists"`
	Path     string `json:"path"`
	Code     string `json:"code"`
	Modified int64  `json:"modified"` // unix seconds, 0 while there's no file
}

func maintenancePagePath(hostPath string) string {
	return filepath.Join(hostPath, "wp-content", "maintenance.php")
}

// readMaintenancePage returns the file's code, a missing file isn't an error
func readMaintenancePage(host string) (maintenancePage, error) {
	res := maintenancePage{Path: maintenancePageShown}
	f, err := openRegular(host, os.O_RDONLY)
	if errors.Is(err, fs.ErrNotExist) {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return res, err
	}
	if st.Size() > maintenancePageMax {
		return res, errMaintenanceTooBig
	}
	data, err := io.ReadAll(io.LimitReader(f, maintenancePageMax+1))
	if err != nil {
		return res, err
	}
	if len(data) > maintenancePageMax {
		return res, errMaintenanceTooBig
	}
	res.Exists, res.Code, res.Modified = true, string(data), st.ModTime().Unix()
	return res, nil
}

// writeMaintenancePage rewrites the file in place so its owner and mode stay, a new file gets wp-content's owner
func writeMaintenancePage(host, code string) error {
	if len(code) > maintenancePageMax {
		return errMaintenanceTooBig
	}
	if !utf8.ValidString(code) || strings.ContainsRune(code, 0) {
		return errMaintenanceText
	}
	f, err := openRegular(host, os.O_WRONLY|os.O_TRUNC)
	if err == nil {
		_, werr := f.WriteString(code)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		return werr
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	dir, err := tenantfs.Lstat(filepath.Dir(host))
	if err != nil || !dir.IsDir() {
		return errNoWPContent
	}
	if err := tenantfs.WriteFile(host, []byte(code), 0o644); err != nil {
		return err
	}
	if st, ok := dir.Sys().(*syscall.Stat_t); ok {
		_ = tenantfs.Lchown(host, int(st.Uid), int(st.Gid))
	}
	return nil
}

func maintenanceError(w http.ResponseWriter, err error, verb string) {
	switch {
	case errors.Is(err, errMaintenanceTooBig), errors.Is(err, errMaintenanceText), errors.Is(err, errNoWPContent):
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, errNotRegular):
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": maintenancePageShown + " isn't a regular file, edit it in the File Manager."})
	default:
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Couldn't " + verb + " " + maintenancePageShown + ": " + errText(err)})
	}
}

// fileManagerAllowed: the page is PHP in the site's folder, so editing it needs the same feature as the File Manager
func fileManagerAllowed(a *appctx.App, r *http.Request) bool {
	userID, ok := auth.UserID(r)
	if !ok {
		return false
	}
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return false
	}
	list, _ := injected["user_allowed"].([]string)
	for _, f := range list {
		if f == "filemanager" {
			return true
		}
	}
	return false
}

func denyWithoutFileManager(a *appctx.App, w http.ResponseWriter, r *http.Request) bool {
	if fileManagerAllowed(a, r) {
		return false
	}
	web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Editing the maintenance page needs the File Manager feature."})
	return true
}

func handleWPToolMaintenancePage(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	if denyWithoutFileManager(a, w, r) {
		return
	}
	res, err := readMaintenancePage(maintenancePagePath(s.HostPath))
	if err != nil {
		maintenanceError(w, err, "read")
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

func handleWPToolMaintenancePageSave(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	if denyWithoutFileManager(a, w, r) {
		return
	}
	code := r.FormValue("code")
	if strings.TrimSpace(code) == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "The page is empty."})
		return
	}
	if err := writeMaintenancePage(maintenancePagePath(s.HostPath), code); err != nil {
		maintenanceError(w, err, "save")
		return
	}
	s.record(a, r, "saved "+maintenancePageShown)
	web.WriteJSON(w, http.StatusOK, map[string]any{"message": "Saved " + maintenancePageShown + ".", "modified": time.Now().Unix()})
}
