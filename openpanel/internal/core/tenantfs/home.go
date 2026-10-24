package tenantfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"golang.org/x/sys/unix"
)

// ErrSymlink is returned when a home-root path resolves through a symlink. The account owns its own
// home dir, so a root-run op that followed one could be redirected onto another account's files, the
// panel's config, or the host - see volumeDir for the stricter volume case.
var ErrSymlink = errors.New("path goes through a symlink in an account home")

// ErrNotRegular is returned when a home-root file we're about to read or write isn't a regular file,
// e.g. the account swapped it for a fifo or device to stall or trick a root-run read.
var ErrNotRegular = errors.New("not a regular file")

// homeSplit reports whether p sits directly under an account home (/home/<ctx>/...) rather than in a
// volume (volumeOf handles those and never reaches here). base is the account home, rel the path
// under it ("" for the home dir itself). The account owns everything under base, so these ops never
// follow a symlink on the way down.
func homeSplit(p string) (base, rel string, ok bool) {
	rest, found := strings.CutPrefix(path.Clean(p), homeRoot)
	if !found {
		return "", "", false
	}
	ctx, sub, _ := strings.Cut(rest, "/")
	if ctx == "" || ctx == "." || ctx == ".." {
		return "", "", false
	}
	return homeRoot + ctx, sub, true
}

// resolveHomeParent opens the parent directory of rel under base, opening each component with
// O_NOFOLLOW so a symlink anywhere on the way fails instead of redirecting us. name is the final
// component the caller then operates on (". " means rel was the home dir itself). The returned fd is
// the caller's to close.
func resolveHomeParent(base, rel string) (parentFd int, name string, err error) {
	bfd, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, "", &fs.PathError{Op: "open", Path: base, Err: err}
	}
	dir, name := path.Split(rel)
	cur := bfd
	for _, comp := range strings.Split(dir, "/") {
		if comp == "" || comp == "." {
			continue
		}
		if comp == ".." {
			unix.Close(cur)
			return -1, "", &fs.PathError{Op: "open", Path: base + "/" + rel, Err: ErrEscape}
		}
		next, oerr := unix.Openat(cur, comp, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(cur)
		if oerr != nil {
			return -1, "", &fs.PathError{Op: "open", Path: base + "/" + rel, Err: mapHomeErr(oerr)}
		}
		cur = next
	}
	if name == "" {
		name = "."
	}
	return cur, name, nil
}

// mapHomeErr turns the raw errno from a refused traversal into our own sentinels so callers and
// messages can tell a planted symlink apart from an ordinary failure.
func mapHomeErr(err error) error {
	switch err {
	case unix.ELOOP:
		return ErrSymlink
	case unix.EXDEV:
		return ErrEscape
	}
	return err
}

// openHomeFile opens name under base/rel for real IO, O_NOFOLLOW on the final component so a symlink
// left in its place is refused rather than followed. O_NONBLOCK keeps a fifo from stalling the open.
func openHomeFile(base, rel string, flag int, perm os.FileMode) (*os.File, error) {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return nil, err
	}
	defer unix.Close(pfd)
	var mode uint32
	if flag&os.O_CREATE != 0 {
		mode = uint32(perm.Perm())
	}
	fd, err := unix.Openat(pfd, name, flag|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, mode)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: base + "/" + rel, Err: mapHomeErr(err)}
	}
	return os.NewFile(uintptr(fd), base+"/"+rel), nil
}

// requireRegular rejects anything that isn't a plain file, so a fifo/device/dir planted at a config
// path can't stall a root-run read or be written through.
func requireRegular(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return &fs.PathError{Op: "open", Path: f.Name(), Err: ErrNotRegular}
	}
	return nil
}

func homeReadFile(base, rel string) ([]byte, error) {
	f, err := openHomeFile(base, rel, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := requireRegular(f); err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

func homeWriteFile(base, rel string, data []byte, perm os.FileMode) error {
	f, err := openHomeFile(base, rel, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if err := requireRegular(f); err != nil {
		f.Close()
		return err
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// homeLstat stats the final component without following it, the same as Lstat.
func homeLstat(base, rel string) (fs.FileInfo, error) {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return nil, err
	}
	defer unix.Close(pfd)
	// O_PATH|O_NOFOLLOW gives a handle to the entry itself even when it's a symlink, so fstat reports the link
	fd, err := unix.Openat(pfd, name, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: base + "/" + rel, Err: mapHomeErr(err)}
	}
	f := os.NewFile(uintptr(fd), base+"/"+rel)
	defer f.Close()
	return f.Stat()
}

// homeStat is Stat but refuses a final symlink instead of following it out of the home.
func homeStat(base, rel string) (fs.FileInfo, error) {
	fi, err := homeLstat(base, rel)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&fs.ModeSymlink != 0 {
		return nil, &fs.PathError{Op: "stat", Path: base + "/" + rel, Err: ErrSymlink}
	}
	return fi, nil
}

func homeReadDir(base, rel string) ([]fs.DirEntry, error) {
	f, err := openHomeFile(base, rel, os.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.ReadDir(-1)
}

func homeChmod(base, rel string, mode os.FileMode) error {
	f, err := openHomeFile(base, rel, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Chmod(mode)
}

// homeChown chowns the entry itself without following a symlink, so a planted link can't retarget the
// chown onto someone else's file. This is stricter than the volume Chown, which may follow in-volume.
func homeChown(base, rel string, uid, gid int) error {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return err
	}
	defer unix.Close(pfd)
	if err := unix.Fchownat(pfd, name, uid, gid, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &fs.PathError{Op: "chown", Path: base + "/" + rel, Err: err}
	}
	return nil
}

func homeRemove(base, rel string) error {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return err
	}
	defer unix.Close(pfd)
	return unlinkatAny(pfd, name, base+"/"+rel)
}

// unlinkatAny removes name under dirfd whether it's a file or a dir, matching os.Remove.
func unlinkatAny(dirfd int, name, full string) error {
	err := unix.Unlinkat(dirfd, name, 0)
	if err == unix.EISDIR || err == unix.EPERM {
		err = unix.Unlinkat(dirfd, name, unix.AT_REMOVEDIR)
	}
	if err != nil {
		return &fs.PathError{Op: "remove", Path: full, Err: err}
	}
	return nil
}

// homeRemoveAll is RemoveAll for a home path: it resolves the parent without following symlinks, then
// deletes name and everything under it, never descending through a symlink.
func homeRemoveAll(base, rel string) error {
	if rel == "" || rel == "." {
		return &fs.PathError{Op: "removeall", Path: base, Err: ErrEscape}
	}
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer unix.Close(pfd)
	return removeAllAt(pfd, name, base+"/"+rel)
}

func removeAllAt(dirfd int, name, full string) error {
	err := unix.Unlinkat(dirfd, name, 0)
	if err == nil || err == unix.ENOENT {
		return nil
	}
	if err != unix.EISDIR && err != unix.EPERM {
		return &fs.PathError{Op: "remove", Path: full, Err: err}
	}
	fd, oerr := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if oerr != nil {
		if oerr == unix.ENOENT {
			return nil
		}
		// not a directory after all, retry as a plain unlink
		if oerr == unix.ENOTDIR || oerr == unix.ELOOP {
			if uerr := unix.Unlinkat(dirfd, name, 0); uerr != nil && uerr != unix.ENOENT {
				return &fs.PathError{Op: "remove", Path: full, Err: uerr}
			}
			return nil
		}
		return &fs.PathError{Op: "open", Path: full, Err: oerr}
	}
	f := os.NewFile(uintptr(fd), full)
	entries, rerr := f.ReadDir(-1)
	if rerr == nil {
		for _, e := range entries {
			if aerr := removeAllAt(int(f.Fd()), e.Name(), full+"/"+e.Name()); aerr != nil {
				f.Close()
				return aerr
			}
		}
	}
	f.Close()
	if uerr := unix.Unlinkat(dirfd, name, unix.AT_REMOVEDIR); uerr != nil && uerr != unix.ENOENT {
		return &fs.PathError{Op: "remove", Path: full, Err: uerr}
	}
	return nil
}

func homeMkdir(base, rel string, perm os.FileMode) error {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return err
	}
	defer unix.Close(pfd)
	if err := unix.Mkdirat(pfd, name, uint32(perm.Perm())); err != nil {
		return &fs.PathError{Op: "mkdir", Path: base + "/" + rel, Err: err}
	}
	return nil
}

// homeMkdirAll is MkdirAll for a home path, creating each missing component while opening every
// existing one with O_NOFOLLOW so a symlink in the chain is refused.
func homeMkdirAll(base, rel string, perm os.FileMode) error {
	if rel == "" || rel == "." {
		return nil // the home dir already exists
	}
	bfd, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return &fs.PathError{Op: "open", Path: base, Err: err}
	}
	cur := bfd
	for _, comp := range strings.Split(rel, "/") {
		if comp == "" || comp == "." {
			continue
		}
		if comp == ".." {
			unix.Close(cur)
			return &fs.PathError{Op: "mkdir", Path: base + "/" + rel, Err: ErrEscape}
		}
		if merr := unix.Mkdirat(cur, comp, uint32(perm.Perm())); merr != nil && merr != unix.EEXIST {
			unix.Close(cur)
			return &fs.PathError{Op: "mkdir", Path: base + "/" + rel, Err: merr}
		}
		next, oerr := unix.Openat(cur, comp, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(cur)
		if oerr != nil {
			return &fs.PathError{Op: "mkdir", Path: base + "/" + rel, Err: mapHomeErr(oerr)}
		}
		cur = next
	}
	unix.Close(cur)
	return nil
}

func homeReadlink(base, rel string) (string, error) {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return "", err
	}
	defer unix.Close(pfd)
	buf := make([]byte, 4096)
	n, err := unix.Readlinkat(pfd, name, buf)
	if err != nil {
		return "", &fs.PathError{Op: "readlink", Path: base + "/" + rel, Err: err}
	}
	return string(buf[:n]), nil
}

func homeSymlink(target, base, rel string) error {
	pfd, name, err := resolveHomeParent(base, rel)
	if err != nil {
		return err
	}
	defer unix.Close(pfd)
	if err := unix.Symlinkat(target, pfd, name); err != nil {
		return &fs.PathError{Op: "symlink", Path: base + "/" + rel, Err: err}
	}
	return nil
}
