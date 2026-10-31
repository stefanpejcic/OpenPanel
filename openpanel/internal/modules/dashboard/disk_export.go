package dashboard

import (
	"context"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// DiskAndInodesForUser is the dashboard's quota reading (disk in KB), nil when the quota report has no row for username
func DiskAndInodesForUser(a *appctx.App, ctx context.Context, username string) (*DiskInodes, error) {
	return getDiskAndInodesForUser(a, ctx, username)
}
