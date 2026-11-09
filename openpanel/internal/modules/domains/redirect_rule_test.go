package domains

import (
	"strings"
	"testing"
)

const confBlock = "http://example.com, http://www.example.com {\n  import domain_log example.com\n  route {\n  }\n}\nhttps://example.com, https://www.example.com {\n  import domain_log example.com\n}"

func TestRedirectRuleLinesRoundTrip(t *testing.T) {
	cases := []struct {
		rule redirectRule
		want []string
	}{
		{redirectRule{To: "https://new.example"}, []string{"redir https://new.example"}},
		{redirectRule{To: "https://new.example/", Permanent: true, KeepPath: true}, []string{"redir https://new.example{uri} 301"}},
		{redirectRule{To: "https://new.example/blog", KeepPath: true, BareOnly: true}, []string{"@op_redir_bare host example.com", "redir @op_redir_bare https://new.example/blog{uri}"}},
	}
	for _, c := range cases {
		got := c.rule.lines("example.com")
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("%+v: got %q, want %q", c.rule, got, c.want)
			continue
		}
		back, ok := parseRedirectLine("    " + got[len(got)-1])
		want := c.rule
		want.To = strings.TrimRight(want.To, "/")
		if !c.rule.KeepPath {
			want.To = c.rule.To
		}
		if !ok || back != want {
			t.Errorf("%+v read back as %+v (ok=%v)", c.rule, back, ok)
		}
	}
	for _, line := range []string{"# redirects", "import /etc/openpanel/caddy/redirects.conf", "redir", "redir @op_redir_bare"} {
		if _, ok := parseRedirectLine(line); ok {
			t.Errorf("%q parsed as a redirect", line)
		}
	}
}

func TestInsertReplaceRemoveRedirectRule(t *testing.T) {
	bare := redirectRule{To: "https://new.example", Permanent: true, KeepPath: true, BareOnly: true}
	added := insertOrReplaceRedirect(confBlock, bare, "example.com")
	if n := strings.Count(added, "@op_redir_bare host example.com"); n != 2 {
		t.Fatalf("expected the matcher in both site blocks, got %d:\n%s", n, added)
	}
	if n := strings.Count(added, "redir @op_redir_bare https://new.example{uri} 301"); n != 2 {
		t.Fatalf("expected the redir line in both site blocks, got %d:\n%s", n, added)
	}

	// switching to a plain redirect replaces both lines, nothing is left behind
	plain := insertOrReplaceRedirect(added, redirectRule{To: "https://other.example"}, "example.com")
	if strings.Contains(plain, bareMatcher) || strings.Contains(plain, "new.example") {
		t.Fatalf("old rule left behind:\n%s", plain)
	}
	if n := strings.Count(plain, "redir https://other.example\n"); n != 2 {
		t.Fatalf("expected the plain redirect twice:\n%s", plain)
	}

	if got := removeRedirect(added, "https://new.example"); got != confBlock {
		t.Errorf("add then remove changed the conf:\n%s", got)
	}
}
