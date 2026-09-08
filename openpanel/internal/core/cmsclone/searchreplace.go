package cmsclone

import (
	"context"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
)

// SearchReplaceDatabase swaps oldStr for newStr across every text/blob column of dbName - the generic wp-cli search-replace equivalent since none of the other CMS CLIs have one built in; dbName is already validated by ValidDB so it's safe to interpolate directly, best-effort per column, and can't safely fix PHP-serialized strings since replacing changes their byte length
func SearchReplaceDatabase(ctx context.Context, userContext, dbName, oldStr, newStr string) {
	if dbName == "" || oldStr == "" || oldStr == newStr {
		return
	}

	// BLOB included alongside TEXT since Drupal stores its config table (the likely home of a base URL) as longblob, and REPLACE() works fine on binary strings too
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
