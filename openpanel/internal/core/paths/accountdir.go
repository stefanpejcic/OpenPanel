package paths

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// ErrAccountSymlink is returned when a dir under an account home is reached through a symlink the account made
var ErrAccountSymlink = errors.New("path goes through a symlink the account owns")

// OpenAccountDir opens dir, which sits under the account home base, without following any symlink the
// account could have made on the way. The account owns every component below base, so a link it swapped
// in would otherwise point the root panel at another account or a host dir. Links owned by root, like
// an admin moving docker-data to another disk, are still followed.
func OpenAccountDir(base, dir string) (*os.Root, error) {
	rest, ok := strings.CutPrefix(dir, strings.TrimSuffix(base, "/")+"/")
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: dir, Err: errors.New("not under the account home")}
	}
	cur, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: base, Err: err}
	}
	for _, comp := range strings.Split(rest, "/") {
		if comp == "" || comp == "." {
			continue
		}
		if comp == ".." {
			unix.Close(cur)
			return nil, &fs.PathError{Op: "open", Path: dir, Err: errors.New("path climbs out of the account home")}
		}
		next, oerr := unix.Openat(cur, comp, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if oerr == unix.ELOOP || oerr == unix.ENOTDIR {
			var st unix.Stat_t
			if serr := unix.Fstatat(cur, comp, &st, unix.AT_SYMLINK_NOFOLLOW); serr == nil && st.Mode&unix.S_IFMT == unix.S_IFLNK && st.Uid == 0 {
				next, oerr = unix.Openat(cur, comp, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
			} else if serr == nil && st.Mode&unix.S_IFMT == unix.S_IFLNK {
				oerr = ErrAccountSymlink
			}
		}
		unix.Close(cur)
		if oerr != nil {
			return nil, &fs.PathError{Op: "open", Path: dir, Err: oerr}
		}
		cur = next
	}
	defer unix.Close(cur)
	var walked unix.Stat_t
	if err := unix.Fstat(cur, &walked); err != nil {
		return nil, &fs.PathError{Op: "open", Path: dir, Err: err}
	}
	// open by name so Root.Name() stays the real path, then make sure it's the dir we walked to and nothing was swapped in between
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	f, err := root.Open(".")
	if err != nil {
		root.Close()
		return nil, err
	}
	defer f.Close()
	var opened unix.Stat_t
	if err := unix.Fstat(int(f.Fd()), &opened); err != nil || opened.Dev != walked.Dev || opened.Ino != walked.Ino {
		root.Close()
		return nil, &fs.PathError{Op: "open", Path: dir, Err: ErrAccountSymlink}
	}
	return root, nil
}
