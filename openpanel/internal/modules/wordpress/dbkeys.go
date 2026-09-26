package wordpress

import (
	"net/http"
	"regexp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// highPerformanceKeys are the ALTERs the Index WP MySQL For Speed plugin (1.5.x, Barracuda/DYNAMIC rows) runs on a fresh install's standard keys, unique keys come before the primary key swap so the auto-increment column always stays indexed
var highPerformanceKeys = []struct{ table, alter string }{
	{"options", "ADD UNIQUE KEY option_id (option_id), DROP PRIMARY KEY, ADD PRIMARY KEY (option_name), DROP KEY option_name"},
	{"comments", "ADD UNIQUE KEY comment_ID (comment_ID), DROP PRIMARY KEY, ADD PRIMARY KEY (comment_post_ID, comment_ID), DROP KEY comment_approved_date_gmt, ADD KEY comment_approved_date_gmt (comment_approved, comment_date_gmt, comment_ID), DROP KEY comment_date_gmt, ADD KEY comment_date_gmt (comment_date_gmt, comment_ID), DROP KEY comment_parent, ADD KEY comment_parent (comment_parent, comment_ID), DROP KEY comment_author_email, ADD KEY comment_author_email (comment_author_email, comment_post_ID, comment_ID), ADD KEY comment_post_parent_approved (comment_post_ID, comment_parent, comment_approved, comment_type, user_id, comment_date_gmt, comment_ID), DROP KEY comment_post_ID"},
	{"postmeta", "ADD UNIQUE KEY meta_id (meta_id), DROP PRIMARY KEY, ADD PRIMARY KEY (post_id, meta_key, meta_id), DROP KEY meta_key, ADD KEY meta_key (meta_key, meta_value(32), post_id, meta_id), ADD KEY meta_value (meta_value(32), meta_id), DROP KEY post_id"},
	{"usermeta", "ADD UNIQUE KEY umeta_id (umeta_id), DROP PRIMARY KEY, ADD PRIMARY KEY (user_id, meta_key, umeta_id), DROP KEY meta_key, ADD KEY meta_key (meta_key, meta_value(32), user_id, umeta_id), ADD KEY meta_value (meta_value(32), umeta_id), DROP KEY user_id"},
	{"termmeta", "ADD UNIQUE KEY meta_id (meta_id), DROP PRIMARY KEY, ADD PRIMARY KEY (term_id, meta_key, meta_id), DROP KEY meta_key, ADD KEY meta_key (meta_key, meta_value(32), term_id, meta_id), ADD KEY meta_value (meta_value(32), meta_id), DROP KEY term_id"},
	{"posts", "DROP KEY post_name, ADD KEY post_name (post_name), DROP KEY post_parent, ADD KEY post_parent (post_parent, post_type, post_status), DROP KEY type_status_date, ADD KEY type_status_date (post_type, post_status, post_date, post_author), DROP KEY post_author, ADD KEY post_author (post_author, post_type, post_status, post_date)"},
	{"commentmeta", "ADD UNIQUE KEY meta_id (meta_id), DROP PRIMARY KEY, ADD PRIMARY KEY (meta_key, comment_id, meta_id), DROP KEY comment_id, ADD KEY comment_id (comment_id, meta_key, meta_value(32)), ADD KEY meta_value (meta_value(32)), DROP KEY meta_key"},
	{"users", "ADD KEY display_name (display_name)"},
}

var tablePrefixSafeRE = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// highPerformanceKeySQL returns the ALTER statement per table for prefix, or nil when the prefix isn't safe to put in SQL
func highPerformanceKeySQL(prefix string) []string {
	if !tablePrefixSafeRE.MatchString(prefix) {
		return nil
	}
	stmts := make([]string, 0, len(highPerformanceKeys))
	for _, k := range highPerformanceKeys {
		stmts = append(stmts, "ALTER TABLE `"+prefix+k.table+"` "+k.alter)
	}
	return stmts
}

// addHighPerformanceKeys rekeys a freshly installed WordPress database, one table at a time so a failure only skips that table
func addHighPerformanceKeys(a *appctx.App, r *http.Request, userContext, dbName, prefix string, emit func(map[string]any)) {
	stmts := highPerformanceKeySQL(prefix)
	if stmts == nil {
		emit(map[string]any{"status": web.Tr(a, r, "Skipped high-performance database keys, unusual table prefix.")})
		return
	}
	emit(map[string]any{"status": web.Tr(a, r, "Adding high-performance database keys")})
	for i, stmt := range stmts {
		table := prefix + highPerformanceKeys[i].table
		if _, err := mysqlmanager.Exec(r.Context(), userContext, stmt, dbName); err != nil {
			emit(map[string]any{"status": web.Tr(a, r, "Could not add high-performance keys to %(table)s: %(error)s", "table", table, "error", err.Error())})
			continue
		}
		_, _ = mysqlmanager.Exec(r.Context(), userContext, "ANALYZE TABLE `"+table+"`", dbName)
	}
}
