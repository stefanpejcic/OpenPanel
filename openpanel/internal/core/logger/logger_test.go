package logger

import (
	"os"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// activityLogDir is a const pointing at a system path, so we can't redirect
// it in a test without touching real filesystem state; these tests only
// exercise it if writable (e.g. running as root in a container), and skip
// otherwise rather than failing the suite in a normal dev sandbox.
func TestRecordUserAction(t *testing.T) {
	testUser := "go-migration-test-user"
	if err := os.MkdirAll(activityLogDir+"/"+testUser, 0o755); err != nil {
		t.Skipf("%s not writable in this environment: %v", activityLogDir, err)
	}
	logFile := activityLogDir + "/" + testUser + "/activity.log"
	t.Cleanup(func() { os.RemoveAll(activityLogDir + "/" + testUser) })

	cfg := config.Config{}
	if err := RecordUserAction(cfg, testUser, "logged in", "1.2.3.4"); err != nil {
		t.Fatalf("RecordUserAction: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !strings.Contains(string(data), "1.2.3.4 User "+testUser+" logged in") {
		t.Errorf("log entry missing expected content: %q", data)
	}
}

func TestRecordUserActionStripsSuspendedPrefix(t *testing.T) {
	testUser := "go-migration-test-user2"
	if err := os.MkdirAll(activityLogDir+"/"+testUser, 0o755); err != nil {
		t.Skipf("%s not writable in this environment: %v", activityLogDir, err)
	}
	t.Cleanup(func() { os.RemoveAll(activityLogDir + "/" + testUser) })

	cfg := config.Config{}
	if err := RecordUserAction(cfg, "SUSPENDED_"+testUser, "logged in", "1.2.3.4"); err != nil {
		t.Fatalf("RecordUserAction: %v", err)
	}

	if _, err := os.Stat(activityLogDir + "/" + testUser + "/activity.log"); err != nil {
		t.Errorf("expected log file for stripped username %q: %v", testUser, err)
	}
}
