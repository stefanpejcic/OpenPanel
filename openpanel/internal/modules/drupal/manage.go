package drupal

import (
	"regexp"
	"strings"
)

var (
	removeDBNameRE = regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)
	removeDBUserRE = regexp.MustCompile(`'username'\s*=>\s*'([^']*)'`)
)

// stripPHPCommentLines drops every line starting with "*" - settings.php ships a doc comment block with placeholder 'database' => 'database_name' lines that would otherwise match removeDBNameRE/removeDBUserRE before the real array drush appends
func stripPHPCommentLines(content string) string {
	var codeLines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	return strings.Join(codeLines, "\n")
}
