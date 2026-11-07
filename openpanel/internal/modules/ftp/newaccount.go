package ftp

import (
	"net/url"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// ftpLocalRe is opencli ftp-add's rule for the part before @domain
var ftpLocalRe = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

var (
	ftpLowerRe = regexp.MustCompile(`[a-z]`)
	ftpUpperRe = regexp.MustCompile(`[A-Z]`)
	ftpDigitRe = regexp.MustCompile(`[0-9]`)
	ftpOtherRe = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

// newAccountProblem is the create form's username and password checks, the same ones opencli ftp-add runs after a slow container exec; "" when fine
func newAccountProblem(local, domain, password string, threshold int, existing []Account) string {
	if !ftpLocalRe.MatchString(local) {
		return "Only lower case letters, numbers, _ and - are allowed in the username, up to 32 characters."
	}
	full := strings.ToLower(local + "@" + domain)
	for _, acc := range existing {
		if strings.ToLower(acc.Username) == full {
			return "FTP account %(ftp_username)s already exists."
		}
	}
	if len(password) < 8 || len(password) > 30 {
		return "Password must be between 8 and 30 characters."
	}
	if !ftpLowerRe.MatchString(password) || !ftpUpperRe.MatchString(password) || !ftpDigitRe.MatchString(password) || !ftpOtherRe.MatchString(password) {
		return "Use at least one lower case letter, upper case letter, number and symbol."
	}
	if !validators.IsPasswordStrongEnough(password, threshold) {
		return "Password does not meet the required strength."
	}
	return ""
}

// newAccountBack is /ftp/new with what was typed, minus the password, so a rejected submit doesn't start over
func newAccountBack(local, domain, folder string) string {
	q := make([]string, 0, 3)
	if domain != "" {
		q = append(q, "domain="+url.QueryEscape(domain))
	}
	if local != "" {
		q = append(q, "username="+url.QueryEscape(local))
	}
	if folder != "" {
		q = append(q, "path="+url.QueryEscape(folder))
	}
	if len(q) == 0 {
		return "/ftp/new"
	}
	return "/ftp/new?" + strings.Join(q, "&")
}
