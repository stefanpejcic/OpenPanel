package errorlogs

import (
	"context"
	"regexp"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// Site is a row from the sites table, Container is set for apps that run in their own container
type Site struct {
	Name      string
	Type      string
	Container string
}

// Target is one option in the website filter, a domain or a site in one of its subfolders
type Target struct {
	Value string
	Sub   bool
}

// appTypeLabels are the sites.type values that run in their own container
var appTypeLabels = map[string]string{
	"NodeJS": "Node.js", "Python": "Python", "Ruby": "Ruby", "Java": "Java", "n8n": "n8n", "Docker": "Container",
}

func loadSites(ctx context.Context, a *appctx.App, userID int) []Site {
	rows, err := a.DB.QueryContext(ctx, `SELECT s.site_name, COALESCE(s.type, ''), COALESCE(s.container, '')
		FROM sites s JOIN domains d ON d.domain_id = s.domain_id WHERE d.user_id = ?`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var sites []Site
	for rows.Next() {
		var s Site
		if rows.Scan(&s.Name, &s.Type, &s.Container) == nil {
			sites = append(sites, s)
		}
	}
	return sites
}

// buildTargets lists every domain with its subfolder sites right under it
func buildTargets(domains []appctx.Domain, sites []Site) []Target {
	subs := map[string][]string{}
	for _, s := range sites {
		if dom, _, ok := strings.Cut(s.Name, "/"); ok {
			subs[dom] = append(subs[dom], s.Name)
		}
	}
	names := make([]string, 0, len(domains))
	for _, d := range domains {
		names = append(names, d.DomainURL)
	}
	sort.Strings(names)
	var out []Target
	for _, n := range names {
		out = append(out, Target{Value: n})
		sort.Strings(subs[n])
		for _, s := range subs[n] {
			out = append(out, Target{Value: s, Sub: true})
		}
	}
	return out
}

// siteDocroots adds a docroot for each subfolder site so its PHP errors aren't credited to the parent domain
func siteDocroots(domains []appctx.Domain, sites []Site) []appctx.Domain {
	docroots := map[string]string{}
	for _, d := range domains {
		docroots[d.DomainURL] = strings.TrimRight(d.Docroot, "/")
	}
	out := append([]appctx.Domain{}, domains...)
	for _, s := range sites {
		dom, sub, ok := strings.Cut(s.Name, "/")
		if root := docroots[dom]; ok && root != "" && sub != "" {
			out = append(out, appctx.Domain{DomainURL: s.Name, Docroot: root + "/" + strings.Trim(sub, "/")})
		}
	}
	return out
}

// appSources maps each existing app or proxied container to the site it serves
func appSources(userContext string, sites []Site, existing map[string]bool) []Source {
	seen := map[string]bool{}
	var out []Source
	add := func(container, site, typ string) {
		container = strings.ToLower(container)
		if container == "" || seen[container] || !existing[container] {
			return
		}
		seen[container] = true
		label := appTypeLabels[typ]
		if label == "" {
			label = "Container"
		}
		out = append(out, Source{Name: container, Label: label + " (" + site + ")", kind: "app", Site: site})
	}
	for _, s := range sites {
		if _, ok := appTypeLabels[s.Type]; ok {
			add(s.Container, s.Name, s.Type)
		}
	}
	// custom containers proxied to a domain may have no sites row
	if compose, err := docker.LoadCompose(userContext); err == nil {
		services, _ := compose["services"].(map[string]any)
		for name, raw := range services {
			svc, _ := raw.(map[string]any)
			if d := docker.ServiceDomain(svc); d != "" {
				cname, _ := svc["container_name"].(string)
				if cname == "" {
					cname = name
				}
				add(cname, d, "Docker")
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

var (
	appErrorRE    = regexp.MustCompile(`(?i)\b(error|exception|fatal|panic|traceback|critical|failed|refused|warn(ing)?)\b|(?i:uncaught|unhandled)`)
	appErrnoRE    = regexp.MustCompile(`\bE[A-Z]{4,}\b`)
	appCriticalRE = regexp.MustCompile(`(?i)\b(fatal|panic|critical)\b|(?i:uncaught|unhandled)`)
	appWarningRE  = regexp.MustCompile(`(?i)\bwarn(ing)?\b`)
	appStackRE    = regexp.MustCompile(`^(\s+\S|Caused by|File ")`)
	// the last line of a python traceback, like "ZeroDivisionError: division by zero"
	pyExceptionRE = regexp.MustCompile(`^[\w.]+(Error|Exception|Interrupt|Exit)\b`)
)

// parseApp keeps the error-looking lines of an app's output plus the stack trace lines right after them
func parseApp(service, site, out string) []Entry {
	var entries []Entry
	keep := false
	for _, line := range strings.Split(out, "\n") {
		ts, text := splitPodmanTS(strings.TrimRight(line, "\r"))
		if strings.TrimSpace(text) == "" {
			continue
		}
		if keep {
			last := &entries[len(entries)-1]
			if appStackRE.MatchString(text) {
				last.Message += "\n" + text
				continue
			}
			if strings.HasPrefix(last.Message, "Traceback") && pyExceptionRE.MatchString(text) {
				last.Message += "\n" + text
				keep = false
				continue
			}
		}
		keep = false
		if !appErrorRE.MatchString(text) && !appErrnoRE.MatchString(text) {
			continue
		}
		lvl := "error"
		if appCriticalRE.MatchString(text) {
			lvl = "critical"
		} else if appWarningRE.MatchString(text) {
			lvl = "warning"
		}
		entries = append(entries, Entry{Time: ts, Source: "app", Service: service, Domain: site, Level: lvl, Message: strings.TrimSpace(text)})
		keep = true
	}
	return entries
}

// subsiteFor picks the subfolder site a request path belongs to, longest prefix wins
func subsiteFor(domain, uri string, sites []Site) string {
	best := ""
	for _, s := range sites {
		dom, sub, ok := strings.Cut(s.Name, "/")
		if !ok || dom != domain {
			continue
		}
		prefix := "/" + strings.Trim(sub, "/")
		if (uri == prefix || strings.HasPrefix(uri, prefix+"/") || strings.HasPrefix(uri, prefix+"?")) && len(s.Name) > len(best) {
			best = s.Name
		}
	}
	return best
}

// inTarget is true when an entry belongs to the picked domain or site, a domain also covers its subfolder sites
func inTarget(entryDomain, target string) bool {
	return target == "" || entryDomain == target || strings.HasPrefix(entryDomain, target+"/")
}
