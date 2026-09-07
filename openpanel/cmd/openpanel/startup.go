package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// startup side effects live here: log dirs, dev_mode logging, restart flag, custom startup script, GOMAXPROCS sizing, Redis flush, and direct TLS termination when Caddy isn't already on it

const (
	restartFlagPath     = "/root/openpanel_restart_needed"
	customStartupScript = "/root/openpanel_run_on_startup"

	errorLogPath   = "/var/log/openpanel/user/error.log"
	accessLogPath  = "/var/log/openpanel/user/access.log"
	opencliLogPath = "/var/log/openpanel/admin/opencli.log"
)

// fatalLogger always writes to stderr, even when configureLogging silenced the default logger, so a real boot failure never gets swallowed
var fatalLogger = log.New(os.Stderr, "", log.LstdFlags)

// configureLogging silences log.Printf diagnostics unless dev_mode=on; fatalLogger stays untouched so boot failures still surface
func configureLogging(a *appctx.App) {
	devMode := strings.EqualFold(a.Config.Get("dev_mode", ""), "on")
	if !devMode {
		log.SetOutput(io.Discard)
	}
}

// runStartupTasks runs one-time setup before the HTTP listener starts accepting connections
func runStartupTasks() {
	ensureLogDirectories()
	emptyRestartFlag()
	runCustomStartupScript()
}

// ensureLogDirectories creates the log dirs upfront so handlers that append directly (like the login rate limiter) don't fail silently
func ensureLogDirectories() {
	for _, p := range []string{errorLogPath, accessLogPath, opencliLogPath} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			log.Printf("BOOTSTRAP - failed to create log directory for %s: %v", p, err)
		}
	}
}

// emptyRestartFlag truncates (not deletes) the "restart needed" marker file the GUI polls for
func emptyRestartFlag() {
	if _, err := os.Stat(restartFlagPath); err != nil {
		return
	}
	if err := os.Truncate(restartFlagPath, 0); err != nil {
		log.Printf("BOOTSTRAP - failed to clear %s: %v", restartFlagPath, err)
	}
}

// runCustomStartupScript runs /root/openpanel_run_on_startup with bash if present, capped at 60s
func runCustomStartupScript() {
	info, err := os.Stat(customStartupScript)
	if err != nil || info.IsDir() {
		return
	}
	log.Printf("BOOTSTRAP - executing custom startup script: %s", customStartupScript)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if out, runErr := exec.CommandContext(ctx, "bash", customStartupScript).CombinedOutput(); runErr != nil {
		log.Printf("BOOTSTRAP - error executing %s: %v\n%s", customStartupScript, runErr, string(out))
	} else {
		log.Printf("BOOTSTRAP - executed custom startup script successfully")
	}
}

// setGOMAXPROCSFromCgroup caps GOMAXPROCS to the container's real CPU quota - runtime.NumCPU() reports the host's core count even inside a quota-limited container
func setGOMAXPROCSFromCgroup() {
	quota, ok := cgroupCPUQuota()
	if !ok || quota <= 0 || quota >= runtime.NumCPU() {
		return
	}
	runtime.GOMAXPROCS(quota)
	log.Printf("BOOTSTRAP - GOMAXPROCS set to %d (container CPU quota), host reports %d cores", quota, runtime.NumCPU())
}

// cgroupCPUQuota reads the container's CPU allotment from cgroup v2 (cpu.max), falling back to cgroup v1 (cfs_quota_us/cfs_period_us).
// fractional quotas round up so a small allotment never collapses to 0.
func cgroupCPUQuota() (int, bool) {
	if data, err := os.ReadFile("/sys/fs/cgroup/cpu.max"); err == nil {
		fields := strings.Fields(strings.TrimSpace(string(data)))
		if len(fields) == 2 && fields[0] != "max" {
			return quotaOverPeriod(fields[0], fields[1])
		}
		return 0, false
	}

	quotaBytes, err1 := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	periodBytes, err2 := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err1 == nil && err2 == nil {
		return quotaOverPeriod(strings.TrimSpace(string(quotaBytes)), strings.TrimSpace(string(periodBytes)))
	}
	return 0, false
}

func quotaOverPeriod(quotaStr, periodStr string) (int, bool) {
	quota, qErr := strconv.ParseFloat(quotaStr, 64)
	period, pErr := strconv.ParseFloat(periodStr, 64)
	if qErr != nil || pErr != nil || quota <= 0 || period <= 0 {
		return 0, false
	}
	cpus := quota / period
	n := int(cpus)
	if cpus > float64(n) {
		n++
	}
	if n < 1 {
		n = 1
	}
	return n, true
}

// flushRedisCache clears every Redis key except active sessions once the listener is up, so a restart doesn't log everyone out
func flushRedisCache(c *cache.Cache) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rdb := c.Raw()
	var keys []string
	iter := rdb.Scan(ctx, 0, "*", 1000).Iterator()
	for iter.Next(ctx) {
		if key := iter.Val(); !strings.HasPrefix(key, "session:") {
			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		log.Printf("BOOTSTRAP - failed to scan Redis keys: %v", err)
		return
	}
	if len(keys) == 0 {
		log.Printf("BOOTSTRAP - Redis cache already empty (sessions preserved)")
		return
	}

	deleted, err := rdb.Del(ctx, keys...).Result()
	if err != nil {
		log.Printf("BOOTSTRAP - failed to clear Redis cache: %v", err)
		return
	}
	log.Printf("BOOTSTRAP - Redis cache cleared: %d key(s) deleted, sessions preserved", deleted)
}

// caddyCertDirs is where Caddy stores certs, checked in this order
var caddyCertDirs = []string{
	"/etc/openpanel/caddy/ssl/custom/",
	"/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/",
}

// opencli runs `opencli <args>` with a 5s timeout; errors are swallowed since callers treat a failed lookup as "not configured", not fatal
func opencli(args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "opencli", args...).Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// checkSSLExists returns the first non-empty <base>/<domain>/ dir across caddyCertDirs
func checkSSLExists(domain string) (dir string, ok bool) {
	for _, base := range caddyCertDirs {
		certDir := filepath.Join(base, domain)
		entries, err := os.ReadDir(certDir)
		if err == nil && len(entries) > 0 {
			return certDir, true
		}
	}
	return "", false
}

// tlsCertPaths decides if this server should terminate TLS itself: only when the panel's domain already has a cert and the admin port isn't 443 (so Caddy isn't already handling it).
// Falls back to plain HTTP otherwise - no domain, no cert, Caddy on 443, or listenAddr not matching the configured admin port (LISTEN_ADDR can be overridden for testing on another port).
func tlsCertPaths(listenAddr string) (certFile, keyFile string, ok bool) {
	domain, domainOK := opencli("domain")
	if !domainOK || domain == "" {
		log.Printf("BOOTSTRAP - could not determine panel domain (opencli domain failed or empty), running plain HTTP.")
		return "", "", false
	}
	port, portOK := opencli("port")
	if port == "" {
		port = defaultAdminPort
		log.Printf("BOOTSTRAP - could not determine admin port (opencli port failed or empty, ok=%v), falling back to default port %s.", portOK, port)
	}
	if !strings.HasSuffix(listenAddr, ":"+port) {
		log.Printf("BOOTSTRAP - listen address %q doesn't match configured admin port %q, running plain HTTP.", listenAddr, port)
		return "", "", false
	}

	certDir, found := checkSSLExists(domain)
	if !found {
		log.Printf("BOOTSTRAP - no SSL certificate found for %q, running plain HTTP.", domain)
		return "", "", false
	}
	if port == "443" {
		log.Printf("BOOTSTRAP - domain %q has SSL but port is 443: Caddy handles TLS, server runs plain HTTP.", domain)
		return "", "", false
	}

	certFile = filepath.Join(certDir, domain+".crt")
	keyFile = filepath.Join(certDir, domain+".key")
	if _, err := os.Stat(certFile); err != nil {
		log.Printf("BOOTSTRAP - SSL cert file missing for %q (%s): %v, running plain HTTP.", domain, certFile, err)
		return "", "", false
	}
	if _, err := os.Stat(keyFile); err != nil {
		log.Printf("BOOTSTRAP - SSL key file missing for %q (%s): %v, running plain HTTP.", domain, keyFile, err)
		return "", "", false
	}
	log.Printf("BOOTSTRAP - domain %q has SSL, server will terminate TLS directly (cert=%s)", domain, certFile)
	return certFile, keyFile, true
}
