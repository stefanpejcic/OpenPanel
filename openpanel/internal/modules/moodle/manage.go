package moodle

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`CFG->dbuser\s*=\s*'([^']*)'`)
)
