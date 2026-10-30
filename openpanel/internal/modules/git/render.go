package git

import (
	"context"
	"log"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var gitPage = web.LoadPage("manager/git.html")

// RepoView is a Repo plus what the UI shows next to it, secrets left out
type RepoView struct {
	Name             string   `json:"name"`
	Source           string   `json:"source"`
	RemoteURL        string   `json:"remote_url"`
	RemoteUser       string   `json:"remote_user"`
	HasRemotePass    bool     `json:"has_remote_pass"`
	IsSSH            bool     `json:"is_ssh"`
	Branch           string   `json:"branch"`
	Branches         []string `json:"branches"`
	Domain           string   `json:"domain"`
	TargetDir        string   `json:"target_dir"`
	DeployMode       string   `json:"deploy_mode"`
	Actions          string   `json:"actions"`
	ActionsContainer string   `json:"actions_container"`
	CloneURL         string   `json:"clone_url"`
	WebhookURL       string   `json:"webhook_url"`
	CreatedAt        string   `json:"created_at"`
	LastPull         string   `json:"last_pull"`
	LastDeploy       string   `json:"last_deploy"`
	LastDeployISO    string   `json:"last_deploy_iso"`
	LastDeployed     string   `json:"last_deployed_commit"`
	LastError        string   `json:"last_error"`
	LastOutput       string   `json:"last_output"`
	Latest           *Commit  `json:"latest"`
	UpToDate         bool     `json:"up_to_date"`
	BulkSkip         string   `json:"-"`
}

func buildViews(ctx context.Context, userContext, username, baseURL string, repos []Repo, domain string) []RepoView {
	views := make([]RepoView, 0, len(repos))
	for _, r := range repos {
		if domain != "" && r.Domain != domain {
			continue
		}
		v := RepoView{
			Name: r.Name, Source: r.Source, RemoteURL: r.RemoteURL, RemoteUser: r.RemoteUser, HasRemotePass: r.RemotePass != "",
			IsSSH: isSSHURL(r.RemoteURL), Branch: r.Branch, Branches: listBranches(ctx, userContext, r.Name), Domain: r.Domain,
			TargetDir: r.TargetDir, DeployMode: r.DeployMode, Actions: r.Actions, ActionsContainer: r.ActionsContainer,
			CreatedAt: r.CreatedAt, LastPull: r.LastPull, LastDeploy: r.LastDeploy, LastDeployISO: isoTime(r.LastDeploy), LastDeployed: r.LastDeployed,
			LastError: r.LastError, LastOutput: r.LastOutput,
		}
		if r.Source == SourceLocal {
			v.CloneURL = baseURL + strings.TrimPrefix(httpPrefix, "/") + username + "/" + r.Name + ".git"
		} else {
			v.WebhookURL = baseURL + strings.TrimPrefix(webhookPrefix, "/") + username + "/" + r.WebhookToken
		}
		var skip []string
		if r.Source == SourceLocal {
			skip = append(skip, "pull")
		}
		if r.DeployMode == DeployNone {
			skip = append(skip, "deploy")
		}
		v.BulkSkip = strings.Join(skip, ",")
		if c, ok := latestCommit(ctx, userContext, r); ok {
			v.Latest = &c
			v.UpToDate = c.Hash == r.LastDeployed
		}
		views = append(views, v)
	}
	return views
}

// pageExtras carries one-off things a POST wants rendered without a redirect
type pageExtras struct {
	NewPassword     string
	NewPasswordRepo string
	LogRepo         string
}

// GitPageData is manager/git.html's template context
type GitPageData struct {
	web.LayoutData
	Repos      []RepoView
	Domains    []appctx.Domain
	Containers []string
	PublicKey  string
	Username   string
	pageExtras
	Log         []Commit
	LogView     *RepoView
	BulkActions []web.BulkAction
}

// RepoNames are the taken names, the add form flags a duplicate before posting
func (d GitPageData) RepoNames() []string {
	names := make([]string, 0, len(d.Repos))
	for _, r := range d.Repos {
		names = append(names, r.Name)
	}
	return names
}

func renderGitPage(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, username, userContext string, extras pageExtras) {
	ctx := r.Context()
	layout, _, err := web.BuildLayoutData(a, w, r, "Git")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	repos, err := loadRepos(userContext)
	if err != nil {
		log.Printf("GIT - %s: could not read repos.json: %v", username, err)
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	pub, err := PublicKey(ctx, userContext)
	if err != nil {
		log.Printf("GIT - %s: could not generate deploy key: %v", username, err)
	}

	data := GitPageData{
		LayoutData:  layout,
		Repos:       buildViews(ctx, userContext, username, web.PublicBaseURL(ctx, a, r), repos, ""),
		Domains:     domains,
		Containers:  containersFor(ctx, userContext),
		PublicKey:   pub,
		Username:    username,
		pageExtras:  extras,
		BulkActions: gitBulkActions(web.RequestTranslator(a, r)),
	}
	if extras.LogRepo != "" {
		for i := range data.Repos {
			if data.Repos[i].Name == extras.LogRepo {
				data.LogView = &data.Repos[i]
				if repo, ok := getRepo(userContext, extras.LogRepo); ok {
					data.Log, _ = commitLog(ctx, userContext, repo, 100)
				}
				break
			}
		}
	}
	if err := gitPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("GIT - template render error: %v", err)
	}
}

// handleSiteJSON feeds the site manager's Git tab with this domain's repos
func handleSiteJSON(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	domain := r.URL.Query().Get("domain")
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": web.Tr(a, r, "You do not own this domain.")})
		return
	}
	repos, err := loadRepos(userContext)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	pub, _ := PublicKey(ctx, userContext)
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"repos":      buildViews(ctx, userContext, username, web.PublicBaseURL(ctx, a, r), repos, domain),
		"containers": containersFor(ctx, userContext),
		"public_key": pub,
	})
}
