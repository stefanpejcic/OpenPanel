package prestashop

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`'database_name'\s*=>\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`'database_user'\s*=>\s*'([^']*)'`)
)
