package podmanmanager

import (
	"strings"
	"testing"
)

func TestIsRootContext(t *testing.T) {
	for _, c := range []string{"", "default", "root"} {
		if !isRootContext(c) {
			t.Errorf("expected %q to be a root context", c)
		}
	}
	if isRootContext("someuser") {
		t.Error("expected 'someuser' to not be a root context")
	}
}

func TestPodmanArgv(t *testing.T) {
	got := PodmanArgv("someuser", "ps", "-a")
	want := []string{"podman", "--remote", "ps", "-a"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("PodmanArgv(someuser) = %v, want %v", got, want)
	}

	got = PodmanArgv("root", "ps", "-a")
	want = []string{"podman", "ps", "-a"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("PodmanArgv(root) = %v, want %v", got, want)
	}
}

func TestPodmanComposeArgvNeverHasRemote(t *testing.T) {
	got := PodmanComposeArgv("up", "-d", "mysql")
	for _, arg := range got {
		if arg == "--remote" {
			t.Errorf("PodmanComposeArgv must never include --remote, got %v", got)
		}
	}
	if got[0] != "podman-compose" {
		t.Errorf("expected argv[0] = podman-compose, got %q", got[0])
	}
}

func TestPodmanEnvRootContextStripsContainerHost(t *testing.T) {
	env := PodmanEnv("root")
	for _, kv := range env {
		if strings.HasPrefix(kv, "CONTAINER_HOST=") {
			t.Errorf("expected no CONTAINER_HOST for root context, got %q", kv)
		}
	}
}

func TestPodmanEnvUserContextSetsContainerHost(t *testing.T) {
	env := PodmanEnv("someuser")
	found := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "CONTAINER_HOST=unix:///hostfs/run/user/") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a CONTAINER_HOST=unix:///hostfs/... entry for a per-user context, got %v", env)
	}
}

func TestBuildComposeUpDownCommand(t *testing.T) {
	argv, dir, ok := BuildComposeUpDownCommand("someuser", "mysql", "activate")
	if !ok {
		t.Fatal("expected ok=true for 'activate'")
	}
	if dir != "/home/someuser" {
		t.Errorf("dir = %q, want /home/someuser", dir)
	}
	want := "podman-compose up -d mysql"
	if strings.Join(argv, " ") != want {
		t.Errorf("argv = %q, want %q", strings.Join(argv, " "), want)
	}

	_, _, ok = BuildComposeUpDownCommand("someuser", "mysql", "bogus")
	if ok {
		t.Error("expected ok=false for an unrecognized action")
	}
}

func TestBuildWPCLIBaseCommand(t *testing.T) {
	argv := BuildWPCLIBaseCommand("someuser", "php")
	if argv[0] != "podman" || argv[len(argv)-1] != "/usr/local/bin/wp" {
		t.Errorf("unexpected argv shape: %v", argv)
	}
}
