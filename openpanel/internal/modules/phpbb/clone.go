package phpbb

import (
	"os/exec"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors flarum/clone.go's shape (file copy, DB create+dump, config rewrite, sites insert) - config.php's dbname/dbuser/dbpasswd need rewriting, but phpBB derives its base URL from $_SERVER at runtime like Drupal rather than storing it, so only SearchReplaceDatabase handles that side, same limitation drupal/clone.go documents

var (
	clonePhpbbDatabaseRE = regexp.MustCompile(`\$dbname\s*=\s*'.*?';`)
	clonePhpbbUsernameRE = regexp.MustCompile(`\$dbuser\s*=\s*'.*?';`)
	clonePhpbbPasswordRE = regexp.MustCompile(`\$dbpasswd\s*=\s*'.*?';`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	_ = exec.CommandContext(c.Ctx, "chmod", "644", filepath.Join(c.DstPath, "config.php")).Run()
	return c.RewriteConfig("config.php", func(s string) string {
		s = clonePhpbbDatabaseRE.ReplaceAllString(s, "$$dbname = '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"';")
		s = clonePhpbbUsernameRE.ReplaceAllString(s, "$$dbuser = '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"';")
		return clonePhpbbPasswordRE.ReplaceAllString(s, "$$dbpasswd = '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"';")
	}, "Failed to set 'dbname' in config.php")
}
