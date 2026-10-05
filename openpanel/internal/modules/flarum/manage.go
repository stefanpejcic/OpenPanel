package flarum

import (
	"regexp"
)

var (
	removeDBNameRE = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`'username'\s*=>\s*'([^']*)'`)
)
