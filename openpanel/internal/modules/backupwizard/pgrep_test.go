package backupwizard

import (
	"regexp"
	"testing"
)

func TestBackupProcessPatternMatchesOnlyThatAccount(t *testing.T) {
	re := regexp.MustCompile(backupProcessPattern("bob"))
	for cmd, want := range map[string]bool{
		"bash /usr/local/bin/opencli user-backup --account bob":         true,
		"bash /usr/local/bin/opencli user-backup --account bob --quiet": true,
		"bash /usr/local/bin/opencli user-backup --account bobby":       false,
		"bash /usr/local/bin/opencli user-backup --account xbob":        false,
		"bash /usr/local/bin/opencli user-backup --account alice":       false,
	} {
		if got := re.MatchString(cmd); got != want {
			t.Errorf("%q: match=%v, want %v", cmd, got, want)
		}
	}
}
