package docker

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// baseLayout builds a representative web.LayoutData shared by every render test below - the same app-shell fields BuildLayoutData would produce, without needing a live *appctx.App/http.Request
func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "docker": true}
	return web.LayoutData{
		Title:           "Test",
		BrandName:       "Test Panel",
		CSRFToken:       "test-csrf-token",
		PanelDir:        "ltr",
		PanelVersion:    "1.0.0",
		NavGroups:       web.BuildClassicSidebarNav(userAllowed, nil, path),
		UserAllowed:     userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed),
		CurrentUsername: "testuser",
		RequestPath:     path,
		AdminPort:       "2087",
		T:               mgr.Translator("en"),
	}
}

// exercises containers.html with a representative mix of rows: unlimited/limited CPU+RAM, a trusted and an untrusted image, and a core service (no Edit/Delete) alongside a user-added one
func TestRenderContainersPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ContainersPageData{
		LayoutData: baseLayout(mgr, "/containers"),
		TotalCPU:   4,
		TotalRAM:   8,
		Rows: []ContainerRow{
			{Service: "nginx", DisplayName: "nginx", Image: "nginx:latest", ImageTrusted: true, CPUUnlimited: true, RAMUnlimited: true, ShowManage: false},
			{Service: "myapp", DisplayName: "myapp", Image: "custom/myapp:v1", ImageTrusted: false, CPUValue: "1.5", RAMGB: "2", ShowManage: true},
		},
	}
	w := httptest.NewRecorder()
	if err := containersPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Test Panel", "nginx:latest", "custom/myapp:v1", "/containers/edit/myapp", "/containers/delete/myapp"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered containers page missing %q", want)
		}
	}
	if strings.Contains(body, "/containers/edit/nginx") {
		t.Error("core service nginx should not show an Edit link")
	}
}

// covers all three form states: fresh add, failed-validation redisplay, and GET-edit prefill
func TestRenderContainerFormPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("add", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ContainerFormPageData{
			LayoutData: baseLayout(mgr, "/containers/new"), Title: "Add service",
			AvailableNetworks: []string{"web"}, CPU: "0.5", RAM: "1G",
		}
		if err := containerFormPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `value="0.5"`) {
			t.Error("expected default CPU value in rendered form")
		}
	})

	t.Run("validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ContainerFormPageData{
			LayoutData: baseLayout(mgr, "/containers/new"), Title: "Add service",
			Error: "Invalid service name.", ServiceName: "1bad", Image: "nginx",
		}
		if err := containerFormPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Invalid service name.") {
			t.Error("expected error message in rendered form")
		}
	})

	t.Run("edit prefill", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ContainerFormPageData{
			LayoutData: baseLayout(mgr, "/containers/edit/myapp"), Title: "Edit service myapp",
			Editing: true, ServiceName: "myapp", ServiceNameReadonly: true, Image: "custom/myapp:v1",
			CPU: "1.5", RAM: "2G", VolumeEntries: []VolumeEntry{{Name: "data", Mount: "/data", ReadOnly: true}},
		}
		if err := containerFormPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "readonly") || !strings.Contains(body, `value="data"`) {
			t.Error("expected readonly service name and prefilled volume in edit form")
		}
	})
}

func TestRenderDeleteConfirmPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	data := DeleteConfirmPageData{LayoutData: baseLayout(mgr, "/containers/delete/myapp"), Service: "myapp"}
	if err := deleteConfirmPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "myapp") {
		t.Error("expected service name in rendered delete confirmation")
	}
}

func TestRenderChangeMySQLPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	choices := []ServerChoice{
		{Value: "mysql", Label: "MySQL", Description: "d", Icon: "mysql", Current: true},
		{Value: "mariadb", Label: "MariaDB", Description: "d", Icon: "mariadb"},
		{Value: "percona", Label: "Percona", Description: "d", Icon: "percona"},
	}
	w := httptest.NewRecorder()
	data := ChangeMySQLPageData{LayoutData: baseLayout(mgr, "/containers/mysql"), Title: "Database Server", MySQLType: "mysql", Choices: choices}
	if err := changeMySQLPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="new_sql" value="percona"`) || !strings.Contains(body, "/static/img/servers/percona.svg") {
		t.Error("expected a percona card with its icon")
	}
	if !strings.Contains(body, `name="new_sql" value="mysql" checked disabled`) || !strings.Contains(body, "Choose a database server") {
		t.Error("expected the current type locked and the switch button")
	}

	w = httptest.NewRecorder()
	data.Blocked, data.DBCount = true, 3
	_ = changeMySQLPage.Render(w, 200, data)
	if body := w.Body.String(); strings.Contains(body, "Choose a database server") || !strings.Contains(body, "Go to Databases") || !strings.Contains(body, "<strong>3</strong>") {
		t.Error("blocked page should show the notice instead of the button")
	}
}

func TestDatabaseFlavorAndChoices(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()

	write := func(mysqlImage string) {
		compose := "services:\n  mysql:\n    image: " + mysqlImage + "\n  mariadb:\n    image: mariadb:latest\n"
		if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("mysql:${MYSQL_VERSION:-latest}")
	if got := databaseFlavor("u", "mysql"); got != "mysql" {
		t.Errorf("flavor = %q, want mysql", got)
	}
	write("percona/percona-server:8.0")
	if got := databaseFlavor("u", "mysql"); got != "percona" {
		t.Errorf("flavor = %q, want percona", got)
	}
	if got := databaseFlavor("u", "mariadb"); got != "mariadb" {
		t.Errorf("flavor = %q, want mariadb", got)
	}

	choices := serverChoicesFor("u", databaseChoices, "percona", databaseService)
	if len(choices) != 3 || !choices[2].Current || choices[0].Current {
		t.Errorf("all three should show with percona current, got %+v", choices)
	}
	web := serverChoicesFor("u", webserverChoices, "apache", func(v string) string { return v })
	if len(web) != 1 || web[0].Value != "apache" {
		t.Errorf("webservers missing from compose should be hidden, got %+v", web)
	}
}

func TestRenderChangeWebserverPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	choices := []ServerChoice{
		{Value: "nginx", Label: "Nginx", Description: "d", Icon: "nginx", Current: true},
		{Value: "apache", Label: "Apache", Description: "d", Icon: "apache"},
	}

	t.Run("no domains", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ChangeWebserverPageData{LayoutData: baseLayout(mgr, "/containers/webserver"), Title: "Webserver", Webserver: "nginx", Choices: choices}
		if err := changeWebserverPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, `name="new_ws" value="apache"`) || !strings.Contains(body, "/static/img/servers/apache.svg") {
			t.Error("expected an apache card with its icon")
		}
		if !strings.Contains(body, `name="new_ws" value="nginx" checked disabled`) {
			t.Error("the current webserver should be shown checked and not selectable")
		}
		if !strings.Contains(body, "Choose a webserver") {
			t.Error("expected the switch button when there are no domains")
		}
	})

	t.Run("has domains", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ChangeWebserverPageData{LayoutData: baseLayout(mgr, "/containers/webserver"), Title: "Webserver", Webserver: "nginx", Choices: choices, DomainCount: 2}
		if err := changeWebserverPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if strings.Contains(body, "Choose a webserver") {
			t.Error("switch button should be hidden while domains still exist")
		}
		if !strings.Contains(body, "Go to Domains") || !strings.Contains(body, "<strong>2</strong>") {
			t.Error("expected the remove domains notice with the count")
		}
	})
}

func TestServerChoicesFor(t *testing.T) {
	// no compose file for this context, so only the current one is kept
	got := serverChoicesFor("choicestest", webserverChoices, "nginx", func(v string) string { return v })
	if len(got) != 1 || got[0].Value != "nginx" || !got[0].Current {
		t.Errorf("got %+v", got)
	}
}

func TestRenderChangeImagePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("single service", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ChangeImagePageData{
			LayoutData: baseLayout(mgr, "/containers/image/change/nginx"),
			Service:    "nginx", CurrentVersion: "1.26",
		}
		if err := changeImagePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `value="1.26"`) {
			t.Error("expected current version prefilled in rendered form")
		}
	})

	t.Run("picker", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ChangeImagePageData{
			LayoutData:         baseLayout(mgr, "/containers/image/change"),
			SelectableServices: []string{"nginx", "myapp"},
		}
		if err := changeImagePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "nginx") || !strings.Contains(body, "myapp") {
			t.Error("expected selectable services in rendered picker")
		}
	})
}

func TestRenderLogsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	data := LogsPageData{LayoutData: baseLayout(mgr, "/containers/logs"), Services: []string{"nginx", "myapp"}}
	if err := logsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `value="nginx"`) || !strings.Contains(body, `value="myapp"`) {
		t.Error("expected service options in rendered logs page")
	}
}

func TestRenderTerminalPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("connected", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := TerminalPageData{
			LayoutData:    baseLayout(mgr, "/containers/terminal/myapp"),
			ContainerName: "myapp", TerminalTimeoutSeconds: 10,
		}
		if err := terminalPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "/ws/containers/terminal/myapp") {
			t.Error("expected websocket path in rendered terminal page")
		}
	})

	t.Run("picker", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := TerminalPageData{
			LayoutData:     baseLayout(mgr, "/containers/terminal"),
			ActiveServices: []string{"nginx", "myapp"},
		}
		if err := terminalPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "nginx") || !strings.Contains(body, "myapp") {
			t.Error("expected active services in rendered picker")
		}
	})
}

// guards against nondeterministic row order (Go map iteration) leaking into the rendered table
func TestBuildContainerRowsSorted(t *testing.T) {
	rows := buildContainerRows(map[string]any{
		"zeta":  map[string]any{"image": "z"},
		"alpha": map[string]any{"image": "a"},
		"mid":   map[string]any{"image": "m"},
	})
	if len(rows) != 3 || rows[0].Service != "alpha" || rows[1].Service != "mid" || rows[2].Service != "zeta" {
		t.Fatalf("expected rows sorted by service name, got %+v", rows)
	}
}

// guards against regressing to the "always shows unlimited/0" bug: `podman-compose config` resolves deploy.resources.limits.memory to a compose-style string like "0.5G", not a raw byte count
func TestBuildContainerRowsMemory(t *testing.T) {
	rows := buildContainerRows(map[string]any{
		"nginx": map[string]any{
			"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
				"cpus": "0.5", "memory": "0.5G",
			}}},
		},
		"unlimited": map[string]any{
			"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
				"cpus": "0", "memory": "0",
			}}},
		},
	})
	byName := map[string]ContainerRow{}
	for _, r := range rows {
		byName[r.Service] = r
	}

	nginx := byName["nginx"]
	if nginx.RAMUnlimited || nginx.RAMGB != "0.5" {
		t.Errorf("nginx: RAMUnlimited=%v RAMGB=%q, want RAMGB=\"0.5\"", nginx.RAMUnlimited, nginx.RAMGB)
	}
	if nginx.CPUUnlimited || nginx.CPUValue != "0.5" {
		t.Errorf("nginx: CPUUnlimited=%v CPUValue=%q, want CPUValue=\"0.5\"", nginx.CPUUnlimited, nginx.CPUValue)
	}

	unlimited := byName["unlimited"]
	if !unlimited.RAMUnlimited || !unlimited.CPUUnlimited {
		t.Errorf("unlimited service: RAMUnlimited=%v CPUUnlimited=%v, want both true", unlimited.RAMUnlimited, unlimited.CPUUnlimited)
	}
}

func TestParseMemoryBytes(t *testing.T) {
	cases := []struct {
		in     string
		wantOK bool
		wantGB string
	}{
		{"0.5G", true, "0.5"},
		{"512M", true, "0.5"},
		{"2G", true, "2"},
		{"1073741824", true, "1"},
		{"1GB", true, "1"},
		{"", false, ""},
		{"unlimited", false, ""},
	}
	for _, c := range cases {
		bytes, ok := parseMemoryBytes(c.in)
		if ok != c.wantOK {
			t.Errorf("parseMemoryBytes(%q) ok = %v, want %v", c.in, ok, c.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if got := formatGB(bytes); got != c.wantGB {
			t.Errorf("parseMemoryBytes(%q) -> formatGB = %q, want %q", c.in, got, c.wantGB)
		}
	}
}
