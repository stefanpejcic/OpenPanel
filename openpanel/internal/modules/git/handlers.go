package git

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// containersFor is swapped out in tests
var containersFor = crons.Containers

func validContainer(ctx context.Context, userContext, c string) bool {
	if !isValidContainer(c) {
		return false
	}
	for _, x := range containersFor(ctx, userContext) {
		if x == c {
			return true
		}
	}
	return false
}

// repoNameFromURL guesses a name from the last URL segment, like Plesk does
func repoNameFromURL(u string) string {
	if i := strings.LastIndex(u, ":"); i != -1 && !strings.Contains(u, "://") {
		u = u[i+1:]
	}
	return strings.TrimSuffix(path.Base(strings.TrimSuffix(u, "/")), ".git")
}

// applyActions reads the deploy actions fields shared by create and settings
func applyActions(ctx context.Context, userContext string, v url.Values, r *Repo) error {
	if v.Get("actions_enabled") == "" {
		r.Actions, r.ActionsContainer = "", ""
		return nil
	}
	actions := strings.TrimSpace(v.Get("actions"))
	container := strings.TrimSpace(v.Get("actions_container"))
	if actions == "" {
		r.Actions, r.ActionsContainer = "", ""
		return nil
	}
	if len(actions) > 4000 {
		return errors.New("Deploy actions are too long.")
	}
	if !validContainer(ctx, userContext, container) {
		return errors.New("Select a valid container for the deploy actions.")
	}
	r.Actions, r.ActionsContainer = actions, container
	return nil
}

// createRepo validates the add form, sets up the bare repo and saves it, returning the push password for local repos
func createRepo(ctx context.Context, a *appctx.App, userID int, userContext string, v url.Values) (Repo, string, error) {
	r := Repo{
		Name:       strings.TrimSpace(v.Get("name")),
		Source:     strings.TrimSpace(v.Get("source")),
		RemoteURL:  strings.TrimSpace(v.Get("remote_url")),
		RemoteUser: strings.TrimSpace(v.Get("remote_user")),
		RemotePass: v.Get("remote_pass"),
		Branch:     strings.TrimSpace(v.Get("branch")),
		Domain:     strings.TrimSpace(v.Get("domain")),
		DeployMode: strings.TrimSpace(v.Get("deploy_mode")),
	}
	if r.DeployMode == "" {
		r.DeployMode = DeployAutomatic
	}

	switch r.Source {
	case SourceRemote:
		if !isValidRemoteURL(r.RemoteURL) {
			return r, "", errors.New("Invalid repository URL. Use an https:// or SSH (git@host:user/repo.git) URL.")
		}
		if isSSHURL(r.RemoteURL) {
			r.RemoteUser, r.RemotePass = "", ""
		}
		if r.Name == "" {
			r.Name = repoNameFromURL(r.RemoteURL)
		}
	case SourceLocal:
		r.RemoteURL, r.RemoteUser, r.RemotePass = "", "", ""
		if r.Name == "" {
			r.Name = r.Domain
		}
		if r.Branch == "" {
			r.Branch = "main"
		}
	default:
		return r, "", errors.New("Select where the repository comes from.")
	}

	if !isValidRepoName(r.Name) {
		return r, "", errors.New("Invalid repository name. Use letters, numbers, dots, dashes and underscores.")
	}
	if r.Branch != "" && !isValidBranch(r.Branch) {
		return r, "", errors.New("Invalid branch name.")
	}
	if !isValidDeployMode(r.DeployMode) {
		return r, "", errors.New("Invalid deployment mode.")
	}

	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		return r, "", err
	}
	docroot, found := "", false
	for _, d := range domains {
		if d.DomainURL == r.Domain {
			docroot, found = d.Docroot, true
			break
		}
	}
	if !found {
		return r, "", errors.New("You do not own this domain.")
	}
	target := strings.TrimSpace(v.Get("target_dir"))
	if target == "" {
		target = docroot
	}
	cleaned, ok := cleanTargetDir(target)
	if !ok {
		return r, "", errors.New("Target directory must be inside /var/www/html/.")
	}
	r.TargetDir = cleaned
	if _, err := targetHostPath(userContext, r.TargetDir); err != nil {
		return r, "", errors.New("Invalid target directory.")
	}
	if err := applyActions(ctx, userContext, v, &r); err != nil {
		return r, "", err
	}

	existing, err := loadRepos(userContext)
	if err != nil {
		return r, "", err
	}
	if _, dup := findRepo(existing, r.Name); dup {
		return r, "", errors.New("A repository with this name already exists.")
	}

	r.WebhookToken = randomToken(32)
	r.CreatedAt = nowStr()
	pushPassword := ""
	if r.Source == SourceLocal {
		pushPassword = randomToken(24)
		if r.PushPassHash, err = hashPassword(pushPassword); err != nil {
			return r, "", err
		}
	}

	unlock := lockRepo(userContext, r.Name)
	err = initRepo(ctx, userContext, &r)
	unlock()
	if err != nil {
		return r, "", errors.New("Could not set up the repository: " + err.Error())
	}

	err = updateRepos(userContext, func(repos []Repo) ([]Repo, error) {
		if _, dup := findRepo(repos, r.Name); dup {
			return nil, errors.New("A repository with this name already exists.")
		}
		return append(repos, r), nil
	})
	if err != nil {
		_ = removeRepo(userContext, r.Name)
		return r, "", err
	}

	if r.Source == SourceRemote && r.DeployMode == DeployAutomatic {
		if derr := deployAndRecord(ctx, userContext, r); derr != nil {
			return r, pushPassword, errors.New("Repository added, but the first deployment failed: " + derr.Error())
		}
	}
	return r, pushPassword, nil
}

// updateSettings applies the settings form to an existing repo
func updateSettings(ctx context.Context, userContext, name string, v url.Values) error {
	current, ok := getRepo(userContext, name)
	if !ok {
		return errors.New("Repository not found.")
	}
	next := current

	if b := strings.TrimSpace(v.Get("branch")); b != "" {
		if !isValidBranch(b) {
			return errors.New("Invalid branch name.")
		}
		if current.Source == SourceRemote {
			known := false
			for _, x := range listBranches(ctx, userContext, name) {
				known = known || x == b
			}
			if !known {
				return errors.New("Branch not found, pull updates first.")
			}
		}
		next.Branch = b
	}
	if t := strings.TrimSpace(v.Get("target_dir")); t != "" {
		cleaned, ok := cleanTargetDir(t)
		if !ok {
			return errors.New("Target directory must be inside /var/www/html/.")
		}
		if _, err := targetHostPath(userContext, cleaned); err != nil {
			return errors.New("Invalid target directory.")
		}
		next.TargetDir = cleaned
	}
	if m := strings.TrimSpace(v.Get("deploy_mode")); m != "" {
		if !isValidDeployMode(m) {
			return errors.New("Invalid deployment mode.")
		}
		next.DeployMode = m
	}
	// only touch the parts the form actually sent, so a bulk mode change keeps deploy actions and credentials
	if v.Get("actions_present") != "" {
		if err := applyActions(ctx, userContext, v, &next); err != nil {
			return err
		}
	}
	if _, sent := v["remote_user"]; sent && current.Source == SourceRemote && !isSSHURL(current.RemoteURL) {
		next.RemoteUser = strings.TrimSpace(v.Get("remote_user"))
		// blank keeps the saved password, so the form never has to echo it back
		if p := v.Get("remote_pass"); p != "" {
			next.RemotePass = p
		}
		if next.RemoteUser == "" {
			next.RemotePass = ""
		}
	}

	return saveRepoFields(userContext, name, func(x *Repo) {
		x.Branch, x.TargetDir, x.DeployMode = next.Branch, next.TargetDir, next.DeployMode
		x.Actions, x.ActionsContainer = next.Actions, next.ActionsContainer
		x.RemoteUser, x.RemotePass = next.RemoteUser, next.RemotePass
	})
}

// resetPushPassword issues a new HTTPS password for a local repo
func resetPushPassword(userContext, name string) (string, error) {
	r, ok := getRepo(userContext, name)
	if !ok {
		return "", errors.New("Repository not found.")
	}
	if r.Source != SourceLocal {
		return "", errors.New("Only local repositories have a push password.")
	}
	pw := randomToken(24)
	hash, err := hashPassword(pw)
	if err != nil {
		return "", err
	}
	return pw, saveRepoFields(userContext, name, func(x *Repo) { x.PushPassHash = hash })
}

func deleteRepo(userContext, name string) error {
	if _, ok := getRepo(userContext, name); !ok {
		return errors.New("Repository not found.")
	}
	unlock := lockRepo(userContext, name)
	defer unlock()
	if err := removeRepo(userContext, name); err != nil {
		return err
	}
	return updateRepos(userContext, func(repos []Repo) ([]Repo, error) {
		i, ok := findRepo(repos, name)
		if !ok {
			return repos, nil
		}
		return append(repos[:i], repos[i+1:]...), nil
	})
}

// actionResult is what one panel action hands back to the page, the site tab or the API
type actionResult struct {
	Message     string `json:"message"`
	Repo        string `json:"repo,omitempty"`
	NewPassword string `json:"push_password,omitempty"`
}

// doAction runs one create/pull/deploy/settings/reset_password/delete action from form values
func doAction(ctx context.Context, a *appctx.App, r *http.Request, userID int, username, userContext string, v url.Values) (actionResult, error) {
	name := strings.TrimSpace(v.Get("repo"))
	ip := reqip.ClientIP(r)
	tr := func(msg string, kv ...any) string { return web.Tr(a, r, msg, kv...) }
	switch v.Get("action") {
	case "create":
		repo, pw, err := createRepo(ctx, a, userID, userContext, v)
		if repo.CreatedAt == "" {
			return actionResult{}, err
		}
		_ = logger.RecordUserAction(a.Config, username, "added git repository "+repo.Name, ip)
		res := actionResult{Repo: repo.Name, NewPassword: pw}
		switch {
		case err != nil:
			// repo exists but its first deploy failed, still hand back the password
			res.Message = tr("Repository %(repo)s was added, but its first deployment failed: %(error)s", "repo", repo.Name, "error", err.Error())
		case repo.Source == SourceLocal:
			res.Message = tr("Local repository %(repo)s created. Push to branch %(branch)s to deploy it to %(dir)s.", "repo", repo.Name, "branch", repo.Branch, "dir", repo.TargetDir)
		case repo.DeployMode == DeployAutomatic:
			res.Message = tr("Repository %(repo)s cloned and branch %(branch)s deployed to %(dir)s.", "repo", repo.Name, "branch", repo.Branch, "dir", repo.TargetDir)
		default:
			res.Message = tr("Repository %(repo)s cloned (branch %(branch)s), nothing was deployed yet.", "repo", repo.Name, "branch", repo.Branch)
		}
		return res, nil

	case "pull":
		repo, ok := getRepo(userContext, name)
		if !ok {
			return actionResult{}, errors.New(tr("Repository %(repo)s not found.", "repo", name))
		}
		before, _ := branchCommit(ctx, userContext, repo.Name, repo.Branch)
		if err := pullAndRecord(ctx, userContext, repo); err != nil {
			return actionResult{}, errors.New(tr("Pulling %(repo)s failed: %(error)s", "repo", name, "error", err.Error()))
		}
		_ = logger.RecordUserAction(a.Config, username, "pulled git repository "+name, ip)
		after, _ := branchCommit(ctx, userContext, repo.Name, repo.Branch)
		if after == before {
			return actionResult{Repo: name, Message: tr("%(repo)s is already up to date, no new commits on %(branch)s.", "repo", name, "branch", repo.Branch)}, nil
		}
		count := newCommitCount(ctx, userContext, repo.Name, before, after)
		if repo.DeployMode == DeployAutomatic {
			return actionResult{Repo: name, Message: tr("Pulled %(count)s new commit(s) into %(repo)s on %(branch)s and deployed %(commit)s to %(dir)s.", "count", count, "repo", name, "branch", repo.Branch, "commit", shortHash(after), "dir", repo.TargetDir)}, nil
		}
		return actionResult{Repo: name, Message: tr("Pulled %(count)s new commit(s) into %(repo)s on %(branch)s, latest is %(commit)s. Not deployed yet.", "count", count, "repo", name, "branch", repo.Branch, "commit", shortHash(after))}, nil

	case "deploy":
		repo, ok := getRepo(userContext, name)
		if !ok {
			return actionResult{}, errors.New(tr("Repository %(repo)s not found.", "repo", name))
		}
		if err := deployAndRecord(ctx, userContext, repo); err != nil {
			return actionResult{}, errors.New(tr("Deploying %(repo)s failed: %(error)s", "repo", name, "error", err.Error()))
		}
		_ = logger.RecordUserAction(a.Config, username, "deployed git repository "+name+" to "+repo.TargetDir, ip)
		commit, _ := branchCommit(ctx, userContext, repo.Name, repo.Branch)
		msg := tr("Deployed %(repo)s (%(branch)s at %(commit)s) to %(dir)s.", "repo", name, "branch", repo.Branch, "commit", shortHash(commit), "dir", repo.TargetDir)
		if repo.Actions != "" {
			msg += " " + tr("Deploy actions ran in %(container)s.", "container", repo.ActionsContainer)
		}
		return actionResult{Repo: name, Message: msg}, nil

	case "settings":
		if err := updateSettings(ctx, userContext, name, v); err != nil {
			return actionResult{}, errors.New(tr("Could not save settings for %(repo)s: %(error)s", "repo", name, "error", tr(err.Error())))
		}
		_ = logger.RecordUserAction(a.Config, username, "changed settings of git repository "+name, ip)
		saved, _ := getRepo(userContext, name)
		return actionResult{Repo: name, Message: tr("Settings saved for %(repo)s: branch %(branch)s, %(mode)s, deploying to %(dir)s.", "repo", name, "branch", saved.Branch, "mode", tr(modeLabel(saved.DeployMode)), "dir", saved.TargetDir)}, nil

	case "reset_password":
		pw, err := resetPushPassword(userContext, name)
		if err != nil {
			return actionResult{}, err
		}
		_ = logger.RecordUserAction(a.Config, username, "reset push password of git repository "+name, ip)
		return actionResult{Repo: name, NewPassword: pw, Message: tr("New push password generated for %(repo)s, the old one no longer works.", "repo", name)}, nil

	case "delete":
		repo, _ := getRepo(userContext, name)
		if err := deleteRepo(userContext, name); err != nil {
			return actionResult{}, errors.New(tr("Could not remove %(repo)s: %(error)s", "repo", name, "error", tr(err.Error())))
		}
		_ = logger.RecordUserAction(a.Config, username, "removed git repository "+name, ip)
		return actionResult{Repo: name, Message: tr("Repository %(repo)s removed. Files already deployed to %(dir)s were left in place.", "repo", name, "dir", repo.TargetDir)}, nil
	}
	return actionResult{}, errors.New("Invalid action.")
}

// handleGitPage serves /git and its form posts, passwords are rendered once instead of going through a flash
func handleGitPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodPost {
		renderGitPage(a, w, r, userID, username, userContext, pageExtras{LogRepo: r.URL.Query().Get("log")})
		return
	}
	parseForm(r)
	res, err := doAction(r.Context(), a, r, userID, username, userContext, r.Form)
	if err != nil {
		web.FlashRedirect(a, w, r, "error", err.Error(), "/git")
		return
	}
	if res.NewPassword != "" {
		web.Flash(a, w, r, "success", res.Message)
		renderGitPage(a, w, r, userID, username, userContext, pageExtras{NewPassword: res.NewPassword, NewPasswordRepo: res.Repo})
		return
	}
	web.FlashRedirect(a, w, r, "success", res.Message, "/git")
}

// handleGitJSON is the site manager tab's twin of the page form posts
func handleGitJSON(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	parseForm(r)
	res, err := doAction(r.Context(), a, r, userID, username, userContext, r.Form)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": web.Tr(a, r, err.Error())})
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

func modeLabel(m string) string {
	switch m {
	case DeployAutomatic:
		return "automatic deployment"
	case DeployManual:
		return "manual deployment"
	}
	return "no deployment"
}

func shortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

// parseForm handles both url-encoded posts and the multipart ones FormData sends from the site tab
func parseForm(r *http.Request) {
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		_ = r.ParseForm()
	}
}
