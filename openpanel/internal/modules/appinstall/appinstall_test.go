package appinstall

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComposeServiceIndent(t *testing.T) {
	cases := map[string]struct {
		lines []string
		want  int
	}{
		"fresh vendored file, 2-space services": {
			lines: []string{"\n", "  openlitespeed:\n", "    image: x\n"},
			want:  2,
		},
		"post-SaveCompose file, 4-space services": {
			lines: []string{"    openlitespeed:\n", "        image: x\n"},
			want:  4,
		},
		"blank lines before the first service key are skipped": {
			lines: []string{"\n", "\n", "  openlitespeed:\n"},
			want:  2,
		},
		"no service lines falls back to the vendored default": {
			lines: []string{"\n"},
			want:  vendoredComposeServiceIndent,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := composeServiceIndent(tc.lines); got != tc.want {
				t.Errorf("composeServiceIndent(%q) = %d, want %d", tc.lines, got, tc.want)
			}
		})
	}
}

func TestIndentComposeService(t *testing.T) {
	template := "  myapp:\n    image: node\n    ports:\n      - 3000\n"

	t.Run("matches a fresh 2-space file: no extra indent added", func(t *testing.T) {
		got := indentComposeService(template, 2)
		if got != template {
			t.Errorf("got:\n%s\nwant unchanged:\n%s", got, template)
		}
	})

	t.Run("matches a post-SaveCompose 4-space file: shifted by 2", func(t *testing.T) {
		got := indentComposeService(template, 4)
		want := "    myapp:\n      image: node\n      ports:\n        - 3000\n"
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})
}

// TestComposeServiceInsertProducesValidYAML reproduces the live bug: a
// vendored, never-saved compose file (2-space service keys) got a new
// service unconditionally pushed to 4 spaces, so the inserted service sat
// deeper than its siblings under services: - invalid YAML ("expected
// <block end>, but found '<block mapping start>'"), which is exactly what
// broke every fresh Python/NodeJS/Ruby install. Assert the merged result
// actually parses and both services show up as siblings.
func TestComposeServiceInsertProducesValidYAML(t *testing.T) {
	existing := "services:\n\n  openlitespeed:\n    image: openlitespeed\n    networks:\n      - www\n"
	composeLines := strings.SplitAfter(existing, "\n")

	insertPosition := -1
	for i, line := range composeLines {
		if strings.HasPrefix(line, "services:") {
			insertPosition = i + 1
			break
		}
	}
	if insertPosition == -1 {
		t.Fatal("services: not found in fixture")
	}

	newService := "  myapp:\n    image: node:25.9.0\n    ports:\n      - \"3000\"\n"
	indented := indentComposeService(newService, composeServiceIndent(composeLines[insertPosition:]))

	newLines := make([]string, 0, len(composeLines)+1)
	newLines = append(newLines, composeLines[:insertPosition]...)
	newLines = append(newLines, "\n"+indented+"\n")
	newLines = append(newLines, composeLines[insertPosition:]...)
	merged := strings.Join(newLines, "")

	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(merged), &parsed); err != nil {
		t.Fatalf("merged compose file is not valid YAML: %v\n%s", err, merged)
	}
	services, _ := parsed["services"].(map[string]any)
	if _, ok := services["openlitespeed"]; !ok {
		t.Errorf("expected existing service preserved, got services: %v", services)
	}
	if _, ok := services["myapp"]; !ok {
		t.Errorf("expected new service inserted as a sibling, got services: %v", services)
	}
}

func TestIsValidServiceName(t *testing.T) {
	cases := map[string]bool{
		"myapp": true, "My_App-1": true, "": false, "with space": false, "slash/here": false,
	}
	for in, want := range cases {
		if got := isValidServiceName(in); got != want {
			t.Errorf("isValidServiceName(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsValidSubdirectory(t *testing.T) {
	cases := map[string]bool{
		"": true, "app": true, "sub/dir": true, "../etc": false, "/abs": false,
	}
	for in, want := range cases {
		if got := isValidSubdirectory(in); got != want {
			t.Errorf("isValidSubdirectory(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsValidVersion(t *testing.T) {
	cases := map[string]bool{
		"3.10": true, "18": true, "1.2.3": true, "latest": false, "3.x": false, "": false,
	}
	for in, want := range cases {
		if got := isValidVersion(in); got != want {
			t.Errorf("isValidVersion(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsValidStartupFile(t *testing.T) {
	cases := map[string]bool{
		"/var/www/html/app.py":        true,
		"/var/www/html/sub/index.js":  true,
		"/var/www/html/app.exe":       false,
		"/etc/passwd":                 false,
		"/var/www/html/../etc/app.py": false,
	}
	for in, want := range cases {
		if got := isValidStartupFile(in); got != want {
			t.Errorf("isValidStartupFile(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsValidCustomCommand(t *testing.T) {
	cases := map[string]bool{
		"gunicorn app:app": true, "../evil": false, "line1\nline2": false,
	}
	for in, want := range cases {
		if got := isValidCustomCommand(in); got != want {
			t.Errorf("isValidCustomCommand(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestGetValidatedFloat(t *testing.T) {
	if v := getValidatedFloat("2.5", "1.0"); v != 2.5 {
		t.Errorf("got %v, want 2.5", v)
	}
	if v := getValidatedFloat("-1", "1.0"); v != 1.0 {
		t.Errorf("got %v, want 1.0 (default on invalid)", v)
	}
	if v := getValidatedFloat("not-a-number", "0.5"); v != 0.5 {
		t.Errorf("got %v, want 0.5 (default on invalid)", v)
	}
}

func TestGetValidatedInt(t *testing.T) {
	if v := getValidatedInt("250", "100"); v != 250 {
		t.Errorf("got %v, want 250", v)
	}
	if v := getValidatedInt("-1", "100"); v != 100 {
		t.Errorf("got %v, want 100 (default on invalid)", v)
	}
	if v := getValidatedInt("not-a-number", "50"); v != 50 {
		t.Errorf("got %v, want 50 (default on invalid)", v)
	}
}

func TestBuildAppRunCommand(t *testing.T) {
	cases := []struct {
		kind                                               Kind
		requirements, customCmd, startupFile, gitURL, want string
	}{
		{Python, "1", "", "", "", "pip install -r requirements.txt && python app.py"},
		{Python, "", "", "app/main.py", "", "python app/main.py"},
		{Python, "", "gunicorn app:app", "main.py", "", "gunicorn app:app"},
		{NodeJS, "1", "", "", "", "npm install && node index.js"},
		{NodeJS, "", "", "server.js", "", "node server.js"},
		{
			NodeJS, "1", "", "", "https://github.com/user/repo.git",
			"(command -v git >/dev/null 2>&1 || (apt-get update -qq && apt-get install -y -qq git)) && " +
				"(git rev-parse --is-inside-work-tree >/dev/null 2>&1 || git init -q) && " +
				"(git remote get-url origin >/dev/null 2>&1 || git remote add origin 'https://github.com/user/repo.git') && " +
				"git fetch --depth 1 origin HEAD && git reset --hard FETCH_HEAD && " +
				"npm install && node index.js",
		},
		{Ruby, "1", "", "", "", "bundle install && ruby app.rb"},
		{Ruby, "", "", "server.rb", "", "ruby server.rb"},
		{Java, "1", "", "", "", "mvn install && java Main.java"},
		{Java, "", "", "App.java", "", "java App.java"},
	}
	for _, c := range cases {
		if got := buildAppRunCommand(c.kind, c.requirements, c.customCmd, c.startupFile, c.gitURL); got != c.want {
			t.Errorf("buildAppRunCommand(%q,%q,%q,%q,%q) = %q, want %q", c.kind.PyOrNode, c.requirements, c.customCmd, c.startupFile, c.gitURL, got, c.want)
		}
	}
}

func TestIsValidGitURL(t *testing.T) {
	cases := map[string]bool{
		"":                                    true,
		"https://github.com/user/repo.git":    true,
		"https://token@github.com/user/r.git": true,
		"http://github.com/user/repo.git":     false,
		"git@github.com:user/repo.git":        false,
		"https://evil.com/x'; rm -rf /":       false,
		"https://has space.com/repo.git":      false,
	}
	for in, want := range cases {
		if got := isValidGitURL(in); got != want {
			t.Errorf("isValidGitURL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNormalizeRequirements(t *testing.T) {
	cases := map[string]string{
		"1": "1", "on": "1", "true": "1", "yes": "1", "TRUE": "1",
		"": "", "0": "", "off": "", "no": "",
	}
	for in, want := range cases {
		if got := normalizeRequirements(in); got != want {
			t.Errorf("normalizeRequirements(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatPyFloat(t *testing.T) {
	cases := map[float64]string{1.0: "1.0", 0.5: "0.5", 2.25: "2.25"}
	for in, want := range cases {
		if got := formatPyFloat(in); got != want {
			t.Errorf("formatPyFloat(%v) = %q, want %q", in, got, want)
		}
	}
}
