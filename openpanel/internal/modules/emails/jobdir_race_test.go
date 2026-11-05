package emails

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"golang.org/x/sys/unix"
)

func TestOpenJobDirSwapRaceNeverReachesAnotherAccount(t *testing.T) {
	home, _ := withMigrateHome(t)
	victim := filepath.Join(home, "victim", "imapsync")
	if err := os.MkdirAll(victim, 0o700); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(victim, "marker.log"), []byte("VICTIM"), 0o600)
	own := filepath.Join(home, "ctx1", "imapsync")
	os.Mkdir(own, 0o700)
	swap := filepath.Join(home, "ctx1", "swap")
	if err := os.Symlink(victim, swap); err != nil {
		t.Fatal(err)
	}
	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		for !stop.Load() {
			_ = unix.Renameat2(unix.AT_FDCWD, own, unix.AT_FDCWD, swap, unix.RENAME_EXCHANGE)
		}
	}()
	for i := 0; i < 5000; i++ {
		root, _, _, err := openJobDir("ctx1")
		if err != nil {
			continue
		}
		if _, err := root.Stat("marker.log"); err == nil {
			stop.Store(true)
			<-done
			t.Fatalf("attempt %d opened the other account's job dir", i)
		}
		root.Close()
	}
	stop.Store(true)
	<-done
}
