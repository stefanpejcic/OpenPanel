package nextcloud

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`'dbname'\s*=>\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`'dbuser'\s*=>\s*'([^']*)'`)
)
