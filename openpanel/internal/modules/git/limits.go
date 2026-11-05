package git

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// git data is written as root and so never counts against the account's disk quota, these caps bound it instead
var (
	gitStorageLimit int64 = 2048 << 20
	gitMaxRepos           = 50
)

var errGitStorageFull = errors.New("the git storage limit for this account is reached")

// loadGitLimits reads git_max_storage_mb and git_max_repos from the panel config, 0 means no limit
func loadGitLimits(a *appctx.App) {
	if mb, err := strconv.ParseInt(a.Config.Get("git_max_storage_mb", "2048"), 10, 64); err == nil && mb >= 0 {
		gitStorageLimit = mb << 20
	}
	if n, err := strconv.Atoi(a.Config.Get("git_max_repos", "50")); err == nil && n >= 0 {
		gitMaxRepos = n
	}
}

// gitUsage sums the account's git dir, WalkDir doesn't follow symlinks
func gitUsage(userContext string) int64 {
	var total int64
	_ = filepath.WalkDir(accountRoot(userContext), func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// gitRoomLeft is how many bytes the account can still add, -1 means unlimited
func gitRoomLeft(userContext string) int64 {
	if gitStorageLimit == 0 {
		return -1
	}
	left := gitStorageLimit - gitUsage(userContext)
	if left < 0 {
		return 0
	}
	return left
}
