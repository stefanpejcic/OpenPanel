package docker

import (
	"context"
	"fmt"
	"os/exec"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// ComposeContainer is a small start/stop/status/restart helper other
// modules use directly on the shared services (redis, memcached, mysql,
// postgres, phpmyadmin, valkey) rather than through the docker-compose.yml
// edit/save flow the rest of this package uses. Exported here (rather
// than internal/app, which can't import this package - see helpers.go's
// package comment) so those callers don't have to duplicate it.
//
// "restart" is fire-and-forget: the underlying process is started but not
// waited on, so this reports whether the process launched, not whether it
// finished.
func ComposeContainer(ctx context.Context, userContext, name, action string) bool {
	switch action {
	case "start", "stop", "restart":
		var sub []string
		switch action {
		case "start":
			sub = []string{"up", "-d"}
		case "stop":
			sub = []string{"down"}
		case "restart":
			sub = []string{"restart"}
		}
		argv := podmanmanager.PodmanComposeArgv(append([]string{"-f", homePath(userContext, "docker-compose.yml")}, append(sub, name)...)...)

		if action == "restart" {
			// Fire-and-forget: use context.Background() rather than ctx,
			// since the request context is canceled as soon as the HTTP
			// response is written and would otherwise kill the process
			// mid-restart.
			cmd := podmanmanager.Command(context.Background(), userContext, argv)
			if err := cmd.Start(); err != nil {
				return false
			}
			go func() { _ = cmd.Wait() }() // reap, avoid a zombie process
			return true
		}

		cmd := podmanmanager.Command(ctx, userContext, argv)
		return cmd.Run() == nil
	case "status":
		argv := podmanmanager.PodmanArgv(userContext, "ps", "-q", "-f", "name="+name)
		out, err := podmanmanager.Command(ctx, userContext, argv).Output()
		return err == nil && len(out) > 0
	default:
		return false
	}
}

// StartComposeServiceIfNotRunning resolves the "sql"/"ws" meta-names to
// the actual configured MySQL/webserver service (also syncing the MySQL
// root password into my.cnf when starting "sql"), then activates it via
// StartOrStopContainer - but only when it isn't already running.
// `podman-compose up -d` on an already-running container isn't always the
// no-op its name implies (e.g. an ephemeral host port in the compose file
// can look like config drift), and needlessly recreating a live database
// container drops its unix socket for the few seconds it takes to come
// back up, breaking any query issued in that window - see the "postgres
// socket not found" class of bug this guards against.
func StartComposeServiceIfNotRunning(ctx context.Context, userContext, serviceName string) {
	switch serviceName {
	case "sql":
		serviceName, _ = GetEnvValue(userContext, "MYSQL_TYPE")
		rootPass, _ := GetEnvValue(userContext, "MYSQL_ROOT_PASSWORD")
		myCnfPath := homePath(userContext, "my.cnf")
		_ = exec.CommandContext(ctx, "sed", "-i",
			fmt.Sprintf("s/^password=.*/password=%s/", rootPass), myCnfPath).Run()
	case "ws":
		serviceName, _ = GetEnvValue(userContext, "WEB_SERVER")
	}
	if GetContainerStatus(ctx, userContext, serviceName).State == "running" {
		return
	}
	StartOrStopContainer(ctx, userContext, serviceName, "activate", "")
}
