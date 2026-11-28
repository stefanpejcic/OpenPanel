package git

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/githubapi"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	ProviderGitHub = "github"

	apiTimeout = 15 * time.Second
	// 100 per page, so the picker lists up to 1000 repositories and 300 branches
	maxRepoPages   = 10
	maxBranchPages = 3
	hostedCacheTTL = time.Minute
)

// githubCloneBase is swapped out in tests that clone from a loopback server
var githubCloneBase = "https://github.com/"

// Account is one connected Git hosting account in accounts.json
type Account struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Login    string `json:"login"`
	Label    string `json:"label,omitempty"`
	Token    string `json:"token"`
	Expires  string `json:"token_expires,omitempty"`
	AddedAt  string `json:"added_at"`
}

// Title is what the UI calls the account
func (a Account) Title() string {
	if a.Label != "" {
		return a.Label
	}
	return a.Login
}

// HostedRepo is one repository an account can reach on GitHub
type HostedRepo struct {
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
	Description   string `json:"description"`
}

var (
	tokenRE    = regexp.MustCompile(`^[A-Za-z0-9_]{20,255}$`)
	loginRE    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,99}$`)
	fullNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,99}/[A-Za-z0-9._-]{1,100}$`)
	accountRE  = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)

	errAccountGone     = errors.New("the GitHub account this repository uses is no longer connected, pick another one in the repository settings")
	errHookUnreachable = errors.New("GitHub can only deliver webhooks to a panel on a domain with a valid certificate")
)

func isValidFullName(s string) bool {
	return fullNameRE.MatchString(s) && !strings.Contains(s, "..") && !strings.HasSuffix(s, ".git")
}

func accountsPath(userContext string) string {
	return filepath.Join(accountRoot(userContext), "accounts.json")
}

// loadAccounts reads accounts.json, a missing file is just an empty list
func loadAccounts(userContext string) ([]Account, error) {
	if err := checkAccountRoot(userContext); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Account{}, nil
		}
		return nil, err
	}
	data, err := os.ReadFile(accountsPath(userContext))
	if errors.Is(err, os.ErrNotExist) {
		return []Account{}, nil
	}
	if err != nil {
		return nil, err
	}
	var accounts []Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// updateAccounts runs fn on the account list under the same lock repos.json uses and saves what it leaves behind
func updateAccounts(userContext string, fn func([]Account) ([]Account, error)) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	f, err := lockFile(userContext)
	if err != nil {
		return err
	}
	defer unlockFile(f)
	accounts, err := loadAccounts(userContext)
	if err != nil {
		return err
	}
	accounts, err = fn(accounts)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}
	tmp := accountsPath(userContext) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, accountsPath(userContext))
}

func getAccount(userContext, id string) (Account, bool) {
	if !accountRE.MatchString(id) {
		return Account{}, false
	}
	accounts, err := loadAccounts(userContext)
	if err != nil {
		return Account{}, false
	}
	for _, a := range accounts {
		if a.ID == id {
			return a, true
		}
	}
	return Account{}, false
}

// cleanLabel keeps an optional display name short and free of control characters
func cleanLabel(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) > 60 {
		return "", false
	}
	for _, c := range s {
		if c < 0x20 || c == 0x7f {
			return "", false
		}
	}
	return s, true
}

// apiError turns the client's errors into something a panel user can act on
func apiError(err error) error {
	switch {
	case errors.Is(err, githubapi.ErrUnauthorized):
		return errors.New("GitHub did not accept the access token. Check that it was copied in full and has not expired or been revoked.")
	case errors.Is(err, githubapi.ErrNotFound):
		return errors.New("GitHub could not find it, or the access token has no permission for it.")
	}
	return err
}

// connectAccount checks the token against GitHub and stores it as a new account
func connectAccount(ctx context.Context, userContext, token, label string) (Account, error) {
	token = strings.TrimSpace(token)
	if !tokenRE.MatchString(token) {
		return Account{}, errors.New("That does not look like a GitHub access token.")
	}
	label, ok := cleanLabel(label)
	if !ok {
		return Account{}, errors.New("The account name is too long.")
	}
	resp, err := githubapi.Do(ctx, http.MethodGet, "/user", token, nil, apiTimeout)
	if err != nil {
		return Account{}, apiError(err)
	}
	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(resp.Body, &user); err != nil || !loginRE.MatchString(user.Login) {
		return Account{}, errors.New("GitHub sent an answer the panel could not read.")
	}
	acct := Account{ID: randomToken(12), Provider: ProviderGitHub, Login: user.Login, Label: label, Token: token, AddedAt: nowStr()}
	if exp := githubapi.TokenExpiry(resp.Header); !exp.IsZero() {
		acct.Expires = exp.Format(timeLayout)
	}
	err = updateAccounts(userContext, func(accounts []Account) ([]Account, error) {
		for _, a := range accounts {
			if a.Token == token {
				return nil, errors.New("This access token is already connected.")
			}
		}
		if gitMaxAccounts > 0 && len(accounts) >= gitMaxAccounts {
			return nil, errors.New("Connected account limit reached for this account.")
		}
		return append(accounts, acct), nil
	})
	return acct, err
}

// disconnectAccount forgets the token, taking the webhooks it registered with it, and reports how many repositories relied on it
func disconnectAccount(ctx context.Context, userContext, id string) (Account, int, error) {
	acct, ok := getAccount(userContext, id)
	if !ok {
		return Account{}, 0, errors.New("Account not found.")
	}
	repos, err := loadRepos(userContext)
	if err != nil {
		return acct, 0, err
	}
	hctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	used := 0
	for _, r := range repos {
		if r.AccountID == id {
			used++
			dropHook(hctx, userContext, r)
		}
	}
	err = updateAccounts(userContext, func(accounts []Account) ([]Account, error) {
		kept := accounts[:0]
		for _, a := range accounts {
			if a.ID != id {
				kept = append(kept, a)
			}
		}
		return kept, nil
	})
	hostedCache.Delete(userContext + "/" + id)
	return acct, used, err
}

type hostedList struct {
	at        time.Time
	repos     []HostedRepo
	truncated bool
}

// hostedCache keeps the picker from spending API quota every time the form is opened
var hostedCache sync.Map

// listHostedRepos returns what the account can reach, most recently pushed first
func listHostedRepos(ctx context.Context, userContext string, acct Account, refresh bool) ([]HostedRepo, bool, error) {
	key := userContext + "/" + acct.ID
	if v, ok := hostedCache.Load(key); ok && !refresh {
		if l := v.(hostedList); time.Since(l.at) < hostedCacheTTL {
			return l.repos, l.truncated, nil
		}
	}
	repos, truncated := []HostedRepo{}, false
	for page := 1; ; page++ {
		resp, err := githubapi.Do(ctx, http.MethodGet, "/user/repos?per_page=100&sort=pushed&page="+strconv.Itoa(page), acct.Token, nil, apiTimeout)
		if err != nil {
			return nil, false, apiError(err)
		}
		var batch []HostedRepo
		if err := json.Unmarshal(resp.Body, &batch); err != nil {
			return nil, false, errors.New("GitHub sent an answer the panel could not read.")
		}
		for _, r := range batch {
			if isValidFullName(r.FullName) {
				repos = append(repos, r)
			}
		}
		if len(batch) < 100 {
			break
		}
		if page == maxRepoPages {
			truncated = true
			break
		}
	}
	hostedCache.Store(key, hostedList{at: time.Now(), repos: repos, truncated: truncated})
	return repos, truncated, nil
}

// hostedRepo looks one repository up, which also proves the account can read it
func hostedRepo(ctx context.Context, acct Account, fullName string) (HostedRepo, error) {
	if !isValidFullName(fullName) {
		return HostedRepo{}, errors.New("invalid repository name")
	}
	resp, err := githubapi.Do(ctx, http.MethodGet, "/repos/"+fullName, acct.Token, nil, apiTimeout)
	if err != nil {
		return HostedRepo{}, apiError(err)
	}
	var r HostedRepo
	// a renamed repository answers with its new name, which is the one to clone
	if err := json.Unmarshal(resp.Body, &r); err != nil || !isValidFullName(r.FullName) {
		return HostedRepo{}, errors.New("GitHub sent an answer the panel could not read.")
	}
	return r, nil
}

func listHostedBranches(ctx context.Context, acct Account, fullName string) ([]string, error) {
	if !isValidFullName(fullName) {
		return nil, errors.New("invalid repository name")
	}
	branches := []string{}
	for page := 1; page <= maxBranchPages; page++ {
		resp, err := githubapi.Do(ctx, http.MethodGet, "/repos/"+fullName+"/branches?per_page=100&page="+strconv.Itoa(page), acct.Token, nil, apiTimeout)
		if err != nil {
			return nil, apiError(err)
		}
		var batch []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(resp.Body, &batch); err != nil {
			return nil, errors.New("GitHub sent an answer the panel could not read.")
		}
		for _, b := range batch {
			if isValidBranch(b.Name) {
				branches = append(branches, b.Name)
			}
		}
		if len(batch) < 100 {
			break
		}
	}
	return branches, nil
}

// hookURLFor is the repo's webhook address when GitHub can deliver to it
func hookURLFor(baseURL, username string, r Repo) (string, bool) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "localhost" || net.ParseIP(u.Hostname()) != nil {
		return "", false
	}
	return baseURL + strings.TrimPrefix(webhookPrefix, "/") + username + "/" + r.WebhookToken, true
}

// ensureHook registers the push webhook on GitHub for a repo added through an account, true when it added one
func ensureHook(ctx context.Context, userContext, username, baseURL string, r Repo) (bool, error) {
	if r.RemoteRepo == "" || r.HookID != 0 {
		return false, nil
	}
	acct, ok := getAccount(userContext, r.AccountID)
	if !ok {
		return false, errAccountGone
	}
	hookURL, ok := hookURLFor(baseURL, username, r)
	if !ok {
		return false, errHookUnreachable
	}
	payload := map[string]any{
		"name":   "web",
		"active": true,
		"events": []string{"push"},
		"config": map[string]string{"url": hookURL, "content_type": "json"},
	}
	resp, err := githubapi.Do(ctx, http.MethodPost, "/repos/"+r.RemoteRepo+"/hooks", acct.Token, payload, apiTimeout)
	if err != nil {
		return false, apiError(err)
	}
	var hook struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(resp.Body, &hook); err != nil || hook.ID == 0 {
		return false, errors.New("GitHub sent an answer the panel could not read.")
	}
	return true, saveRepoFields(userContext, r.Name, func(x *Repo) { x.HookID = hook.ID })
}

// dropHook removes the webhook the panel registered, best effort since the token may already be gone
func dropHook(ctx context.Context, userContext string, r Repo) {
	if r.HookID == 0 || r.RemoteRepo == "" {
		return
	}
	if acct, ok := getAccount(userContext, r.AccountID); ok {
		_, _ = githubapi.Do(ctx, http.MethodDelete, "/repos/"+r.RemoteRepo+"/hooks/"+strconv.FormatInt(r.HookID, 10), acct.Token, nil, apiTimeout)
	}
	_ = saveRepoFields(userContext, r.Name, func(x *Repo) { x.HookID = 0 })
}

// AccountView is an Account for the UI, the token never leaves the server
type AccountView struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	Login       string `json:"login"`
	Title       string `json:"title"`
	AddedAt     string `json:"added_at"`
	Expires     string `json:"token_expires"`
	Expired     bool   `json:"token_expired"`
	ExpiresSoon bool   `json:"token_expires_soon"`
	Repos       int    `json:"repos"`
}

func accountViews(accounts []Account, repos []Repo) []AccountView {
	views := make([]AccountView, 0, len(accounts))
	for _, a := range accounts {
		v := AccountView{ID: a.ID, Provider: a.Provider, Login: a.Login, Title: a.Title(), AddedAt: shortDate(a.AddedAt), Expires: a.Expires}
		if t, err := time.Parse(timeLayout, a.Expires); err == nil {
			left := time.Until(t)
			v.Expired = left <= 0
			v.ExpiresSoon = left > 0 && left < 14*24*time.Hour
			v.Expires = t.Format("2006-01-02")
		}
		for _, r := range repos {
			if r.AccountID == a.ID {
				v.Repos++
			}
		}
		views = append(views, v)
	}
	return views
}

// handleHostedRepos feeds the add form's repository picker
func handleHostedRepos(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	acct, ok := getAccount(userContext, r.URL.Query().Get("account"))
	if !ok {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": web.Tr(a, r, "Account not found.")})
		return
	}
	repos, truncated, err := listHostedRepos(r.Context(), userContext, acct, r.URL.Query().Get("refresh") != "")
	if err != nil {
		web.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": web.Tr(a, r, err.Error())})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"repos": repos, "truncated": truncated})
}

// handleHostedBranches lists the picked repository's branches
func handleHostedBranches(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	acct, ok := getAccount(userContext, r.URL.Query().Get("account"))
	if !ok {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": web.Tr(a, r, "Account not found.")})
		return
	}
	full := r.URL.Query().Get("repo")
	hosted, err := hostedRepo(r.Context(), acct, full)
	if err != nil {
		web.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": web.Tr(a, r, err.Error())})
		return
	}
	branches, err := listHostedBranches(r.Context(), acct, hosted.FullName)
	if err != nil {
		web.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": web.Tr(a, r, err.Error())})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"branches": branches, "default_branch": hosted.DefaultBranch})
}

// shortDate keeps just the day of a stored timestamp
func shortDate(stored string) string {
	if t, err := time.Parse(timeLayout, stored); err == nil {
		return t.Format("2006-01-02")
	}
	return stored
}

// repoLink is what a theme needs to know about a repo's account without touching the page data
type repoLink struct {
	RemoteRepo  string `json:"remote_repo"`
	AccountID   string `json:"account_id"`
	Account     string `json:"account"`
	AccountGone bool   `json:"account_gone"`
	HookAdded   bool   `json:"hook_added"`
}

// handleAccountsJSON lists the connected accounts and which repos use them, themes that draw their own Git page read it
func handleAccountsJSON(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	accounts, err := loadAccounts(userContext)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	repos, _ := loadRepos(userContext)
	web.WriteJSON(w, http.StatusOK, map[string]any{"accounts": accountViews(accounts, repos), "repos": repoLinks(accounts, repos), "max_accounts": gitMaxAccounts})
}

func repoLinks(accounts []Account, repos []Repo) map[string]repoLink {
	titles := map[string]string{}
	for _, a := range accounts {
		titles[a.ID] = a.Title()
	}
	links := map[string]repoLink{}
	for _, r := range repos {
		if r.RemoteRepo != "" {
			links[r.Name] = repoLink{RemoteRepo: r.RemoteRepo, AccountID: r.AccountID, Account: titles[r.AccountID], AccountGone: titles[r.AccountID] == "", HookAdded: r.HookID != 0}
		}
	}
	return links
}
