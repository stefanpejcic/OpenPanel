package emails

import (
	"context"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func validMigrateRequest() MigrateRequest {
	return MigrateRequest{
		SourceHost: " imap.example.com ", SourceUser: "old@example.com", SourcePassword: "s3cret-src",
		Destination: "New@Demo.rs", DestPassword: "s3cret-dst",
	}
}

func TestValidateMigrateRequest(t *testing.T) {
	req := validMigrateRequest()
	if err := validateMigrateRequest(&req); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if req.SourceHost != "imap.example.com" || req.SourcePort != 993 || req.SourceSecurity != "ssl" || req.Destination != "new@demo.rs" {
		t.Errorf("defaults not applied: %+v", req)
	}

	starttls := validMigrateRequest()
	starttls.SourceSecurity = "STARTTLS"
	_ = validateMigrateRequest(&starttls)
	if starttls.SourcePort != 143 {
		t.Errorf("starttls default port = %d, want 143", starttls.SourcePort)
	}

	bad := map[string]func(*MigrateRequest){
		"empty host":      func(r *MigrateRequest) { r.SourceHost = "" },
		"host with space": func(r *MigrateRequest) { r.SourceHost = "imap example.com" },
		"host injection":  func(r *MigrateRequest) { r.SourceHost = "a.com;rm -rf" },
		"bad security":    func(r *MigrateRequest) { r.SourceSecurity = "tls13" },
		"port too high":   func(r *MigrateRequest) { r.SourcePort = 70000 },
		"empty user":      func(r *MigrateRequest) { r.SourceUser = " " },
		"newline in pass": func(r *MigrateRequest) { r.SourcePassword = "a\nb" },
		"empty dest pass": func(r *MigrateRequest) { r.DestPassword = "" },
		"bad destination": func(r *MigrateRequest) { r.Destination = "nope" },
		"newline in user": func(r *MigrateRequest) { r.SourceUser = "a\n--host2=x" },
	}
	for name, mutate := range bad {
		r := validMigrateRequest()
		mutate(&r)
		if err := validateMigrateRequest(&r); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestIsPublicIP(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.0.0.5", "192.168.1.1", "172.17.0.1", "169.254.169.254", "100.64.1.1", "0.0.0.0", "::1", "fd00::1", "fe80::1", "::ffff:127.0.0.1", "64:ff9b::a00:1"} {
		if isPublicIP(net.ParseIP(ip)) {
			t.Errorf("%s should not be public", ip)
		}
	}
	for _, ip := range []string{"1.1.1.1", "185.199.108.153", "2606:4700::1111"} {
		if !isPublicIP(net.ParseIP(ip)) {
			t.Errorf("%s should be public", ip)
		}
	}
}

func TestResolveSource(t *testing.T) {
	orig := lookupIP
	t.Cleanup(func() { lookupIP = orig })
	lookupIP = func(_ context.Context, host string) ([]net.IP, error) {
		switch host {
		case "mixed.example":
			return []net.IP{net.ParseIP("2606:4700::1111"), net.ParseIP("1.1.1.1")}, nil
		case "rebind.example":
			return []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("10.0.0.1")}, nil
		}
		return nil, &net.DNSError{Err: "no such host", Name: host}
	}

	if ip, err := resolveSource(context.Background(), "mixed.example", false); err != nil || ip != "1.1.1.1" {
		t.Errorf("mixed: got %q, %v, want IPv4", ip, err)
	}
	if _, err := resolveSource(context.Background(), "rebind.example", false); err == nil {
		t.Error("a private address in the answer should be refused")
	}
	if ip, err := resolveSource(context.Background(), "rebind.example", true); err != nil || ip == "" {
		t.Errorf("allowPrivate: got %q, %v", ip, err)
	}
	if _, err := resolveSource(context.Background(), "127.0.0.1", false); err == nil {
		t.Error("loopback literal should be refused")
	}
	if _, err := resolveSource(context.Background(), "missing.example", false); err == nil {
		t.Error("unresolvable host should error")
	}
}

func TestImapsyncArgs(t *testing.T) {
	a := &appctx.App{Config: config.Config{"imapsync_host": "mail.local", "imapsync_port": "9993", "imapsync_security": "ssl"}}
	destHost, destPort, destSec := migrateDest(context.Background(), a, "user1")
	req := validMigrateRequest()
	req.SourceSecurity = "starttls"
	req.SourceUser = "-weird"
	_ = validateMigrateRequest(&req)
	args := strings.Join(imapsyncArgs(req, "1.1.1.1", destHost, destPort, destSec, "job1"), " ")
	for _, want := range []string{"--host1=1.1.1.1", "--port1=143", "--tls1", "--user1=-weird", "--passfile1=/data/job1.p1", "--host2=mail.local", "--port2=9993", "--ssl2", "--user2=new@demo.rs", "--passfile2=/data/job1.p2"} {
		if !strings.Contains(args, want) {
			t.Errorf("args missing %q: %s", want, args)
		}
	}
	if strings.Contains(args, "s3cret") {
		t.Errorf("passwords leaked into args: %s", args)
	}
}

// fakePodman records calls and pretends the imapsync container runs until stopped
type fakePodman struct {
	calls   [][]string
	running string
}

func (f *fakePodman) run(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	switch args[0] {
	case "container":
		if f.running == "" {
			return []byte("Error: no such container"), os.ErrNotExist
		}
		return []byte("true|" + f.running + "\n"), nil
	case "run":
		for i, a := range args {
			if a == "--label" {
				f.running = strings.TrimPrefix(args[i+1], migrateJobLabel+"=")
			}
		}
		return []byte("abc123\n"), nil
	case "stop":
		f.running = ""
		return nil, nil
	}
	return nil, nil
}

func withMigrateHome(t *testing.T) (string, *fakePodman) {
	t.Helper()
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "ctx1"), 0o755); err != nil {
		t.Fatal(err)
	}
	origHome, origPodman := migrateHomeRoot, runPodman
	fake := &fakePodman{}
	migrateHomeRoot, runPodman = home, fake.run
	t.Cleanup(func() { migrateHomeRoot, runPodman = origHome, origPodman })
	return home, fake
}

func TestOpenJobDirRefusesSymlink(t *testing.T) {
	home, _ := withMigrateHome(t)
	if err := os.Symlink("/etc", filepath.Join(home, "ctx1", "imapsync")); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := openJobDir("ctx1"); err == nil {
		t.Error("a symlinked job dir should be refused")
	}
	if _, _, _, err := openJobDir("../ctx1"); err == nil {
		t.Error("a context with a slash should be refused")
	}
}

func TestMigrationLifecycle(t *testing.T) {
	home, fake := withMigrateHome(t)
	dir := filepath.Join(home, "ctx1", "imapsync")
	ctx := context.Background()

	root, uid, gid, err := openJobDir("ctx1")
	if err != nil {
		t.Fatalf("openJobDir: %v", err)
	}
	defer root.Close()

	// a finished job, a failed one and a cancelled one
	for _, j := range []struct {
		job MigrateJob
		log string
	}{
		{MigrateJob{ID: "20261001-100000", SourceHost: "a", SourceUser: "u", Destination: "d@x.rs"}, "Host1: success login\nExiting with return value 0 (EX_OK: successful termination) 0/50 nb_errors/max_errors\n"},
		{MigrateJob{ID: "20261002-100000"}, "Host1 failure: Error login\nExiting with return value 161 (EXIT_AUTHENTICATION_FAILURE_USER1) 1/50 nb_errors/max_errors\n"},
		{MigrateJob{ID: "20261003-100000", Cancelled: true}, "copying...\n"},
		{MigrateJob{ID: "20261004-100000"}, "copying...\n"},
	} {
		if err := writeJob(root, j.job, uid, gid); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, j.job.ID+".log"), []byte(j.log), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(dir, "20261001-100000.p1"), []byte("left over"), 0o600)

	st, err := migrationStatus(ctx, "ctx1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	want := map[string]string{"20261001-100000": "completed", "20261002-100000": "failed", "20261003-100000": "cancelled", "20261004-100000": "failed"}
	if len(st.Jobs) != 4 || st.Jobs[0].ID != "20261004-100000" {
		t.Fatalf("jobs not newest first: %+v", st.Jobs)
	}
	for _, j := range st.Jobs {
		if j.Status != want[j.ID] {
			t.Errorf("%s status = %s, want %s (%s)", j.ID, j.Status, want[j.ID], j.Message)
		}
	}
	if st.Jobs[3].Message != "" {
		t.Errorf("completed job should have no message, got %q", st.Jobs[3].Message)
	}
	if st.Jobs[2].ExitCode != 161 || st.Jobs[2].Message != "EXIT_AUTHENTICATION_FAILURE_USER1" {
		t.Errorf("failed job exit info = %d %q", st.Jobs[2].ExitCode, st.Jobs[2].Message)
	}
	if _, err := os.Stat(filepath.Join(dir, "20261001-100000.p1")); !os.IsNotExist(err) {
		t.Error("leftover passfile of a finished job should be removed")
	}

	// a running job shows as running and can be cancelled
	fake.running = "20261004-100000"
	st, _ = migrationStatus(ctx, "ctx1")
	if !st.Running || st.Jobs[0].Status != "running" || st.Log != "copying...\n" {
		t.Errorf("running status = %+v", st)
	}
	if err := cancelMigration(ctx, "ctx1"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	st, _ = migrationStatus(ctx, "ctx1")
	if st.Running || st.Jobs[0].Status != "cancelled" {
		t.Errorf("after cancel = %+v", st.Jobs[0])
	}
	if err := cancelMigration(ctx, "ctx1"); err == nil {
		t.Error("cancel with nothing running should error")
	}

	if _, err := readJobLog("ctx1", "../../etc/passwd"); err == nil {
		t.Error("a bad log id should be refused")
	}
	if text, err := readJobLog("ctx1", "20261002-100000"); err != nil || !strings.Contains(text, "Host1 failure") {
		t.Errorf("readJobLog = %q, %v", text, err)
	}
}

func TestPruneJobsKeepsNewest(t *testing.T) {
	home, _ := withMigrateHome(t)
	root, uid, gid, err := openJobDir("ctx1")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for i := 0; i < migrateKeepJobs+3; i++ {
		_ = writeJob(root, MigrateJob{ID: "20261001-1000" + twoDigits(i)}, uid, gid)
	}
	pruneJobs(root, "")
	entries, _ := os.ReadDir(filepath.Join(home, "ctx1", "imapsync"))
	if len(entries) != migrateKeepJobs {
		t.Errorf("kept %d jobs, want %d", len(entries), migrateKeepJobs)
	}
	if _, err := root.Stat("20261001-100000.json"); err == nil {
		t.Error("oldest job should be pruned")
	}
}

func twoDigits(i int) string {
	return string([]byte{byte('0' + i/10), byte('0' + i%10)})
}

func TestMigrateStartWritesPassfilesNotArgs(t *testing.T) {
	_, fake := withMigrateHome(t)
	root, uid, gid, _ := openJobDir("ctx1")
	defer root.Close()

	// what startMigration does once the request passed the checks that need the DB
	req := validMigrateRequest()
	_ = validateMigrateRequest(&req)
	id := newJobID(root)
	_ = writeOwned(root, id+".p1", []byte(req.SourcePassword+"\n"), uid, gid)
	data, err := root.ReadFile(id + ".p1")
	if err != nil || string(data) != "s3cret-src\n" {
		t.Fatalf("passfile = %q, %v", data, err)
	}
	info, _ := root.Stat(id + ".p1")
	if info.Mode().Perm() != 0o600 {
		t.Errorf("passfile mode = %v", info.Mode().Perm())
	}
	args := imapsyncArgs(req, "1.1.1.1", "185.7.32.112", "993", "ssl", id)
	_, _ = fake.run(context.Background(), "ctx1", append([]string{"run", "--label", migrateJobLabel + "=" + id}, args...)...)
	for _, arg := range fake.calls[0] {
		if strings.Contains(arg, "s3cret") {
			t.Errorf("password in podman args: %v", fake.calls[0])
		}
	}
}

func TestRenderMigratePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := MigratePageData{
		LayoutData: baseLayout(mgr, "/emails/migrate"),
		Accounts:   []string{"info@demo.rs"},
		Status:     MigrateStatus{Jobs: []MigrateJob{{ID: "20261001-100000", SourceHost: "imap.old.com", Status: "completed"}}},
	}
	w := httptest.NewRecorder()
	if err := migratePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"info@demo.rs", "source_host", "destination_password", "imap.old.com", "Test connection"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}

	empty := MigratePageData{LayoutData: baseLayout(mgr, "/emails/migrate")}
	w = httptest.NewRecorder()
	if err := migratePage.Render(w, 200, empty); err != nil {
		t.Fatalf("Render empty: %v", err)
	}
	if !strings.Contains(w.Body.String(), "No email accounts yet") {
		t.Error("expected empty state")
	}
}

func TestMigrateDestFallsBackTo143(t *testing.T) {
	orig := dialIMAP
	t.Cleanup(func() { dialIMAP = orig })
	a := &appctx.App{Config: config.Config{"imapsync_host": "203.0.113.5"}}

	dialIMAP = func(string) bool { return true }
	if _, port, sec := migrateDest(context.Background(), a, "u"); port != "993" || sec != "ssl" {
		t.Errorf("with 993 open got %s/%s", port, sec)
	}
	dialIMAP = func(string) bool { return false }
	host, port, sec := migrateDest(context.Background(), a, "u")
	if host != "203.0.113.5" || port != "143" || sec != "auto" {
		t.Errorf("without 993 got %s:%s/%s", host, port, sec)
	}
	args := strings.Join(imapsyncArgs(validMigrateRequest(), "1.1.1.1", host, port, sec, "j"), " ")
	if strings.Contains(args, "ssl2") || strings.Contains(args, "tls2") {
		t.Errorf("auto should leave host2 security to imapsync: %s", args)
	}
}

func TestLoginLinesKeepsOnlyResults(t *testing.T) {
	out := "Host1: will try to use LOGIN authentication on host1\nHost1 failure: Error login on [x] with user [u] auth [LOGIN]: 2 NO [AUTHENTICATIONFAILED] Authentication failed.\nHost2: success login on [x] with user [d] auth [LOGIN]\nErr 1/1: Host1 failure: Error login on [x]\n"
	got := loginLines(out)
	if strings.Count(got, "\n") != 1 || !strings.Contains(got, "AUTHENTICATIONFAILED") || !strings.Contains(got, "Host2: success login") {
		t.Errorf("loginLines = %q", got)
	}
}

func TestParseDovecotSize(t *testing.T) {
	for in, want := range map[string]int64{"10M": 10 << 20, "10240000\n": 10240000, "0": 0, "unlimited": 0, "2 G": 2 << 30, "512kB": 512 << 10} {
		if got, ok := parseDovecotSize(in); !ok || got != want {
			t.Errorf("parseDovecotSize(%q) = %d, %v, want %d", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "abc", "-5", "10X", "1x0M"} {
		if _, ok := parseDovecotSize(in); ok {
			t.Errorf("parseDovecotSize(%q) should fail", in)
		}
	}
}

func TestStatusCountsTooBig(t *testing.T) {
	home, _ := withMigrateHome(t)
	dir := filepath.Join(home, "ctx1", "imapsync")
	root, uid, gid, err := openJobDir("ctx1")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	skip := "msg INBOX/8 skipped (11479514 exceeds maxsize limit 10485761 bytes)\n"
	for _, j := range []struct {
		job MigrateJob
		log string
	}{
		{MigrateJob{ID: "20261001-100000", MaxSize: 10 << 20}, "copied\n" + skip + skip + "Exiting with return value 0 (EX_OK: successful termination) 0/50 nb_errors/max_errors\n"},
		// killed before imapsync wrote its summary, so the count isn't final yet
		{MigrateJob{ID: "20261002-100000", MaxSize: 10 << 20}, skip},
		{MigrateJob{ID: "20261003-100000"}, skip + "Exiting with return value 0 (EX_OK: successful termination) 0/50 nb_errors/max_errors\n"},
	} {
		_ = writeJob(root, j.job, uid, gid)
		_ = os.WriteFile(filepath.Join(dir, j.job.ID+".log"), []byte(j.log), 0o600)
	}

	st, err := migrationStatus(context.Background(), "ctx1")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, j := range st.Jobs {
		got[j.ID] = j.TooBig
	}
	if got["20261001-100000"] != 2 || got["20261002-100000"] != 0 || got["20261003-100000"] != 0 {
		t.Errorf("too_big = %v", got)
	}
	if st.Jobs[2].Status != "completed" || st.Jobs[2].Message != "" {
		t.Errorf("skipped messages shouldn't fail the job: %+v", st.Jobs[2])
	}
	saved, _ := readJob(root, "20261001-100000")
	if !saved.SizeChecked || saved.TooBig != 2 || saved.Status != "" {
		t.Errorf("count not saved in the job file: %+v", saved)
	}
	if unfinished, _ := readJob(root, "20261002-100000"); unfinished.SizeChecked {
		t.Error("an unfinished job shouldn't be counted yet")
	}
}
