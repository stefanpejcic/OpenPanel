package filemanager

import (
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

// copyFile copies file contents and permission bits inside root - it doesn't itself reject an existing destination, callers are expected to have already disambiguated dst to a non-existent path or a directory target
func copyFile(root *os.Root, src, dst string) error {
	in, err := root.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := root.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// copyTree recursively copies a directory inside root, failing if dst already exists.
func copyTree(root *os.Root, src, dst string) error {
	if _, err := root.Lstat(dst); err == nil {
		return os.ErrExist
	}

	return fs.WalkDir(root.FS(), src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := path.Join(dst, relTo(src, p))

		switch {
		case d.IsDir():
			info, err := d.Info()
			if err != nil {
				return err
			}
			return root.MkdirAll(target, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			linkTarget, err := root.Readlink(p)
			if err != nil {
				return err
			}
			return root.Symlink(linkTarget, target)
		case d.Type().IsRegular():
			return copyFile(root, p, target)
		}
		return nil
	})
}

// renameNoReplace is root.Rename but won't replace an existing dst
func renameNoReplace(root *os.Root, src, dst string) error {
	if src != dst {
		if _, err := root.Lstat(dst); err == nil {
			return os.ErrExist
		}
	}
	return root.Rename(src, dst)
}

// relTo is p relative to base for slash paths that came out of fs.WalkDir(base)
func relTo(base, p string) string {
	if p == base {
		return "."
	}
	if base == "." {
		return p
	}
	return strings.TrimPrefix(p, base+"/")
}
