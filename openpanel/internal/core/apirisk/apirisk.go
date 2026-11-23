// Package apirisk sorts /api/ routes into read, change and destructive, shared by the AI assistant and scoped MCP tokens.
package apirisk

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"gist.github.com/stefanpejcic/openpanel/internal/core/apidocs"
)

type Risk string

const (
	Read        Risk = "read"
	Change      Risk = "change"
	Destructive Risk = "destructive"
)

var destructivePathRE = regexp.MustCompile(`(?i)kill|stop|suspend|restore|reset|restart|delete|remove|password|/raw|import|/mysql$|/webserver$|/image$|vhost|webserver-conf|/ini$|configuration|detach|edit-file|rename|/move$|permissions|extract|wp-cli|bulk|resources|timezone|/default$|docroot|/dns$|/secure$|/destination$|/settings$|/remote-access$|/cache/[a-z]+$|/services/<|/extensions$|website-builder/<|/containers/<service>$`)

var destructiveDescRE = regexp.MustCompile(`(?i)\b(wipes?|uninstall|deletes?|removes?|drops? (database|user)|overwrites?|restarts?|recreates?|kills?|disable|password)\b`)

// Classify takes a documented route path (with <param> placeholders) and its description
func Classify(method, path, desc string) Risk {
	switch method {
	case http.MethodGet, http.MethodHead:
		return Read
	case http.MethodDelete:
		return Destructive
	}
	if destructivePathRE.MatchString(path) || destructiveDescRE.MatchString(desc) {
		return Destructive
	}
	return Change
}

var (
	docOnce  sync.Once
	docRisks map[string]Risk
)

var docParamRE = regexp.MustCompile(`<[^<>]+>`)
var muxParamRE = regexp.MustCompile(`\{[^{}]+\}`)

// docKey makes "GET /api/x/<int:id>" and the mux pattern "GET /api/x/{id}" compare equal
func docKey(method, path string) string {
	return method + " " + muxParamRE.ReplaceAllString(docParamRE.ReplaceAllString(path, "{}"), "{}")
}

func loadDocRisks() map[string]Risk {
	docOnce.Do(func() {
		docRisks = map[string]Risk{}
		var groups []struct {
			Endpoints []struct {
				Method      string `json:"method"`
				Path        string `json:"path"`
				Description string `json:"description"`
			} `json:"endpoints"`
		}
		if err := json.Unmarshal(apidocs.EndpointsJSON, &groups); err != nil {
			return
		}
		for _, g := range groups {
			for _, ep := range g.Endpoints {
				docRisks[docKey(ep.Method, ep.Path)] = Classify(ep.Method, ep.Path, ep.Description)
			}
		}
	})
	return docRisks
}

// ForRequest classifies a live request by the mux pattern it matched, falling back to its URL path for undocumented routes
func ForRequest(r *http.Request) Risk {
	if pattern := r.Pattern; pattern != "" {
		path := pattern
		if i := strings.IndexByte(pattern, ' '); i != -1 {
			path = pattern[i+1:]
		}
		if risk, ok := loadDocRisks()[docKey(r.Method, path)]; ok {
			return risk
		}
		// the pattern keeps {param} names, which the regexes treat like the docs' <param>
		return Classify(r.Method, muxParamRE.ReplaceAllStringFunc(path, func(s string) string { return "<" + strings.Trim(s, "{}.") + ">" }), "")
	}
	return Classify(r.Method, r.URL.Path, "")
}
