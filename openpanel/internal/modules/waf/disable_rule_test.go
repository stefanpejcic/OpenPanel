package waf

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testDomainConf = `https://example.com {
    coraza_waf {
        directives ` + "`" + `
            Include /etc/openpanel/caddy/coraza_rules.conf
            SecRuleEngine On
            SecRuleRemoveById 007 920350
            SecRuleRemoveByTag example
            SecAuditLog /var/log/caddy/coraza_waf/example.com.log
        ` + "`" + `
    }
}
`

func withTestDomain(t *testing.T) (path string, reloads *int) {
	t.Helper()
	dir := t.TempDir()
	origDir, origReload := caddyDomainsDir, reloadCaddy
	caddyDomainsDir = dir + "/"
	n := 0
	reloadCaddy = func(context.Context) error { n++; return nil }
	t.Cleanup(func() { caddyDomainsDir, reloadCaddy = origDir, origReload })
	path = filepath.Join(dir, "example.com.conf")
	if err := os.WriteFile(path, []byte(testDomainConf), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, &n
}

func TestDisableRuleForDomainAppends(t *testing.T) {
	path, reloads := withTestDomain(t)
	already, err := disableRuleForDomain(context.Background(), "example.com", "942100")
	if err != nil || already {
		t.Fatalf("unexpected result already=%v err=%v", already, err)
	}
	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "SecRuleRemoveById 007 920350 942100") {
		t.Fatalf("rule not appended:\n%s", content)
	}
	if !strings.Contains(string(content), `SecRuleRemoveByTag "example"`) {
		t.Fatalf("tags lost:\n%s", content)
	}
	if *reloads != 1 {
		t.Fatalf("expected one caddy reload, got %d", *reloads)
	}
}

func TestDisableRuleForDomainAlreadyDisabled(t *testing.T) {
	_, reloads := withTestDomain(t)
	already, err := disableRuleForDomain(context.Background(), "example.com", "920350")
	if err != nil || !already {
		t.Fatalf("expected already disabled, got already=%v err=%v", already, err)
	}
	if *reloads != 0 {
		t.Fatalf("no reload expected for a no-op, got %d", *reloads)
	}
}

func TestDisableRuleForDomainRefusesScoringAndBadIDs(t *testing.T) {
	path, _ := withTestDomain(t)
	for _, id := range []string{"949110", "959100", "980170"} {
		if _, err := disableRuleForDomain(context.Background(), "example.com", id); !errors.Is(err, errScoringRule) {
			t.Fatalf("%s: expected errScoringRule, got %v", id, err)
		}
	}
	for _, id := range []string{"", "abc", "942100 949110", "007"} {
		if _, err := disableRuleForDomain(context.Background(), "example.com", id); !errors.Is(err, errInvalidRuleID) {
			t.Fatalf("%q: expected errInvalidRuleID, got %v", id, err)
		}
	}
	content, _ := os.ReadFile(path)
	if string(content) != testDomainConf {
		t.Fatalf("config changed by refused requests:\n%s", content)
	}
}
