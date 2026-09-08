// Package podmanmanager builds the right `podman`/`podman-compose` argv and environment for a hosting account's own rootless podman socket, versus the root/default context's own podman. CLI only, no REST client yet.
package podmanmanager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

// isRootContext reports whether userContext means "talk to the local/root podman directly" instead of a per-user rootless one
func isRootContext(userContext string) bool {
	switch userContext {
	case "", "default", "root":
		return true
	default:
		return false
	}
}

// GetUID returns the owning UID of /home/<context>, used to find that user's rootless podman socket
func GetUID(userContext string) (int, error) {
	info, err := os.Stat("/home/" + userContext)
	if err != nil {
		return 0, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("could not read UID for /home/%s", userContext)
	}
	return int(stat.Uid), nil
}

// PodmanUserSocket returns the unix:// URL for <context>'s rootless podman socket - the /hostfs prefix matters since this binary runs in a container with the host rootfs bind-mounted there
func PodmanUserSocket(userContext string) string {
	uid, err := GetUID(userContext)
	if err != nil {
		uid = 0
	}
	return fmt.Sprintf("unix:///hostfs/run/user/%d/podman/podman.sock", uid)
}

// PodmanEnv returns the process environment with CONTAINER_HOST set for per-user contexts and stripped for the local/root context
func PodmanEnv(userContext string) []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, "CONTAINER_HOST=") {
			filtered = append(filtered, kv)
		}
	}
	if !isRootContext(userContext) {
		filtered = append(filtered, "CONTAINER_HOST="+PodmanUserSocket(userContext))
	}
	return filtered
}

// PodmanArgv returns argv for a `podman` CLI call against <context>'s instance - root/default hits the local socket directly, per-user goes through --remote + CONTAINER_HOST
func PodmanArgv(userContext string, args ...string) []string {
	if isRootContext(userContext) {
		return append([]string{"podman"}, args...)
	}
	return append([]string{"podman", "--remote"}, args...)
}

// PodmanComposeArgv returns argv for a `podman-compose` call, never with --remote - podman-compose puts extra podman-args after the subcommand, which breaks a global-only flag like --remote, so CONTAINER_HOST via PodmanEnv() carries it instead
func PodmanComposeArgv(args ...string) []string {
	return append([]string{"podman-compose"}, args...)
}

// Command builds an *exec.Cmd for argv (from PodmanArgv/PodmanComposeArgv) with userContext's environment already applied
func Command(ctx context.Context, userContext string, argv []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = PodmanEnv(userContext)
	return cmd
}

// LoadComposeConfig runs `podman-compose config` for <context> and returns the merged compose file, parsing its default YAML output since it has no --format flag like docker-compose does
func LoadComposeConfig(ctx context.Context, userContext string) (map[string]any, error) {
	argv := PodmanComposeArgv("-f", "/home/"+userContext+"/docker-compose.yml", "config")
	cmd := Command(ctx, userContext, argv)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := yaml.Unmarshal(out, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BuildWPCLIBaseCommand returns the `podman exec ... php ... /usr/local/bin/wp` argv prefix shared by all WP-CLI calls; run it via Command(ctx, userContext, argv) to reach the right socket
func BuildWPCLIBaseCommand(userContext, phpContainer string) []string {
	return PodmanArgv(userContext, "exec", phpContainer,
		"php",
		"-d", "memory_limit=-1",
		"-d", "open_basedir=none",
		"-d", "disable_functions=",
		"-d", "display_errors=0",
		"-d", "error_log=/dev/null",
		"/usr/local/bin/wp",
	)
}

// BuildComposeUpDownCommand returns the `podman-compose up -d`/`down` argv and working directory for a container (ok is false for an unrecognized action) - "up" always passes --no-deps since callers already start real dependencies themselves, and the vendored podman-compose crashes with a KeyError resolving a `${VAR:-default}` depends_on entry otherwise, which looks like activate silently failed
func BuildComposeUpDownCommand(userContext, containerName, action string) (argv []string, dir string, ok bool) {
	dir = "/home/" + userContext
	switch action {
	case "activate":
		return PodmanComposeArgv("up", "-d", "--no-deps", containerName), dir, true
	case "deactivate":
		return PodmanComposeArgv("down", containerName), dir, true
	default:
		return nil, dir, false
	}
}
