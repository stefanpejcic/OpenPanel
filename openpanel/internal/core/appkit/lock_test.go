package appkit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateLockFileExclusive(t *testing.T) {
	old := lockBaseDir
	lockBaseDir = t.TempDir() + "/"
	t.Cleanup(func() { lockBaseDir = old })

	if err := CreateLockFile("alice"); err != nil {
		t.Fatalf("first lock: %v", err)
	}
	// a second acquire while alice's lock is held must be refused
	if err := CreateLockFile("alice"); !errors.Is(err, ErrInstallInProgress) {
		t.Fatalf("second lock = %v, want ErrInstallInProgress", err)
	}
	// a different user is unaffected
	if err := CreateLockFile("bob"); err != nil {
		t.Fatalf("other user lock: %v", err)
	}
	// releasing lets the next one through
	RemoveLockFile("alice")
	if err := CreateLockFile("alice"); err != nil {
		t.Fatalf("re-lock after release: %v", err)
	}
}

func TestCreateLockFileTakesOverStaleLock(t *testing.T) {
	old := lockBaseDir
	lockBaseDir = t.TempDir() + "/"
	t.Cleanup(func() { lockBaseDir = old })

	path := LockFilePath("carol")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-installLockStaleAfter - time.Minute)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}

	if err := CreateLockFile("carol"); err != nil {
		t.Fatalf("stale lock should be taken over, got %v", err)
	}
}
