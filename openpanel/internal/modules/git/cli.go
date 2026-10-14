package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// RunCLI is `openpanel git <pull|deploy> <context> <repo>`, how OpenAdmin runs repo actions without its own copy of the deploy code
func RunCLI(args []string, out io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(out, "usage: openpanel git <pull|deploy> <context> <repo>")
		return 2
	}
	action, userContext, name := args[0], args[1], args[2]
	msg, err := runCLIAction(context.Background(), action, userContext, name)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return 1
	}
	fmt.Fprintln(out, msg)
	return 0
}

func runCLIAction(ctx context.Context, action, userContext, name string) (string, error) {
	if userContext == "" || strings.ContainsAny(userContext, "/\\.") {
		return "", errors.New("invalid account")
	}
	if !isValidRepoName(name) {
		return "", errors.New("invalid repository name")
	}
	repo, ok := getRepo(userContext, name)
	if !ok {
		return "", errors.New("repository " + name + " not found")
	}
	switch action {
	case "pull":
		if repo.Source != SourceRemote {
			return "", errors.New(name + " is a local repository, there is nothing to pull")
		}
		if err := pullAndRecord(ctx, userContext, repo); err != nil {
			return "", fmt.Errorf("pulling %s failed: %w", name, err)
		}
		if repo.DeployMode == DeployAutomatic {
			return "pulled and deployed " + name, nil
		}
		return "pulled " + name, nil
	case "deploy":
		if repo.DeployMode == DeployNone {
			return "", errors.New("deployment is turned off for " + name)
		}
		if err := deployAndRecord(ctx, userContext, repo); err != nil {
			return "", fmt.Errorf("deploying %s failed: %w", name, err)
		}
		return "deployed " + name + " to " + repo.TargetDir, nil
	}
	return "", errors.New("unknown action " + action)
}
