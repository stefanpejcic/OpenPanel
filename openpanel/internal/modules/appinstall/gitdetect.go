package appinstall

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-git/go-git/v5"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// nodeEntryCandidates/pythonEntryCandidates/rubyEntryCandidates are checked in order when there's no manifest to read (or package.json has no "main"), matching the default filenames buildAppRunCommand() falls back to
var (
	nodeEntryCandidates   = []string{"index.js", "server.js", "app.js", "main.js"}
	pythonEntryCandidates = []string{"app.py", "main.py", "manage.py", "run.py"}
	rubyEntryCandidates   = []string{"app.rb", "main.rb", "server.rb", "config.ru"}
)

// detectStartupFile shallow-clones gitURL into a throwaway temp dir just to guess the entry point file, then discards it - never touches a user's actual app container or docroot. appType picks the candidate list/manifest format.
func detectStartupFile(ctx context.Context, gitURL string, appType string) (string, error) {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or missing git repository URL."})
		return
	}
	if appType != "nodejs" && appType != "python" && appType != "ruby" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid app type."})
		return
	}

	startupFile, err := detectStartupFile(r.Context(), gitURL, appType)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Could not read repository: " + err.Error()})
		return
	}
	if startupFile == "" {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Could not detect a startup file, please set it manually."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"startup_file": "/var/www/html/" + startupFile})
}
