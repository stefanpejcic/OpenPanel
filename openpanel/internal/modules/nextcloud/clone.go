package nextcloud

import (
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors wordpress/manage.go's handleCloneWordPress in shape (file copy, DB create+dump, config rewrite, sites insert), sharing everything but docroot copy and config rewrite with every other CMS via internal/core/cmsclone; config.php's datadirectory, trusted_domains (index 1), overwrite.cli.url and db creds are fixed via regex text-replace rather than occ, since occ needs a matching PHP runtime and this hit PHP-version failures elsewhere (see maintenance.go); instanceid is left untouched so the copied data/appdata_<instanceid>/ dir keeps matching it; cmsclone.ValidDocroot accepts the real absolute "/var/www/html/..." form used everywhere here, unlike wordpress's own validateDocroot which would reject it

var (
	cloneNCDataDirRE   = regexp.MustCompile(`'datadirectory'\s*=>\s*'.*?',`)
	cloneNCOverwriteRE = regexp.MustCompile(`'overwrite\.cli\.url'\s*=>\s*'.*?',`)
	cloneNCDBNameRE    = regexp.MustCompile(`'dbname'\s*=>\s*'.*?',`)
	cloneNCDBUserRE    = regexp.MustCompile(`'dbuser'\s*=>\s*'.*?',`)
	cloneNCDBPasswdRE  = regexp.MustCompile(`'dbpassword'\s*=>\s*'.*?',`)
	// matches trusted_domains index 1, right after the always-present "0 => 'localhost'," line - see install.go for why index 1 holds the site's domain
	cloneNCTrustedDomainRE = regexp.MustCompile(`(0 => 'localhost',\s*\n\s*1 => )'.*?'(,)`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	return c.RewriteConfig("config/config.php", func(s string) string {
		s = cloneNCDataDirRE.ReplaceAllString(s, "'datadirectory' => '"+cmsapp.EscapePHPSingleQuoted(strings.TrimSuffix(c.Docroot, "/")+"/data")+"',")
		s = cloneNCOverwriteRE.ReplaceAllString(s, "'overwrite.cli.url' => 'https://"+cmsapp.EscapePHPSingleQuoted(c.DstDomainWithSubdir)+"',")
		s = cloneNCDBNameRE.ReplaceAllString(s, "'dbname' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"',")
		s = cloneNCDBUserRE.ReplaceAllString(s, "'dbuser' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"',")
		s = cloneNCDBPasswdRE.ReplaceAllString(s, "'dbpassword' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"',")
		return cloneNCTrustedDomainRE.ReplaceAllString(s, "${1}'"+cmsapp.EscapePHPSingleQuoted(c.DstDomain)+"'${2}")
	}, "Failed to set dbname in config.php")
}
