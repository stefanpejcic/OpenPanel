package docker

import (
	"context"
	"os/exec"
	"testing"
)

// A canceled request (client gone between down and up) must not stop the restart halfway
func TestRestartContainerOutlivesRequest(t *testing.T) {
	var calls []string
	var ctxErrs []error
	orig := restartCommand
	restartCommand = func(ctx context.Context, userContext string, argv []string) *exec.Cmd {
		calls = append(calls, argv[len(argv)-2]+" "+argv[len(argv)-1])
		ctxErrs = append(ctxErrs, ctx.Err())
		if _, ok := ctx.Deadline(); !ok {
			t.Error("restart commands should have a deadline")
		}
		return exec.CommandContext(ctx, "true")
	}
	t.Cleanup(func() { restartCommand = orig })

	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()
	// the commands run in /home/<context>, point that at a temp dir
	res := RestartContainer(reqCtx, ".."+t.TempDir(), "flaskapp")
	if !res.Success {
		t.Fatalf("restart failed: %+v", res)
	}
	if len(calls) != 2 {
		t.Fatalf("want down and up, got %v", calls)
	}
	for i, err := range ctxErrs {
		if err != nil {
			t.Errorf("command %d (%s) got a canceled context: %v", i, calls[i], err)
		}
	}
}
