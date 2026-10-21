// Package dockertemplates loads the one-click Docker app templates shown on the autoinstaller, each one just prefills /containers/new
package dockertemplates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

const (
	SourceURL = "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/templates/applications.yml"
	// the configuration repo is checked out in /etc/openpanel, used when github can't be reached
	localPath = "/etc/openpanel/docker/templates/applications.yml"
	cacheKey  = "docker_templates"
	cacheTTL  = 24 * time.Hour
	maxSize   = 1 << 20
)

// Template is one entry from applications.yml
type Template struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Category    string   `yaml:"category" json:"category"`
	Logo        string   `yaml:"logo" json:"logo"`
	Docs        string   `yaml:"docs" json:"docs"`
	Image       string   `yaml:"image" json:"image"`
	Service     string   `yaml:"service" json:"service"`
	CPU         string   `yaml:"cpu" json:"cpu"`
	RAM         string   `yaml:"ram" json:"ram"`
	PIDs        int      `yaml:"pids" json:"pids"`
	Port        int      `yaml:"port" json:"port"`
	Networks    []string `yaml:"networks" json:"networks"`
	Volumes     []string `yaml:"volumes" json:"volumes"`
	Env         []string `yaml:"env" json:"env"`
	Healthcheck string   `yaml:"healthcheck" json:"healthcheck"`
	Socket      bool     `yaml:"socket" json:"socket"`
}

type file struct {
	Applications []Template `yaml:"applications"`
}

var (
	idRE      = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,40}$`)
	serviceRE = regexp.MustCompile(`^[a-z][a-z0-9]{2,}$`)
	imageRE   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/:@-]{0,254}$`)
)

// InstallURL is the /containers/new link that prefills the form with this template
func (t Template) InstallURL() string {
	q := url.Values{}
	q.Set("image", t.Image)
	q.Set("name", t.Service)
	if t.CPU != "" {
		q.Set("cpu", t.CPU)
	}
	if t.RAM != "" {
		q.Set("ram", t.RAM)
	}
	if t.PIDs > 0 {
		q.Set("pids", strconv.Itoa(t.PIDs))
	}
	if t.Port > 0 {
		q.Set("port", strconv.Itoa(t.Port))
	}
	if len(t.Networks) > 0 {
		q.Set("network", strings.Join(t.Networks, ","))
	}
	for _, v := range t.Volumes {
		q.Add("volume", v)
	}
	for _, e := range t.Env {
		q.Add("env", e)
	}
	if t.Healthcheck != "" {
		q.Set("healthcheck", t.Healthcheck)
	}
	if t.Socket {
		q.Set("socket", "1")
	}
	return "/containers/new?" + q.Encode()
}

// Load returns the templates from cache, refreshed from github every 24h, the local copy is used when github fails
func Load(ctx context.Context, a *appctx.App) []Template {
	list, err := cache.Memoize(ctx, a.Cache, cacheKey, cacheTTL, func() ([]Template, error) {
		var list []Template
		data, err := fetch(ctx)
		if err == nil {
			list, err = Parse(data)
		}
		if err == nil {
			return list, nil
		}
		log.Printf("DOCKER TEMPLATES - %s: %v, using %s", SourceURL, err, localPath)
		local, readErr := tenantfs.ReadFile(localPath)
		if readErr != nil {
			return nil, readErr
		}
		return Parse(local)
	})
	if err != nil {
		return nil
	}
	return list
}

func fetch(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, SourceURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxSize))
}

// Parse reads applications.yml and drops entries that wouldn't give a usable form
func Parse(data []byte) ([]Template, error) {
	var f file
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	var out []Template
	seen := map[string]bool{}
	for _, t := range f.Applications {
		if !idRE.MatchString(t.ID) || seen[t.ID] || t.Name == "" || !imageRE.MatchString(t.Image) || !serviceRE.MatchString(t.Service) {
			continue
		}
		if t.Logo != "" && !strings.HasPrefix(t.Logo, "https://") {
			t.Logo = ""
		}
		if t.Docs != "" && !strings.HasPrefix(t.Docs, "https://") {
			t.Docs = ""
		}
		seen[t.ID] = true
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil, errors.New("no valid templates")
	}
	return out, nil
}
