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
)

// This file implements the server's startup side effects: log directory
// creation, dev_mode-gated logging, the restart-needed flag, the
// admin-provided custom startup script, container-aware worker/GOMAXPROCS
// sizing, a Redis cache flush once the server is ready, and direct TLS
// termination when the panel's own domain already has a cert Caddy isn't
// already serving on 443.

const (
	restartFlagPath     = "/root/openpanel_restart_needed"
	customStartupScript = "/root/openpanel_run_on_startup"

	errorLogPath   = "/var/log/openpanel/user/error.log"
	accessLogPath  = "/var/log/openpanel/user/access.log"
	opencliLogPath = "/var/log/openpanel/admin/opencli.log"
)

// fatalLogger always writes to stderr, even when configureLogging has
// silenced the default logger for non-dev-mode, so a genuine boot
// failure is never swallowed.
var fatalLogger = log.New(os.Stderr, "", log.LstdFlags)

// configureLogging silences all log.Printf diagnostic output unless
// dev_mode=on (unrecoverable boot failures still surface via untouched
// stderr, see fatalLogger below). log.Printf across this codebase is
// used for deliberate diagnostic output, so it's what gets silenced
// here; fatalLogger is a separate os.Stderr writer, unaffected.
func configureLogging(a *appctx.App) {
	devMode := strings.EqualFold(a.Config.Get("dev_mode", ""), "on")
	if !devMode {
		log.SetOutput(io.Discard)
	}
}

// runStartupTasks performs one-time startup side effects (directory
// creation, the restart flag, the custom startup script), run once
// before the HTTP listener starts accepting connections.
func runStartupTasks() {
	ensureLogDirectories()
	emptyRestartFlag()
	runCustomStartupScript()
}

// ensureLogDirectories creates the parent directories for
// error.log/access.log/opencli.log: several handlers in this codebase
// already append to error.log directly (e.g. internal/modules/account's
// login rate limiter) and would otherwise fail silently if the directory
// never got created.
func ensureLogDirectories() {
	for _, p := range []string{errorLogPath, accessLogPath, opencliLogPath} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			log.Printf("BOOTSTRAP - failed to create log directory for %s: %v", p, err)
		}
	}
}

// emptyRestartFlag truncates (not deletes) the "a restart is needed"
// marker file the GUI polls for.
func emptyRestartFlag() {
	if _, err := os.Stat(restartFlagPath); err != nil {
		return
	}
	if err := os.Truncate(restartFlagPath, 0); err != nil {
		log.Printf("BOOTSTRAP - failed to clear %s: %v", restartFlagPath, err)
	}
}

// runCustomStartupScript runs /root/openpanel_run_on_startup with bash
// if present, capped at 60s.
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

// setGOMAXPROCSFromCgroup caps GOMAXPROCS to the container's actual CPU
// allotment: Go's runtime.NumCPU()-based GOMAXPROCS default reports the
// HOST's core count even inside a CPU-quota-limited container, not the
// container's real allotment, so this reads the cgroup v2/v1 CPU quota
// when present and caps GOMAXPROCS to it instead.
func setGOMAXPROCSFromCgroup() {
	quota, ok := cgroupCPUQuota()
	if !ok || quota <= 0 || quota >= runtime.NumCPU() {
		return
	}
	runtime.GOMAXPROCS(quota)
	log.Printf("BOOTSTRAP - GOMAXPROCS set to %d (container CPU quota), host reports %d cores", quota, runtime.NumCPU())
}

// cgroupCPUQuota reads the number of CPUs allotted to this container from
// cgroup v2 (/sys/fs/cgroup/cpu.max: "<quota> <period>", or "max" for
// unlimited) or, failing that, cgroup v1
// (cpu.cfs_quota_us/cpu.cfs_period_us, -1 quota meaning unlimited).
// Fractional quotas round up (a container with 1.2 CPUs still gets
// GOMAXPROCS=2 minimum) so a small allotment never collapses to 0.
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

// flushRedisCache clears the Redis cache once the HTTP listener has
// bound its socket (see main()'s call site).
func flushRedisCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "podman", "exec", "openpanel_redis", "redis-cli", "FLUSHDB").CombinedOutput()
	if err != nil {
		log.Printf("BOOTSTRAP - failed to clear Redis cache: %v: %s", err, strings.TrimSpace(string(out)))
		return
	}
	log.Printf("BOOTSTRAP - Redis cache cleared: %s", strings.TrimSpace(string(out)))
}

// caddyCertDirs lists the directories Caddy stores certificates under,
// searched in order when looking up an existing cert for a domain.
var caddyCertDirs = []string{
	"/etc/openpanel/caddy/ssl/custom/",
	"/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/",
}

// opencli runs `opencli <args>` with a 5s timeout, returning ok=false on
// any failure (missing binary, non-zero exit, timeout). Errors are
// swallowed rather than propagated because callers treat a failed
// lookup as "not configured" rather than a fatal condition.
func opencli(args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "opencli", args...).Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// checkSSLExists returns the first non-empty <base>/<domain>/ directory
// across caddyCertDirs, if any.
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

// tlsCertPaths decides whether this server should terminate TLS itself:
// if the panel's own domain (`opencli domain`) already has a cert and
// the admin port (`opencli port`) isn't 443 - meaning Caddy isn't the
// one terminating TLS for it - the server terminates TLS itself using
// that same cert. Returns ok=false whenever plain HTTP is correct (no
// domain configured, no cert found, or Caddy already handles port 443) -
// including whenever listenAddr doesn't actually match `opencli port`'s
// admin port: LISTEN_ADDR can be overridden (e.g. for running this build
// alongside an existing install on a different port during migration
// testing), and terminating TLS for the *admin* port's cert on some other,
// plain-HTTP-only port would just break that port instead of securing it.
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
