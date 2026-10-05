package phpapp

import (
	"regexp"
	"strings"
)

var archiveURLRE = regexp.MustCompile(`(?i)^https://[^\s'"]+\.(zip|tar\.gz|tgz|tar)$`)

// isArchiveURL reports whether initialProject looks like a downloadable archive URL (as opposed to a Composer package name like "laravel/laravel")
func isArchiveURL(initialProject string) bool {
	return archiveURLRE.MatchString(initialProject)
}

// composerPackageRE matches a Composer package name: vendor/package, each segment lowercase alphanumeric plus ., _, -
var composerPackageRE = regexp.MustCompile(`^[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?/[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?$`)

// isValidInitialProject accepts empty (no initial project), an archive URL, or a valid-looking Composer package name
func isValidInitialProject(initialProject string) bool {
	if initialProject == "" {
		return true
	}
	if len(initialProject) > 500 {
		return false
	}
	if strings.ContainsAny(initialProject, "\n\r") {
		return false
	}
	if isArchiveURL(initialProject) {
		return true
	}
	return composerPackageRE.MatchString(initialProject)
}
