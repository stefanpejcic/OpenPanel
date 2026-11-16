package backups

import (
	"context"
	"io"
	"sync"
	"time"
)

// restoreProgress is what a running restore reports for the queue page, every method is a no-op on nil so restores outside the queue don't care
type restoreProgress struct {
	mu        sync.Mutex
	size      int64
	phase     int
	phases    int
	done      int64
	step      string
	detail    string
	items     int
	started   time.Time
	phaseFrom time.Time
	stepFrom  time.Time
	steps     []stepTiming
	now       func() time.Time
}

// stepTiming is how long one step took, kept in the log so it shows where the time went
type stepTiming struct {
	Step    string  `json:"step"`
	Detail  string  `json:"detail,omitempty"`
	Seconds float64 `json:"seconds"`
}

// progressView is the snapshot the page gets, Percent and ETA are -1 while they can't be known yet
type progressView struct {
	Step    string       `json:"step"`
	Detail  string       `json:"detail,omitempty"`
	Phase   int          `json:"phase"`
	Phases  int          `json:"phases"`
	Done    int64        `json:"done"`
	Size    int64        `json:"size"`
	Items   int          `json:"items"`
	Percent float64      `json:"percent"`
	Rate    float64      `json:"rate"`
	ETA     float64      `json:"eta"`
	Steps   []stepTiming `json:"steps,omitempty"`
}

func newRestoreProgress(size int64) *restoreProgress {
	p := &restoreProgress{size: size, phases: 1, now: time.Now}
	p.started = p.now()
	p.stepFrom = p.started
	return p
}

type progressKey struct{}

func withProgress(ctx context.Context, p *restoreProgress) context.Context {
	return context.WithValue(ctx, progressKey{}, p)
}

func progressFrom(ctx context.Context) *restoreProgress {
	p, _ := ctx.Value(progressKey{}).(*restoreProgress)
	return p
}

// setPhases is how many passes over the archive this restore makes, a full restore downloads, then reads it twice
func (p *restoreProgress) setPhases(n int) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phases = n
}

// startPhase begins the next pass over the archive with its own byte count
func (p *restoreProgress) startPhase(step string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase++
	p.done = 0
	p.phaseFrom = p.now()
	p.setStepLocked(step, "")
}

// setStep names what's happening now, the previous step's time goes to the log
func (p *restoreProgress) setStep(step, detail string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setStepLocked(step, detail)
}

func (p *restoreProgress) setStepLocked(step, detail string) {
	if step == p.step && detail == p.detail {
		return
	}
	p.closeStepLocked()
	p.step, p.detail, p.stepFrom = step, detail, p.now()
}

func (p *restoreProgress) closeStepLocked() {
	if p.step == "" {
		return
	}
	secs := p.now().Sub(p.stepFrom).Seconds()
	// the same step can come back, like the archive read again after files
	for i := range p.steps {
		if p.steps[i].Step == p.step && p.steps[i].Detail == p.detail {
			p.steps[i].Seconds += secs
			return
		}
	}
	p.steps = append(p.steps, stepTiming{Step: p.step, Detail: p.detail, Seconds: secs})
}

func (p *restoreProgress) item() {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.items++
	p.mu.Unlock()
}

func (p *restoreProgress) add(n int) {
	p.mu.Lock()
	p.done += int64(n)
	p.mu.Unlock()
}

// reader counts what's read from the archive toward the current phase
func (p *restoreProgress) reader(r io.Reader) io.Reader {
	if p == nil {
		return r
	}
	return &progressReader{r: r, p: p}
}

type progressReader struct {
	r io.Reader
	p *restoreProgress
}

func (pr *progressReader) Read(b []byte) (int, error) {
	n, err := pr.r.Read(b)
	if n > 0 {
		pr.p.add(n)
	}
	return n, err
}

// finish closes the last step, the log keeps the timings
func (p *restoreProgress) finish() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeStepLocked()
	p.step, p.detail = "", ""
}

func (p *restoreProgress) view() *progressView {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	v := &progressView{Step: p.step, Detail: p.detail, Phase: p.phase, Phases: p.phases, Done: p.done, Size: p.size, Items: p.items, Percent: -1, ETA: -1}
	v.Steps = append([]stepTiming(nil), p.steps...)
	now := p.now()
	if p.phase > 0 {
		if secs := now.Sub(p.phaseFrom).Seconds(); secs > 0 {
			v.Rate = float64(p.done) / secs
		}
	}
	if p.size > 0 && p.phases > 0 && p.phase > 0 {
		frac := float64(p.done) / float64(p.size)
		if frac > 1 {
			frac = 1
		}
		overall := (float64(p.phase-1) + frac) / float64(p.phases)
		v.Percent = overall * 100
		// too early and the guess jumps around
		if elapsed := now.Sub(p.started).Seconds(); overall >= 0.02 && elapsed >= 3 {
			v.ETA = elapsed * (1 - overall) / overall
		}
	}
	return v
}
