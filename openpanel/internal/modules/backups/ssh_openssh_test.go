package backups

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fakeSSH puts a shell script in place of ssh, it gets the same argv the real one would
func fakeSSH(t *testing.T, script string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ssh")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	old := opensshBinary
	opensshBinary = bin
	t.Cleanup(func() { opensshBinary = old })
}

func opensshTestConfig(t *testing.T) map[string]string {
	t.Helper()
	key := filepath.Join(t.TempDir(), "id")
	if err := os.WriteFile(key, []byte("KEY"), 0o644); err != nil {
		t.Fatal(err)
	}
	return map[string]string{allowPrivateKey: "1", "SSH_IDENTITY_FILE": key, "SSH_HOST_NAME": "backup.example.com", "SSH_USER": "u", "SSH_REMOTE_PATH": "/b"}
}

func gzipBytes(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	_, _ = zw.Write([]byte(s))
	_ = zw.Close()
	return b.Bytes()
}

func TestOpensshAllowed(t *testing.T) {
	fakeSSH(t, "exit 0\n")
	base := opensshTestConfig(t)
	if !opensshAllowed(base) {
		t.Fatal("admin-managed key login should use openssh")
	}
	for name, change := range map[string]func(map[string]string){
		"account managed": func(c map[string]string) { delete(c, allowPrivateKey) },
		"password":        func(c map[string]string) { c["SSH_PASSWORD"] = "x" },
		"passphrase":      func(c map[string]string) { c["SSH_IDENTITY_PASSPHRASE"] = "x" },
		"no key":          func(c map[string]string) { delete(c, "SSH_IDENTITY_FILE") },
	} {
		c := map[string]string{}
		for k, v := range base {
			c[k] = v
		}
		change(c)
		if opensshAllowed(c) {
			t.Errorf("%s: openssh allowed", name)
		}
	}
	opensshBinary = filepath.Join(t.TempDir(), "missing")
	if opensshAllowed(base) {
		t.Error("openssh allowed without the binary")
	}
}

func TestKnownHostsLineAndArgs(t *testing.T) {
	c := map[string]string{"SSH_HOST_NAME": "h.example.com", "SSH_HOST_PUBLIC_KEY": "ssh-ed25519 AAAA", "SSH_USER": "u"}
	if got := knownHostsLine(c); got != "h.example.com ssh-ed25519 AAAA\n" {
		t.Errorf("port 22 line = %q", got)
	}
	c["SSH_PORT"] = "23"
	if got := knownHostsLine(c); got != "[h.example.com]:23 ssh-ed25519 AAAA\n" {
		t.Errorf("port 23 line = %q", got)
	}
	args := opensshArgs(c, "/k", "/kh", "cat -- '/x'")
	if !slices.Contains(args, "StrictHostKeyChecking=yes") || !slices.Contains(args, "UserKnownHostsFile=/kh") {
		t.Errorf("pinned key not enforced: %v", args)
	}
	if i := slices.Index(args, "--"); i < 0 || args[i+1] != "h.example.com" || args[i+2] != "cat -- '/x'" {
		t.Errorf("destination not after --: %v", args)
	}
	if args := opensshArgs(c, "/k", "", "cat"); !slices.Contains(args, "StrictHostKeyChecking=no") {
		t.Errorf("no pinned key should skip the check like the Go client: %v", args)
	}
}

func TestOpenWithOpenSSHStreamsAndCleansUp(t *testing.T) {
	c := opensshTestConfig(t)
	archive := filepath.Join(t.TempDir(), "a.tar.gz")
	want := gzipBytes(t, strings.Repeat("dump ", 1000))
	if err := os.WriteFile(archive, want, 0o644); err != nil {
		t.Fatal(err)
	}
	// prints the key mode so the 0600 copy is checked too
	fakeSSH(t, `for a; do [ "$prev" = "-i" ] && keyfile=$a; prev=$a; done
[ "$(stat -c %a "$keyfile")" = 600 ] || { echo "key mode" >&2; exit 9; }
cat "`+archive+`"
`)
	rc, err := openSFTPOrSSH(context.Background(), c, "/b/a.tar.gz", func() (io.ReadCloser, error) {
		t.Fatal("fell back to sftp")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rc)
	rc.Close()
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("read %d bytes, err %v", len(got), err)
	}
	if dir := rc.(*opensshFile).tmpDir; dirExists(dir) {
		t.Errorf("temp dir %s left behind", dir)
	}
}

func TestOpenWithOpenSSHFallsBack(t *testing.T) {
	c := opensshTestConfig(t)
	for name, script := range map[string]string{
		"command rejected": "echo 'This service allows sftp connections only.'; exit 1\n",
		"auth failed":      "echo 'Permission denied (publickey).' >&2; exit 255\n",
	} {
		fakeSSH(t, script)
		fellBack := false
		rc, err := openSFTPOrSSH(context.Background(), c, "/b/a.tar.gz", func() (io.ReadCloser, error) {
			fellBack = true
			return io.NopCloser(strings.NewReader("sftp")), nil
		})
		if err != nil || !fellBack {
			t.Errorf("%s: fellBack=%v err=%v", name, fellBack, err)
			continue
		}
		rc.Close()
	}
}

func TestOpenWithOpenSSHCutOffIsAnError(t *testing.T) {
	c := opensshTestConfig(t)
	fakeSSH(t, "printf '\\037\\213partial'; echo 'Connection reset' >&2; exit 255\n")
	rc, err := openSFTPOrSSH(context.Background(), c, "/b/a.tar.gz", func() (io.ReadCloser, error) {
		t.Fatal("fell back after data arrived")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if _, err := io.ReadAll(rc); err == nil || !strings.Contains(err.Error(), "Connection reset") {
		t.Fatalf("cut off download read without error: %v", err)
	}
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
