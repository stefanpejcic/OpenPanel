package waf

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http/httptest"
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

func TestEnableRuleForDomain(t *testing.T) {
	path, reloads := withTestDomain(t)
	if err := enableRuleForDomain(context.Background(), "example.com", "920350"); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "920350") || !strings.Contains(string(content), "SecRuleRemoveById 007") {
		t.Fatalf("rule not removed or sentinel lost:\n%s", content)
	}
	if err := enableRuleForDomain(context.Background(), "example.com", "920350"); err != nil || *reloads != 1 {
		t.Fatalf("second enable should be a no-op, err=%v reloads=%d", err, *reloads)
	}
	if err := enableRuleForDomain(context.Background(), "example.com", "007"); !errors.Is(err, errInvalidRuleID) {
		t.Fatalf("sentinel must not be removable, got %v", err)
	}
}

func TestRuleIDFromMultipartAndURLEncoded(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("rule_id", "930130")
	_ = mw.Close()
	r := httptest.NewRequest("POST", "/server/waf/disable-rule/example.com", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		_ = r.ParseForm()
	}
	if r.Form.Get("rule_id") != "930130" {
		t.Fatal("multipart rule_id not read")
	}
	r = httptest.NewRequest("POST", "/server/waf/disable-rule/example.com", strings.NewReader("rule_id=1"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		_ = r.ParseForm()
	}
	if r.Form.Get("rule_id") != "1" {
		t.Fatal("urlencoded rule_id not read")
	}
}
