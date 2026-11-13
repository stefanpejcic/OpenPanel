package botcontrol

import (
	"bufio"
	"encoding/json"
	"io"
	"net/netip"
	"os"
	"sort"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

const (
	maxLogTailBytes = 8 * 1024 * 1024
	maxTopRows      = 25
)

// TrafficReport holds the aggregated "Who's hitting me" breakdown for one domain.
type TrafficReport struct {
	Domain         string       `json:"domain"`
	Window         string       `json:"window"`
	TotalRequests  int          `json:"total_requests"`
	AIRequests     int          `json:"ai_requests"`
	BlockedCount   int          `json:"blocked_count"`
	TopBots        []BotStat    `json:"top_bots"`
	TopNetworks    []NetStat    `json:"top_networks"`
	TopURLs        []URLStat    `json:"top_urls"`
	LogMissing     bool         `json:"log_missing"`
}

// BotStat ranks traffic by classified bot or user-agent.
type BotStat struct {
	Name       string      `json:"name"`
	Operator   string      `json:"operator"`
	Category   BotCategory `json:"category"`
	Requests   int         `json:"requests"`
	Blocked    int         `json:"blocked"`
	SharePct   float64     `json:"share_pct"`
}

// NetStat ranks traffic by ASN (when available) or /24 / /48 IP network.
type NetStat struct {
	Network   string `json:"network"`
	ASN       string `json:"asn"`
	Requests  int    `json:"requests"`
	Blocked   int    `json:"blocked"`
	UniqueIPs int    `json:"unique_ips"`
}

// URLStat ranks the most-hit URLs and how much of their traffic is bot traffic.
type URLStat struct {
	Path        string `json:"path"`
	Requests    int    `json:"requests"`
	BotRequests int    `json:"bot_requests"`
	Blocked     int    `json:"blocked"`
}

type caddyLogLine struct {
	TS      float64 `json:"ts"`
	Status  int     `json:"status"`
	Request struct {
		ClientIP string              `json:"client_ip"`
		URI      string              `json:"uri"`
		Headers  map[string][]string `json:"headers"`
	} `json:"request"`
}

func domainAccessLogPath(domainName string) string {
	return "/var/log/caddy/domlogs/" + domainName + "/access.log"
}

func parseWindowDuration(window string) (string, time.Duration) {
	switch window {
	case "1h":
		return "1h", time.Hour
	case "7d":
		return "7d", 7 * 24 * time.Hour
	default:
		return "24h", 24 * time.Hour
	}
}

// analyzeDomainTraffic reads the tail of a domain's Caddy JSON log and ranks traffic by bot, network/ASN, and URL.
func analyzeDomainTraffic(domainName, window string, now time.Time) (TrafficReport, error) {
	winLabel, dur := parseWindowDuration(window)
	report := TrafficReport{
		Domain:      domainName,
		Window:      winLabel,
		TopBots:     []BotStat{},
		TopNetworks: []NetStat{},
		TopURLs:     []URLStat{},
	}

	f, err := tenantfs.Open(domainAccessLogPath(domainName))
	if err != nil {
		if os.IsNotExist(err) {
			report.LogMissing = true
			return report, nil
		}
		return report, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return report, err
	}
	if info.Size() == 0 {
		return report, nil
	}

	// Read at most maxLogTailBytes from the end so huge logs don't spike memory or CPU.
	var r io.Reader = f
	if info.Size() > maxLogTailBytes {
		if _, seekErr := f.Seek(-maxLogTailBytes, io.SeekEnd); seekErr == nil {
			br := bufio.NewReader(f)
			_, _ = br.ReadString('\n') // drop partial first line after seek
			r = br
		}
	}

	cutoff := float64(now.Add(-dur).UnixNano()) / 1e9
	asnTable := loadASNPrefixes("/etc/openpanel/asn_prefixes.tsv")

	return aggregateLogReader(r, domainName, winLabel, cutoff, asnTable)
}

func aggregateLogReader(r io.Reader, domainName, winLabel string, cutoff float64, asnTable []asnPrefixEntry) (TrafficReport, error) {
	report := TrafficReport{
		Domain: domainName,
		Window: winLabel,
	}

	type botKey struct {
		name     string
		operator string
		cat      BotCategory
	}
	type netAcc struct {
		network string
		asn     string
		reqs    int
		blocked int
		ips     map[string]struct{}
	}
	type urlAcc struct {
		reqs    int
		botReqs int
		blocked int
	}

	bots := make(map[botKey]*BotStat)
	nets := make(map[string]*netAcc)
	urls := make(map[string]*urlAcc)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry caddyLogLine
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		if entry.TS > 0 && entry.TS < cutoff {
			continue
		}

		report.TotalRequests++
		isBlocked := entry.Status == 403 || entry.Status == 429
		if isBlocked {
			report.BlockedCount++
		}

		ua := firstHeader(entry.Request.Headers, "User-Agent")
		botName, op, cat, isAI := classifyUserAgent(ua)
		if isAI {
			report.AIRequests++
		}

		bk := botKey{name: botName, operator: op, cat: cat}
		bs, ok := bots[bk]
		if !ok {
			bs = &BotStat{Name: botName, Operator: op, Category: cat}
			bots[bk] = bs
		}
		bs.Requests++
		if isBlocked {
			bs.Blocked++
		}

		ipStr := strings.TrimSpace(entry.Request.ClientIP)
		netLabel, asnLabel := lookupNetworkAndASN(ipStr, asnTable)
		na, ok := nets[netLabel]
		if !ok {
			na = &netAcc{network: netLabel, asn: asnLabel, ips: make(map[string]struct{})}
			nets[netLabel] = na
		}
		na.reqs++
		if isBlocked {
			na.blocked++
		}
		if ipStr != "" {
			na.ips[ipStr] = struct{}{}
		}

		path := normalizePath(entry.Request.URI)
		uaAcc, ok := urls[path]
		if !ok {
			uaAcc = &urlAcc{}
			urls[path] = uaAcc
		}
		uaAcc.reqs++
		if cat != CatHuman {
			uaAcc.botReqs++
		}
		if isBlocked {
			uaAcc.blocked++
		}
	}

	report.TopBots = make([]BotStat, 0, len(bots))
	for _, b := range bots {
		if report.TotalRequests > 0 {
			b.SharePct = float64(int(float64(b.Requests)*1000/float64(report.TotalRequests)+0.5)) / 10
		}
		report.TopBots = append(report.TopBots, *b)
	}
	sort.Slice(report.TopBots, func(i, j int) bool {
		if report.TopBots[i].Requests != report.TopBots[j].Requests {
			return report.TopBots[i].Requests > report.TopBots[j].Requests
		}
		return report.TopBots[i].Name < report.TopBots[j].Name
	})
	if len(report.TopBots) > maxTopRows {
		report.TopBots = report.TopBots[:maxTopRows]
	}

	report.TopNetworks = make([]NetStat, 0, len(nets))
	for _, n := range nets {
		report.TopNetworks = append(report.TopNetworks, NetStat{
			Network:   n.network,
			ASN:       n.asn,
			Requests:  n.reqs,
			Blocked:   n.blocked,
			UniqueIPs: len(n.ips),
		})
	}
	sort.Slice(report.TopNetworks, func(i, j int) bool {
		if report.TopNetworks[i].Requests != report.TopNetworks[j].Requests {
			return report.TopNetworks[i].Requests > report.TopNetworks[j].Requests
		}
		return report.TopNetworks[i].Network < report.TopNetworks[j].Network
	})
	if len(report.TopNetworks) > maxTopRows {
		report.TopNetworks = report.TopNetworks[:maxTopRows]
	}

	report.TopURLs = make([]URLStat, 0, len(urls))
	for p, u := range urls {
		report.TopURLs = append(report.TopURLs, URLStat{
			Path:        p,
			Requests:    u.reqs,
			BotRequests: u.botReqs,
			Blocked:     u.blocked,
		})
	}
	sort.Slice(report.TopURLs, func(i, j int) bool {
		if report.TopURLs[i].Requests != report.TopURLs[j].Requests {
			return report.TopURLs[i].Requests > report.TopURLs[j].Requests
		}
		return report.TopURLs[i].Path < report.TopURLs[j].Path
	})
	if len(report.TopURLs) > maxTopRows {
		report.TopURLs = report.TopURLs[:maxTopRows]
	}

	return report, scanner.Err()
}

func firstHeader(headers map[string][]string, name string) string {
	for k, vals := range headers {
		if strings.EqualFold(k, name) && len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

func normalizePath(rawURI string) string {
	if rawURI == "" {
		return "/"
	}
	if idx := strings.IndexByte(rawURI, '?'); idx >= 0 {
		rawURI = rawURI[:idx]
	}
	if rawURI == "" {
		return "/"
	}
	if len(rawURI) > 120 {
		return rawURI[:120] + "..."
	}
	return rawURI
}

type asnPrefixEntry struct {
	prefix netip.Prefix
	asn    string
}

func loadASNPrefixes(path string) []asnPrefixEntry {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []asnPrefixEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		pfx, err := netip.ParsePrefix(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		label := strings.TrimSpace(parts[1])
		if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
			label = label + " (" + strings.TrimSpace(parts[2]) + ")"
		}
		entries = append(entries, asnPrefixEntry{prefix: pfx, asn: label})
	}
	return entries
}

func lookupNetworkAndASN(ipStr string, table []asnPrefixEntry) (network string, asn string) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		if ipStr == "" {
			return "Unknown", "-"
		}
		return ipStr, "-"
	}
	for _, e := range table {
		if e.prefix.Contains(addr) {
			return e.prefix.String(), e.asn
		}
	}
	if addr.Is4() {
		return netip.PrefixFrom(addr, 24).Masked().String(), "-"
	}
	return netip.PrefixFrom(addr, 48).Masked().String(), "-"
}
