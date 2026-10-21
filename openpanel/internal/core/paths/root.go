package paths

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// UserRoot is an open handle on a user's HOME/HTML dir, every op through it stays inside even if a dir gets swapped for a symlink mid-request
type UserRoot struct {
	*os.Root
	Home string
	// dir is Home as callers spell it, before symlinks in the prefix were resolved
	dir string
}

// OpenUserRoot opens the user's base dir, callers must Close it
func OpenUserRoot(base, context string) (*UserRoot, error) {
	if context == "" || strings.ContainsAny(context, "/\\") {
		return nil, abort(403, "Invalid context")
	}
	home, err := baseDir(base, context)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return nil, abort(404, "User root directory not found")
	}
	return &UserRoot{Root: root, Home: home, dir: home}, nil
}

// VolumeDir is where an account's named podman volume keeps its data on the host
func VolumeDir(userContext, volume string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_" + volume + "/_data"
}

// OpenVolume opens an account volume like "html_data", the account's containers can write anywhere under it so host-side ops on it must go through the root
func OpenVolume(userContext, volume string) (*UserRoot, error) {
	if userContext == "" || strings.ContainsAny(userContext, "/\\") || volume == "" || strings.ContainsAny(volume, "/\\") {
		return nil, abort(403, "Invalid context")
	}
	if volume == "html_data" {
		root, err := OpenUserRoot("HOME", userContext)
		if err == nil && homeVolumeOverride == "" {
			root.dir = VolumeDir(userContext, volume)
		}
		return root, err
	}
	return OpenDir(VolumeDir(userContext, volume))
}

// OpenDir opens a root at a host dir whose contents aren't trusted, the dir path itself has to be
func OpenDir(dir string) (*UserRoot, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, abort(404, "Directory not found")
	}
	root, err := os.OpenRoot(resolved)
	if err != nil {
		return nil, abort(404, "Directory not found")
	}
	return &UserRoot{Root: root, Home: resolved, dir: filepath.Clean(dir)}, nil
}

// Under turns a host path the caller built under the root into a checked relative one, refusing .. escapes and symlinks on the way
func (u *UserRoot) Under(hostPath string) (string, error) {
	clean := filepath.Clean(hostPath)
	for _, base := range []string{u.dir, u.Home} {
		rel, err := filepath.Rel(base, clean)
		if err != nil || rel == ".." || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) {
			continue
		}
		return u.relParts(rel, false, 4096)
	}
	return "", abort(403, "Path traversal or symlink escape detected")
}

// Rel runs the same checks as SecureUserPath and returns the path relative to the root, "." for the root itself
func (u *UserRoot) Rel(userInputPath string, checkExists bool) (string, error) {
	return u.relParts(userInputPath, checkExists, 200)
}

func (u *UserRoot) relParts(userInputPath string, checkExists bool, maxLen int) (string, error) {
	parts, err := splitPath(userInputPath, maxLen)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return ".", nil
	}
	for i := range parts {
		info, statErr := u.Lstat(path.Join(parts[:i+1]...))
		if statErr != nil {
			if errors.Is(statErr, fs.ErrNotExist) {
				break
			}
			return "", abort(403, "Path traversal or symlink escape detected")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", abort(403, "Symlinks are not allowed")
		}
	}
	rel := path.Join(parts...)
	if checkExists {
		if _, statErr := u.Stat(rel); statErr != nil {
			return "", abort(404, "File or directory not found")
		}
	}
	return rel, nil
}

// Abs is the host path for rel, only for logs and display
func (u *UserRoot) Abs(rel string) string {
	return filepath.Join(u.Home, rel)
}

// WriteFileAtomic writes to a temp file next to name and renames it over, both inside the root
func (u *UserRoot) WriteFileAtomic(name string, data []byte, perm os.FileMode) error {
	tmp := name + ".tmp"
	if err := u.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := u.Rename(tmp, name); err != nil {
		_ = u.Remove(tmp)
		return err
	}
	return nil
}

// ChownTree lchowns name and everything under it without following symlinks
func ChownTree(root *os.Root, name string, uid int) error {
	return fs.WalkDir(root.FS(), name, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // best-effort, keep chowning the rest
		}
		_ = root.Lchown(p, uid, uid)
		return nil
	})
}

// RenameBetween moves src in one root to dst in another via renameat on the already opened parent dirs
func RenameBetween(srcRoot *os.Root, src string, dstRoot *os.Root, dst string) error {
	srcDir, err := srcRoot.Open(path.Dir(src))
	if err != nil {
		return err
	}
	defer srcDir.Close()
	dstDir, err := dstRoot.Open(path.Dir(dst))
	if err != nil {
		return err
	}
	defer dstDir.Close()
	if err := syscall.Renameat(int(srcDir.Fd()), path.Base(src), int(dstDir.Fd()), path.Base(dst)); err != nil {
		return &os.LinkError{Op: "renameat", Old: src, New: dst, Err: err}
	}
	return nil
}

// WriteFileFrom creates or truncates name from r, replacing a symlink sitting there instead of writing through it
func WriteFileFrom(root *os.Root, name string, r io.Reader, perm os.FileMode) error {
	if err := root.MkdirAll(path.Dir(name), 0o755); err != nil {
		return err
	}
	if info, err := root.Lstat(name); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := root.Remove(name); err != nil {
			return err
		}
	}
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// ReplaceSymlink creates a symlink at name, removing whatever non-directory is already there
func ReplaceSymlink(root *os.Root, target, name string) error {
	if err := root.MkdirAll(path.Dir(name), 0o755); err != nil {
		return err
	}
	if info, err := root.Lstat(name); err == nil && !info.IsDir() {
		if err := root.Remove(name); err != nil {
			return err
		}
	}
	return root.Symlink(target, name)
}

// CopyDirInto copies the trusted host dir src over dest under root, used to publish a staged tree into a user's dir
func CopyDirInto(root *os.Root, dest, src string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := path.Join(dest, filepath.ToSlash(rel))
		switch {
		case d.IsDir():
			return root.MkdirAll(target, 0o755)
		case d.Type()&fs.ModeSymlink != 0:
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return ReplaceSymlink(root, link, target)
		case d.Type().IsRegular():
			info, err := d.Info()
			if err != nil {
				return err
			}
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := WriteFileFrom(root, target, f, info.Mode().Perm()); err != nil {
				return err
			}
			return root.Chmod(target, info.Mode().Perm())
		}
		return nil
	})
}

// ChmodTree is chmod -R inside root, symlinks are skipped
func ChmodTree(root *os.Root, name string, mode os.FileMode) error {
	return fs.WalkDir(root.FS(), name, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		return root.Chmod(p, mode)
	})
}

// CopyContents is cp -a src/. dst/ inside root: merges into dst, keeps modes and owners, copies symlinks as links
func CopyContents(root *os.Root, src, dst string) error {
	return fs.WalkDir(root.FS(), src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := path.Join(dst, relTo(src, p))
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			if err := root.MkdirAll(target, info.Mode().Perm()); err != nil {
				return err
			}
			_ = root.Chmod(target, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			link, err := root.Readlink(p)
			if err != nil {
				return err
			}
			if err := ReplaceSymlink(root, link, target); err != nil {
				return err
			}
		case d.Type().IsRegular():
			f, err := root.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)
			if err != nil {
				return err
			}
			err = WriteFileFrom(root, target, f, info.Mode().Perm())
			_ = f.Close()
			if err != nil {
				return err
			}
			_ = root.Chmod(target, info.Mode().Perm())
		default:
			return nil
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			_ = root.Lchown(target, int(st.Uid), int(st.Gid))
		}
		return nil
	})
}
