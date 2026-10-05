package opencart

import (
	"os"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
)

// mirrors wordpress/manage.go's handleCloneWordPress in shape (file copy, DB create+dump, config rewrite, sites insert), sharing everything but docroot copy and config rewrite with every other CMS via internal/core/cmsclone; OpenCart hardcodes its URL and filesystem path in two files, config.php and admin/config.php, so only HTTP_SERVER/HTTP_CATALOG and DIR_OPENCART need rewriting since every other DIR_* constant derives from DIR_OPENCART, and oc_setting stores no url-related key so no DB-side fix is needed; cmsclone.ValidDocroot accepts the real absolute "/var/www/html/..." form used everywhere here, unlike wordpress's own validateDocroot which would reject it

var (
	cloneOCHTTPServerRE  = regexp.MustCompile(`define\('HTTP_SERVER',\s*'.*?'\);`)
	cloneOCHTTPCatalogRE = regexp.MustCompile(`define\('HTTP_CATALOG',\s*'.*?'\);`)
	cloneOCDirOpenCartRE = regexp.MustCompile(`define\('DIR_OPENCART',\s*'.*?'\);`)
	cloneOCDBHostnameRE  = regexp.MustCompile(`define\('DB_HOSTNAME',\s*'.*?'\);`)
	cloneOCDBUsernameRE  = regexp.MustCompile(`define\('DB_USERNAME',\s*'.*?'\);`)
	cloneOCDBPasswordRE  = regexp.MustCompile(`define\('DB_PASSWORD',\s*'.*?'\);`)
	cloneOCDBDatabaseRE  = regexp.MustCompile(`define\('DB_DATABASE',\s*'.*?'\);`)
)

// cloneConfig rewrites both config.php and admin/config.php for the new domain, path and db
func cloneConfig(c *cmsapp.Clone) map[string]any {
	dstBaseURLPath := ""
	if c.DstFolder != "" {
		dstBaseURLPath = c.DstFolder + "/"
	}
	newHTTPServer := "https://" + c.DstDomain + "/" + dstBaseURLPath
	newHTTPCatalog := "https://" + c.DstDomain + "/" + dstBaseURLPath
	newDirOpenCart := "/var/www/html/" + c.DstDomainWithSubdir + "/"

	rewriteConfig := func(relPath string, isAdmin bool) error {
		fp := filepath.Join(c.DstPath, relPath)
		content, readErr := os.ReadFile(fp)
		if readErr != nil {
			return readErr
		}
		strContent := string(content)
		if isAdmin {
			strContent = cloneOCHTTPServerRE.ReplaceAllString(strContent, "define('HTTP_SERVER', '"+cmsapp.EscapePHPSingleQuoted(newHTTPServer+"admin/")+"');")
			strContent = cloneOCHTTPCatalogRE.ReplaceAllString(strContent, "define('HTTP_CATALOG', '"+cmsapp.EscapePHPSingleQuoted(newHTTPCatalog)+"');")
		} else {
			strContent = cloneOCHTTPServerRE.ReplaceAllString(strContent, "define('HTTP_SERVER', '"+cmsapp.EscapePHPSingleQuoted(newHTTPServer)+"');")
		}
		strContent = cloneOCDirOpenCartRE.ReplaceAllString(strContent, "define('DIR_OPENCART', '"+cmsapp.EscapePHPSingleQuoted(newDirOpenCart)+"');")
		strContent = cloneOCDBHostnameRE.ReplaceAllString(strContent, "define('DB_HOSTNAME', '"+mysql.AppDBHost(c.UserContext, c.MySQLVersion)+"');")
		strContent = cloneOCDBUsernameRE.ReplaceAllString(strContent, "define('DB_USERNAME', '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUser)+"');")
		strContent = cloneOCDBPasswordRE.ReplaceAllString(strContent, "define('DB_PASSWORD', '"+cmsapp.EscapePHPSingleQuoted(c.DstDBUserPassword)+"');")
		strContent = cloneOCDBDatabaseRE.ReplaceAllString(strContent, "define('DB_DATABASE', '"+cmsapp.EscapePHPSingleQuoted(c.DstDB)+"');")
		return os.WriteFile(fp, []byte(strContent), 0o644)
	}

	if rwErr := rewriteConfig("config.php", false); rwErr != nil {
		return map[string]any{"status": "error", "details": "config.php: " + rwErr.Error()}
	}
	if rwErr := rewriteConfig(filepath.Join("admin", "config.php"), true); rwErr != nil {
		return map[string]any{"status": "error", "details": "admin/config.php: " + rwErr.Error()}
	}
	return nil
}
