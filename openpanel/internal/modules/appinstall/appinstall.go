// Package appinstall ports the near-identical modules/python.py and
// modules/nodejs.py (a Python/NodeJS app is installed by generating a
// docker-compose service entry from a shared template, starting the
// container, then wiring a reverse-proxy into the domain's vhost config),
// plus the two always-on helper routes from modules/json/helpers.py that
// only these two install forms use (/docker/tags/<type>,
// /json/check_if_file_exists). The two Python modules differ only in a
// handful of string constants, captured here as a Kind.
package appinstall

import (
	"os"
	stdpath "path"
	"regexp"
	"strconv"
	"strings"
)

// Kind captures every point where python.py and nodejs.py diverge.
type Kind struct {
	AppType        string // "python" | "nodejs", used in URLs and the compose template filename
	DisplayAppType string // "Python" | "NodeJS", stored in sites.type and shown in messages
	PyOrNode       string // "PY" | "NODE", the env-var/compose-template placeholder infix
	Title          string // page title / DISPLAY_APP_TYPE-based heading
}

var (
	Python = Kind{AppType: "python", DisplayAppType: "Python", PyOrNode: "PY", Title: "Install Python Application"}
	NodeJS = Kind{AppType: "nodejs", DisplayAppType: "NodeJS", PyOrNode: "NODE", Title: "Install NodeJS Application"}
)

var validServiceNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// isValidServiceName mirrors helpers.is_valid_service_name() - NOT
// docker.IsValidServiceName's stricter lowercase-only regex; python.py and
// nodejs.py both import the looser helpers.py version.
func isValidServiceName(name string) bool {
	return validServiceNameRE.MatchString(name)
}

// isValidSubdirectory mirrors helpers.is_valid_subdirectory().
func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

var versionRE = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

// isValidVersion mirrors helpers.is_valid_version().
func isValidVersion(version string) bool {
	return versionRE.MatchString(version)
}

// noPathTraversal mirrors helpers.no_path_traversal(): normalize the path
// (resolving "." and ".." the same way Python's os.path.normpath does,
// which for an absolute path always fully resolves any ".." that stays
// within the tree) then require it lands under /var/www/html/.
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

// isValidStartupFile mirrors helpers.is_valid_startup_file(): shared
// between both app types (accepts either .py or .js regardless of which
// install form submitted it, matching Python's own lack of
// type-specificity here).
func isValidStartupFile(path string) bool {
	if !noPathTraversal(path) {
		return false
	}
	return strings.HasSuffix(path, ".py") || strings.HasSuffix(path, ".js")
}

// isValidCustomCommand mirrors helpers.is_valid_custom_command().
func isValidCustomCommand(cmd string) bool {
	if strings.Contains(cmd, "..") {
		return false
	}
	return !strings.ContainsAny(cmd, "\n\r")
}

// isValidRequirements mirrors helpers.is_valid_requirements(): only "" (No)
// or "1" (Yes) are accepted.
func isValidRequirements(req string) bool {
	return req == "" || req == "1"
}

// isValidWorkdir mirrors helpers.is_valid_workdir(), a direct alias of
// no_path_traversal().
func isValidWorkdir(path string) bool {
	return noPathTraversal(path)
}

// isPositiveNumber mirrors helpers.is_positive_number().
func isPositiveNumber(value string) bool {
	v, err := strconv.ParseFloat(value, 64)
	return err == nil && v > 0
}

// getValidatedFloat mirrors helpers.get_validated_float(): a positive
// float, or the parsed default (assumed itself valid, matching every call
// site passing a literal like "1.0").
func getValidatedFloat(value, def string) float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	d, _ := strconv.ParseFloat(def, 64)
	return d
}

// buildAppRunCommand mirrors helpers.build_app_run_command().
func buildAppRunCommand(pyOrNode, requirements, customCmd, startupFile string) string {
	var installCmd, defaultRun string
	if pyOrNode == "NODE" {
		if requirements == "1" {
			installCmd = "npm install && "
		}
		if startupFile != "" {
			defaultRun = "node " + startupFile
		} else {
			defaultRun = "node index.js"
		}
	} else {
		if requirements == "1" {
			installCmd = "pip install -r requirements.txt && "
		}
		if startupFile != "" {
			defaultRun = "python " + startupFile
		} else {
			defaultRun = "python app.py"
		}
	}

	runCmd := customCmd
	if runCmd == "" {
		runCmd = defaultRun
	}
	return installCmd + runCmd
}

// normalizeRequirements mirrors helpers.normalize_requirements() (inlined
// at python.py/nodejs.py's call site as `'1' if str(req_raw).lower() in
// (...) else ”`).
func normalizeRequirements(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "on", "true", "yes":
		return "1"
	default:
		return ""
	}
}

// fileExists mirrors `os.path.exists` for the small paths this package
// checks (backup files during the install's rollback path).
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
