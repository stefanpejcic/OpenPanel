package summary

import (
	"strings"
	"testing"
	"time"
)

func TestParseMySQLLog(t *testing.T) {
	out := "2026-11-12 10:00:00 0 [Note] mariadbd: ready for connections.\n" +
		"2026-11-12 11:00:00 5 [ERROR] Incorrect definition of table mysql.column_stats\n" +
		"2026-11-12T12:00:00.000000Z 0 [ERROR] [MY-010119] [Server] Aborting\n" +
		"2026-11-12 13:00:00 7 [Warning] Access denied for user 'x'@'localhost'\n"
	n, last := parseMySQLLog(out)
	if n != 2 || last != "Aborting" {
		t.Fatalf("got %d %q", n, last)
	}
}

func TestParsePostgresLog(t *testing.T) {
	out := `2026-11-12 10:00:00.000 UTC [12] FATAL:  password authentication failed for user "app"
2026-11-12 10:00:01.000 UTC [13] FATAL:  sorry, too many clients already
2026-11-12 10:00:02.000 UTC [14] ERROR:  relation "users" does not exist at character 15
2026-11-12 10:00:03.000 UTC [14] STATEMENT:  SELECT * FROM users`
	errs, failed, refused, last := parsePostgresLog(out)
	if errs != 1 || failed != 1 || refused != 1 || !strings.Contains(last, `relation "users" does not exist`) {
		t.Fatalf("got %d %d %d %q", errs, failed, refused, last)
	}
}

func TestParseMongoLog(t *testing.T) {
	out := `{"t":{"$date":"2026-11-12T10:00:00.000+00:00"},"s":"I","c":"ACCESS","id":20250,"ctx":"conn1","msg":"Authentication failed","attr":{"mechanism":"SCRAM-SHA-256"}}
{"t":{"$date":"2026-11-12T10:00:01.000+00:00"},"s":"E","c":"STORAGE","id":22435,"ctx":"conn2","msg":"WiredTiger error","attr":{"error":"No space left"}}
{"t":{"$date":"2026-11-12T10:00:02.000+00:00"},"s":"I","c":"NETWORK","id":22942,"ctx":"listener","msg":"Connection refused because there are too many open connections"}
not json`
	errs, failed, refused, last := parseMongoLog(out)
	if errs != 1 || failed != 1 || refused != 1 || last != "WiredTiger error: No space left" {
		t.Fatalf("got %d %d %d %q", errs, failed, refused, last)
	}
}

func TestCounterDiff(t *testing.T) {
	day := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	at := day.Add(26 * time.Hour)
	prevAt := at.Add(-24 * time.Hour)
	prev := &Database{Uptime: 100_000, Counters: map[string]int64{"Connections": 1000, "Max_used_connections": 10}}

	// same server, a day later
	d := newCounterDiff(map[string]int64{"Connections": 1500, "Max_used_connections": 10}, 100_000+86_400, prev, prevAt, at, day)
	if got := d.get("Connections"); got != 500 {
		t.Fatalf("same server: got %d", got)
	}
	if d.grew("Max_used_connections") {
		t.Fatal("peak didn't change, it shouldn't count as the day's")
	}

	// restarted an hour ago, the counter only covers that hour
	d = newCounterDiff(map[string]int64{"Connections": 40, "Max_used_connections": 3}, 3600, prev, prevAt, at, day)
	if got := d.get("Connections"); got != 40 || !d.grew("Max_used_connections") {
		t.Fatalf("restart: got %d", got)
	}

	// no summary the day before and a server up for weeks, the day's part is unknown
	d = newCounterDiff(map[string]int64{"Connections": 40}, 3_000_000, nil, time.Time{}, at, day)
	if got := d.get("Connections"); got != -1 {
		t.Fatalf("unknown: got %d", got)
	}
}

func TestCheckDatabases(t *testing.T) {
	b := &builder{s: &Summary{}, allowed: map[string]bool{"mysql_conf": true, "docker": true}}
	d := unknownDatabase("MariaDB", "mysql", "/mysql")
	d.Refused, d.MaxConnections, d.Errors, d.LastError, d.FailedLogins = 3, 100, 2, "Aborting", 25
	b.s.Databases = []Database{d}
	b.checkDatabases()
	if len(b.s.Attention) != 1 || b.s.Attention[0].Link != "/mysql/configuration" {
		t.Fatalf("attention: %+v", b.s.Attention)
	}
	if len(b.s.Look) != 2 || b.s.Look[0].Link != "/containers/logs?container=mariadb&lines=100&ctx=service" {
		t.Fatalf("look: %+v", b.s.Look)
	}

	b = &builder{s: &Summary{Databases: []Database{unknownDatabase("PostgreSQL", "postgresql", "/postgresql")}}, allowed: map[string]bool{}}
	b.checkDatabases()
	if len(b.s.Good) != 1 || len(b.s.Look) != 0 {
		t.Fatalf("quiet day: %+v %+v", b.s.Good, b.s.Look)
	}
}
