package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// withHomeVolume creates a temp "home volume" directory for context "alice"
// and points homeVolumeOverride at its parent for the duration of the test.
func withHomeVolume(t *testing.T) (root string, context string) {
	t.Helper()
	tmp := t.TempDir()
	context = "alice"
	root = filepath.Join(tmp, context)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	homeVolumeOverride = tmp
	t.Cleanup(func() { homeVolumeOverride = "" })
	return root, context
}

func TestSecureUserPathEmptyReturnsHome(t *testing.T) {
	root, context := withHomeVolume(t)
	got, err := SecureUserPath("HOME", context, "", false)
	if err != nil {
		t.Fatalf("SecureUserPath: %v", err)
	}
	wantRoot, _ := filepath.EvalSymlinks(root)
	if got != wantRoot {
		t.Errorf("got %q, want %q", got, wantRoot)
	}
}

func TestSecureUserPathValidNestedPath(t *testing.T) {
	root, context := withHomeVolume(t)
	if err := os.MkdirAll(filepath.Join(root, "documents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "documents", "report.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := SecureUserPath("HOME", context, "documents/report.txt", true)
	if err != nil {
		t.Fatalf("SecureUserPath: %v", err)
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(root, "documents", "report.txt"))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSecureUserPathCheckExistsTrueMissing(t *testing.T) {
	_, context := withHomeVolume(t)
	_, err := SecureUserPath("HOME", context, "does/not/exist.txt", true)
	perr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %v (%T)", err, err)
	}
	if perr.Code != 404 {
		t.Errorf("Code = %d, want 404", perr.Code)
	}
}

func TestSecureUserPathCheckExistsFalseAllowsMissing(t *testing.T) {
	root, context := withHomeVolume(t)
	got, err := SecureUserPath("HOME", context, "new-file.txt", false)
	if err != nil {
		t.Fatalf("SecureUserPath: %v", err)
	}
	want := filepath.Join(root, "new-file.txt")
	wantResolvedRoot, _ := filepath.EvalSymlinks(root)
	wantResolved := filepath.Join(wantResolvedRoot, "new-file.txt")
	if got != want && got != wantResolved {
		t.Errorf("got %q, want %q or %q", got, want, wantResolved)
	}
}

func TestSecureUserPathDotDotTraversalRejected(t *testing.T) {
	_, context := withHomeVolume(t)
	for _, attempt := range []string{
		"../etc/passwd",
		"documents/../../etc/passwd",
		"..%2f etc/passwd",
		"..%5cetc",
	} {
		_, err := SecureUserPath("HOME", context, attempt, false)
		perr, ok := err.(*Error)
		if !ok || perr.Code != 403 {
			t.Errorf("SecureUserPath(%q) = %v, want a 403 *Error", attempt, err)
		}
	}
}

func TestSecureUserPathSymlinkEscapeRejected(t *testing.T) {
	root, context := withHomeVolume(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks not supported in this environment: %v", err)
	}

	_, err := SecureUserPath("HOME", context, "escape/secret.txt", true)
	perr, ok := err.(*Error)
	if !ok || perr.Code != 403 {
		t.Errorf("SecureUserPath through a symlink = %v, want a 403 *Error", err)
	}
}

func TestSecureUserPathControlCharsRejected(t *testing.T) {
	_, context := withHomeVolume(t)
	_, err := SecureUserPath("HOME", context, "bad\x00name.txt", false)
	perr, ok := err.(*Error)
	if !ok || perr.Code != 403 {
		t.Errorf("SecureUserPath with a control char = %v, want a 403 *Error", err)
	}
}

func TestSecureUserPathTooLongRejected(t *testing.T) {
	_, context := withHomeVolume(t)
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	_, err := SecureUserPath("HOME", context, string(long), false)
	perr, ok := err.(*Error)
	if !ok || perr.Code != 403 {
		t.Errorf("SecureUserPath with an overlong path = %v, want a 403 *Error", err)
	}
}

func TestSecureUserPathInvalidContextRejected(t *testing.T) {
	for _, ctx := range []string{"", "a/b", "a\\b"} {
		_, err := SecureUserPath("HOME", ctx, "file.txt", false)
		perr, ok := err.(*Error)
		if !ok || perr.Code != 403 {
			t.Errorf("SecureUserPath with context %q = %v, want a 403 *Error", ctx, err)
		}
	}
}

func TestSecureUserPathMissingVolumeIs404(t *testing.T) {
	homeVolumeOverride = t.TempDir() // exists, but "nobody" subdir under it doesn't
	t.Cleanup(func() { homeVolumeOverride = "" })

	_, err := SecureUserPath("HOME", "nobody", "file.txt", false)
	perr, ok := err.(*Error)
	if !ok || perr.Code != 404 {
		t.Errorf("SecureUserPath with a missing volume = %v, want a 404 *Error", err)
	}
}

func TestSecureUserPathAbsoluteLikeInputTreatedAsRelative(t *testing.T) {
	root, context := withHomeVolume(t)
	if err := os.MkdirAll(filepath.Join(root, "etc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "etc", "passwd"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A leading "/" isn't rejected - it's deliberately treated as relative
	// and just joined, staying inside userHome.
	got, err := SecureUserPath("HOME", context, "/etc/passwd", true)
	if err != nil {
		t.Fatalf("SecureUserPath: %v", err)
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(root, "etc", "passwd"))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSecureUserPathInvalidBaseDirRejected(t *testing.T) {
	_, context := withHomeVolume(t)
	_, err := SecureUserPath("VOLUME", context, "file.txt", false)
	perr, ok := err.(*Error)
	if !ok || perr.Code != 403 {
		t.Errorf("SecureUserPath with base VOLUME = %v, want a 403 *Error", err)
	}
}
