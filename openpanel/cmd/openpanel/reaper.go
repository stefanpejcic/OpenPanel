package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// set on the child so it doesn't try to become init again
const initChildEnv = "OPENPANEL_INIT_CHILD"

// shouldRunAsInit is true when we're pid 1 in the container, where nobody else reaps the orphans opencli and podman-compose leave behind
func shouldRunAsInit() bool {
	return os.Getpid() == 1 && os.Getenv(initChildEnv) == ""
}

// runAsInit starts the real panel as a child, forwards signals to it and reaps every exited process, returning the panel's exit code
func runAsInit() int {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	sigs := make(chan os.Signal, 16)
	signal.Notify(sigs, syscall.SIGCHLD, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	cmd := exec.Command(exe)
	cmd.Args = os.Args
	cmd.Env = append(os.Environ(), initChildEnv+"=1")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		fatalLogger.Printf("BOOTSTRAP - failed to start panel process: %v", err)
		return 1
	}
	child := cmd.Process.Pid

	for sig := range sigs {
		if sig != syscall.SIGCHLD {
			_ = syscall.Kill(child, sig.(syscall.Signal))
			continue
		}
		if code, done := reapChildren(child); done {
			return code
		}
	}
	return 0
}

// reapChildren waits on everything that has exited and reports the panel's exit code once it's among them
func reapChildren(child int) (int, bool) {
	for {
		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
		if err == syscall.EINTR {
			continue
		}
		if err != nil || pid <= 0 {
			return 0, false
		}
		if pid != child {
			continue
		}
		if ws.Signaled() {
			return 128 + int(ws.Signal()), true
		}
		return ws.ExitStatus(), true
	}
}
