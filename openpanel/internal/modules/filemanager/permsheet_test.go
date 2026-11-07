package filemanager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
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

// the Permissions column and the permissions sheet, on testinguser's real wp.tests.openpanel.org listing plus a world-writable file and a symlink
func TestRenderFilesPermissions(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	raw := []ViewRow{
		{Name: "wp-admin", Type: "directory", Size: "4096", Date: "Oct 29 07:53", Permissions: "drwxr-xr-x"},
		{Name: "wp-content", Type: "directory", Size: "127", Date: "Oct 29 17:44", Permissions: "drwxr-xr-x"},
		{Name: "wp-includes", Type: "directory", Size: "12288", Date: "Oct 29 07:53", Permissions: "drwxr-xr-x"},
		{Name: "cache", Type: "directory", Size: "57", Date: "Oct 29 18:50", Permissions: "drwxrwxrwt"},
		{Name: ".htaccess", Type: "file", Size: "234", Date: "Oct 29 07:53", Permissions: "-rw-r--r--"},
		{Name: "index.php", Type: "file", Size: "405", Date: "Oct 29 07:53", Permissions: "-rw-r--r--"},
		{Name: "license.txt", Type: "file", Size: "19903", Date: "Oct 29 07:53", Permissions: "-rw-r--r--"},
		{Name: "upload.php", Type: "file", Size: "1210", Date: "Oct 29 18:02", Permissions: "-rw-rw-rw-"},
		{Name: "wp-config.php", Type: "file", Size: "3115", Date: "Oct 29 17:49", Permissions: "-rw-------"},
		{Name: "wp-cron.php", Type: "file", Size: "5617", Date: "Oct 29 07:53", Permissions: "-rw-r--r--"},
		{Name: "wp-login.php", Type: "file", Size: "52536", Date: "Oct 29 07:53", Permissions: "-rw-r--r--"},
		{Name: "current", Type: "symlink", Size: "7", Date: "Oct 29 07:53", Permissions: "lrwxrwxrwx", LinkTarget: "wp-content"},
	}
	var dirs, files, rows []ViewRow
	for i, r := range raw {
		r.Index, r.Owner, r.Group, r.Links = i, "1002", "1002", "1"
		r = r.withMode()
		if r.Type == "directory" {
			r.Href = "/files/wp.tests.openpanel.org/" + r.Name + "?view=modern"
			dirs = append(dirs, r)
		} else {
			r.DownloadURL = "/file-manager/download-file/" + r.Name + "?path_param=wp.tests.openpanel.org"
			files = append(files, r)
		}
		rows = append(rows, r)
	}
	layout := baseLayout(mgr, "/files/wp.tests.openpanel.org")
	layout.UserAllowed["fix_permissions"] = true
	data := FilesPageData{
		LayoutData: layout, PathParam: "wp.tests.openpanel.org", View: "modern",
		Breadcrumbs: buildBreadcrumbs("wp.tests.openpanel.org"),
		Directories: dirs, Files: files, Rows: rows,
		ShowUpLink: true, UpLink: "/files?view=modern",
		FilemanagerUploadSize: 500, FilemanagerEditSize: 5, FilemanagerViewSize: 5, FilemanagerDownloadSize: 500, FilemanagerUploadBatchSize: 1000,
		Extensions: ".txt .php .htaccess", Images: ".jpg .png", Archives: ".zip .tar",
		CurrentPage: 1, TotalPages: 1, TotalFiles: len(rows), StartLineNumber: 1, EndLineNumber: len(rows),
		PageEntries: buildPageEntries(1, 1),
	}
	w := httptest.NewRecorder()
	if err := filesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		`data-mode="600" data-owner="1002" data-group="1002"`,
		`<span class="op-perm">0600</span><span class="fm-perm-sym">-rw-------</span>`,
		`aria-label="Permissions 0644, change for index.php"`,
		`title="Change permissions of wp-config.php"`,
		// the sheet, its presets and the server's validation rule
		`x-data="fmPermSheet()"`, "Change permissions", "Apply to everything inside", `href="/fix-permissions"`,
		"Group reads, others get nothing. Fits wp-config.php.",
		"Special bits (setuid, setgid, sticky) aren't supported. Use three digits like 755.",
		`"Permissions {mode}, change for {name}"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered files page missing %q", want)
		}
	}
	// world-writable modes are red, the symlink's lrwxrwxrwx is plain text and not a button
	risky := regexp.MustCompile(`class="fm-perm is-risky"[^>]*title="Change permissions of ([^"]+)"`).FindAllStringSubmatch(body, -1)
	var names []string
	for _, m := range risky {
		names = append(names, m[1])
	}
	if strings.Join(names, ",") != "cache,upload.php" {
		t.Errorf("risky cells = %v, want cache and upload.php", names)
	}
	if strings.Contains(body, "Change permissions of current") || !strings.Contains(body, `<span class="fm-perm is-static" title="A symlink uses the permissions of what it points to.">`) {
		t.Error("the symlink's permissions should be plain text")
	}
	if strings.Contains(body, "permissionCell") {
		t.Error("the old hover-to-see-rwx cell is still on /files")
	}

	// without the Fix Permissions feature the sheet doesn't link to it
	delete(layout.UserAllowed, "fix_permissions")
	data.LayoutData = layout
	w = httptest.NewRecorder()
	if err := filesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), `href="/fix-permissions"`) {
		t.Error("Fix Permissions link shown to a user without the feature")
	}
}
