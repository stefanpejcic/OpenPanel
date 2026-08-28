package appinstall

import "testing"

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
