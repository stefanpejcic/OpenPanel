package ftp

import (
	"context"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ftpLimitReached is the plan's ftp_limit against the accounts the user already has, 0 means unlimited
func ftpLimitReached(ctx context.Context, a *appctx.App, userID int, userContext string) bool {
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	limit := 0
	if plan, err := a.QueryPlanDetailsByID(ctx, planID); err == nil {
		limit = web.AtoiDefault(plan.FTPLimit, 0)
	}
	if limit == 0 {
		return false
	}
	return len(loadAccounts("/etc/openpanel/ftp/users/"+userContext+"/users.list")) >= limit
}

const ftpLimitMessage = "FTP account limit reached for your hosting plan."
