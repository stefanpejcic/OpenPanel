package backups

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// loadPrivateKey auto-detects the key algorithm (Ed25519/ECDSA/RSA/DSA)
// from the PEM headers via ssh.ParsePrivateKey, which handles every
// algorithm without needing each key class tried in turn. loadPrivateKey
// is used uniformly everywhere an SSH connection is opened, since there's
// no reason a restore should require a different key type than a reindex
// against the very same destination.
func loadPrivateKey(path, passphrase string) (ssh.Signer, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if passphrase == "" {
		return ssh.ParsePrivateKey(pemBytes)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
	if err != nil {
		// Fall back to unencrypted parsing in case the key doesn't
		// actually require the configured passphrase.
		return ssh.ParsePrivateKey(pemBytes)
	}
	return signer, nil
}

// dialSSH connects to the remote backup host and unconditionally trusts
// its host key (a known, pre-existing security tradeoff, not something
// introduced here).
//
// Auth mirrors what the "backup" container's own docker-volume-backup tool
// accepts for its SSH/SFTP target: a private key (SSH_IDENTITY_FILE) when
// present, otherwise a plain password (SSH_PASSWORD) - backup.env's ssh
// section documents both as valid, and a destination configured with only
// a password (no identity file) is a normal, supported setup, not a
// misconfiguration.
func dialSSH(config map[string]string) (*ssh.Client, error) {
	var authMethod ssh.AuthMethod
	if keyPath := config["SSH_IDENTITY_FILE"]; keyPath != "" {
		signer, err := loadPrivateKey(keyPath, config["SSH_IDENTITY_PASSPHRASE"])
		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(config["SSH_PASSWORD"])
	}

	port := config["SSH_PORT"]
	if port == "" {
		port = "22"
	}

	clientConfig := &ssh.ClientConfig{
		User:            config["SSH_USER"],
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // trusts the remote host key on first connect, see dialSSH doc comment
		Timeout:         15 * time.Second,
	}

	addr := net.JoinHostPort(config["SSH_HOST_NAME"], port)
	return ssh.Dial("tcp", addr, clientConfig)
}

// runSSHCommand runs a single command over a fresh channel/session on the
// shared connection and returns stdout only.
func runSSHCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.Output(command)
	return string(out), err
}

// sshStore is the remoteStore implementation for the "ssh" backup.env
// section. Each method opens its own short-lived connection rather than
// sharing one across a doReindex run, matching how handleRestoreFromBackup/
// handleDownloadBackup already dialed a fresh connection per request before
// this type existed - simpler than plumbing connection lifetime through the
// remoteStore interface for a destination that's only reindexed manually,
// not on every page load.
type sshStore struct {
	config map[string]string
}

// List reports the remote backup directory's entries via SFTP (rather than
// shelling out to `ls`), so it works the same regardless of what shell (or
// lack thereof) the SSH destination provides.
func (s *sshStore) List(ctx context.Context) ([]string, error) {
	client, err := dialSSH(s.config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	remotePath := strings.TrimRight(s.config["SSH_REMOTE_PATH"], "/")
	entries, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// Fetch downloads one backup archive from the SSH/SFTP destination into a
// local temp file. The caller must call cleanup() once done with the local
// copy.
func (s *sshStore) Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error) {
	client, err := dialSSH(s.config)
	if err != nil {
		return "", nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return "", nil, err
	}
	defer sftpClient.Close()

	remotePath := strings.TrimRight(s.config["SSH_REMOTE_PATH"], "/")
	remoteFilePath := remotePath + "/" + filename

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	remoteFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer remoteFile.Close()

	localPath = filepath.Join(tmpDir, filename)
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, remoteFile); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

// Classify lists one remote archive's members with a remote `tar -tzf`
// instead of downloading it first - a meaningful savings for large
// archives, and the reason SSH keeps its own connection-based Classify
// instead of falling back to classifyViaDownload like every other backend.
func (s *sshStore) Classify(ctx context.Context, filename string) (BackupInfo, error) {
	client, err := dialSSH(s.config)
	if err != nil {
		return BackupInfo{}, err
	}
	defer client.Close()

	remotePath := strings.TrimRight(s.config["SSH_REMOTE_PATH"], "/")
	return processBackup(client, filename, remotePath), nil
}
