package mediawiki

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// mirrors drupal/clone.go in structure (site-limit check, file copy, DB create+dump-pipe, config rewrite, sites-table insert) via internal/core/cmsclone, with two MediaWiki-specific differences: LocalSettings.php hardcodes $wgServer/$wgScriptPath to the source domain (unlike Drupal's runtime-derived base URL), so both must be rewritten to the clone's own domain or it keeps generating/accepting links only for the source; and MediaWiki needs its own per-site cron job (maintenance/runJobs.php) registered for the clone, mirroring install.go's own crons.AddJob call, or the clone's job queue never runs

var (
	cloneMediaWikiDBNameRE     = regexp.MustCompile(`\$wgDBname\s*=\s*"[^"]*"`)
	cloneMediaWikiDBUserRE     = regexp.MustCompile(`\$wgDBuser\s*=\s*"[^"]*"`)
	cloneMediaWikiDBPasswordRE = regexp.MustCompile(`\$wgDBpassword\s*=\s*"[^"]*"`)
	cloneMediaWikiServerRE     = regexp.MustCompile(`\$wgServer\s*=\s*"[^"]*"`)
	cloneMediaWikiScriptPathRE = regexp.MustCompile(`\$wgScriptPath\s*=\s*"[^"]*"`)
)

func cloneConfig(c *cmsapp.Clone) map[string]any {
	_ = exec.CommandContext(c.Ctx, "chmod", "644", filepath.Join(c.DstPath, "LocalSettings.php")).Run()
	escapedPassword := strings.ReplaceAll(c.DstDBUserPassword, `"`, `\"`)
	dstScriptPath := ""
	if c.DstFolder != "" {
		dstScriptPath = "/" + c.DstFolder
	}
	// ReplaceAllStringFunc since ReplaceAllString would expand the "$wg..." prefix as a submatch reference and drop it
	return c.RewriteConfig("LocalSettings.php", func(s string) string {
		s = cloneMediaWikiDBNameRE.ReplaceAllStringFunc(s, func(string) string { return `$wgDBname = "` + c.DstDB + `"` })
		s = cloneMediaWikiDBUserRE.ReplaceAllStringFunc(s, func(string) string { return `$wgDBuser = "` + c.DstDBUser + `"` })
		s = cloneMediaWikiDBPasswordRE.ReplaceAllStringFunc(s, func(string) string { return `$wgDBpassword = "` + escapedPassword + `"` })
		s = cloneMediaWikiServerRE.ReplaceAllStringFunc(s, func(string) string { return `$wgServer = "https://` + c.DstDomain + `"` })
		return cloneMediaWikiScriptPathRE.ReplaceAllStringFunc(s, func(string) string { return `$wgScriptPath = "` + dstScriptPath + `"` })
	}, "Failed to set 'wgDBname' in LocalSettings.php")
}

// cloneAddCron registers the clone's own runJobs.php job, the job queue doesn't run without it
func cloneAddCron(c *cmsapp.Clone) {
	webServer := webserver.GetEnvFileValue(c.UserContext, "WEB_SERVER")
	phpContainer := webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + c.PHPVersion
	}
	cronCommand := "php " + c.Docroot + "/maintenance/runJobs.php --maxjobs=50"
	_ = crons.AddJob(c.Ctx, c.UserContext, mediawikiCronComment(c.DstDomainWithSubdir), "0 * * * * *", phpContainer, cronCommand, true)
}
