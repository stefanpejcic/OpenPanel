package joomla

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`\$db\s*=\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`\$user\s*=\s*'([^']*)'`)
)
