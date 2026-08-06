package domains

import (
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

var httpURLRE = regexp.MustCompile(`^(https?://)`)

func domainConfPath(domainURL string) string {
	return "/etc/openpanel/caddy/domains/" + domainURL + ".conf"
}

func reloadCaddyWebserver(r *http.Request) {
	_ = exec.CommandContext(r.Context(), "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run()
}

// handleDeleteRedirect removes a domain's redirect rule.
func handleDeleteRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	domainURL := r.Form.Get("domain_name")
	redirectURL := r.Form.Get("redirect_url")

	if !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	if redirectURL == "" {
		flashAndRedirect(a, w, r, "error", "Domain "+domainURL+" does not have a redirect URL configured.", "/domains")
		return
	}

	var ownerUserID int
	var actualDomainURL string
	if scanErr := a.DB.QueryRowContext(ctx, "SELECT user_id, domain_url FROM domains WHERE domain_url = ?", domainURL).Scan(&ownerUserID, &actualDomainURL); scanErr != nil || ownerUserID != userID {
		flashAndRedirect(a, w, r, "error", "Unauthorized", "/domains")
		return
	}
	domainURL = actualDomainURL

	path := domainConfPath(domainURL)
	content, readErr := readTextFile(path)
	if readErr != nil {
		flashAndRedirect(a, w, r, "error", readErr.Error(), "/domains")
		return
	}

	redirectLine := "redir " + redirectURL
	newContent := strings.ReplaceAll(content, redirectLine, "")

	if writeErr := writeTextFile(path, newContent); writeErr != nil {
		flashAndRedirect(a, w, r, "error", writeErr.Error(), "/domains")
		return
	}

	reloadCaddyWebserver(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted the redirect link "+redirectURL+" for domain "+domainURL, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", "Successfully deleted the redirect link "+redirectURL+" for domain "+domainURL, "/domains")
}

// handleSetRedirect sets or replaces a domain's redirect rule.
func handleSetRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainURL := r.Form.Get("domain_name")
		redirectURL := r.Form.Get("redirect_url")

		if domainURL == "" {
			flashAndRedirect(a, w, r, "error", "Domain name not provided.", "/domains")
			return
		}
		if !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}
		if !httpURLRE.MatchString(redirectURL) {
			flashAndRedirect(a, w, r, "error", "Invalid URL. Please provide a URL with 'http://' or 'https://' prefix.", "/domains")
			return
		}

		path := domainConfPath(domainURL)
		content, readErr := readTextFile(path)
		if readErr != nil {
			http.Error(w, readErr.Error(), http.StatusInternalServerError)
			return
		}

		updated := insertOrReplaceRedirect(content, redirectURL)
		if writeErr := writeTextFile(path, updated); writeErr != nil {
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}

		reloadCaddyWebserver(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a redirect link "+redirectURL+" for domain "+domainURL, reqip.ClientIP(r))
		flashAndRedirect(a, w, r, "success", "Successfully created redirect from domain "+domainURL+" to "+redirectURL, "/domains")
		return
	}

	domainName := r.URL.Query().Get("domain")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderRedirectSelectPage(a, w, r, domainsList)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	redirectURL := getRedirectURL(domainName)
	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	renderRedirectEditPage(a, w, r, domainName, redirectURL, domainsList)
}

// insertOrReplaceRedirect finds the `log {`/`import domain_log` anchor
// line(s) in a vhost config and either replaces an existing `redir ` line
// immediately before the anchor, or inserts a new one - supporting both the
// legacy inline `log {` block and the newer `import domain_log` style (see
// openpanel/openpanel#645).
func insertOrReplaceRedirect(content, redirectURL string) string {
	lines := strings.Split(content, "\n")
	redirectLine := "redir " + redirectURL

	var anchors []int
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "log {") || strings.HasPrefix(trimmed, "import domain_log") {
			anchors = append(anchors, i)
		}
	}

	for i := len(anchors) - 1; i >= 0; i-- {
		idx := anchors[i]
		indent := lines[idx][:len(lines[idx])-len(strings.TrimLeft(lines[idx], " \t"))]
		if idx > 0 && strings.HasPrefix(strings.TrimLeft(lines[idx-1], " \t"), "redir ") {
			lines[idx-1] = indent + redirectLine
		} else {
			before := append([]string{}, lines[:idx]...)
			after := append([]string{}, lines[idx:]...)
			lines = append(before, append([]string{indent + redirectLine}, after...)...)
		}
	}

	return strings.Join(lines, "\n")
}

func readTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	return string(content), err
}

func writeTextFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
