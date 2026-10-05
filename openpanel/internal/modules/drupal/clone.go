package drupal

import (
	"os/exec"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors wordpress/manage.go's handleCloneWordPress in shape (site-limit check, file copy, DB create+dump-pipe, config rewrite, sites-table insert) via internal/core/cmsclone, but differs from WordPress in two ways: Drupal's base URL is request-derived with no trusted_host_patterns set, so there's no host-allowlist to update and no wp-cli-style DB search-replace - hardcoded URLs in node content still point at the source domain after cloning, same known limitation as joomla/clone.go; and cmsclone.ValidDocroot accepts the real "/var/www/html/..." absolute-path form .Docroot actually uses, unlike WordPress's own validateDocroot() which rejects a leading "/"

var (
	cloneDrupalDatabaseRE = regexp.MustCompile(`'database'\s*=>\s*'.*?',`)
	cloneDrupalUsernameRE = regexp.MustCompile(`'username'\s*=>\s*'.*?',`)
	cloneDrupalPasswordRE = regexp.MustCompile(`'password'\s*=>\s*'.*?',`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	_ = exec.CommandContext(c.Ctx, "chmod", "644", filepath.Join(c.DstPath, "sites", "default", "settings.php")).Run()
	return c.RewriteConfig("sites/default/settings.php", func(s string) string {
		s = cloneDrupalDatabaseRE.ReplaceAllString(s, "'database' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"',")
		s = cloneDrupalUsernameRE.ReplaceAllString(s, "'username' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"',")
		return cloneDrupalPasswordRE.ReplaceAllString(s, "'password' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"',")
	}, "Failed to set 'database' in settings.php")
}
