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

// HandleDockerTags proxies endoflife.date's release feed for the "version" dropdown on the nodejs/python install forms (cached 24h). For ruby it queries Docker Hub's own tag list directly instead, reshaped into the same [{"latest": "X.Y.Z"}, ...] shape so the frontend doesn't need a ruby-specific branch.
func HandleDockerTags(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("type")
	if appType != "nodejs" && appType != "python" && appType != "ruby" && appType != "java" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid type. Use 'nodejs', 'python', 'ruby', or 'java'."})
		return
	}

	ctx := r.Context()

	// ruby and java both come straight from Docker Hub's tag list. Java has no plain "X.Y.Z" tags at all (eclipse-temurin only ships suffixed tags like "21-jdk-jammy"), so it gets its own cache key and filter, reshaped into the same [{"latest": "X"}, ...] shape - no java-specific branch needed in the frontend.
	if appType == "ruby" || appType == "java" {
		cacheKey, fetch := "docker_tags_for_ruby", fetchRubyDockerHubVersions
		if appType == "java" {
			cacheKey, fetch = "docker_tags_for_java", fetchJavaDockerHubVersions
		}
		versions, err := cache.Memoize(ctx, a.Cache, cacheKey, 24*time.Hour, func() ([]string, error) {
			return fetch(ctx)
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch " + appType + " versions."})
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

// fetchRubyDockerHubVersions queries Docker Hub's registry API for the official `ruby` image's tags, keeping only plain "X.Y.Z" ones - variant tags like "3.3.6-slim" are excluded since the compose template substitutes this value as-is into `ruby:${...TAG}`. Doesn't rely on the API's `ordering` param since neither -name nor -last_updated actually sorts by version in practice, so this pages through all ~1700 tags and sorts properly in Go.
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

// javaCleanTagRE matches eclipse-temurin's plain major-version LTS tags ("8-jdk-jammy", "17-jdk-jammy", ...) - eclipse-temurin ships no bare "X.Y.Z" tags at all, only ones suffixed with a JVM variant and base OS, and "-jdk-jammy" is the one the compose template actually substitutes, same reasoning as rubyCleanTagRE.
var javaCleanTagRE = regexp.MustCompile(`^(\d+)-jdk-jammy$`)

// fetchJavaDockerHubVersions queries Docker Hub's registry API for the official `eclipse-temurin` image's tags, keeping only clean major-version LTS tags (see javaCleanTagRE), newest first
func fetchJavaDockerHubVersions(ctx context.Context) ([]string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	seen := make(map[string]bool)
	var versions []string

	url := "https://hub.docker.com/v2/repositories/library/eclipse-temurin/tags?page_size=100"
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
			if javaCleanTagRE.MatchString(r.Name) && !seen[r.Name] {
				seen[r.Name] = true
				versions = append(versions, r.Name)
			}
		}
		url = payload.Next
	}
	if len(versions) == 0 {
		return nil, errors.New("no java versions found on Docker Hub")
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareJavaVersions(versions[i], versions[j]) > 0
	})
	return versions, nil
}

func compareJavaVersions(a, b string) int {
	na, _ := strconv.Atoi(javaCleanTagRE.FindStringSubmatch(a)[1])
	nb, _ := strconv.Atoi(javaCleanTagRE.FindStringSubmatch(b)[1])
	return na - nb
}

// HandleCheckFileExists mirrors helpers.check_file_exists(), used by both install forms' startup-file input to show a live exists/does-not-exist indicator
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
	if ext != ".py" && ext != ".js" && ext != ".rb" && ext != ".java" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file extension. Only .py, .js, .rb, or .java are allowed."})
		return
	}

	if runErr := exec.CommandContext(r.Context(), "test", "-f", realFilePath).Run(); runErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": file + " does not exist."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": file + " exists."})
}

// RegisterShared wires the two always-on routes only the Python/NodeJS install forms use. Gated on "helpers", not "python"/"nodejs" - "helpers" is unconditionally granted (see baselineFeatures), so in practice this is login-only.
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
