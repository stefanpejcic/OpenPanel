// Package appinstall handles the near-identical Python, NodeJS, and Ruby app install flows (generate a docker-compose entry from a shared template, start the container, wire a reverse-proxy into the vhost config), plus the always-on helper routes the install forms use. The three types differ only in a handful of string constants, captured as a Kind.
package appinstall

import (
	"os"
	stdpath "path"
	"regexp"
	"strconv"
	"strings"
)

// Kind captures every point where the Python, NodeJS, and Ruby install flows diverge. The values themselves live in nodejs.go/python.go/ruby.go, one file per type; the rest of the package works the same regardless of which Kind it's handed.
type Kind struct {
	AppType        string // "python" | "nodejs" | "ruby", used in URLs and the compose template filename
	DisplayAppType string // "Python" | "NodeJS" | "Ruby", stored in sites.type and shown in messages
	PyOrNode       string // "PY" | "NODE" | "RUBY", the env-var/compose-template placeholder infix
	Title          string // page title / DISPLAY_APP_TYPE-based heading

	InstallToken       string // e.g. "npm install", "pip install -r requirements.txt", "bundle install"
	RunToken           string // e.g. "node", "python", "ruby"
	DefaultStartupFile string // e.g. "index.js", "app.py", "app.rb"
}

// kindsByAppType resolves a sites.type value (lowercase) to its Kind, replacing what used to be a repeated switch statement at every PM2 call site
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

// isValidServiceName is deliberately looser than docker.IsValidServiceName's stricter lowercase-only regex - both python and nodejs share this looser rule
func isValidServiceName(name string) bool {
	return validServiceNameRE.MatchString(name)
}

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// versionRE accepts plain dotted-numeric tags ("3.3.6", "18") used by ruby/python/nodejs, plus the "_NN"/"-jdk-jammy" suffixes eclipse-temurin's Java tags always carry (see shared.go's javaCleanTagRE)
var versionRE = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*(_[0-9]+)?(-jdk-jammy)?$`)

func isValidVersion(version string) bool {
	return versionRE.MatchString(version)
}

// noPathTraversal normalizes the path (resolving "." and "..") then requires it lands under /var/www/html/
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

// isValidStartupFile is shared between all app types: accepts .py, .js, .rb, or .java regardless of which install form submitted it
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

// isValidGitURL is deliberately restrictive - the URL ends up single-quoted in the container's startup shell command (buildAppRunCommand), so a quote would break out of it, and only https:// is supported since there's no SSH deploy-key setup. Empty is valid (git deploy is optional).
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

// getValidatedFloat returns a positive float, or the parsed default (assumed valid - every call site passes a literal like "1.0")
func getValidatedFloat(value, def string) float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	d, _ := strconv.ParseFloat(def, 64)
	return d
}

// getValidatedInt returns a positive int, or the parsed default (assumed valid - every call site passes a literal like "100")
func getValidatedInt(value, def string) int {
	if v, err := strconv.Atoi(value); err == nil && v > 0 {
		return v
	}
	d, _ := strconv.Atoi(def)
	return d
}

// gitBootstrapCmd returns the shell snippet that makes sure `git` is on PATH (the official node/python images don't include it), then fetches the repo's default branch and hard-resets the working tree to it every time the container starts - gitURL is already validated by isValidGitURL so single-quoting it is safe. Avoids `git clone` (refuses a non-empty directory) and `git pull` (needs upstream tracking info reset --hard FETCH_HEAD doesn't set up); `init`+`remote add` are idempotent, and `fetch origin HEAD` resolves the default branch without naming it.
func gitBootstrapCmd(gitURL string) string {
	if gitURL == "" {
		return ""
	}
	return "(command -v git >/dev/null 2>&1 || (apt-get update -qq && apt-get install -y -qq git)) && " +
		"(git rev-parse --is-inside-work-tree >/dev/null 2>&1 || git init -q) && " +
		"(git remote get-url origin >/dev/null 2>&1 || git remote add origin '" + gitURL + "') && " +
		"git fetch --depth 1 origin HEAD && git reset --hard FETCH_HEAD && "
}

// buildAppRunCommand assembles the container's startup shell command from kind's install/run tokens - the algorithm is identical across all three types, only the tokens differ
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

// normalizeRequirements maps common truthy strings to "1", everything else to ""
func normalizeRequirements(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "on", "true", "yes":
		return "1"
	default:
		return ""
	}
}

// fileExists checks the small paths this package cares about (backup files during the install's rollback path)
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
