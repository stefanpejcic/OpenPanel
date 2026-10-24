package backups

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestSSHAuthMethodsMissingKeyFallsBackToPassword(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "id_rsa")

	methods, err := sshAuthMethods(map[string]string{"SSH_IDENTITY_FILE": missing, "SSH_PASSWORD": "secret"})
	if err != nil || len(methods) != 1 {
		t.Fatalf("missing key with a password: got %d methods, err %v", len(methods), err)
	}
	if _, err := sshAuthMethods(map[string]string{"SSH_IDENTITY_FILE": missing}); err == nil {
		t.Error("missing key and no password should fail")
	}
	if _, err := sshAuthMethods(map[string]string{}); err == nil {
		t.Error("no credentials at all should fail")
	}
}

func newHostKey(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func TestHostKeyCallback(t *testing.T) {
	pinned, other := newHostKey(t), newHostKey(t)
	authorized := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pinned.PublicKey())))
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 22}

	cb, err := hostKeyCallback(map[string]string{"SSH_HOST_PUBLIC_KEY": authorized})
	if err != nil {
		t.Fatal(err)
	}
	if err := cb("h:22", addr, pinned.PublicKey()); err != nil {
		t.Errorf("pinned key rejected: %v", err)
	}
	if err := cb("h:22", addr, other.PublicKey()); err == nil {
		t.Error("a different host key was accepted")
	}
	if _, err := hostKeyCallback(map[string]string{"SSH_HOST_PUBLIC_KEY": "not a key"}); err == nil {
		t.Error("an invalid key should be an error")
	}
	if cb, err := hostKeyCallback(map[string]string{}); err != nil || cb == nil {
		t.Error("no pinned key should still connect")
	}
}

func TestFetchHostKey(t *testing.T) {
	hostKey := newHostKey(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	serverConfig := &ssh.ServerConfig{PasswordCallback: func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) {
		return nil, errors.New("no logins")
	}}
	serverConfig.AddHostKey(hostKey)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _, _ = ssh.NewServerConn(conn, serverConfig)
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	got, err := fetchHostKey(map[string]string{"SSH_HOST_NAME": host, "SSH_PORT": port, "SSH_USER": "backup"})
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(hostKey.PublicKey()))); got != want {
		t.Errorf("fetchHostKey = %q, want %q", got, want)
	}
}

func TestPinSSHHostKey(t *testing.T) {
	const fetched = "ssh-ed25519 AAAAfetched"
	cases := []struct {
		name          string
		before, after map[string]string
		typed         string
		wantFetch     bool
	}{
		{"no ssh destination", map[string]string{}, map[string]string{"AWS_S3_BUCKET_NAME": "b"}, "", false},
		{"no key yet", map[string]string{}, map[string]string{"SSH_HOST_NAME": "h"}, "", true},
		{"key kept for the same host", map[string]string{"SSH_HOST_NAME": "h", "SSH_HOST_PUBLIC_KEY": `"ssh-rsa old"`}, map[string]string{"SSH_HOST_NAME": "h", "SSH_HOST_PUBLIC_KEY": `"ssh-rsa old"`}, "", false},
		{"host changed", map[string]string{"SSH_HOST_NAME": "h", "SSH_HOST_PUBLIC_KEY": `"ssh-rsa old"`}, map[string]string{"SSH_HOST_NAME": "h2", "SSH_HOST_PUBLIC_KEY": `"ssh-rsa old"`}, "", true},
		{"port changed", map[string]string{"SSH_HOST_NAME": "h", "SSH_PORT": "22", "SSH_HOST_PUBLIC_KEY": "k"}, map[string]string{"SSH_HOST_NAME": "h", "SSH_PORT": "2222", "SSH_HOST_PUBLIC_KEY": "k"}, "", true},
		{"key typed in", map[string]string{}, map[string]string{"SSH_HOST_NAME": "h2"}, "ssh-rsa typed", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			updated := map[string]string{}
			if c.typed != "" {
				updated["SSH_HOST_PUBLIC_KEY"] = c.typed
				c.after["SSH_HOST_PUBLIC_KEY"] = c.typed
			}
			called := false
			pinSSHHostKey(c.before, c.after, updated, func(map[string]string) (string, error) {
				called = true
				return fetched, nil
			})
			if called != c.wantFetch {
				t.Fatalf("fetched = %v, want %v", called, c.wantFetch)
			}
			if c.wantFetch && updated["SSH_HOST_PUBLIC_KEY"] != `"`+fetched+`"` {
				t.Errorf("updated key = %q", updated["SSH_HOST_PUBLIC_KEY"])
			}
		})
	}

	updated := map[string]string{}
	pinSSHHostKey(map[string]string{}, map[string]string{"SSH_HOST_NAME": "h"}, updated, func(map[string]string) (string, error) {
		return "", errors.New("unreachable")
	})
	if _, set := updated["SSH_HOST_PUBLIC_KEY"]; set {
		t.Error("a failed fetch must not write a key")
	}
}
