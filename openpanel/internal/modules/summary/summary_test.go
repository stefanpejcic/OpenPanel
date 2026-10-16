package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/errorlogs"
)

func testBuilder(t *testing.T, allowed ...string) *builder {
	t.Helper()
	dir := t.TempDir()
	caddyLogDir, wafLogDir = dir+"/domlogs/", dir+"/waf/"
	usersDir, vulnDir, homeDir = dir+"/users/", dir+"/vuln/", dir+"/home/"
	quotaReportPath = dir + "/quota_report.json"
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	set := map[string]bool{}
	for _, f := range allowed {
		set[f] = true
	}
	s := &Summary{Username: "stefan", Date: "2026-10-15", From: from, To: from.AddDate(0, 0, 1)}
	return &builder{
		username: "stefan", userContext: "stefan", allowed: set, from: s.From, to: s.To, s: s, buckets: map[string]*Website{},
		domains: []appctx.Domain{{DomainURL: "a.com", Docroot: "/var/www/html/a.com"}},
		sites:   []errorlogs.Site{{Name: "a.com", Type: "wordpress"}, {Name: "a.com/blog", Type: "wordpress"}},
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func access(ts time.Time, ip, uri string, status int, size int64) string {
	return fmt.Sprintf(`{"ts":%d.5,"request":{"client_ip":"%s","uri":"%s"},"size":%d,"status":%d}`+"\n", ts.Unix(), ip, uri, size, status)
}

func TestCollectTraffic(t *testing.T) {
	b := testBuilder(t, "ip_blocker")
	at := b.from.Add(10 * time.Hour)
	var log strings.Builder
	log.WriteString(access(at, "1.1.1.1", "/", 200, 1000))
	log.WriteString(access(at, "1.1.1.1", "/shop?x=1", 200, 1000))
	log.WriteString(access(at, "2.2.2.2", "/shop", 200, 1000))
	log.WriteString(access(at, "2.2.2.2", "/missing", 404, 10))
	log.WriteString(access(at, "3.3.3.3", "/blog/post", 500, 10))
	log.WriteString(access(b.from.Add(-time.Hour), "9.9.9.9", "/", 200, 1000)) // day before
	for i := 0; i < 8; i++ {
		log.WriteString(access(at, "6.6.6.6", "/", 429, 0))
	}
	log.WriteString(access(at, "7.7.7.7", "/", 429, 0))
	write(t, caddyLogDir+"a.com/access.log", log.String())

	b.collectTraffic()
	b.finishWebsites()

	if len(b.s.Websites) != 2 {
		t.Fatalf("websites = %+v", b.s.Websites)
	}
	a, blog := b.s.Websites[0], b.s.Websites[1]
	if a.Name != "a.com" || a.Requests != 13 || a.UniqueIPs != 4 || a.NotFound != 1 || a.RateLimited != 9 || a.Bytes != 3010 {
		t.Errorf("a.com = %+v", a)
	}
	if strings.Join(a.TopPages, ",") != "/shop,/" {
		t.Errorf("top pages = %v", a.TopPages)
	}
	if blog.Name != "a.com/blog" || blog.ServerErrors != 1 {
		t.Errorf("blog = %+v", blog)
	}
	if len(b.s.Look) != 1 || !strings.Contains(b.s.Look[0].Detail, "6.6.6.6") || b.s.Look[0].Link != "/security/ip-blocker" {
		t.Errorf("rate limit item = %+v", b.s.Look)
	}
}

func TestRateLimitWithoutIPBlocker(t *testing.T) {
	b := testBuilder(t)
	at := b.from.Add(time.Hour)
	write(t, caddyLogDir+"a.com/access.log", access(at, "6.6.6.6", "/", 429, 0)+access(at, "6.6.6.6", "/", 429, 0))
	b.collectTraffic()
	if len(b.s.Look) != 1 || b.s.Look[0].Link != "" || !strings.Contains(b.s.Look[0].Action, "hosting provider") {
		t.Errorf("without ip_blocker the user should be pointed at the provider: %+v", b.s.Look)
	}
}

func TestServerErrorsWithoutErrorLogs(t *testing.T) {
	b := testBuilder(t, "domain_logs")
	w := b.website("a.com")
	w.Requests, w.ServerErrors = 100, 12
	b.checkErrors(nil)
	if len(b.s.Attention) != 1 || b.s.Attention[0].Link != "/domains/log/a.com" {
		t.Fatalf("12 of 100 should need attention with the access log link: %+v", b.s.Attention)
	}

	b = testBuilder(t)
	w = b.website("a.com/blog")
	w.Requests, w.ServerErrors = 1000, 2
	b.checkErrors(nil)
	if len(b.s.Look) != 1 || b.s.Look[0].Link != "" || !strings.Contains(b.s.Look[0].Action, "hosting provider") {
		t.Errorf("a couple of errors is only worth a look, no link without features: %+v", b.s.Look)
	}
}

func TestCheckDiskSuggestions(t *testing.T) {
	report := `{"users":[{"username":"stefan","disk_used":9000000,"disk_hard":10000000,"inodes_used":10,"inodes_hard":1000}]}`

	b := testBuilder(t, "trash", "disk_usage")
	write(t, quotaReportPath, report)
	write(t, homeDir+"stefan/.local/share/Trash/files/big.zip", strings.Repeat("x", 2*1024*1024))
	// trash is 2 MB of a ~8.6 GB disk, too small to be the fix
	b.checkDisk(nil)
	if len(b.s.Look) != 1 || b.s.Look[0].Link != "/disk-usage/" {
		t.Fatalf("90%% full should point at disk usage: %+v", b.s.Look)
	}

	b = testBuilder(t)
	write(t, quotaReportPath, strings.Replace(report, `"disk_used":9000000`, `"disk_used":9600000`, 1))
	b.checkDisk(nil)
	if len(b.s.Attention) != 1 || b.s.Attention[0].Link != "" || !strings.Contains(b.s.Attention[0].Action, "hosting provider") {
		t.Errorf("96%% full without disk features should go to the provider: %+v", b.s.Attention)
	}
	if len(b.s.Usage) != 2 || b.s.Usage[0].Title != "Disk space" || b.s.Usage[0].Percent != 96 || b.s.Usage[0].Level() != "warning" || b.s.Usage[1].Total != "1,000 files" {
		t.Errorf("usage = %+v", b.s.Usage)
	}
}

func TestCollectActivity(t *testing.T) {
	b := testBuilder(t, "twofa")
	write(t, usersDir+"stefan/activity.log", strings.Join([]string{
		"2026-10-01 10:00:00  1.1.1.1 User stefan logged in with a password",
		"2026-10-15 08:00:00  1.1.1.1 User stefan logged in with a password",
		"2026-10-15 09:00:00  5.5.5.5 User stefan logged in with a password",
		"2026-10-15 09:05:00  5.5.5.5 User stefan added domain b.com.",
		"2026-10-15 09:06:00  5.5.5.5 User stefan logged out.",
		"2026-10-15 10:00:00  10.0.0.1 User stefan logged in via admin API",
		"2026-10-16 01:00:00  1.1.1.1 User stefan changed PHP version",
	}, "\n")+"\n")
	b.a = nil
	b.allowed["twofa"] = false
	b.collectActivity()
	if len(b.s.Logins) != 2 || b.s.Logins[0].New || !b.s.Logins[1].New {
		t.Errorf("logins = %+v", b.s.Logins)
	}
	if len(b.s.Changes) != 1 || b.s.Changes[0] != "09:05 added domain b.com" {
		t.Errorf("changes = %v", b.s.Changes)
	}
	if len(b.s.Look) != 1 || !strings.Contains(b.s.Look[0].Title, "5.5.5.5") || b.s.Look[0].Link != "/account" {
		t.Errorf("new IP item = %+v", b.s.Look)
	}
}

func TestCheckVulnerabilities(t *testing.T) {
	b := testBuilder(t, "wordpress")
	write(t, vulnDir+"a.com_blog.json", `{"timestamp":"x","plugin_cf7":{"data":{"vulnerability":[{"name":"CF7 < 5.3"},{"name":"CF7 < 5.9"},{"name":"CF7 < 6.0"}]}}}`)
	b.checkVulnerabilities()
	if len(b.s.Attention) != 1 || b.s.Attention[0].Link != "/website?domain=a.com/blog" || b.s.Attention[0].Detail != "They're in CF7. An update usually fixes them." {
		t.Fatalf("vulnerable site = %+v", b.s.Attention)
	}
	if len(b.s.Good) != 0 {
		t.Errorf("no all-clear when a site is vulnerable: %v", b.s.Good)
	}

	b = testBuilder(t)
	write(t, vulnDir+"a.com_blog.json", `{"plugin_cf7":{"data":{"vulnerability":[{"name":"x"}]}}}`)
	if b.checkVulnerabilities(); len(b.s.Attention) != 0 {
		t.Error("vulnerabilities shown without the wordpress feature")
	}
}

func TestParseCronRuns(t *testing.T) {
	out := `2026-10-15T01:00:00Z common.go:125 ▶ NOTICE [Job "backup-db" (a1b2c3)] Started - mysqldump
2026-10-15T01:00:02Z common.go:125 ▶ NOTICE [Job "backup-db" (a1b2c3)] Finished in "2.1s", failed: true, skipped: false, error: error non-zero exit code: 255
2026-10-15T02:00:02Z common.go:125 ▶ NOTICE [Job "backup-db" (a1b2c3)] Finished in "2.1s", failed: true, skipped: false, error: error non-zero exit code: 255
2026-10-15T02:00:02Z common.go:125 ▶ NOTICE [Job "wp-cron" (d4e5f6)] Finished in "0.4s", failed: false, skipped: false, error: none`
	jobs := parseCronRuns(out)
	if j := jobs["backup-db"]; j == nil || j.runs != 2 || j.failed != 2 || j.lastError != "error non-zero exit code: 255" {
		t.Errorf("backup-db = %+v", j)
	}
	if j := jobs["wp-cron"]; j == nil || j.runs != 1 || j.failed != 0 {
		t.Errorf("wp-cron = %+v", j)
	}
}

func TestContainerEvents(t *testing.T) {
	oom, crashed := containerEvents("died php-fpm-8.2 0\noom mariadb \ndied mariadb 137\ndied nodeapp 1\ndied nodeapp 1\ndied apache 143\n")
	if oom["mariadb"] != 1 || len(oom) != 1 {
		t.Errorf("oom = %v", oom)
	}
	if crashed["nodeapp"] != 2 || len(crashed) != 1 {
		t.Errorf("crashed = %v", crashed)
	}
}

func TestSizeBytes(t *testing.T) {
	for in, want := range map[string]int64{"1G": 1 << 30, "512M": 512 << 20, "1.5G": 3 << 29, "2048K": 2 << 20, "100": 100, "": 0, "abc": 0, "1GB": 1 << 30} {
		if got := sizeBytes(in); got != want {
			t.Errorf("sizeBytes(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestSaveLoadAndPrune(t *testing.T) {
	testBuilder(t)
	old := time.Now().AddDate(0, 0, -keepDays-5).Format("2006-01-02")
	write(t, summaryDir("stefan")+old+".json", "{}")
	s := Summary{Username: "stefan", Date: time.Now().Format("2006-01-02"), Attention: []Item{{Title: "x"}}}
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	if d := Dates("stefan"); len(d) != 1 || d[0] != s.Date {
		t.Errorf("old summary not pruned: %v", d)
	}
	if got, ok := Load("stefan", s.Date); !ok || got.Attention[0].Title != "x" {
		t.Errorf("load = %+v %v", got, ok)
	}
	if _, ok := Load("stefan", "../../etc/passwd"); ok {
		t.Error("path outside the folder loaded")
	}
}

func TestEmail(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	s := Summary{Username: "stefan", Date: "2026-10-15", From: from,
		Attention: []Item{{Title: "Malware found on your account", Action: "Review the quarantined files", Link: "/malware-scanner/quarantine"}},
		Look:      []Item{{Title: "Disk is 90% full", Action: "Ask your hosting provider for more disk space."}},
		Good:      []string{"SSL certificates valid"},
		Websites:  []Website{{Name: "a.com", Requests: 10, UniqueIPs: 3, Bytes: 2048}},
		Logins:    []Login{{IP: "1.1.1.1", Country: "RS"}, {IP: "1.1.1.1", Country: "RS"}},
	}
	if got := s.Subject(); got != "⚠ 1 thing needs your attention – 15 Oct summary for stefan" {
		t.Errorf("subject = %q", got)
	}
	body := s.Body("https://panel.example.com:2083/")
	for _, want := range []string{
		"NEEDS YOUR ATTENTION\n\n1. Malware found on your account\n   → Review the quarantined files: https://panel.example.com:2083/malware-scanner/quarantine",
		"• Disk is 90% full\n   → Ask your hosting provider for more disk space.\n",
		"✓ SSL certificates valid", "- a.com: 10 requests from 3 visitors, 2 KB of bandwidth", "You logged in 2 times:", "  00:00 from 1.1.1.1 (RS)",
		"See the full summary: https://panel.example.com:2083/summary/2026-10-15", "Hi stefan,",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q:\n%s", want, body)
		}
	}
	if (Summary{Username: "stefan", From: from}).Subject() != "✓ All good – 15 Oct summary for stefan" {
		t.Error("all good subject wrong")
	}
}

func TestRateLimitFromServerItself(t *testing.T) {
	b := testBuilder(t, "ip_blocker", "crons")
	at := b.from.Add(time.Hour)
	write(t, caddyLogDir+"a.com/access.log", access(at, "127.0.0.1", "/", 429, 0)+access(at, "127.0.0.1", "/", 429, 0))
	b.collectTraffic()
	if len(b.s.Look) != 1 || strings.Contains(b.s.Look[0].Action, "Block") || b.s.Look[0].Link != "/cronjobs" {
		t.Errorf("the server's own IP must never be suggested for blocking: %+v", b.s.Look)
	}
}

func TestHasCronJobs(t *testing.T) {
	b := testBuilder(t)
	if b.hasCronJobs() {
		t.Error("no crons.ini should mean no jobs")
	}
	write(t, homeDir+"stefan/crons.ini", "# [job-exec \"old\"]\n")
	if b.hasCronJobs() {
		t.Error("a commented out job is not a job")
	}
	write(t, homeDir+"stefan/crons.ini", "[job-exec \"backup\"]\nschedule = @daily\n")
	if !b.hasCronJobs() {
		t.Error("job not found")
	}
}

func TestWordsHelpers(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 4812: "4,812", 1234567: "1,234,567", -1500: "-1,500"} {
		if got := num(n); got != want {
			t.Errorf("num(%d) = %q, want %q", n, got, want)
		}
	}
	if joinAnd([]string{"A"}) != "A" || joinAnd([]string{"A", "B"}) != "A and B" || joinAnd([]string{"A", "B", "C"}) != "A, B and C" {
		t.Error("joinAnd wrong")
	}
	got := vulnerableNames([]string{"Contact Form 7 [contact-form-7] < 5.3.2", "Contact Form 7 [contact-form-7] < 5.9.2", "Astra < 4.1.6"})
	if strings.Join(got, "|") != "Astra|Contact Form 7" {
		t.Errorf("vulnerableNames = %v", got)
	}
	for _, c := range [][3]any{{112, 100, "12% more than the day before"}, {80, 100, "20% fewer than the day before"}, {102, 100, "about the same as the day before"}, {50, 0, ""}, {50, -1, ""}} {
		if got := change(c[0].(int), c[1].(int)); got != c[2].(string) {
			t.Errorf("change(%v, %v) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
	line := websiteLine(Website{Name: "a.com", Requests: 4812, UniqueIPs: 1236, Bytes: 2 << 20, NotFound: 14, ServerErrors: 1, PrevRequests: 4000})
	if line != "a.com: 4,812 requests from 1,236 visitors, 2 MB of bandwidth (20% more than the day before, 14 missing pages, 1 error page)" {
		t.Errorf("websiteLine = %q", line)
	}
}

func TestHeadlineAndBusiestPage(t *testing.T) {
	b := testBuilder(t)
	b.s.Websites = []Website{{Name: "a.com", Requests: 900, UniqueIPs: 300, TopPages: []string{"/shop"}, TopPageViews: 120}, {Name: "b.com", Requests: 100, UniqueIPs: 40}}
	b.prevTotal = 800
	b.busiestPage()
	if len(b.s.Good) != 1 || b.s.Good[0] != "Your busiest page was /shop on a.com, with 120 views." {
		t.Errorf("busiest page = %v", b.s.Good)
	}
	if got := b.headline(); got != "Your 2 websites handled 1,000 requests from 340 visitors, 25% more than the day before. Nothing needs your attention." {
		t.Errorf("headline = %q", got)
	}
	b.s.Attention = []Item{{Title: "x"}}
	if got := b.headline(); !strings.HasSuffix(got, " 1 thing needs your attention.") {
		t.Errorf("headline with attention = %q", got)
	}
	b.s.Attention, b.s.Look, b.s.Websites = nil, []Item{{}, {}}, nil
	if got := b.headline(); got != "Your websites had no visitors. Nothing urgent, but 2 things are worth a look." {
		t.Errorf("quiet headline = %q", got)
	}
}

func TestErrorPagesGrouped(t *testing.T) {
	b := testBuilder(t, "domain_logs")
	for i, n := range []int{64, 8, 5} {
		w := b.website([]string{"a.com", "b.com", "c.com"}[i])
		w.Requests, w.ServerErrors = 100, n
	}
	b.checkErrors(nil)
	if len(b.s.Attention) != 1 || b.s.Attention[0].Title != "Visitors saw error pages on 3 websites" || b.s.Attention[0].Link != "/domains/log" ||
		!strings.HasPrefix(b.s.Attention[0].Detail, "On a.com (64 times), b.com (8) and c.com (5).") {
		t.Fatalf("grouped item = %+v", b.s.Attention)
	}

	b = testBuilder(t, "domain_logs")
	w := b.website("a.com")
	w.Requests, w.ServerErrors, w.errorExample = 100, 20, "PHP Fatal error: boom"
	for _, name := range []string{"b.com", "c.com", "d.com"} {
		x := b.website(name)
		x.Requests, x.ServerErrors = 100, 15
	}
	b.checkErrors(nil)
	if len(b.s.Attention) != 2 || !strings.Contains(b.s.Attention[0].Detail, "boom") || b.s.Attention[1].Title != "Visitors saw error pages on 3 websites" {
		t.Errorf("a known cause should stay its own item: %+v", b.s.Attention)
	}
}

func TestLoginCountriesAndBars(t *testing.T) {
	b := testBuilder(t)
	write(t, usersDir+"stefan/.lastlogin", "IP: 5.5.5.5 - Country: rs - Login Time: 2026-10-15 09:00:00\nIP: 6.6.6.6 - Country: UNKNOWN - Login Time: x\njunk\n")
	write(t, usersDir+"stefan/activity.log", "2026-10-15 09:00:00  5.5.5.5 User stefan logged in with a password\n2026-10-15 10:00:00  6.6.6.6 User stefan logged in with a password\n")
	b.collectActivity()
	if len(b.s.Logins) != 2 || b.s.Logins[0].Country != "RS" || b.s.Logins[0].Flag() != "rs" || b.s.Logins[1].Flag() != "" {
		t.Errorf("logins = %+v", b.s.Logins)
	}
	for pct, want := range map[int]string{40: "ok", 85: "warning", 100: "full", 120: "full"} {
		u := UsageBar{Percent: pct, Limit: 85}
		if u.Level() != want || u.Width() > 100 {
			t.Errorf("bar %d%% = %s width %d, want %s", pct, u.Level(), u.Width(), want)
		}
	}
	b.s.Websites = []Website{{Name: "a.com", TopPages: []string{"/"}, TopPageViews: 50}}
	b.busiestPage()
	if b.s.Good[len(b.s.Good)-1] != "Your busiest page was the homepage of a.com, with 50 views." {
		t.Errorf("homepage line = %v", b.s.Good)
	}
}

func TestCheckResources(t *testing.T) {
	line := func(ts string, cpu, mem, tasks int) string {
		return fmt.Sprintf(`{"timestamp":"%s","memory":{"total":{"bytes":2147483648},"used":{"bytes":1932735283},"usage_pct":%d},"cpu":{"usage":{"pct":%d,"human":"x cores"},"total":{"pct":200,"human":"2.0 cores"}},"tasks":{"current":%d,"limit":2000,"usage_pct":%d}}`, ts, mem, cpu, tasks*20, tasks)
	}
	b := testBuilder(t, "usage")
	write(t, homeDir+"stefan/resource_usage.txt", strings.Join([]string{
		line("2026-10-14T23:00:00Z", 200, 99, 99), // day before, ignored
		line("2026-10-15T08:00:00Z", 100, 50, 10),
		line("2026-10-15T14:00:00Z", 190, 60, 10),
		line("2026-10-15T15:00:00Z", 80, 70, 10),
		line("2026-10-15T16:00:00Z", 80, 70, 10),
		line("2026-10-15T17:00:00Z", 80, 70, 10),
	}, "\n"))
	b.checkResources()
	if len(b.s.Look) != 1 || !strings.Contains(b.s.Look[0].Detail, "14:00") || !strings.Contains(b.s.Look[0].Detail, "1 of the 5") || b.s.Look[0].Link != "/server/usage" {
		t.Fatalf("one CPU peak should be worth a look: %+v", b.s.Look)
	}
	if len(b.s.Usage) != 2 || b.s.Usage[0].Percent != 95 || b.s.Usage[1].Percent != 70 || len(b.s.Attention) != 0 || len(b.s.Good) != 0 {
		t.Errorf("usage = %+v attention = %+v", b.s.Usage, b.s.Attention)
	}

	b = testBuilder(t)
	write(t, homeDir+"stefan/resource_usage.txt", line("2026-10-15T08:00:00Z", 50, 40, 95))
	b.checkResources()
	if len(b.s.Attention) != 1 || !strings.Contains(b.s.Attention[0].Title, "process limit") || b.s.Attention[0].Link != "" {
		t.Errorf("process limit should need attention: %+v", b.s.Attention)
	}
}
