package dashboard

import (
	"context"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
)

// DashboardList stands in for the websites card on plans without sites
type DashboardList struct {
	Kind, Title, Desc, Href, Icon, NameLabel, MetaLabel, Empty string
	HasState, HasType                                          bool
	Rows                                                       []DashboardListRow
}

// DashboardListRow is one line of that card, State is Running or Stopped for containers
type DashboardListRow struct {
	Name, Sub, Meta, Href, State string
	// TypeLabel and TypeIcon say which engine a database row is on
	TypeLabel string
	TypeIcon  template.HTML
	size      int64
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
		l := &DashboardList{Kind: "emails", Title: "Email accounts", Desc: "Your mailboxes, the fullest first.", Href: "/emails", Icon: "ri-mail-line", NameLabel: "Address", MetaLabel: "Used", Empty: "No email accounts yet"}
		l.Rows = emailRows(emails.GetEmailList(ctx, a, userID, username, set))
		return l
	case allowed["mysql"] || allowed["postgresql"] || allowed["mongodb"]:
		href := "/mysql"
		switch {
		case !allowed["mysql"] && allowed["postgresql"]:
			href = "/postgresql"
		case !allowed["mysql"] && !allowed["postgresql"]:
			href = "/mongodb"
		}
		l := &DashboardList{Kind: "databases", Title: "Databases", Desc: "Your databases on every engine, the biggest first.", Href: href, Icon: "ri-database-2-line", NameLabel: "Database", MetaLabel: "Size", HasType: true, Empty: "No databases yet"}
		l.Rows = databaseRows(ctx, a, username, userContext, allowed)
		return l
	case allowed["ftp"]:
		l := &DashboardList{Kind: "ftp", Title: "FTP accounts", Desc: "FTP users and their folders, the most used first.", Href: "/ftp", Icon: "ri-folder-shared-line", NameLabel: "User", MetaLabel: "Used", Empty: "No FTP accounts yet"}
		l.Rows = ftpRows(userContext)
		return l
	case allowed["docker"]:
		l := &DashboardList{Kind: "containers", Title: "Containers", Desc: "The services on your account, running ones first.", Href: "/containers", Icon: "ri-stack-line", NameLabel: "Service", HasState: true, Empty: "No containers yet"}
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

// databaseRows merges MySQL/MariaDB, PostgreSQL and MongoDB databases, biggest first, cached so the dashboard doesn't wait on them
func databaseRows(ctx context.Context, a *appctx.App, username, userContext string, allowed map[string]bool) []DashboardListRow {
	if a.Cache == nil {
		return nil
	}
	key := fmt.Sprintf("dashboard_db_rows:%s:%t%t%t", username, allowed["mysql"], allowed["postgresql"], allowed["mongodb"])
	rows, _ := cache.Memoize(ctx, a.Cache, key, 10*time.Minute, func() ([]DashboardListRow, error) {
		var out []DashboardListRow
		if allowed["mysql"] {
			out = append(out, mysqlRows(ctx, a, userContext)...)
		}
		if allowed["postgresql"] {
			out = append(out, postgresRows(ctx, userContext)...)
		}
		if allowed["mongodb"] {
			out = append(out, mongoRows(ctx, userContext)...)
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].size > out[j].size })
		return firstN(out, altListRows), nil
	})
	return rows
}

func dbRow(name, kind, href string, size int64) DashboardListRow {
	return DashboardListRow{Name: name, Meta: humanSize(size), Href: href, TypeLabel: dbTypeLabels[kind], TypeIcon: dbTypeIcons[kind], size: size}
}

// mysqlRows only asks when the socket is there, a stopped MySQL isn't woken up for the dashboard
func mysqlRows(ctx context.Context, a *appctx.App, userContext string) []DashboardListRow {
	if !mysqlmanager.SocketExists(userContext) {
		return nil
	}
	kind := "mysql"
	if strings.EqualFold(strings.Trim(webserver.GetEnvFileValue(userContext, "MYSQL_TYPE"), `"' `), "mariadb") {
		kind = "mariadb"
	}
	q := "SELECT s.schema_name, COALESCE(SUM(t.data_length + t.index_length), 0) FROM information_schema.schemata s " +
		"LEFT JOIN information_schema.tables t ON t.table_schema = s.schema_name WHERE s.schema_name NOT IN (" +
		restrictedDatabasesSQL(a.Config) + ") GROUP BY s.schema_name"
	res, err := mysqlmanager.Exec(ctx, userContext, q, "")
	if err != nil {
		return nil
	}
	var out []DashboardListRow
	for _, r := range res {
		if len(r) >= 2 {
			out = append(out, dbRow(mysqlmanager.ToString(r[0]), kind, "/mysql", int64(mysqlmanager.ToInt(r[1]))))
		}
	}
	return out
}

// postgresRows reads only a running container, the dashboard never starts PostgreSQL
func postgresRows(ctx context.Context, userContext string) []DashboardListRow {
	if docker.GetContainerStatus(ctx, userContext, "postgres").State != "running" {
		return nil
	}
	res, err := postgresmanager.Exec(ctx, userContext, "SELECT datname, pg_database_size(datname) FROM pg_database WHERE datname NOT IN ('postgres','template0','template1') AND NOT datistemplate", "postgres")
	if err != nil {
		return nil
	}
	var out []DashboardListRow
	for _, r := range res {
		if len(r) >= 2 {
			n, _ := strconv.ParseInt(fmt.Sprint(r[1]), 10, 64)
			out = append(out, dbRow(fmt.Sprint(r[0]), "postgresql", "/postgresql", n))
		}
	}
	return out
}

// mongoRows reads only a running container, system databases left out
func mongoRows(ctx context.Context, userContext string) []DashboardListRow {
	if docker.GetContainerStatus(ctx, userContext, "mongodb").State != "running" {
		return nil
	}
	dbs, err := mongomanager.ListDatabases(ctx, userContext)
	if err != nil {
		return nil
	}
	var out []DashboardListRow
	for _, d := range dbs {
		if !mongomanager.IsSystemDatabase(d.Name) {
			out = append(out, dbRow(d.Name, "mongodb", "/mongodb", d.SizeBytes))
		}
	}
	return out
}

// ftpRows reads users.list (user|password|path|uid|gid) and sorts by quota.usage next to it, the most used first
func ftpRows(userContext string) []DashboardListRow {
	dir := fmt.Sprintf("/etc/openpanel/ftp/users/%s/", userContext)
	data, err := tenantfs.ReadFile(dir + "users.list")
	if err != nil {
		return nil
	}
	used := map[string]int64{}
	if q, qerr := tenantfs.ReadFile(dir + "quota.usage"); qerr == nil {
		for _, line := range strings.Split(string(q), "\n") {
			if p := strings.Split(strings.TrimSpace(line), "|"); len(p) >= 2 && p[0] != "" {
				kb, _ := strconv.ParseInt(p[1], 10, 64)
				used[p[0]] = kb
			}
		}
	}
	var out []DashboardListRow
	var sizes []int64
	for _, line := range strings.Split(string(data), "\n") {
		p := strings.Split(strings.TrimSpace(line), "|")
		if len(p) < 3 || p[0] == "" {
			continue
		}
		row := DashboardListRow{Name: p[0], Sub: p[2], Href: "/ftp"}
		if kb, ok := used[p[0]]; ok {
			row.Meta = humanSize(kb * 1024)
		}
		out = append(out, row)
		sizes = append(sizes, used[p[0]])
	}
	return firstN(sortBySize(out, sizes), altListRows)
}

// emailRows turns "* info@demo.rs ( 28M / 2.0G ) [1%]" lines into rows, the fullest mailbox first
func emailRows(lines []string) []DashboardListRow {
	var out []DashboardListRow
	var sizes []int64
	for _, line := range lines {
		addr, usedStr, _ := emails.ParseMailboxQuota(line)
		if addr == "" {
			continue
		}
		dom := addr
		if i := strings.LastIndex(addr, "@"); i >= 0 {
			dom = addr[i+1:]
		}
		out = append(out, DashboardListRow{Name: addr, Sub: dom, Meta: usedStr, Href: "/emails"})
		sizes = append(sizes, parseHumanBytes(usedStr))
	}
	return firstN(sortBySize(out, sizes), altListRows)
}

// sortBySize orders rows by their size, biggest first, and keeps the file order for ties
func sortBySize(rows []DashboardListRow, sizes []int64) []DashboardListRow {
	idx := make([]int, len(rows))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return sizes[idx[a]] > sizes[idx[b]] })
	out := make([]DashboardListRow, len(rows))
	for i, j := range idx {
		out[i] = rows[j]
	}
	return out
}

// parseHumanBytes reads mailserver sizes like 7.0K, 28M or 2.0G, anything else is 0
func parseHumanBytes(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "~" {
		return 0
	}
	mult := map[byte]float64{'K': 1 << 10, 'M': 1 << 20, 'G': 1 << 30, 'T': 1 << 40}
	m := 1.0
	if f, ok := mult[s[len(s)-1]]; ok {
		m, s = f, s[:len(s)-1]
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(n * m)
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
	}
	sort.SliceStable(rows, func(a, b int) bool { return rows[a].State == "Running" && rows[b].State != "Running" })
	return firstN(rows, altListRows)
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
