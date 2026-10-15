package docker

import (
	"sort"
	"strings"
)

// knownWebservers is the display order for webservers we have cards for, anything else found in compose goes after these
var knownWebservers = []string{"apache", "nginx", "openresty", "openlitespeed", "litespeed"}

// fallbackWebservers is used when the compose file can't be read or has no webserver services
var fallbackWebservers = []string{"apache", "nginx", "openresty", "openlitespeed"}

// WebserverServices returns services that mount webserver_data and publish ports, backup mounts the volume too but has no ports
func WebserverServices(composeData map[string]any) []string {
	var names []string
	for name, svc := range servicesOf(composeData) {
		ports, _ := svc["ports"].([]any)
		volumes, _ := svc["volumes"].([]any)
		if len(ports) > 0 && mountsVolume(volumes, "webserver_data") {
			names = append(names, name)
		}
	}
	rank := map[string]int{}
	for i, n := range knownWebservers {
		rank[n] = i
	}
	sort.Slice(names, func(i, j int) bool {
		ri, iKnown := rank[names[i]]
		rj, jKnown := rank[names[j]]
		switch {
		case iKnown && jKnown:
			return ri < rj
		case iKnown != jKnown:
			return iKnown
		}
		return names[i] < names[j]
	})
	return names
}

// mountsVolume handles both the short "name:/path" and long {source: name} volume syntax
func mountsVolume(volumes []any, volume string) bool {
	for _, v := range volumes {
		switch v := v.(type) {
		case string:
			if src, _, _ := strings.Cut(v, ":"); src == volume {
				return true
			}
		case map[string]any:
			if src, _ := v["source"].(string); src == volume {
				return true
			}
		}
	}
	return false
}

// AvailableWebservers lists the webservers in the user's own docker-compose.yml, or the built-in list if there are none
func AvailableWebservers(userContext string) []string {
	if composeData, err := LoadCompose(userContext); err == nil {
		if names := WebserverServices(composeData); len(names) > 0 {
			return names
		}
	}
	return append([]string(nil), fallbackWebservers...)
}

// webserverChoicesFor builds the switch page cards from the user's compose file, keeping the current webserver even if it's gone from there
func webserverChoicesFor(userContext, current string) []ServerChoice {
	names := AvailableWebservers(userContext)
	if current != "" && !containsString(names, current) {
		names = append([]string{current}, names...)
	}
	known := map[string]ServerChoice{}
	for _, c := range webserverChoices {
		known[c.Value] = c
	}
	out := make([]ServerChoice, 0, len(names))
	for _, name := range names {
		c, ok := known[name]
		if !ok {
			c = ServerChoice{Value: name, Label: name}
		}
		c.Current = name == current
		out = append(out, c)
	}
	return out
}
