// Package tenantfs mirrors the os file calls for host paths, anything inside an account's podman volume goes through an os.Root on that volume so a symlink or .. planted by the account can't send a root-run op outside it. Paths outside volumes pass straight through to os.
package tenantfs

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
)

// homeRoot is where account homes live, tests point it at a temp dir
var homeRoot = "/home/"

// ErrEscape is returned when a host path climbs out of the volume it starts in
var ErrEscape = errors.New("path escapes the account volume")

// volumeDir reports whether clean is exactly a volume's data dir, /home/<ctx>/docker-data/volumes/<ctx>_<vol>/_data
func volumeDir(clean string) bool {
	parts := strings.Split(strings.TrimPrefix(clean, homeRoot), "/")
	return strings.HasPrefix(clean, homeRoot) && len(parts) == 5 && parts[0] != "" && parts[1] == "docker-data" &&
		parts[2] == "volumes" && strings.HasPrefix(parts[3], parts[0]+"_") && parts[4] == "_data"
}

// volumeOf finds the volume p is in, going by the path as written so a .. can't step back out of it unnoticed
func volumeOf(p string) (dir, rel string, ok bool, err error) {
	if !filepath.IsAbs(p) {
		return "", "", false, nil
	}
	raw := strings.Split(p, "/")
	for i := 1; i <= len(raw); i++ {
		if prefix := filepath.Clean("/" + strings.Join(raw[:i], "/")); volumeDir(prefix) {
			dir = prefix
			break
		}
	}
	clean := filepath.Clean(p)
	if dir == "" {
		// written some other way but still lands in a volume
		if rest, found := strings.CutPrefix(clean, homeRoot); found {
			parts := strings.SplitN(rest, "/", 6)
			if len(parts) >= 5 && volumeDir(homeRoot+strings.Join(parts[:5], "/")) {
				dir = homeRoot + strings.Join(parts[:5], "/")
			}
		}
	}
	if dir == "" {
		return "", "", false, nil
	}
	rel, relErr := filepath.Rel(dir, clean)
	if relErr != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", "", false, &fs.PathError{Op: "open", Path: p, Err: ErrEscape}
	}
	return dir, rel, true, nil
}

// open resolves p, root is nil when p is outside every volume
func open(p string) (*os.Root, string, error) {
	dir, rel, ok, err := volumeOf(p)
	if err != nil || !ok {
		return nil, "", err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, "", err
	}
	return root, rel, nil
}

// IsVolumePath reports whether p is inside an account volume
func IsVolumePath(p string) bool {
	_, _, ok, err := volumeOf(p)
	return ok || err != nil
}

// do routes an op three ways: through the volume's os.Root, through the strict home resolver for a
// path under an account home, or straight to os for everything else (system paths like /etc/openpanel).
func do(p string, inRoot func(*os.Root, string) error, inHome func(base, rel string) error, plain func() error) error {
	root, rel, err := open(p)
	if err != nil {
		return err
	}
	if root != nil {
		defer root.Close()
		return inRoot(root, rel)
	}
	if base, hrel, ok := homeSplit(p); ok {
		return inHome(base, hrel)
	}
	return plain()
}

func ReadFile(p string) ([]byte, error) {
	var data []byte
	err := do(p, func(r *os.Root, rel string) (err error) { data, err = r.ReadFile(rel); return },
		func(base, rel string) (err error) { data, err = homeReadFile(base, rel); return },
		func() (err error) { data, err = os.ReadFile(p); return })
	return data, err
}

func WriteFile(p string, data []byte, perm os.FileMode) error {
	return do(p, func(r *os.Root, rel string) error { return r.WriteFile(rel, data, perm) },
		func(base, rel string) error { return homeWriteFile(base, rel, data, perm) },
		func() error { return os.WriteFile(p, data, perm) })
}

func MkdirAll(p string, perm os.FileMode) error {
	return do(p, func(r *os.Root, rel string) error { return r.MkdirAll(rel, perm) },
		func(base, rel string) error { return homeMkdirAll(base, rel, perm) },
		func() error { return os.MkdirAll(p, perm) })
}

func Mkdir(p string, perm os.FileMode) error {
	return do(p, func(r *os.Root, rel string) error { return r.Mkdir(rel, perm) },
		func(base, rel string) error { return homeMkdir(base, rel, perm) },
		func() error { return os.Mkdir(p, perm) })
}

func Remove(p string) error {
	return do(p, func(r *os.Root, rel string) error { return r.Remove(rel) },
		func(base, rel string) error { return homeRemove(base, rel) },
		func() error { return os.Remove(p) })
}

func RemoveAll(p string) error {
	return do(p, func(r *os.Root, rel string) error {
		if rel == "." {
			return &fs.PathError{Op: "removeall", Path: p, Err: ErrEscape}
		}
		return r.RemoveAll(rel)
	},
		func(base, rel string) error { return homeRemoveAll(base, rel) },
		func() error { return os.RemoveAll(p) })
}

func Stat(p string) (fs.FileInfo, error) {
	var fi fs.FileInfo
	err := do(p, func(r *os.Root, rel string) (err error) { fi, err = r.Stat(rel); return },
		func(base, rel string) (err error) { fi, err = homeStat(base, rel); return },
		func() (err error) { fi, err = os.Stat(p); return })
	return fi, err
}

func Lstat(p string) (fs.FileInfo, error) {
	var fi fs.FileInfo
	err := do(p, func(r *os.Root, rel string) (err error) { fi, err = r.Lstat(rel); return },
		func(base, rel string) (err error) { fi, err = homeLstat(base, rel); return },
		func() (err error) { fi, err = os.Lstat(p); return })
	return fi, err
}

func Chmod(p string, mode os.FileMode) error {
	return do(p, func(r *os.Root, rel string) error { return r.Chmod(rel, mode) },
		func(base, rel string) error { return homeChmod(base, rel, mode) },
		func() error { return os.Chmod(p, mode) })
}

// Chown never follows a final symlink on a home path; in a volume it may follow while staying inside
func Chown(p string, uid, gid int) error {
	return do(p, func(r *os.Root, rel string) error { return r.Chown(rel, uid, gid) },
		func(base, rel string) error { return homeChown(base, rel, uid, gid) },
		func() error { return os.Chown(p, uid, gid) })
}

func Lchown(p string, uid, gid int) error {
	return do(p, func(r *os.Root, rel string) error { return r.Lchown(rel, uid, gid) },
		func(base, rel string) error { return homeChown(base, rel, uid, gid) },
		func() error { return os.Lchown(p, uid, gid) })
}

func Symlink(target, p string) error {
	return do(p, func(r *os.Root, rel string) error { return r.Symlink(target, rel) },
		func(base, rel string) error { return homeSymlink(target, base, rel) },
		func() error { return os.Symlink(target, p) })
}

func Readlink(p string) (string, error) {
	var s string
	err := do(p, func(r *os.Root, rel string) (err error) { s, err = r.Readlink(rel); return },
		func(base, rel string) (err error) { s, err = homeReadlink(base, rel); return },
		func() (err error) { s, err = os.Readlink(p); return })
	return s, err
}

func ReadDir(p string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	err := do(p, func(r *os.Root, rel string) (err error) { entries, err = fs.ReadDir(r.FS(), rel); return },
		func(base, rel string) (err error) { entries, err = homeReadDir(base, rel); return },
		func() (err error) { entries, err = os.ReadDir(p); return })
	return entries, err
}

// OpenFile hands back a plain file, it stays valid after the volume root is closed
func OpenFile(p string, flag int, perm os.FileMode) (*os.File, error) {
	root, rel, err := open(p)
	if err != nil {
		return nil, err
	}
	if root != nil {
		defer root.Close()
		return root.OpenFile(rel, flag, perm)
	}
	if base, hrel, ok := homeSplit(p); ok {
		return openHomeFile(base, hrel, flag, perm)
	}
	return os.OpenFile(p, flag, perm)
}

func Open(p string) (*os.File, error) { return OpenFile(p, os.O_RDONLY, 0) }

func Create(p string) (*os.File, error) {
	return OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

// Rename handles both ends in volumes, in home dirs, in different ones or outside any
func Rename(oldPath, newPath string) error {
	if !isVolumeOrHome(oldPath) && !isVolumeOrHome(newPath) {
		return os.Rename(oldPath, newPath)
	}
	oldDir, oldName, err := renameEnd(oldPath)
	if err != nil {
		return err
	}
	defer oldDir.Close()
	newDir, newName, err := renameEnd(newPath)
	if err != nil {
		return err
	}
	defer newDir.Close()
	if err := syscall.Renameat(int(oldDir.Fd()), oldName, int(newDir.Fd()), newName); err != nil {
		return &os.LinkError{Op: "rename", Old: oldPath, New: newPath, Err: err}
	}
	return nil
}

// isVolumeOrHome reports whether p is inside a volume or directly under an account home
func isVolumeOrHome(p string) bool {
	if IsVolumePath(p) {
		return true
	}
	_, _, ok := homeSplit(p)
	return ok
}

// renameEnd opens the parent dir of p for a renameat, going through the volume root, the strict home
// resolver, or plain os. name is the final component. The caller closes the returned dir.
func renameEnd(p string) (dir *os.File, name string, err error) {
	root, rel, err := open(p)
	if err != nil {
		return nil, "", err
	}
	if root != nil {
		defer root.Close()
		if rel == "." {
			return nil, "", &fs.PathError{Op: "rename", Path: p, Err: ErrEscape}
		}
		d, derr := root.Open(path.Dir(rel))
		if derr != nil {
			return nil, "", derr
		}
		return d, filepath.Base(p), nil
	}
	if base, hrel, ok := homeSplit(p); ok {
		if hrel == "" || hrel == "." {
			return nil, "", &fs.PathError{Op: "rename", Path: p, Err: ErrEscape}
		}
		pfd, hname, herr := resolveHomeParent(base, hrel)
		if herr != nil {
			return nil, "", herr
		}
		return os.NewFile(uintptr(pfd), base+"/"+hrel), hname, nil
	}
	d, derr := os.Open(filepath.Dir(p))
	if derr != nil {
		return nil, "", derr
	}
	return d, filepath.Base(p), nil
}

// WalkDir is filepath.WalkDir, inside a volume symlinks are never followed out of it
func WalkDir(p string, fn fs.WalkDirFunc) error {
	root, rel, err := open(p)
	if err != nil {
		return err
	}
	if root == nil {
		return filepath.WalkDir(p, fn)
	}
	defer root.Close()
	base := filepath.Clean(p)
	return fs.WalkDir(root.FS(), rel, func(q string, d fs.DirEntry, err error) error {
		host := base
		if q != rel {
			host = filepath.Join(base, relTo(rel, q))
		}
		return fn(host, d, err)
	})
}

// ChownR is chown -R uid:gid p without following symlinks
func ChownR(p string, uid, gid int) error {
	return WalkDir(p, func(q string, _ fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // best-effort like chown -R, keep going
		}
		_ = Lchown(q, uid, gid)
		return nil
	})
}

// ChmodR is chmod -R mode p, symlinks are skipped
func ChmodR(p string, mode os.FileMode) error {
	return WalkDir(p, func(q string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // best-effort like chmod -R, keep going
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		return Chmod(q, mode)
	})
}

// CopyFile copies one regular file's contents, dst gets perm when it's created
func CopyFile(src, dst string, perm os.FileMode) error {
	in, err := OpenFile(src, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	defer in.Close()
	if fi, err := in.Stat(); err != nil || !fi.Mode().IsRegular() {
		if err == nil {
			err = &fs.PathError{Op: "copy", Path: src, Err: syscall.EINVAL}
		}
		return err
	}
	out, err := OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// CopyDir is cp -a src/. dst/: merges into dst, keeps modes and owners, copies symlinks as links
func CopyDir(src, dst string) error {
	return WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.Clean(src), p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			if err := MkdirAll(target, info.Mode().Perm()); err != nil {
				return err
			}
			_ = Chmod(target, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			link, err := Readlink(p)
			if err != nil {
				return err
			}
			if fi, err := Lstat(target); err == nil && !fi.IsDir() {
				_ = Remove(target)
			}
			if err := Symlink(link, target); err != nil {
				return err
			}
		case d.Type().IsRegular():
			if fi, err := Lstat(target); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
				_ = Remove(target)
			}
			if err := CopyFile(p, target, info.Mode().Perm()); err != nil {
				return err
			}
			_ = Chmod(target, info.Mode().Perm())
		default:
			return nil
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			_ = Lchown(target, int(st.Uid), int(st.Gid))
		}
		return nil
	})
}

// ExtractTar unpacks a .tar/.tar.gz at archivePath into dest, like tar -xf archive -C dest --strip-components=strip
func ExtractTar(ctx context.Context, archivePath, dest string, strip int) error {
	f, err := Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return withDest(dest, func(root *os.Root, rel string) error {
		return paths.ExtractTarReader(ctx, root, rel, f, archivePath, strip)
	})
}

// ExtractZip unpacks a zip at archivePath into dest, like unzip -o archive -d dest
func ExtractZip(ctx context.Context, archivePath, dest string, strip int) error {
	f, err := Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return err
	}
	return withDest(dest, func(root *os.Root, rel string) error {
		return paths.ExtractZip(ctx, root, rel, zr, strip)
	})
}

// WriteTarGz is tar -czf out -C dir . reading dir only through its volume root
func WriteTarGz(ctx context.Context, dir, out string) (err error) {
	w, err := Create(out)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); err == nil {
			err = cerr
		}
	}()
	return withDest(dir, func(root *os.Root, rel string) error {
		return paths.WriteArchive(ctx, root, "tar.gz", rel, []string{"."}, w)
	})
}

// withDest gives fn a root to work under for dest, outside volumes that's a root on dest itself
func withDest(dest string, fn func(*os.Root, string) error) error {
	root, rel, err := open(dest)
	if err != nil {
		return err
	}
	if root == nil {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		if root, err = os.OpenRoot(dest); err != nil {
			return err
		}
		rel = "."
	}
	defer root.Close()
	if err := root.MkdirAll(rel, 0o755); err != nil {
		return err
	}
	return fn(root, rel)
}

func relTo(base, p string) string {
	if base == "." {
		return p
	}
	return strings.TrimPrefix(p, base+"/")
}

// ExtractZipSubdir unpacks only what's under prefix/ in the zip straight into dest, like unzip "prefix/*" and moving prefix's contents up
func ExtractZipSubdir(ctx context.Context, archivePath, prefix, dest string) error {
	f, err := Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return err
	}
	prefix = strings.Trim(prefix, "/")
	sub := &zip.Reader{}
	for _, zf := range zr.File {
		if name, ok := paths.MemberName(zf.Name); ok && strings.HasPrefix(name+"/", prefix+"/") {
			sub.File = append(sub.File, zf)
		}
	}
	if len(sub.File) == 0 {
		return &fs.PathError{Op: "unzip", Path: archivePath, Err: fs.ErrNotExist}
	}
	return withDest(dest, func(root *os.Root, rel string) error {
		return paths.ExtractZip(ctx, root, rel, sub, strings.Count(prefix, "/")+1)
	})
}

// ExtractNestedZip unpacks the zip stored as member inside the outer zip into dest, the inner one is staged in the panel's own temp dir
func ExtractNestedZip(ctx context.Context, archivePath, member, dest string) error {
	outer, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer outer.Close()
	inner, err := outer.Open(member)
	if err != nil {
		return err
	}
	defer inner.Close()
	tmp, err := os.CreateTemp("", "nested-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, copyErr := io.Copy(tmp, paths.CtxReader(ctx, inner))
	if closeErr := tmp.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		return copyErr
	}
	return ExtractZip(ctx, tmp.Name(), dest, 0)
}

// ReplaceContents moves every top-level entry of src over the same name in dst, leaving the keep names in dst alone, like the app updaters' rm -rf + mv loop
func ReplaceContents(src, dst string, keep ...string) error {
	entries, err := ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if slices.Contains(keep, e.Name()) {
			continue
		}
		target := filepath.Join(dst, e.Name())
		if err := RemoveAll(target); err != nil {
			return err
		}
		if err := Rename(filepath.Join(src, e.Name()), target); err != nil {
			return err
		}
	}
	return nil
}

// OutputIn runs name args... with its working dir set to dir, which is opened through a root on base first so a symlink on the way can't move it outside base
func OutputIn(ctx context.Context, base, dir, name string, args ...string) ([]byte, error) {
	rel, err := filepath.Rel(filepath.Clean(base), filepath.Clean(dir))
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return nil, &fs.PathError{Op: "chdir", Path: dir, Err: ErrEscape}
	}
	root, err := os.OpenRoot(base)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	d, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	if info, err := d.Stat(); err != nil || !info.IsDir() {
		return nil, &fs.PathError{Op: "chdir", Path: dir, Err: syscall.ENOTDIR}
	}
	// the cd happens after exec, once the opened dir is fd 3 in the child
	cmd := exec.CommandContext(ctx, "sh", append([]string{"-c", `cd /proc/self/fd/3 && exec "$@"`, "sh", name}, args...)...)
	cmd.ExtraFiles = []*os.File{d}
	return cmd.Output()
}
