package matomo

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors drupal/clone.go in structure (site-limit check, file copy, DB create+dump-pipe, config rewrite, sites-table insert) via internal/core/cmsclone - Matomo's own difference: config/config.ini.php is INI, not a PHP array/define() list, and it hardcodes a [General] trusted_hosts[] array of allowed HTTP Host headers, so the clone's own domain must be appended there or Matomo rejects every request to the clone with "untrusted host"

var (
	cloneMatomoDBNameRE       = regexp.MustCompile(`(?m)^dbname\s*=\s*"[^"]*"`)
	cloneMatomoDBUserRE       = regexp.MustCompile(`(?m)^username\s*=\s*"[^"]*"`)
	cloneMatomoDBPasswordRE   = regexp.MustCompile(`(?m)^password\s*=\s*"[^"]*"`)
	cloneMatomoTrustedHostsRE = regexp.MustCompile(`(?m)^trusted_hosts\[\]\s*=\s*"[^"]*"$`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	_ = exec.CommandContext(c.Ctx, "chmod", "644", filepath.Join(c.DstPath, "config", "config.ini.php")).Run()
	escapedPassword := strings.ReplaceAll(c.DstDBUserPassword, `"`, `\"`)
	return c.RewriteConfig("config/config.ini.php", func(s string) string {
		s = cloneMatomoDBNameRE.ReplaceAllString(s, `dbname = "`+c.DstDB+`"`)
		s = cloneMatomoDBUserRE.ReplaceAllString(s, `username = "`+c.DstDBUser+`"`)
		s = cloneMatomoDBPasswordRE.ReplaceAllString(s, `password = "`+escapedPassword+`"`)
		if cloneMatomoTrustedHostsRE.MatchString(s) {
			if !strings.Contains(s, `trusted_hosts[] = "`+c.DstDomain+`"`) {
				lastMatch := cloneMatomoTrustedHostsRE.FindAllStringIndex(s, -1)
				insertAt := lastMatch[len(lastMatch)-1][1]
				s = s[:insertAt] + "\ntrusted_hosts[] = \"" + c.DstDomain + "\"" + s[insertAt:]
			}
		} else {
			s = strings.Replace(s, "[General]", "[General]\ntrusted_hosts[] = \""+c.DstDomain+"\"", 1)
		}
		return s
	}, "Failed to set 'dbname' in config.ini.php")
}
