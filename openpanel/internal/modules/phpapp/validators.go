package phpapp

import (
	stdpath "path"
	"regexp"
	"strings"
)

// isValidSubdirectory mirrors appinstall.isValidSubdirectory: empty is
// valid (install at docroot), otherwise no path traversal and no leading
// slash.
func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// noPathTraversal mirrors appinstall.noPathTraversal: normalizes the path
// then requires it lands under /var/www/html/.
func noPathTraversal(p string) bool {
	if strings.ContainsAny(p, "\n\r") {
		return false
	}
	cleaned := stdpath.Clean(p)
	if !strings.HasPrefix(cleaned, "/var/www/html/") {
		return false
	}
	for _, part := range strings.Split(cleaned, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

var archiveURLRE = regexp.MustCompile(`(?i)^https://[^\s'"]+\.(zip|tar\.gz|tgz|tar)$`)

// isArchiveURL reports whether initialProject looks like a downloadable
// archive URL (as opposed to a Composer package name like "laravel/laravel").
func isArchiveURL(initialProject string) bool {
	return archiveURLRE.MatchString(initialProject)
}

// composerPackageRE matches a Composer package name: vendor/package, each
// segment lowercase alphanumeric plus ., _, -.
var composerPackageRE = regexp.MustCompile(`^[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?/[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?$`)

// isValidInitialProject accepts empty (no initial project), an archive URL,
// or a valid-looking Composer package name.
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
