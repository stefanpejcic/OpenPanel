package mediawiki

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`\$wgDBname\s*=\s*"([^"]*)"`)
	removeDBUserRE = regexp.MustCompile(`\$wgDBuser\s*=\s*"([^"]*)"`)
)
