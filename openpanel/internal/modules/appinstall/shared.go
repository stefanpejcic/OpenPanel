package appinstall

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	stdpath "path"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// HandleDockerTags proxies endoflife.date's release feed for the
// "version" dropdown on both install forms, cached 24h.
func HandleDockerTags(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("type")
	if appType != "nodejs" && appType != "python" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid type. Use 'nodejs' or 'python'."})
		return
	}

	ctx := r.Context()
	body, err := cache.Memoize(ctx, a.Cache, "docker_tags_for_py_n_node:"+appType, 24*time.Hour, func() ([]byte, error) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, getErr := client.Get("https://endoflife.date/api/" + appType + ".json")
		if getErr != nil {
			return nil, getErr
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unexpected status code")
		}
		return io.ReadAll(resp.Body)
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch " + appType + " versions."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

// HandleCheckFileExists mirrors helpers.check_file_exists(): used by both
// install forms' startup-file input to show a live exists/does-not-exist
// indicator.
func HandleCheckFileExists(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injectedContext(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	file := r.FormValue("file")
	if file == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "File path is missing"})
		return
	}

	const prefix = "/var/www/html/"
	if len(file) >= len(prefix) && file[:len(prefix)] == prefix {
		file = file[len(prefix):]
	}

	realFilePath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + file

	ext := stdpath.Ext(file)
	if ext != ".py" && ext != ".js" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file extension. Only .py, or .js are allowed."})
		return
	}

	if runErr := exec.CommandContext(r.Context(), "test", "-f", realFilePath).Run(); runErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": file + " does not exist."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": file + " exists."})
}

// RegisterShared wires the two always-on routes only the Python/NodeJS
// install forms use (/docker/tags/<type>, /json/check_if_file_exists).
// Both are gated on the "helpers" feature, not "python"/"nodejs" -
// "helpers" is unconditionally granted to every user (see
// baselineFeatures), so in practice this is login-only, matching why it
// belongs in alwaysOn rather than configured.
func RegisterShared(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "helpers")(h)
	}
	mux.Handle("GET /docker/tags/{type}", requireLogin(func(w http.ResponseWriter, r *http.Request) { HandleDockerTags(a, w, r) }))
	mux.Handle("POST /json/check_if_file_exists", requireLogin(func(w http.ResponseWriter, r *http.Request) { HandleCheckFileExists(a, w, r) }))
}

func injectedContext(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}
