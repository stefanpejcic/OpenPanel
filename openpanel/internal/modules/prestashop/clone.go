package prestashop

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// mirrors wordpress/manage.go's handleCloneWordPress in shape (file copy, DB create+dump, config rewrite, sites insert), sharing everything but docroot copy and config rewrite with every other CMS via internal/core/cmsclone
// unlike Joomla/Drupal, PrestaShop does hardcode its domain in the DB - the `{prefix}shop_url` table's domain/domain_ssl/physical_uri columns - so after the DB import this updates that row to point the clone at its own new domain/subdirectory, the PrestaShop equivalent of wp-cli's search-replace step

var (
	clonePrestaDBNameRE   = regexp.MustCompile(`'database_name'\s*=>\s*'.*?',`)
	clonePrestaDBUserRE   = regexp.MustCompile(`'database_user'\s*=>\s*'.*?',`)
	clonePrestaDBPasswdRE = regexp.MustCompile(`'database_password'\s*=>\s*'.*?',`)
)

// var/cache/{prod,dev} holds a compiled service container with the source site's DB credentials baked in, which 500s the clone's first request until cleared
func cloneClearCache(c *cmsapp.Clone) {
	_ = os.RemoveAll(filepath.Join(c.DstPath, "var", "cache", "prod"))
	_ = os.RemoveAll(filepath.Join(c.DstPath, "var", "cache", "dev"))
}

func cloneConfig(c *cmsapp.Clone) map[string]any {
	return c.RewriteConfig("app/config/parameters.php", func(s string) string {
		s = clonePrestaDBNameRE.ReplaceAllString(s, "'database_name' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"',")
		s = clonePrestaDBUserRE.ReplaceAllString(s, "'database_user' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"',")
		return clonePrestaDBPasswdRE.ReplaceAllString(s, "'database_password' => '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"',")
	}, "Failed to set database_name in parameters.php")
}

// cloneShopURL points the clone's ps_shop_url row at its new domain/subdirectory, same fix install.go applies via --domain/--base_uri
func cloneShopURL(c *cmsapp.Clone) {
	dstSubdirURI := "/"
	if c.DstFolder != "" {
		dstSubdirURI = "/" + c.DstFolder + "/"
	}
	escapedDstDomain := escapeMySQLString(c.DstDomain)
	escapedURI := escapeMySQLString(dstSubdirURI)
	_, _ = mysqlmanager.Exec(c.Ctx, c.UserContext,
		"UPDATE `ps_shop_url` SET domain = '"+escapedDstDomain+"', domain_ssl = '"+escapedDstDomain+"', physical_uri = '"+escapedURI+"'",
		c.DstDB)
}

func escapeMySQLString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
