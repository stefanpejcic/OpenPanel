package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Database is one database server's day, -1 means the number isn't known for this day
type Database struct {
	Engine          string `json:"engine"`
	Feature         string `json:"feature"`
	Link            string `json:"link"`
	Databases       int    `json:"databases"`
	Size            int64  `json:"size"`
	PrevSize        int64  `json:"prev_size"`
	Connections     int64  `json:"connections"`
	PeakConnections int64  `json:"peak_connections"`
	MaxConnections  int64  `json:"max_connections"`
	Queries         int64  `json:"queries"`
	SlowQueries     int64  `json:"slow_queries"`
	SlowSeconds     string `json:"slow_seconds,omitempty"`
	FailedLogins    int64  `json:"failed_logins"`
	Refused         int64  `json:"refused"`
	Deadlocks       int64  `json:"deadlocks"`
	Errors          int    `json:"errors"`
	LastError       string `json:"last_error,omitempty"`
	// raw counters and uptime when the summary was made, tomorrow's summary subtracts them to get its own day
	Uptime   int64            `json:"uptime"`
	Counters map[string]int64 `json:"counters,omitempty"`
}

// Sized is the size for templates, with the day's growth when there's a day before to compare with
func (d Database) Sized() string {
	out := humanBytes(d.Size)
	if d.PrevSize >= 0 && d.Size != d.PrevSize {
		diff := d.Size - d.PrevSize
		if diff > 0 {
			out += " (+" + humanBytes(diff) + ")"
		} else {
			out += " (-" + humanBytes(-diff) + ")"
		}
	}
	return out
}

func unknownDatabase(engine, feature, link string) Database {
	return Database{Engine: engine, Feature: feature, Link: link, PrevSize: -1, Connections: -1, PeakConnections: -1, Queries: -1, SlowQueries: -1, FailedLogins: -1, Refused: -1, Deadlocks: -1}
}

// counterDiff gives the day's part of an ever-growing counter, it's the whole value when the server started within the day, -1 when it can't tell
type counterDiff struct {
	now, prev   map[string]int64
	sameServer  bool
	sinceStart  bool
	hasPrevious bool
}

func newCounterDiff(now map[string]int64, uptime int64, prev *Database, prevAt, at time.Time, dayStart time.Time) counterDiff {
	d := counterDiff{now: now, sinceStart: uptime > 0 && at.Add(-time.Duration(uptime)*time.Second).After(dayStart)}
	if prev != nil && prev.Counters != nil && !prevAt.IsZero() {
		d.hasPrevious, d.prev = true, prev.Counters
		// a few minutes of slack, uptime and the generation time aren't read at the same instant
		d.sameServer = uptime+300 >= prev.Uptime+int64(at.Sub(prevAt).Seconds())
	}
	return d
}

func (d counterDiff) get(key string) int64 {
	v, ok := d.now[key]
	if !ok {
		return -1
	}
	switch {
	case d.hasPrevious && d.sameServer:
		if p, ok := d.prev[key]; ok && v >= p {
			return v - p
		}
		return -1
	case d.hasPrevious || d.sinceStart:
		// restarted since the last summary, so the counter only covers the time since then
		return v
	}
	return -1
}

// grew reports whether a high-water mark like Max_used_connections went up during the day
func (d counterDiff) grew(key string) bool {
	if d.sinceStart || (d.hasPrevious && !d.sameServer) {
		return true
	}
	return d.hasPrevious && d.now[key] > d.prev[key]
}

// trimLogLine keeps a log message short enough for one line of the summary
func trimLogLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 200 {
		s = s[:197] + "..."
	}
	return s
}

// parseMySQLLog counts [ERROR] lines, MySQL 8 and MariaDB both tag them that way
func parseMySQLLog(out string) (errors int, last string) {
	for _, line := range strings.Split(out, "\n") {
		i := strings.Index(line, "[ERROR]")
		if i < 0 {
			continue
		}
		errors++
		msg := strings.TrimSpace(line[i+len("[ERROR]"):])
		// MySQL 8 adds "[MY-010000] [Server]" in front of the message
		for strings.HasPrefix(msg, "[") {
			end := strings.Index(msg, "]")
			if end < 0 {
				break
			}
			msg = strings.TrimSpace(msg[end+1:])
		}
		last = trimLogLine(msg)
	}
	return errors, last
}

// parsePostgresLog counts ERROR, FATAL and PANIC lines, failed password logins and connections refused at the limit
func parsePostgresLog(out string) (errors int, failedLogins, refused int64, last string) {
	for _, line := range strings.Split(out, "\n") {
		level := ""
		for _, l := range []string{"ERROR:", "FATAL:", "PANIC:"} {
			if strings.Contains(line, l) {
				level = l
				break
			}
		}
		if level == "" {
			continue
		}
		msg := strings.TrimSpace(line[strings.Index(line, level)+len(level):])
		switch {
		case strings.Contains(msg, "password authentication failed"), strings.Contains(msg, "no pg_hba.conf entry"):
			failedLogins++
			continue
		case strings.Contains(msg, "too many clients"), strings.Contains(msg, "remaining connection slots are reserved"):
			refused++
			continue
		}
		errors++
		last = trimLogLine(msg)
	}
	return errors, failedLogins, refused, last
}

// parseMongoLog reads mongod's JSON log, s is the severity and E or F an error
func parseMongoLog(out string) (errors int, failedLogins, refused int64, last string) {
	for _, line := range strings.Split(out, "\n") {
		start := strings.Index(line, "{")
		if start < 0 {
			continue
		}
		var entry struct {
			S    string         `json:"s"`
			Msg  string         `json:"msg"`
			Attr map[string]any `json:"attr"`
		}
		if json.Unmarshal([]byte(line[start:]), &entry) != nil {
			continue
		}
		switch {
		case entry.Msg == "Authentication failed":
			failedLogins++
		case strings.Contains(entry.Msg, "too many open connections"):
			refused++
		case entry.S == "E" || entry.S == "F":
			errors++
			msg := entry.Msg
			if e, ok := entry.Attr["error"]; ok {
				msg += ": " + fmt.Sprint(e)
			}
			last = trimLogLine(msg)
		}
	}
	return errors, failedLogins, refused, last
}

// liveStats is true when the summary is made right after its day, older days can't read the servers' counters anymore
func (b *builder) liveStats() bool { return time.Since(b.to) < 36*time.Hour }

// previousDatabases is the day before's database snapshots by engine feature
func (b *builder) previousDatabases() (map[string]*Database, time.Time) {
	prev, ok := Load(b.username, b.from.AddDate(0, 0, -1).Format("2006-01-02"))
	out := map[string]*Database{}
	if !ok {
		return out, time.Time{}
	}
	for i := range prev.Databases {
		out[prev.Databases[i].Feature] = &prev.Databases[i]
	}
	return out, prev.GeneratedAt
}

// collectDatabases reads each database server the account can use and has running, it never starts one
func (b *builder) collectDatabases(ctx context.Context) {
	prev, prevAt := b.previousDatabases()
	if b.can("mysql") {
		if d, ok := b.mysqlStats(ctx, prev["mysql"], prevAt); ok {
			b.s.Databases = append(b.s.Databases, d)
		}
	}
	if b.can("postgresql") {
		if d, ok := b.postgresStats(ctx, prev["postgresql"], prevAt); ok {
			b.s.Databases = append(b.s.Databases, d)
		}
	}
	if b.can("mongodb") {
		if d, ok := b.mongoStats(ctx, prev["mongodb"], prevAt); ok {
			b.s.Databases = append(b.s.Databases, d)
		}
	}
	for i := range b.s.Databases {
		if p := prev[b.s.Databases[i].Feature]; p != nil && b.s.Databases[i].Counters != nil {
			b.s.Databases[i].PrevSize = p.Size
		}
	}
	b.checkDatabases()
}

func (b *builder) mysqlStats(ctx context.Context, prev *Database, prevAt time.Time) (Database, bool) {
	service := "mysql"
	if strings.EqualFold(strings.Trim(webserver.GetEnvFileValue(b.userContext, "MYSQL_TYPE"), `"' `), "mariadb") {
		service = "mariadb"
	}
	d := unknownDatabase("MySQL", "mysql", "/mysql")
	if service == "mariadb" {
		d.Engine = "MariaDB"
	}
	d.Errors, d.LastError = parseMySQLLog(b.podman(ctx, append(append([]string{"logs"}, b.window()...), service)...))
	if !mysqlmanager.SocketExists(b.userContext) {
		// stopped all day and quiet in the log, nothing to say about it
		return d, d.Errors > 0
	}
	if !b.liveStats() {
		return d, true
	}
	qctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	rows, err := mysqlmanager.Exec(qctx, b.userContext, "SHOW GLOBAL STATUS WHERE Variable_name IN ('Uptime','Connections','Max_used_connections','Questions','Slow_queries','Aborted_connects','Connection_errors_max_connections')", "")
	if err != nil {
		return d, true
	}
	counters := map[string]int64{}
	for _, r := range rows {
		if len(r) >= 2 {
			n, _ := strconv.ParseInt(mysqlmanager.ToString(r[1]), 10, 64)
			counters[mysqlmanager.ToString(r[0])] = n
		}
	}
	d.Uptime = counters["Uptime"]
	delete(counters, "Uptime")
	d.Counters = counters
	if res, err := mysqlmanager.Exec(qctx, b.userContext, "SELECT @@max_connections, @@long_query_time", ""); err == nil && len(res) > 0 && len(res[0]) >= 2 {
		d.MaxConnections = int64(mysqlmanager.ToInt(res[0][0]))
		if f, err := strconv.ParseFloat(mysqlmanager.ToString(res[0][1]), 64); err == nil {
			d.SlowSeconds = strconv.FormatFloat(f, 'f', -1, 64)
		}
	}
	system := strings.Fields(strings.Trim(b.a.Config.Get("mysql_restricted_databases", "information_schema performance_schema mysql phpmyadmin sys mariadb.sys"), `"'`))
	quoted := make([]string, len(system))
	for i, s := range system {
		quoted[i] = "'" + strings.ReplaceAll(strings.Trim(s, `"'`), "'", "") + "'"
	}
	q := "SELECT COUNT(DISTINCT s.schema_name), COALESCE(SUM(t.data_length + t.index_length), 0) FROM information_schema.schemata s " +
		"LEFT JOIN information_schema.tables t ON t.table_schema = s.schema_name WHERE s.schema_name NOT IN (" + strings.Join(quoted, ", ") + ")"
	if res, err := mysqlmanager.Exec(qctx, b.userContext, q, ""); err == nil && len(res) > 0 && len(res[0]) >= 2 {
		d.Databases = mysqlmanager.ToInt(res[0][0])
		d.Size, _ = strconv.ParseInt(mysqlmanager.ToString(res[0][1]), 10, 64)
	}

	diff := newCounterDiff(counters, d.Uptime, prev, prevAt, time.Now(), b.from)
	d.Connections = diff.get("Connections")
	d.Queries = diff.get("Questions")
	d.SlowQueries = diff.get("Slow_queries")
	d.FailedLogins = diff.get("Aborted_connects")
	d.Refused = diff.get("Connection_errors_max_connections")
	if diff.grew("Max_used_connections") {
		d.PeakConnections = counters["Max_used_connections"]
	}
	return d, true
}

func (b *builder) postgresStats(ctx context.Context, prev *Database, prevAt time.Time) (Database, bool) {
	d := unknownDatabase("PostgreSQL", "postgresql", "/postgresql")
	errs, failed, refused, last := parsePostgresLog(b.podman(ctx, append(append([]string{"logs"}, b.window()...), "postgres")...))
	d.Errors, d.FailedLogins, d.Refused, d.LastError = errs, failed, refused, last
	if docker.GetContainerStatus(ctx, b.userContext, "postgres").State != "running" {
		return d, errs > 0 || failed > 0 || refused > 0
	}
	if !b.liveStats() {
		return d, true
	}
	qctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	res, err := postgresmanager.Exec(qctx, b.userContext, "SELECT extract(epoch FROM now() - pg_postmaster_start_time())::bigint, "+
		"COALESCE(sum(xact_commit + xact_rollback), 0)::bigint, COALESCE(sum(deadlocks), 0)::bigint, "+
		"(SELECT setting::bigint FROM pg_settings WHERE name = 'max_connections') FROM pg_stat_database", "postgres")
	if err != nil || len(res) == 0 || len(res[0]) < 4 {
		return d, true
	}
	toInt := func(v any) int64 { n, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64); return n }
	d.Uptime = toInt(res[0][0])
	d.Counters = map[string]int64{"transactions": toInt(res[0][1]), "deadlocks": toInt(res[0][2])}
	d.MaxConnections = toInt(res[0][3])
	if sz, err := postgresmanager.Exec(qctx, b.userContext, "SELECT count(*), COALESCE(sum(pg_database_size(datname)), 0)::bigint FROM pg_database WHERE NOT datistemplate AND datname <> 'postgres'", "postgres"); err == nil && len(sz) > 0 && len(sz[0]) >= 2 {
		d.Databases = int(toInt(sz[0][0]))
		d.Size = toInt(sz[0][1])
	}
	diff := newCounterDiff(d.Counters, d.Uptime, prev, prevAt, time.Now(), b.from)
	d.Queries = diff.get("transactions")
	d.Deadlocks = diff.get("deadlocks")
	return d, true
}

func (b *builder) mongoStats(ctx context.Context, prev *Database, prevAt time.Time) (Database, bool) {
	d := unknownDatabase("MongoDB", "mongodb", "/mongodb")
	errs, failed, refused, last := parseMongoLog(b.podman(ctx, append(append([]string{"logs"}, b.window()...), "mongodb")...))
	d.Errors, d.FailedLogins, d.Refused, d.LastError = errs, failed, refused, last
	if docker.GetContainerStatus(ctx, b.userContext, "mongodb").State != "running" {
		return d, errs > 0 || failed > 0 || refused > 0
	}
	if !b.liveStats() {
		return d, true
	}
	qctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	client, err := mongomanager.GetClient(qctx, b.userContext)
	if err != nil {
		return d, true
	}
	var status struct {
		Uptime      float64 `bson:"uptime"`
		Connections struct {
			Current      int64 `bson:"current"`
			Available    int64 `bson:"available"`
			TotalCreated int64 `bson:"totalCreated"`
		} `bson:"connections"`
		Opcounters map[string]int64 `bson:"opcounters"`
	}
	if client.Database("admin").RunCommand(qctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&status) != nil {
		return d, true
	}
	var ops int64
	for _, n := range status.Opcounters {
		ops += n
	}
	d.Uptime = int64(status.Uptime)
	d.Counters = map[string]int64{"connections": status.Connections.TotalCreated, "operations": ops}
	d.MaxConnections = status.Connections.Current + status.Connections.Available
	if dbs, err := mongomanager.ListDatabases(qctx, b.userContext); err == nil {
		for _, db := range dbs {
			if !mongomanager.IsSystemDatabase(db.Name) {
				d.Databases++
				d.Size += db.SizeBytes
			}
		}
	}
	diff := newCounterDiff(d.Counters, d.Uptime, prev, prevAt, time.Now(), b.from)
	d.Connections = diff.get("connections")
	d.Queries = diff.get("operations")
	return d, true
}

// checkDatabases turns the day's database numbers into findings, refused connections first since visitors saw errors
func (b *builder) checkDatabases() {
	if len(b.s.Databases) == 0 {
		return
	}
	found := false
	for _, d := range b.s.Databases {
		conf, logs := b.databaseLinks(d)
		if d.Refused > 0 {
			found = true
			it := Item{Title: fmt.Sprintf("%s turned away %s because it was at its connection limit", d.Engine, plural(int(d.Refused), "connection", "connections")),
				Detail: "Websites that couldn't connect showed a database error."}
			if d.MaxConnections > 0 {
				it.Detail = fmt.Sprintf("The limit is %d connections. Websites that couldn't connect showed a database error.", d.MaxConnections)
			}
			if conf != "" {
				it.Action, it.Link = "Raise the connection limit", conf
			} else {
				it.Action = "Ask your hosting provider to raise the connection limit."
			}
			b.attention(it)
		} else if d.PeakConnections > 0 && d.MaxConnections > 0 && d.PeakConnections*100 >= d.MaxConnections*90 {
			found = true
			it := Item{Title: fmt.Sprintf("%s came close to its connection limit (%d of %d)", d.Engine, d.PeakConnections, d.MaxConnections),
				Detail: "Once it's reached, new connections are refused and websites show a database error."}
			if conf != "" {
				it.Action, it.Link = "Raise the connection limit", conf
			}
			b.look(it)
		}
		if d.Errors > 0 {
			found = true
			it := Item{Title: fmt.Sprintf("%s logged %s", d.Engine, plural(d.Errors, "error", "errors"))}
			if d.LastError != "" {
				it.Detail = "Last one: " + d.LastError
			}
			if logs != "" {
				it.Action, it.Link = "See the logs", logs
			}
			b.look(it)
		}
		// a few are typos, dozens mean a site has the wrong password or someone is guessing
		if d.FailedLogins >= 10 {
			found = true
			b.look(Item{Title: fmt.Sprintf("%s refused %s with a wrong password or from a host that isn't allowed", d.Engine, plural(int(d.FailedLogins), "login", "logins")),
				Detail: "Usually a website with an old database password in its config, or someone guessing passwords.",
				Action: "Check database users", Link: d.Link + "/users"})
		}
		if d.SlowQueries > 0 {
			found = true
			it := Item{Title: fmt.Sprintf("%s took longer than %s seconds", plural(int(d.SlowQueries), "database query", "database queries"), d.SlowSeconds),
				Detail: "Slow queries make pages wait, a missing index or a very large table is the usual cause."}
			if d.SlowSeconds == "" {
				it.Title = fmt.Sprintf("%s ran slower than the slow query limit", plural(int(d.SlowQueries), "database query", "database queries"))
			}
			if b.can("mysql_processlist") {
				it.Action, it.Link = "See running queries", "/mysql/processlist"
			}
			b.look(it)
		}
		if d.Deadlocks > 0 {
			found = true
			b.look(Item{Title: fmt.Sprintf("%s had %s", d.Engine, plural(int(d.Deadlocks), "deadlock", "deadlocks")),
				Detail: "Two queries waited on each other and one was cancelled, the app usually retries or shows an error.",
				Action: "See running queries", Link: d.Link + "/processlist"})
		}
	}
	if !found {
		b.good("Your databases ran without errors.")
	}
}

// databaseLinks are the configuration page and the logs for one server, empty when the user can't open them
func (b *builder) databaseLinks(d Database) (conf, logs string) {
	switch d.Feature {
	case "mysql":
		if b.can("mysql_conf") {
			conf = "/mysql/configuration"
		}
	case "postgresql":
		if b.can("postgresql_conf") {
			conf = "/postgresql/configuration"
		}
	}
	if b.can("docker") {
		container := map[string]string{"mysql": strings.ToLower(d.Engine), "postgresql": "postgres", "mongodb": "mongodb"}[d.Feature]
		logs = "/containers/logs?container=" + container + "&lines=100&ctx=service"
	}
	return conf, logs
}

// databaseLine is one server's day in a sentence for the email
func databaseLine(d Database) string {
	line := d.Engine
	if d.Counters != nil {
		line += ": " + plural(d.Databases, "database", "databases") + ", " + d.Sized()
	}
	var extra []string
	if d.Queries >= 0 {
		word := "queries"
		if d.Feature != "mysql" {
			word = map[string]string{"postgresql": "transactions", "mongodb": "operations"}[d.Feature]
		}
		extra = append(extra, num(int(d.Queries))+" "+word)
	}
	if d.Connections >= 0 {
		extra = append(extra, plural(int(d.Connections), "connection", "connections"))
	}
	if d.PeakConnections >= 0 && d.MaxConnections > 0 {
		extra = append(extra, fmt.Sprintf("peak %d of %d connections", d.PeakConnections, d.MaxConnections))
	}
	if d.Errors > 0 {
		extra = append(extra, plural(d.Errors, "error", "errors"))
	}
	if len(extra) > 0 {
		line += " (" + strings.Join(extra, ", ") + ")"
	}
	return line
}

// QueriesLabel names what Queries counts for this server, for the page
func (d Database) QueriesLabel() string {
	switch d.Feature {
	case "postgresql":
		return "Transactions"
	case "mongodb":
		return "Operations"
	}
	return "Queries"
}

// Num is num for templates, they hand over int64
func (d Database) Num(n int64) string { return num(int(n)) }
