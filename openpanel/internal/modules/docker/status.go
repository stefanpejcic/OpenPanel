// Package docker handles container status/control, compose-file and .env
// editing, MySQL/webserver switching, image management, log viewing, and
// the interactive web terminal.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// ContainerStatus holds a container's (state, health) pair. State is one
// of: created, restarting, running, removing, paused, exited, dead, or
// "not_found" if inspect failed (container doesn't exist, socket
// unreachable, etc.).
type ContainerStatus struct {
	State  string
	Health string
}

// GetContainerStatus inspects a container and returns its state and health.
func GetContainerStatus(ctx context.Context, userContext, serviceName string) ContainerStatus {
	argv := podmanmanager.PodmanArgv(userContext, "inspect",
		"--format", "{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
		serviceName)

	cmd := podmanmanager.Command(ctx, userContext, argv)
	out, err := cmd.Output()
	if err != nil {
		return ContainerStatus{State: "not_found", Health: "none"}
	}

	parts := strings.Fields(string(out))
	if len(parts) == 0 {
		return ContainerStatus{State: "not_found", Health: "none"}
	}
	health := "none"
	if len(parts) > 1 {
		health = parts[1]
	}
	return ContainerStatus{State: parts[0], Health: health}
}

// IsServiceRunning reports whether a container by that name is currently
// running.
func IsServiceRunning(ctx context.Context, userContext, serviceName string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "ps",
		"--filter", "name="+serviceName,
		"--filter", "status=running",
		"--format", "{{.Names}}")

	cmd := podmanmanager.Command(ctx, userContext, argv)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// StartStopResult is the outcome of a start/stop/restart operation.
type StartStopResult struct {
	Success bool
	Message string
}

// StartOrStopContainer starts, stops, or restarts a compose service. flag
// may be "" (wait for completion), "detached" (fire-and-forget), or "pull"
// (only meaningful with action="activate": adds --pull).
func StartOrStopContainer(ctx context.Context, userContext, containerName, action, flag string) StartStopResult {
	argv, dir, ok := podmanmanager.BuildComposeUpDownCommand(userContext, containerName, action)
	if !ok {
		return StartStopResult{Success: false, Message: fmt.Sprintf("Invalid action: %s", action)}
	}
	if action == "activate" && flag == "pull" {
		argv = insertAfter(argv, "up", "--pull")
	}

	if flag == "detached" {
		// Deliberately NOT exec.CommandContext(ctx, ...): ctx is the HTTP
		// request's context, which net/http cancels once the response is
		// written - exec.CommandContext kills the child on cancellation,
		// so the compose-up would get SIGKILLed moments after starting,
		// often before the container is even up. context.Background()
		// decouples the child process from the request lifecycle so it
		// keeps running as a true fire-and-forget background job.
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Dir = dir
		cmd.Env = podmanmanager.PodmanEnv(userContext)
		if err := cmd.Start(); err != nil {
			return StartStopResult{Success: false, Message: "Unexpected error. Contact Administrator."}
		}
		go func() { _ = cmd.Wait() }()
		return StartStopResult{Success: true, Message: fmt.Sprintf("Container '%s' %s in background.", containerName, action)}
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = podmanmanager.PodmanEnv(userContext)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return StartStopResult{Success: false, Message: "Command failed with error. Please try again."}
		}
		return StartStopResult{Success: false, Message: "Unexpected error. Contact Administrator."}
	}

	msg := stdout.String()
	if msg == "" {
		msg = fmt.Sprintf("Container '%s' %sd successfully.", containerName, action)
	}
	return StartStopResult{Success: true, Message: msg}
}

func insertAfter(argv []string, after, insert string) []string {
	for i, a := range argv {
		if a == after {
			result := make([]string, 0, len(argv)+1)
			result = append(result, argv[:i+1]...)
			result = append(result, insert)
			result = append(result, argv[i+1:]...)
			return result
		}
	}
	return argv
}

// allowedVolumes is the set of data volumes deleteDockerVolume is allowed
// to remove.
func allowedVolumes(userContext string) map[string]bool {
	return map[string]bool{
		userContext + "_mysql_data": true,
		userContext + "_pg_data":    true,
	}
}

// deleteDockerVolume only ever removes one of the two data volumes it
// itself owns, never an arbitrary name.
func deleteDockerVolume(ctx context.Context, userContext, volumeName string) bool {
	if !allowedVolumes(userContext)[volumeName] {
		return false
	}
	argv := podmanmanager.PodmanArgv(userContext, "volume", "rm", volumeName)
	cmd := podmanmanager.Command(ctx, userContext, argv)
	return cmd.Run() == nil
}

// removeImage force-removes a container image by name, ignoring errors.
func removeImage(ctx context.Context, userContext, imageName string) {
	argv := podmanmanager.PodmanArgv(userContext, "rmi", "-f", imageName)
	cmd := podmanmanager.Command(ctx, userContext, argv)
	_ = cmd.Run()
}
