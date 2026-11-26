package appinstall

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoWith(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestDetectStack(t *testing.T) {
	cases := []struct {
		name      string
		files     map[string]string
		appType   string
		framework string
		cmd       string
		port      int
		install   bool
	}{
		{"next", map[string]string{"package.json": `{"scripts":{"build":"next build","start":"next start"},"dependencies":{"next":"15"},"engines":{"node":">=20.9"}}`},
			"nodejs", "next", buildOnce("npm run build") + " && HOST=0.0.0.0 PORT=3000 npm start", 3000, true},
		{"express start", map[string]string{"package.json": `{"scripts":{"start":"node server.js"},"dependencies":{"express":"4"}}`},
			"nodejs", "express", "HOST=0.0.0.0 PORT=3000 npm start", 3000, true},
		{"vite static", map[string]string{"package.json": `{"scripts":{"build":"vite build","dev":"vite"},"devDependencies":{"vite":"5"}}`},
			"nodejs", "vite", buildOnce("npm run build") + " && npx -y serve -s dist -l 3000", 3000, true},
		{"cra static", map[string]string{"package.json": `{"scripts":{"build":"react-scripts build"},"dependencies":{"react-scripts":"5"}}`},
			"nodejs", "create-react-app", buildOnce("npm run build") + " && npx -y serve -s build -l 3000", 3000, true},
		{"plain node", map[string]string{"package.json": `{"main":"app.js"}`, "app.js": ""},
			"nodejs", "node", "HOST=0.0.0.0 PORT=3000 node app.js", 3000, true},
		{"static html", map[string]string{"index.html": "<h1>hi</h1>"},
			"nodejs", "static", "npx -y serve -l 3000 .", 3000, false},
		{"django", map[string]string{"manage.py": "", "requirements.txt": "django\n"},
			"python", "django", "python manage.py migrate --noinput && python manage.py runserver 0.0.0.0:8000", 8000, true},
		{"fastapi", map[string]string{"main.py": "from fastapi import FastAPI\napi = FastAPI()\n", "requirements.txt": "fastapi\nuvicorn\n"},
			"python", "fastapi", "uvicorn main:api --host 0.0.0.0 --port 8000", 8000, true},
		{"fastapi without uvicorn", map[string]string{"app.py": "app = FastAPI()\n", "requirements.txt": "fastapi\n"},
			"python", "fastapi", "pip install uvicorn && uvicorn app:app --host 0.0.0.0 --port 8000", 8000, true},
		{"flask pyproject", map[string]string{"app.py": "app = Flask(__name__)\n", "pyproject.toml": "[project]\ndependencies=['flask']\n"},
			"python", "flask", "pip install . && flask --app app:app run --host 0.0.0.0 --port 8000", 8000, false},
		{"rails", map[string]string{"Gemfile": "gem 'rails'", "bin/rails": ""},
			"ruby", "rails", "bundle exec rails server -b 0.0.0.0 -p 3000", 3000, true},
		{"rack", map[string]string{"Gemfile": "gem 'sinatra'", "config.ru": "run App"},
			"ruby", "rack", "bundle exec rackup -o 0.0.0.0 -p 3000", 3000, true},
		{"spring", map[string]string{"pom.xml": "<artifactId>spring-boot-starter-web</artifactId>"},
			"java", "spring-boot", "mvn -q -DskipTests spring-boot:run -Dspring-boot.run.arguments=--server.port=8080", 8080, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan, err := detectStackInDir(repoWith(t, c.files))
			if err != nil {
				t.Fatalf("detect: %v", err)
			}
			if plan.AppType != c.appType || plan.Framework != c.framework || plan.StartCmd != c.cmd || plan.Port != c.port || plan.Install != c.install {
				t.Errorf("got %+v", plan)
			}
			// whatever detection writes has to pass the same check as a caller's start_cmd
			if msg := applyOverrides(&StackPlan{}, deployRequest{StartCmd: plan.StartCmd}); msg != "" {
				t.Errorf("detected command %q fails the start_cmd check: %s", plan.StartCmd, msg)
			}
		})
	}
}

func TestDetectStackNodeVersionFromEngines(t *testing.T) {
	plan, err := detectStackInDir(repoWith(t, map[string]string{"package.json": `{"scripts":{"start":"x"},"engines":{"node":"^22.1.0"}}`}))
	if err != nil || plan.Version != "22" {
		t.Fatalf("version = %v, %v", plan, err)
	}
	if !isValidVersion(plan.Version) {
		t.Error("detected version isn't accepted by the installer")
	}
}

func TestDetectStackUnsupported(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"dockerfile only": {"Dockerfile": "FROM nginx"},
		"empty":           {"README.md": "hi"},
		"node no entry":   {"package.json": `{"name":"x"}`},
	} {
		if _, err := detectStackInDir(repoWith(t, files)); !errors.Is(err, ErrUnsupportedStack) {
			t.Errorf("%s: err = %v, want ErrUnsupportedStack", name, err)
		}
	}
	if _, err := detectStackInDir(repoWith(t, map[string]string{"package.json": "{"})); err == nil || errors.Is(err, ErrUnsupportedStack) {
		t.Errorf("broken package.json: err = %v", err)
	}
}

func TestApplyOverrides(t *testing.T) {
	bad := []deployRequest{
		{AppType: "php"},
		{StartCmd: `node -e "x"`},
		{StartCmd: "node a.js\nrm -rf /"},
		{StartCmd: "cat ../../etc/passwd"},
		{StartCmd: "echo `id`"},
		{StartCmd: `a\b`},
		{Port: 70000},
		{Version: "20-slim; id"},
	}
	for _, req := range bad {
		if msg := applyOverrides(&StackPlan{}, req); msg == "" {
			t.Errorf("accepted %+v", req)
		}
	}
	no := false
	plan := &StackPlan{AppType: "nodejs", Port: 3000, Install: true, Version: "latest"}
	if msg := applyOverrides(plan, deployRequest{AppType: "python", StartCmd: "gunicorn app:app -b 0.0.0.0:9000", Port: 9000, Version: "3.12", Install: &no}); msg != "" {
		t.Fatal(msg)
	}
	if plan.AppType != "python" || plan.Port != 9000 || plan.Version != "3.12" || plan.Install || !strings.HasPrefix(plan.StartCmd, "gunicorn") {
		t.Errorf("overrides not applied: %+v", plan)
	}
}

func TestDefaultServiceName(t *testing.T) {
	cases := map[[2]string]string{
		{"example.com", ""}:                        "example",
		{"shop.example.com", "blog"}:               "shopblog",
		{"123.example.com", ""}:                    "app123",
		{"a-really-long-domain-name-here.com", ""}: "areallylongdomainnamehere"[:24],
	}
	for in, want := range cases {
		got := defaultServiceName(in[0], in[1])
		if got != want || !deployServiceRE.MatchString(got) {
			t.Errorf("defaultServiceName(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

func TestParseInstallStream(t *testing.T) {
	ok := parseInstallStream([]byte("{\"status\":\"Validating\"}\n{\"status\":\"Service added successfully\",\"success\":true}\n"), 200)
	if !ok.OK || len(ok.Steps) != 2 {
		t.Errorf("success stream = %+v", ok)
	}
	failed := parseInstallStream([]byte("{\"status\":\"Validating\"}\n{\"error\":\"Service 'x' already exists\"}\n"), 200)
	if failed.OK || !strings.Contains(failed.Error, "already exists") {
		t.Errorf("error stream = %+v", failed)
	}
	denied := parseInstallStream([]byte(`{"error":"Access denied"}`), 403)
	if denied.OK || denied.Error != "Access denied" {
		t.Errorf("403 = %+v", denied)
	}
	var long strings.Builder
	for i := 0; i < 30; i++ {
		long.WriteString("{\"status\":\"step\"}\n")
	}
	if got := parseInstallStream([]byte(long.String()), 200); len(got.Steps) != deployStepsKept {
		t.Errorf("kept %d steps", len(got.Steps))
	}
}

func TestProbeApp(t *testing.T) {
	oldRun, oldState := probeRunner, containerStateFn
	defer func() { probeRunner, containerStateFn = oldRun, oldState }()

	state := "running"
	out := "200\n"
	var gotArgv []string
	containerStateFn = func(context.Context, string, string) string { return state }
	probeRunner = func(_ context.Context, _, _ string, argv []string) (string, error) {
		gotArgv = argv
		return out, nil
	}

	h := probeApp(context.Background(), "u", "shop", "nodejs", 3000)
	if !h.Listening || h.HTTPStatus != 200 || gotArgv[0] != "node" || !strings.Contains(gotArgv[2], "127.0.0.1:3000") {
		t.Errorf("node probe: %+v argv=%v", h, gotArgv)
	}
	out = "npm warn something\n404\n"
	if h := probeApp(context.Background(), "u", "shop", "python", 8000); !h.Listening || h.HTTPStatus != 404 || gotArgv[0] != "python" {
		t.Errorf("python probe with noise: %+v", h)
	}
	out = "ERR ECONNREFUSED\n"
	if h := probeApp(context.Background(), "u", "shop", "ruby", 3000); h.Listening || !strings.Contains(h.ProbeError, "3000") {
		t.Errorf("refused probe: %+v", h)
	}
	if h := probeApp(context.Background(), "u", "shop", "java", 8080); h.Listening || h.ProbeError == "" {
		t.Errorf("java has no probe: %+v", h)
	}
	state = "exited"
	gotArgv = nil
	if h := probeApp(context.Background(), "u", "shop", "nodejs", 3000); h.Container != "exited" || gotArgv != nil {
		t.Errorf("stopped container was probed: %+v", h)
	}
}

func TestWaitForAppStopsWhenContainerExits(t *testing.T) {
	oldRun, oldState := probeRunner, containerStateFn
	defer func() { probeRunner, containerStateFn = oldRun, oldState }()
	containerStateFn = func(context.Context, string, string) string { return "exited" }
	probeRunner = func(context.Context, string, string, []string) (string, error) { return "", nil }
	if h := waitForApp(context.Background(), "u", "x", "nodejs", 3000, 0); h.Container != "exited" {
		t.Errorf("got %+v", h)
	}
}

func TestBuildOnceIsSafeForComposeAndEnv(t *testing.T) {
	cmd := buildOnce("npm run build")
	// compose interpolates $ and the command sits inside double quotes in both compose and .env
	if strings.ContainsAny(cmd, "$\"\\`\n") || strings.Contains(cmd, "..") {
		t.Fatalf("unsafe characters in %q", cmd)
	}
	if !isValidCustomCommand(cmd) {
		t.Fatal("the installer would reject it")
	}
	if strings.Count(cmd, "(") != strings.Count(cmd, ")") {
		t.Fatal("unbalanced parentheses")
	}
}
