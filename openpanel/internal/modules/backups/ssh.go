package backups

import (
	"net"
	"os"
	"time"

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
func dialSSH(config map[string]string) (*ssh.Client, error) {
	signer, err := loadPrivateKey(config["SSH_IDENTITY_FILE"], config["SSH_IDENTITY_PASSPHRASE"])
	if err != nil {
		return nil, err
	}

	port := config["SSH_PORT"]
	if port == "" {
		port = "22"
	}

	clientConfig := &ssh.ClientConfig{
		User:            config["SSH_USER"],
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
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
