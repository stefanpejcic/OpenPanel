package appinstall

import (
	"encoding/json"
	"net/http"
	"os/exec"
	stdpath "path"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterSharedAPI wires /api/ equivalents of the three helper endpoints
// the Python/NodeJS install forms call over AJAX (docker/tags,
// check_if_file_exists, detect_git_startup_file). Gated on the "helpers"
// feature, same as RegisterShared - unconditionally granted, so in
// practice this is API-key-only.
func RegisterSharedAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "helpers", "GET /api/docker/tags/{type}", func(w http.ResponseWriter, r *http.Request) { HandleDockerTags(a, w, r) })
	apiregistry.Handle(mux, a, "helpers", "POST /api/helpers/check-file-exists", func(w http.ResponseWriter, r *http.Request) { apiCheckFileExists(a, w, r) })
	apiregistry.Handle(mux, a, "helpers", "POST /api/helpers/detect-git-startup-file", func(w http.ResponseWriter, r *http.Request) { apiDetectGitStartupFile(a, w, r) })
}

// apiCheckFileExists is HandleCheckFileExists (shared.go) with a
// JSON-body request instead of a form post.
func apiCheckFileExists(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		File string `json:"file"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.File = r.Form.Get("file")
	}
	file := body.File
	if file == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "File path is missing"})
		return
	}

	const prefix = "/var/www/html/"
	if strings.HasPrefix(file, prefix) {
		file = file[len(prefix):]
	}

	realFilePath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + file

	ext := stdpath.Ext(file)
	if ext != ".py" && ext != ".js" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file extension. Only .py, or .js are allowed."})
		return
	}

	if runErr := exec.CommandContext(r.Context(), "test", "-f", realFilePath).Run(); runErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{"file": file, "exists": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"file": file, "exists": true})
}

// apiDetectGitStartupFile is HandleDetectGitStartupFile (gitdetect.go)
// with a JSON-body request instead of a form post.
func apiDetectGitStartupFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		GitRepoURL string `json:"git_repo_url"`
		AppType    string `json:"app_type"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.GitRepoURL = r.Form.Get("git_repo_url")
		body.AppType = r.Form.Get("app_type")
	}

	if !isValidGitURL(body.GitRepoURL) || body.GitRepoURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or missing git repository URL."})
		return
	}
	if body.AppType != "nodejs" && body.AppType != "python" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid app type."})
		return
	}

	startupFile, err := detectStartupFile(r.Context(), body.GitRepoURL, body.AppType == "nodejs")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Could not read repository: " + err.Error()})
		return
	}
	if startupFile == "" {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Could not detect a startup file, please set it manually."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"startup_file": startupFile})
}
