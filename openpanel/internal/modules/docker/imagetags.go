package docker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// imageVarRE splits a compose image like "redis:${REDIS_VERSION:-8.6.2-alpine}" into repo, env var and default tag
var imageVarRE = regexp.MustCompile(`^([^\s:${}]+(?::\d+)?(?:/[^\s:${}]+)*):\$\{([A-Za-z_][A-Za-z0-9_]*)(?::?-([^}]*))?\}$`)

// imageTagRE is the tag format Docker accepts, it also keeps quotes and newlines out of .env
var imageTagRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

// ServiceImage is where a service's image comes from and which .env variable holds its tag
type ServiceImage struct {
	Repo       string
	EnvVar     string
	DefaultTag string
}

// parseServiceImage reads the image of one compose service, ok is false when its tag isn't set through a variable
func parseServiceImage(userContext, service string) (ServiceImage, bool) {
	composeData, err := LoadCompose(userContext)
	if err != nil {
		return ServiceImage{}, false
	}
	image, _ := servicesOf(composeData)[service]["image"].(string)
	m := imageVarRE.FindStringSubmatch(strings.TrimSpace(image))
	if m == nil {
		return ServiceImage{}, false
	}
	return ServiceImage{Repo: m[1], EnvVar: strings.ToUpper(m[2]), DefaultTag: m[3]}, true
}

// fixedImageServices are internal services whose image must not be changed from the panel
var fixedImageServices = map[string]bool{"cron": true, "backup": true, "docker-proxy": true}

// imageChangeable reports whether a service's tag may be changed from the Software Versions page
func imageChangeable(service string) bool {
	return service != "" && !fixedImageServices[service] && !strings.HasPrefix(service, "php-")
}

// dockerHubPage links to the tags of an image on Docker Hub, "" for other registries
func dockerHubPage(userContext, service string) string {
	img, ok := parseServiceImage(userContext, service)
	if !ok {
		return ""
	}
	hubRepo, ok := dockerHubRepo(img.Repo)
	if !ok {
		return ""
	}
	if name, official := strings.CutPrefix(hubRepo, "library/"); official {
		return "https://hub.docker.com/_/" + name + "/tags"
	}
	return "https://hub.docker.com/r/" + hubRepo + "/tags"
}

// fixedTagVars are the services whose tag variable is fixed, whatever the compose image says
func fixedTagVar(userContext, service string) (string, bool) {
	switch service {
	case "phpmyadmin":
		return "PMA_VERSION", true
	case "mysql", "mariadb":
		return "MYSQL_VERSION", true
	case userContext:
		return "OS", true
	}
	return "", false
}

// imageTagVar is the .env variable that sets the service's tag, reading and saving always use the same one
func imageTagVar(userContext, service string) (envVar, defaultTag string) {
	img, parsed := parseServiceImage(userContext, service)
	if v, ok := fixedTagVar(userContext, service); ok {
		return v, img.DefaultTag
	}
	if parsed {
		return img.EnvVar, img.DefaultTag
	}
	return strings.ToUpper(service) + "_VERSION", ""
}

// dockerHubRepo turns an image repo into Docker Hub's namespace/name, false for other registries
func dockerHubRepo(repo string) (string, bool) {
	repo = strings.TrimPrefix(strings.TrimPrefix(repo, "docker.io/"), "index.docker.io/")
	parts := strings.Split(repo, "/")
	if strings.ContainsAny(parts[0], ".:") || parts[0] == "localhost" {
		return "", false
	}
	switch len(parts) {
	case 1:
		return "library/" + parts[0], true
	case 2:
		return repo, true
	}
	return "", false
}

// keepTag drops signature/attestation tags that aren't real images
func keepTag(tag string) bool {
	return !strings.HasPrefix(tag, "sha256-") && !strings.HasSuffix(tag, ".sig") && !strings.HasSuffix(tag, ".att") && imageTagRE.MatchString(tag)
}

// fetchDockerHubTags lists up to maxTags of a repo's tags, most recently pushed first
func fetchDockerHubTags(ctx context.Context, hubRepo string, maxTags int) ([]string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	next := "https://hub.docker.com/v2/repositories/" + hubRepo + "/tags?page_size=100&ordering=last_updated"
	var tags []string
	for page := 0; page < 5 && next != "" && len(tags) < maxTags; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, errors.New("Docker Hub returned " + resp.Status)
		}
		var payload struct {
			Next    string `json:"next"`
			Results []struct {
				Name string `json:"name"`
			} `json:"results"`
		}
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		for _, r := range payload.Results {
			if keepTag(r.Name) {
				tags = append(tags, r.Name)
			}
		}
		// only follow Docker Hub's own next links
		if u, err := url.Parse(payload.Next); err != nil || u.Host != "hub.docker.com" {
			next = ""
		} else {
			next = payload.Next
		}
	}
	if len(tags) > maxTags {
		tags = tags[:maxTags]
	}
	return tags, nil
}

// handleContainerImageTags returns the Docker Hub tags for a service's image, cached in redis for 12h per image so all users share one fetch
func handleContainerImageTags(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	service := r.PathValue("service")
	if !imageChangeable(service) {
		writeJSON(w, map[string]any{"tags": []string{}, "error": "The image of this service can't be changed."})
		return
	}

	img, ok := parseServiceImage(userContext, service)
	if !ok {
		writeJSON(w, map[string]any{"tags": []string{}, "error": "This service's image tag isn't set through a variable."})
		return
	}
	hubRepo, ok := dockerHubRepo(img.Repo)
	if !ok {
		writeJSON(w, map[string]any{"tags": []string{}, "repo": img.Repo, "error": "Tags can only be listed for Docker Hub images."})
		return
	}
	tags, err := cache.Memoize(r.Context(), a.Cache, "dockerhub_tags:"+hubRepo, 12*time.Hour, func() ([]string, error) {
		return fetchDockerHubTags(r.Context(), hubRepo, 300)
	})
	if err != nil {
		writeJSON(w, map[string]any{"tags": []string{}, "repo": img.Repo, "error": "Could not load tags from Docker Hub."})
		return
	}
	writeJSON(w, map[string]any{"tags": tags, "repo": img.Repo, "default": img.DefaultTag})
}
