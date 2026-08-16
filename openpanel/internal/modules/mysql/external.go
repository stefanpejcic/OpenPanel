package mysql

import (
	"context"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// DatabaseLimitReached checks the current user's plan db_limit against
// their actual database count (same source getDatabaseCount/dbLimit checks
// already use for /mysql/new and the setup wizard), exported so an app
// installer (any CMS module's install.go, which is about to CREATE
// DATABASE + CREATE USER on the user's behalf) can refuse up front instead
// of only checking the unrelated website-count limit and letting a
// database-limit failure surface later, mid-install, as a raw SQL error.
func DatabaseLimitReached(ctx context.Context, a *appctx.App, userID int, currentUsername, userContext string) bool {
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	dbLimit := 0
	if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
		dbLimit = atoiDefault(plan.DBLimit, 0)
	}
	if dbLimit == 0 {
		return false
	}
	dbUsage := getDatabaseCount(ctx, a, currentUsername, userContext)
	return dbUsage >= dbLimit
}
