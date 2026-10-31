package wordpress

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// the Debugging tab shows the end of debug.log, never the whole thing
const (
	debugLogMaxBytes = 256 << 10
	debugLogMaxLines = 200
)

// errDebugLogPath is a WP_DEBUG_LOG the panel won't read, the message says why
type errDebugLogPath struct{ msg string }

func (e *errDebugLogPath) Error() string { return e.msg }

var errNotRegular = errors.New("debug.log isn't a regular file.")

// parseDebugLogSetting reads `wp config get WP_DEBUG_LOG --format=json`, a custom path comes back as a JSON string, anything else means wp-content/debug.log
func parseDebugLogSetting(out string) string {
	out = strings.TrimSpace(out)
	var v any
	if json.Unmarshal([]byte(out), &v) != nil {
		v = out
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "true", "1", "false", "0":
		return ""
	}
	return strings.TrimSpace(s)
}

// debugLogLocation maps the setting to a host path, a custom path has to be absolute and inside the site's docroot
func debugLogLocation(docroot, hostPath, setting string) (host, shown string, custom bool, err error) {
	if setting == "" {
		return filepath.Join(hostPath, "wp-content", "debug.log"), "wp-content/debug.log", false, nil
	}
	if strings.ContainsRune(setting, 0) || !strings.HasPrefix(setting, "/") {
		return "", "", true, &errDebugLogPath{"WP_DEBUG_LOG is set to a relative path (" + setting + "), the panel can only show a log inside this site's folder."}
	}
	rel, ok := strings.CutPrefix(path.Clean(setting), strings.TrimSuffix(docroot, "/")+"/")
	if !ok || rel == "" {
		return "", "", true, &errDebugLogPath{"WP_DEBUG_LOG points outside this site's folder (" + setting + "), the panel can only show a log inside it."}
	}
	return filepath.Join(hostPath, filepath.FromSlash(rel)), rel, true, nil
}

// debugLogFile asks wp-cli where the site logs to, if that fails it still looks in wp-content
func (s *wpToolSite) debugLogFile(ctx context.Context) (host, shown string, custom bool, err error) {
	setting := ""
	out, cliErr := s.wp(ctx, time.Minute, "config", "get", "WP_DEBUG_LOG", "--type=constant", "--format=json", "--skip-plugins", "--skip-themes")
	if cliErr == nil {
		setting = parseDebugLogSetting(out)
	}
	return debugLogLocation(s.Docroot, s.HostPath, setting)
}

// openRegular refuses fifos, devices and symlinks before opening, a fifo would hang the read
func openRegular(p string, flag int) (*os.File, error) {
	info, err := tenantfs.Lstat(p)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errNotRegular
	}
	f, err := tenantfs.OpenFile(p, flag, 0)
	if err != nil {
		return nil, err
	}
	if st, statErr := f.Stat(); statErr != nil || !st.Mode().IsRegular() {
		f.Close()
		return nil, errNotRegular
	}
	return f, nil
}

// tailLines reads at most maxBytes from the end of p and returns the last maxLines lines, more is true when older lines were left out
func tailLines(p string, maxBytes int64, maxLines int) (lines []string, size int64, more bool, err error) {
	f, err := openRegular(p, os.O_RDONLY)
	if err != nil {
		return nil, 0, false, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, 0, false, err
	}
	size = st.Size()
	start := max(size-maxBytes, 0)
	data, err := io.ReadAll(io.NewSectionReader(f, start, size-start))
	if err != nil {
		return nil, size, false, err
	}
	if start > 0 {
		more = true
		// the first line is cut in half, drop it
		if i := strings.IndexByte(string(data), '\n'); i >= 0 {
			data = data[i+1:]
		} else {
			data = nil
		}
	}
	text := strings.TrimRight(string(data), "\r\n")
	lines = []string{}
	if text != "" {
		lines = strings.Split(text, "\n")
	}
	if len(lines) > maxLines {
		lines, more = lines[len(lines)-maxLines:], true
	}
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines, size, more, nil
}

type debugLogResult struct {
	Exists bool     `json:"exists"`
	Path   string   `json:"path"`
	Custom bool     `json:"custom"`
	Size   int64    `json:"size"`
	Lines  []string `json:"lines"`
	More   bool     `json:"more"`
}

func debugLogPathError(w http.ResponseWriter, err error) bool {
	var pe *errDebugLogPath
	if errors.As(err, &pe) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": pe.msg})
		return true
	}
	return false
}

// readDebugLog builds the GET answer for host, a missing file isn't an error
func readDebugLog(host, shown string, custom bool) (debugLogResult, error) {
	res := debugLogResult{Path: shown, Custom: custom, Lines: []string{}}
	lines, size, more, err := tailLines(host, debugLogMaxBytes, debugLogMaxLines)
	if errors.Is(err, fs.ErrNotExist) {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	res.Exists, res.Size, res.Lines, res.More = true, size, lines, more
	return res, nil
}

func handleWPToolDebugLog(w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	host, shown, custom, err := s.debugLogFile(r.Context())
	if debugLogPathError(w, err) {
		return
	}
	res, err := readDebugLog(host, shown, custom)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Couldn't read " + shown + ": " + errText(err)})
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

// clearDebugLog empties the file in place so its owner and mode stay as they were
func clearDebugLog(host string) error {
	f, err := openRegular(host, os.O_WRONLY)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Truncate(0)
}

func handleWPToolDebugLogClear(a *appctx.App, w http.ResponseWriter, r *http.Request, s *wpToolSite) {
	host, shown, _, err := s.debugLogFile(r.Context())
	if debugLogPathError(w, err) {
		return
	}
	if err := clearDebugLog(host); errors.Is(err, fs.ErrNotExist) {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "There's no " + shown + " to clear."})
		return
	} else if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Couldn't clear " + shown + ": " + errText(err)})
		return
	}
	s.record(a, r, "cleared "+shown)
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": shown + " cleared."})
}

// errText keeps host paths out of messages shown in the browser
func errText(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
