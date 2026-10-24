package backups

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// loadPrivateKey auto-detects the key algorithm (Ed25519/ECDSA/RSA/DSA) from the PEM headers via ssh.ParsePrivateKey, and is used uniformly everywhere an SSH connection is opened - a restore shouldn't need a different key type than a reindex against the same destination
func loadPrivateKey(path, passphrase string) (ssh.Signer, error) {
	pemBytes, err := tenantfs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if passphrase == "" {
		return ssh.ParsePrivateKey(pemBytes)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
	if err != nil {
		// fall back to unencrypted parsing in case the key doesn't actually require the configured passphrase
		return ssh.ParsePrivateKey(pemBytes)
	}
	return signer, nil
}

// sshAuthMethods mirrors docker-volume-backup: the key file when it can be read, the password otherwise, the template's default SSH_IDENTITY_FILE used to fail every password-only destination
func sshAuthMethods(config map[string]string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if keyPath := config["SSH_IDENTITY_FILE"]; keyPath != "" {
		signer, err := loadPrivateKey(keyPath, config["SSH_IDENTITY_PASSPHRASE"])
		switch {
		case err == nil:
			methods = append(methods, ssh.PublicKeys(signer))
		case !errors.Is(err, fs.ErrNotExist) || config["SSH_PASSWORD"] == "":
			return nil, err
		}
	}
	if password := config["SSH_PASSWORD"]; password != "" {
		methods = append(methods, ssh.Password(password))
	}
	if len(methods) == 0 {
		return nil, errors.New("no SSH password or identity file configured")
	}
	return methods, nil
}

// hostKeyCallback checks SSH_HOST_PUBLIC_KEY when there is one, older backup.env files don't have it so those still trust the host on connect
func hostKeyCallback(config map[string]string) (ssh.HostKeyCallback, error) {
	raw := strings.TrimSpace(config["SSH_HOST_PUBLIC_KEY"])
	if raw == "" {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // no pinned key to check against yet
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("invalid SSH_HOST_PUBLIC_KEY: %w", err)
	}
	return ssh.FixedHostKey(key), nil
}

func sshAddr(config map[string]string) string {
	port := config["SSH_PORT"]
	if port == "" {
		port = "22"
	}
	return net.JoinHostPort(config["SSH_HOST_NAME"], port)
}

// dialSSH connects to the remote backup host with the same auth and host key the backup container uses
func dialSSH(config map[string]string) (*ssh.Client, error) {
	auth, err := sshAuthMethods(config)
	if err != nil {
		return nil, err
	}
	hostKey, err := hostKeyCallback(config)
	if err != nil {
		return nil, err
	}
	clientConfig := &ssh.ClientConfig{
		User:            config["SSH_USER"],
		Auth:            auth,
		HostKeyCallback: hostKey,
		Timeout:         15 * time.Second,
	}
	return ssh.Dial("tcp", sshAddr(config), clientConfig)
}

// errGotHostKey stops the handshake once the host key is in, fetchHostKey doesn't need to log in
var errGotHostKey = errors.New("got host key")

// fetchHostKey reads the destination's host key in authorized_keys form, docker-volume-backup refuses SSH without one
func fetchHostKey(config map[string]string) (string, error) {
	var found string
	clientConfig := &ssh.ClientConfig{
		User: config["SSH_USER"],
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			found = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
			return errGotHostKey
		},
		Timeout: 15 * time.Second,
	}
	_, err := ssh.Dial("tcp", sshAddr(config), clientConfig)
	if found != "" {
		return found, nil
	}
	if err == nil {
		err = errors.New("no host key received")
	}
	return "", err
}

// runSSHCommand runs a single command over a fresh channel/session on the shared connection and returns stdout only
func runSSHCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.Output(command)
	return string(out), err
}

// sshStore is the remoteStore implementation for the "ssh" backup.env section. Each method opens its own short-lived connection rather than sharing one across a doReindex run - simpler than plumbing connection lifetime through the remoteStore interface for a destination that's only reindexed manually.
type sshStore struct {
	config map[string]string
}

// List reports the remote backup directory's entries via SFTP rather than shelling out to `ls`, so it works the same regardless of what shell (or lack thereof) the SSH destination provides
func (s *sshStore) List(ctx context.Context) ([]remoteFile, error) {
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

	var files []remoteFile
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, remoteFile{Name: entry.Name(), Size: entry.Size()})
		}
	}
	return files, nil
}

// sftpFile closes the sftp session and ssh connection along with the file
type sftpFile struct {
	*sftp.File
	sftpClient *sftp.Client
	client     *ssh.Client
}

func (f *sftpFile) Close() error {
	err := f.File.Close()
	f.sftpClient.Close()
	f.client.Close()
	return err
}

func (s *sshStore) Open(ctx context.Context, filename string) (io.ReadCloser, error) {
	client, err := dialSSH(s.config)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	remotePath := strings.TrimRight(s.config["SSH_REMOTE_PATH"], "/")
	remoteFile, err := sftpClient.Open(remotePath + "/" + filename)
	if err != nil {
		sftpClient.Close()
		client.Close()
		return nil, err
	}
	return &sftpFile{File: remoteFile, sftpClient: sftpClient, client: client}, nil
}

// ListMembers runs `tar -tzf` on the remote host, so archives without a manifest can still be indexed without downloading them
func (s *sshStore) ListMembers(ctx context.Context, filename string) (string, error) {
	client, err := dialSSH(s.config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	remotePath := strings.TrimRight(s.config["SSH_REMOTE_PATH"], "/")
	return runSSHCommand(client, "tar -tzf "+shellQuoteArg(remotePath+"/"+filename)+" 2>/dev/null")
}
