package usersocket

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestDialRefusesSymlinks(t *testing.T) {
	tmp := t.TempDir()
	old := HomeRoot
	HomeRoot = tmp + "/"
	defer func() { HomeRoot = old }()
	for _, d := range []string{"alice/sockets/mysqld", "bob/sockets/mysqld", "carol/sockets"} {
		os.MkdirAll(filepath.Join(tmp, d), 0o755)
	}
	l, err := net.Listen("unix", filepath.Join(tmp, "bob/sockets/mysqld/mysqld.sock"))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	ctx := context.Background()
	if c, err := Dial(ctx, "bob", "sockets/mysqld/mysqld.sock"); err != nil {
		t.Fatalf("own socket: %v", err)
	} else {
		c.Close()
	}
	// leaf symlink to another account's socket
	os.Symlink(filepath.Join(tmp, "bob/sockets/mysqld/mysqld.sock"), filepath.Join(tmp, "alice/sockets/mysqld/mysqld.sock"))
	if _, err := Dial(ctx, "alice", "sockets/mysqld/mysqld.sock"); err == nil {
		t.Fatal("followed a leaf symlink")
	}
	// directory symlink
	os.Symlink(filepath.Join(tmp, "bob/sockets/mysqld"), filepath.Join(tmp, "carol/sockets/mysqld"))
	if _, err := Dial(ctx, "carol", "sockets/mysqld/mysqld.sock"); err == nil {
		t.Fatal("followed a directory symlink")
	}
	if Exists("alice", "sockets/mysqld/mysqld.sock") || !Exists("bob", "sockets/mysqld/mysqld.sock") {
		t.Fatal("Exists is wrong")
	}
	if _, err := Dial(ctx, "../bob", "sockets/mysqld/mysqld.sock"); err == nil {
		t.Fatal("took a context with a slash")
	}
}
