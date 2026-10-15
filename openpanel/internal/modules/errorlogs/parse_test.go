package errorlogs

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// captured from php-fpm-8.5 (shinsenter image) on the test server
const phpSample = `2026-10-15T17:05:36Z NOTICE: fpm is running, pid 1
2026-10-15T19:42:42Z NOTICE: PHP message: PHP Warning:  Undefined variable $undefined_var in /var/www/html/example.test.rs/zz.php on line 2
2026-10-15T19:42:42Z NOTICE: PHP message: claude plain error_log call
2026-10-15T19:42:42Z NOTICE: PHP message: PHP Fatal error:  Uncaught Error: Call to undefined function nonexistent_fn() in /var/www/html/example.test.rs/zz.php:5
2026-10-15T19:42:42Z Stack trace:
2026-10-15T19:42:42Z #0 {main}
2026-10-15T19:42:42Z   thrown in /var/www/html/example.test.rs/zz.php on line 5
2026-10-15T19:43:00Z WARNING: [pool www] server reached pm.max_children setting (5), consider raising it
`

func TestParsePHP(t *testing.T) {
	got := parseLines("php", "php-fpm-8.5", phpSample)
	if len(got) != 4 {
		t.Fatalf("want 4 entries, got %d: %+v", len(got), got)
	}
	if got[0].Level != "warning" || !strings.HasPrefix(got[0].Message, "Warning: Undefined variable") {
		t.Errorf("warning parsed wrong: %+v", got[0])
	}
	if got[1].Level != "info" || got[1].Message != "claude plain error_log call" {
		t.Errorf("plain error_log parsed wrong: %+v", got[1])
	}
	if got[2].Level != "critical" || !strings.Contains(got[2].Message, "Stack trace:\n#0 {main}\n  thrown in") {
		t.Errorf("fatal with stack trace parsed wrong: %+v", got[2])
	}
	if got[3].Level != "warning" || !strings.Contains(got[3].Message, "pm.max_children") {
		t.Errorf("fpm warning parsed wrong: %+v", got[3])
	}
	if want := time.Date(2026, 10, 15, 19, 42, 42, 0, time.UTC); !got[0].Time.Equal(want) {
		t.Errorf("timestamp = %v, want %v", got[0].Time, want)
	}
}

func TestParseWebservers(t *testing.T) {
	apache := `2026-10-15T17:05:37Z [Thu Oct 15 17:05:37.525906 2026] [ssl:warn] [pid 1:tid 1] AH01906: example.test.rs:443:0 server certificate is a CA certificate
2026-10-15T17:05:37Z [Thu Oct 15 17:05:37.529719 2026] [core:notice] [pid 1:tid 1] AH00094: Command line: 'httpd -D FOREGROUND'
2026-10-15T17:06:00Z [Thu Oct 15 17:06:00.000000 2026] [proxy_fcgi:error] [pid 9:tid 3] [client 1.2.3.4:5555] AH01071: Got error 'Primary script unknown'
2026-10-15T17:06:01Z 1.2.3.4 - - [15/Oct/2026:17:06:01 +0000] "GET / HTTP/1.1" 200 45`
	got := parseLines("apache", "apache", apache)
	if len(got) != 3 || got[0].Level != "warning" || got[1].Level != "notice" || got[2].Level != "error" {
		t.Fatalf("apache parsed wrong: %+v", got)
	}
	if !strings.HasPrefix(got[2].Message, "AH01071") {
		t.Errorf("apache prefixes not stripped: %q", got[2].Message)
	}

	nginx := `2026-10-15T17:06:00Z 2026/10/15 17:06:00 [error] 29#29: *1 connect() failed (111: Connection refused) while connecting to upstream, client: 1.2.3.4, server: example.com, host: "example.com"
2026-10-15T17:06:01Z 10.0.0.1 - - [15/Oct/2026:17:06:01 +0000] "GET / HTTP/1.1" 200 612`
	got = parseLines("nginx", "nginx", nginx)
	if len(got) != 1 || got[0].Level != "error" || !strings.HasPrefix(got[0].Message, "connect() failed") {
		t.Fatalf("nginx parsed wrong: %+v", got)
	}
}

func TestParseDatabases(t *testing.T) {
	got := parseLines("mysql", "mariadb", "2026-10-15T17:06:00Z 2026-10-15 17:06:00 0 [Warning] Access denied for user 'x'@'localhost'\n2026-10-15T17:06:00Z 2026-10-15 17:06:00 0 [Note] InnoDB: ready")
	if len(got) != 1 || got[0].Level != "warning" {
		t.Fatalf("mysql parsed wrong: %+v", got)
	}
	got = parseLines("postgres", "postgres", "2026-10-15T17:06:00Z 2026-10-15 17:06:00.000 UTC [12] ERROR:  relation \"x\" does not exist\n2026-10-15T17:06:00Z 2026-10-15 17:06:00.000 UTC [12] STATEMENT:  select * from x")
	if len(got) != 1 || got[0].Level != "error" || !strings.Contains(got[0].Message, "STATEMENT: select * from x") {
		t.Fatalf("postgres parsed wrong: %+v", got)
	}
}

func TestParseCaddy5xx(t *testing.T) {
	from := time.Unix(1792085000, 0)
	content := `{"level":"info","ts":1792085164.9,"request":{"client_ip":"1.2.3.4","method":"GET","host":"a.com","uri":"/"},"status":200}
{"level":"error","ts":1792085165.5,"request":{"client_ip":"1.2.3.4","method":"POST","host":"a.com","uri":"/wp-login.php"},"status":502}
{"level":"error","ts":1792084000,"request":{"client_ip":"1.2.3.4","method":"GET","host":"a.com","uri":"/old"},"status":500}`
	got := parseCaddy5xx("a.com", content, from, time.Time{})
	if len(got) != 1 || got[0].Message != "HTTP 502 POST /wp-login.php (client 1.2.3.4)" || got[0].Domain != "a.com" {
		t.Fatalf("caddy parsed wrong: %+v", got)
	}
}

func TestDomainMatcher(t *testing.T) {
	m := newDomainMatcher([]appctx.Domain{
		{DomainURL: "example.com", Docroot: "/var/www/html/example.com"},
		{DomainURL: "shop.example.com", Docroot: "/var/www/html/example.com/shop"},
		{DomainURL: "example.com.au", Docroot: "/var/www/html/example.com.au"},
	})
	cases := map[string]string{
		"PHP Warning: x in /var/www/html/example.com/index.php on line 1":      "example.com",
		"PHP Warning: x in /var/www/html/example.com/shop/index.php on line 1": "shop.example.com",
		"PHP Warning: x in /var/www/html/example.com.au/index.php:3":           "example.com.au",
		`connect() failed, server: example.com, host: "example.com"`:           "example.com",
		"AH01906: shop.example.com:443:0 server certificate":                   "shop.example.com",
		"nothing to see at notexample.com":                                     "",
	}
	for msg, want := range cases {
		if got := m.match(msg); got != want {
			t.Errorf("match(%q) = %q, want %q", msg, got, want)
		}
	}
}

func TestFilterEntries(t *testing.T) {
	base := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	all := []Entry{
		{Time: base, Service: "php-fpm-8.2", Level: "notice", Message: "Deprecated: x in /var/www/html/a.com/i.php"},
		{Time: base.Add(time.Minute), Service: "php-fpm-8.2", Level: "critical", Message: "Fatal error: y in /var/www/html/b.com/i.php"},
		{Time: base.Add(2 * time.Minute), Service: "apache", Level: "warning", Message: "AH01909: a.com:443:0 cert mismatch"},
	}
	m := newDomainMatcher([]appctx.Domain{{DomainURL: "a.com", Docroot: "/var/www/html/a.com"}, {DomainURL: "b.com", Docroot: "/var/www/html/b.com"}})

	res := filterEntries(append([]Entry{}, all...), m, Query{Domain: "a.com"})
	if res.Total != 2 || res.Entries[0].Service != "apache" {
		t.Errorf("domain filter / newest-first wrong: %+v", res)
	}
	sub := append(append([]Entry{}, all...), Entry{Time: base.Add(3 * time.Minute), Domain: "a.com/blog", Level: "error", Message: "x"})
	if res := filterEntries(sub, m, Query{Domain: "a.com"}); res.Total != 3 {
		t.Errorf("domain filter should include subfolder sites: %+v", res)
	}
	sub = append(append([]Entry{}, all...), Entry{Time: base.Add(3 * time.Minute), Domain: "a.com/blog", Level: "error", Message: "x"})
	if res := filterEntries(sub, m, Query{Domain: "a.com", Exact: true}); res.Total != 2 {
		t.Errorf("exact filter should leave subfolder sites out: %+v", res)
	}
	res = filterEntries(append([]Entry{}, all...), m, Query{Level: "warning"})
	if res.Total != 2 {
		t.Errorf("level filter wrong: %+v", res)
	}
	res = filterEntries(append([]Entry{}, all...), m, Query{Search: "FATAL"})
	if res.Total != 1 || res.Entries[0].Domain != "b.com" {
		t.Errorf("search filter wrong: %+v", res)
	}
	res = filterEntries(append([]Entry{}, all...), m, Query{Limit: 1})
	if !res.Truncated || len(res.Entries) != 1 || res.Total != 3 {
		t.Errorf("limit wrong: %+v", res)
	}
}

func TestParseQuery(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	q := parseQuery(httptest.NewRequest("GET", "/error-logs?level=nope&limit=999999&tz=99999", nil), now)
	if !q.From.IsZero() || !q.To.IsZero() || q.Level != "" || q.Limit != defaultLimit || q.TZ != 0 {
		t.Errorf("defaults wrong: %+v", q)
	}

	// picked days start at the viewer's midnight, UTC+2 here
	q = parseQuery(httptest.NewRequest("GET", "/error-logs?from=2026-10-01&to=2026-10-03&tz=120", nil), now)
	if want := time.Date(2026, 9, 30, 22, 0, 0, 0, time.UTC); !q.From.Equal(want) {
		t.Errorf("from = %v, want %v", q.From.UTC(), want)
	}
	if want := time.Date(2026, 10, 3, 21, 59, 59, 999999999, time.UTC); !q.To.Equal(want) {
		t.Errorf("to = %v, want %v", q.To.UTC(), want)
	}
	if q.FromDay != "2026-10-01" || q.ToDay != "2026-10-03" {
		t.Errorf("days not kept for the picker: %+v", q)
	}

	q = parseQuery(httptest.NewRequest("GET", "/error-logs?from=2026-10-10&to=2026-10-02", nil), now)
	if q.FromDay != "2026-10-02" || q.ToDay != "2026-10-10" || !q.From.Before(q.To) {
		t.Errorf("reversed range not swapped: %+v", q)
	}

	q = parseQuery(httptest.NewRequest("GET", "/error-logs?from=2026-10-14&to=2026-10-15", nil), now)
	if !q.To.IsZero() {
		t.Errorf("range ending today should stay open: %+v", q)
	}

	q = parseQuery(httptest.NewRequest("GET", "/error-logs?from=2026-10-15T08:00:00Z&to=2026-10-15T09:00:00Z", nil), now)
	if q.From.Hour() != 8 || q.To.Hour() != 9 || q.FromDay != "" {
		t.Errorf("RFC3339 from the API wrong: %+v", q)
	}

	q = parseQuery(httptest.NewRequest("GET", "/error-logs?from=garbage", nil), now)
	if !q.From.IsZero() {
		t.Errorf("bad from should be ignored: %+v", q)
	}
}

func TestTailFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "access.log")
	if err := os.WriteFile(p, []byte("first line\nsecond line\nthird line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := tailFile(p, 15); got != "third line\n" {
		t.Errorf("tailFile = %q", got)
	}
	if got := tailFile(p, 1<<20); !strings.HasPrefix(got, "first line") {
		t.Errorf("tailFile whole = %q", got)
	}
}

// captured from openlitespeed's /usr/local/lsws/logs/error.log on the test server
const olsSample = `2026-10-15 20:00:44.923807 [NOTICE] [80] LiteSpeed/1.9.3 Open starts successfully!
2026-10-15 20:00:59.905046 [NOTICE] [77] [10.89.2.2:48506#ols.test.rs] [STDERR] PHP Warning:  Undefined variable $undef in /var/www/html/ols.test.rs/zz.php on line 2
2026-10-15 20:00:59.905922 [NOTICE] [77] [10.89.2.2:48506#ols.test.rs] [STDERR] claude ols plain
2026-10-15 20:00:59.905941 [NOTICE] [77] [10.89.2.2:48506#ols.test.rs] [STDERR] PHP Fatal error:  Uncaught Error: Call to undefined function nonexistent_fn() in /var/www/html/ols.test.rs/zz.php:5
Stack trace:
#0 {main}
  thrown in /var/www/html/ols.test.rs/zz.php on line 5
2026-10-15 20:01:10.000001 [NOTICE] [80] [config:server] reload requested
Stack trace that belongs to a skipped notice
2026-10-15 20:01:12.500000 [ERROR] [80] [10.89.2.2:48600#ols.test.rs] [Module:Cache] cache dir not writable
2026-10-15 19:00:00.000000 [ERROR] [80] too old, outside the window`

func TestParseOLS(t *testing.T) {
	from := time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)
	got := parseOLS("openlitespeed", olsSample, from, time.Time{})
	if len(got) != 4 {
		t.Fatalf("want 4 entries, got %d: %+v", len(got), got)
	}
	if got[0].Source != "php" || got[0].Level != "warning" || got[0].Domain != "ols.test.rs" || !strings.HasPrefix(got[0].Message, "Warning: Undefined variable") {
		t.Errorf("php warning parsed wrong: %+v", got[0])
	}
	if got[1].Level != "info" || got[1].Message != "claude ols plain" {
		t.Errorf("plain error_log parsed wrong: %+v", got[1])
	}
	if got[2].Level != "critical" || !strings.HasSuffix(got[2].Message, "Stack trace:\n#0 {main}\n  thrown in /var/www/html/ols.test.rs/zz.php on line 5") {
		t.Errorf("fatal with stack trace parsed wrong: %q", got[2].Message)
	}
	if got[3].Source != "webserver" || got[3].Level != "error" || got[3].Message != "[Module:Cache] cache dir not writable" {
		t.Errorf("ols error parsed wrong: %+v", got[3])
	}
	if want := time.Date(2026, 10, 15, 20, 0, 59, 905046000, time.UTC); !got[0].Time.Equal(want) {
		t.Errorf("timestamp = %v, want %v", got[0].Time, want)
	}
}

func TestSourceKind(t *testing.T) {
	cases := []struct{ name, ws, sql, kind string }{
		{"openlitespeed", "openlitespeed", "mysql", "ols"},
		{"openlitespeed", "apache", "mysql", ""},
		{"apache", "openlitespeed", "mysql", ""},
		{"php-fpm-8.2", "nginx", "mysql", "php"},
		{"mariadb", "nginx", "mariadb", "mysql"},
		{"mysql", "nginx", "mariadb", ""},
	}
	for _, c := range cases {
		if kind, _ := sourceKind(c.name, c.ws, c.sql); kind != c.kind {
			t.Errorf("sourceKind(%q, ws=%q, sql=%q) = %q, want %q", c.name, c.ws, c.sql, kind, c.kind)
		}
	}
}

func TestSitesAttribution(t *testing.T) {
	domains := []appctx.Domain{{DomainURL: "a.com", Docroot: "/var/www/html/a.com"}, {DomainURL: "b.com", Docroot: "/var/www/html/b.com"}}
	sites := []Site{{Name: "a.com", Type: "WordPress"}, {Name: "a.com/blog", Type: "WordPress"}, {Name: "a.com/shop/v2", Type: "PHP"}, {Name: "b.com/api", Type: "NodeJS", Container: "B-API"}}

	targets := buildTargets(domains, sites)
	var got []string
	for _, tg := range targets {
		got = append(got, tg.Value)
	}
	if strings.Join(got, ",") != "a.com,a.com/blog,a.com/shop/v2,b.com,b.com/api" || targets[0].Sub || !targets[1].Sub {
		t.Errorf("targets = %+v", targets)
	}

	m := newDomainMatcher(siteDocroots(domains, sites))
	for msg, want := range map[string]string{
		"x in /var/www/html/a.com/index.php":           "a.com",
		"x in /var/www/html/a.com/blog/wp-load.php":    "a.com/blog",
		"x in /var/www/html/a.com/shop/v2/index.php:3": "a.com/shop/v2",
		"x in /var/www/html/a.com/blogger/index.php":   "a.com",
	} {
		if g := m.match(msg); g != want {
			t.Errorf("match(%q) = %q, want %q", msg, g, want)
		}
	}

	for uri, want := range map[string]string{"/blog/wp-admin": "a.com/blog", "/blog": "a.com/blog", "/blog?x=1": "a.com/blog", "/blogger": "", "/shop/v2/cart": "a.com/shop/v2", "/": ""} {
		if g := subsiteFor("a.com", uri, sites); g != want {
			t.Errorf("subsiteFor(%q) = %q, want %q", uri, g, want)
		}
	}

	if !inTarget("a.com/blog", "a.com") || inTarget("a.com", "a.com/blog") || inTarget("a.com.au", "a.com") || !inTarget("x", "") {
		t.Error("inTarget wrong")
	}
}

func TestParseApp(t *testing.T) {
	out := `2026-10-15T20:00:00Z Server listening on port 3000
2026-10-15T20:00:01Z GET /health 200
2026-10-15T20:00:02Z Error: connect ECONNREFUSED 127.0.0.1:3306
2026-10-15T20:00:02Z     at TCPConnectWrap.afterConnect [as oncomplete] (node:net:1555:16)
2026-10-15T20:00:02Z     at Protocol._enqueue (/app/node_modules/mysql/lib/protocol/Protocol.js:144:48)
2026-10-15T20:00:03Z (node:1) Warning: something is deprecated
2026-10-15T20:00:04Z Traceback (most recent call last):
2026-10-15T20:00:04Z   File "/app/main.py", line 3, in <module>
2026-10-15T20:00:04Z ZeroDivisionError: division by zero
2026-10-15T20:00:05Z process.on uncaughtException handler fired
2026-10-15T20:00:06Z all good again`
	got := parseApp("b-api", "b.com/api", out)
	if len(got) != 4 {
		t.Fatalf("want 4 entries, got %d: %+v", len(got), got)
	}
	if got[0].Level != "error" || got[0].Domain != "b.com/api" || strings.Count(got[0].Message, "\n") != 2 {
		t.Errorf("node error with stack wrong: %+v", got[0])
	}
	if got[1].Level != "warning" {
		t.Errorf("warning wrong: %+v", got[1])
	}
	if !strings.HasPrefix(got[2].Message, "Traceback") || !strings.HasSuffix(got[2].Message, "\nZeroDivisionError: division by zero") {
		t.Errorf("python traceback wrong: %q", got[2].Message)
	}
	if got[3].Level != "critical" {
		t.Errorf("uncaught wrong: %+v", got[3])
	}
	if n := len(parseApp("x", "y", "2026-10-15T20:00:00Z Everything is fine, every request served")); n != 0 {
		t.Errorf("plain words starting with e matched as errno, got %d", n)
	}
}
