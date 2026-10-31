package ftp

import (
	"strings"
	"testing"
)

func TestNewAccountProblem(t *testing.T) {
	existing := []Account{{Username: "deploy@example.com"}}
	good := "Str0ng!" + "Tw0Value"
	cases := []struct {
		local, pass, want string
	}{
		{"backup_1", good, ""},
		{"Deploy", good, "lower case"},
		{"a.b", good, "lower case"},
		{strings.Repeat("a", 33), good, "lower case"},
		{"deploy", good, "already exists"},
		{"other", "Ab1!", "between 8 and 30"},
		{"other", "Ab1!" + strings.Repeat("x", 27), "between 8 and 30"},
		{"other", "abcdefgh1234!", "upper case letter"},
		{"other", "Abcdefgh1234", "upper case letter"},
	}
	for _, c := range cases {
		got := newAccountProblem(c.local, "example.com", c.pass, 50, existing)
		if c.want == "" && got != "" || c.want != "" && !strings.Contains(got, c.want) {
			t.Errorf("%q/%q: got %q, want %q", c.local, c.pass, got, c.want)
		}
	}
	if got := newAccountProblem("other", "example.com", "Abcdef1!", 100, nil); !strings.Contains(got, "strength") {
		t.Errorf("threshold: got %q", got)
	}
}

func TestNewAccountBack(t *testing.T) {
	if got := newAccountBack("", "", ""); got != "/ftp/new" {
		t.Errorf("empty: %q", got)
	}
	if got := newAccountBack("deploy", "example.com", "/var/www/html/a b"); got != "/ftp/new?domain=example.com&username=deploy&path=%2Fvar%2Fwww%2Fhtml%2Fa+b" {
		t.Errorf("filled: %q", got)
	}
}
