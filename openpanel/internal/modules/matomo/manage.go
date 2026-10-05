package matomo

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`(?m)^dbname\s*=\s*"([^"]*)"`)
	removeDBUserRE = regexp.MustCompile(`(?m)^username\s*=\s*"([^"]*)"`)
)
