package waf

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestTagREBlocksCaddyfileBreakout(t *testing.T) {
	for _, bad := range []string{"x\"\n`\n}\nroot * /", "`", "a\"b", "a b", "a{b}", "a\\b", ""} {
		if tagRE.MatchString(bad) {
			t.Errorf("accepted %q", bad)
		}
	}
	for _, ok := range []string{"attack-sqli", "paranoia-level/1", "capec/1000/153/267", "OWASP_CRS"} {
		if !tagRE.MatchString(ok) {
			t.Errorf("rejected %q", ok)
		}
	}
	out := rewriteDirectivesBlock(testDomainConf, []string{"007"}, []string{"example", "x\"\n`\nroot * /"})
	if strings.Contains(out, "root * /") {
		t.Fatalf("invalid tag was written:\n%s", out)
	}
}

func TestRejectedReloadRestoresConf(t *testing.T) {
	path, _ := withTestDomain(t)
	before, _ := os.ReadFile(path)
	reloadCaddy = func(context.Context) error { return errors.New("caddy said no") }
	if _, err := disableRuleForDomain(context.Background(), "example.com", "942100"); !errors.Is(err, errCaddyReload) {
		t.Fatalf("err = %v, want errCaddyReload", err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatalf("conf was not restored:\n%s", after)
	}
}
