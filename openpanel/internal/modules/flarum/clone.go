package flarum

import (
	"os/exec"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors drupal/clone.go's shape (site-limit check, file copy, DB create+dump-pipe, config rewrite, sites-table insert) via internal/core/cmsclone - unlike Drupal's request-derived URL, Flarum's config.php stores its base URL explicitly, so a clone must also rewrite the 'url' key or the cloned forum keeps pointing at the source domain, handled below

var (
	cloneFlarumDatabaseRE = regexp.MustCompile(`'database'\s*=>\s*'.*?',`)
	cloneFlarumUsernameRE = regexp.MustCompile(`'username'\s*=>\s*'.*?',`)
	cloneFlarumPasswordRE = regexp.MustCompile(`'password'\s*=>\s*'.*?',`)
	cloneFlarumURLRE      = regexp.MustCompile(`'url'\s*=>\s*'.*?',`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	_ = exec.CommandContext(c.Ctx, "chmod", "644", filepath.Join(c.DstPath, "config.php")).Run()
	return c.RewriteConfig("config.php", func(s string) string {
		s = cloneFlarumDatabaseRE.ReplaceAllString(s, "'database' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"',")
		s = cloneFlarumUsernameRE.ReplaceAllString(s, "'username' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"',")
		s = cloneFlarumPasswordRE.ReplaceAllString(s, "'password' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"',")
		return cloneFlarumURLRE.ReplaceAllString(s, "'url' => 'https://"+cmsapp.EscapePHPSingleQuoted(c.DstDomainWithSubdir)+"',")
	}, "Failed to set 'database' in config.php")
}
