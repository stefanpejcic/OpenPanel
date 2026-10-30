package errorlogs

import (
	"context"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

const (
	tailPerContainer = 2000
	caddyTailBytes   = 2 << 20
	caddyService     = "caddy"
	olsErrorLog      = "/usr/local/lsws/logs/error.log"
	olsTailLines     = 5000
)

// caddyLogDir is a var so tests can point it at a temp dir
var caddyLogDir = "/var/log/caddy/domlogs/"

// Source is one place errors get read from, shown in the service filter
type Source struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Site  string `json:"site,omitempty"`
	kind  string
}

// Query is the parsed filter set shared by the page, the txt export and the API
type Query struct {
	Domain, Service, Level, Search string
	// Severity keeps one exact level, Level is "this level and up"
	Severity       string
	FromDay, ToDay string
	From, To       time.Time
	TZ, Limit      int
	Exact          bool
}

// Result is what a Query returns after merging and filtering every source
type Result struct {
	Entries   []Entry  `json:"entries"`
	Total     int      `json:"total"`
	Truncated bool     `json:"truncated"`
	Sources   []Source `json:"sources"`
}

func sourceKind(name, webserver, mysqlType string) (kind, label string) {
	switch {
	case strings.HasPrefix(name, "php-fpm-"):
		return "php", "PHP " + strings.TrimPrefix(name, "php-fpm-")
	case name != webserver && (name == "nginx" || name == "openresty" || name == "apache" || name == "openlitespeed" || name == "litespeed"):
		return "", ""
	case name == "nginx":
		return "nginx", "Nginx"
	case name == "openresty":
		return "nginx", "OpenResty"
	case name == "apache":
		return "apache", "Apache"
	case name == "openlitespeed":
		return "ols", "OpenLiteSpeed"
	case name == "litespeed":
		return "ols", "LiteSpeed"
	case name == mysqlType && name == "mysql":
		return "mysql", "MySQL"
	case name == mysqlType && name == "mariadb":
		return "mysql", "MariaDB"
	case name == "postgres":
		return "postgres", "PostgreSQL"
	}
	return "", ""
}

// listSources finds which of the user's containers actually exist and are worth reading, app containers last
func listSources(ctx context.Context, userContext string, sites []Site) []Source {
	argv := podmanmanager.PodmanArgv(userContext, "ps", "-a", "--format", "{{.Names}}")
	cmdCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, _ := podmanmanager.Command(cmdCtx, userContext, argv).Output()

	webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE")

	var sources []Source
	existing := map[string]bool{}
	for _, name := range strings.Fields(string(out)) {
		existing[name] = true
		if kind, label := sourceKind(name, webserver, mysqlType); kind != "" {
			sources = append(sources, Source{Name: name, Label: label, kind: kind})
		}
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Name < sources[j].Name })
	sources = append(sources, appSources(userContext, sites, existing)...)
	return append(sources, Source{Name: caddyService, Label: "HTTP 5xx responses", kind: "caddy"})
}

// cacheWindow is the picked range as typed, so the 30s cache gets hits while "now" keeps moving
func (q Query) cacheWindow() string {
	return q.FromDay + "_" + q.ToDay + "_" + strconv.Itoa(q.TZ) + "_" + q.From.Format(time.RFC3339) + "_" + q.To.Format(time.RFC3339)
}

func fetchContainer(ctx context.Context, a *appctx.App, userContext string, src Source, q Query) []Entry {
	key := "_error_logs:" + userContext + ":" + src.Name + ":" + q.cacheWindow()
	out, _ := cache.Memoize(ctx, a.Cache, key, 30*time.Second, func() (string, error) {
		args := []string{"logs", "--timestamps"}
		if !q.From.IsZero() {
			args = append(args, "--since", q.From.Format(time.RFC3339))
		}
		if !q.To.IsZero() {
			args = append(args, "--until", q.To.Format(time.RFC3339))
		}
		args = append(args, "--tail", strconv.Itoa(tailPerContainer), src.Name)
		if src.kind == "ols" {
			// OLS keeps everything in its own error.log, podman logs only has the entrypoint output
			args = []string{"exec", src.Name, "tail", "-n", strconv.Itoa(olsTailLines), olsErrorLog}
		}
		cmdCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		b, _ := podmanmanager.Command(cmdCtx, userContext, podmanmanager.PodmanArgv(userContext, args...)).CombinedOutput()
		return string(b), nil
	})
	switch src.kind {
	case "ols":
		return parseOLS(src.Name, out, q.From, q.To)
	case "app":
		return parseApp(src.Name, src.Site, out)
	}
	return parseLines(src.kind, src.Name, out)
}

// tailFile reads at most the last n bytes of path, dropping the first partial line
func tailFile(path string, n int64) string {
	f, err := tenantfs.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	offset := info.Size() - n
	if offset < 0 {
		offset = 0
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return ""
	}
	b, _ := io.ReadAll(f)
	s := string(b)
	if offset > 0 {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		}
	}
	return s
}

func fetchCaddy(domains []appctx.Domain, sites []Site, q Query) []Entry {
	var entries []Entry
	for _, d := range domains {
		if q.Domain != "" && d.DomainURL != q.Domain && !strings.HasPrefix(q.Domain, d.DomainURL+"/") {
			continue
		}
		content := tailFile(caddyLogDir+d.DomainURL+"/access.log", caddyTailBytes)
		for _, e := range parseCaddy5xx(d.DomainURL, content, q.From, q.To) {
			if sub := subsiteFor(d.DomainURL, e.uri, sites); sub != "" {
				e.Domain = sub
			}
			entries = append(entries, e)
		}
	}
	return entries
}

// Collect reads every source in parallel, tags domains, filters and sorts newest first
func Collect(ctx context.Context, a *appctx.App, userContext string, domains []appctx.Domain, sites []Site, q Query) Result {
	sources := listSources(ctx, userContext, sites)

	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		all []Entry
	)
	for _, src := range sources {
		if q.Service != "" && q.Service != src.Name {
			continue
		}
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			var got []Entry
			if src.kind == "caddy" {
				got = fetchCaddy(domains, sites, q)
			} else {
				got = fetchContainer(ctx, a, userContext, src, q)
			}
			mu.Lock()
			all = append(all, got...)
			mu.Unlock()
		}(src)
	}
	wg.Wait()

	matcher := newDomainMatcher(siteDocroots(domains, sites))
	matcher.backends = appBackends(sources)
	res := filterEntries(all, matcher, q)
	res.Sources = sources
	return res
}

func filterEntries(all []Entry, matcher domainMatcher, q Query) Result {
	minRank := map[string]int{"warning": levelRank["warning"], "error": levelRank["error"]}[q.Level]
	search := strings.ToLower(q.Search)

	kept := all[:0]
	for _, e := range all {
		if e.Domain == "" {
			e.Domain = matcher.match(e.Message)
		}
		if q.Exact && e.Domain != q.Domain || !q.Exact && !inTarget(e.Domain, q.Domain) {
			continue
		}
		if levelRank[e.Level] < minRank || q.Severity != "" && e.Level != q.Severity {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(e.Message), search) {
			continue
		}
		kept = append(kept, e)
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Time.After(kept[j].Time) })

	res := Result{Total: len(kept), Entries: kept}
	if q.Limit > 0 && len(kept) > q.Limit {
		res.Entries, res.Truncated = kept[:q.Limit], true
	}
	if res.Entries == nil {
		res.Entries = []Entry{}
	}
	return res
}

// appBackends maps each app container to its site so proxy errors naming the container land on that site
func appBackends(sources []Source) map[string]string {
	out := map[string]string{}
	for _, s := range sources {
		if s.kind == "app" && s.Site != "" {
			out[strings.ToLower(s.Name)] = s.Site
		}
	}
	return out
}
