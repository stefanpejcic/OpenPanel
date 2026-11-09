package filemanager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func TestLsModeOctal(t *testing.T) {
	for perm, want := range map[string]string{
		"-rw-r--r--":  "644",
		"drwxr-xr-x":  "755",
		"drwxrwxr-x":  "775",
		"-rw-r-----":  "640",
		"-rw-------":  "600",
		"----------":  "000",
		"-rwxrwxrwx":  "777",
		"lrwxrwxrwx":  "777",
		"-rw-r--r--.": "644",
		"drwxr-xr-x+": "755",
		// special bits only count for their execute bit, the panel never sets them
		"drwxrwxrwt": "777",
		"drwxrwxrwT": "776",
		"-rwsr-xr-x": "755",
		"-rwSr--r--": "644",
		"":           "",
		"-rw-r--r":   "",
		"-rw-r?-r--": "",
	} {
		if got := lsModeOctal(perm); got != want {
			t.Errorf("lsModeOctal(%q) = %q, want %q", perm, got, want)
		}
	}
}

func TestViewRowWithMode(t *testing.T) {
	for _, tc := range []struct {
		row       ViewRow
		mode, sym string
		world     bool
	}{
		{ViewRow{Type: "file", Permissions: "-rw-r--r--"}, "644", "-rw-r--r--", false},
		{ViewRow{Type: "directory", Permissions: "drwxrwxrwx."}, "777", "drwxrwxrwx", true},
		{ViewRow{Type: "file", Permissions: "-rw-rw--w-"}, "662", "-rw-rw--w-", true},
		// a symlink is always lrwxrwxrwx, that isn't a risk of its own
		{ViewRow{Type: "symlink", Permissions: "lrwxrwxrwx"}, "777", "lrwxrwxrwx", false},
		{ViewRow{Type: "unknown", Permissions: "?"}, "", "?", false},
	} {
		got := tc.row.withMode()
		if got.Mode != tc.mode || got.Sym != tc.sym || got.WorldWritable != tc.world {
			t.Errorf("withMode(%q %s) = mode %q sym %q world %v, want %q %q %v", tc.row.Permissions, tc.row.Type, got.Mode, got.Sym, got.WorldWritable, tc.mode, tc.sym, tc.world)
		}
	}
}

// the permissions sheet posts with Accept: application/json and shows the server's refusal as a toast
func TestChangePermissionsJSONRefusal(t *testing.T) {
	a := &appctx.App{
		Sessions: session.NewStore([]byte("test-secret-key-0123456789abcdef")),
		I18n:     i18n.NewManager(t.TempDir(), cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))),
	}
	for value, want := range map[string]string{
		"4755": "Special permission bits (setuid, setgid, sticky) are not supported, use three digits like 755.",
		"888":  "Invalid permissions format.",
	} {
		form := url.Values{"permissions": {value}, "filename": {"a.txt"}, "path_param": {"site"}}
		req := httptest.NewRequest(http.MethodPost, "/file-manager/permissions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		handleChangePermissions(a, w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status %d, want 400", value, w.Code)
		}
		var res struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("%s: body isn't JSON: %v %q", value, err, w.Body.String())
		}
		if res.Success || res.Error != want {
			t.Errorf("%s: got %+v, want error %q", value, res, want)
		}
	}

	// the plain form post keeps redirecting with a flash
	form := url.Values{"permissions": {"4755"}, "path_param": {"site"}}
	req := httptest.NewRequest(http.MethodPost, "/file-manager/permissions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleChangePermissions(a, w, req)
	if w.Code != http.StatusFound || w.Header().Get("Location") != "/files/site" {
		t.Errorf("form post: status %d location %q, want 302 to /files/site", w.Code, w.Header().Get("Location"))
	}
}
