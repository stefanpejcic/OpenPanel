package git

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

const (
	httpPrefix    = "/git-http/"
	webhookPrefix = "/git-webhook/"
)

// gitAccount is swapped out in tests
var gitAccount = lookupGitAccount

// lookupGitAccount resolves a username to its context, only when the account exists, isn't suspended and has git on its plan
func lookupGitAccount(ctx context.Context, a *appctx.App, username string) (string, bool) {
	if username == "" || strings.ContainsAny(username, "/\\") {
		return "", false
	}
	var server sql.NullString
	if err := a.DB.QueryRowContext(ctx, "SELECT server FROM users WHERE username = ?", username).Scan(&server); err != nil {
		return "", false
	}
	userContext := server.String
	if userContext == "" {
		userContext = username
	}
	features, err := a.LoadUserFeatures(ctx, username, userContext)
	if err != nil {
		return "", false
	}
	for _, f := range features {
		if f == "git" {
			return userContext, true
		}
	}
	return "", false
}

// splitHTTPPath turns /git-http/<user>/<name>.git/<rest> into its parts
func splitHTTPPath(p string) (user, name, rest string, ok bool) {
	parts := strings.SplitN(strings.TrimPrefix(p, httpPrefix), "/", 3)
	if len(parts) < 3 || !strings.HasSuffix(parts[1], ".git") {
		return "", "", "", false
	}
	name = strings.TrimSuffix(parts[1], ".git")
	if !isValidRepoName(name) {
		return "", "", "", false
	}
	return parts[0], name, parts[2], true
}

// isSmartRequest only lets the smart protocol through, no dumb object fetches
func isSmartRequest(r *http.Request, rest string) (service string, ok bool) {
	switch {
	case rest == "info/refs" && r.Method == http.MethodGet:
		s := r.URL.Query().Get("service")
		return s, s == "git-upload-pack" || s == "git-receive-pack"
	case rest == "git-upload-pack" && r.Method == http.MethodPost:
		return rest, true
	case rest == "git-receive-pack" && r.Method == http.MethodPost:
		return rest, true
	}
	return "", false
}

func askAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="OpenPanel Git", charset="UTF-8"`)
	http.Error(w, "Authentication required", http.StatusUnauthorized)
}

// handleSmartHTTP serves clone/fetch/push for local repos via git http-backend
func handleSmartHTTP(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	user, name, rest, ok := splitHTTPPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	service, ok := isSmartRequest(r, rest)
	if !ok {
		http.Error(w, "Only the smart HTTP protocol is supported", http.StatusForbidden)
		return
	}

	authUser, authPass, hasAuth := r.BasicAuth()
	if !hasAuth || authUser != user {
		askAuth(w)
		return
	}
	userContext, ok := gitAccount(r.Context(), a, user)
	if !ok {
		askAuth(w)
		return
	}
	repo, ok := getRepo(userContext, name)
	if !ok || repo.Source != SourceLocal || !checkPassword(repo.PushPassHash, authPass) {
		// bcrypt already makes guessing slow, this just blunts tight loops
		time.Sleep(500 * time.Millisecond)
		askAuth(w)
		return
	}

	gitBin, err := exec.LookPath("git")
	if err != nil {
		http.Error(w, "git is not installed", http.StatusInternalServerError)
		return
	}
	h := &cgi.Handler{
		Path: gitBin,
		Args: []string{"http-backend"},
		Root: httpPrefix + user,
		Env: []string{
			"GIT_PROJECT_ROOT=" + accountRoot(userContext),
			"GIT_HTTP_EXPORT_ALL=1",
			"GIT_CONFIG_NOSYSTEM=1",
			"REMOTE_USER=" + user,
			"HOME=" + accountRoot(userContext),
			"PATH=" + os.Getenv("PATH"),
		},
	}

	if service != "git-receive-pack" || rest != "git-receive-pack" {
		h.ServeHTTP(w, r)
		return
	}

	unlock := lockRepo(userContext, name)
	h.ServeHTTP(w, r)
	unlock()

	_ = saveRepoFields(userContext, name, func(x *Repo) { x.LastPull = nowStr() })
	if repo.DeployMode == DeployAutomatic {
		go func() {
			if err := deployAndRecord(context.Background(), userContext, repo); err != nil {
				log.Printf("GIT - %s/%s: deploy after push failed: %v", user, name, err)
			}
		}()
	}
}

// handleWebhook is what GitHub/GitLab/Bitbucket call on push, it pulls and deploys in the background
func handleWebhook(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	user, token := r.PathValue("user"), r.PathValue("token")
	userContext, ok := gitAccount(r.Context(), a, user)
	if !ok {
		http.NotFound(w, r)
		return
	}
	repos, err := loadRepos(userContext)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var repo Repo
	found := false
	for _, x := range repos {
		if tokensEqual(x.WebhookToken, token) {
			repo, found = x, true
			break
		}
	}
	if !found || repo.Source != SourceRemote {
		http.NotFound(w, r)
		return
	}

	go func() {
		if err := pullAndRecord(context.Background(), userContext, repo); err != nil {
			log.Printf("GIT - %s/%s: webhook pull failed: %v", user, repo.Name, err)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("queued\n"))
}
