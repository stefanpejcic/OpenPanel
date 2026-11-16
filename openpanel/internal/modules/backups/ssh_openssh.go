package backups

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// opensshBinary is swapped in tests, x/crypto/ssh tops out around 5 MB/s on links where OpenSSH does 15
var opensshBinary = "ssh"

// opensshAllowed is true only for admin-managed key logins, accounts that set their own destination stay on the Go client so safenet keeps blocking private addresses
func opensshAllowed(config map[string]string) bool {
	if config[allowPrivateKey] != "1" || config["SSH_IDENTITY_FILE"] == "" {
		return false
	}
	if config["SSH_PASSWORD"] != "" || config["SSH_IDENTITY_PASSPHRASE"] != "" {
		return false
	}
	_, err := exec.LookPath(opensshBinary)
	return err == nil
}

// knownHostsLine turns the pinned SSH_HOST_PUBLIC_KEY into a known_hosts entry, OpenSSH wants [host]:port for anything but 22
func knownHostsLine(config map[string]string) string {
	key := strings.TrimSpace(config["SSH_HOST_PUBLIC_KEY"])
	if key == "" {
		return ""
	}
	host := config["SSH_HOST_NAME"]
	if port := config["SSH_PORT"]; port != "" && port != "22" {
		host = "[" + host + "]:" + port
	}
	return host + " " + key + "\n"
}

// opensshArgs builds the ssh command line, -F /dev/null so nothing from a config file sneaks in
func opensshArgs(config map[string]string, keyPath, knownHosts, remoteCmd string) []string {
	port := config["SSH_PORT"]
	if port == "" {
		port = "22"
	}
	args := []string{
		"-F", "/dev/null",
		"-o", "BatchMode=yes",
		"-o", "IdentitiesOnly=yes",
		"-o", "PasswordAuthentication=no",
		"-o", "ConnectTimeout=15",
		"-o", "ServerAliveInterval=15",
		"-o", "ServerAliveCountMax=4",
		"-o", "LogLevel=ERROR",
		"-o", "GlobalKnownHostsFile=/dev/null",
	}
	if knownHosts != "" {
		args = append(args, "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+knownHosts)
	} else {
		// same as the Go client, no pinned key means the host isn't checked
		args = append(args, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null")
	}
	args = append(args, "-i", keyPath, "-p", port, "-l", config["SSH_USER"], "--", config["SSH_HOST_NAME"], remoteCmd)
	return args
}

// opensshFile is the stdout of `ssh host cat file`, a non-zero exit turns the final EOF into an error so a cut off download never looks complete
type opensshFile struct {
	r       *bufio.Reader
	cmd     *exec.Cmd
	stderr  *bytes.Buffer
	tmpDir  string
	once    sync.Once
	waitErr error
}

func (f *opensshFile) wait() error {
	f.once.Do(func() {
		if err := f.cmd.Wait(); err != nil {
			if msg := strings.TrimSpace(f.stderr.String()); msg != "" {
				err = fmt.Errorf("ssh: %s", lastLine(msg))
			}
			f.waitErr = err
		}
		os.RemoveAll(f.tmpDir)
	})
	return f.waitErr
}

func (f *opensshFile) Read(p []byte) (int, error) {
	n, err := f.r.Read(p)
	if errors.Is(err, io.EOF) {
		if werr := f.wait(); werr != nil {
			return n, werr
		}
	}
	return n, err
}

func (f *opensshFile) Close() error {
	if f.cmd.ProcessState == nil && f.cmd.Process != nil {
		_ = f.cmd.Process.Kill()
	}
	_ = f.wait()
	return nil
}

func lastLine(s string) string {
	lines := strings.Split(s, "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// openWithOpenSSH streams the file through `ssh host cat`, any error before the first bytes arrive means the caller falls back to sftp
func openWithOpenSSH(ctx context.Context, config map[string]string, remoteFile string) (io.ReadCloser, error) {
	keyBytes, err := tenantfs.ReadFile(config["SSH_IDENTITY_FILE"])
	if err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp("", "backup-ssh-")
	if err != nil {
		return nil, err
	}
	fail := func(err error) (io.ReadCloser, error) {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// ssh refuses keys other users can read, the admin's copy is often 0644
	keyPath := filepath.Join(tmpDir, "key")
	if err := os.WriteFile(keyPath, keyBytes, 0o600); err != nil {
		return fail(err)
	}
	knownHosts := ""
	if line := knownHostsLine(config); line != "" {
		knownHosts = filepath.Join(tmpDir, "known_hosts")
		if err := os.WriteFile(knownHosts, []byte(line), 0o600); err != nil {
			return fail(err)
		}
	}

	cmd := exec.CommandContext(ctx, opensshBinary, opensshArgs(config, keyPath, knownHosts, "cat -- "+shellQuoteArg(remoteFile))...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fail(err)
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	f := &opensshFile{r: bufio.NewReaderSize(stdout, 1<<20), cmd: cmd, stderr: stderr, tmpDir: tmpDir}

	// SFTP-only accounts reject the command, a gzip archive has to start with its magic bytes or this isn't the file
	want := 1
	if strings.HasSuffix(remoteFile, ".gz") {
		want = 2
	}
	head, err := f.r.Peek(want)
	if err != nil || (want == 2 && !bytes.Equal(head, []byte{0x1f, 0x8b})) {
		_ = f.Close()
		if f.waitErr != nil {
			return nil, f.waitErr
		}
		return nil, fmt.Errorf("ssh: unexpected output for %s", remoteFile)
	}
	return f, nil
}

// openSFTPOrSSH tries OpenSSH first when allowed and logs why it fell back to the Go client
func openSFTPOrSSH(ctx context.Context, config map[string]string, remotePath string, viaGo func() (io.ReadCloser, error)) (io.ReadCloser, error) {
	if opensshAllowed(config) {
		rc, err := openWithOpenSSH(ctx, config, remotePath)
		if err == nil {
			return rc, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		log.Printf("backups: openssh download of %s from %s failed, using sftp: %v", remotePath, config["SSH_HOST_NAME"], err)
	}
	return viaGo()
}
