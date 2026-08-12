package docker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// getCPUAndRAMForPlanID looks up a hosting plan's CPU/RAM limits, cached 30s.
func getCPUAndRAMForPlanID(a *appctx.App, ctx context.Context, planID int) (int, int) {
	type limits struct{ CPU, RAM int }
	result, _ := cache.Memoize(ctx, a.Cache, fmt.Sprintf("get_cpu_and_ram_for_plan_id:%d", planID), 30*time.Second, func() (limits, error) {
		var cpu, ram sql.NullString
		row := a.DB.QueryRowContext(ctx, "SELECT cpu, ram FROM plans WHERE id = ?", planID)
		if err := row.Scan(&cpu, &ram); err != nil {
			return limits{}, nil //nolint:nilerr // a missing plan yields (0, 0), not an error
		}
		totalCPU := atoiDefault(cpu.String, 0)
		totalRAM := atoiDefault(strings.TrimRight(ram.String, "gG"), 0)
		return limits{CPU: totalCPU, RAM: totalRAM}, nil
	})
	return result.CPU, result.RAM
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

// GetRunningContainers lists the names of currently running containers.
func GetRunningContainers(ctx context.Context, userContext string) ([]string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "ps", "--format", "{{.Names}}")
	cmd := podmanmanager.Command(ctx, userContext, argv)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []string{}, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// containerExists reports whether a container by that name can be inspected.
func containerExists(ctx context.Context, userContext, containerName string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "inspect", containerName)
	cmd := podmanmanager.Command(ctx, userContext, argv)
	return cmd.Run() == nil
}

// updateContainerRAMOrCPU updates a container's RAM or CPU limit, both in
// its .env entry and live via `podman update`. action is "ram" or "cpu".
// Returns (message, success).
func updateContainerRAMOrCPU(a *appctx.App, ctx context.Context, userContext string, planID int, containerName, action, providedValue string) (string, bool) {
	action = strings.ToLower(action)
	if action != "ram" && action != "cpu" {
		return fmt.Sprintf("Unsupported action: %s. Use 'ram' or 'cpu'.", action), false
	}

	envVar := ServiceKeyPrefix(containerName) + "_" + strings.ToUpper(action)

	totalCPU, totalRAM := getCPUAndRAMForPlanID(a, ctx, planID)

	var value string
	if providedValue == "0" {
		if action == "ram" {
			value = strconv.Itoa(totalRAM)
		} else {
			value = strconv.Itoa(totalCPU)
		}
	} else {
		value = providedValue
	}

	if action == "ram" && !strings.HasSuffix(strings.ToUpper(value), "G") {
		value += "G"
	}

	envPath := homePath(userContext, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Sprintf(".env file not found at %s", envPath), false
	}
	lines := strings.Split(string(data), "\n")
	prefix := envVar + "="
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = envVar + `="` + value + `"`
		}
	}
	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Sprintf("Failed to update .env: %s", err), false
	}

	if containerExists(ctx, userContext, containerName) {
		var argv []string
		if action == "ram" {
			argv = podmanmanager.PodmanArgv(userContext, "update", "--memory-swap", value, "--memory", value, containerName)
		} else {
			argv = podmanmanager.PodmanArgv(userContext, "update", "--cpus", value, containerName)
		}
		cmd := podmanmanager.Command(ctx, userContext, argv)
		cmd.Dir = homePath(userContext)
		if err := cmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return "Docker command failed. Try again.", false
			}
			return "Unexpected error. Contact Administrator.", false
		}
	}

	if providedValue == "0" || providedValue == "0G" {
		return fmt.Sprintf("%s limits for container %s removed and set to max available on the plan.", strings.ToUpper(action), containerName), true
	}
	return fmt.Sprintf("Max %s for container %s set to %s", strings.ToUpper(action), containerName, value), true
}

// RestartContainer stops and re-starts a single compose service via
// `podman-compose down`/`up -d --no-deps`.
func RestartContainer(ctx context.Context, userContext, containerName string) StartStopResult {
	downArgv, dir, _ := podmanmanager.BuildComposeUpDownCommand(userContext, containerName, "deactivate")
	down := podmanmanager.Command(ctx, userContext, downArgv)
	down.Dir = dir
	if err := down.Run(); err != nil {
		return StartStopResult{Success: false, Message: "Command failed with error. Please try again."}
	}

	upArgv, _, _ := podmanmanager.BuildComposeUpDownCommand(userContext, containerName, "activate")
	up := podmanmanager.Command(ctx, userContext, upArgv)
	up.Dir = dir
	if err := up.Run(); err != nil {
		return StartStopResult{Success: false, Message: "Command failed with error. Please try again."}
	}

	return StartStopResult{Success: true, Message: fmt.Sprintf("Container '%s' restarted successfully.", containerName)}
}

func handleContainersStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	if username == "" {
		writeJSONError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	running, err := GetRunningContainers(r.Context(), username)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Could not retrieve container status")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"running_containers": running})
}

// handleManageContainer mirrors manage_container(). action is one of
// "start"/"stop"/"restart"/"cpu"/"ram", fixed by the caller (docker.go
// registers one literal route per action, rather than a single
// "{action}/{container_name}" wildcard route, which would create an
// unresolvable net/http.ServeMux registration-time ambiguity against the
// other literal-prefixed /containers/* routes like /containers/image/).
func handleManageContainer(a *appctx.App, w http.ResponseWriter, r *http.Request, action string) {
	containerName := r.PathValue("container_name")
	ctx := r.Context()

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)
	if username == "" {
		writeJSONError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if action == "cpu" || action == "ram" {
		_ = r.ParseForm()
		value := r.Form.Get("value")
		planID, _ := injected["hosting_plan"].(int)
		message, success := updateContainerRAMOrCPU(a, ctx, userContext, planID, containerName, action, value)
		if success {
			_ = logger.RecordUserAction(a.Config, username, fmt.Sprintf("updated %s limit for %s to: %s", action, containerName, value), reqip.ClientIP(r))
		}
		flashAndRedirect(a, w, r, successCategory(success), message, "/containers")
		return
	}

	var response StartStopResult
	pullRequested := false

	switch action {
	case "restart":
		response = RestartContainer(ctx, userContext, containerName)
	case "start":
		_ = r.ParseForm()
		if r.Form.Get("pull") == "true" {
			pullRequested = true
			response = StartOrStopContainer(ctx, userContext, containerName, "activate", "pull")
		} else {
			response = StartOrStopContainer(ctx, userContext, containerName, "activate", "")
		}
	case "stop":
		response = StartOrStopContainer(ctx, userContext, containerName, "deactivate", "")
	}

	actionType := map[string]string{"start": "activate", "stop": "deactivate", "restart": "restart"}[action]

	if response.Success {
		var logMsg, flashMsg string
		switch {
		case pullRequested:
			logMsg = fmt.Sprintf("pulled image for %s and %sd", containerName, actionType)
			flashMsg = fmt.Sprintf("Image pulled and container %s %sd successfully.", containerName, actionType)
		case action == "restart":
			logMsg = fmt.Sprintf("restarted %s", containerName)
			flashMsg = fmt.Sprintf("Container %s restarted successfully.", containerName)
		default:
			logMsg = fmt.Sprintf("%sd %s", actionType, containerName)
			flashMsg = fmt.Sprintf("Container %s %sd successfully.", containerName, actionType)
		}
		_ = logger.RecordUserAction(a.Config, username, logMsg, reqip.ClientIP(r))
		flashAndRedirect(a, w, r, "success", flashMsg, "/containers")
		return
	}

	msg := response.Message
	if msg == "" {
		msg = "Error occurred!"
	}
	flashAndRedirect(a, w, r, "error", msg, "/containers")
}

func successCategory(success bool) string {
	if success {
		return "success"
	}
	return "error"
}
