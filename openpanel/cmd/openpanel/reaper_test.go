package main

import (
	"os/exec"
	"testing"
	"time"
)

func TestReapChildrenReportsPanelExit(t *testing.T) {
	other := exec.Command("true")
	panel := exec.Command("sh", "-c", "exit 3")
	if err := other.Start(); err != nil {
		t.Fatal(err)
	}
	if err := panel.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if code, done := reapChildren(panel.Process.Pid); done {
			if code != 3 {
				t.Fatalf("exit code %d, want 3", code)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("panel exit never reaped")
}
