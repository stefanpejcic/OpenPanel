package git

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

const (
	fetchTimeout   = 5 * time.Minute
	deployTimeout  = 5 * time.Minute
	actionsTimeout = 10 * time.Minute
)

func sshDir(userContext string) string {
	return filepath.Join(accountRoot(userContext), ".ssh")
}

func sshKeyPath(userContext string) string {
	return filepath.Join(sshDir(userContext), "id_ed25519")
}

// PublicKey returns the account's deploy key, generating it on first use
func PublicKey(ctx context.Context, userContext string) (string, error) {
	pub := sshKeyPath(userContext) + ".pub"
	if data, err := os.ReadFile(pub); err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if err := ensureAccountRoot(userContext); err != nil {
		return "", err
	}
	if err := os.MkdirAll(sshDir(userContext), 0o700); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-C", "openpanel-"+userContext, "-f", sshKeyPath(userContext))
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", errors.New(strings.TrimSpace(string(out)))
	}
	data, err := os.ReadFile(pub)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// remoteEnv carries the remote's credentials without putting them in argv or the repo config
func remoteEnv(userContext string, r Repo) []string {
	if isSSHURL(r.RemoteURL) {
		return []string{"GIT_SSH_COMMAND=ssh -i " + sshKeyPath(userContext) +
			" -o IdentitiesOnly=yes -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=" +
			filepath.Join(sshDir(userContext), "known_hosts")}
	}
	if r.RemoteUser != "" || r.RemotePass != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(r.RemoteUser + ":" + r.RemotePass))
		return []string{"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=http.extraHeader", "GIT_CONFIG_VALUE_0=Authorization: Basic " + auth}
	}
	return nil
}

// initRepo creates the bare repo on disk, cloning it for remotes, and fills in the branch when it was left empty
func initRepo(ctx context.Context, userContext string, r *Repo) error {
	if err := ensureAccountRoot(userContext); err != nil {
		return err
	}
	dir := repoDir(userContext, r.Name)
	if _, err := os.Stat(dir); err == nil {
		return errors.New("a repository with this name already exists")
	}

	if r.Source == SourceLocal {
		if _, err := runGit(ctx, userContext, nil, "init", "--bare", "-q", "-b", r.Branch, dir); err != nil {
			_ = os.RemoveAll(dir)
			return err
		}
		if _, err := runGit(ctx, userContext, nil, "-C", dir, "config", "http.receivepack", "true"); err != nil {
			_ = os.RemoveAll(dir)
			return err
		}
		return nil
	}

	if isSSHURL(r.RemoteURL) {
		if _, err := PublicKey(ctx, userContext); err != nil {
			return err
		}
	}
	cctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	if _, err := runGit(cctx, userContext, remoteEnv(userContext, *r), "clone", "--bare", "-q", "--", r.RemoteURL, dir); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	// bare clones have no fetch refspec, add one so pulls update every branch
	if _, err := runGit(ctx, userContext, nil, "-C", dir, "config", "remote.origin.fetch", "+refs/heads/*:refs/heads/*"); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	if r.Branch == "" {
		head, err := runGit(ctx, userContext, nil, "-C", dir, "symbolic-ref", "--short", "HEAD")
		if err != nil || head == "" {
			head = "main"
		}
		r.Branch = head
	}
	if _, ok := branchCommit(ctx, userContext, r.Name, r.Branch); !ok {
		_ = os.RemoveAll(dir)
		return errors.New("branch " + r.Branch + " was not found in the remote repository")
	}
	return nil
}

// fetchRepo pulls every branch from the remote
func fetchRepo(ctx context.Context, userContext string, r Repo) (string, error) {
	if r.Source != SourceRemote {
		return "", errors.New("only remote repositories can be pulled")
	}
	cctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	return runGit(cctx, userContext, remoteEnv(userContext, r), "-C", repoDir(userContext, r.Name), "fetch", "--prune", "origin")
}

// branchCommit returns the branch tip, false when the branch doesn't exist yet (e.g. nothing pushed)
func branchCommit(ctx context.Context, userContext, name, branch string) (string, bool) {
	out, err := runGit(ctx, userContext, nil, "-C", repoDir(userContext, name), "rev-parse", "--verify", "-q", "refs/heads/"+branch+"^{commit}")
	if err != nil || out == "" {
		return "", false
	}
	return out, true
}

// chownTree is swapped out in tests, deployed files must belong to the account's uid or its containers can't touch them
var chownTree = func(ctx context.Context, userContext, hostPath string) error {
	uid, err := podmanmanager.GetUID(userContext)
	if err != nil {
		return err
	}
	ids := strconv.Itoa(uid) + ":" + strconv.Itoa(uid)
	return exec.CommandContext(ctx, "chown", "-R", ids, hostPath).Run()
}

// runActions is swapped out in tests, it runs the post-deploy commands inside the chosen container
var runActions = func(ctx context.Context, userContext, container, workdir, script string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", `cd "$1" 2>/dev/null; `+script, "sh", workdir)
	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	return string(out), err
}

// deployRepo checks the branch out into the target dir, then runs the extra deploy actions
func deployRepo(ctx context.Context, userContext string, r Repo) (commit, output string, err error) {
	commit, ok := branchCommit(ctx, userContext, r.Name, r.Branch)
	if !ok {
		return "", "", errors.New("branch " + r.Branch + " has no commits yet, push to it first")
	}
	hostPath, err := targetHostPath(userContext, r.TargetDir)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(hostPath, 0o755); err != nil {
		return "", "", err
	}

	cctx, cancel := context.WithTimeout(ctx, deployTimeout)
	defer cancel()
	dir := repoDir(userContext, r.Name)
	out, err := runGit(cctx, userContext, nil, "--git-dir="+dir, "--work-tree="+hostPath, "checkout", "-f", "-q", r.Branch)
	if err != nil {
		return "", out, err
	}
	if err := chownTree(cctx, userContext, hostPath); err != nil {
		return commit, out, errors.New("files deployed but setting their owner failed: " + err.Error())
	}

	if strings.TrimSpace(r.Actions) != "" && r.ActionsContainer != "" {
		actx, acancel := context.WithTimeout(ctx, actionsTimeout)
		defer acancel()
		script := strings.ReplaceAll(strings.TrimSpace(r.Actions), "\r\n", "\n")
		actionsOut, actionsErr := runActions(actx, userContext, r.ActionsContainer, r.TargetDir, script)
		out = strings.TrimSpace(out + "\n" + actionsOut)
		if actionsErr != nil {
			return commit, out, errors.New("files deployed but deploy actions failed: " + actionsErr.Error())
		}
	}
	return commit, out, nil
}

// deployAndRecord deploys under the repo lock and stores the outcome on the repo
func deployAndRecord(ctx context.Context, userContext string, r Repo) error {
	unlock := lockRepo(userContext, r.Name)
	commit, out, err := deployRepo(ctx, userContext, r)
	unlock()
	_ = saveRepoFields(userContext, r.Name, func(x *Repo) {
		x.LastDeploy = nowStr()
		x.LastOutput = truncateOutput(redact(out))
		if commit != "" {
			x.LastDeployed = commit
		}
		x.LastError = ""
		if err != nil {
			x.LastError = err.Error()
		}
	})
	return err
}

// pullAndRecord fetches, then deploys too when the repo is on automatic
func pullAndRecord(ctx context.Context, userContext string, r Repo) error {
	unlock := lockRepo(userContext, r.Name)
	out, err := fetchRepo(ctx, userContext, r)
	unlock()
	_ = saveRepoFields(userContext, r.Name, func(x *Repo) {
		x.LastPull = nowStr()
		x.LastError = ""
		if err != nil {
			x.LastError = err.Error()
			x.LastOutput = truncateOutput(redact(out))
		}
	})
	if err != nil {
		return err
	}
	if r.DeployMode == DeployAutomatic {
		return deployAndRecord(ctx, userContext, r)
	}
	return nil
}

// Commit is one line of the commit log
type Commit struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	Date    string
	ISO     string
	Subject string
}

func commitLog(ctx context.Context, userContext string, r Repo, limit int) ([]Commit, error) {
	if _, ok := branchCommit(ctx, userContext, r.Name, r.Branch); !ok {
		return nil, nil
	}
	out, err := runGit(ctx, userContext, nil, "-C", repoDir(userContext, r.Name), "log", "-n", strconv.Itoa(limit),
		"--format=%H%x1f%h%x1f%an%x1f%ae%x1f%ad%x1f%aI%x1f%s", "--date=format:%Y-%m-%d %H:%M", "refs/heads/"+r.Branch, "--")
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\x1f")
		if len(f) != 7 {
			continue
		}
		commits = append(commits, Commit{Hash: f[0], Short: f[1], Author: f[2], Email: f[3], Date: f[4], ISO: f[5], Subject: f[6]})
	}
	return commits, nil
}

// newCommitCount is how many commits after has that before didn't, "?" if git can't tell
func newCommitCount(ctx context.Context, userContext, name, before, after string) string {
	rng := after
	if before != "" {
		rng = before + ".." + after
	}
	out, err := runGit(ctx, userContext, nil, "-C", repoDir(userContext, name), "rev-list", "--count", rng)
	if err != nil {
		return "?"
	}
	return out
}

func listBranches(ctx context.Context, userContext, name string) []string {
	out, err := runGit(ctx, userContext, nil, "-C", repoDir(userContext, name), "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// latestCommit is the branch tip for the list view
func latestCommit(ctx context.Context, userContext string, r Repo) (Commit, bool) {
	commits, err := commitLog(ctx, userContext, r, 1)
	if err != nil || len(commits) == 0 {
		return Commit{}, false
	}
	return commits[0], true
}

func removeRepo(userContext, name string) error {
	if !isValidRepoName(name) {
		return errors.New("invalid repository name")
	}
	return os.RemoveAll(repoDir(userContext, name))
}
