package botcontrol

import (
	"context"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyUserAgent(t *testing.T) {
	cases := []struct {
		ua       string
		wantName string
		wantCat  BotCategory
		wantAI   bool
	}{
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)", "GPTBot", CatAITraining, true},
		{"ClaudeBot/1.0; +claudebot@anthropic.com", "ClaudeBot", CatAITraining, true},
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", "Googlebot", CatSearchEngine, false},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/130.0.0.0", "Browser / Other", CatHuman, false},
		{"curl/8.5.0", "CLI / Script (curl/8.5.0)", CatScraper, false},
	}

	for _, tc := range cases {
		name, _, cat, isAI := classifyUserAgent(tc.ua)
		if name != tc.wantName || cat != tc.wantCat || isAI != tc.wantAI {
			t.Fatalf("classifyUserAgent(%q) = (%q, %q, %v), want (%q, %q, %v)", tc.ua, name, cat, isAI, tc.wantName, tc.wantCat, tc.wantAI)
		}
	}
}

func TestRewritePolicy(t *testing.T) {
	base := `http://example.com, http://www.example.com {
    route {
        import ratelimit example.com
        coraza_waf {
            SecRuleEngine Off
        }
        reverse_proxy http://127.0.0.1:8080 {
        }
    }
}

https://example.com, https://www.example.com {
    route {
        import ratelimit example.com
        coraza_waf {
            SecRuleEngine Off
        }
        reverse_proxy https://127.0.0.1:8443 {
        }
    }
}`

	if got := policyFromConfig(base); got != PolicyAllow {
		t.Fatalf("initial policy = %q, want allow", got)
	}

	blocked := mustRewrite(t, base, PolicyBlock)
	if got := policyFromConfig(blocked); got != PolicyBlock {
		t.Fatalf("after block policy = %q, want block", got)
	}
	if strings.Count(blocked, "import botcontrol_block") != 2 {
		t.Fatalf("expected two botcontrol_block directives (HTTP + HTTPS), got:\n%s", blocked)
	}

	limited := mustRewrite(t, blocked, PolicyRateLimit)
	if got := policyFromConfig(limited); got != PolicyRateLimit {
		t.Fatalf("after ratelimit policy = %q, want ratelimit", got)
	}
	if strings.Count(limited, "import botcontrol_ratelimit example.com") != 2 {
		t.Fatalf("expected two ratelimit imports in:\n%s", limited)
	}

	allowed := mustRewrite(t, limited, PolicyAllow)
	if got := policyFromConfig(allowed); got != PolicyAllow {
		t.Fatalf("after allow policy = %q, want allow", got)
	}
	if strings.Count(allowed, "# botcontrol: allow") != 2 {
		t.Fatalf("expected two allow markers in:\n%s", allowed)
	}
	if again := mustRewrite(t, allowed, PolicyBlock); strings.Count(again, "import botcontrol_block") != 2 {
		t.Fatalf("block after allow lost a block:\n%s", again)
	}
}

func mustRewrite(t *testing.T, content string, p Policy) string {
	t.Helper()
	out, err := rewritePolicy(content, "example.com", p)
	if err != nil {
		t.Fatalf("rewritePolicy(%s): %v", p, err)
	}
	return out
}

// a config with nothing to anchor on used to get a "# botcontrol: block" comment appended, which blocked nothing
func TestRewritePolicyNoAnchor(t *testing.T) {
	content := "example.com {\n    respond \"hi\"\n}\n"
	if _, err := rewritePolicy(content, "example.com", PolicyBlock); !errors.Is(err, errNoAnchor) {
		t.Fatalf("block without anchor: err = %v, want errNoAnchor", err)
	}
	if out, err := rewritePolicy(content, "example.com", PolicyAllow); err != nil || out != content {
		t.Fatalf("allow without anchor should be a no-op, got %q, %v", out, err)
	}
}

func TestSetDomainPolicyRollsBackWhenCaddyRejects(t *testing.T) {
	dir := t.TempDir()
	oldDir, oldReload := caddyDomainsDir, reloadCaddy
	t.Cleanup(func() { caddyDomainsDir, reloadCaddy = oldDir, oldReload })
	caddyDomainsDir = dir + string(filepath.Separator)

	orig := "https://example.com {\n    route {\n        # botcontrol: allow\n        reverse_proxy http://127.0.0.1:8080\n    }\n}\n"
	path := filepath.Join(dir, "example.com.conf")
	if err := os.WriteFile(path, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	reloadCaddy = func(context.Context) error { return errors.New("unrecognized directive: rate_limit") }
	err := setDomainPolicy(context.Background(), "example.com", PolicyRateLimit)
	if err == nil || !strings.Contains(err.Error(), "rate_limit") {
		t.Fatalf("err = %v, want caddy's message", err)
	}
	if got, _ := os.ReadFile(path); string(got) != orig {
		t.Fatalf("config not restored:\n%s", got)
	}

	reloadCaddy = func(context.Context) error { return nil }
	if err := setDomainPolicy(context.Background(), "example.com", PolicyBlock); err != nil {
		t.Fatal(err)
	}
	if getDomainPolicy("example.com") != PolicyBlock {
		t.Fatal("block not saved")
	}
}

func TestCaddyErrorLine(t *testing.T) {
	jsonOut := `{"level":"info","ts":1794591082.760779,"msg":"using config from file","file":"/etc/caddy/Caddyfile"}` + "\n" + `{"level":"error","ts":1794591082.7608736,"msg":"adapting config using caddyfile: File to import not found: /etc/openpanel/caddy/botcontrol.conf, at /etc/caddy/Caddyfile:79"}` + "\n"
	if got := caddyErrorLine(jsonOut); got != "adapting config using caddyfile: File to import not found: /etc/openpanel/caddy/botcontrol.conf, at /etc/caddy/Caddyfile:79" {
		t.Fatalf("json: got %q", got)
	}
	if got := caddyErrorLine("Error: adapting config using caddyfile: boom\n"); got != "adapting config using caddyfile: boom" {
		t.Fatalf("plain: got %q", got)
	}
}

func TestAggregateLogReader(t *testing.T) {
	logs := strings.Join([]string{
		`{"ts":1000,"status":403,"request":{"client_ip":"203.0.113.10","uri":"/?s=test","headers":{"User-Agent":["GPTBot/1.0"]}}}`,
		`{"ts":1001,"status":403,"request":{"client_ip":"203.0.113.11","uri":"/wp-login.php","headers":{"User-Agent":["GPTBot/1.0"]}}}`,
		`{"ts":1002,"status":200,"request":{"client_ip":"198.51.100.5","uri":"/","headers":{"User-Agent":["Mozilla/5.0"]}}}`,
		`{"ts":500,"status":200,"request":{"client_ip":"198.51.100.9","uri":"/old","headers":{"User-Agent":["ClaudeBot/1.0"]}}}`,
	}, "\n")

	table := []asnPrefixEntry{
		{prefix: netip.MustParsePrefix("203.0.113.0/24"), asn: "AS12345 (OpenAI)"},
	}

	report, err := aggregateLogReader(strings.NewReader(logs), "example.com", "24h", 900, table)
	if err != nil {
		t.Fatalf("aggregateLogReader error: %v", err)
	}
	if report.TotalRequests != 3 || report.AIRequests != 2 || report.BlockedCount != 2 {
		t.Fatalf("unexpected totals: %+v", report)
	}
	if len(report.TopBots) == 0 || report.TopBots[0].Name != "GPTBot" || report.TopBots[0].Requests != 2 {
		t.Fatalf("unexpected TopBots: %+v", report.TopBots)
	}
	if len(report.TopNetworks) == 0 || report.TopNetworks[0].ASN != "AS12345 (OpenAI)" || report.TopNetworks[0].UniqueIPs != 2 {
		t.Fatalf("unexpected TopNetworks: %+v", report.TopNetworks)
	}
	if len(report.TopURLs) == 0 || report.TopURLs[0].Path != "/" {
		t.Fatalf("unexpected TopURLs: %+v", report.TopURLs)
	}
}

// a suspended domain's config and an empty one got the marker after the site block, block there broke every reload
func TestRewritePolicyDropsTopLevelLines(t *testing.T) {
	suspended := "old.example.org {\n  route {\n    file_server\n  }\n}\n    # botcontrol: allow\n"
	if _, err := rewritePolicy(suspended, "old.example.org", PolicyBlock); !errors.Is(err, errNoAnchor) {
		t.Fatalf("suspended: err = %v, want errNoAnchor", err)
	}
	if out, err := rewritePolicy(suspended, "old.example.org", PolicyAllow); err != nil || strings.Contains(out, "botcontrol") {
		t.Fatalf("allow should drop the stray marker, got %q, %v", out, err)
	}
	if _, err := rewritePolicy("\n    # botcontrol: allow\n", "x.org", PolicyBlock); !errors.Is(err, errNoAnchor) {
		t.Fatalf("empty: err = %v, want errNoAnchor", err)
	}
	waf := "https://a.com {\n  route {\n    coraza_waf {\n      directives `\n        SecRuleEngine Off\n      `\n    }\n    reverse_proxy x\n  }\n}\n# botcontrol: block\n"
	out, err := rewritePolicy(waf, "a.com", PolicyBlock)
	if err != nil || strings.Count(out, "import botcontrol_block") != 1 || strings.Contains(out, "# botcontrol") {
		t.Fatalf("waf + stray: %q, %v", out, err)
	}
}
