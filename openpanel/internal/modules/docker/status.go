package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// ContainerStatus holds a container's (state, health) pair. State is one of: created, restarting, running, removing, paused, exited, dead, or "not_found" if inspect failed (container doesn't exist, socket unreachable, etc.)
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

// IsServiceRunning reports whether a container by that name is currently running
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

// waitForServiceRunningAttempts/waitForServiceRunningInterval bound how long WaitForServiceRunning polls before giving up
const (
	waitForServiceRunningAttempts = 15
	waitForServiceRunningInterval = 2 * time.Second
)

// WaitForServiceRunning polls IsServiceRunning until it's true or the poll budget runs out. A container whose entrypoint does setup work before the real process starts (e.g. varnish rewrites its VCL before exec'ing varnishd), or whose image still needs pulling, can take a few seconds to reach State.Running=true after `podman-compose up` returns - checking only once would give a false negative.
func WaitForServiceRunning(ctx context.Context, userContext, serviceName string) bool {
	if IsServiceRunning(ctx, userContext, serviceName) {
		return true
	}
	for attempt := 0; attempt < waitForServiceRunningAttempts; attempt++ {
		time.Sleep(waitForServiceRunningInterval)
		if IsServiceRunning(ctx, userContext, serviceName) {
			return true
		}
	}
	return false
}

// StartStopResult is the outcome of a start/stop/restart operation.
type StartStopResult struct {
	Success bool
	Message string
}

// StartOrStopContainer starts, stops, or restarts a compose service. flag may be "" (wait for completion), "detached" (fire-and-forget), or "pull" (only meaningful with action="activate": adds --pull).
func StartOrStopContainer(ctx context.Context, userContext, containerName, action, flag string) StartStopResult {
	argv, dir, ok := podmanmanager.BuildComposeUpDownCommand(userContext, containerName, action)
	if !ok {
		return StartStopResult{Success: false, Message: fmt.Sprintf("Invalid action: %s", action)}
	}
	if action == "activate" && flag == "pull" {
		argv = insertAfter(argv, "up", "--pull")
	}

	if flag == "detached" {
		// deliberately not exec.CommandContext(ctx, ...): ctx is the HTTP request's context, canceled once the response is written, which would SIGKILL the compose-up moments after starting - plain exec.Command decouples the child from the request lifecycle so it actually runs in the background
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Dir = dir
		cmd.Env = podmanmanager.PodmanEnv(userContext)
		if err := cmd.Start(); err != nil {
			return StartStopResult{Success: false, Message: "Unexpected error. Contact Administrator."}
		}
		go func() {
			_ = cmd.Wait()
			if action == "activate" {
				fixSearchEngineOwnership(context.Background(), userContext, containerName)
			}
		}()
		return StartStopResult{Success: true, Message: fmt.Sprintf("Container '%s' %s in background.", containerName, action)}
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = podmanmanager.PodmanEnv(userContext)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if detail := strings.TrimSpace(stderr.String()); detail != "" {
				return StartStopResult{Success: false, Message: detail}
			}
			return StartStopResult{Success: false, Message: "Command failed with error. Please try again."}
		}
		return StartStopResult{Success: false, Message: "Unexpected error. Contact Administrator."}
	}

	if action == "activate" {
		fixSearchEngineOwnership(ctx, userContext, containerName)
	}

	msg := stdout.String()
	if msg == "" {
		msg = fmt.Sprintf("Container '%s' %sd successfully.", containerName, action)
	}
	return StartStopResult{Success: true, Message: msg}
}

// searchEngineFix maps a search-engine service's container name to the ownership (in the container's own uid/gid namespace, matching its image's baked-in default user) its dedicated data/config/logs volumes need to end up under
var searchEngineFix = map[string]struct {
	uid, gid int
	volumes  []string
}{
	"elasticsearch": {1000, 0, []string{"es_data", "es_config", "es_logs"}},
	"opensearch":    {1000, 1000, []string{"opensearch_data", "opensearch_config", "opensearch_logs"}},
}

// idMapEntry is one [ContainerID, ContainerID+Size) -> HostID range from `podman info`'s Host.IDMappings.{UID,GID}Map - translates a tenant's user-namespace ids to real host ids
type idMapEntry struct {
	ContainerID int `json:"container_id"`
	HostID      int `json:"host_id"`
	Size        int `json:"size"`
}

// mapToHostID translates a container-namespace id to its real host id using entries from `podman info`'s IDMappings, or ok=false if id falls outside every mapped range
func mapToHostID(entries []idMapEntry, id int) (hostID int, ok bool) {
	for _, e := range entries {
		if id >= e.ContainerID && id < e.ContainerID+e.Size {
			return e.HostID + (id - e.ContainerID), true
		}
	}
	return 0, false
}

// fixSearchEngineOwnership works around a shared-image-store limitation that leaves elasticsearch/opensearch stuck crash-looping in "starting": rootless podman pulls each image once into a store shared across every tenant, so it can't be UID-shifted per tenant, and the image-baked non-root uid gets a flat permission denied in every tenant's namespace.
// The compose template gives both services dedicated data/config/logs volumes, which are plain host directories (unlike the container's own overlay mount, which lives in a private mount namespace nothing outside it can chown) - so this resolves the tenant's container-uid -> host-uid mapping via `podman info`, finds each volume's real host path via `podman volume inspect`, and chowns it directly over the remote API, no unshare or local execution involved.
func fixSearchEngineOwnership(ctx context.Context, userContext, containerName string) {
	fix, ok := searchEngineFix[containerName]
	if !ok {
		return
	}

	var idMappings struct {
		UIDMap []idMapEntry
		GIDMap []idMapEntry
	}
	infoArgv := podmanmanager.PodmanArgv(userContext, "info", "--format", "{{json .Host.IDMappings}}")
	infoOut, err := podmanmanager.Command(ctx, userContext, infoArgv).Output()
	if err != nil || json.Unmarshal(infoOut, &idMappings) != nil {
		return
	}
	hostUID, ok := mapToHostID(idMappings.UIDMap, fix.uid)
	if !ok {
		return
	}
	hostGID, ok := mapToHostID(idMappings.GIDMap, fix.gid)
	if !ok {
		return
	}
	ownership := fmt.Sprintf("%d:%d", hostUID, hostGID)

	for _, volume := range fix.volumes {
		volArgv := podmanmanager.PodmanArgv(userContext, "volume", "inspect", userContext+"_"+volume, "--format", "{{.Mountpoint}}")
		mountOut, mountErr := podmanmanager.Command(ctx, userContext, volArgv).Output()
		if mountErr != nil {
			continue
		}
		mountpoint := strings.TrimSpace(string(mountOut))
		if mountpoint == "" {
			continue
		}
		_ = exec.CommandContext(ctx, "chown", "-R", ownership, mountpoint).Run()
	}
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

// allowedVolumes is the set of data volumes deleteDockerVolume is allowed to remove
func allowedVolumes(userContext string) map[string]bool {
	return map[string]bool{
		userContext + "_mysql_data": true,
		userContext + "_pg_data":    true,
	}
}

// deleteDockerVolume only ever removes one of the two data volumes it itself owns, never an arbitrary name
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

// ForceRemoveContainer stops and removes exactly one container by name via plain `podman rm -f`, deliberately bypassing `podman-compose down` - that subcommand tears down the whole depends_on chain too. The vendored template chains varnish -> webserver -> php-fpm, so `podman-compose down varnish` also removes php-fpm, and since the recreate step uses `up -d --no-deps` it never comes back, breaking the webserver's own recreation. Plain `podman rm -f` doesn't know about compose dependencies, so it only touches the one container named.
func ForceRemoveContainer(ctx context.Context, userContext, containerName string) {
	argv := podmanmanager.PodmanArgv(userContext, "rm", "-f", containerName)
	_ = podmanmanager.Command(ctx, userContext, argv).Run()
}
