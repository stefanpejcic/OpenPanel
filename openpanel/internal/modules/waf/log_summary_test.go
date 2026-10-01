package waf

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func logLine(uri string, interrupted bool, status int, msgs ...string) json.RawMessage {
	type msg struct {
		ErrorMessage string `json:"error_message"`
	}
	var m []msg
	for _, s := range msgs {
		m = append(m, msg{s})
	}
	b, _ := json.Marshal(map[string]any{
		"transaction": map[string]any{
			"timestamp": "2026/10/01 09:00:36", "client_ip": "1.2.3.4", "is_interrupted": interrupted,
			"request": map[string]any{"method": "GET", "uri": uri}, "response": map[string]any{"status": status},
		},
		"messages": m,
	})
	return b
}

const (
	lfiMsg   = `Coraza: Warning. [id "930130"] [msg "Restricted File Access Attempt"] [data "Matched Data: /.env found within REQUEST_FILENAME: /.env"] [tag "attack-lfi"] [tag "OWASP_CRS"]`
	sqliMsg  = `Coraza: Warning. [id "942100"] [msg "SQL Injection Attack Detected via libinjection"] [tag "attack-sqli"]`
	scoreMsg = `Coraza: Access denied (phase 2). [id "949110"] [msg "Inbound Anomaly Score Exceeded"] [tag "anomaly-evaluation"]`
	hostMsg  = `Coraza: Warning. [id "920350"] [msg "Host header is a numeric IP address"] [tag "attack-protocol"]`
)

func TestSummarizeLogs(t *testing.T) {
	s := summarizeLogs([]json.RawMessage{
		logLine("/.env", true, 403, lfiMsg, scoreMsg),
		logLine("/?id=1", false, 200, sqliMsg, scoreMsg), // monitor only
		logLine("/?id=2", false, 200, sqliMsg, sqliMsg),  // duplicate rule in one request
		logLine("/about", false, 200, hostMsg),           // flagged, not enough score
		logLine("/favicon.ico", false, 404),              // relevant status, no rule
		json.RawMessage(`not json`),
	})
	if s.Blocked != 1 || s.WouldBlock != 1 || s.Flagged != 2 || s.Other != 1 || len(s.Events) != 4 {
		t.Fatalf("counts: %+v", s)
	}
	if s.Events[0].Result != "blocked" || s.Events[0].Rules[0].Category != "Access to restricted files" || s.Events[0].Rules[0].Data != "/.env found within REQUEST_FILENAME: /.env" {
		t.Fatalf("event: %+v", s.Events[0])
	}
	if s.Events[1].Result != "would_block" || len(s.Events[2].Rules) != 1 {
		t.Fatalf("monitor/dupe: %+v %+v", s.Events[1], s.Events[2])
	}
	for _, e := range s.Events {
		for _, r := range e.Rules {
			if isScoringRule(r.ID) {
				t.Fatalf("scoring rule %s shown as a reason", r.ID)
			}
		}
	}
	// 942100: 2 hits, 1 would-block; 930130: 1 blocked; 920350: 1 flagged
	if len(s.Groups) != 3 || s.Groups[0].ID != "930130" && s.Groups[0].ID != "942100" || s.Groups[2].ID != "920350" {
		t.Fatalf("groups order: %+v", s.Groups)
	}
	for _, g := range s.Groups {
		if g.ID == "942100" && (g.Count != 2 || g.Blocked != 1 || len(g.Paths) != 1 || g.Paths[0] != "/") {
			t.Fatalf("sqli group: %+v", g)
		}
	}
}

func TestRuleCategory(t *testing.T) {
	if ruleCategory([]string{"OWASP_CRS", "attack-xss"}) != "Cross-site scripting (XSS)" {
		t.Fatal("xss")
	}
	if ruleCategory([]string{"attack-new-thing"}) != "new thing" {
		t.Fatal("fallback")
	}
	if ruleCategory(nil) != "Suspicious request" {
		t.Fatal("none")
	}
}

func TestReadLogTail(t *testing.T) {
	path := t.TempDir() + "/x.log"
	var b strings.Builder
	for i := 0; i < 3000; i++ {
		b.WriteString(`{"n":` + strconv.Itoa(i) + `}` + "\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	got := readLogTail(path, 500)
	if len(got) != 500 || string(got[0]) != `{"n":2999}` || string(got[499]) != `{"n":2500}` {
		t.Fatalf("len=%d first=%s last=%s", len(got), got[0], got[len(got)-1])
	}
	all := readLogTail(path, 10000)
	if len(all) != 3000 || string(all[2999]) != `{"n":0}` {
		t.Fatalf("all: len=%d last=%s", len(all), all[len(all)-1])
	}
	if readLogTail(t.TempDir()+"/missing.log", 10) != nil {
		t.Fatal("missing file")
	}
}

func TestRenderWAFDomainPageRecentBlocked(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	logs := []json.RawMessage{logLine("/.env", true, 403, lfiMsg, scoreMsg)}
	data := WAFDomainPageData{
		LayoutData: baseLayout(mgr, "/server/waf/example.com"), Domain: "example.com", Status: "On",
		Summary: summarizeLogs(logs), RemovedRules: []string{"920350"},
	}
	w := httptest.NewRecorder()
	if err := wafDomainPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Recent blocked requests", "Why requests were stopped", "Access to restricted files", "View full logs", "function fmt(", "/server/waf/log/example.com", "ruleRequest('enable-rule'", "wafRecent("} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestPurgeDomainLogAndRecentSummary(t *testing.T) {
	dir := t.TempDir() + "/"
	orig := wafLogDir
	wafLogDir = dir
	t.Cleanup(func() { wafLogDir = orig })

	var b strings.Builder
	for i := 0; i < 15; i++ {
		b.Write(logLine("/.env", true, 403, lfiMsg, scoreMsg))
		b.WriteString("\n")
	}
	if err := os.WriteFile(wafLogPath("example.com"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	s := recentSummary("example.com")
	if len(s.Events) != latestEventsShown || s.Blocked != 15 || s.Groups[0].Count != 15 {
		t.Fatalf("events=%d blocked=%d group=%d", len(s.Events), s.Blocked, s.Groups[0].Count)
	}
	if err := purgeDomainLog("example.com"); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(wafLogPath("example.com")); err != nil || info.Size() != 0 {
		t.Fatal("log not emptied in place")
	}
	empty := recentSummary("example.com")
	b2, _ := json.Marshal(empty)
	if len(empty.Events) != 0 || !strings.Contains(string(b2), `"events":[]`) || !strings.Contains(string(b2), `"groups":[]`) {
		t.Fatalf("empty log must give empty lists, not null: %s", b2)
	}
	if b3, _ := json.Marshal(recentSummary("never-logged.com")); !strings.Contains(string(b3), `"events":[]`) {
		t.Fatalf("missing log must give empty lists too: %s", b3)
	}
	if err := purgeDomainLog("missing.com"); err != nil {
		t.Fatalf("missing log should be a no-op, got %v", err)
	}
	if _, err := os.Stat(wafLogPath("missing.com")); !os.IsNotExist(err) {
		t.Fatal("purge must not create a log file")
	}
}
