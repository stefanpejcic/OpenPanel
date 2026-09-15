package crons

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// runWsUpgrader mirrors docker/terminal.go's wsUpgrader - same-origin-only (no CheckOrigin override) since this streams command output from inside a user's container
var runWsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// cronRunTimeout caps how long a manual "Run" can execute before being killed
func cronRunTimeout(a *appctx.App) time.Duration {
	n, err := strconv.Atoi(a.Config.Get("cron_run_timeout_seconds", "120"))
	if err != nil || n <= 0 {
		n = 120
	}
	return time.Duration(n) * time.Second
}

// findCronJob looks up the exact job the "Run" button was clicked for, so the websocket only ever runs a command already saved in the user's own crons.ini
func findCronJob(jobs []CronJob, comment, schedule, container, command string) (CronJob, bool) {
	for _, j := range jobs {
		if j.Comment == comment && j.Schedule == schedule && j.Container == container && j.Command == command {
			return j, true
		}
	}
	return CronJob{}, false
}

// handleCronjobsRunWS upgrades to a websocket and streams the live output of a manual "Run now" of one cron job, executed non-interactively inside its configured container
func handleCronjobsRunWS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	content, readErr := os.ReadFile(cronFilePath(userContext))
	if readErr != nil {
		http.Error(w, "cron job not found", http.StatusNotFound)
		return
	}
	job, found := findCronJob(ParseCronFile(string(content)), q.Get("comment"), q.Get("schedule"), q.Get("container"), q.Get("command"))
	if !found {
		http.Error(w, "cron job not found", http.StatusNotFound)
		return
	}

	conn, err := runWsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = logger.RecordUserAction(a.Config, username, "manually ran cron job "+job.Comment, reqip.ClientIP(r))

	runCtx, cancel := context.WithTimeout(context.Background(), cronRunTimeout(a))
	defer cancel()

	// keep reading so gorilla processes control frames, and bail out early if the browser closes the modal/tab
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	argv := podmanmanager.PodmanArgv(userContext, "exec", job.Container, "sh", "-c", job.Command)
	cmd := podmanmanager.Command(runCtx, userContext, argv)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var writeMu sync.Mutex
	writeOutput := func(data string) bool {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(map[string]any{"type": "output", "data": data}) == nil
	}

	if startErr := cmd.Start(); startErr != nil {
		_ = conn.WriteJSON(map[string]any{"type": "exit", "code": -1, "error": startErr.Error()})
		return
	}

	pump := func(reader io.Reader, wg *sync.WaitGroup) {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, readErr := reader.Read(buf)
			if n > 0 && !writeOutput(string(buf[:n])) {
				cancel()
				return
			}
			if readErr != nil {
				return
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go pump(stdout, &wg)
	go pump(stderr, &wg)
	wg.Wait()

	exitCode := 0
	if waitErr := cmd.Wait(); waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	writeMu.Lock()
	_ = conn.WriteJSON(map[string]any{"type": "exit", "code": exitCode})
	writeMu.Unlock()
}
