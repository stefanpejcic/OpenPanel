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

// nodeEntryCandidates/pythonEntryCandidates are checked in order when
// package.json has no "main" field (Node) or there's no manifest to read
// at all (Python) - matching the same default filenames
// buildAppRunCommand() falls back to.
var (
	nodeEntryCandidates   = []string{"index.js", "server.js", "app.js", "main.js"}
	pythonEntryCandidates = []string{"app.py", "main.py", "manage.py", "run.py"}
)

// detectStartupFile shallow-clones gitURL into a throwaway temp dir purely
// to guess the entry point file, then discards the clone - this never
// touches a user's actual app container or docroot. isNode picks which
// candidate list / manifest format to look for.
func detectStartupFile(ctx context.Context, gitURL string, isNode bool) (string, error) {
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

	if isNode {
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
	}

	for _, candidate := range pythonEntryCandidates {
		if fileExists(tmpDir + "/" + candidate) {
			return candidate, nil
		}
	}
	return "", nil
}

// HandleDetectGitStartupFile powers the install form's "Git repository
// URL" field: on blur, it tries to guess the startup file from the repo so
// the user doesn't have to know it up front, but the field always stays
// editable - this is a best-effort suggestion, not a requirement.
func HandleDetectGitStartupFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	gitURL := r.FormValue("git_repo_url")
	appType := r.FormValue("app_type")

	if !isValidGitURL(gitURL) || gitURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or missing git repository URL."})
		return
	}
	if appType != "nodejs" && appType != "python" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid app type."})
		return
	}

	startupFile, err := detectStartupFile(r.Context(), gitURL, appType == "nodejs")
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
