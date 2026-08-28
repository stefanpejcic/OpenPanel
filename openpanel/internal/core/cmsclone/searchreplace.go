package cmsclone

import (
	"context"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
)

// SearchReplaceDatabase runs a database-wide search-replace across every
// text-like column of dbName, swapping oldStr for newStr - the generic
// equivalent of `wp search-replace`, for every CMS clone flow whose own
// CLI has no built-in equivalent (Drupal's drush dropped it years ago,
// and none of Joomla/OpenCart/Nextcloud/PrestaShop/Matomo/Moodle/
// MediaWiki/Flarum/SofaWiki's CLIs ever had one). Without this, a clone
// only copies files+DB as-is: any URL hardcoded into page/content body
// text - not just the handful of config-file constants each clone.go
// already rewrites - keeps pointing at the source domain.
//
// dbName is trusted to already be validated by ValidDB (^[a-zA-Z0-9_]+$)
// by every call site, so it's safe to interpolate directly into SQL
// rather than needing placeholder binding here.
//
// Best-effort and column-at-a-time: one column's UPDATE failing (a
// generated/virtual column, a CHECK constraint the replaced value would
// violate) doesn't abort the rest. Like wp-cli's own tool, this can't
// safely fix PHP-serialized strings whose byte-length the replacement
// changes (a serialized array's length prefix would then mismatch and
// corrupt on unserialize) - fine for the common case (plain URLs in
// content/config), but a site that serializes URLs into blobs may still
// need manual follow-up after cloning.
func SearchReplaceDatabase(ctx context.Context, userContext, dbName, oldStr, newStr string) {
	if dbName == "" || oldStr == "" || oldStr == newStr {
		return
	}

	// BLOB types included alongside TEXT ones - confirmed live that
	// Drupal's own config table (the most likely place a base URL
	// actually lives) stores its serialized data as longblob, not
	// longtext, and MySQL's REPLACE() works identically on binary
	// strings, so excluding blobs would silently skip exactly the
	// tables most likely to need this.
	rows, err := mysqlmanager.Exec(ctx, userContext, `
		SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = '`+dbName+`'
		  AND DATA_TYPE IN ('char','varchar','tinytext','text','mediumtext','longtext',
		                     'tinyblob','blob','mediumblob','longblob')`, "")
	if err != nil {
		return
	}

	escapedOld := strings.ReplaceAll(oldStr, "'", "''")
	escapedNew := strings.ReplaceAll(newStr, "'", "''")

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		table := toStringCell(row[0])
		column := toStringCell(row[1])
		if table == "" || column == "" {
			continue
		}
		safeTable := "`" + strings.ReplaceAll(table, "`", "``") + "`"
		safeColumn := "`" + strings.ReplaceAll(column, "`", "``") + "`"

		query := "UPDATE " + safeTable + " SET " + safeColumn + " = REPLACE(" + safeColumn + ", '" + escapedOld + "', '" + escapedNew + "')" +
			" WHERE " + safeColumn + " LIKE '%" + escapedOld + "%'"
		_, _ = mysqlmanager.Exec(ctx, userContext, query, dbName)
	}
}

func toStringCell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	default:
		return ""
	}
}
