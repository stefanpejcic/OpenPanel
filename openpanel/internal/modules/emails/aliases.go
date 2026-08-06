package emails

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// AliasEntry is one parsed alias source -> targets mapping.
type AliasEntry struct {
	Source  string
	Targets []string
}

func aliasesCacheFile(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/aliases.yml"
}

// readAliasesFile parses the cached aliases.yml, keeping only entries whose source domain belongs to userDomains.
func readAliasesFile(username string, userDomains map[string]bool) []AliasEntry {
	data, err := os.ReadFile(aliasesCacheFile(username))
	if err != nil {
		return nil
	}
	var aliases []AliasEntry
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "*") {
			continue
		}
		parts := strings.Fields(strings.TrimLeft(line, "* "))
		if len(parts) < 2 {
			continue
		}
		source := parts[0]
		targets := strings.Split(parts[1], ",")
		sourceDomain := ""
		if _, d, ok := strings.Cut(source, "@"); ok {
			sourceDomain = d
		}
		if userDomains[sourceDomain] {
			aliases = append(aliases, AliasEntry{Source: source, Targets: targets})
		}
	}
	return aliases
}

// GetAliasList returns the user's alias list, cached 1h; the cache is
// (re)populated from `opencli` on first access.
func GetAliasList(ctx context.Context, a *appctx.App, userID int, username string, userDomains map[string]bool) []AliasEntry {
	result, _ := cache.Memoize(ctx, a.Cache, "alias_list:"+username, time.Hour, func() ([]AliasEntry, error) {
		if _, err := os.Stat(aliasesCacheFile(username)); err != nil {
			domains, _ := a.AllDomainsForUser(ctx, userID)
			ImportUserAliases(username, domainSet(domains))
		}
		return readAliasesFile(username, userDomains), nil
	})
	return result
}

// InvalidateAliasCache drops the cached alias list and re-imports it from `opencli` immediately.
func InvalidateAliasCache(ctx context.Context, a *appctx.App, userID int, username string) {
	_ = a.Cache.Delete(ctx, "alias_list:"+username)
	domains, _ := a.AllDomainsForUser(ctx, userID)
	ImportUserAliases(username, domainSet(domains))
}

// ImportUserAliases rewrites the user's cached aliases.yml from the
// output of `opencli email-setup alias list`, keeping only aliases whose
// source domain is in userDomains.
func ImportUserAliases(currentUsername string, userDomains map[string]bool) {
	path := aliasesCacheFile(currentUsername)
	if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, nil, 0o644)
	}

	out, err := exec.Command("opencli", "email-setup", "alias", "list").Output()
	if err != nil {
		return
	}

	var filtered []string
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "*") {
			continue
		}
		parts := strings.Fields(strings.TrimLeft(line, "* "))
		if len(parts) == 0 {
			continue
		}
		source := parts[0]
		sourceDomain := ""
		if _, d, ok := strings.Cut(source, "@"); ok {
			sourceDomain = d
		}
		if userDomains[sourceDomain] {
			filtered = append(filtered, line)
		}
	}

	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(strings.Join(filtered, "\n")+"\n"), 0o644)
}

// ---------------------------------------------------------------------------
// routes
// ---------------------------------------------------------------------------

// handleAliases lists all aliases for the current user's domains.
func handleAliases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	aliasList := GetAliasList(ctx, a, userID, currentUsername, userDomains)

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"aliases": aliasList})
		return
	}

	renderAliasesPage(a, w, r, aliasList, domains)
}

// handleAliasDetail dispatches GET/POST/DELETE /emails/aliases/{email}.
func handleAliasDetail(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.PathValue("email")
	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	_, domain, _ := strings.Cut(email, "@")
	if !userDomains[domain] {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getAlias(a, w, r, email, userID, currentUsername, userDomains)
	case http.MethodPost:
		postAlias(a, w, r, email, userID, currentUsername)
	case http.MethodDelete:
		deleteAlias(a, w, r, email, userID, currentUsername, userDomains)
	}
}

func getAlias(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string, userDomains map[string]bool) {
	ctx := r.Context()
	aliasList := GetAliasList(ctx, a, userID, currentUsername, userDomains)
	var entry *AliasEntry
	for i := range aliasList {
		if aliasList[i].Source == email {
			entry = &aliasList[i]
			break
		}
	}

	if r.URL.Query().Get("output") == "json" {
		targets := []string{}
		if entry != nil {
			targets = entry.Targets
		}
		writeJSON(w, http.StatusOK, map[string]any{"source": email, "targets": targets})
		return
	}

	renderAliasDetailPage(a, w, r, email, entry)
}

func postAlias(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string) {
	ctx := r.Context()
	_ = r.ParseForm()
	target := strings.TrimSpace(r.Form.Get("target"))

	if target == "" || !isValidEmail(target) {
		flashAndRedirect(a, w, r, "error", "Error: a valid target email address is required.", "/emails/aliases/"+email)
		return
	}

	out, err := exec.CommandContext(ctx, "opencli", "email-setup", "alias", "add", email, target).CombinedOutput()
	if err != nil {
		flashAndRedirect(a, w, r, "error", "Error adding alias: "+strings.TrimSpace(string(out)), "/emails/aliases/"+email)
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "added alias "+email+" -> "+target, ipAddress)
	InvalidateAliasCache(ctx, a, userID, currentUsername)
	flashAndRedirect(a, w, r, "success", "Alias "+email+" → "+target+" added successfully.", "/emails/aliases/"+email)
}

func deleteAlias(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, userID int, currentUsername string, userDomains map[string]bool) {
	ctx := r.Context()

	var target string
	var deleteAll bool

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body struct {
			Target    string `json:"target"`
			DeleteAll bool   `json:"delete_all"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		target = strings.TrimSpace(body.Target)
		deleteAll = body.DeleteAll
	} else {
		_ = r.ParseForm()
		target = strings.TrimSpace(r.Form.Get("target"))
	}

	if deleteAll {
		aliasList := GetAliasList(ctx, a, userID, currentUsername, userDomains)
		var entry *AliasEntry
		for i := range aliasList {
			if aliasList[i].Source == email {
				entry = &aliasList[i]
				break
			}
		}
		if entry != nil {
			for _, t := range entry.Targets {
				if out, err := exec.CommandContext(ctx, "opencli", "email-setup", "alias", "del", email, t).CombinedOutput(); err != nil {
					flashAndRedirect(a, w, r, "error", "Error deleting alias "+email+" -> "+t+": "+strings.TrimSpace(string(out)), "/emails/aliases")
					return
				}
			}
		}
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "deleted alias "+email, ipAddress)
		InvalidateAliasCache(ctx, a, userID, currentUsername)
		flashAndRedirect(a, w, r, "success", "Alias "+email+" deleted successfully.", "/emails/aliases")
		return
	}

	if target == "" || !isValidEmail(target) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "A valid target email address is required."})
		return
	}

	_, _ = exec.CommandContext(ctx, "opencli", "email-setup", "alias", "del", email, target).CombinedOutput()

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted alias "+email+" -> "+target, ipAddress)
	InvalidateAliasCache(ctx, a, userID, currentUsername)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// ---------------------------------------------------------------------------
// create alias (new source address)
// ---------------------------------------------------------------------------

// handleAliasNew creates a new alias from a username/domain/target form submission.
func handleAliasNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		username := strings.TrimSpace(r.Form.Get("username"))
		domain := strings.TrimSpace(r.Form.Get("domain"))
		target := strings.TrimSpace(r.Form.Get("target"))

		var missing []string
		if username == "" {
			missing = append(missing, "username")
		}
		if domain == "" {
			missing = append(missing, "domain")
		}
		if target == "" {
			missing = append(missing, "target")
		}
		if len(missing) > 0 {
			for _, field := range missing {
				flashSess(a, w, r, "error", "Error: "+field+" not provided.")
			}
			http.Redirect(w, r, "/emails/aliases/new", http.StatusFound)
			return
		}

		if !isValidEmail(target) {
			flashAndRedirect(a, w, r, "error", "Error: target must be a valid email address.", "/emails/aliases/new")
			return
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		source := username + "@" + domain

		out, cmdErr := exec.CommandContext(ctx, "opencli", "email-setup", "alias", "add", source, target).CombinedOutput()
		if cmdErr != nil {
			flashAndRedirect(a, w, r, "error", "Error creating alias: "+strings.TrimSpace(string(out)), "/emails/aliases/new")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created alias "+source+" -> "+target, ipAddress)
		InvalidateAliasCache(ctx, a, userID, currentUsername)
		flashAndRedirect(a, w, r, "success", "Alias "+source+" → "+target+" created successfully.", "/emails/aliases")
		return
	}

	renderAliasNewPage(a, w, r, domains)
}

// ---------------------------------------------------------------------------
// delete alias page (confirmation view)
// ---------------------------------------------------------------------------

// handleAliasDeletePage renders the alias deletion confirmation view,
// either for a single alias or the full list.
func handleAliasDeletePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	email := r.PathValue("email")
	var aliasList []AliasEntry
	if email == "" {
		domains, _ := a.AllDomainsForUser(ctx, userID)
		aliasList = GetAliasList(ctx, a, userID, currentUsername, domainSet(domains))
	} else if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	renderAliasDeletePage(a, w, r, email, aliasList)
}
