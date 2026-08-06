package account

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// loginRateLimiter enforces a fixed-window request limit ("N per minute")
// on POST /login, keyed per IP. State is kept in-process rather than in a
// shared store - correct for a single Go binary instance, but won't share
// state across multiple instances behind a load balancer if the deployment
// ever grows that way.
type loginRateLimiter struct {
	limit int

	mu      sync.Mutex
	windows map[string]*window
}

type window struct {
	start time.Time
	count int
}

func newLoginRateLimiter(a *appctx.App) *loginRateLimiter {
	limit := atoiDefault(a.Config.Get("login_ratelimit", ""), 5)
	return &loginRateLimiter{limit: limit, windows: map[string]*window{}}
}

func (l *loginRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, ok := l.windows[ip]
	if !ok || now.Sub(w.start) >= time.Minute {
		w = &window{start: now}
		l.windows[ip] = w
	}
	w.count++
	return w.count <= l.limit
}

const failedLoginLogPath = "/var/log/openpanel/user/failed_login.log"

var (
	failedAttemptsMu sync.Mutex
	failedAttempts   = map[string]int{}
)

func clearFailedAttempts(ip string) {
	failedAttemptsMu.Lock()
	delete(failedAttempts, ip)
	failedAttemptsMu.Unlock()
}

// handleRateLimitExceeded logs the throttled attempt, tracks a separate
// (unbounded, in-memory) failure counter per IP, and temporarily blocks
// the IP via CSF once that counter passes login_blocklimit.
func handleRateLimitExceeded(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ip := reqip.ClientIP(r)
	postLimit := atoiDefault(a.Config.Get("login_ratelimit", ""), 5)
	blockLimit := atoiDefault(a.Config.Get("login_blocklimit", ""), 20)

	appendLine(failedLoginLogPath, fmt.Sprintf("%s Rate limit for login: '%d per minute' exceeded from IP: %s",
		time.Now().Format("2006-01-02 15:04:05"), postLimit, ip))

	failedAttemptsMu.Lock()
	failedAttempts[ip]++
	count := failedAttempts[ip]
	failedAttemptsMu.Unlock()

	if count > blockLimit {
		blockIPTemporarily(ip, blockLimit)
	}

	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{"error":"rate_limit_exceeded","message":"Too many requests. Please try again later."}`)
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	locale := resolveLocale(a, r, sess)
	t := a.I18n.Translator(locale)
	flash.Add(sess, "danger", t.Get("Too many failed login attempts. Please try again later."))
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// blockIPTemporarily appends a CSF tempban entry if CSF is installed,
// otherwise just logs that it couldn't.
func blockIPTemporarily(ip string, blockLimit int) {
	const tempbanPath = "/var/lib/csf/csf.tempban"
	const errorLogPath = "/var/log/openpanel/user/error.log"

	if _, err := os.Stat(tempbanPath); err != nil {
		appendLine(errorLogPath, fmt.Sprintf("%s - Failed to block IP %s (after %d failed logins) on Firewall: CSF is not installed.",
			time.Now().Format(time.RFC3339), ip, blockLimit))
		return
	}

	entry := fmt.Sprintf("%d|%s||in|3600|Too many failed login attempts on OpenPanel",
		time.Now().Unix(), ip)
	if err := appendLine(tempbanPath, entry); err != nil {
		appendLine(errorLogPath, fmt.Sprintf("%s - An unexpected error occurred: %v", time.Now().Format(time.RFC3339), err))
		return
	}

	appendLine(failedLoginLogPath, fmt.Sprintf("%s IP: %s temporary blocked on CSF due to %d failed logins.",
		time.Now().Format("2006-01-02 15:04:05"), ip, blockLimit))
}

func appendLine(path, line string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}
