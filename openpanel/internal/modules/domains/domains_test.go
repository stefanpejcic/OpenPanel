package domains

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPageEntries(t *testing.T) {
	entries := buildPageEntries(5, 10)
	var numbers []int
	ellipses := 0
	for _, e := range entries {
		if e.IsEllipsis {
			ellipses++
		} else {
			numbers = append(numbers, e.Number)
		}
	}
	// 1, ellipsis, 3,4,5,6,7, ellipsis, 10
	want := []int{1, 3, 4, 5, 6, 7, 10}
	if len(numbers) != len(want) {
		t.Fatalf("expected numbers %v, got %v", want, numbers)
	}
	for i, n := range want {
		if numbers[i] != n {
			t.Errorf("position %d: expected %d, got %d", i, n, numbers[i])
		}
	}
	if ellipses != 2 {
		t.Errorf("expected 2 ellipses, got %d", ellipses)
	}
}

func TestBuildLogPageEntries(t *testing.T) {
	t.Run("small total, no ellipsis needed", func(t *testing.T) {
		entries := buildLogPageEntries(1, 3)
		for _, e := range entries {
			if e.IsEllipsis {
				t.Errorf("did not expect an ellipsis for a 3-page total, got %+v", entries)
			}
		}
	})

	t.Run("large total, both ellipses", func(t *testing.T) {
		entries := buildLogPageEntries(10, 20)
		if entries[0].Number != 1 {
			t.Errorf("expected first entry to be page 1, got %+v", entries[0])
		}
		if !entries[1].IsEllipsis {
			t.Errorf("expected an ellipsis after page 1, got %+v", entries[1])
		}
		last := entries[len(entries)-1]
		if last.Number != 20 {
			t.Errorf("expected last entry to be page 20, got %+v", last)
		}
	})
}

func TestResolveUnderVarWWWHTML(t *testing.T) {
	cases := []struct {
		in       string
		wantOK   bool
		wantPath string
	}{
		{"/var/www/html/", true, "/var/www/html"},
		{"/var/www/html/example.com", true, "/var/www/html/example.com"},
		{"/etc/passwd", false, "/etc/passwd"},
		// filepath.Clean resolves ".." lexically: html/../.. only pops back
		// to /var, landing on /var/etc/passwd (not /etc/passwd) - still
		// correctly rejected as outside /var/www/html/.
		{"/var/www/html/../../etc/passwd", false, "/var/etc/passwd"},
	}
	for _, c := range cases {
		got, ok := resolveUnderVarWWWHTML(c.in)
		if ok != c.wantOK {
			t.Errorf("resolveUnderVarWWWHTML(%q) ok = %v, want %v", c.in, ok, c.wantOK)
		}
		if got != c.wantPath {
			t.Errorf("resolveUnderVarWWWHTML(%q) = %q, want %q", c.in, got, c.wantPath)
		}
	}
}

func TestInsertOrReplaceRedirect(t *testing.T) {
	t.Run("insert before log block", func(t *testing.T) {
		content := "example.com {\n    root * /srv\n    log {\n        output file /var/log/x.log\n    }\n}\n"
		got := insertOrReplaceRedirect(content, "https://target.example")
		if !containsLine(got, "redir https://target.example") {
			t.Errorf("expected inserted redir line, got:\n%s", got)
		}
	})

	t.Run("replace existing redir before anchor", func(t *testing.T) {
		content := "example.com {\n    redir https://old.example\n    import domain_log\n}\n"
		got := insertOrReplaceRedirect(content, "https://new.example")
		if containsLine(got, "redir https://old.example") {
			t.Errorf("expected old redir line replaced, got:\n%s", got)
		}
		if !containsLine(got, "redir https://new.example") {
			t.Errorf("expected new redir line present, got:\n%s", got)
		}
	})
}

func containsLine(content, substr string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == substr {
			return true
		}
	}
	return false
}

func TestCapitalizedDomainsFile(t *testing.T) {
	got := capitalizedDomainsFile("alice")
	want := "/home/alice/capitalized_domains.json"
	if got != want {
		t.Errorf("capitalizedDomainsFile(%q) = %q, want %q", "alice", got, want)
	}
}

func TestVhostFilePath(t *testing.T) {
	got := vhostFilePath("alice", "example.com")
	want := "/home/alice/docker-data/volumes/alice_webserver_data/_data/example.com.conf"
	if got != want {
		t.Errorf("vhostFilePath = %q, want %q", got, want)
	}
}

func TestReadWriteTextFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.conf")
	if err := writeTextFile(path, "hello world"); err != nil {
		t.Fatalf("writeTextFile: %v", err)
	}
	got, err := readTextFile(path)
	if err != nil {
		t.Fatalf("readTextFile: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestSplitMax(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want []string
	}{
		{"redir https://x.com permanent", 3, []string{"redir", "https://x.com", "permanent"}},
		{"redir https://x.com", 3, []string{"redir", "https://x.com"}},
		{"", 3, nil},
	}
	for _, c := range cases {
		got := splitMax(c.in, c.max)
		if len(got) != len(c.want) {
			t.Errorf("splitMax(%q, %d) = %v, want %v", c.in, c.max, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitMax(%q, %d)[%d] = %q, want %q", c.in, c.max, i, got[i], c.want[i])
			}
		}
	}
}

func TestGetRedirectURLFromRealFile(t *testing.T) {
	// getRedirectURL() hardcodes /etc/openpanel/caddy/domains/, which isn't
	// writable in a sandboxed test environment, so this just confirms the
	// not-found path returns "" rather than panicking.
	if got := getRedirectURL("no-such-domain-ever.example"); got != "" {
		t.Errorf("expected empty string for a nonexistent domain conf, got %q", got)
	}
}
