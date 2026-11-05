package docker

import (
	"context"
	"fmt"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// ComposeContainer is a small start/stop/status/restart helper other modules use directly on the shared services (redis, memcached, mysql, postgres, phpmyadmin, valkey) rather than through the docker-compose.yml edit/save flow the rest of this package uses, exported here since internal/app can't import this package - "restart" is fire-and-forget, so this reports whether it launched, not whether it finished.
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
			// fire-and-forget: use context.Background() rather than ctx, since the request context is canceled as soon as the HTTP response is written and would otherwise kill the process mid-restart
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

// StartComposeServiceIfNotRunning resolves the "sql"/"ws" meta-names to the actual configured MySQL/webserver service (also syncing the MySQL root password into my.cnf when starting "sql"), then activates it via StartOrStopContainer - but only when it isn't already running. `podman-compose up -d` on an already-running container isn't always a no-op, and needlessly recreating a live database container drops its unix socket for a few seconds, breaking any query issued in that window - see the "postgres socket not found" class of bug this guards against.
func StartComposeServiceIfNotRunning(ctx context.Context, userContext, serviceName string) {
	switch serviceName {
	case "sql":
		serviceName, _ = GetEnvValue(userContext, "MYSQL_TYPE")
		rootPass, _ := GetEnvValue(userContext, "MYSQL_ROOT_PASSWORD")
		_ = syncMyCnfPassword(homePath(userContext, "my.cnf"), rootPass)
	case "ws":
		serviceName, _ = GetEnvValue(userContext, "WEB_SERVER")
	}
	if GetContainerStatus(ctx, userContext, serviceName).State == "running" {
		return
	}
	StartOrStopContainer(ctx, userContext, serviceName, "activate", "")
}

// syncMyCnfPassword sets the password= lines in my.cnf, done in Go because the value comes from the account's .env and must never be read as a sed script
func syncMyCnfPassword(path, password string) error {
	if strings.ContainsAny(password, "\r\n") {
		return fmt.Errorf("invalid password")
	}
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(line, "password=") && line != "password="+password {
			lines[i] = "password=" + password
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return tenantfs.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}
