package aiassistant

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"sync"

	"gist.github.com/stefanpejcic/openpanel/internal/core/apidocs"
)

// what the assistant may do: off answers only, read looks things up, all can also make changes
const (
	modeOff  = "off"
	modeRead = "read"
	modeAll  = "all"
)

type risk string

const (
	riskRead        risk = "read"
	riskChange      risk = "change"
	riskDestructive risk = "destructive"
)

// action is one documented /api/ endpoint the assistant can call
type action struct {
	Name        string
	Method      string
	Path        string
	Group       string
	Features    []string
	Title       string
	Description string
	Body        json.RawMessage
	Params      string
	PathParams  []string
	Risk        risk
	// output can hold text written by others (files, logs, visitors), so it may carry injected instructions
	Untrusted bool
	terms     map[string]int
}

var pathParamRE = regexp.MustCompile(`<(?:[^:<>]+:)?([^<>]+)>`)

// never exposed: logins, credentials, tokens, autologin links, downloads and uploads the chat can't handle
var excludedPrefixes = []string{
	"/api/login", "/api/endpoints", "/api/account/sessions", "/api/account/passkeys", "/api/account/2fa",
	"/api/account/mcp", "/api/mysql/root-password", "/api/phpmyadmin", "/api/webmail/", "/api/dashboard/tour", "/api/helpers/",
}

var excludedRE = regexp.MustCompile(`(?i)/temporary-link$|/wp-info$|/favicon$|/screenshot$|/download|/export$|/configuration/<|^/api/stats/<`)

var excludedDescRE = regexp.MustCompile(`(?i)multipart/form-data|form-encoded|^download`)

var destructivePathRE = regexp.MustCompile(`(?i)kill|stop|suspend|restore|reset|restart|delete|remove|password|/raw|import|/mysql$|/webserver$|/image$|vhost|webserver-conf|/ini$|configuration|detach|edit-file|rename|/move$|permissions|extract|wp-cli|bulk|resources|timezone|/default$|docroot|/dns$|/secure$|/destination$|/settings$|/remote-access$|/cache/[a-z]+$|/services/<|/extensions$|website-builder/<|/containers/<service>$`)

var destructiveDescRE = regexp.MustCompile(`(?i)\b(wipes?|uninstall|deletes?|removes?|drops? (database|user)|overwrites?|restarts?|recreates?|kills?|disable|password)\b`)

var untrustedRE = regexp.MustCompile(`(?i)log|file|folder|trash|quarantine|visitors|vhost|/raw|filters|processlist|website-builder|wp-cli|activity|/search/|/stats|scan|safebrowsing|vulnerability|pagespeed|tables|/ini$`)

var (
	actionsOnce sync.Once
	actionList  []*action
	actionByKey map[string]*action
)

func classify(method, path, desc string) risk {
	switch method {
	case "GET", "HEAD":
		return riskRead
	case "DELETE":
		return riskDestructive
	}
	if destructivePathRE.MatchString(path) || destructiveDescRE.MatchString(desc) {
		return riskDestructive
	}
	return riskChange
}

func excluded(method, path, desc string) bool {
	if method == "PATCH" && path == "/api/account" {
		return true
	}
	for _, p := range excludedPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return excludedRE.MatchString(path) || excludedDescRE.MatchString(desc)
}

func actionSlug(path string) string {
	slug := pathParamRE.ReplaceAllString(path, "$1")
	slug = strings.TrimPrefix(slug, "/api/")
	slug = strings.NewReplacer("/", "_", "-", "_").Replace(slug)
	return strings.Trim(slug, "_")
}

// loadActions parses the api docs once, names match the MCP tool names
func loadActions() []*action {
	actionsOnce.Do(func() {
		var groups []struct {
			Group     string          `json:"group"`
			Feature   json.RawMessage `json:"feature"`
			Endpoints []struct {
				Method      string          `json:"method"`
				Path        string          `json:"path"`
				Description string          `json:"description"`
				Body        json.RawMessage `json:"body"`
				Params      string          `json:"params"`
			} `json:"endpoints"`
		}
		actionByKey = map[string]*action{}
		if err := json.Unmarshal(apidocs.EndpointsJSON, &groups); err != nil {
			return
		}
		perSlug := map[string]int{}
		for _, g := range groups {
			for _, ep := range g.Endpoints {
				perSlug[actionSlug(ep.Path)]++
			}
		}
		for _, g := range groups {
			var features []string
			if json.Unmarshal(g.Feature, &features) != nil {
				var one string
				_ = json.Unmarshal(g.Feature, &one)
				features = []string{one}
			}
			for _, ep := range g.Endpoints {
				if excluded(ep.Method, ep.Path, ep.Description) {
					continue
				}
				name := actionSlug(ep.Path)
				if perSlug[name] > 1 {
					name += "_" + strings.ToLower(ep.Method)
				}
				act := &action{
					Name: name, Method: ep.Method, Path: ep.Path, Group: g.Group, Features: features,
					Title: shortTitle(ep.Description), Description: ep.Description, Body: ep.Body, Params: ep.Params,
					Risk: classify(ep.Method, ep.Path, ep.Description),
				}
				act.Untrusted = act.Risk == riskRead && untrustedRE.MatchString(ep.Path)
				for _, m := range pathParamRE.FindAllStringSubmatch(ep.Path, -1) {
					act.PathParams = append(act.PathParams, m[1])
				}
				act.terms = map[string]int{}
				for _, t := range tokenize(g.Group + " " + ep.Description + " " + strings.ReplaceAll(name, "_", " ")) {
					act.terms[t]++
				}
				actionList = append(actionList, act)
				actionByKey[name] = act
			}
		}
	})
	return actionList
}

// shortTitle is the first clause of the description, for cards and the activity log
func shortTitle(desc string) string {
	t := desc
	if i := strings.Index(t, " ("); i > 0 {
		t = t[:i]
	}
	if i := strings.Index(t, ". "); i > 0 {
		t = t[:i]
	}
	if i := strings.Index(t, ": "); i > 0 {
		t = t[:i]
	}
	return strings.TrimSuffix(strings.TrimSpace(t), ".")
}

func findAction(name string) *action {
	loadActions()
	return actionByKey[name]
}

// allowedFor reports whether the plan has the action's feature and the admin's mode lets it run
func (act *action) allowedFor(features map[string]bool, mode string) bool {
	if mode == modeOff || (mode == modeRead && act.Risk != riskRead) {
		return false
	}
	for _, f := range act.Features {
		if features[f] {
			return true
		}
	}
	return false
}

// searchActions returns the actions that best match the query, only ones this user can run
func searchActions(query string, features map[string]bool, mode string, limit int) []*action {
	q := tokenize(query)
	type scored struct {
		act   *action
		score int
	}
	var res []scored
	for _, act := range loadActions() {
		if !act.allowedFor(features, mode) {
			continue
		}
		score := 0
		for _, t := range q {
			score += act.terms[t]
		}
		if score > 0 {
			res = append(res, scored{act, score})
		}
	}
	sort.SliceStable(res, func(i, j int) bool { return res[i].score > res[j].score })
	var out []*action
	for i := 0; i < len(res) && i < limit; i++ {
		out = append(out, res[i].act)
	}
	return out
}

// describe is what the model sees for one action in find_actions results
func (act *action) describe() map[string]any {
	d := map[string]any{
		"action":      act.Name,
		"method":      act.Method,
		"path":        act.Path,
		"description": act.Description,
		"kind":        string(act.Risk),
	}
	if len(act.PathParams) > 0 {
		d["params"] = act.PathParams
	}
	if len(act.Body) > 0 {
		d["body_example"] = act.Body
	}
	if act.Params != "" {
		d["query_example"] = act.Params
	}
	return d
}

// groupsFor lists the action groups available to this user, so the model knows what it can look for
func groupsFor(features map[string]bool, mode string) []string {
	seen := map[string]bool{}
	var out []string
	for _, act := range loadActions() {
		if act.allowedFor(features, mode) && !seen[act.Group] {
			seen[act.Group] = true
			out = append(out, act.Group)
		}
	}
	return out
}
