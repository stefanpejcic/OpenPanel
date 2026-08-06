package filemanager

import (
	"io"
	"os"
	"path/filepath"
)

// copyFile copies file contents and permission bits. It does not itself
// reject an existing destination - callers are expected to have already
// disambiguated dst to a non-existent path or a directory target;
// os.IsExist below is only reachable via the O_EXCL open, kept as defense
// in depth.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// copyTree recursively copies a directory, failing if dst already exists.
func copyTree(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return os.ErrExist
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		}
		return copyFile(path, target)
	})
}
