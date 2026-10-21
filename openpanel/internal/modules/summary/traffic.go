package summary

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/errorlogs"
)

// both are vars so tests can point them at a temp dir
var (
	caddyLogDir = "/var/log/caddy/domlogs/"
	wafLogDir   = "/var/log/caddy/coraza_waf/"
)

// accessLine is the subset of a Caddy JSON access-log line the summary counts
type accessLine struct {
	TS      float64 `json:"ts"`
	Status  int     `json:"status"`
	Size    int64   `json:"size"`
	Request struct {
		ClientIP string `json:"client_ip"`
		URI      string `json:"uri"`
	} `json:"request"`
}

// logFiles are the domain's current log plus rotated ones that may still hold lines from the day, oldest first
func logFiles(dir string, from time.Time) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, "access*.log*"))
	type file struct {
		path string
		mod  time.Time
	}
	var files []file
	for _, m := range matches {
		info, err := tenantfs.Stat(m)
		if err != nil || info.ModTime().Before(from) {
			continue
		}
		files = append(files, file{m, info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.path
	}
	return out
}

// eachLine calls fn for every line of a plain or gzipped log
func eachLine(path string, fn func([]byte)) {
	f, err := tenantfs.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return
		}
		defer gz.Close()
		r = gz
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		fn(sc.Bytes())
	}
}

// collectTraffic counts requests, unique IPs, bytes, 404s, 5xx and 429s per domain and subfolder site
func (b *builder) collectTraffic() {
	for _, d := range b.domains {
		for _, path := range logFiles(caddyLogDir+d.DomainURL, b.from) {
			eachLine(path, func(raw []byte) {
				var l accessLine
				if json.Unmarshal(raw, &l) != nil {
					return
				}
				ts := time.Unix(int64(l.TS), 0)
				if ts.Before(b.from) || !ts.Before(b.to) {
					return
				}
				name := d.DomainURL
				if sub := errorlogs.SubsiteFor(d.DomainURL, l.Request.URI, b.sites); sub != "" {
					name = sub
				}
				w := b.website(name)
				w.Requests++
				w.Bytes += l.Size
				w.ips[l.Request.ClientIP]++
				switch {
				case l.Status == 404:
					w.NotFound++
				case l.Status == 429:
					w.RateLimited++
					w.limitedIPs[l.Request.ClientIP]++
				case l.Status >= 500:
					w.ServerErrors++
				case l.Status < 400:
					page, _, _ := strings.Cut(l.Request.URI, "?")
					if len(w.pages) < 5000 || w.pages[page] > 0 {
						w.pages[page]++
					}
				}
			})
		}
	}
	b.rateLimitItems()
}

// rateLimitItems explains 429s, one IP doing most of them is a bot worth blocking, many IPs means real traffic hit the limit
func (b *builder) rateLimitItems() {
	for _, d := range b.domains {
		w, ok := b.buckets[d.DomainURL]
		if !ok || w.RateLimited == 0 {
			continue
		}
		ip := topKeys(w.limitedIPs, 1)[0]
		share := w.limitedIPs[ip] * 100 / w.RateLimited
		it := Item{}
		switch {
		case isLocalIP(ip):
			it.Title = fmt.Sprintf("Something on your own server called %s too often", w.Name)
			it.Detail = fmt.Sprintf("%s requests from %s were turned away for going over the request limit. That's usually a cron job or a plugin calling the site.", num(w.RateLimited), ip)
			if b.can("crons") {
				it.Action, it.Link = "Check cron jobs that call the site", "/cronjobs"
			} else {
				it.Action = "Check plugins or scripts that call the site, or ask your hosting provider to raise the rate limit for " + w.Name + "."
			}
		case share >= 50:
			it.Title = fmt.Sprintf("A bot hit %s hard and was slowed down", w.Name)
			it.Detail = fmt.Sprintf("%s made %s of the %s requests that went over the limit, so it looks like a bot or a scraper, not a visitor.", ip, num(w.limitedIPs[ip]), num(w.RateLimited))
			if b.can("ip_blocker") {
				it.Action, it.Link = "Block "+ip, "/security/ip-blocker"
			} else {
				it.Action = "Ask your hosting provider to block " + ip + " if it keeps coming back."
			}
		default:
			it.Title = fmt.Sprintf("%s turned away %s requests during a busy moment", w.Name, num(w.RateLimited))
			it.Detail = fmt.Sprintf("They came from %s, so real visitors may have seen a 'too many requests' error.", plural(len(w.limitedIPs), "visitor", "different visitors"))
			it.Action = "If this keeps happening, ask your hosting provider to raise the rate limit for " + w.Name + "."
		}
		b.look(it)
	}
}

// isLocalIP is the server or a private network, blocking those would break the site
func isLocalIP(ip string) bool {
	p := net.ParseIP(ip)
	return p == nil || p.IsLoopback() || p.IsPrivate() || p.IsUnspecified()
}

// wafEntry is the subset of a Coraza audit log line needed to count blocks
type wafEntry struct {
	Transaction struct {
		UnixTimestamp int64  `json:"unix_timestamp"`
		ClientIP      string `json:"client_ip"`
		Request       struct {
			URI string `json:"uri"`
		} `json:"request"`
		IsInterrupted bool `json:"is_interrupted"`
	} `json:"transaction"`
}

// collectWAF counts requests the web firewall blocked, a working firewall is good news so it only adds to the numbers
func (b *builder) collectWAF() {
	if !b.can("waf") {
		return
	}
	total := 0
	for _, d := range b.domains {
		eachLine(wafLogDir+d.DomainURL+".log", func(raw []byte) {
			// most lines are flagged, not blocked, skip those before decoding
			if !bytes.Contains(raw, []byte(`"is_interrupted":true`)) {
				return
			}
			var e wafEntry
			if json.Unmarshal(raw, &e) != nil || !e.Transaction.IsInterrupted {
				return
			}
			ts := time.Unix(0, e.Transaction.UnixTimestamp)
			if ts.Before(b.from) || !ts.Before(b.to) {
				return
			}
			name := d.DomainURL
			if sub := errorlogs.SubsiteFor(d.DomainURL, e.Transaction.Request.URI, b.sites); sub != "" {
				name = sub
			}
			w := b.website(name)
			w.WAFBlocked++
			w.wafIPs[e.Transaction.ClientIP]++
			total++
		})
	}
	if total > 0 {
		b.good(fmt.Sprintf("The web firewall stopped %s on your websites.", plural(total, "attack", "attacks")))
	}
}

// checkErrors ties server errors to their cause from the Error Logs, without that module the user only gets the count
func (b *builder) checkErrors(ctx context.Context) {
	if b.can("error_logs") {
		q := errorlogs.Query{Level: "error", From: b.from.UTC()}
		// podman logs returns nothing for an end in the future
		if b.to.Before(time.Now()) {
			q.To = b.to.UTC()
		}
		res := errorlogs.Collect(ctx, b.a, b.userContext, b.domains, b.sites, q)
		for _, e := range res.Entries {
			// 5xx lines are already counted from the access log, the cause is in the PHP, web server or app errors
			if e.Domain == "" || e.Source == "http" {
				continue
			}
			w := b.website(e.Domain)
			w.errorCount++
			// entries are newest first, keep the newest as the example
			if w.errorExample == "" {
				msg, _, _ := strings.Cut(e.Message, "\n")
				if len(msg) > 200 {
					msg = msg[:200] + "..."
				}
				w.errorExample = msg
			}
		}
	}

	names := make([]string, 0, len(b.buckets))
	for name := range b.buckets {
		names = append(names, name)
	}
	sort.Strings(names)
	// error pages without a known cause read the same for every site, so 3 or more get one item
	plain := map[bool][]*Website{}
	for _, name := range names {
		w := b.buckets[name]
		if w.ServerErrors == 0 && w.errorCount == 0 {
			continue
		}
		urgent := w.ServerErrors >= 10 || w.Requests > 0 && w.ServerErrors*100 >= w.Requests
		if w.ServerErrors > 0 && w.errorExample == "" {
			plain[urgent] = append(plain[urgent], w)
			continue
		}
		it := Item{}
		if w.ServerErrors > 0 {
			it.Title = fmt.Sprintf("Visitors saw an error page on %s (%s)", name, plural(w.ServerErrors, "time", "times"))
			it.Detail = "The server returned an error instead of the page."
		} else {
			it.Title = fmt.Sprintf("%s showed up in the logs of %s", plural(w.errorCount, "error", "errors"), name)
		}
		switch {
		case w.errorExample != "":
			it.Detail = strings.TrimSpace(it.Detail + " The latest error was: " + w.errorExample)
			it.Action, it.Link = "See what went wrong", "/error-logs?domain="+name+"&exact=1"
		case b.can("domain_logs"):
			domain, _, _ := strings.Cut(name, "/")
			it.Action, it.Link = "Check the visits that failed", "/domains/log/"+domain
		default:
			it.Action = "If visitors keep getting errors, contact your hosting provider."
		}
		// a few errors happen, many or a share of real traffic means the site is broken for people
		if urgent {
			b.attention(it)
		} else {
			b.look(it)
		}
	}
	for _, urgent := range []bool{true, false} {
		for _, it := range b.errorPageItems(plain[urgent]) {
			if urgent {
				b.attention(it)
			} else {
				b.look(it)
			}
		}
	}
}

// errorPageItems is one item per site, or one for all of them when there are 3 or more
func (b *builder) errorPageItems(sites []*Website) []Item {
	link := func(site string) (string, string) {
		switch {
		case b.can("error_logs") && site == "":
			return "See what went wrong", "/error-logs"
		case b.can("error_logs"):
			return "See what went wrong", "/error-logs?domain=" + site + "&exact=1"
		case b.can("domain_logs") && site == "":
			return "Check the visits that failed", "/domains/log"
		case b.can("domain_logs"):
			domain, _, _ := strings.Cut(site, "/")
			return "Check the visits that failed", "/domains/log/" + domain
		}
		return "If visitors keep getting errors, contact your hosting provider.", ""
	}
	if len(sites) < 3 {
		var items []Item
		for _, w := range sites {
			it := Item{Title: fmt.Sprintf("Visitors saw an error page on %s (%s)", w.Name, plural(w.ServerErrors, "time", "times")),
				Detail: "The server returned an error instead of the page."}
			it.Action, it.Link = link(w.Name)
			items = append(items, it)
		}
		return items
	}
	sort.SliceStable(sites, func(i, j int) bool { return sites[i].ServerErrors > sites[j].ServerErrors })
	var parts []string
	for i, w := range sites {
		if i == 0 {
			parts = append(parts, fmt.Sprintf("%s (%s)", w.Name, plural(w.ServerErrors, "time", "times")))
		} else {
			parts = append(parts, fmt.Sprintf("%s (%s)", w.Name, num(w.ServerErrors)))
		}
	}
	it := Item{Title: fmt.Sprintf("Visitors saw error pages on %d websites", len(sites)),
		Detail: "On " + joinAnd(parts) + ". The server returned an error instead of the page."}
	it.Action, it.Link = link("")
	return []Item{it}
}
