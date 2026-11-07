// Package usersocket dials a unix socket inside an account's home without following anything the account put there.
package usersocket

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// HomeRoot is swapped in tests
var HomeRoot = "/home/"

var ErrNotSocket = errors.New("account socket path is not a plain socket")

// open walks /home/<ctx>/<rel> one component at a time with O_NOFOLLOW and returns an O_PATH fd of the socket inode
func open(userContext, rel string) (int, error) {
	if userContext == "" || strings.ContainsAny(userContext, "/\x00") || userContext == "." || userContext == ".." {
		return -1, fmt.Errorf("bad account context %q", userContext)
	}
	cur, err := unix.Open(HomeRoot+userContext, unix.O_PATH|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	parts := strings.Split(rel, "/")
	for i, p := range parts {
		if p == "" || p == "." || p == ".." {
			unix.Close(cur)
			return -1, ErrNotSocket
		}
		flags := unix.O_PATH | unix.O_NOFOLLOW | unix.O_CLOEXEC
		if i < len(parts)-1 {
			flags |= unix.O_DIRECTORY
		}
		next, err := unix.Openat(cur, p, flags, 0)
		unix.Close(cur)
		if err != nil {
			return -1, err
		}
		cur = next
	}
	var st unix.Stat_t
	if err := unix.Fstat(cur, &st); err != nil || st.Mode&unix.S_IFMT != unix.S_IFSOCK {
		unix.Close(cur)
		return -1, ErrNotSocket
	}
	return cur, nil
}

// Exists reports whether a real socket sits at /home/<ctx>/<rel>
func Exists(userContext, rel string) bool {
	fd, err := open(userContext, rel)
	if err != nil {
		return false
	}
	unix.Close(fd)
	return true
}

// Dial connects to /home/<ctx>/<rel> through the checked fd, so a swap after the check can't redirect it
func Dial(ctx context.Context, userContext, rel string) (net.Conn, error) {
	fd, err := open(userContext, rel)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	var d net.Dialer
	return d.DialContext(ctx, "unix", "/proc/self/fd/"+strconv.Itoa(fd))
}
