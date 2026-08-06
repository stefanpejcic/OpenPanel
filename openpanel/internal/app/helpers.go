package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/net/idna"

	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

// This file holds the shared lookup helpers used widely across other
// modules: AllDomainsForUser, CheckDomainBelongsToUser,
// GetCachedIPForUserOrPublicIPv4, GetResourceUsage, GetLastLoginData,
// QueryPlanDetailsByID, QueryPlanEmailMailboxLimits,
// ReadDedicatedIPFromFile, and GetUID. Standalone JSON routes are not
// here; they belong to whichever later phase owns each one. The
// podman/docker-CLI dependent shared helpers (container compose/start/stop,
// log fetching, the undeletable-services list) live in
// internal/modules/docker instead, since this package can't import that
// one without an import cycle (docker already imports appctx).

// AllDomainsForUser queries the domains table for every domain belonging
// to userID. Uncached - each call hits the database directly.
func (a *App) AllDomainsForUser(ctx context.Context, userID int) ([]Domain, error) {
	rows, err := a.DB.QueryContext(ctx,
		"SELECT domain_id, docroot, domain_url, php_version FROM domains WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain
	for rows.Next() {
		var (
			d          Domain
			docroot    sql.NullString
			phpVersion sql.NullString
		)
		if err := rows.Scan(&d.DomainID, &docroot, &d.DomainURL, &phpVersion); err != nil {
			return nil, err
		}
		d.Docroot, d.PHPVersion = docroot.String, phpVersion.String
		result = append(result, d)
	}
	return result, rows.Err()
}

// CheckDomainBelongsToUser reports whether domainParam (a numeric
// domain_id or a domain name) belongs to userID. Domain names are
// IDNA-encoded before the lookup. Any error (DB or IDNA conversion) is
// treated as "doesn't belong to this user" rather than surfaced as an
// error.
func (a *App) CheckDomainBelongsToUser(ctx context.Context, userID int, domainParam string) bool {
	if idx := strings.Index(domainParam, "/"); idx != -1 {
		domainParam = domainParam[:idx]
	}

	isNumeric := domainParam != "" && strings.IndexFunc(domainParam, func(r rune) bool { return r < '0' || r > '9' }) == -1

	if !isNumeric {
		encoded, err := idna.ToASCII(domainParam)
		if err != nil {
			return false
		}
		domainParam = encoded
	}

	query := "SELECT user_id FROM domains WHERE domain_url = ?"
	if isNumeric {
		query = "SELECT user_id FROM domains WHERE domain_id = ?"
	}

	var gotUserID int
	if err := a.DB.QueryRowContext(ctx, query, domainParam).Scan(&gotUserID); err != nil {
		return false
	}
	return gotUserID == userID
}

// ReadDedicatedIPFromFile reads a user's dedicated-IP JSON file, if
// present. Returns ("", false) if the file doesn't exist or the "ip" key
// is absent/empty.
func ReadDedicatedIPFromFile(username string) (string, bool) {
	path := fmt.Sprintf("/etc/openpanel/openpanel/core/users/%s/ip.json", username)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var parsed struct {
		IP string `json:"ip"`
	}
	if json.Unmarshal(data, &parsed) != nil || parsed.IP == "" {
		return "", false
	}
	return parsed.IP, true
}

// GetCachedIPForUserOrPublicIPv4 returns a user's dedicated IP if set,
// else the server's public IPv4, cached 10 minutes. Used on dashboard,
// temporary links, ftp and emails.
func (a *App) GetCachedIPForUserOrPublicIPv4(ctx context.Context, username string) string {
	ip, _ := cache.Memoize(ctx, a.Cache, "get_cached_ip_for_user_or_public_ipv4:"+username, 600*time.Second, func() (string, error) {
		if ip, ok := ReadDedicatedIPFromFile(username); ok {
			return ip, nil
		}
		return sysinfo.FetchPublicIP(ctx, a.Cache), nil
	})
	return ip
}

// GetResourceUsage returns the last JSON line of
// /home/<context>/resource_usage.txt, cached 6 minutes.
func (a *App) GetResourceUsage(ctx context.Context, username, userContext string) (map[string]any, error) {
	return cache.Memoize(ctx, a.Cache, "get_resource_usage:"+username, 360*time.Second, func() (map[string]any, error) {
		path := fmt.Sprintf("/home/%s/resource_usage.txt", userContext)
		lastLine, err := lastLineOf(path)
		if err != nil {
			return nil, nil //nolint:nilerr // deliberate: a read/parse failure means "no usage data yet", not an error response
		}
		var stats map[string]any
		if err := json.Unmarshal([]byte(lastLine), &stats); err != nil {
			return nil, nil //nolint:nilerr
		}
		return stats, nil
	})
}

func lastLineOf(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	if last == "" {
		return "", fmt.Errorf("empty file")
	}
	return last, nil
}

// PlanDetails holds the full plans-table row shape returned by
// QueryPlanDetailsByID.
type PlanDetails struct {
	Description   string
	DomainsLimit  string
	WebsitesLimit string
	DBLimit       string
	CPU           string
	RAM           string
	EmailLimit    string
	FTPLimit      string
	DiskLimit     string
	InodesLimit   string
	Bandwidth     string
	MaxEmailQuota string
}

// QueryPlanDetailsByID returns the full plans-table row for planID, cached
// 5 minutes. Every column is read as nullable so a NULL column doesn't
// fail the whole scan - it just comes back as an empty string.
func (a *App) QueryPlanDetailsByID(ctx context.Context, planID int) (PlanDetails, error) {
	return cache.Memoize(ctx, a.Cache, fmt.Sprintf("query_plan_details_by_id:%d", planID), 300*time.Second, func() (PlanDetails, error) {
		var (
			description, domainsLimit, websitesLimit, dbLimit sql.NullString
			cpu, ram, emailLimit, ftpLimit, diskLimit         sql.NullString
			inodesLimit, bandwidth, maxEmailQuota             sql.NullString
		)
		row := a.DB.QueryRowContext(ctx, `
			SELECT description, domains_limit, websites_limit, db_limit,
			       cpu, ram, email_limit, ftp_limit, disk_limit,
			       inodes_limit, bandwidth, max_email_quota
			FROM plans WHERE id = ?`, planID)
		err := row.Scan(&description, &domainsLimit, &websitesLimit, &dbLimit,
			&cpu, &ram, &emailLimit, &ftpLimit, &diskLimit,
			&inodesLimit, &bandwidth, &maxEmailQuota)
		if err != nil {
			return PlanDetails{}, nil //nolint:nilerr // deliberate: a missing plan returns a zero-value result, not an error
		}
		return PlanDetails{
			Description: description.String, DomainsLimit: domainsLimit.String,
			WebsitesLimit: websitesLimit.String, DBLimit: dbLimit.String,
			CPU: cpu.String, RAM: ram.String,
			EmailLimit: emailLimit.String, FTPLimit: ftpLimit.String, DiskLimit: diskLimit.String,
			InodesLimit: inodesLimit.String, Bandwidth: bandwidth.String, MaxEmailQuota: maxEmailQuota.String,
		}, nil
	})
}

// QueryPlanEmailMailboxLimits returns the (email_limit, max_email_quota)
// pair for a plan, cached 6 minutes.
func (a *App) QueryPlanEmailMailboxLimits(ctx context.Context, planID int) (emailLimit, maxEmailQuota string, err error) {
	type limits struct{ EmailLimit, MaxEmailQuota string }
	result, err := cache.Memoize(ctx, a.Cache, fmt.Sprintf("query_plan_email_mailbox_limits:%d", planID), 360*time.Second, func() (limits, error) {
		var el, meq sql.NullString
		row := a.DB.QueryRowContext(ctx, "SELECT email_limit, max_email_quota FROM plans WHERE id = ?", planID)
		if scanErr := row.Scan(&el, &meq); scanErr != nil {
			return limits{}, nil //nolint:nilerr // deliberate: a missing plan returns empty limits, not an error
		}
		return limits{EmailLimit: el.String, MaxEmailQuota: meq.String}, nil
	})
	return result.EmailLimit, result.MaxEmailQuota, err
}

// LastLogin is one parsed entry returned by GetLastLoginData.
type LastLogin struct {
	IP          string
	CountryCode string
	LoginTime   string
}

// GetLastLoginData returns a user's parsed login history, cached 10
// minutes.
func (a *App) GetLastLoginData(ctx context.Context, username string) ([]LastLogin, error) {
	return cache.Memoize(ctx, a.Cache, "get_last_login_data:"+username, 600*time.Second, func() ([]LastLogin, error) {
		path := fmt.Sprintf("/etc/openpanel/openpanel/core/users/%s/.lastlogin", username)
		data, err := os.ReadFile(path)
		if err != nil {
			_ = os.WriteFile(path, nil, 0o644) // create the empty file if missing, so later reads don't need to special-case it
			return nil, nil
		}
		return parseLastLoginLines(strings.Split(string(data), "\n")), nil
	})
}

// parseLastLoginLines parses "IP: <ip> - Country: <cc> - Login Time: <ts>"
// entries, one per line. Split out as a pure function so it's testable
// without the filesystem/cache.
func parseLastLoginLines(lines []string) []LastLogin {
	var entries []LastLogin
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " - ")
		if len(parts) != 3 ||
			!strings.HasPrefix(parts[0], "IP: ") ||
			!strings.HasPrefix(parts[1], "Country: ") ||
			!strings.HasPrefix(parts[2], "Login Time: ") {
			continue
		}
		entries = append(entries, LastLogin{
			IP:          strings.TrimPrefix(parts[0], "IP: "),
			CountryCode: strings.TrimSpace(strings.TrimPrefix(parts[1], "Country: ")),
			LoginTime:   strings.TrimPrefix(parts[2], "Login Time: "),
		})
	}
	return entries
}

// GetUID returns the numeric UID that owns /home/<username>, cached 2h.
// podmanmanager.GetUID does the same os.Stat lookup uncached (it's called
// inline when building podman socket paths); this wraps it with a
// longer-lived cache for the separate (less latency-sensitive) callers
// that just need the UID for display/URL purposes.
func (a *App) GetUID(ctx context.Context, username string) (int, error) {
	return cache.Memoize(ctx, a.Cache, "get_uid:"+username, 2*time.Hour, func() (int, error) {
		uid, err := podmanmanager.GetUID(username)
		if err != nil {
			return 0, nil //nolint:nilerr // deliberate: a lookup failure returns 0, not an error
		}
		return uid, nil
	})
}
