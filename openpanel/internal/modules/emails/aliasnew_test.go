package emails

import (
	"strings"
	"testing"
)

func TestAliasProblem(t *testing.T) {
	existing := []AliasEntry{{Source: "Sales@b.com", Targets: []string{"x@y.com"}}}
	cases := []struct{ user, target, want string }{
		{"info", "me@gmail.com", ""},
		{"in fo", "me@gmail.com", "Username can only"},
		{strings.Repeat("a", 61), "me@gmail.com", "Username can only"},
		{"info", "not-an-email", "valid email"},
		{"info", "info@b.com", "itself"},
		{"info", "INFO@B.com", "itself"},
		{"sales", "me@gmail.com", "already exists"},
	}
	for _, c := range cases {
		got := aliasProblem(c.user, "b.com", c.target, existing)
		if c.want == "" && got != "" || c.want != "" && !strings.Contains(got, c.want) {
			t.Errorf("%q -> %q: got %q, want %q", c.user, c.target, got, c.want)
		}
	}
}

func TestAliasSheetURL(t *testing.T) {
	if got := aliasSheetURL("", ""); got != "/emails/aliases?new=1" {
		t.Errorf("empty: %q", got)
	}
	if got := aliasSheetURL("b.com", "a+b"); got != "/emails/aliases?domain=b.com&new=1&username=a%2Bb" {
		t.Errorf("filled: %q", got)
	}
}
