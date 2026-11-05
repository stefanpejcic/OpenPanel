package web

import "fmt"

// DomainPick is one row of the shared domain list on pages that act on a single domain (DNS, SSL, logs, vhosts...)
type DomainPick struct {
	Domain string
	// Detail is plain data for the second column (a path, a size, a date)
	Detail string
	// Badge is a msgid shown as a status badge, BadgeArg is appended untranslated (a date)
	Badge    string
	BadgeArg string
	// Tone picks the badge colour: neutral, success, warning, error
	Tone string
	// Disabled greys out the row action, Note (a msgid) says why
	Disabled bool
	Note     string
}

// FormatBytes renders a byte count as "1.2 MB".
func FormatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	size := float64(n)
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if size < 1024 {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%.1f TB", size)
}
