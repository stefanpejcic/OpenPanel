package docker

import (
	"context"
	"testing"
)

// These tests run without a real podman installation (this sandbox has
// none), so they only exercise the "podman unavailable/container doesn't
// exist" failure paths - which is real, meaningful behavior in its own
// right (e.g. a freshly-provisioned account with no mysqld container yet).

func TestGetContainerStatusNoPodman(t *testing.T) {
	status := GetContainerStatus(context.Background(), "someuser", "mysql")
	if status.State != "not_found" || status.Health != "none" {
		t.Errorf("GetContainerStatus() = %+v, want {not_found none}", status)
	}
}

func TestIsServiceRunningNoPodman(t *testing.T) {
	if IsServiceRunning(context.Background(), "someuser", "mysql") {
		t.Error("expected IsServiceRunning to be false when podman isn't available")
	}
}

func TestStartOrStopContainerInvalidAction(t *testing.T) {
	result := StartOrStopContainer(context.Background(), "someuser", "mysql", "bogus", "")
	if result.Success {
		t.Error("expected Success=false for an invalid action")
	}
	if result.Message != "Invalid action: bogus" {
		t.Errorf("Message = %q, want %q", result.Message, "Invalid action: bogus")
	}
}

func TestInsertAfter(t *testing.T) {
	got := insertAfter([]string{"podman-compose", "up", "-d", "mysql"}, "up", "--pull")
	want := []string{"podman-compose", "up", "--pull", "-d", "mysql"}
	if len(got) != len(want) {
		t.Fatalf("insertAfter() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("insertAfter()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestInsertAfterNoMatch(t *testing.T) {
	argv := []string{"podman-compose", "down", "mysql"}
	got := insertAfter(argv, "up", "--pull")
	if len(got) != len(argv) {
		t.Errorf("expected argv unchanged when 'up' isn't present, got %v", got)
	}
}
