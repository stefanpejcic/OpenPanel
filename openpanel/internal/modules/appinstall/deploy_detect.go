package appinstall

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// StackPlan is what deploy detection decided for a repo, every field can be overridden by the caller
type StackPlan struct {
	AppType     string   `json:"app_type"`
	Framework   string   `json:"framework"`
	StartupFile string   `json:"startup_file,omitempty"`
	StartCmd    string   `json:"start_cmd,omitempty"`
	Install     bool     `json:"install_dependencies"`
	Port        int      `json:"port"`
	Version     string   `json:"version"`
	Notes       []string `json:"notes,omitempty"`
}

// ErrUnsupportedStack means the repo has nothing a runtime installer can run
var ErrUnsupportedStack = errors.New("unsupported stack")

// these ports go into the start command as PORT and into the reverse proxy, so the two always match
const (
	deployNodePort   = 3000
	deployPythonPort = 8000
	deployRubyPort   = 3000
	deployJavaPort   = 8080
)

var nodeMajorRE = regexp.MustCompile(`(\d{2})`)

// buildOnce runs build only when the checked-out commit differs from the last one that built, so a restart without new
// commits starts in seconds; no $ or quotes since it lands inside sh -c "..." in compose and in .env, and a repo without
// git (or a failed rev-parse) simply builds every time like before
func buildOnce(build string) string {
	return "((git rev-parse HEAD 2>/dev/null > .git/op-head && cmp -s .git/op-head .git/op-built) || (" + build +
		" && (cp .git/op-head .git/op-built 2>/dev/null || true)))"
}

// cloneForDetect shallow-clones a public https repo into a temp dir the caller removes
func cloneForDetect(ctx context.Context, gitURL string) (string, error) {
	if err := resolvePublicGitHost(gitURL); err != nil {
		return "", err
	}
	tmpDir, err := os.MkdirTemp("", "opdeploy-*")
	if err != nil {
		return "", err
	}
	cloneCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err := git.PlainCloneContext(cloneCtx, tmpDir, false, &git.CloneOptions{URL: gitURL, Depth: 1, SingleBranch: true}); err != nil {
		_ = tenantfs.RemoveAll(tmpDir)
		return "", err
	}
	return tmpDir, nil
}

// DetectStack clones gitURL and works out how to run it
func DetectStack(ctx context.Context, gitURL string) (*StackPlan, error) {
	dir, err := cloneForDetect(ctx, gitURL)
	if err != nil {
		return nil, err
	}
	defer tenantfs.RemoveAll(dir)
	return detectStackInDir(dir)
}

func has(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func readSmall(dir, name string) string {
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := make([]byte, 256*1024)
	n, _ := f.Read(buf)
	return string(buf[:n])
}

// detectStackInDir looks at the files at the repo root, first match wins
func detectStackInDir(dir string) (*StackPlan, error) {
	switch {
	case has(dir, "package.json"):
		return detectNode(dir)
	case has(dir, "manage.py") || has(dir, "requirements.txt") || has(dir, "pyproject.toml"):
		return detectPython(dir), nil
	case has(dir, "Gemfile"):
		return detectRuby(dir), nil
	case has(dir, "pom.xml"):
		return detectJava(dir), nil
	case has(dir, "index.html"):
		return &StackPlan{
			AppType: "nodejs", Framework: "static", Port: deployNodePort, Version: "latest",
			StartCmd: "npx -y serve -l 3000 .",
			Notes:    []string{"No app manifest, serving the files as a static site."},
		}, nil
	case has(dir, "Dockerfile") || has(dir, "docker-compose.yml") || has(dir, "compose.yml"):
		return nil, errors.Join(ErrUnsupportedStack, errors.New("only a Dockerfile or compose file was found, building images isn't supported by deploy yet, add the service through the Containers API instead"))
	}
	return nil, errors.Join(ErrUnsupportedStack, errors.New("no package.json, requirements.txt, pyproject.toml, manage.py, Gemfile, pom.xml or index.html at the repository root"))
}

func detectNode(dir string) (*StackPlan, error) {
	var pkg struct {
		Main            string            `json:"main"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Engines         map[string]string `json:"engines"`
	}
	if err := json.Unmarshal([]byte(readSmall(dir, "package.json")), &pkg); err != nil {
		return nil, errors.New("package.json is not valid JSON")
	}
	dep := func(name string) bool {
		_, a := pkg.Dependencies[name]
		_, b := pkg.DevDependencies[name]
		return a || b
	}

	plan := &StackPlan{AppType: "nodejs", Install: true, Port: deployNodePort, Version: "latest"}
	if m := nodeMajorRE.FindString(pkg.Engines["node"]); m != "" {
		plan.Version = m
	}

	switch {
	case dep("next"):
		plan.Framework = "next"
	case dep("nuxt"):
		plan.Framework = "nuxt"
	case dep("@sveltejs/kit"):
		plan.Framework = "sveltekit"
	case dep("@remix-run/node") || dep("@remix-run/serve"):
		plan.Framework = "remix"
	case dep("astro"):
		plan.Framework = "astro"
	case dep("@nestjs/core"):
		plan.Framework = "nestjs"
	case dep("vite"):
		plan.Framework = "vite"
	case dep("react-scripts"):
		plan.Framework = "create-react-app"
	case dep("express"):
		plan.Framework = "express"
	case dep("fastify"):
		plan.Framework = "fastify"
	default:
		plan.Framework = "node"
	}

	_, hasBuild := pkg.Scripts["build"]
	_, hasStart := pkg.Scripts["start"]
	// most node servers and frameworks read HOST and PORT, so set both instead of guessing per framework
	env := "HOST=0.0.0.0 PORT=3000 "
	switch {
	case hasBuild && hasStart:
		plan.StartCmd = buildOnce("npm run build") + " && " + env + "npm start"
	case hasStart:
		plan.StartCmd = env + "npm start"
	case hasBuild:
		// a build with nothing to start is a static frontend, serve what it outputs
		out := "dist"
		if plan.Framework == "create-react-app" {
			out = "build"
		}
		plan.StartCmd = buildOnce("npm run build") + " && npx -y serve -s " + out + " -l 3000"
		plan.Notes = append(plan.Notes, "Static build, serving ./"+out+" after npm run build.")
	default:
		entry := pkg.Main
		if entry == "" || !strings.HasSuffix(entry, ".js") || strings.Contains(entry, "..") {
			entry = ""
			for _, c := range nodeEntryCandidates {
				if has(dir, c) {
					entry = c
					break
				}
			}
		}
		if entry == "" {
			return nil, errors.Join(ErrUnsupportedStack, errors.New("package.json has no start or build script and no entry file was found"))
		}
		plan.StartupFile = "/var/www/html/" + strings.TrimPrefix(entry, "./")
		plan.StartCmd = env + "node " + strings.TrimPrefix(entry, "./")
	}
	if hasBuild {
		plan.Notes = append(plan.Notes, "npm run build only runs when the commit changed since the last build, so restarts without new commits skip it.")
	}
	if plan.Framework == "next" || plan.Framework == "nuxt" || plan.Framework == "sveltekit" || plan.Framework == "astro" {
		plan.Notes = append(plan.Notes, "Framework builds can need more than 1 GB of RAM, raise mem_limit if npm run build is killed.")
	}
	if !has(dir, "package-lock.json") && !has(dir, "npm-shrinkwrap.json") {
		plan.Notes = append(plan.Notes, "No package-lock.json, npm install will pick the newest matching versions.")
	}
	return plan, nil
}

var pyAppVarRE = regexp.MustCompile(`(?m)^(\w+)\s*=\s*(FastAPI|Flask)\(`)

func detectPython(dir string) *StackPlan {
	reqs := strings.ToLower(readSmall(dir, "requirements.txt") + "\n" + readSmall(dir, "pyproject.toml"))
	plan := &StackPlan{AppType: "python", Install: has(dir, "requirements.txt"), Port: deployPythonPort, Version: "latest"}
	install := ""
	if !plan.Install && has(dir, "pyproject.toml") {
		install = "pip install . && "
	}
	if has(dir, "manage.py") {
		plan.Framework = "django"
		plan.StartCmd = install + "python manage.py migrate --noinput && python manage.py runserver 0.0.0.0:8000"
		plan.Notes = append(plan.Notes, "Runs Django's development server, add gunicorn and your own start_cmd for production. ALLOWED_HOSTS must include the domain.")
		return plan
	}
	for _, c := range pythonEntryCandidates {
		src := readSmall(dir, c)
		m := pyAppVarRE.FindStringSubmatch(src)
		if m == nil {
			continue
		}
		module := strings.TrimSuffix(c, ".py")
		plan.StartupFile = "/var/www/html/" + c
		if m[2] == "FastAPI" {
			plan.Framework = "fastapi"
			if !strings.Contains(reqs, "uvicorn") {
				install += "pip install uvicorn && "
			}
			plan.StartCmd = install + "uvicorn " + module + ":" + m[1] + " --host 0.0.0.0 --port 8000"
		} else {
			plan.Framework = "flask"
			plan.StartCmd = install + "flask --app " + module + ":" + m[1] + " run --host 0.0.0.0 --port 8000"
			plan.Notes = append(plan.Notes, "Runs Flask's development server, add gunicorn and your own start_cmd for production.")
		}
		return plan
	}
	plan.Framework = "python"
	for _, c := range pythonEntryCandidates {
		if has(dir, c) {
			plan.StartupFile = "/var/www/html/" + c
			plan.StartCmd = install + "PORT=8000 python " + c
			plan.Notes = append(plan.Notes, "Plain Python app, it has to listen on 0.0.0.0 port 8000 (PORT is set).")
			return plan
		}
	}
	plan.Notes = append(plan.Notes, "No app.py, main.py, manage.py or run.py found, pass start_cmd.")
	return plan
}

func detectRuby(dir string) *StackPlan {
	plan := &StackPlan{AppType: "ruby", Install: true, Port: deployRubyPort, Version: "latest"}
	switch {
	case has(dir, "bin/rails") || has(dir, "config/application.rb"):
		plan.Framework = "rails"
		plan.StartCmd = "bundle exec rails server -b 0.0.0.0 -p 3000"
		plan.Notes = append(plan.Notes, "Rails runs in development mode unless RAILS_ENV is set in start_cmd, production also needs SECRET_KEY_BASE and a database.")
	case has(dir, "config.ru"):
		plan.Framework = "rack"
		plan.StartCmd = "bundle exec rackup -o 0.0.0.0 -p 3000"
	default:
		plan.Framework = "ruby"
		for _, c := range rubyEntryCandidates {
			if has(dir, c) {
				plan.StartupFile = "/var/www/html/" + c
				plan.StartCmd = "bundle exec ruby " + c + " -o 0.0.0.0 -p 3000"
				break
			}
		}
		if plan.StartCmd == "" {
			plan.Notes = append(plan.Notes, "No config.ru or app.rb found, pass start_cmd.")
		}
	}
	return plan
}

func detectJava(dir string) *StackPlan {
	plan := &StackPlan{AppType: "java", Install: true, Port: deployJavaPort, Framework: "maven"}
	if strings.Contains(readSmall(dir, "pom.xml"), "spring-boot") {
		plan.Framework = "spring-boot"
		plan.StartCmd = "mvn -q -DskipTests spring-boot:run -Dspring-boot.run.arguments=--server.port=8080"
	}
	plan.Notes = append(plan.Notes, "Java apps are expected to listen on port 8080.")
	return plan
}
