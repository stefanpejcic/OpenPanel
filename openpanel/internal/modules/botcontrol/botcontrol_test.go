package botcontrol

import (
	"net/netip"
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

	blocked := rewritePolicy(base, "example.com", PolicyBlock)
	if got := policyFromConfig(blocked); got != PolicyBlock {
		t.Fatalf("after block policy = %q, want block", got)
	}
	if strings.Count(blocked, "import botcontrol_block") != 2 {
		t.Fatalf("expected two botcontrol_block directives (HTTP + HTTPS), got:\n%s", blocked)
	}

	limited := rewritePolicy(blocked, "example.com", PolicyRateLimit)
	if got := policyFromConfig(limited); got != PolicyRateLimit {
		t.Fatalf("after ratelimit policy = %q, want ratelimit", got)
	}
	if strings.Count(limited, "import botcontrol_ratelimit example.com") != 2 {
		t.Fatalf("expected two ratelimit imports in:\n%s", limited)
	}

	allowed := rewritePolicy(limited, "example.com", PolicyAllow)
	if got := policyFromConfig(allowed); got != PolicyAllow {
		t.Fatalf("after allow policy = %q, want allow", got)
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
