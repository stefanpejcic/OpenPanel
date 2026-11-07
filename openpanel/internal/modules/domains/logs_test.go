package domains

import (
	"net/url"
	"os"
	"strings"
	"testing"
)

// testdata/access.log is real Caddy lines from the test server (nodejs, waf and a redirect)
func readAccessLog(t *testing.T) []AccessLogEntry {
	t.Helper()
	b, err := os.ReadFile("testdata/access.log")
	if err != nil {
		t.Fatal(err)
	}
	return parseAccessLog(string(b) + "not json\n\n")
}

func TestParseAccessLog(t *testing.T) {
	entries := readAccessLog(t)
	if len(entries) != 8 {
		t.Fatalf("want 8 entries, got %d", len(entries))
	}
	// newest first means the file read backwards
	if entries[0].Status != 302 || entries[7].Request.URI != "/health" {
		t.Errorf("order wrong: first %d, last %s", entries[0].Status, entries[7].Request.URI)
	}
	e := entries[7]
	if e.UserAgent() != "curl/8.12.1" || e.Referer() != "" || e.Request.Proto != "HTTP/2.0" || e.TLSLabel() != "TLS 1.3 · h2" {
		t.Errorf("request fields wrong: ua=%q ref=%q proto=%q tls=%q", e.UserAgent(), e.Referer(), e.Request.Proto, e.TLSLabel())
	}
	if e.ContentType() != "application/json; charset=utf-8" || e.Server() != "Apache/2.4.69 (Unix) OpenSSL/3.5.9" || e.BytesRead != 0 {
		t.Errorf("response fields wrong: ctype=%q server=%q", e.ContentType(), e.Server())
	}
	if !strings.HasPrefix(e.Raw, `{"level":"info","ts":1793275664.6688836,`) {
		t.Errorf("raw line not kept: %q", e.Raw[:40])
	}
	waf := entries[2]
	if waf.Status != 403 || !strings.Contains(waf.UserAgent(), "sqlmap-test") {
		t.Errorf("waf line wrong: %+v", waf)
	}
	if (AccessLogEntry{}).TLSLabel() != "none" {
		t.Error("plain http should read none")
	}
	if headerValue(map[string][]string{"referer": {"https://x.test/"}}, "Referer") != "https://x.test/" {
		t.Error("header lookup should ignore case")
	}
}

func TestBuildAccessLogRows(t *testing.T) {
	entries := readAccessLog(t)
	rows, details := buildAccessLogRows("waf.tests.openpanel.org", entries)
	waf := rows[2]
	if waf.Path != "/?id=1' OR '1'='1" || waf.Tone != "is-warning" || waf.DurText != "413 ms" {
		t.Errorf("waf row wrong: path=%q tone=%q dur=%q", waf.Path, waf.Tone, waf.DurText)
	}
	if details[2].URL != "https://waf.tests.openpanel.org/?id=1%27%20OR%20%271%27=%271" || details[2].Shown != "https://waf.tests.openpanel.org/?id=1' OR '1'='1" {
		t.Errorf("detail url wrong: %q / %q", details[2].URL, details[2].Shown)
	}
	slow := rows[5]
	if !slow.Slow || slow.DurText != "17.7 s" || slow.Day != "Oct 29" || slow.Clock != "12:21:49" || slow.Stamp != "2026-10-29 12:21:49.783 UTC" {
		t.Errorf("slow row wrong: %+v", slow)
	}
	if rows[4].Tone != "is-danger" || rows[0].Tone != "is-accent" || rows[7].Tone != "is-success" || rows[3].SizeText != "299 B" {
		t.Errorf("tones/sizes wrong")
	}
	if humanBytes(15874) != "15.9 KB" || humanBytes(2_400_000) != "2.4 MB" || humanDuration(5.012) != "5.01 s" {
		t.Error("human formats wrong")
	}
	if details[7].Response != "33 B in 9 ms" || details[7].CType != "application/json; charset=utf-8" {
		t.Errorf("detail response wrong: %+v", details[7])
	}
}

func TestBuildDomainLogsData(t *testing.T) {
	entries := readAccessLog(t)

	d := buildDomainLogsData("n.test", entries, url.Values{}, 3, 10000)
	if d.TotalLines != 8 || d.Matched != 8 || d.TotalPages != 3 || len(d.Rows) != 3 || d.Counts[""] != 8 || d.Counts["2"] != 3 || d.Counts["3"] != 2 || d.Counts["4"] != 1 || d.Counts["5"] != 2 {
		t.Fatalf("unfiltered wrong: %+v", d.Counts)
	}

	d = buildDomainLogsData("n.test", entries, url.Values{"status": {"5"}, "page": {"9"}}, 3, 10000)
	if d.Matched != 2 || d.CurrentPage != 1 || d.Rows[0].Status != 500 || d.Counts["2"] != 3 {
		t.Errorf("status chip wrong: matched=%d page=%d", d.Matched, d.CurrentPage)
	}

	d = buildDomainLogsData("n.test", entries, url.Values{"q": {"SQLMAP"}, "view": {"raw"}}, 3, 10000)
	if d.Matched != 1 || d.Counts[""] != 1 || d.View != "raw" || d.Rows[0].Status != 403 {
		t.Errorf("search should match the user agent: matched=%d", d.Matched)
	}
	if got := d.URL(2, "4"); got != "/domains/log/n.test?page=2&q=SQLMAP&status=4&view=raw" {
		t.Errorf("URL = %q", got)
	}

	d = buildDomainLogsData("n.test", entries, url.Values{"status": {"9"}, "show_all": {"true"}}, 3, 10000)
	if d.Status != "" || !d.ShowAll || d.TotalPages != 1 || len(d.Rows) != 8 {
		t.Errorf("show all / bad status wrong: status=%q pages=%d rows=%d", d.Status, d.TotalPages, len(d.Rows))
	}
	if d := buildDomainLogsData("n.test", entries, url.Values{"show_all": {"true"}}, 3, 5); d.ShowAll {
		t.Error("show all should be refused past the cap")
	}
}

func TestAccessLogPath(t *testing.T) {
	for _, bad := range []string{"", ".", "..", "../etc", "a/b", `a\b`, "a\x00b"} {
		if _, ok := accessLogPath(bad); ok {
			t.Errorf("accessLogPath(%q) allowed", bad)
		}
	}
	if p, ok := accessLogPath("example.com"); !ok || p != "/var/log/caddy/domlogs/example.com/access.log" {
		t.Errorf("accessLogPath = %q %v", p, ok)
	}
}
