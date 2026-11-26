package tenantfs

import (
	"os/exec"
	"slices"
	"testing"
)

func TestLowPrio(t *testing.T) {
	name, args := LowPrio(true, "find", ".", "-printf", "%h\n")
	all := append([]string{name}, args...)
	if !slices.Equal(all[len(all)-4:], []string{"find", ".", "-printf", "%h\n"}) {
		t.Fatalf("command not kept at the end: %v", all)
	}
	if _, err := exec.LookPath("ionice"); err == nil && !slices.Contains(all, "-c3") {
		t.Fatalf("idle class missing: %v", all)
	}
	// du of one folder only gets nice, the user is waiting on it
	_, args = LowPrio(false, "du", "-sh")
	if slices.Contains(args, "ionice") {
		t.Fatalf("ionice without idle: %v", args)
	}
}
