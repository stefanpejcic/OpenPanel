package backups

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func testQueue(t *testing.T) *accountQueue {
	t.Helper()
	return &accountQueue{path: filepath.Join(t.TempDir(), "backup_restores.json")}
}

func waitIdle(t *testing.T, q *accountQueue) {
	t.Helper()
	for i := 0; i < 200; i++ {
		q.mu.Lock()
		idle := !q.working
		q.mu.Unlock()
		if idle {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("queue never finished")
}

func TestQueueRunsOneAtATimeInOrder(t *testing.T) {
	var running, maxRunning int32
	var mu sync.Mutex
	var order []string
	runRestore = func(ctx context.Context, a *appctx.App, req restoreRequest) (int, map[string]any) {
		n := atomic.AddInt32(&running, 1)
		for {
			m := atomic.LoadInt32(&maxRunning)
			if n <= m || atomic.CompareAndSwapInt32(&maxRunning, m, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		order = append(order, req.Form.Get("database"))
		mu.Unlock()
		atomic.AddInt32(&running, -1)
		if req.Form.Get("database") == "bad" {
			return http.StatusNotFound, map[string]any{"error": "Database 'bad' not found in this backup."}
		}
		return http.StatusOK, map[string]any{"success": true, "message": "ok"}
	}
	defer func() { runRestore = performRestore }()

	q := testQueue(t)
	for _, db := range []string{"one", "bad", "three"} {
		q.enqueue(nil, restoreRequest{Form: url.Values{"backup_file": {"u-2026-10-15T10-58-27.tar.gz"}, "restore_target": {"database"}, "database": {db}}})
	}
	waitIdle(t, q)

	if maxRunning != 1 {
		t.Errorf("%d restores ran at the same time, want 1", maxRunning)
	}
	if len(order) != 3 || order[0] != "one" || order[1] != "bad" || order[2] != "three" {
		t.Errorf("order = %v", order)
	}
	queue, logs := q.snapshot(nil)
	if len(queue) != 0 || len(logs) != 3 {
		t.Fatalf("queue %d, logs %d", len(queue), len(logs))
	}
	if logs[0].Database != "three" || logs[1].Status != "failed" || logs[1].Message == "" || logs[0].Created != "2026-10-15T10:58:27Z" {
		t.Errorf("logs = %+v", logs)
	}
}

func TestQueueMarksInterruptedJobsFailed(t *testing.T) {
	q := testQueue(t)
	content := `[{"id":"a","request":{"form":{"restore_target":["html"]}},"status":"running","queued":"2026-10-15T10:00:00Z"}]`
	if err := os.WriteFile(q.path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	queue, logs := q.snapshot(nil)
	if len(queue) != 0 || len(logs) != 1 || logs[0].Status != "failed" || logs[0].Finished == nil {
		t.Errorf("queue %+v logs %+v", queue, logs)
	}
}

func TestQueueKeepsLogSize(t *testing.T) {
	q := testQueue(t)
	now := time.Now().UTC()
	q.mu.Lock()
	q.loaded = true
	for i := 0; i < restoreLogSize+20; i++ {
		q.jobs = append(q.jobs, &restoreJob{ID: newJobID(), Status: "done", Finished: &now})
	}
	q.jobs = append(q.jobs, &restoreJob{ID: "waiting", Status: "queued"})
	q.save()
	q.mu.Unlock()
	if len(q.jobs) != restoreLogSize+1 || q.jobs[len(q.jobs)-1].ID != "waiting" {
		t.Errorf("kept %d jobs", len(q.jobs))
	}
}
