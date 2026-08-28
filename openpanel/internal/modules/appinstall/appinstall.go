// Package appinstall handles the near-identical Python, NodeJS, and Ruby
// app install flows (an app is installed by generating a docker-compose
// service entry from a shared template, starting the container, then
// wiring a reverse-proxy into the domain's vhost config), plus the
// always-on helper routes these install forms use (/docker/tags/<type>,
// /json/check_if_file_exists). The three app types differ only in a
// handful of string constants, captured here as a Kind.
package appinstall

import (
	"os"
	stdpath "path"
	"regexp"
	"strconv"
	"strings"
)

// Kind captures every point where the Python, NodeJS, and Ruby install
// flows diverge. The three Kind values themselves (NodeJS, Python, Ruby)
// live in nodejs.go/python.go/ruby.go respectively, one file per type, so
// each type's specifics are easy to find and change independently of the
// others - everything in this file and the rest of the package works the
// same regardless of which Kind it's handed.
type Kind struct {
	AppType        string // "python" | "nodejs" | "ruby", used in URLs and the compose template filename
	DisplayAppType string // "Python" | "NodeJS" | "Ruby", stored in sites.type and shown in messages
	PyOrNode       string // "PY" | "NODE" | "RUBY", the env-var/compose-template placeholder infix
	Title          string // page title / DISPLAY_APP_TYPE-based heading

	InstallToken       string // e.g. "npm install", "pip install -r requirements.txt", "bundle install"
	RunToken           string // e.g. "node", "python", "ruby"
	DefaultStartupFile string // e.g. "index.js", "app.py", "app.rb"
}

// kindsByAppType resolves a sites.type value (lowercase) to its Kind -
// replaces what used to be a "nodejs"/"python"/"ruby" -> pyOrNode switch
// statement repeated at every PM2 call site (handlePM2Action,
// handlePM2Delete, apiPM2Update, apiPM2Delete).
var kindsByAppType = map[string]Kind{
	NodeJS.AppType: NodeJS,
	Python.AppType: Python,
	Ruby.AppType:   Ruby,
	Java.AppType:   Java,
}

// kindByAppType looks up a Kind by its lowercase sites.type value.
func kindByAppType(appType string) (Kind, bool) {
	kind, ok := kindsByAppType[strings.ToLower(appType)]
	return kind, ok
}

var validServiceNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// isValidServiceName is deliberately looser than docker.IsValidServiceName's
// stricter lowercase-only regex - both install types (python, nodejs)
// share this looser rule.
func isValidServiceName(name string) bool {
	return validServiceNameRE.MatchString(name)
}

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// versionRE accepts plain dotted-numeric tags ("3.3.6", "18") used by
// ruby/python/nodejs, plus the "_NN" and "-jdk-jammy" suffix eclipse-
// temurin's own Java tags always carry (see shared.go's javaCleanTagRE -
// that image ships no plain numeric tags at all).
var versionRE = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*(_[0-9]+)?(-jdk-jammy)?$`)

func isValidVersion(version string) bool {
	return versionRE.MatchString(version)
}

// noPathTraversal normalizes the path (resolving "." and ".." - for an
// absolute path this always fully resolves any ".." that stays within the
// tree) then requires it lands under /var/www/html/.
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

// isValidStartupFile is shared between all app types: it accepts .py, .js,
// .rb, or .java regardless of which install form submitted it.
func isValidStartupFile(path string) bool {
	if !noPathTraversal(path) {
		return false
	}
	return strings.HasSuffix(path, ".py") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".rb") || strings.HasSuffix(path, ".java")
}

func isValidCustomCommand(cmd string) bool {
	if strings.Contains(cmd, "..") {
		return false
	}
	return !strings.ContainsAny(cmd, "\n\r")
}

// isValidGitURL is deliberately restrictive: the URL ends up single-quoted
// inside the container's startup shell command (see buildAppRunCommand), so
// a single quote would break out of that quoting, and only https:// is
// supported - there's no SSH deploy-key infrastructure to authenticate a
// git:// or ssh:// clone with. Empty is valid (git deploy is optional).
func isValidGitURL(url string) bool {
	if url == "" {
		return true
	}
	if len(url) > 500 {
		return false
	}
	if !strings.HasPrefix(url, "https://") {
		return false
	}
	return !strings.ContainsAny(url, "'\n\r \t")
}

// isValidRequirements accepts only "" (No) or "1" (Yes).
func isValidRequirements(req string) bool {
	return req == "" || req == "1"
}

// isValidWorkdir is a direct alias of noPathTraversal().
func isValidWorkdir(path string) bool {
	return noPathTraversal(path)
}

func isPositiveNumber(value string) bool {
	v, err := strconv.ParseFloat(value, 64)
	return err == nil && v > 0
}

// getValidatedFloat returns a positive float, or the parsed default
// (assumed itself valid - every call site passes a literal like "1.0").
func getValidatedFloat(value, def string) float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	d, _ := strconv.ParseFloat(def, 64)
	return d
}

// getValidatedInt returns a positive int, or the parsed default (assumed
// itself valid - every call site passes a literal like "100").
func getValidatedInt(value, def string) int {
	if v, err := strconv.Atoi(value); err == nil && v > 0 {
		return v
	}
	d, _ := strconv.Atoi(def)
	return d
}

// gitBootstrapCmd returns the shell snippet that makes sure `git` is on
// PATH (the official node/python images it runs in don't include it, and
// - being official images - can't be edited, only the compose file that
// launches them), then fetches the repo's default branch and hard-resets
// the working tree to it, every time the container starts (first install
// and every later restart alike). gitURL is trusted to already be
// validated by isValidGitURL (https://-only, no quote/whitespace/newline
// chars), so single-quoting it here is safe.
//
// This deliberately avoids both `git clone` and `git pull`:
//   - `git clone` refuses to clone into a non-empty directory, but a repo
//     can be connected later from an already-installed app (see the
//     manage page), whose docroot then already has files in it - `init`
//     (idempotent - skipped if already a repo from a prior start) +
//     `remote add` (idempotent - skipped if already configured) + `fetch`
//   - `reset --hard` adopts the existing directory in place instead.
//   - `git pull` depends on the current branch having upstream tracking
//     info, which `reset --hard FETCH_HEAD` deliberately doesn't set up
//     (it leaves a detached HEAD) - so a plain `pull` on the next restart
//     would fail outright. Fetching+resetting every time sidesteps branch
//     tracking entirely and works the same way on every start.
//
// `fetch origin HEAD` resolves the remote's actual default branch without
// needing to name it.
func gitBootstrapCmd(gitURL string) string {
	if gitURL == "" {
		return ""
	}
	return "(command -v git >/dev/null 2>&1 || (apt-get update -qq && apt-get install -y -qq git)) && " +
		"(git rev-parse --is-inside-work-tree >/dev/null 2>&1 || git init -q) && " +
		"(git remote get-url origin >/dev/null 2>&1 || git remote add origin '" + gitURL + "') && " +
		"git fetch --depth 1 origin HEAD && git reset --hard FETCH_HEAD && "
}

// buildAppRunCommand assembles the container's startup shell command from
// kind's install/run tokens (see nodejs.go/python.go/ruby.go) - the
// algorithm itself is identical across all three types, only the tokens
// differ.
func buildAppRunCommand(kind Kind, requirements, customCmd, startupFile, gitURL string) string {
	var installCmd string
	if requirements == "1" {
		installCmd = kind.InstallToken + " && "
	}

	defaultRun := kind.RunToken + " " + kind.DefaultStartupFile
	if startupFile != "" {
		defaultRun = kind.RunToken + " " + startupFile
	}

	runCmd := customCmd
	if runCmd == "" {
		runCmd = defaultRun
	}
	return gitBootstrapCmd(gitURL) + installCmd + runCmd
}

// normalizeRequirements maps common truthy strings to "1", everything
// else to "".
func normalizeRequirements(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "on", "true", "yes":
		return "1"
	default:
		return ""
	}
}

// fileExists checks the small paths this package cares about (backup
// files during the install's rollback path).
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
