package appinstall

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	stdpath "path"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
// "version" dropdown on the nodejs/python install forms (cached 24h), or -
// for ruby - queries Docker Hub's own tag list directly (per explicit
// request: ruby's versions should come from Docker Hub, not endoflife.date)
// and reshapes it into the same [{"latest": "X.Y.Z"}, ...] shape the
// frontend already expects, so python_node_apps.html's fetch/render code
// doesn't need a ruby-specific branch.
func HandleDockerTags(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("type")
	if appType != "nodejs" && appType != "python" && appType != "ruby" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid type. Use 'nodejs', 'python', or 'ruby'."})
		return
	}

	ctx := r.Context()

	if appType == "ruby" {
		versions, err := cache.Memoize(ctx, a.Cache, "docker_tags_for_ruby", 24*time.Hour, func() ([]string, error) {
			return fetchRubyDockerHubVersions(ctx)
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch ruby versions."})
			return
		}
		type tagEntry struct {
			Latest string `json:"latest"`
		}
		entries := make([]tagEntry, len(versions))
		for i, v := range versions {
			entries[i] = tagEntry{Latest: v}
		}
		writeJSON(w, http.StatusOK, entries)
		return
	}

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

var rubyCleanTagRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// fetchRubyDockerHubVersions queries Docker Hub's own registry API for the
// official `ruby` image's every tag, keeping only plain "X.Y.Z" tags -
// Docker Hub also lists variant tags like "3.3.6-slim", "3.3.6-alpine",
// "3-bookworm" etc, which are deliberately excluded here since the compose
// template (configuration/docker/compose/ruby.yml) only substitutes this
// value as-is into `ruby:${...TAG}`, not as a suffix - a user picking a
// version from this list expects the plain "ruby:X.Y.Z" image, not a
// slim/alpine variant they didn't ask for.
//
// Deliberately does NOT rely on the API's `ordering` param to find the
// newest versions quickly: confirmed live that `ordering=-name` fails to
// surface current major versions first (a page_size=10&ordering=-name
// request returned "1", "1.9", "1.9.3", ... i.e. ascending from the
// oldest tag - the "-" descending prefix appears to be ignored entirely),
// and `ordering=-last_updated` is no better, since older patch tags get
// rebuilt for base-image security patches more often than a stable recent
// release does, so "most recently updated" doesn't track "highest
// version" either. The library/ruby repo has ~1700 tags total, so this
// pages through all of them (default ordering, whatever Docker Hub
// actually returns) and sorts properly in Go below instead.
func fetchRubyDockerHubVersions(ctx context.Context) ([]string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	seen := make(map[string]bool)
	var versions []string

	url := "https://hub.docker.com/v2/repositories/library/ruby/tags?page_size=100"
	for page := 0; page < 25 && url != ""; page++ {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		resp, getErr := client.Do(req)
		if getErr != nil {
			return nil, getErr
		}
		var payload struct {
			Next    string `json:"next"`
			Results []struct {
				Name string `json:"name"`
			} `json:"results"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		for _, r := range payload.Results {
			if rubyCleanTagRE.MatchString(r.Name) && !seen[r.Name] {
				seen[r.Name] = true
				versions = append(versions, r.Name)
			}
		}
		url = payload.Next
	}
	if len(versions) == 0 {
		return nil, errors.New("no ruby versions found on Docker Hub")
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareRubyVersions(versions[i], versions[j]) > 0
	})
	return versions, nil
}

func compareRubyVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		na, _ := strconv.Atoi(pa[i])
		nb, _ := strconv.Atoi(pb[i])
		if na != nb {
			return na - nb
		}
	}
	return 0
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
	if ext != ".py" && ext != ".js" && ext != ".rb" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file extension. Only .py, .js, or .rb are allowed."})
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
	mux.Handle("POST /json/detect_git_startup_file", requireLogin(func(w http.ResponseWriter, r *http.Request) { HandleDetectGitStartupFile(a, w, r) }))
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
