package backups

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"

	"gist.github.com/stefanpejcic/openpanel/internal/core/safenet"
)

func writeAcctEnv(t *testing.T, tmp, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(tmp, "home", "acct", "backup.env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSSHIdentityFileStaysInOwnVolume(t *testing.T) {
	tmp := envFixture(t)
	ownDir := "/home/acct/docker-data/volumes/acct_html_data/_data/"
	for _, v := range []string{
		"/var/www/html/id_rsa",
		"/home/victim/docker-data/volumes/victim_html_data/_data/id_rsa",
		"/etc/ssh/ssh_host_ed25519_key",
		"id_rsa",
	} {
		writeAcctEnv(t, tmp, "SSH_HOST_NAME=example.com\nSSH_IDENTITY_FILE="+v+"\n")
		config, _ := readBackupEnv("acct")
		// another account's key or a host key used to be read as root here
		if got, want := config["SSH_IDENTITY_FILE"], ownDir+filepath.Base(v); got != want {
			t.Errorf("SSH_IDENTITY_FILE=%q resolved to %q, want %q", v, got, want)
		}
	}
}

func postSettings(t *testing.T, form url.Values) error {
	t.Helper()
	r := httptest.NewRequest("POST", "/backups/settings", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return saveBackupSettings(nil, r, envFilePath("acct"), "acct", "acct")
}

func TestBackupSettingsRejectBadValues(t *testing.T) {
	tmp := envFixture(t)
	writeAcctEnv(t, tmp, "SSH_HOST_NAME=example.com\n")
	for name, form := range map[string]url.Values{
		"foreign key file": {"values[SSH_IDENTITY_FILE]": {"/home/victim/id_rsa"}},
		"climbing key":     {"values[SSH_IDENTITY_FILE]": {"/var/www/html/../../etc/shadow"}},
		"newline in value": {"values[SSH_USER]": {"bob\nWEBDAV_URL=http://127.0.0.1/"}},
	} {
		err := postSettings(t, form)
		if _, bad := err.(badSettingError); !bad {
			t.Errorf("%s: got %v, want a badSettingError", name, err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(tmp, "home", "acct", "backup.env"))
	if strings.Contains(string(data), "victim") || strings.Contains(string(data), "WEBDAV_URL") {
		t.Errorf("rejected values were written: %q", data)
	}
}

func TestAccountBackupEndpointsSkipInternalAddresses(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "INTERNAL")
	}))
	defer internal.Close()

	store := &webdavStore{config: map[string]string{"WEBDAV_URL": internal.URL + "/"}}
	if rc, err := store.Open(context.Background(), "x"); err == nil {
		rc.Close()
		t.Fatal("webdav store opened a loopback url from an account config")
	} else if !errors.Is(err, safenet.ErrBlockedAddress) {
		t.Fatalf("err = %v, want ErrBlockedAddress", err)
	}

	// admin-managed destinations may sit on a private network
	store.config[allowPrivateKey] = "1"
	rc, err := store.Open(context.Background(), "x")
	if err != nil {
		t.Fatalf("admin-managed webdav on loopback: %v", err)
	}
	rc.Close()
}

func TestBackupEnvCantGrantPrivateDial(t *testing.T) {
	tmp := envFixture(t)
	writeAcctEnv(t, tmp, allowPrivateKey+"=1\nWEBDAV_URL=http://127.0.0.1/\n")
	if config, _ := readBackupEnv("acct"); config[allowPrivateKey] != "" {
		t.Fatal("backup.env turned on private-address dialing")
	}
}

func tarGz(t *testing.T, write func(tw *tar.Writer)) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	write(tw)
	_ = tw.Close()
	_ = gz.Close()
	return bytes.NewReader(buf.Bytes())
}

func TestReindexRefusesOversizedFileList(t *testing.T) {
	stream := tarGz(t, func(tw *tar.Writer) {
		// only the header matters, a destination can declare or stream any size here
		_ = tw.WriteHeader(&tar.Header{Name: "backup/00-meta/files.tsv", Typeflag: tar.TypeReg, Size: maxListBytes + 1, Mode: 0o644})
	})
	saved := 0
	_, _, err := classifyArchiveStream("a.tar.gz", stream, true, func(kind string, r io.Reader) error { saved++; return nil })
	if !errors.Is(err, errListTooLarge) || saved != 0 {
		t.Fatalf("err=%v saved=%d, want errListTooLarge before saving", err, saved)
	}
}

func TestSaveListStopsAtCap(t *testing.T) {
	tmp := envFixture(t)
	home := filepath.Join(tmp, "home", "acct")
	err := saveList(home, "a.tar.gz", "files", io.LimitReader(zeroReader{}, maxListBytes+10))
	if !errors.Is(err, errListTooLarge) {
		t.Fatalf("err=%v, want errListTooLarge", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(home, "backup_files")); len(entries) != 0 {
		t.Fatalf("oversized list left files behind: %v", entries)
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestReindexCapsRetainedNames(t *testing.T) {
	old := maxListedNames
	maxListedNames = 50
	t.Cleanup(func() { maxListedNames = old })
	stream := tarGz(t, func(tw *tar.Writer) {
		for i := 0; i < 100; i++ {
			_ = tw.WriteHeader(&tar.Header{Name: fmt.Sprintf("backup/html/f%d", i), Typeflag: tar.TypeReg, Mode: 0o644})
		}
	})
	if _, _, err := classifyArchiveStream("a.tar.gz", stream, true, nil); !errors.Is(err, errListingTooLarge) {
		t.Fatalf("err=%v, want errListingTooLarge", err)
	}
}

func TestRestoredCronsIniGetsEditorPolicy(t *testing.T) {
	plan, tmp := testPlan(t)
	cronsPath := filepath.Join(tmp, "crons.ini")
	plan.sections["backup/crons.ini"] = cronsPath
	for _, bad := range []string{
		"[job-local \"x\"]\nschedule = @every 1m\ncommand = id\n",
		"[job-run \"x\"]\nimage = alpine\ncommand = id\n",
		"[job-exec \"x\"]\nschedule = @daily\ncontainer = php\nnetwork = host\ncommand = id\n",
	} {
		if _, err := plan.extractEntry(strings.NewReader(bad), "backup/crons.ini", 0); err == nil {
			t.Errorf("restore accepted a crons.ini the editor refuses: %q", bad)
		}
	}
	if _, err := os.Stat(cronsPath); err == nil {
		t.Error("a refused crons.ini was written")
	}
	ok := "[job-exec \"backup\"]\nschedule = @daily\ncontainer = php\ncommand = php cron.php\n"
	if dest, err := plan.extractEntry(strings.NewReader(ok), "backup/crons.ini", 0); err != nil || dest == "" {
		t.Errorf("a normal crons.ini was refused: %v", err)
	}
}

func TestQueuedRestoreRunsAsQueueOwner(t *testing.T) {
	var got restoreRequest
	runRestore = func(ctx context.Context, a *appctx.App, req restoreRequest) (int, map[string]any) {
		got = req
		return http.StatusOK, map[string]any{"message": "ok"}
	}
	defer func() { runRestore = performRestore }()

	q := testQueue(t)
	// a job saved in the account's own home naming another account
	forged := `[{"id":"x","request":{"user_id":2,"username":"victim","user_context":"victim","form":{"backup_file":["a.tar.gz"],"restore_target":["database"]}},"status":"queued","queued":"2026-10-15T10:00:00Z"}]`
	if err := os.WriteFile(q.path, []byte(forged), 0o600); err != nil {
		t.Fatal(err)
	}
	q.snapshot(nil, restoreRequest{UserID: 7, Username: "attacker", UserContext: "attacker"})
	waitIdle(t, q)
	if got.UserID != 7 || got.UserContext != "attacker" || got.Username != "attacker" {
		t.Fatalf("restore ran as %+v, want the queue owner", got)
	}
}
