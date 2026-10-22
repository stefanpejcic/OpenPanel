package emails

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	migrateContainer     = "imapsync"
	migrateTestContainer = "imapsync_test"
	migrateJobLabel      = "openpanel.imapsync.job"
	migrateActiveKey     = "imapsync:active"
	migrateKeepJobs      = 10
	migrateLogTailBytes  = 64 * 1024
	migrateLogViewBytes  = 1024 * 1024
)

// migrateHomeRoot is where account homes live, tests point it at a temp dir
var migrateHomeRoot = "/home"

// runPodman runs podman against the account's rootless instance, swapped out in tests
var runPodman = func(ctx context.Context, userContext string, args ...string) ([]byte, error) {
	argv := podmanmanager.PodmanArgv(userContext, args...)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

// lookupIP is swapped out in tests so they don't hit DNS
var lookupIP = func(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

// one start at a time so the per-user and server-wide limits can't race
var migrateStartMu sync.Mutex

// MigrateRequest is one migration's input, shared by the web form and the API
type MigrateRequest struct {
	SourceHost     string `json:"source_host"`
	SourcePort     int    `json:"source_port"`
	SourceSecurity string `json:"source_security"`
	SourceUser     string `json:"source_user"`
	SourcePassword string `json:"source_password"`
	Destination    string `json:"destination"`
	DestPassword   string `json:"destination_password"`
	ExcludeJunk    bool   `json:"exclude_junk"`
}

// MigrateJob is the <id>.json kept next to each log, Status/ExitCode/Message are filled in when read
type MigrateJob struct {
	ID          string    `json:"id"`
	SourceHost  string    `json:"source_host"`
	SourceUser  string    `json:"source_user"`
	Destination string    `json:"destination"`
	Started     time.Time `json:"started"`
	Cancelled   bool      `json:"cancelled,omitempty"`
	StartError  string    `json:"start_error,omitempty"`
	Status      string    `json:"status"`
	ExitCode    int       `json:"exit_code"`
	Message     string    `json:"message,omitempty"`
}

// MigrateStatus is what the page polls while a job runs
type MigrateStatus struct {
	Running bool         `json:"running"`
	Jobs    []MigrateJob `json:"jobs"`
	Log     string       `json:"log"`
}

// migrateError is a message that's safe to show the user as is
type migrateError struct {
	status int
	msg    string
}

func (e *migrateError) Error() string { return e.msg }

func userErr(msg string) error { return &migrateError{http.StatusBadRequest, msg} }

// errStatus maps an error to its HTTP status and the text to show for it
func errStatus(err error) (int, string) {
	var me *migrateError
	if errors.As(err, &me) {
		return me.status, me.msg
	}
	return http.StatusInternalServerError, "Internal error, please contact support."
}

var (
	migrateHostRE  = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,62})(\.[A-Za-z0-9]([A-Za-z0-9-]{0,62}))*\.?$`)
	migrateJobIDRE = regexp.MustCompile(`^\d{8}-\d{6}(-\d{1,2})?$`)
	migrateExitRE  = regexp.MustCompile(`Exiting with return value (\d+)(?: \(([^)]*)\))?`)
	// folders nobody wants copied over, matched against the last part of the folder path
	migrateJunkFolders = `(?i)(^|[./])(spam|junk|junk e-?mail|bulk mail|trash|bin|deleted items|deleted messages)$`
)

// validateMigrateRequest trims req in place and returns the first problem as a user-facing error
func validateMigrateRequest(req *MigrateRequest) error {
	req.SourceHost = strings.TrimSpace(req.SourceHost)
	req.SourceUser = strings.TrimSpace(req.SourceUser)
	req.Destination = strings.ToLower(strings.TrimSpace(req.Destination))
	req.SourceSecurity = strings.ToLower(strings.TrimSpace(req.SourceSecurity))

	if req.SourceHost == "" || len(req.SourceHost) > 253 {
		return userErr("Enter the source IMAP server.")
	}
	if net.ParseIP(req.SourceHost) == nil && !migrateHostRE.MatchString(req.SourceHost) {
		return userErr("The source IMAP server is not a valid hostname or IP address.")
	}
	switch req.SourceSecurity {
	case "":
		req.SourceSecurity = "ssl"
	case "ssl", "starttls", "none":
	default:
		return userErr("Invalid security option, use ssl, starttls or none.")
	}
	if req.SourcePort == 0 {
		req.SourcePort = 993
		if req.SourceSecurity != "ssl" {
			req.SourcePort = 143
		}
	}
	if req.SourcePort < 1 || req.SourcePort > 65535 {
		return userErr("The source port must be between 1 and 65535.")
	}
	if req.SourceUser == "" || len(req.SourceUser) > 254 || strings.ContainsAny(req.SourceUser, "\r\n\x00") {
		return userErr("Enter a valid source username.")
	}
	// imapsync only reads the first line of a passfile
	for _, p := range []string{req.SourcePassword, req.DestPassword} {
		if p == "" || len(p) > 1024 || strings.ContainsAny(p, "\r\n\x00") {
			return userErr("Enter both passwords, without line breaks.")
		}
	}
	if !isValidEmail(req.Destination) {
		return userErr("Select a destination email account.")
	}
	return nil
}

// isPublicIP is false for anything a user shouldn't be able to point imapsync at from inside the server
func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() {
		return false
	}
	for _, cidr := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "198.18.0.0/15", "240.0.0.0/4", "64:ff9b::/96", "64:ff9b:1::/48", "2001:db8::/32"} {
		if _, n, _ := net.ParseCIDR(cidr); n.Contains(ip) {
			return false
		}
	}
	return true
}

// resolveSource returns the IP imapsync should connect to, resolved here so the hostname can't rebind to an internal address later
func resolveSource(ctx context.Context, host string, allowPrivate bool) (string, error) {
	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else {
		lctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		found, err := lookupIP(lctx, strings.TrimSuffix(host, "."))
		if err != nil || len(found) == 0 {
			return "", userErr("Could not resolve the source IMAP server " + host + ".")
		}
		ips = found
	}
	var picked net.IP
	for _, ip := range ips {
		if !allowPrivate && !isPublicIP(ip) {
			return "", userErr("The source IMAP server resolves to a private or reserved address, which is not allowed.")
		}
		// prefer IPv4, rootless container networking doesn't always have IPv6
		if picked == nil || (picked.To4() == nil && ip.To4() != nil) {
			picked = ip
		}
	}
	return picked.String(), nil
}

// dialIMAP is swapped out in tests
var dialIMAP = func(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// migrateDest returns where the destination mailbox is reached from inside the account's containers, the account's own IP since pasta doesn't route host.containers.internal to published ports
func migrateDest(ctx context.Context, a *appctx.App, username string) (host, port, security string) {
	host = a.Config.Get("imapsync_host", "")
	if host == "" {
		host = a.GetCachedIPForUserOrPublicIPv4(ctx, username)
	}
	port, security = a.Config.Get("imapsync_port", ""), a.Config.Get("imapsync_security", "")
	if port != "" && security != "" {
		return host, port, security
	}
	// mailservers set up with just an IP have no SSL and only listen on 143
	if dialIMAP(net.JoinHostPort(host, "993")) {
		return host, "993", "ssl"
	}
	return host, "143", "auto"
}

func migrateImage(a *appctx.App) string {
	return a.Config.Get("imapsync_image", "docker.io/gilleslamiral/imapsync:latest")
}

func migrateConfigInt(a *appctx.App, key string, def int) int {
	if n, err := strconv.Atoi(a.Config.Get(key, "")); err == nil && n > 0 {
		return n
	}
	return def
}

// imapsyncArgs builds imapsync's own arguments, always in --opt=value form so a value starting with - can't be read as a flag
func imapsyncArgs(req MigrateRequest, sourceIP, destHost, destPort, destSecurity, passPrefix string) []string {
	args := []string{
		"--host1=" + sourceIP,
		"--port1=" + strconv.Itoa(req.SourcePort),
	}
	switch req.SourceSecurity {
	case "ssl":
		args = append(args, "--ssl1")
	case "starttls":
		args = append(args, "--tls1")
	default:
		args = append(args, "--notls1")
	}
	args = append(args,
		"--user1="+req.SourceUser,
		"--passfile1=/data/"+passPrefix+".p1",
		"--host2="+destHost,
		"--port2="+destPort,
	)
	// auto leaves it to imapsync, which uses STARTTLS when the server offers it
	switch destSecurity {
	case "ssl":
		args = append(args, "--ssl2")
	case "starttls":
		args = append(args, "--tls2")
	case "none":
		args = append(args, "--notls2")
	}
	args = append(args,
		"--user2="+req.Destination,
		"--passfile2=/data/"+passPrefix+".p2",
		"--noreleasecheck",
	)
	return args
}

// accountUID returns the owner of /home/<ctx>, which is who the account's rootless containers run as
func accountUID(userContext string) (int, int, error) {
	info, err := os.Stat(filepath.Join(migrateHomeRoot, userContext))
	if err != nil {
		return 0, 0, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("no owner for %s", userContext)
	}
	return int(st.Uid), int(st.Gid), nil
}

// jobDir is the host dir mounted into the imapsync container as /data
func jobDir(userContext string) string {
	return filepath.Join(migrateHomeRoot, userContext, "imapsync")
}

// openJobDir creates /home/<ctx>/imapsync if needed and opens it as an os.Root, so nothing planted inside can send our root writes elsewhere
func openJobDir(userContext string) (*os.Root, int, int, error) {
	if userContext == "" || strings.ContainsAny(userContext, "/\x00") || userContext == "." || userContext == ".." {
		return nil, 0, 0, fmt.Errorf("invalid context %q", userContext)
	}
	uid, gid, err := accountUID(userContext)
	if err != nil {
		return nil, 0, 0, err
	}
	dir := jobDir(userContext)
	info, err := os.Lstat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.Mkdir(dir, 0o700); err != nil {
			return nil, 0, 0, err
		}
		if err := os.Lchown(dir, uid, gid); err != nil {
			return nil, 0, 0, err
		}
	case err != nil:
		return nil, 0, 0, err
	case !info.IsDir():
		return nil, 0, 0, fmt.Errorf("%s is not a directory", dir)
	}
	root, err := os.OpenRoot(dir)
	return root, uid, gid, err
}

// writeOwned writes a new file in root owned by the account, so the container (running as the account) can read it
func writeOwned(root *os.Root, name string, data []byte, uid, gid int) error {
	_ = root.Remove(name)
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return root.Lchown(name, uid, gid)
}

func removePassfiles(root *os.Root, prefix string) {
	_ = root.Remove(prefix + ".p1")
	_ = root.Remove(prefix + ".p2")
}

// readTail returns up to max bytes from the end of name, starting on a full line
func readTail(root *os.Root, name string, max int64) string {
	f, err := root.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	start := info.Size() - max
	if start < 0 {
		start = 0
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	data, _ := io.ReadAll(io.LimitReader(f, max))
	s := string(data)
	if start > 0 {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		}
	}
	return s
}

// containerState reports whether the account's container name is running and which job it was started for
func containerState(ctx context.Context, userContext, name string) (bool, string) {
	out, err := runPodman(ctx, userContext, "container", "inspect", "-f", `{{.State.Running}}|{{index .Config.Labels "`+migrateJobLabel+`"}}`, name)
	if err != nil {
		return false, ""
	}
	running, job, _ := strings.Cut(strings.TrimSpace(string(out)), "|")
	return running == "true", job
}

func readJob(root *os.Root, id string) (MigrateJob, error) {
	var job MigrateJob
	f, err := root.Open(id + ".json")
	if err != nil {
		return job, err
	}
	defer f.Close()
	err = json.NewDecoder(io.LimitReader(f, 64*1024)).Decode(&job)
	return job, err
}

func writeJob(root *os.Root, job MigrateJob, uid, gid int) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return writeOwned(root, job.ID+".json", data, uid, gid)
}

// jobIDs lists the jobs in root, newest first
func jobIDs(root *os.Root) []string {
	dir, err := root.Open(".")
	if err != nil {
		return nil
	}
	defer dir.Close()
	names, _ := dir.Readdirnames(-1)
	var ids []string
	for _, n := range names {
		if id, ok := strings.CutSuffix(n, ".json"); ok && migrateJobIDRE.MatchString(id) {
			ids = append(ids, id)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids
}

// fillJobStatus works out how a job ended from its log, since --rm takes the exit code with the container
func fillJobStatus(root *os.Root, job *MigrateJob, runningID string) {
	switch {
	case job.ID == runningID:
		job.Status = "running"
		return
	case job.StartError != "":
		job.Status, job.Message = "failed", job.StartError
		return
	}
	m := migrateExitRE.FindAllStringSubmatch(readTail(root, job.ID+".log", 8*1024), -1)
	switch {
	case job.Cancelled:
		job.Status, job.Message = "cancelled", "Cancelled by user."
	case m == nil:
		job.Status, job.Message = "failed", "Stopped before finishing."
	default:
		last := m[len(m)-1]
		job.ExitCode, _ = strconv.Atoi(last[1])
		job.Status = "completed"
		if job.ExitCode != 0 {
			// imapsync names the exit code, e.g. EXIT_AUTHENTICATION_FAILURE_USER1
			job.Status, job.Message = "failed", strings.TrimSpace(last[2])
		}
	}
}

// pruneJobs keeps the newest migrateKeepJobs jobs and drops the rest along with their logs
func pruneJobs(root *os.Root, keepID string) {
	ids := jobIDs(root)
	for i, id := range ids {
		if i < migrateKeepJobs || id == keepID {
			continue
		}
		_ = root.Remove(id + ".json")
		_ = root.Remove(id + ".log")
		removePassfiles(root, id)
	}
}

// migrationStatus lists the account's jobs and the log of the newest one, and cleans up passfiles left by jobs that are no longer running
func migrationStatus(ctx context.Context, userContext string) (MigrateStatus, error) {
	st := MigrateStatus{Jobs: []MigrateJob{}}
	root, _, _, err := openJobDir(userContext)
	if err != nil {
		return st, err
	}
	defer root.Close()

	running, runningID := containerState(ctx, userContext, migrateContainer)
	if !running {
		runningID = ""
	}
	st.Running = running
	for _, id := range jobIDs(root) {
		job, err := readJob(root, id)
		if err != nil {
			continue
		}
		job.ID = id
		if id != runningID {
			removePassfiles(root, id)
		}
		fillJobStatus(root, &job, runningID)
		st.Jobs = append(st.Jobs, job)
	}
	if len(st.Jobs) > 0 {
		st.Log = readTail(root, st.Jobs[0].ID+".log", migrateLogTailBytes)
	}
	return st, nil
}

// readJobLog returns the end of one job's log
func readJobLog(userContext, id string) (string, error) {
	if !migrateJobIDRE.MatchString(id) {
		return "", &migrateError{http.StatusNotFound, "Log not found."}
	}
	root, _, _, err := openJobDir(userContext)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if _, err := root.Stat(id + ".json"); err != nil {
		return "", &migrateError{http.StatusNotFound, "Log not found."}
	}
	return readTail(root, id+".log", migrateLogViewBytes), nil
}

// allowRate counts one hit against key and reports whether it's still under limit for this hour, failing open if redis is down
func allowRate(ctx context.Context, a *appctx.App, key string, limit int) bool {
	if a.Cache == nil {
		return true
	}
	n, err := a.Cache.Raw().Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if n == 1 {
		_ = a.Cache.Raw().Expire(ctx, key, time.Hour).Err()
	}
	return n <= int64(limit)
}

// activeMigrations counts accounts with a running import across the server, dropping ones that have finished
func activeMigrations(ctx context.Context, a *appctx.App) int {
	if a.Cache == nil {
		return 0
	}
	members, err := a.Cache.Raw().SMembers(ctx, migrateActiveKey).Result()
	if err != nil {
		return 0
	}
	count := 0
	for _, uc := range members {
		if running, _ := containerState(ctx, uc, migrateContainer); running {
			count++
		} else {
			_ = a.Cache.Raw().SRem(ctx, migrateActiveKey, uc).Err()
		}
	}
	return count
}

// ownsMailbox checks the destination is one of the account's own mailboxes
func ownsMailbox(ctx context.Context, a *appctx.App, userID int, username, email string) bool {
	domains, _ := a.AllDomainsForUser(ctx, userID)
	for _, addr := range addressesOf(GetEmailList(ctx, a, userID, username, domainSet(domains))) {
		if strings.EqualFold(addr, email) {
			return true
		}
	}
	return false
}

// newJobID is a sortable timestamp, with a suffix if two land in the same second
func newJobID(root *os.Root) string {
	base := time.Now().UTC().Format("20060102-150405")
	id := base
	for i := 1; i < 100; i++ {
		if _, err := root.Stat(id + ".json"); errors.Is(err, fs.ErrNotExist) {
			break
		}
		id = base + "-" + strconv.Itoa(i)
	}
	return id
}

// startMigration validates req and starts the account's imapsync container in the background
func startMigration(ctx context.Context, a *appctx.App, userID int, username, userContext string, req MigrateRequest) (MigrateJob, error) {
	var job MigrateJob
	if err := validateMigrateRequest(&req); err != nil {
		return job, err
	}
	if !ownsMailbox(ctx, a, userID, username, req.Destination) {
		return job, &migrateError{http.StatusForbidden, "The destination must be one of your email accounts."}
	}
	if !allowRate(ctx, a, fmt.Sprintf("imapsync:starts:%d", userID), migrateConfigInt(a, "imapsync_starts_per_hour", 10)) {
		return job, &migrateError{http.StatusTooManyRequests, "Too many migrations started in the last hour, please try again later."}
	}
	sourceIP, err := resolveSource(ctx, req.SourceHost, a.Config.Get("imapsync_allow_private", "no") == "yes")
	if err != nil {
		return job, err
	}

	migrateStartMu.Lock()
	defer migrateStartMu.Unlock()

	if running, _ := containerState(ctx, userContext, migrateContainer); running {
		return job, &migrateError{http.StatusConflict, "A migration is already running. Wait for it to finish or cancel it first."}
	}
	if activeMigrations(ctx, a) >= migrateConfigInt(a, "imapsync_max_jobs", 10) {
		return job, &migrateError{http.StatusServiceUnavailable, "The server is busy with other migrations, please try again in a few minutes."}
	}

	root, uid, gid, err := openJobDir(userContext)
	if err != nil {
		return job, err
	}
	defer root.Close()

	job = MigrateJob{
		ID: newJobID(root), SourceHost: req.SourceHost, SourceUser: req.SourceUser,
		Destination: req.Destination, Started: time.Now().UTC(),
	}
	if err := writeOwned(root, job.ID+".p1", []byte(req.SourcePassword+"\n"), uid, gid); err != nil {
		return job, err
	}
	if err := writeOwned(root, job.ID+".p2", []byte(req.DestPassword+"\n"), uid, gid); err != nil {
		removePassfiles(root, job.ID)
		return job, err
	}
	if err := writeJob(root, job, uid, gid); err != nil {
		removePassfiles(root, job.ID)
		return job, err
	}

	args := []string{
		"run", "-d", "--rm", "--name", migrateContainer,
		"--label", migrateJobLabel + "=" + job.ID,
		"--user", "0:0", "--cap-drop", "all", "--security-opt", "no-new-privileges",
		"--memory", a.Config.Get("imapsync_memory", "512m"), "--cpus", a.Config.Get("imapsync_cpus", "1"), "--pids-limit", "128",
		"--timeout", strconv.Itoa(migrateConfigInt(a, "imapsync_timeout_hours", 24) * 3600),
		"-v", jobDir(userContext) + ":/data",
		"--entrypoint", "imapsync",
		migrateImage(a),
	}
	destHost, destPort, destSecurity := migrateDest(ctx, a, username)
	args = append(args, imapsyncArgs(req, sourceIP, destHost, destPort, destSecurity, job.ID)...)
	// file logging is off by default inside docker
	args = append(args, "--automap", "--nofoldersizes", "--log", "--logdir=/data", "--logfile="+job.ID+".log")
	if req.ExcludeJunk {
		args = append(args, "--exclude="+migrateJunkFolders)
	}

	// not tied to the request, the first run pulls the image and a closed tab shouldn't abort that
	runCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if out, err := runPodman(runCtx, userContext, args...); err != nil {
		removePassfiles(root, job.ID)
		job.StartError = "Could not start the migration: " + lastLine(string(out))
		_ = writeJob(root, job, uid, gid)
		return job, &migrateError{http.StatusInternalServerError, job.StartError}
	}

	if a.Cache != nil {
		_ = a.Cache.Raw().SAdd(ctx, migrateActiveKey, userContext).Err()
	}
	pruneJobs(root, job.ID)
	go dropPassfilesAfterLogin(userContext, job.ID)
	job.Status = "running"
	return job, nil
}

// dropPassfilesAfterLogin removes a job's passfiles once imapsync has read them, it reads both before logging in to host2
func dropPassfilesAfterLogin(userContext, id string) {
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)
		root, _, _, err := openJobDir(userContext)
		if err != nil {
			return
		}
		log := readTail(root, id+".log", migrateLogTailBytes)
		running, _ := containerState(context.Background(), userContext, migrateContainer)
		done := !running || strings.Contains(log, "Host2: success login") || strings.Contains(log, "Exiting with return value")
		if done {
			removePassfiles(root, id)
		}
		root.Close()
		if done {
			return
		}
	}
	if root, _, _, err := openJobDir(userContext); err == nil {
		removePassfiles(root, id)
		root.Close()
	}
}

// cancelMigration stops the running job and marks it cancelled
func cancelMigration(ctx context.Context, userContext string) error {
	running, id := containerState(ctx, userContext, migrateContainer)
	if !running {
		return &migrateError{http.StatusConflict, "No migration is running."}
	}
	root, uid, gid, err := openJobDir(userContext)
	if err != nil {
		return err
	}
	defer root.Close()
	if migrateJobIDRE.MatchString(id) {
		if job, err := readJob(root, id); err == nil {
			job.Cancelled = true
			_ = writeJob(root, job, uid, gid)
		}
	}
	if out, err := runPodman(ctx, userContext, "stop", "-t", "10", migrateContainer); err != nil {
		return &migrateError{http.StatusInternalServerError, "Could not stop the migration: " + lastLine(string(out))}
	}
	if migrateJobIDRE.MatchString(id) {
		removePassfiles(root, id)
	}
	return nil
}

// MigrateTestResult is what a connection test reports for each side
type MigrateTestResult struct {
	Source      bool   `json:"source"`
	Destination bool   `json:"destination"`
	Output      string `json:"output"`
}

// testMigration runs imapsync --justlogin against both servers and waits for it
func testMigration(ctx context.Context, a *appctx.App, userID int, username, userContext string, req MigrateRequest) (MigrateTestResult, error) {
	var res MigrateTestResult
	if err := validateMigrateRequest(&req); err != nil {
		return res, err
	}
	if !ownsMailbox(ctx, a, userID, username, req.Destination) {
		return res, &migrateError{http.StatusForbidden, "The destination must be one of your email accounts."}
	}
	if !allowRate(ctx, a, fmt.Sprintf("imapsync:tests:%d", userID), migrateConfigInt(a, "imapsync_tests_per_hour", 30)) {
		return res, &migrateError{http.StatusTooManyRequests, "Too many connection tests in the last hour, please try again later."}
	}
	sourceIP, err := resolveSource(ctx, req.SourceHost, a.Config.Get("imapsync_allow_private", "no") == "yes")
	if err != nil {
		return res, err
	}

	root, uid, gid, err := openJobDir(userContext)
	if err != nil {
		return res, err
	}
	defer root.Close()

	prefix := "test-" + randomURLToken(6)
	defer removePassfiles(root, prefix)
	if err := writeOwned(root, prefix+".p1", []byte(req.SourcePassword+"\n"), uid, gid); err != nil {
		return res, err
	}
	if err := writeOwned(root, prefix+".p2", []byte(req.DestPassword+"\n"), uid, gid); err != nil {
		return res, err
	}

	args := []string{
		"run", "--rm", "--name", migrateTestContainer,
		"--user", "0:0", "--cap-drop", "all", "--security-opt", "no-new-privileges",
		"--memory", "256m", "--pids-limit", "64", "--timeout", "120",
		"-v", jobDir(userContext) + ":/data",
		"--entrypoint", "imapsync",
		migrateImage(a),
	}
	destHost, destPort, destSecurity := migrateDest(ctx, a, username)
	args = append(args, imapsyncArgs(req, sourceIP, destHost, destPort, destSecurity, prefix)...)
	args = append(args, "--justlogin", "--nolog")

	runCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	out, runErr := runPodman(runCtx, userContext, args...)
	text := string(out)
	if strings.Contains(text, "the container name \""+migrateTestContainer+"\" is already in use") {
		return res, &migrateError{http.StatusConflict, "A connection test is already running."}
	}
	res.Source = strings.Contains(text, "Host1: success login")
	res.Destination = strings.Contains(text, "Host2: success login")
	res.Output = loginLines(text)
	if runErr != nil && res.Output == "" {
		res.Output = lastLine(text)
	}
	return res, nil
}

// loginLines keeps just the lines that say how each login went
func loginLines(out string) string {
	var keep []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "success login") || strings.Contains(line, "failure:") && !strings.HasPrefix(line, "Err ") {
			keep = append(keep, line)
		}
	}
	return strings.Join(keep, "\n")
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// web UI

// handleMigratePage renders the form, the running job and past jobs
func handleMigratePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	accounts := addressesOf(GetEmailList(ctx, a, userID, username, domainSet(domains)))
	st, err := migrationStatus(ctx, userContext)
	if err != nil {
		web.FlashRedirect(a, w, r, "error", "Error: Could not open the migrations folder.", "/emails")
		return
	}
	renderMigratePage(a, w, r, accounts, st)
}

// migrateFormRequest reads the web form into a MigrateRequest
func migrateFormRequest(r *http.Request) MigrateRequest {
	port, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("source_port")))
	return MigrateRequest{
		SourceHost:     r.FormValue("source_host"),
		SourcePort:     port,
		SourceSecurity: r.FormValue("source_security"),
		SourceUser:     r.FormValue("source_user"),
		SourcePassword: r.FormValue("source_password"),
		Destination:    r.FormValue("destination"),
		DestPassword:   r.FormValue("destination_password"),
		ExcludeJunk:    r.FormValue("exclude_junk") != "",
	}
}

func writeMigrateErr(w http.ResponseWriter, err error) {
	status, msg := errStatus(err)
	web.WriteJSON(w, status, map[string]string{"error": msg})
}

func handleMigrateStart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	job, err := startMigration(r.Context(), a, userID, username, userContext, migrateFormRequest(r))
	if err != nil {
		writeMigrateErr(w, err)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "started email migration from "+job.SourceUser+"@"+job.SourceHost+" to "+job.Destination, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "job": job})
}

func handleMigrateTest(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	res, err := testMigration(r.Context(), a, userID, username, userContext, migrateFormRequest(r))
	if err != nil {
		writeMigrateErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

func handleMigrateCancel(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := cancelMigration(r.Context(), userContext); err != nil {
		writeMigrateErr(w, err)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "cancelled email migration", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

func handleMigrateStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	st, err := migrationStatus(r.Context(), userContext)
	if err != nil {
		writeMigrateErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, st)
}

func handleMigrateLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	text, err := readJobLog(userContext, r.PathValue("id"))
	if err != nil {
		status, msg := errStatus(err)
		http.Error(w, msg, status)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = io.WriteString(w, text)
}
