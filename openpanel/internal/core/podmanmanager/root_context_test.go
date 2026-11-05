package podmanmanager

import (
	"strings"
	"testing"
)

func TestUserSocketNeverFallsBackToRoot(t *testing.T) {
	if s := PodmanUserSocket("no-such-account-xyz"); strings.Contains(s, "/user/0/") {
		t.Fatalf("missing home fell back to root's socket: %s", s)
	}
}

func TestPodmanEnvDropsRootContainerHost(t *testing.T) {
	t.Setenv("CONTAINER_HOST", "unix:///run/podman/podman.sock")
	for _, ctx := range []string{"", "default", "root"} {
		if !IsRootContext(ctx) {
			t.Fatalf("%q should be a root context", ctx)
		}
		for _, kv := range PodmanEnv(ctx) {
			if strings.HasPrefix(kv, "CONTAINER_HOST=") {
				t.Fatalf("root context %q kept %s", ctx, kv)
			}
		}
	}
}
