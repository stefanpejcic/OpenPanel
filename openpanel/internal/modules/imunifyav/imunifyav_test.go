package imunifyav

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// fakeAgent answers on a unix socket like imunify's simple_rpc.sock and records what it got
func fakeAgent(t *testing.T, reply func(req map[string]any) string) *[]map[string]any {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rpc.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Skipf("unix sockets not available: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	old := socketPath
	socketPath = path
	t.Cleanup(func() { socketPath = old })

	var got []map[string]any
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			line, _ := bufio.NewReader(conn).ReadBytes('\n')
			var req map[string]any
			_ = json.Unmarshal(line, &req)
			got = append(got, req)
			conn.Write([]byte(strings.ReplaceAll(reply(req), "\n", "") + "\n"))
			conn.Close()
		}
	}()
	return &got
}

func TestCallSendsCommandAndDecodesData(t *testing.T) {
	got := fakeAgent(t, func(map[string]any) string {
		return `{"result": "success", "messages": [], "data": {"max_count": 1, "items": [{"user": "demo", "infected": 2, "infected_db": 1, "scan_date": 1794297231, "scan_status": "stopped"}]}}`
	})
	if !available() {
		t.Fatal("socket should count as available")
	}
	st, err := fetchStatus("demo")
	if err != nil {
		t.Fatal(err)
	}
	if st.Infected+st.InfectedDB != 3 || st.ScanDate == nil || st.scanning() {
		t.Fatalf("bad status %+v", st)
	}
	req := (*got)[0]
	if cmd, _ := json.Marshal(req["command"]); string(cmd) != `["malware","user","list"]` {
		t.Fatalf("command = %s", cmd)
	}
	if req["params"].(map[string]any)["user"] != "demo" {
		t.Fatalf("params = %v", req["params"])
	}
}

func TestCallReturnsAgentErrors(t *testing.T) {
	fakeAgent(t, func(map[string]any) string {
		return `{"result": "error", "messages": ["User name not found"], "data": null}`
	})
	if _, err := fetchStatus("ghost"); err == nil || err.Error() != "User name not found" {
		t.Fatalf("err = %v", err)
	}
}

func TestNotRunningWithoutSocket(t *testing.T) {
	old := socketPath
	socketPath = filepath.Join(t.TempDir(), "missing.sock")
	defer func() { socketPath = old }()
	if available() {
		t.Fatal("missing socket should not be available")
	}
	if _, err := call([]string{"malware", "user", "list"}, nil); err != errNotRunning {
		t.Fatalf("err = %v", err)
	}
}

func TestFetchScansKeepsOnlyTheUsersHome(t *testing.T) {
	fakeAgent(t, func(map[string]any) string {
		return `{"result": "success", "data": {"items": [
			{"path": "/home", "started": 10},
			{"path": "/home/demo2", "started": 20},
			{"path": "/home/demo", "started": 30, "completed": 31},
			{"path": "/home/demo/docker-data/volumes/demo_html_data/_data", "started": 40.51, "completed": null, "scan_status": "running"}
		]}}`
	})
	scans, err := fetchScans("demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(scans) != 2 || scans[0].Started != 40.51 || scans[1].Started != 30 {
		t.Fatalf("scans = %+v", scans)
	}
}

func TestQueueScanPutsHome(t *testing.T) {
	got := fakeAgent(t, func(map[string]any) string { return `{"result": "success", "data": {"items": null}}` })
	if err := queueScan("demo"); err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal((*got)[0]["params"])
	if string(params) != `{"paths":["/home/demo"]}` {
		t.Fatalf("params = %s", params)
	}
}

func TestScanAllowedReadsPermission(t *testing.T) {
	for reply, want := range map[string]bool{
		`{"result": "success", "data": {"items": {"PERMISSIONS": {"allow_malware_scan": true}}}}`:  true,
		`{"result": "success", "data": {"items": {"PERMISSIONS": {"allow_malware_scan": false}}}}`: false,
		`{"result": "success", "data": {"items": {}}}`:                                             false,
	} {
		got := fakeAgent(t, func(map[string]any) string { return reply })
		allowed, err := scanAllowed()
		if err != nil || allowed != want {
			t.Errorf("%s: allowed = %v, err = %v", reply, allowed, err)
		}
		if cmd, _ := json.Marshal((*got)[0]["command"]); string(cmd) != `["config","show"]` {
			t.Errorf("command = %s", cmd)
		}
	}
}

func TestDisplayPath(t *testing.T) {
	for in, want := range map[string]string{
		"/home/demo/docker-data/volumes/demo_html_data/_data/example.org/x.php": "/var/www/html/example.org/x.php",
		"/home/demo/docker-data/volumes/demo_html_data/_database/x":             "/home/demo/docker-data/volumes/demo_html_data/_database/x",
		"/home/demo/mail/x": "/home/demo/mail/x",
	} {
		if got := displayPath("demo", in); got != want {
			t.Errorf("displayPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFillPage(t *testing.T) {
	running, aborted, boom := "running", "aborted", "scan crashed"
	date := float64(1794297231)
	done := float64(1794297231)
	st := userStatus{Infected: 1, ScanDate: &date, ScanStatus: &running}
	hits := []hit{{File: "/home/demo/docker-data/volumes/demo_html_data/_data/a.php", Type: "SMW-EICAR", Detected: 1794297231, Status: "found"}, {File: "/home/demo/b.php", Status: "cleanup_done"}}
	scans := []scan{
		{Started: 1794297214.4, Completed: &done, TotalMalicious: 1, TotalResources: 23401},
		{Started: 1794297213, Error: &aborted},
		{Started: 1794297212, Error: &boom},
		{Started: 1794297211},
	}
	var data PageData
	fillPage(&data, "demo", st, hits, scans, time.Unix(1794297300, 0))
	if !data.Scanning || data.Infected != 1 || data.LastScan == "" {
		t.Fatalf("summary = %+v", data)
	}
	if data.Hits[0].Path != "/var/www/html/a.php" || !data.Hits[0].Found || data.Hits[1].Found || data.Hits[1].Status != "cleanup done" {
		t.Fatalf("hits = %+v", data.Hits)
	}
	statuses := []string{}
	for _, s := range data.Scans {
		statuses = append(statuses, s.Status)
	}
	if strings.Join(statuses, ",") != "done,aborted,failed,running" || data.Scans[0].Duration != "17s" || data.Scans[2].Message != boom {
		t.Fatalf("scans = %+v", data.Scans)
	}
}

func baseLayout(mgr *i18n.Manager) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "imunifyav": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, pagePath), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: pagePath, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	cases := map[string]struct {
		data PageData
		want []string
		not  []string
	}{
		"not running": {PageData{NotRunning: true}, []string{`id="imunifyav-not-running"`}, []string{`action="/imunifyav/scan"`}},
		"clean":       {PageData{LastScan: "2026-11-10 08:00", GUIReady: true, ScanAllowed: true}, []string{`action="/imunifyav/scan"`, `name="csrf_token" value="test-csrf-token"`, `id="imunifyav-no-hits"`, `href="/imunifyav"`, `id="imunifyav-open-gui"`, `id="imunifyav-scan"`}, []string{"}, 15000)", `id="imunifyav-scan-off"`}},
		"scans off":   {PageData{GUIReady: true}, []string{`id="imunifyav-scan-off"`, `id="imunifyav-open-gui"`}, []string{`id="imunifyav-scan"`}},
		"infected": {PageData{Infected: 1, Scanning: true, ScanAllowed: true, Hits: []HitRow{{Path: "/var/www/html/<x>.php", Signature: "SMW-EICAR", Found: true}}, Scans: []ScanRow{{Started: "2026-11-10 08:00", Status: "done", Found: 1}}},
			[]string{`id="imunifyav-hits"`, "/var/www/html/&lt;x&gt;.php", "SMW-EICAR", `id="imunifyav-scans"`, "}, 15000)", "disabled"}, nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			c.data.LayoutData = baseLayout(mgr)
			w := httptest.NewRecorder()
			if err := page.Render(w, 200, c.data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			for _, s := range c.want {
				if !strings.Contains(body, s) {
					t.Errorf("missing %q", s)
				}
			}
			for _, s := range c.not {
				if strings.Contains(body, s) {
					t.Errorf("unexpected %q", s)
				}
			}
		})
	}
}

func TestDefaultGateway(t *testing.T) {
	route := "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n" +
		"eth0\t000014AC\t00000000\t0001\t0\t0\t0\t00FFFFFF\t0\t0\t0\n" +
		"eth0\t00000000\t010014AC\t0003\t0\t0\t100\t00000000\t0\t0\t0\n"
	if got := defaultGateway(strings.NewReader(route)); got != "172.20.0.1" {
		t.Fatalf("gateway = %q", got)
	}
	if got := defaultGateway(strings.NewReader("Iface\tDestination\n")); got != "" {
		t.Fatalf("no default route should give empty, got %q", got)
	}
}

func TestLoginTokenAsksForTheUsersContext(t *testing.T) {
	got := fakeAgent(t, func(map[string]any) string { return `{"result": "success", "data": {"items": "jwt.token.here"}}` })
	token, err := loginToken("demo")
	if err != nil || token != "jwt.token.here" {
		t.Fatalf("token = %q, err = %v", token, err)
	}
	if cmd, _ := json.Marshal((*got)[0]["command"]); string(cmd) != `["login","get"]` {
		t.Fatalf("command = %s", cmd)
	}
	if params, _ := json.Marshal((*got)[0]["params"]); string(params) != `{"username":"demo"}` {
		t.Fatalf("params = %s", params)
	}
}

func TestGUIProxySendsKeyAndDropsPanelCookies(t *testing.T) {
	var seen *http.Request
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r
		w.Write([]byte("imav ok"))
	}))
	defer upstream.Close()
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(upstream.URL, "http://"))

	oldGW, oldPort, oldKey := gatewayRun, guiPort, guiKeyPath
	defer func() { gatewayRun, guiPort, guiKeyPath = oldGW, oldPort, oldKey }()
	gatewayRun = func() string { return host }
	guiPort = port
	guiKeyPath = filepath.Join(t.TempDir(), "panel.key")
	os.WriteFile(guiKeyPath, []byte("sekret\n"), 0600)

	req := httptest.NewRequest("POST", "/imav/api.php?x=1", strings.NewReader("{}"))
	req.Header.Set("Cookie", "session=panel; OPENPANEL_CSRF=x")
	req.Header.Set("Authorization", "Bearer panel")
	req.Header.Set("X-OpenPanel-Key", "forged")
	w := httptest.NewRecorder()
	newGUIProxy().ServeHTTP(w, req)

	if w.Body.String() != "imav ok" || seen == nil {
		t.Fatalf("proxy response %d %q", w.Code, w.Body.String())
	}
	if seen.URL.Path != "/imav/api.php" || seen.URL.RawQuery != "x=1" || seen.Method != "POST" {
		t.Fatalf("upstream got %s %s?%s", seen.Method, seen.URL.Path, seen.URL.RawQuery)
	}
	if seen.Header.Get("Cookie") != "" || seen.Header.Get("Authorization") != "" || seen.Header.Get("X-OpenPanel-Key") != "sekret" {
		t.Fatalf("upstream headers %v", seen.Header)
	}
	if !guiReady() {
		t.Fatal("gui should be ready with a key and a listening port")
	}
	os.Remove(guiKeyPath)
	if guiReady() {
		t.Fatal("gui without a key should not be ready")
	}
}

func TestRenderGUIPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	if err := guiPage.Render(w, 200, GUIPageData{LayoutData: baseLayout(mgr), Token: `a.b"c`}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="imunifyav-gui"`) || !strings.Contains(body, `href="/imunifyav/summary"`) || strings.Contains(body, `a.b"c`) {
		t.Fatalf("bad gui page")
	}
}
