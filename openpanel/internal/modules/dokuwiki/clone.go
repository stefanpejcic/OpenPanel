package dokuwiki

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// cloneDokuwikiTitleRE matches the $conf['title'] line writeDokuwikiConfig writes, the only place a source domain ends up in conf/local.php
var cloneDokuwikiTitleRE = regexp.MustCompile(`\$conf\['title'\]\s*=\s*'.*?';`)

// cloneSetTitle points conf/local.php's title at the new domain, hardcoded links inside data/pages/*.txt stay as-is
func cloneSetTitle(c *cmsapp.Clone) {
	localConfigFile := filepath.Join(c.DstPath, "conf", "local.php")
	if content, readErr := os.ReadFile(localConfigFile); readErr == nil {
		// "$$" is a literal "$", without it ReplaceAllString treats "$conf" as a submatch ref and drops it
		newContent := cloneDokuwikiTitleRE.ReplaceAllString(string(content), "$$conf['title'] = '"+cmsapp.EscapePHPSingleQuoted(c.DstDomainWithSubdir)+"';")
		_ = os.WriteFile(localConfigFile, []byte(newContent), 0o644)
	}
}

func cloneVersion(c *cmsapp.Clone) string {
	if versionBytes, readErr := os.ReadFile(filepath.Join(c.DstPath, "VERSION")); readErr == nil {
		return strings.TrimSpace(string(versionBytes))
	}
	return "unknown"
}
