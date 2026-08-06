// Package validators provides input-validation helpers for panel forms and
// API requests.
package validators

import (
	"math"
	"regexp"
	"strconv"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	validHostRegex = regexp.MustCompile(`^[\w.\-]+$`)
	lowerRegex     = regexp.MustCompile(`[a-z]`)
	upperRegex     = regexp.MustCompile(`[A-Z]`)
	digitRegex     = regexp.MustCompile(`\d`)
	specialRegex   = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

var allowedMySQLHosts = map[string]bool{"%": true, "localhost": true, "127.0.0.1": true}

func IsValidIdentifier(name string) bool {
	return validNameRegex.MatchString(name)
}

func IsValidHost(host string) bool {
	return allowedMySQLHosts[host] || validHostRegex.MatchString(host)
}

// ClampPasswordStrength parses raw as an int and clamps it to [1, 100],
// falling back to def on a parse failure.
func ClampPasswordStrength(raw string, def int) int {
	value := def
	if v, err := strconv.Atoi(raw); err == nil {
		value = v
	}
	if value < 1 {
		return 1
	}
	if value > 100 {
		return 100
	}
	return value
}

// PasswordStrengthScore mirrors the 6-check rubric shared with
// static/js/password-strength.js and opencli/lib/password_strength.sh -
// keep all three in sync if this changes.
func PasswordStrengthScore(password string) int {
	if password == "" {
		return 0
	}
	score := 0
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if lowerRegex.MatchString(password) {
		score++
	}
	if upperRegex.MatchString(password) {
		score++
	}
	if digitRegex.MatchString(password) {
		score++
	}
	if specialRegex.MatchString(password) {
		score++
	}
	return int(math.Round(float64(score) / 6 * 100))
}

func IsPasswordStrongEnough(password string, threshold int) bool {
	return PasswordStrengthScore(password) >= threshold
}
