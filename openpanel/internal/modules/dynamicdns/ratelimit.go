package dynamicdns

import (
	"sync"
	"time"
)

// ipLimiter is a fixed-window counter per client IP for the public update URL, which needs no login
type ipLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*ipWindow
}

type ipWindow struct {
	start time.Time
	n     int
}

var updateLimiter = &ipLimiter{limit: 30, window: time.Minute, hits: map[string]*ipWindow{}}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if len(l.hits) > 10000 {
		for k, v := range l.hits {
			if now.Sub(v.start) > l.window {
				delete(l.hits, k)
			}
		}
	}
	w, ok := l.hits[ip]
	if !ok || now.Sub(w.start) > l.window {
		l.hits[ip] = &ipWindow{start: now, n: 1}
		return true
	}
	w.n++
	return w.n <= l.limit
}
