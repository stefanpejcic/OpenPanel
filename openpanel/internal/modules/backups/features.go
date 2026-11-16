package backups

import (
	"context"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// targetFeatures is the feature a restore or download target needs, everything else (all, files, queue, logs) is always there
var targetFeatures = map[string][]string{
	"html":      {"filemanager"},
	"paths":     {"filemanager"},
	"mail":      {"emails"},
	"mailboxes": {"emails"},
	"vhosts":    {"domains"},
	"domains":   {"domains"},
	"dns":       {"domains", "dns"},
	"dns_zones": {"domains", "dns"},
	"crons":     {"crons"},
	"mysql":     {"mysql"},
	"postgres":  {"postgresql"},
}

// loadAllowed is swapped in tests, the real one reads the account's feature set
var loadAllowed = func(ctx context.Context, a *appctx.App, username, userContext string) map[string]bool {
	allowed := map[string]bool{}
	if a == nil {
		return allowed
	}
	features, _ := a.LoadUserFeatures(ctx, username, userContext)
	for _, f := range features {
		allowed[strings.TrimSpace(f)] = true
	}
	return allowed
}

// targetAllowed says if the account may restore or download target, a database target goes by its engine and needs either when the engine isn't known
func targetAllowed(allowed map[string]bool, target, engine string) bool {
	if target == "database" || target == "databases" {
		switch engine {
		case "mysql":
			return allowed["mysql"]
		case "postgres":
			return allowed["postgresql"]
		}
		return allowed["mysql"] || allowed["postgresql"]
	}
	need, ok := targetFeatures[target]
	if !ok {
		return true
	}
	for _, f := range need {
		if !allowed[f] {
			return false
		}
	}
	return true
}

const targetNotAllowed = "This is not available on your plan."
