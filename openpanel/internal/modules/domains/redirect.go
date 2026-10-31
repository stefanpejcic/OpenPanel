package domains

import (
	"context"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// httpURLRE requires the whole value to be a scheme plus at least one non-whitespace char, not just a prefix check, so a redirect_url can't smuggle a newline into the Caddyfile line insertOrReplaceRedirect() splices this into
var httpURLRE = regexp.MustCompile(`^https?://\S+$`)

func domainConfPath(domainURL string) string {
	return "/etc/openpanel/caddy/domains/" + domainURL + ".conf"
}

func reloadCaddyWebserver(r *http.Request) {
	_ = exec.CommandContext(r.Context(), "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run()
}

// redirectBack is where a redirect form goes after posting: the Redirects page when it came from there, the domains list otherwise
func redirectBack(r *http.Request, domain string, failed bool) string {
	if r.Form.Get("return_to") != "redirects" {
		return "/domains"
	}
	if failed && domain != "" {
		return "/domains/redirect?domain=" + url.QueryEscape(domain)
	}
	return "/domains/redirect"
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

	if !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	var ownerUserID int
	var actualDomainURL string
	if scanErr := a.DB.QueryRowContext(ctx, "SELECT user_id, domain_url FROM domains WHERE domain_url = ?", domainURL).Scan(&ownerUserID, &actualDomainURL); scanErr != nil || ownerUserID != userID {
		web.FlashRedirect(a, w, r, "error", "Unauthorized", redirectBack(r, "", true))
		return
	}
	domainURL = actualDomainURL

	// the conf is the truth, a stale form value would otherwise "delete" nothing and still report success
	redirectURL := getRedirectURL(domainURL)
	if redirectURL == "" {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Domain %(domain_url)s does not have a redirect URL configured.", "domain_url", domainURL), redirectBack(r, "", true))
		return
	}

	path := domainConfPath(domainURL)
	content, readErr := readTextFile(path)
	if readErr != nil {
		web.FlashRedirect(a, w, r, "error", readErr.Error(), redirectBack(r, "", true))
		return
	}

	newContent := removeRedirect(content, redirectURL)

	if writeErr := writeTextFile(path, newContent); writeErr != nil {
		web.FlashRedirect(a, w, r, "error", writeErr.Error(), redirectBack(r, "", true))
		return
	}

	reloadCaddyWebserver(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted the redirect link "+redirectURL+" for domain "+domainURL, reqip.ClientIP(r))
	web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "Successfully deleted the redirect link %(redirect_url)s for domain %(domain_url)s", "redirect_url", redirectURL, "domain_url", domainURL), redirectBack(r, "", false))
}

// handleSetRedirect lists redirects on GET and sets or replaces a domain's redirect rule on POST
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
		domainURL := strings.TrimSpace(r.Form.Get("domain_name"))
		redirectURL := strings.TrimSpace(r.Form.Get("redirect_url"))
		// left out by older forms and the API, which then get the old behaviour: 302, exact URL, www included
		rule := redirectRule{To: redirectURL, Permanent: r.Form.Get("redirect_type") == "301", KeepPath: r.Form.Get("keep_path") == "on", BareOnly: r.Form.Get("www") == "no"}

		if domainURL == "" {
			web.FlashRedirect(a, w, r, "error", "Domain name not provided.", redirectBack(r, "", true))
			return
		}
		// the ownership check stops at the first slash, the conf path doesn't
		if strings.ContainsAny(domainURL, "/\\") || !a.CheckDomainBelongsToUser(ctx, userID, domainURL) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}
		existing, _ := userRedirects(ctx, a, userID)
		if msg, other := redirectTargetProblem(domainURL, redirectURL, redirectTargets(existing)); msg != "" {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, msg, "domain", domainURL, "other", other), redirectBack(r, domainURL, true))
			return
		}

		path := domainConfPath(domainURL)
		content, readErr := readTextFile(path)
		if readErr != nil {
			web.FlashRedirect(a, w, r, "error", readErr.Error(), redirectBack(r, domainURL, true))
			return
		}

		updated := insertOrReplaceRedirect(content, rule, domainURL)
		if writeErr := writeTextFile(path, updated); writeErr != nil {
			web.FlashRedirect(a, w, r, "error", writeErr.Error(), redirectBack(r, domainURL, true))
			return
		}

		reloadCaddyWebserver(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a "+rule.describe()+" redirect link "+redirectURL+" for domain "+domainURL, reqip.ClientIP(r))
		web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "Successfully created redirect from domain %(domain_url)s to %(redirect_url)s", "domain_url", domainURL, "redirect_url", redirectURL), redirectBack(r, domainURL, false))
		return
	}

	// ?domain= opens that domain's form straight away, old links to the per-domain page land there
	open := r.URL.Query().Get("domain")
	if open != "" && !a.CheckDomainBelongsToUser(ctx, userID, open) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	existing, domainsList := userRedirects(ctx, a, userID)
	renderRedirectsPage(a, w, r, domainsList, existing, open)
}

// userRedirects reads the redirect of every domain the user has, keyed by domain
func userRedirects(ctx context.Context, a *appctx.App, userID int) (map[string]redirectRule, []appctx.Domain) {
	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	out := map[string]redirectRule{}
	for _, d := range domainsList {
		if rule, ok := getRedirectRule(d.DomainURL); ok && rule.To != "" {
			out[d.DomainURL] = rule
		}
	}
	return out, domainsList
}

// redirectTargets keeps just the target of each rule, what the loop check needs
func redirectTargets(rules map[string]redirectRule) map[string]string {
	out := make(map[string]string, len(rules))
	for d, rule := range rules {
		out[d] = rule.To
	}
	return out
}

// redirectRule is a domain's redirect as the panel writes it: `redir [@op_redir_bare] <to>[{uri}] [301]`, with `@op_redir_bare host <domain>` on the line above when www keeps serving the site
type redirectRule struct {
	To        string `json:"to"`
	Permanent bool   `json:"permanent"` // 301 instead of caddy's default 302
	KeepPath  bool   `json:"keep_path"` // /some/page goes to <to>/some/page
	BareOnly  bool   `json:"bare_only"` // www keeps serving the site
}

const bareMatcher = "@op_redir_bare"

// lines are the conf lines for rule, the matcher first when there is one
func (rule redirectRule) lines(domain string) []string {
	target := rule.To
	if rule.KeepPath {
		target = strings.TrimRight(target, "/") + "{uri}"
	}
	var out []string
	redir := "redir "
	if rule.BareOnly {
		out = append(out, bareMatcher+" host "+domain)
		redir += bareMatcher + " "
	}
	redir += target
	if rule.Permanent {
		redir += " 301"
	}
	return append(out, redir)
}

// describe is the rule's options for the activity log
func (rule redirectRule) describe() string {
	parts := []string{"302"}
	if rule.Permanent {
		parts[0] = "301"
	}
	if rule.KeepPath {
		parts = append(parts, "path-keeping")
	}
	if rule.BareOnly {
		parts = append(parts, "bare-domain-only")
	}
	return strings.Join(parts, " ")
}

// parseRedirectLine reads a `redir` line back into a rule, ok is false for any other line
func parseRedirectLine(line string) (redirectRule, bool) {
	f := strings.Fields(line)
	if len(f) < 2 || f[0] != "redir" {
		return redirectRule{}, false
	}
	f = f[1:]
	var rule redirectRule
	if f[0] == bareMatcher {
		rule.BareOnly, f = true, f[1:]
		if len(f) == 0 {
			return redirectRule{}, false
		}
	}
	target := f[0]
	if len(f) > 1 {
		rule.Permanent = f[1] == "301" || f[1] == "permanent"
	}
	if t, ok := strings.CutSuffix(target, "{uri}"); ok {
		rule.KeepPath, target = true, t
	}
	rule.To = target
	return rule, true
}

func isBareMatcherLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), bareMatcher+" host ")
}

// bareHost lowercases a host and drops a www. prefix and a trailing dot, since a redirect covers the domain and its www
func bareHost(h string) string {
	h = strings.TrimSuffix(strings.ToLower(h), ".")
	return strings.TrimPrefix(h, "www.")
}

// servedHere reports whether a target URL lands on the ports Caddy serves domains on, a different port is another service
func servedHere(u *url.URL) bool {
	p := u.Port()
	return p == "" || p == "80" || p == "443"
}

// redirectTargetProblem returns why target can't be domain's redirect as an untranslated message with %(domain)s and %(other)s slots, or "" when it's fine
func redirectTargetProblem(domain, target string, existing map[string]string) (msg, other string) {
	if target == "" {
		return "Enter the URL visitors should go to.", ""
	}
	if !httpURLRE.MatchString(target) {
		return "Start with http:// or https:// and leave out spaces.", ""
	}
	// caddy expands {placeholders} in redir targets and quotes would break the conf
	if strings.ContainsAny(target, "{}\"`\\") {
		return "The URL can't contain { } \" ` or \\ characters.", ""
	}
	u, err := url.Parse(target)
	if err != nil || u.Hostname() == "" {
		return "That doesn't look like a full URL.", ""
	}
	self := bareHost(domain)
	host := bareHost(u.Hostname())
	if host == self && servedHere(u) {
		return "This points back at %(domain)s, so visitors would loop forever.", ""
	}
	for from, to := range existing {
		if bareHost(from) != host || host == self {
			continue
		}
		back, err := url.Parse(to)
		if err == nil && bareHost(back.Hostname()) == self && servedHere(back) && servedHere(u) {
			return "%(other)s already redirects to %(domain)s. Saving this would send visitors in a circle.", from
		}
	}
	return "", ""
}

// insertOrReplaceRedirect finds the `log {`/`import domain_log` anchor line(s) in a vhost config and either replaces the redirect lines right before it or inserts new ones, supporting both the legacy inline block and the newer import style (see openpanel/openpanel#645)
func insertOrReplaceRedirect(content string, rule redirectRule, domain string) string {
	lines := strings.Split(content, "\n")
	ruleLines := rule.lines(domain)

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
		// what the panel wrote before here: a redir line, maybe with its matcher above
		start := idx
		if start > 0 {
			if _, ok := parseRedirectLine(lines[start-1]); ok {
				start--
				if start > 0 && isBareMatcherLine(lines[start-1]) {
					start--
				}
			}
		}
		insert := make([]string, len(ruleLines))
		for k, l := range ruleLines {
			insert[k] = indent + l
		}
		before := append([]string{}, lines[:start]...)
		after := append([]string{}, lines[idx:]...)
		lines = append(before, append(insert, after...)...)
	}

	return strings.Join(lines, "\n")
}

// removeRedirect drops the redir lines to redirectURL and the matcher lines insertOrReplaceRedirect added, so add then delete leaves the conf as it was
func removeRedirect(content, redirectURL string) string {
	lines := strings.Split(content, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if rule, ok := parseRedirectLine(line); ok && rule.To == redirectURL {
			continue
		}
		if isBareMatcherLine(line) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func readTextFile(path string) (string, error) {
	content, err := tenantfs.ReadFile(path)
	return string(content), err
}

func writeTextFile(path, content string) error {
	return tenantfs.WriteFile(path, []byte(content), 0o644)
}
