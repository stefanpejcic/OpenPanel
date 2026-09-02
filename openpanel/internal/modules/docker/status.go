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

// waitForServiceRunningAttempts/waitForServiceRunningInterval bound how
// long WaitForServiceRunning polls before giving up.
const (
	waitForServiceRunningAttempts = 15
	waitForServiceRunningInterval = 2 * time.Second
)

// WaitForServiceRunning polls IsServiceRunning until it reports true or the
// poll budget is exhausted. A container whose entrypoint does setup work
// before the real process starts (e.g. varnish copies and rewrites its VCL
// file before exec'ing varnishd), or whose image still needs to be pulled,
// can take a few seconds to reach State.Running=true after `podman-compose
// up` returns - checking once immediately is a false negative, not a
// slow-but-correct one.
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

// searchEngineFix maps a search-engine service's container name to the
// ownership (in the CONTAINER's own uid/gid namespace, matching its
// image's baked-in default user) that its dedicated data/config/logs
// volumes (see the compose template) need to end up under.
var searchEngineFix = map[string]struct {
	uid, gid int
	volumes  []string
}{
	"elasticsearch": {1000, 0, []string{"es_data", "es_config", "es_logs"}},
	"opensearch":    {1000, 1000, []string{"opensearch_data", "opensearch_config", "opensearch_logs"}},
}

// idMapEntry is one [ContainerID, ContainerID+Size) -> HostID range from
// `podman info`'s Host.IDMappings.{UID,GID}Map, translating this tenant's
// own user-namespace ids to real host ids.
type idMapEntry struct {
	ContainerID int `json:"container_id"`
	HostID      int `json:"host_id"`
	Size        int `json:"size"`
}

// mapToHostID translates a container-namespace id to its real host id
// using entries from `podman info`'s IDMappings, or ok=false if id falls
// outside every mapped range.
func mapToHostID(entries []idMapEntry, id int) (hostID int, ok bool) {
	for _, e := range entries {
		if id >= e.ContainerID && id < e.ContainerID+e.Size {
			return e.HostID + (id - e.ContainerID), true
		}
	}
	return 0, false
}

// fixSearchEngineOwnership works around a shared-image-store limitation
// that leaves elasticsearch/opensearch permanently crash-looping in
// "starting": this host's rootless podman pulls every image once into a
// read-only store shared across every tenant (additionalimagestores in
// storage.conf) rather than once per tenant, and that shared store can't
// be UID-shifted per tenant for image-baked non-root ownership - a path
// baked in as uid 1000 shows up as the kernel's unmappable-uid fallback
// in every tenant's own user namespace, and a non-root container uid gets
// a flat permission denied against it (both images also explicitly
// refuse to run as root - "can not run elasticsearch as root" - so
// that's not a usable workaround either).
//
// The compose template gives both services their own dedicated
// data/config/logs volumes (rather than leaving config/logs as
// image-baked paths, which only data/ would normally need) specifically
// so this can be fixed with a plain chown: Podman auto-populates a named
// volume from the image path the first time it's used (so the default
// config files are still there), but the volume itself is a plain host
// directory - unlike a container's own overlay mount, which lives in a
// private per-tenant mount namespace no chown from outside it can reach
// (confirmed live: even real root, from outside that namespace, sees the
// mountpoint's pre-mount contents, not the actual overlay; fixing that
// would need `podman mount`/`unshare` to run as the tenant's own Linux
// login, which isn't reachable from here - this binary runs in its own
// container with no account for any tenant and no local runtime dir for
// their rootless podman, only the remote API socket via CONTAINER_HOST).
//
// So the whole fix stays on that same remote API this package already
// uses everywhere else: resolve the tenant's container-uid -> host-uid
// mapping via `podman info`, resolve each volume's real host path via
// `podman volume inspect`, and chown it directly - no unshare, su, or
// local execution anywhere in this call.
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

// ForceRemoveContainer stops and removes exactly one container by name via
// plain `podman rm -f`, deliberately bypassing `podman-compose down` -
// that subcommand takes a service name but, per its own implementation
// (compose_down in podman_compose.py), doesn't scope to it: it tears down
// that service's full transitive depends_on chain too. The vendored
// compose template chains varnish -> webserver -> php-fpm, so
// `podman-compose down varnish` (or `down <webserver>`) also removes
// php-fpm - confirmed live, that's what left the webserver container
// completely gone after a Varnish enable/disable, since the recreate step
// deliberately uses `up -d --no-deps` (to avoid restarting an
// already-running php-fpm unnecessarily) and so never brings php-fpm back,
// and the webserver's own `--requires=<php-fpm>` then makes its own
// recreation fail outright when that dependency doesn't exist. Plain
// `podman rm -f` has no concept of compose dependencies, so it can only
// ever affect the one container named.
func ForceRemoveContainer(ctx context.Context, userContext, containerName string) {
	argv := podmanmanager.PodmanArgv(userContext, "rm", "-f", containerName)
	_ = podmanmanager.Command(ctx, userContext, argv).Run()
}
