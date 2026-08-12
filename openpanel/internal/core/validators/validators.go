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
	// emailUsernameRegex mirrors the client-side pattern on
	// templates/emails/new.html's username field - notably, it excludes
	// '@' so a mailbox's local part can't smuggle in a second address.
	emailUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+$`)
	// panelUsernameRegex mirrors the client-side pattern on
	// templates/user/account.html's username field.
	panelUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)
)

var allowedMySQLHosts = map[string]bool{"%": true, "localhost": true, "127.0.0.1": true}

func IsValidIdentifier(name string) bool {
	return validNameRegex.MatchString(name)
}

func IsValidHost(host string) bool {
	return allowedMySQLHosts[host] || validHostRegex.MatchString(host)
}

// IsValidEmailUsername reports whether username is a valid mailbox local
// part: letters, digits, and '.', '_', '%', '+', '-'. Critically, this
// excludes '@' - without a server-side check, a value like "@example.com"
// concatenates into a malformed two-address string wherever callers build
// "username@domain" themselves.
func IsValidEmailUsername(username string) bool {
	return emailUsernameRegex.MatchString(username)
}

// IsValidPanelUsername reports whether username is a valid OpenPanel
// account username: 3-20 letters/digits, no other characters. Without a
// server-side check here, a rename to a username outside this set could
// confuse opencli's own path/identity assumptions downstream, since the
// username becomes part of filesystem paths (e.g. /home/<username>).
func IsValidPanelUsername(username string) bool {
	return panelUsernameRegex.MatchString(username)
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
