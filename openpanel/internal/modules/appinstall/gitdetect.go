package appinstall

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-git/go-git/v5"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterGitDetect wires the startup-file detect helper, only registered when the "git" module is enabled
func RegisterGitDetect(mux *http.ServeMux, a *appctx.App) {
	mux.Handle("POST /json/detect_git_startup_file", auth.RequireLogin(a, "git")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleDetectGitStartupFile(a, w, r)
	})))
}

// RegisterGitDetectAPI is RegisterGitDetect's /api/ twin
func RegisterGitDetectAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "git", "POST /api/helpers/detect-git-startup-file", func(w http.ResponseWriter, r *http.Request) { apiDetectGitStartupFile(a, w, r) })
}

// gitAllowed reports whether the git module is enabled and on this user's plan
func gitAllowed(a *appctx.App, r *http.Request) bool {
	if !a.ModuleEnabled("git") {
		return false
	}
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return false
	}
	allowed, _ := injected["user_allowed"].([]string)
	for _, m := range allowed {
		if m == "git" {
			return true
		}
	}
	return false
}

// nodeEntryCandidates/pythonEntryCandidates/rubyEntryCandidates/javaEntryCandidates are checked in order when there's no manifest to read (or package.json has no "main"), matching the default filenames buildAppRunCommand() falls back to
var (
	nodeEntryCandidates   = []string{"index.js", "server.js", "app.js", "main.js"}
	pythonEntryCandidates = []string{"app.py", "main.py", "manage.py", "run.py"}
	rubyEntryCandidates   = []string{"app.rb", "main.rb", "server.rb"}
	javaEntryCandidates   = []string{"Main.java", "App.java", "Application.java"}
)

// isDetectableAppType gates both detect endpoints, n8n has no startup file to guess
func isDetectableAppType(appType string) bool {
	return appType == "nodejs" || appType == "python" || appType == "ruby" || appType == "java"
}

// detectStartupFile shallow-clones gitURL into a throwaway temp dir just to guess the entry point file, then discards it - never touches a user's actual app container or docroot. appType picks the candidate list/manifest format. Returns the full /var/www/html/ path, or "" if nothing matched.
func detectStartupFile(ctx context.Context, gitURL string, appType string) (string, error) {
	file, err := detectEntryFile(ctx, gitURL, appType)
	if err != nil || file == "" {
		return "", err
	}
	return "/var/www/html/" + file, nil
}

func detectEntryFile(ctx context.Context, gitURL string, appType string) (string, error) {
	tmpDir, mkErr := os.MkdirTemp("", "opdetect-*")
	if mkErr != nil {
		return "", mkErr
	}
	defer os.RemoveAll(tmpDir)

	cloneCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	_, cloneErr := git.PlainCloneContext(cloneCtx, tmpDir, false, &git.CloneOptions{
		URL:          gitURL,
		Depth:        1,
		SingleBranch: true,
	})
	if cloneErr != nil {
		return "", cloneErr
	}

	switch appType {
	case "nodejs":
		if pkgBytes, readErr := os.ReadFile(tmpDir + "/package.json"); readErr == nil {
			var pkg struct {
				Main string `json:"main"`
			}
			if json.Unmarshal(pkgBytes, &pkg) == nil && pkg.Main != "" {
				if fileExists(tmpDir + "/" + pkg.Main) {
					return pkg.Main, nil
				}
			}
		}
		for _, candidate := range nodeEntryCandidates {
			if fileExists(tmpDir + "/" + candidate) {
				return candidate, nil
			}
		}
		return "", nil
	case "ruby":
		for _, candidate := range rubyEntryCandidates {
			if fileExists(tmpDir + "/" + candidate) {
				return candidate, nil
			}
		}
		return "", nil
	case "java":
		for _, candidate := range javaEntryCandidates {
			if fileExists(tmpDir + "/" + candidate) {
				return candidate, nil
			}
		}
		return "", nil
	default:
		for _, candidate := range pythonEntryCandidates {
			if fileExists(tmpDir + "/" + candidate) {
				return candidate, nil
			}
		}
		return "", nil
	}
}

// HandleDetectGitStartupFile powers the install form's "Git repository URL" field: on blur, it guesses the startup file from the repo, but the field stays editable - a best-effort suggestion, not a requirement
func HandleDetectGitStartupFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	gitURL := r.FormValue("git_repo_url")
	appType := r.FormValue("app_type")

	if !isValidGitURL(gitURL) || gitURL == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or missing git repository URL."})
		return
	}
	if !isDetectableAppType(appType) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid app type."})
		return
	}

	startupFile, err := detectStartupFile(r.Context(), gitURL, appType)
	if err != nil {
		web.WriteJSON(w, http.StatusOK, map[string]string{"error": "Could not read repository: " + err.Error()})
		return
	}
	if startupFile == "" {
		web.WriteJSON(w, http.StatusOK, map[string]string{"error": "Could not detect a startup file, please set it manually."})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"startup_file": startupFile})
}
