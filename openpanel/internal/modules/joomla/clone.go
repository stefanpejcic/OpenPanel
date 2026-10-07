package joomla

import (
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors wordpress/manage.go's handleCloneWordPress in shape (site-limit check, file copy, DB create+dump-pipe, config rewrite, sites-table insert) via internal/core/cmsclone, but differs from WordPress in that Joomla's configuration.php has no hardcoded site URL (derived from the request at runtime), so there's no wp-cli search-replace step - hardcoded URLs in article/module content still point at the source domain after cloning, same as a manual domain move on a stock Joomla site

var (
	cloneJoomlaUserRE     = regexp.MustCompile(`\$user\s*=\s*'.*?';`)
	cloneJoomlaPasswordRE = regexp.MustCompile(`\$password\s*=\s*'.*?';`)
	cloneJoomlaDBRE       = regexp.MustCompile(`\$db\s*=\s*'.*?';`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	// "$$" is a literal "$" in ReplaceAllString, needed for configuration.php's "public $propertyName = ...;" syntax
	return c.RewriteConfig("configuration.php", func(s string) string {
		s = cloneJoomlaUserRE.ReplaceAllString(s, "$$user = '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"';")
		s = cloneJoomlaPasswordRE.ReplaceAllString(s, "$$password = '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"';")
		return cloneJoomlaDBRE.ReplaceAllString(s, "$$db = '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"';")
	}, "Failed to set $db in configuration.php")
}
