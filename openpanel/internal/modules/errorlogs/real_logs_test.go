package errorlogs

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// testdata/*.log are `podman logs --timestamps` output copied from the test server's containers

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestParsePHPRealContainer(t *testing.T) {
	got := parseLines("php", "php-fpm-8.5", readTestdata(t, "php-fpm-8.5.log"))
	var levels []string
	for _, e := range got {
		levels = append(levels, e.Level)
	}
	// the JIT warning fpm logs with its own [date] prefix on start, 3 requests with 2 warnings and a fatal each, then the next start's JIT warning
	want := "warning,warning,warning,critical,warning,warning,critical,warning,warning,critical,warning"
	if strings.Join(levels, ",") != want {
		t.Fatalf("levels = %v, want %s", levels, want)
	}
	fatal := got[9].Message
	if !strings.HasSuffix(fatal, "index.php:1\nStack trace:\n#0 {main}\n  thrown in /var/www/html/php.tests.openpanel.org/composerapp/index.php on line 1") {
		t.Errorf("fatal error lost its stack trace or picked up more lines: %q", fatal)
	}
	for _, e := range got {
		for _, junk := range []string{"usermod", "shinsenter", "░", "Container     :", "PHP-FPM tuned", "ready to handle"} {
			if strings.Contains(e.Message, junk) {
				t.Errorf("container banner line %q glued onto %q", junk, e.Message[:min(len(e.Message), 60)])
			}
		}
	}
	jit := got[10]
	if !strings.HasPrefix(jit.Message, "Warning: JIT is incompatible") || strings.Contains(jit.Message, "\n") {
		t.Errorf("bracket-stamped fpm line parsed wrong: %q", jit.Message)
	}
}

func TestPHPContinuationRule(t *testing.T) {
	out := `2026-10-29T10:00:00.1Z NOTICE: PHP message: PHP Fatal error:  Uncaught Exception: a in /app/x.php:3
2026-10-29T10:00:00.1Z Stack trace:
2026-10-29T10:00:00.1Z #0 /app/y.php(9): f()
2026-10-29T10:00:00.1Z #1 {main}
2026-10-29T10:00:00.1Z Next RuntimeException: b in /app/x.php:5
2026-10-29T10:00:00.1Z   thrown in /app/x.php on line 5
2026-10-29T10:00:01.1Z Composer version 2.10.3
2026-10-29T10:00:01.1Z     with Zend OPcache v8.5.12
2026-10-29T10:00:02.1Z WARNING: [pool www] child 7 said into stderr: "x"
2026-10-29T10:00:02.1Z     indented right after a warning`
	got := parseLines("php", "php-fpm-8.5", out)
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d: %+v", len(got), got)
	}
	if strings.Count(got[0].Message, "\n") != 5 || strings.Contains(got[0].Message, "Composer") || strings.Contains(got[0].Message, "OPcache") {
		t.Errorf("fatal continuation wrong: %q", got[0].Message)
	}
	if !strings.HasSuffix(got[1].Message, "\n    indented right after a warning") {
		t.Errorf("indented line after an fpm warning should attach: %q", got[1].Message)
	}
}

func TestParseMariaDBRealContainer(t *testing.T) {
	got := parseLines("mysql", "mariadb", readTestdata(t, "mariadb.log"))
	if len(got) != 1 || got[0].Level != "warning" {
		t.Fatalf("want the one io_uring warning, got %+v", got)
	}
	if got[0].Message != "mariadbd: io_uring_queue_init() failed with ENOSYS: check seccomp filters, and the kernel version (newer than 5.1 required)" {
		t.Errorf("podman stamp left inside the message: %q", got[0].Message)
	}
	if c := cleanText("recovered pages: 0%2026-10-29T11:34:29.579538946Z  10%2026-10-29T11:34:29.582958337Z  21% 100%"); c != "recovered pages: 0% 10% 21% 100%" {
		t.Errorf("cleanText = %q", c)
	}
	// an app's own ISO time with millis stays
	if c := cleanText("at 2026-10-29T11:34:29.579Z the job ran"); c != "at 2026-10-29T11:34:29.579Z the job ran" {
		t.Errorf("cleanText ate a normal timestamp: %q", c)
	}
}

func TestParseAppStripsANSI(t *testing.T) {
	got := parseApp("uptimekuma", "files.tests.openpanel.org", readTestdata(t, "uptimekuma.log"))
	if len(got) != 1 {
		t.Fatalf("want the ENOENT line only, got %+v", got)
	}
	if got[0].Message != "2026-10-29T12:15:32Z [SETUP-DATABASE] INFO: db-config.json is not found or invalid: ENOENT: no such file or directory, open 'data/db-config.json'" {
		t.Errorf("colour codes left in: %q", got[0].Message)
	}
}

func TestProxyErrorsGetTheirSite(t *testing.T) {
	got := parseLines("apache", "apache", readTestdata(t, "apache-proxy.log"))
	if len(got) != 7 {
		t.Fatalf("want 7 apache errors, got %d", len(got))
	}
	m := newDomainMatcher([]appctx.Domain{{DomainURL: "nodejs.tests.openpanel.org", Docroot: "/var/www/html/nodejs.tests.openpanel.org"}})
	m.backends = appBackends([]Source{
		{Name: "nodeapp", Site: "nodejs.tests.openpanel.org", kind: "app"},
		{Name: "php-fpm-8.5", kind: "php"},
	})
	res := filterEntries(got, m, Query{})
	sites := map[string]int{}
	for _, e := range res.Entries {
		sites[e.Domain]++
	}
	// nodeapp maps to its site, pyap has no container any more so stays unassigned
	if sites["nodejs.tests.openpanel.org"] != 5 || sites[""] != 2 {
		t.Errorf("proxy errors per site = %v", sites)
	}
	if res := filterEntries(append([]Entry{}, got...), m, Query{Domain: "nodejs.tests.openpanel.org"}); res.Total != 5 {
		t.Errorf("website filter should find the proxy errors, got %d", res.Total)
	}
}

func TestSeverityFilter(t *testing.T) {
	all := []Entry{{Level: "warning", Message: "a"}, {Level: "error", Message: "b"}, {Level: "critical", Message: "c"}}
	if res := filterEntries(append([]Entry{}, all...), domainMatcher{}, Query{Severity: "error"}); res.Total != 1 || res.Entries[0].Message != "b" {
		t.Errorf("severity keeps one exact level: %+v", res)
	}
}

func TestParseQuerySeverity(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	if q := parseQuery(httptest.NewRequest("GET", "/error-logs?severity=notice", nil), now); q.Severity != "notice" {
		t.Errorf("severity = %q", q.Severity)
	}
	if q := parseQuery(httptest.NewRequest("GET", "/error-logs?severity=loud", nil), now); q.Severity != "" {
		t.Errorf("unknown severity kept: %q", q.Severity)
	}
}
