// Package processmanager lists the processes running inside every
// container of a user's stack (via `podman top`) and lets the user
// terminate one.
package processmanager

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// topDescriptors are the fields requested from `podman top`.
var topDescriptors = []string{"uid", "pid", "ppid", "pcpu", "stime", "tty", "time", "args"}

// timeFieldRE matches the elapsed-TIME field ("7s", "23m13s", "1h2m3s") or
// classic "HH:MM:SS" - used to re-anchor field parsing since STIME isn't
// reliably a single token (see getPodmanProcesses).
var timeFieldRE = regexp.MustCompile(`^(?:(?:\d+h)?(?:\d+m)?\d+s|\d+:\d{2}(?::\d{2})?)$`)

// Process is one row of `podman top` output, tagged with its container.
type Process struct {
	Container string `json:"Container"`
	UID       string `json:"UID"`
	PID       string `json:"PID"`
	PPID      string `json:"PPID"`
	C         string `json:"C"`
	STIME     string `json:"STIME"`
	TTY       string `json:"TTY"`
	TIME      string `json:"TIME"`
	CMD       string `json:"CMD"`
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

// serviceNamesFromCompose extracts every service name from a parsed
// docker-compose.yml.
func serviceNamesFromCompose(compose map[string]any) []string {
	services, ok := compose["services"].(map[string]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	return names
}

// getPodmanProcesses runs `podman top` against every running service
// container in the user's compose stack and returns the combined,
// PID-sorted process list.
func getPodmanProcesses(ctx context.Context, userContext string) ([]Process, error) {
	compose, err := podmanmanager.LoadComposeConfig(ctx, userContext)
	if err != nil {
		return nil, err
	}
	serviceNames := serviceNamesFromCompose(compose)

	var processes []Process

	for _, name := range serviceNames {
		if !docker.IsServiceRunning(ctx, userContext, name) {
			continue
		}

		argv := podmanmanager.PodmanArgv(userContext, append([]string{"top", name}, topDescriptors...)...)
		out, cmdErr := podmanmanager.Command(ctx, userContext, argv).Output()
		if cmdErr != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		if len(lines) < 2 {
			continue
		}

		// First line is the header (UID PID PPID %CPU STIME TTY TIME COMMAND); skip it.
		for _, line := range lines[1:] {
			stripped := strings.TrimSpace(line)
			if stripped == "" {
				continue
			}

			tokens := strings.Fields(stripped)
			if len(tokens) < 7 {
				continue
			}
			if _, digitErr := strconv.Atoi(tokens[1]); digitErr != nil {
				continue
			}

			timeIndex := -1
			for i := 6; i < len(tokens)-1; i++ {
				if timeFieldRE.MatchString(tokens[i]) {
					timeIndex = i
					break
				}
			}
			if timeIndex == -1 {
				continue
			}

			processes = append(processes, Process{
				Container: name,
				UID:       tokens[0],
				PID:       tokens[1],
				PPID:      tokens[2],
				C:         tokens[3],
				STIME:     strings.Join(tokens[4:timeIndex-1], " "),
				TTY:       tokens[timeIndex-1],
				TIME:      tokens[timeIndex],
				CMD:       strings.Join(tokens[timeIndex+1:], " "),
			})
		}
	}

	sort.SliceStable(processes, func(i, j int) bool {
		pi, _ := strconv.Atoi(processes[i].PID)
		pj, _ := strconv.Atoi(processes[j].PID)
		return pi < pj
	})

	return processes, nil
}

// isDisplayableCmd filters out entrypoint/healthcheck noise rows that
// aren't useful to show the user.
func isDisplayableCmd(cmd string) bool {
	return !strings.Contains(cmd, "/etc/entrypoint.sh") &&
		!strings.Contains(cmd, "ps -eo pid,%cpu,time,cmd") &&
		!strings.Contains(cmd, "/dev/null")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func processKey(container, pid string) string {
	return container + "\x00" + pid
}
