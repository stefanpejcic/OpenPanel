package paths

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"syscall"
)

// MemberName cleans an archive member name to a relative path, false if it would climb out
func MemberName(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return "", false
		}
	}
	return strings.TrimPrefix(path.Clean("/"+name), "/"), true
}

// memberTarget maps a member under destRel after dropping strip leading dirs, "" when nothing is left of it
func memberTarget(destRel, name string, strip int) (string, bool) {
	clean, ok := MemberName(name)
	if !ok {
		return "", false
	}
	if strip > 0 {
		parts := strings.Split(clean, "/")
		if len(parts) <= strip {
			return "", true
		}
		clean = path.Join(parts[strip:]...)
	}
	if clean == "" {
		return destRel, true
	}
	return path.Join(destRel, clean), true
}

// CtxReader stops a long copy once the extract/compress deadline passes
func CtxReader(ctx context.Context, r io.Reader) io.Reader { return ctxReader{ctx, r} }

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

// extracted files get the archive's mode minus a 022 umask, like tar --no-same-permissions
func extractMode(mode os.FileMode, dir bool) os.FileMode {
	perm := mode.Perm() &^ 0o022
	if dir {
		return perm | 0o700
	}
	if perm == 0 {
		return 0o644
	}
	return perm
}

// ExtractTar unpacks tr under destRel through root, strip works like tar --strip-components
func ExtractTar(ctx context.Context, root *os.Root, destRel string, tr *tar.Reader, strip int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, ok := memberTarget(destRel, hdr.Name, strip)
		if !ok {
			return fmt.Errorf("unsafe path in archive: %s", hdr.Name)
		}
		if target == "" {
			continue
		}
		mode := hdr.FileInfo().Mode()
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = root.MkdirAll(target, extractMode(mode, true))
		case tar.TypeReg:
			err = WriteFileFrom(root, target, ctxReader{ctx, tr}, extractMode(mode, false))
		case tar.TypeSymlink:
			err = ReplaceSymlink(root, hdr.Linkname, target)
		case tar.TypeLink:
			src, ok := memberTarget(destRel, hdr.Linkname, strip)
			if !ok || src == "" {
				return fmt.Errorf("unsafe hard link in archive: %s", hdr.Name)
			}
			if fi, statErr := root.Lstat(target); statErr == nil && !fi.IsDir() {
				_ = root.Remove(target)
			}
			err = root.Link(src, target)
		}
		if err != nil {
			return err
		}
	}
}

// ExtractTarFile unpacks a .tar/.tar.gz the panel owns (like a cached app download) under destRel through root
func ExtractTarFile(ctx context.Context, root *os.Root, destRel, archivePath string, strip int) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return ExtractTarReader(ctx, root, destRel, f, archivePath, strip)
}

// ExtractTarReader is ExtractTar for a raw stream, gunzipping when name ends in .gz/.tgz
func ExtractTarReader(ctx context.Context, root *os.Root, destRel string, r io.Reader, name string, strip int) error {
	if strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".tgz") {
		gz, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gz.Close()
		r = gz
	}
	return ExtractTar(ctx, root, destRel, tar.NewReader(r), strip)
}

// ExtractZip unpacks zr under destRel through root, strip works like tar --strip-components
func ExtractZip(ctx context.Context, root *os.Root, destRel string, zr *zip.Reader, strip int) error {
	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		target, ok := memberTarget(destRel, f.Name, strip)
		if !ok {
			return fmt.Errorf("unsafe path in archive: %s", f.Name)
		}
		if target == "" {
			continue
		}
		mode := f.Mode()
		switch {
		case mode.IsDir() || strings.HasSuffix(f.Name, "/"):
			if err := root.MkdirAll(target, extractMode(mode, true)); err != nil {
				return err
			}
		case mode&os.ModeSymlink != 0:
			rc, err := f.Open()
			if err != nil {
				return err
			}
			link, err := io.ReadAll(io.LimitReader(rc, 4096))
			_ = rc.Close()
			if err != nil {
				return err
			}
			if err := ReplaceSymlink(root, string(link), target); err != nil {
				return err
			}
		case mode.IsRegular():
			rc, err := f.Open()
			if err != nil {
				return err
			}
			err = WriteFileFrom(root, target, ctxReader{ctx, rc}, extractMode(mode, false))
			_ = rc.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ExtractZipFile is ExtractZip for a zip the panel owns
func ExtractZipFile(ctx context.Context, root *os.Root, destRel, archivePath string, strip int) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()
	return ExtractZip(ctx, root, destRel, &zr.Reader, strip)
}

// archiveWriter is the bit of tar/zip that WriteArchive needs
type archiveWriter interface {
	add(name string, info fs.FileInfo, link string, body io.Reader) error
	Close() error
}

type tarArchive struct {
	tw *tar.Writer
	gz *gzip.Writer
}

func (t *tarArchive) add(name string, info fs.FileInfo, link string, body io.Reader) error {
	hdr, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	hdr.Name = name
	if info.IsDir() {
		hdr.Name += "/"
	}
	if err := t.tw.WriteHeader(hdr); err != nil {
		return err
	}
	if body != nil {
		if _, err := io.CopyN(t.tw, body, hdr.Size); err != nil {
			return err
		}
	}
	return nil
}

func (t *tarArchive) Close() error {
	err := t.tw.Close()
	if t.gz != nil {
		if gzErr := t.gz.Close(); err == nil {
			err = gzErr
		}
	}
	return err
}

type zipArchive struct{ zw *zip.Writer }

func (z *zipArchive) add(name string, info fs.FileInfo, link string, body io.Reader) error {
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	switch {
	case info.IsDir():
		hdr.Name += "/"
	case info.Mode().IsRegular():
		hdr.Method = zip.Deflate
	}
	w, err := z.zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	if link != "" {
		_, err = io.WriteString(w, link)
		return err
	}
	if body != nil {
		_, err = io.CopyN(w, body, info.Size())
	}
	return err
}

func (z *zipArchive) Close() error { return z.zw.Close() }

// WriteArchive writes names (relative to dirRel) as a zip/tar/tar.gz to out, reading only through root and storing symlinks as links
func WriteArchive(ctx context.Context, root *os.Root, ext, dirRel string, names []string, out io.Writer) error {
	var aw archiveWriter
	switch ext {
	case "zip":
		aw = &zipArchive{zw: zip.NewWriter(out)}
	case "tar":
		aw = &tarArchive{tw: tar.NewWriter(out)}
	case "tar.gz":
		gz := gzip.NewWriter(out)
		aw = &tarArchive{tw: tar.NewWriter(gz), gz: gz}
	default:
		return fmt.Errorf("unsupported archive format %q", ext)
	}

	// skip the archive itself when it's being written inside the tree
	var outInfo fs.FileInfo
	if f, ok := out.(*os.File); ok {
		outInfo, _ = f.Stat()
	}

	addEntry := func(p string, info fs.FileInfo) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := relTo(dirRel, p)
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			link, err := root.Readlink(p)
			if err != nil {
				return err
			}
			return aw.add(name, info, link, nil)
		case info.IsDir():
			if name == "." {
				return nil
			}
			return aw.add(name, info, "", nil)
		case info.Mode().IsRegular():
			// nonblock so a file swapped for a fifo can't hang us, and re-check the type on the open handle
			f, err := root.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)
			if err != nil {
				return err
			}
			defer f.Close()
			fi, err := f.Stat()
			if err != nil || !fi.Mode().IsRegular() || (outInfo != nil && os.SameFile(fi, outInfo)) {
				return err
			}
			return aw.add(name, fi, "", ctxReader{ctx, f})
		}
		return nil
	}

	for _, n := range names {
		start := path.Join(dirRel, n)
		info, err := root.Lstat(start)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := addEntry(start, info); err != nil {
				return err
			}
			continue
		}
		if err := fs.WalkDir(root.FS(), start, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			fi, err := d.Info()
			if err != nil {
				return err
			}
			return addEntry(p, fi)
		}); err != nil {
			return err
		}
	}
	return aw.Close()
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
