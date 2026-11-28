// Package git is the Plesk-style Git module: per-account repos (cloned from a remote or pushed to over HTTPS) kept as bare repos in /home/<context>/git and checked out into a docroot on deploy.
package git

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
)

const (
	SourceRemote = "remote"
	SourceLocal  = "local"

	DeployAutomatic = "automatic"
	DeployManual    = "manual"
	DeployNone      = "none"

	htmlRoot = "/var/www/html"
)

// Repo is one entry in repos.json
type Repo struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	RemoteURL  string `json:"remote_url,omitempty"`
	RemoteUser string `json:"remote_user,omitempty"`
	RemotePass string `json:"remote_pass,omitempty"`
	// AccountID and RemoteRepo are set for repos picked from a connected account, its token then replaces RemoteUser/RemotePass
	AccountID  string `json:"account_id,omitempty"`
	RemoteRepo string `json:"remote_repo,omitempty"`
	HookID     int64  `json:"hook_id,omitempty"`
	Branch     string `json:"branch"`
	Domain     string `json:"domain"`
	TargetDir  string `json:"target_dir"`
	DeployMode string `json:"deploy_mode"`

	Actions          string `json:"actions,omitempty"`
	ActionsContainer string `json:"actions_container,omitempty"`

	WebhookToken string `json:"webhook_token"`
	PushPassHash string `json:"push_pass_hash,omitempty"`

	CreatedAt    string `json:"created_at"`
	LastPull     string `json:"last_pull,omitempty"`
	LastDeploy   string `json:"last_deploy,omitempty"`
	LastDeployed string `json:"last_deployed_commit,omitempty"`
	LastError    string `json:"last_error,omitempty"`
	LastOutput   string `json:"last_output,omitempty"`
}

// rootOverride lets tests keep repos in a temp dir
var rootOverride string

// accountRoot is the root-only dir holding an account's bare repos, repos.json, accounts.json and deploy key
func accountRoot(userContext string) string {
	if rootOverride != "" {
		return filepath.Join(rootOverride, userContext)
	}
	return filepath.Join("/home", userContext, "git")
}

func repoDir(userContext, name string) string {
	return filepath.Join(accountRoot(userContext), name+".git")
}

// stageDir is the root-only work tree a deploy checks out into before it's copied to the target
func stageDir(userContext, name string) string {
	return filepath.Join(accountRoot(userContext), name+".deploy")
}

func storePath(userContext string) string {
	return filepath.Join(accountRoot(userContext), "repos.json")
}

// errNoContext is for an account with no context yet, without it every such account would land in the same /home/git
var errNoContext = errors.New("Git is not available until this account's home directory exists")

func validContext(userContext string) bool {
	return userContext != "" && userContext != "." && userContext != ".." && !strings.ContainsAny(userContext, "/\\")
}

func ensureAccountRoot(userContext string) error {
	if !validContext(userContext) {
		return errNoContext
	}
	root := accountRoot(userContext)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	if err := checkAccountRoot(userContext); err != nil {
		return err
	}
	// rootless containers map to the account's uid, so keep this out of their reach
	return os.Chmod(root, 0o700)
}

// checkAccountRoot refuses a git dir the account could have planted, since the parent home is theirs and git runs as root in here
func checkAccountRoot(userContext string) error {
	if !validContext(userContext) {
		return errNoContext
	}
	root := accountRoot(userContext)
	fi, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("%s is not a plain directory", root)
	}
	if st, ok := fi.Sys().(*syscall.Stat_t); ok && int(st.Uid) != os.Geteuid() {
		return fmt.Errorf("%s is not owned by the panel", root)
	}
	return nil
}

// gitSafeEnv keeps root git from reading any global or XDG config, hooks path or attributes out of the account's home
var gitSafeEnv = []string{
	"GIT_CONFIG_NOSYSTEM=1",
	"GIT_CONFIG_GLOBAL=/dev/null",
	"HOME=/nonexistent",
	"XDG_CONFIG_HOME=/nonexistent",
}

var storeMu sync.Mutex

// loadRepos reads repos.json, a missing file is just an empty list
func loadRepos(userContext string) ([]Repo, error) {
	if err := checkAccountRoot(userContext); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Repo{}, nil
		}
		return nil, err
	}
	data, err := os.ReadFile(storePath(userContext))
	if errors.Is(err, os.ErrNotExist) {
		return []Repo{}, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func saveRepos(userContext string, repos []Repo) error {
	if err := ensureAccountRoot(userContext); err != nil {
		return err
	}
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	tmp := storePath(userContext) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, storePath(userContext))
}

// updateRepos runs fn on the account's repo list under a lock and saves whatever it leaves behind
func updateRepos(userContext string, fn func([]Repo) ([]Repo, error)) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	f, err := lockFile(userContext)
	if err != nil {
		return err
	}
	defer unlockFile(f)
	repos, err := loadRepos(userContext)
	if err != nil {
		return err
	}
	repos, err = fn(repos)
	if err != nil {
		return err
	}
	return saveRepos(userContext, repos)
}

// lockFile flocks repos.json's sibling so the panel and a webhook/push don't write over each other
func lockFile(userContext string) (*os.File, error) {
	if err := ensureAccountRoot(userContext); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(accountRoot(userContext), ".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func unlockFile(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

func findRepo(repos []Repo, name string) (int, bool) {
	for i, r := range repos {
		if r.Name == name {
			return i, true
		}
	}
	return -1, false
}

// getRepo is a read-only lookup by name
func getRepo(userContext, name string) (Repo, bool) {
	repos, err := loadRepos(userContext)
	if err != nil {
		return Repo{}, false
	}
	i, ok := findRepo(repos, name)
	if !ok {
		return Repo{}, false
	}
	return repos[i], true
}

// saveRepoFields reloads the repo under lock and lets fn change it, so concurrent status writes don't clobber settings edits
func saveRepoFields(userContext, name string, fn func(*Repo)) error {
	return updateRepos(userContext, func(repos []Repo) ([]Repo, error) {
		i, ok := findRepo(repos, name)
		if !ok {
			return nil, errors.New("repository not found")
		}
		fn(&repos[i])
		return repos, nil
	})
}

// repoLocks serializes fetch/push/deploy on the same repo
var repoLocks sync.Map

func lockRepo(userContext, name string) func() {
	v, _ := repoLocks.LoadOrStore(userContext+"/"+name, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

var (
	repoNameRE  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	branchRE    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,199}$`)
	httpsURLRE  = regexp.MustCompile(`^https?://[A-Za-z0-9.-]+(:[0-9]{1,5})?/[A-Za-z0-9._~/%+-]+$`)
	scpURLRE    = regexp.MustCompile(`^[A-Za-z0-9._-]+@[A-Za-z0-9.-]+:[A-Za-z0-9._~/+-]+$`)
	sshURLRE    = regexp.MustCompile(`^ssh://[A-Za-z0-9._-]+@[A-Za-z0-9.-]+(:[0-9]{1,5})?/[A-Za-z0-9._~/+-]+$`)
	containerRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)
)

func isValidRepoName(name string) bool {
	return repoNameRE.MatchString(name) && !strings.HasSuffix(name, ".git") && !strings.Contains(name, "..") && !strings.HasSuffix(name, ".lock")
}

func isValidBranch(branch string) bool {
	return branchRE.MatchString(branch) && !strings.Contains(branch, "..") && !strings.HasSuffix(branch, "/") &&
		!strings.HasSuffix(branch, ".lock") && !strings.Contains(branch, "//") && !strings.Contains(branch, "/.")
}

// isValidRemoteURL only lets through http(s) and ssh remotes, never file://, ext:: or local paths
func isValidRemoteURL(u string) bool {
	if u == "" || strings.HasPrefix(u, "-") || strings.Contains(u, "..") {
		return false
	}
	return httpsURLRE.MatchString(u) || scpURLRE.MatchString(u) || sshURLRE.MatchString(u)
}

func isSSHURL(u string) bool {
	return scpURLRE.MatchString(u) || sshURLRE.MatchString(u)
}

func isValidDeployMode(m string) bool {
	return m == DeployAutomatic || m == DeployManual || m == DeployNone
}

func isValidContainer(c string) bool {
	return containerRE.MatchString(c)
}

// cleanTargetDir normalizes a deploy target to /var/www/html/<rel>, rejecting anything outside it
func cleanTargetDir(dir string) (string, bool) {
	dir = strings.TrimSpace(dir)
	if dir == "" || strings.Contains(dir, "..") {
		return "", false
	}
	if !strings.HasPrefix(dir, "/") {
		dir = htmlRoot + "/" + dir
	}
	dir = filepath.Clean(dir)
	if dir != htmlRoot && !strings.HasPrefix(dir, htmlRoot+"/") {
		return "", false
	}
	return dir, true
}

// targetHostPath resolves a /var/www/html/... target to its host path, refusing symlink escapes
func targetHostPath(userContext, targetDir string) (string, error) {
	rel := strings.TrimPrefix(strings.TrimPrefix(targetDir, htmlRoot), "/")
	if rel == "" {
		return paths.SecureUserPath("HOME", userContext, "", false)
	}
	return paths.SecureUserPath("HOME", userContext, rel, false)
}

const tokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func randomToken(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			panic(err)
		}
		b[i] = tokenAlphabet[idx.Int64()]
	}
	return string(b)
}

func hashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

func checkPassword(hash, pw string) bool {
	return hash != "" && bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func tokensEqual(a, b string) bool {
	return a != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

const timeLayout = "2006-01-02 15:04:05 UTC"

func nowStr() string {
	return time.Now().UTC().Format(timeLayout)
}

// isoTime turns a stored timestamp into something the browser's relative-date helper can parse
func isoTime(stored string) string {
	t, err := time.Parse(timeLayout, stored)
	if err != nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// runGit runs git against a bare repo with prompts and odd transports disabled
func runGit(ctx context.Context, userContext string, env []string, args ...string) (string, error) {
	if err := checkAccountRoot(userContext); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ALLOW_PROTOCOL=http:https:ssh",
	)
	cmd.Env = append(cmd.Env, gitSafeEnv...)
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		return text, fmt.Errorf("%s", redact(text))
	}
	return text, nil
}

var userinfoRE = regexp.MustCompile(`(https?://)[^/@\s]+@`)

// redact strips user:pass@ from any URL git echoes back
func redact(s string) string {
	return userinfoRE.ReplaceAllString(s, "${1}***@")
}

// truncateOutput keeps stored output small
func truncateOutput(s string) string {
	const max = 8000
	if len(s) > max {
		return s[len(s)-max:]
	}
	return s
}
