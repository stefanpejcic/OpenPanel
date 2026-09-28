package account

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"strings"
	"time"
)

// weakpassScript is opencli's shared check, mounted read-only into the container
var weakpassScript = "/usr/local/opencli/lib/weakpass.sh"

// isCommonPassword is a var so tests can stub it
var isCommonPassword = func(ctx context.Context, password string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", weakpassScript)
	cmd.Stdin = strings.NewReader(password + "\n")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true
	}
	// any other failure (script missing, timeout) lets the password through, same as opencli when the list can't be fetched
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
		if err := exec.CommandContext(ctx, "bash", weakpassScript, "--update").Run(); err != nil {
			log.Printf("ACCOUNT - weakpass list prefetch failed: %v", err)
		}
	}()
}
