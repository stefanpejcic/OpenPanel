package opencart

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`DB_DATABASE'\s*,\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`DB_USERNAME'\s*,\s*'([^']*)'`)
)
