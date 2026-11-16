package backups

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func TestRestoreProgressPhasesPercentAndSteps(t *testing.T) {
	clock := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	p := newRestoreProgress(100)
	p.now = func() time.Time { return clock }
	p.started, p.stepFrom = clock, clock

	if v := p.view(); v.Percent != -1 || v.ETA != -1 {
		t.Fatalf("before any phase: %+v", v)
	}
	p.setPhases(2)
	p.startPhase("download")
	clock = clock.Add(10 * time.Second)
	if _, err := io.Copy(io.Discard, p.reader(strings.NewReader(strings.Repeat("x", 50)))); err != nil {
		t.Fatal(err)
	}
	v := p.view()
	if v.Percent != 25 || v.Rate != 5 || v.ETA != 30 || v.Phase != 1 || v.Phases != 2 {
		t.Errorf("half of phase 1 of 2: %+v", v)
	}

	p.startPhase("files")
	p.item()
	clock = clock.Add(5 * time.Second)
	p.setStep("apply", "")
	clock = clock.Add(time.Second)
	p.finish()
	v = p.view()
	if v.Items != 1 || len(v.Steps) != 3 || v.Steps[0] != (stepTiming{Step: "download", Seconds: 10}) || v.Steps[1].Seconds != 5 || v.Steps[2].Step != "apply" {
		t.Errorf("steps = %+v items = %d", v.Steps, v.Items)
	}

	var nilProgress *restoreProgress
	nilProgress.startPhase("x")
	nilProgress.setStep("x", "y")
	nilProgress.item()
	nilProgress.finish()
	if nilProgress.view() != nil || nilProgress.reader(strings.NewReader("")) == nil {
		t.Error("nil progress should be a no-op")
	}
}

func TestQueueShowsLiveProgressAndKeepsTimings(t *testing.T) {
	release := make(chan struct{})
	reading := make(chan struct{})
	runRestore = func(ctx context.Context, a *appctx.App, req restoreRequest) (int, map[string]any) {
		p := progressFrom(ctx)
		p.startPhase("reading")
		_, _ = io.Copy(io.Discard, p.reader(strings.NewReader(strings.Repeat("x", 400))))
		p.setStep("database", "wp")
		close(reading)
		<-release
		return http.StatusOK, map[string]any{"message": "ok"}
	}
	defer func() { runRestore = performRestore }()

	q := testQueue(t)
	archiveSize = func(string, string) int64 { return 1000 }
	q.enqueue(nil, restoreRequest{Username: "stefan", IP: "192.0.2.1", Form: url.Values{"backup_file": {"stefan-2026-10-15T10-58-27.tar.gz"}, "restore_target": {"database"}, "database": {"wp"}}})
	<-reading
	queue, _ := q.snapshot(nil, restoreRequest{Username: "stefan"})
	if len(queue) != 1 || queue[0].Progress == nil {
		t.Fatalf("running job without progress: %+v", queue)
	}
	if p := queue[0].Progress; p.Done != 400 || p.Size != 1000 || p.Percent != 40 || p.Step != "database" || p.Detail != "wp" {
		t.Errorf("live progress = %+v", p)
	}
	if queue[0].By != "stefan" || queue[0].IP != "192.0.2.1" || queue[0].Size != 1000 {
		t.Errorf("who/size = %q %q %d", queue[0].By, queue[0].IP, queue[0].Size)
	}
	close(release)
	waitIdle(t, q)

	_, logs := q.snapshot(nil, restoreRequest{Username: "stefan"})
	if len(logs) != 1 || logs[0].Progress == nil || len(logs[0].Progress.Steps) != 2 || logs[0].Progress.Step != "" {
		t.Fatalf("finished job timings = %+v", logs[0].Progress)
	}
}
