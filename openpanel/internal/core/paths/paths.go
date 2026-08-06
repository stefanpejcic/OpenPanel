// Package paths implements the path-traversal guard used by every
// file-manager route that touches a user-supplied path. Security-critical -
// the checks and their ordering are kept deliberately exact rather than
// "equivalent but simplified", since subtle behavior differences here are
// exploitable.
package paths

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Error is one of SecureUserPath's abort(code, message) failures.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string { return e.Message }

func abort(code int, message string) error {
	return &Error{Code: code, Message: message}
}

var controlCharsRE = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f<>:"|?*]`)

// dangerousPatterns is the substring blocklist checked against the
// lowercased full path.
var dangerousPatterns = []string{
	"..", "/..", "\\..", "../", "..\\",
	"%2e%2e", "%252e%252e", "..%2f", "..%5c",
	"…..", "..%c0%af", "..%c1%9c",
}

// homeVolumeOverride lets tests point the HOME base directory computation
// at a temp directory instead of the real (root-owned) /home. Production
// code never sets this; it's a test-only seam, matching the same pattern
// used by internal/modules/docker's homeDirOverride.
var homeVolumeOverride string

// htmlRootOverride is the equivalent test seam for the HTML base
// directory ("/var/www/html" in production).
var htmlRootOverride string

// baseDir resolves "HOME" or "HTML" to the per-user root directory. HOME
// must already exist (strict resolve); a missing HOME volume is a 404, not
// a 403, since it's a legitimate not-yet-provisioned state rather than an
// attack.
func baseDir(kind, context string) (string, error) {
	switch strings.ToUpper(kind) {
	case "HTML":
		htmlRoot := htmlRootOverride
		if htmlRoot == "" {
			htmlRoot = "/var/www/html"
		}
		// Resolved non-strict here: the /var/www/html/<context> tenant
		// directory may not exist yet - that's handled later.
		return filepath.Join(htmlRoot, context), nil
	case "HOME":
		var dir string
		if homeVolumeOverride != "" {
			dir = filepath.Join(homeVolumeOverride, context)
		} else {
			dir = "/home/" + context + "/docker-data/volumes/" + context + "_html_data/_data"
		}
		resolved, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return "", abort(404, "Volume is not yet created. Please add a domain name first.")
		}
		return resolved, nil
	default:
		return "", abort(403, "Invalid base directory")
	}
}

// SecureUserPath validates and resolves a user-supplied path within the
// account's HOME or HTML root, guarding against path traversal and
// symlink escapes. base is "HOME" or "HTML".
func SecureUserPath(base, context, userInputPath string, checkExists bool) (string, error) {
	const maxPathLength = 200

	if context == "" || strings.ContainsAny(context, "/\\") {
		return "", abort(403, "Invalid context")
	}

	userHome, err := baseDir(base, context)
	if err != nil {
		return "", err
	}

	if userInputPath == "" {
		// The HOME branch always exists by this point (baseDir above
		// already 404s otherwise), so the "fall back to /var/www/html"
		// branch below is effectively HTML-only.
		if filepath.Base(userHome) == context {
			if _, statErr := os.Stat(userHome); statErr != nil {
				return filepath.EvalSymlinks("/var/www/html")
			}
		}
		resolved, err := filepath.EvalSymlinks(userHome)
		if err != nil {
			return "", abort(404, "User root directory not found")
		}
		return resolved, nil
	}

	if len(userInputPath) > maxPathLength {
		return "", abort(403, "Path too long")
	}

	normalized := strings.ReplaceAll(userInputPath, "\\", "/")
	pathParts := strings.Split(normalized, "/")
	for _, part := range pathParts {
		if part == "" {
			continue
		}
		if controlCharsRE.MatchString(part) {
			return "", abort(403, "Invalid characters in filename")
		}
	}

	lowered := strings.ToLower(userInputPath)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowered, pattern) {
			return "", abort(403, "Suspicious path pattern detected")
		}
	}

	// Walk each path segment from userHome, rejecting any intermediate
	// component that's a symlink. This catches an escape attempt even when
	// the final resolved path would still (accidentally) land inside
	// userHome.
	current := userHome
	var nonEmptyParts []string
	for _, part := range pathParts {
		if part == "" {
			continue
		}
		nonEmptyParts = append(nonEmptyParts, part)
		current = filepath.Join(current, part)
		if info, statErr := os.Lstat(current); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", abort(403, "Symlinks are not allowed")
		}
	}

	targetPath := filepath.Join(append([]string{userHome}, nonEmptyParts...)...)

	var resolvedTarget string
	if checkExists {
		resolvedTarget, err = filepath.EvalSymlinks(targetPath)
		if err != nil {
			return "", abort(404, "Not found")
		}
	} else {
		resolvedTarget, err = resolveNonStrict(targetPath)
		if err != nil {
			return "", abort(403, "Forbidden")
		}
	}

	if resolvedTarget != userHome && !isWithin(userHome, resolvedTarget) {
		return "", abort(403, "Path traversal or symlink escape detected")
	}

	if checkExists {
		if _, statErr := os.Stat(resolvedTarget); statErr != nil {
			return "", abort(404, "File or directory not found")
		}
	}

	return resolvedTarget, nil
}

// isWithin reports whether target is a strict descendant of home.
func isWithin(home, target string) bool {
	rel, err := filepath.Rel(home, target)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

// resolveNonStrict resolves symlinks for the longest existing prefix of
// path, then appends the remaining (not-yet-existing) components
// literally, cleaned but not symlink-resolved.
func resolveNonStrict(path string) (string, error) {
	clean := filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		return resolved, nil
	}

	dir, base := filepath.Split(clean)
	dir = strings.TrimSuffix(dir, string(filepath.Separator))
	if dir == "" || dir == clean {
		return clean, nil
	}
	resolvedDir, err := resolveNonStrict(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedDir, base), nil
}
