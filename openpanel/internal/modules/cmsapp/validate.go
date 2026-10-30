package cmsapp

import (
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// PasswordStrength is the admin-set minimum score the wizard holds admin passwords to
func PasswordStrength(a *appctx.App) int {
	return validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
}

// same rules as the install wizard's client checks (static/js/install-wizard.js)
var (
	installDBNameRE = regexp.MustCompile(`^[A-Za-z0-9_]{1,32}$`)
	installUserRE   = regexp.MustCompile(`^[A-Za-z0-9._@-]{1,60}$`)
	installEmailRE  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// ValidEmail is the wizard's loose email shape check
func ValidEmail(v string) bool { return installEmailRE.MatchString(v) }

// ValidDBPassword allows printable ASCII up to 64 chars minus quotes, backslash and backtick, the installers drop it into SQL and config files
func ValidDBPassword(p string) bool {
	if p == "" || len(p) > 64 || strings.ContainsAny(p, "'\"\\`") {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] < 0x21 || p[i] > 0x7e {
			return false
		}
	}
	return true
}

// ValidateInstallForm checks the shared install fields an app form posts, get reads one form value, empty fields are left to the installer's defaults, "" means it's fine
func ValidateInstallForm(get func(string) string, strength int) string {
	for _, f := range []struct{ key, label string }{{"db_name", "database name"}, {"db_user", "database user"}} {
		if v := get(f.key); v != "" && !installDBNameRE.MatchString(v) {
			return "Invalid " + f.label + ": use letters, numbers and underscores only (up to 32)."
		}
	}
	if v := get("db_password"); v != "" && !ValidDBPassword(v) {
		return "Invalid database password: use up to 64 characters without spaces, quotes, backslashes or backticks."
	}
	if v := strings.TrimSpace(get("admin_email")); v != "" && !ValidEmail(v) {
		return "Invalid admin email address."
	}
	for _, key := range []string{"admin_username", "admin_login", "admin_user"} {
		if v := strings.TrimSpace(get(key)); v != "" && (!installUserRE.MatchString(v) || strings.HasPrefix(v, "-")) {
			return "Invalid admin username: use letters, numbers and . _ @ - only, not starting with -."
		}
	}
	if v := get("admin_password"); v != "" {
		// installers pass it as its own CLI argument, a leading - reads as an option (#1207)
		if strings.ContainsAny(v, " \t\r\n") || strings.HasPrefix(v, "-") || len(v) > 128 {
			return "Invalid admin password: no spaces, and it can not start with -."
		}
		if strength > 0 && !validators.IsPasswordStrongEnough(v, strength) {
			return "Admin password does not meet the required strength."
		}
	}
	return ""
}
