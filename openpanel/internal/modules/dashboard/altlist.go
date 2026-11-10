package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
)

// DashboardList stands in for the websites card on plans without sites
type DashboardList struct {
	Kind, Title, Desc, Href, Icon, NameLabel, MetaLabel, Empty string
	HasState                                                   bool
	Rows                                                       []DashboardListRow
}

// DashboardListRow is one line of that card, State is Running or Stopped for containers
type DashboardListRow struct {
	Name, Sub, Meta, Href, State string
}

const altListRows = 6

// sitesAllowed matches the Websites tile in PlanTiles
func sitesAllowed(allowed map[string]bool) bool {
	return allowed["wordpress"] || allowed["website_builder"] || allowed["nodejs"] || allowed["python"]
}

// dashboardAltList picks emails, then databases, then FTP, then containers, whichever the plan has first
func dashboardAltList(ctx context.Context, a *appctx.App, userID int, username, userContext string, allowed map[string]bool, domains []appctx.Domain) *DashboardList {
	switch {
	case allowed["emails"]:
		set := make(map[string]bool, len(domains))
		for _, d := range domains {
			set[d.DomainURL] = true
		}
		l := &DashboardList{Kind: "emails", Title: "Email accounts", Desc: "Your mailboxes and the domain each one belongs to.", Href: "/emails", Icon: "ri-mail-line", NameLabel: "Address", Empty: "No email accounts yet"}
		for _, addr := range firstN(emails.GetEmailList(ctx, a, userID, username, set), altListRows) {
			dom := addr
			if i := strings.LastIndex(addr, "@"); i >= 0 {
				dom = addr[i+1:]
			}
			l.Rows = append(l.Rows, DashboardListRow{Name: addr, Sub: dom, Href: "/emails"})
		}
		return l
	case allowed["mysql"] || allowed["postgresql"]:
		href := "/postgresql"
		if allowed["mysql"] {
			href = "/mysql"
		}
		l := &DashboardList{Kind: "databases", Title: "Databases", Desc: "Your databases and how much space each one takes.", Href: href, Icon: "ri-database-2-line", NameLabel: "Database", MetaLabel: "Size", Empty: "No databases yet"}
		if allowed["mysql"] {
			l.Rows = mysqlRows(ctx, a, username, userContext)
		}
		return l
	case allowed["ftp"]:
		l := &DashboardList{Kind: "ftp", Title: "FTP accounts", Desc: "FTP users and the folder each one opens in.", Href: "/ftp", Icon: "ri-folder-shared-line", NameLabel: "User", Empty: "No FTP accounts yet"}
		l.Rows = ftpRows(userContext)
		return l
	case allowed["docker"]:
		l := &DashboardList{Kind: "containers", Title: "Containers", Desc: "The services on your account and whether they run.", Href: "/containers", Icon: "ri-stack-line", NameLabel: "Service", HasState: true, Empty: "No containers yet"}
		l.Rows = containerRows(ctx, a, userContext)
		return l
	}
	return nil
}

func firstN[T any](s []T, n int) []T {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// mysqlRows lists databases with their size, cached like the count so the dashboard doesn't wait on MySQL
func mysqlRows(ctx context.Context, a *appctx.App, username, userContext string) []DashboardListRow {
	if !mysqlmanager.SocketExists(userContext) || a.Cache == nil {
		return nil
	}
	rows, _ := cache.Memoize(ctx, a.Cache, "dashboard_db_rows:"+username, 10*time.Minute, func() ([]DashboardListRow, error) {
		q := "SELECT s.schema_name, COALESCE(SUM(t.data_length + t.index_length), 0) FROM information_schema.schemata s " +
			"LEFT JOIN information_schema.tables t ON t.table_schema = s.schema_name WHERE s.schema_name NOT IN (" +
			restrictedDatabasesSQL(a.Config) + ") GROUP BY s.schema_name ORDER BY s.schema_name LIMIT " + fmt.Sprint(altListRows)
		res, err := mysqlmanager.Exec(ctx, userContext, q, "")
		if err != nil {
			return nil, nil //nolint:nilerr // an empty card is fine when MySQL doesn't answer
		}
		var out []DashboardListRow
		for _, r := range res {
			if len(r) < 2 {
				continue
			}
			out = append(out, DashboardListRow{Name: fmt.Sprint(r[0]), Meta: humanSize(int64(mysqlmanager.ToInt(r[1]))), Href: "/mysql"})
		}
		return out, nil
	})
	return rows
}

// ftpRows reads users.list, user|password|path|uid|gid
func ftpRows(userContext string) []DashboardListRow {
	data, err := tenantfs.ReadFile(fmt.Sprintf("/etc/openpanel/ftp/users/%s/users.list", userContext))
	if err != nil {
		return nil
	}
	var out []DashboardListRow
	for _, line := range strings.Split(string(data), "\n") {
		p := strings.Split(strings.TrimSpace(line), "|")
		if len(p) < 3 || p[0] == "" {
			continue
		}
		out = append(out, DashboardListRow{Name: p[0], Sub: p[2], Href: "/ftp"})
		if len(out) == altListRows {
			break
		}
	}
	return out
}

// containerRows is one podman ps for the account, cached 30s like the site states
func containerRows(ctx context.Context, a *appctx.App, userContext string) []DashboardListRow {
	if userContext == "" || a.Cache == nil {
		return nil
	}
	rows, _ := cache.Memoize(ctx, a.Cache, "dashboard_container_rows:"+userContext, 30*time.Second, func() ([]DashboardListRow, error) {
		pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		argv := podmanmanager.PodmanArgv(userContext, "ps", "-a", "--format", "{{.Names}}|{{.State}}|{{.Image}}")
		out, err := podmanmanager.Command(pctx, userContext, argv).Output()
		if err != nil {
			return nil, nil //nolint:nilerr // no list is shown as the empty card
		}
		return parseContainerRows(string(out)), nil
	})
	return rows
}

func parseContainerRows(out string) []DashboardListRow {
	var rows []DashboardListRow
	for _, line := range strings.Split(out, "\n") {
		p := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(p) < 3 || p[0] == "" {
			continue
		}
		state := "Stopped"
		if strings.EqualFold(p[1], "running") {
			state = "Running"
		}
		rows = append(rows, DashboardListRow{Name: p[0], Sub: p[2], State: state, Href: "/containers"})
		if len(rows) == altListRows {
			break
		}
	}
	return rows
}

func humanSize(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	f, i := float64(b)/1024, 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	if f < 10 {
		return strings.TrimSuffix(fmt.Sprintf("%.1f", f), ".0") + " " + units[i]
	}
	return fmt.Sprintf("%.0f %s", f, units[i])
}
