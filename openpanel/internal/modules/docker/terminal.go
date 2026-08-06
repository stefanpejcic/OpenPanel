package docker

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

const disableTerminalFlagFile = "/etc/openpanel/disable_openpanel_terminal_ui"

// handleDockerTerminal renders either the service picker (no
// container_name) or the terminal page itself.
func handleDockerTerminal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(disableTerminalFlagFile); err == nil {
		http.Error(w, "Web Terminal access is disabled.", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	containerName := r.PathValue("container_name")
	terminalTimeout := terminalCommandTimeout(a)

	if containerName != "" {
		renderTerminalPage(a, w, r, terminalTimeout, "Docker Terminal", containerName, nil)
		return
	}

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)

	var activeServiceNames []string
	if composeData, err := podmanmanager.LoadComposeConfig(ctx, userContext); err == nil {
		if services, ok := composeData["services"].(map[string]any); ok {
			for name := range services {
				if IsServiceRunning(ctx, userContext, name) {
					activeServiceNames = append(activeServiceNames, name)
				}
			}
		}
	}

	renderTerminalPage(a, w, r, terminalTimeout, "Select Docker service", "", activeServiceNames)
}

func terminalCommandTimeout(a *appctx.App) time.Duration {
	n, err := strconv.Atoi(a.Config.Get("terminal_timeout", "10"))
	if err != nil {
		n = 10
	}
	return time.Duration(n) * time.Second
}

var wsUpgrader = websocket.Upgrader{
	// Same-origin-only by default (no CheckOrigin override), so
	// cross-origin pages cannot open a terminal websocket.
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// handleDockerTerminalWS upgrades the request to a websocket and streams
// an interactive shell session inside the target container.
func handleDockerTerminalWS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(disableTerminalFlagFile); err == nil {
		http.Error(w, "Web Terminal access is disabled.", http.StatusForbidden)
		return
	}

	containerName := r.PathValue("container_name")

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, initRaw, err := conn.ReadMessage()
	if err != nil {
		return
	}
	var init struct {
		Shell string `json:"shell"`
		Rows  int    `json:"rows"`
		Cols  int    `json:"cols"`
	}
	if json.Unmarshal(initRaw, &init) != nil {
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	shell := init.Shell
	if shell != "bash" && shell != "sh" {
		shell = "sh"
	}
	rows, cols := init.Rows, init.Cols
	if rows <= 0 {
		rows = 24
	}
	if cols <= 0 {
		cols = 80
	}

	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	_ = logger.RecordUserAction(a.Config, username, "opened interactive terminal for service "+containerName, reqip.ClientIP(r))

	argv := podmanmanager.PodmanArgv(userContext, "exec", "-it", "-e", "TERM=xterm-256color", containerName, shell)
	runPTYSession(conn, argv, rows, cols, podmanmanager.PodmanEnv(userContext), terminalCommandTimeout(a))
}

// runPTYSession forks a PTY running argv, pumps its output to the
// websocket, and forwards websocket input (keystrokes, or
// {"type":"resize",...} control messages) to the PTY.
func runPTYSession(conn *websocket.Conn, argv []string, rows, cols int, extraEnv []string, readTimeout time.Duration) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = append(append([]string{}, os.Environ()...), "TERM=xterm-256color", "COLORTERM=truecolor")
	cmd.Env = append(cmd.Env, extraEnv...)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
	if err != nil {
		return
	}
	defer func() {
		_ = ptmx.Close()
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	done := make(chan struct{})
	var closeOnce sync.Once
	stop := func() {
		closeOnce.Do(func() { close(done) })
	}

	// pump: PTY output -> websocket (the sole writer to conn, matching
	// gorilla/websocket's one-writer-at-a-time requirement).
	go func() {
		defer stop()
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.TextMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// main loop: websocket input -> PTY (the sole reader of conn).
	for {
		select {
		case <-done:
			return
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			stop()
			return
		}

		if msgType == websocket.TextMessage && len(data) > 0 && data[0] == '{' {
			var msg struct {
				Type string `json:"type"`
				Rows int    `json:"rows"`
				Cols int    `json:"cols"`
			}
			if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" {
				_ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(msg.Rows), Cols: uint16(msg.Cols)})
				continue
			}
		}

		if _, err := ptmx.Write(data); err != nil {
			stop()
			return
		}
	}
}
