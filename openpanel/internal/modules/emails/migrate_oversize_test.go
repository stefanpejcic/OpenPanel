package emails

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestDecodeMUTF7(t *testing.T) {
	for in, want := range map[string]string{"INBOX": "INBOX", "Projects &APw-": "Projects ü", "Entw&APw-rfe": "Entwürfe", "A &- B": "A & B", "&ZeVnLIqe-": "日本語"} {
		if got, ok := decodeMUTF7(in); !ok || got != want {
			t.Errorf("decodeMUTF7(%q) = %q, %v, want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"&APw", "&A-"} {
		if _, ok := decodeMUTF7(in); ok {
			t.Errorf("decodeMUTF7(%q) should fail", in)
		}
	}
}

// fakeMailserver stands in for podman exec into the mailserver container
type fakeMailserver struct {
	calls  [][]string
	saved  map[string][]string
	exists map[string]bool
}

func (f *fakeMailserver) run(_ context.Context, stdin io.Reader, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	switch {
	case args[0] == "doveconf" && args[2] == "quota_mail_size":
		return []byte("10M\n"), nil
	case args[0] == "doveconf":
		return nil, errors.New("unknown setting")
	case args[1] == "mailbox":
		return []byte("INBOX\nAAA-big\nProjects ü\n"), nil
	case args[1] == "search":
		if f.exists[args[len(args)-1]] {
			return []byte("a1b2 3\n"), nil
		}
		return nil, nil
	case args[1] == "-o":
		data, _ := io.ReadAll(stdin)
		folder := args[len(args)-1]
		f.saved[folder] = append(f.saved[folder], string(data))
		return nil, nil
	}
	return nil, errors.New("unexpected call")
}

func withFakeMailserver(t *testing.T) *fakeMailserver {
	t.Helper()
	fake := &fakeMailserver{saved: map[string][]string{}, exists: map[string]bool{"<dup@example.org>": true}}
	orig := runMailserver
	runMailserver = fake.run
	t.Cleanup(func() { runMailserver = orig })
	return fake
}

// oversizeJob writes a finished job whose fetch script left a manifest and downloads
func oversizeJob(t *testing.T, home, id string, manifest string, files map[string]string) {
	t.Helper()
	root, uid, gid, err := openJobDir("ctx1")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	_ = writeJob(root, MigrateJob{ID: id, Destination: "new@demo.rs", MaxSize: 10 << 20, ImportOversize: true, OversizeCap: 4096}, uid, gid)
	dir := filepath.Join(home, "ctx1", "imapsync")
	log := "Folder     1/2 [AAA-big]                -> [AAA-big]\nFolder     2/2 [Projects &APw-]         -> [Projects &APw-]\n" +
		"msg AAA-big/1 skipped (2000 exceeds maxsize limit 10485761 bytes)\nExiting with return value 0 (EX_OK: successful termination) 0/50 nb_errors/max_errors\n"
	_ = os.WriteFile(filepath.Join(dir, id+".log"), []byte(log), 0o600)
	_ = os.Mkdir(filepath.Join(dir, oversizeDir(id)), 0o700)
	_ = os.WriteFile(filepath.Join(dir, oversizeDir(id), "manifest.jsonl"), []byte(manifest), 0o600)
	for name, body := range files {
		_ = os.WriteFile(filepath.Join(dir, oversizeDir(id), name), []byte(body), 0o600)
	}
}

func TestImportOversize(t *testing.T) {
	home, _ := withMigrateHome(t)
	fake := withFakeMailserver(t)
	id := "20261010-100000"
	msg := func(mid string) string { return "Message-ID: " + mid + "\r\nSubject: big\r\n\r\nbody\r\n" }
	manifest := `{"folder":"AAA-big","uid":1,"size":2000,"file":"1.eml"}
{"folder":"Projects &APw-","uid":2,"size":2000,"file":"2.eml"}
{"folder":"AAA-big","uid":3,"size":2000,"file":"3.eml"}
{"folder":"AAA-big","uid":4,"size":5000,"skip":"cap"}
{"folder":"AAA-big","uid":5,"size":2000,"error":"download failed"}
{"folder":"Gone","uid":6,"size":2000,"file":"6.eml"}
{"folder":"AAA-big","uid":7,"size":2000,"file":"../../etc/passwd"}
{"folder":"AAA-big","uid":8,"size":2000,"file":"8.eml"}
`
	oversizeJob(t, home, id, manifest, map[string]string{
		"1.eml": msg("<one@example.org>"), "2.eml": msg("<two@example.org>"), "3.eml": msg("<dup@example.org>"),
		"6.eml": msg("<six@example.org>"), "8.eml": msg("<eight@example.org>") + strings.Repeat("x", 5000),
	})

	imp := &oversizeImport{id: id}
	importOversize(context.Background(), "ctx1", imp, 4096, func(dest string) bool { return dest == "new@demo.rs" })

	root, _, _, _ := openJobDir("ctx1")
	defer root.Close()
	job, err := readJob(root, id)
	if err != nil {
		t.Fatal(err)
	}
	// 1 and 2 saved, 3 was already there, 4 over the cap, 5 not downloaded, 6 no such folder, 7 bad name, 8 bigger than it claimed
	if !job.BigDone || job.BigCopied != 3 || job.BigOverCap != 1 || job.BigFailed != 4 {
		t.Errorf("counts = copied %d, over cap %d, failed %d, done %v", job.BigCopied, job.BigOverCap, job.BigFailed, job.BigDone)
	}
	if len(fake.saved["AAA-big"]) != 1 || !strings.Contains(fake.saved["AAA-big"][0], "<one@example.org>") || len(fake.saved["Projects ü"]) != 1 {
		t.Errorf("saved = %v", fake.saved)
	}
	for _, c := range fake.calls {
		if len(c) > 3 && c[3] == "save" && (c[2] != "quota_mail_size=0" || c[5] != "new@demo.rs") {
			t.Errorf("save call = %v", c)
		}
	}
	if _, err := os.Stat(filepath.Join(home, "ctx1", "imapsync", oversizeDir(id))); !os.IsNotExist(err) {
		t.Error("the download folder should be removed")
	}
	text, _ := readJobLog("ctx1", id)
	if !strings.Contains(text, "3 copied, 1 over 0 MB, 4 failed") || !strings.Contains(text, `folder "Gone" does not exist`) {
		t.Errorf("log = %q", text)
	}
	if st, _ := migrationStatus(context.Background(), "ctx1"); st.Jobs[0].Status != "completed" {
		t.Errorf("job after import = %+v", st.Jobs[0])
	}
}

func TestImportOversizeChecksDestinationAgain(t *testing.T) {
	home, _ := withMigrateHome(t)
	fake := withFakeMailserver(t)
	id := "20261011-100000"
	oversizeJob(t, home, id, `{"folder":"AAA-big","uid":1,"size":2000,"file":"1.eml"}`+"\n", map[string]string{"1.eml": "Subject: x\r\n\r\nx\r\n"})

	importOversize(context.Background(), "ctx1", &oversizeImport{id: id}, 4096, func(string) bool { return false })

	root, _, _, _ := openJobDir("ctx1")
	defer root.Close()
	job, _ := readJob(root, id)
	if job.BigFailed != 1 || len(fake.saved) != 0 {
		t.Errorf("a mailbox the account doesn't own got an import: %+v %v", job, fake.saved)
	}
}

func TestStatusAndCancelWhileImporting(t *testing.T) {
	home, _ := withMigrateHome(t)
	id := "20261012-100000"
	oversizeJob(t, home, id, "", nil)
	imp := &oversizeImport{id: id}
	oversizeImports.Store("ctx1", imp)
	t.Cleanup(func() { oversizeImports.Delete("ctx1") })

	st, err := migrationStatus(context.Background(), "ctx1")
	if err != nil {
		t.Fatal(err)
	}
	if !st.Running || st.Jobs[0].Status != "running" {
		t.Errorf("importing job status = %+v", st)
	}
	if _, err := os.Stat(filepath.Join(home, "ctx1", "imapsync", oversizeDir(id))); err != nil {
		t.Error("the downloads of the job being imported must stay")
	}
	if err := cancelMigration(context.Background(), "ctx1"); err != nil || !imp.cancelled.Load() {
		t.Errorf("cancel during the import = %v, cancelled %v", err, imp.cancelled.Load())
	}
}

func TestStartRefusesOversizeWhenOff(t *testing.T) {
	withMigrateHome(t)
	req := validMigrateRequest()
	req.ImportOversize = true
	_, err := startMigration(context.Background(), &appctx.App{Config: config.Config{}}, 1, "user1", "ctx1", req)
	if status, msg := errStatus(err); status != 400 || !strings.Contains(msg, "not turned on") {
		t.Errorf("start with the import off = %d %q", status, msg)
	}
	if n := oversizeCap(&appctx.App{Config: config.Config{"imapsync_oversize_mb": "50"}}); n != 50<<20 {
		t.Errorf("oversizeCap = %d", n)
	}
}

func TestRenderMigratePageOversizeOption(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	for mb, want := range map[int64]bool{0: false, 50: true} {
		data := MigratePageData{LayoutData: baseLayout(mgr, "/emails/migrate"), Accounts: []string{"info@demo.rs"}, OversizeMB: mb}
		w := httptest.NewRecorder()
		if err := migratePage.Render(w, 200, data); err != nil {
			t.Fatal(err)
		}
		body := w.Body.String()
		if got := strings.Contains(body, `name="import_oversize"`); got != want {
			t.Errorf("OversizeMB %d: checkbox shown = %v", mb, got)
		}
		if want && !strings.Contains(body, "up to 50 MB") {
			t.Error("checkbox label should name the cap")
		}
	}
}
