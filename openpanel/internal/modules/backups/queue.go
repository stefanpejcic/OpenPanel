package backups

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// restoreJob is one restore in an account's queue, finished ones stay as its restore log
type restoreJob struct {
	ID       string         `json:"id"`
	Request  restoreRequest `json:"request"`
	Status   string         `json:"status"` // queued, running, done, failed
	Message  string         `json:"message,omitempty"`
	Queued   time.Time      `json:"queued"`
	Started  *time.Time     `json:"started,omitempty"`
	Finished *time.Time     `json:"finished,omitempty"`
}

// how many finished restores the log keeps
const restoreLogSize = 100

// accountQueue runs one restore at a time for an account, the API's direct restores take the same lock
type accountQueue struct {
	mu      sync.Mutex
	restore sync.Mutex
	path    string
	jobs    []*restoreJob
	loaded  bool
	working bool
}

var (
	queuesMu sync.Mutex
	queues   = map[string]*accountQueue{}
	// swapped in tests
	runRestore = performRestore
)

func queueFor(userContext string) *accountQueue {
	queuesMu.Lock()
	defer queuesMu.Unlock()
	q := queues[userContext]
	if q == nil {
		q = &accountQueue{path: filepath.Join("/home", userContext, "backup_restores.json")}
		queues[userContext] = q
	}
	return q
}

// load reads the saved queue once, a restore that was running when the panel stopped can't be resumed so it's marked failed, call with mu held
func (q *accountQueue) load() {
	if q.loaded {
		return
	}
	q.loaded = true
	content, err := os.ReadFile(q.path)
	if err != nil || json.Unmarshal(content, &q.jobs) != nil {
		return
	}
	changed := false
	for _, j := range q.jobs {
		if j.Status == "running" {
			now := time.Now().UTC()
			j.Status, j.Message, j.Finished = "failed", "Interrupted, the panel restarted while this was running.", &now
			changed = true
		}
	}
	if changed {
		q.save()
	}
}

// save writes the queue and the newest finished jobs, call with mu held
func (q *accountQueue) save() {
	var active, finished []*restoreJob
	for _, j := range q.jobs {
		if j.Status == "queued" || j.Status == "running" {
			active = append(active, j)
		} else {
			finished = append(finished, j)
		}
	}
	if len(finished) > restoreLogSize {
		finished = finished[len(finished)-restoreLogSize:]
	}
	q.jobs = append(finished, active...)
	b, err := json.MarshalIndent(q.jobs, "", "  ")
	if err != nil {
		return
	}
	tmp := q.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, q.path)
	}
}

func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (q *accountQueue) enqueue(a *appctx.App, req restoreRequest) *restoreJob {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.load()
	job := &restoreJob{ID: newJobID(), Request: req, Status: "queued", Queued: time.Now().UTC()}
	q.jobs = append(q.jobs, job)
	q.save()
	q.startLocked(a)
	return job
}

// startLocked starts the worker when there's queued work and none is running, call with mu held
func (q *accountQueue) startLocked(a *appctx.App) {
	if q.working {
		return
	}
	for _, j := range q.jobs {
		if j.Status == "queued" {
			q.working = true
			go q.work(a)
			return
		}
	}
}

// work runs the queued restores one by one, oldest first
func (q *accountQueue) work(a *appctx.App) {
	for {
		q.mu.Lock()
		var job *restoreJob
		for _, j := range q.jobs {
			if j.Status == "queued" {
				job = j
				break
			}
		}
		if job == nil {
			q.working = false
			q.mu.Unlock()
			return
		}
		now := time.Now().UTC()
		job.Status, job.Started = "running", &now
		q.save()
		q.mu.Unlock()

		q.restore.Lock()
		status, body := runRestore(context.Background(), a, job.Request)
		q.restore.Unlock()

		q.mu.Lock()
		done := time.Now().UTC()
		job.Finished = &done
		job.Status = "done"
		if status != http.StatusOK {
			job.Status = "failed"
		}
		if msg, ok := body["message"].(string); ok {
			job.Message = msg
		}
		if msg, ok := body["error"].(string); ok {
			job.Message = msg
		}
		q.save()
		q.mu.Unlock()
	}
}

// jobView is what the restore page shows for a job, without the request internals
type jobView struct {
	ID       string     `json:"id"`
	Created  string     `json:"created,omitempty"`
	Target   string     `json:"target"`
	Database string     `json:"database,omitempty"`
	Items    []string   `json:"items,omitempty"`
	Status   string     `json:"status"`
	Message  string     `json:"message,omitempty"`
	Queued   time.Time  `json:"queued"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
}

func viewOf(j *restoreJob) jobView {
	target := j.Request.Form.Get("restore_target")
	if target == "" {
		target = "all"
	}
	return jobView{
		ID: j.ID, Created: archiveCreated(j.Request.Form.Get("backup_file")), Target: target,
		Database: j.Request.Form.Get("database"), Items: j.Request.Form["items"],
		Status: j.Status, Message: j.Message, Queued: j.Queued, Started: j.Started, Finished: j.Finished,
	}
}

// snapshot returns the queue oldest first and the log newest first
func (q *accountQueue) snapshot(a *appctx.App) (queue, logs []jobView) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.load()
	// queued jobs saved before a panel restart get picked up again here
	q.startLocked(a)
	queue, logs = []jobView{}, []jobView{}
	for _, j := range q.jobs {
		if j.Status == "queued" || j.Status == "running" {
			queue = append(queue, viewOf(j))
		} else {
			logs = append(logs, viewOf(j))
		}
	}
	sort.SliceStable(logs, func(i, k int) bool {
		return logs[i].Finished != nil && logs[k].Finished != nil && logs[i].Finished.After(*logs[k].Finished)
	})
	return queue, logs
}

// handleQueueRestore adds a restore to the account's queue, the restore page posts the same fields as /backups/restore
func handleQueueRestore(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseMultipartForm(1 << 20)
	backup := strings.TrimSpace(r.Form.Get("backup_file"))
	if backup == "" || filepath.Base(backup) != backup {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup filename."})
		return
	}
	form := r.Form
	form.Del("csrf_token")
	job := queueFor(userContext).enqueue(a, restoreRequest{UserID: userID, Username: username, UserContext: userContext, IP: reqip.ClientIP(r), Form: form})
	web.WriteJSON(w, http.StatusOK, map[string]any{"queued": true, "id": job.ID})
}

// handleRestoreJobs lists the account's restore queue and log for the restore page
func handleRestoreJobs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queue, logs := queueFor(userContext).snapshot(a)
	web.WriteJSON(w, http.StatusOK, map[string]any{"queue": queue, "logs": logs})
}
