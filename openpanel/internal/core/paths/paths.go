// Package paths makes sure a user-supplied path can't escape their home dir. Security-critical, be careful changing this.
package paths

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Error is returned when SecureUserPath rejects a path.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string { return e.Message }

func abort(code int, message string) error {
	return &Error{Code: code, Message: message}
}

var controlCharsRE = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f<>:"|?*]`)

// dangerousPatterns are blocked substrings, checked against the lowercased path.
var dangerousPatterns = []string{
	"..", "/..", "\\..", "../", "..\\",
	"%2e%2e", "%252e%252e", "..%2f", "..%5c",
	"…..", "..%c0%af", "..%c1%9c",
}

// homeVolumeOverride lets tests point HOME at a temp dir instead of the real /home.
var homeVolumeOverride string

// htmlRootOverride is the same thing but for the HTML root.
var htmlRootOverride string

// baseDir resolves "HOME" or "HTML" to the user's root dir. A missing HOME volume is a 404, not a 403 - it just isn't provisioned yet.
func baseDir(kind, context string) (string, error) {
	switch strings.ToUpper(kind) {
	case "HTML":
		htmlRoot := htmlRootOverride
		if htmlRoot == "" {
			htmlRoot = "/var/www/html"
		}
		// non-strict since this dir might not exist yet
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

// SecureUserPath resolves a user path inside HOME/HTML, blocking traversal and symlink escapes.
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
		// HOME always exists here, this fallback is really just for HTML
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

	// walk the path and reject any symlink along the way, even if the end result would land inside userHome anyway
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

// resolveNonStrict resolves symlinks as far as the path actually exists, then just appends what's left.
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
