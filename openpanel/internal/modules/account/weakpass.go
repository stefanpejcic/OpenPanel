package account

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// opencli is mounted read-only into the container, the check is `opencli weakpass`
var opencliBin = "/usr/local/opencli/opencli"

func weakpassCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, opencliBin, append([]string{"weakpass"}, args...)...)
	// skips opencli's log, hooks and cluster routing, none of them exist in the container
	cmd.Env = append(os.Environ(), "OPENCLI_RAW=1")
	return cmd
}

// isCommonPassword is a var so tests can stub it
var isCommonPassword = func(ctx context.Context, password string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := weakpassCmd(ctx)
	cmd.Stdin = strings.NewReader(password + "\n")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true
	}
	// any other failure (opencli missing, timeout) lets the password through, same as opencli when the list can't be fetched
	if err != nil {
		log.Printf("ACCOUNT - weakpass check skipped: %v", err)
	}
	return false
}

// warmWeakpassList fetches the list in the background so the first password change doesn't wait on the download
func warmWeakpassList() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := weakpassCmd(ctx, "--update").Run(); err != nil {
			log.Printf("ACCOUNT - weakpass list prefetch failed: %v", err)
		}
	}()
}
