package domains

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// DomainWithSite is one row of a user's domains joined with their site count.
type DomainWithSite struct {
	DomainID   int
	Docroot    string
	DomainURL  string
	PHPVersion string
	SiteCount  int
}

// domainsWithSites lists a user's domains along with their site count.
func domainsWithSites(ctx context.Context, a *appctx.App, userID int) ([]DomainWithSite, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT d.domain_id, d.docroot, d.domain_url, d.php_version, COUNT(s.id) AS site_count
		FROM domains d
		LEFT JOIN sites s ON d.domain_id = s.domain_id
		WHERE d.user_id = ?
		GROUP BY d.domain_id, d.docroot, d.domain_url, d.php_version
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DomainWithSite
	for rows.Next() {
		var d DomainWithSite
		if err := rows.Scan(&d.DomainID, &d.Docroot, &d.DomainURL, &d.PHPVersion, &d.SiteCount); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// getRedirectURL returns the first "redir <url>" line (not a comment)
// inside a domain's Caddy config, if any.
func getRedirectURL(domainURL string) string {
	f, err := os.Open("/etc/openpanel/caddy/domains/" + domainURL + ".conf")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "redir ") {
			continue
		}
		// Split the whole line (not just the part after "redir ") into at
		// most 3 whitespace fields and take the second - matches even a
		// naive "redir " match inside a comment line.
		parts := splitMax(line, 3)
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// SSLStatus is a domain's SSL/suspend state as shown on the domains list.
type SSLStatus struct {
	HTTPS          string // "Unknown" | "Automatic" | "Custom"
	Suspended      string // "Not Suspended" | "Suspended"
	SuspendComment string
}

// isRewriteCondEnabled reads a domain's SSL/suspend status from its Caddy
// config, cached 30s.
func isRewriteCondEnabled(ctx context.Context, a *appctx.App, domainURL string) SSLStatus {
	status, _ := cache.Memoize(ctx, a.Cache, "is_rewrite_cond_enabled:"+domainURL, 30*time.Second, func() (SSLStatus, error) {
		return computeRewriteCondEnabled(domainURL), nil
	})
	return status
}

func computeRewriteCondEnabled(domainURL string) SSLStatus {
	status := SSLStatus{HTTPS: "Unknown", Suspended: "Not Suspended"}

	f, err := os.Open("/etc/openpanel/caddy/domains/" + domainURL + ".conf")
	if err != nil {
		return status
	}
	defer f.Close()

	onDemandFound := false
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if lineNum == 0 && strings.HasPrefix(line, "# comment:") {
			status.SuspendComment = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
		if strings.Contains(line, "on_demand") {
			onDemandFound = true
		}
		if strings.HasPrefix(line, "reverse_proxy") {
			status.Suspended = "Not Suspended"
		} else if strings.HasPrefix(line, "file_server") {
			status.Suspended = "Suspended"
		}
		lineNum++
	}

	if onDemandFound {
		status.HTTPS = "Automatic"
	} else {
		status.HTTPS = "Custom"
	}
	return status
}

// invalidateRewriteCondCache busts isRewriteCondEnabled's cache entry for one domain.
func invalidateRewriteCondCache(ctx context.Context, a *appctx.App, domainURL string) {
	_ = a.Cache.Delete(ctx, "is_rewrite_cond_enabled:"+domainURL)
}

// splitMax splits on runs of whitespace, stopping after maxFields-1 splits
// so the final field keeps any remaining whitespace-separated content intact.
func splitMax(s string, maxFields int) []string {
	var fields []string
	rest := s
	for len(fields) < maxFields-1 {
		rest = strings.TrimLeft(rest, " \t")
		if rest == "" {
			return fields
		}
		idx := strings.IndexAny(rest, " \t")
		if idx == -1 {
			fields = append(fields, rest)
			return fields
		}
		fields = append(fields, rest[:idx])
		rest = rest[idx:]
	}
	rest = strings.TrimLeft(rest, " \t")
	if rest != "" {
		fields = append(fields, rest)
	}
	return fields
}

func injected(a *appctx.App, ctx context.Context, userID int) (username, userContext string, err error) {
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}
