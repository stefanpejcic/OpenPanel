package websites

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"golang.org/x/net/idna"
)

// wpCacheSite is everything the caching tab needs to know about one WordPress install
type wpCacheSite struct {
	Username     string
	UserContext  string
	Domain       string // ascii root domain
	Site         string // domain or domain/folder
	Folder       string
	Docroot      string // container path
	HostPath     string
	PHPVersion   string
	PHPContainer string
	Webserver    string
	Litespeed    bool
}

var (
	wpCacheDomainRE    = regexp.MustCompile(`^[a-z0-9.-]+$`)
	wpCacheContainerRE = regexp.MustCompile(`^[a-z0-9._-]+$`)
	redisPrefixRE      = regexp.MustCompile(`[^a-z0-9]+`)
	vclTTLLineRE       = regexp.MustCompile(`bereq\.http\.host == "([a-z0-9.-]+)".*set beresp\.ttl = (\d+)s;`)
)

const (
	vclTTLBegin       = "    # openpanel-site-ttl-begin"
	vclTTLEnd         = "    # openpanel-site-ttl-end"
	vclDefaultTTL     = 3600
	userIniMarker     = "; openpanel: opcache disabled for this site"
	redisDropinMarker = "Redis Object Cache Drop-In"
)

var wpCacheTTLChoices = map[int]bool{60: true, 300: true, 900: true, 1800: true, 3600: true, 21600: true, 86400: true}

// resolveWPCacheSite checks ownership and works out docroot, php container and webserver for a WordPress site
func resolveWPCacheSite(a *appctx.App, r *http.Request, siteParam string) (*wpCacheSite, int, error) {
	ctx := r.Context()
	userID, username, userContext, err := injected(a, r)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("internal error")
	}
	siteParam = strings.Trim(strings.TrimSpace(siteParam), "/")
	if siteParam == "" {
		return nil, http.StatusBadRequest, errors.New("domain is required")
	}
	rawDomain, folder := splitDomainAndFolder(siteParam)
	if !isSafeWebsiteSubpath(folder) {
		return nil, http.StatusBadRequest, errors.New("invalid folder")
	}
	domain, err := idna.ToASCII(strings.ToLower(rawDomain))
	if err != nil || !wpCacheDomainRE.MatchString(domain) {
		return nil, http.StatusBadRequest, errors.New("invalid domain")
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, rawDomain) {
		return nil, http.StatusForbidden, errors.New("you do not own this domain")
	}

	var siteType sql.NullString
	if err := a.DB.QueryRowContext(ctx, "SELECT type FROM sites WHERE site_name = ? LIMIT 1", siteParam).Scan(&siteType); err != nil {
		return nil, http.StatusNotFound, errors.New("site not found")
	}
	if !strings.EqualFold(siteType.String, "wordpress") {
		return nil, http.StatusBadRequest, errors.New("caching controls are only available for WordPress sites")
	}

	var docroot string
	if err := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domain).Scan(&docroot); err != nil {
		return nil, http.StatusNotFound, errors.New("unable to detect docroot for the domain")
	}
	if folder != "" {
		docroot = strings.TrimRight(docroot, "/") + "/" + folder
	}
	docroot = filepath.Clean(docroot)
	if !strings.HasPrefix(docroot, "/var/www/html/") {
		return nil, http.StatusBadRequest, errors.New("invalid document root")
	}

	s := &wpCacheSite{
		Username: username, UserContext: userContext,
		Domain: domain, Site: siteParam, Folder: folder, Docroot: docroot,
		HostPath: "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/"),
	}
	s.Webserver = strings.TrimSpace(webserver.GetEnvFileValue(userContext, "WEB_SERVER"))
	s.Litespeed = strings.Contains(strings.ToLower(s.Webserver), "litespeed")
	s.PHPVersion = php.GetPHPVForDomain(ctx, a, userContext, domain)
	if s.Litespeed {
		s.PHPContainer = s.Webserver
	} else if s.PHPVersion != "" && s.PHPVersion != "/" {
		s.PHPContainer = "php-fpm-" + s.PHPVersion
	}
	if !wpCacheContainerRE.MatchString(s.PHPContainer) || !wpCacheContainerRE.MatchString(s.Webserver) {
		return nil, http.StatusInternalServerError, errors.New("could not detect the PHP container for this site")
	}
	return s, http.StatusOK, nil
}

// userAllowed is the backend twin of the templates' .UserAllowed map: module on for the server AND in the user's plan
func userAllowed(ctx context.Context, a *appctx.App, s *wpCacheSite, module string) bool {
	if !a.ModuleEnabled(module) {
		return false
	}
	features, err := a.LoadUserFeatures(ctx, s.Username, s.UserContext)
	if err != nil {
		return false
	}
	for _, f := range features {
		if f == module {
			return true
		}
	}
	return false
}

func podmanOutput(ctx context.Context, userContext string, timeout time.Duration, args ...string) (string, error) {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := podmanmanager.Command(runCtx, userContext, podmanmanager.PodmanArgv(userContext, args...)).CombinedOutput()
	return string(out), err
}

func (s *wpCacheSite) wp(ctx context.Context, args ...string) (string, error) {
	argv := podmanmanager.BuildWPCLIBaseCommand(s.UserContext, s.PHPContainer)
	argv = append(argv, "--path="+s.Docroot, "--allow-root")
	argv = append(argv, args...)
	runCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	out, err := podmanmanager.Command(runCtx, s.UserContext, argv).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// redisPrefix keeps each site's keys apart inside the account's shared redis
func (s *wpCacheSite) redisPrefix() string {
	return "op_" + redisPrefixRE.ReplaceAllString(strings.ToLower(s.Site), "_") + ":"
}

// ---------------------- VARNISH TTL (VCL) ---------------------- //

func vclPath(userContext string) string { return "/home/" + userContext + "/default.vcl" }

// parseVCLTTLs returns the per-host lifetimes from our managed block
func parseVCLTTLs(vcl string) map[string]int {
	ttls := map[string]int{}
	begin := strings.Index(vcl, vclTTLBegin)
	end := strings.Index(vcl, vclTTLEnd)
	if begin == -1 || end == -1 || end < begin {
		return ttls
	}
	for _, line := range strings.Split(vcl[begin:end], "\n") {
		if m := vclTTLLineRE.FindStringSubmatch(line); m != nil {
			if v, err := strconv.Atoi(m[2]); err == nil {
				ttls[m[1]] = v
			}
		}
	}
	return ttls
}

// setVCLTTL rewrites the managed block with domain's lifetime, ttl 0 drops the override
func setVCLTTL(vcl, domain string, ttl int) (string, error) {
	if !wpCacheDomainRE.MatchString(domain) {
		return "", errors.New("invalid domain")
	}
	ttls := parseVCLTTLs(vcl)
	if ttl > 0 {
		ttls[domain] = ttl
	} else {
		delete(ttls, domain)
	}

	begin := strings.Index(vcl, vclTTLBegin)
	end := strings.Index(vcl, vclTTLEnd)
	if begin != -1 && end != -1 && end > begin {
		rest := vcl[end+len(vclTTLEnd):]
		for i := 0; i < 2 && strings.HasPrefix(rest, "\n"); i++ {
			rest = rest[1:]
		}
		vcl = vcl[:begin] + rest
	}
	if len(ttls) == 0 {
		return vcl, nil
	}

	// goes right before the grace setting in vcl_backend_response so it overrides the template's forced 1h
	anchor := strings.Index(vcl, "set beresp.grace")
	if anchor == -1 {
		return "", errors.New("custom VCL detected, could not find where to add the cache lifetime")
	}
	lineStart := strings.LastIndex(vcl[:anchor], "\n") + 1

	hosts := make([]string, 0, len(ttls))
	for h := range ttls {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	var b strings.Builder
	b.WriteString(vclTTLBegin + "\n")
	b.WriteString("    if (!beresp.http.Cache-Control && bereq.http.X-Static-File != \"true\") {\n")
	for _, h := range hosts {
		fmt.Fprintf(&b, "        if (bereq.http.host == \"%s\" || bereq.http.host == \"www.%s\") { set beresp.ttl = %ds; }\n", h, h, ttls[h])
	}
	b.WriteString("    }\n")
	b.WriteString(vclTTLEnd + "\n\n")
	return vcl[:lineStart] + b.String() + vcl[lineStart:], nil
}

// reloadVarnishVCL re-renders the template inside the container the same way its entrypoint does, then hot-loads it
func reloadVarnishVCL(ctx context.Context, userContext string) error {
	// pinned to one ipv4, vcc refuses a backend name that resolves to several addresses
	script := `cp /etc/varnish/default.vcl.template /etc/varnish/default.vcl && B=$(getent ahostsv4 "$WEB_SERVER" | awk 'NR==1{print $1}') && sed -i "s|VARNISH_BACKEND_HOST|${B:-$WEB_SERVER}|g" /etc/varnish/default.vcl && varnishreload`
	out, err := podmanOutput(ctx, userContext, 30*time.Second, "exec", "varnish", "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(out))
	}
	return nil
}

// banVarnishSite drops the site's cached pages, using the x-host/x-url headers the template stores on every object
func banVarnishSite(ctx context.Context, s *wpCacheSite) error {
	for _, host := range []string{s.Domain, "www." + s.Domain} {
		expr := []string{"obj.http.x-host", "==", host}
		if s.Folder != "" {
			expr = append(expr, "&&", "obj.http.x-url", "~", "^/"+regexp.QuoteMeta(s.Folder)+"(/|$)")
		}
		args := append([]string{"exec", "varnish", "varnishadm", "ban"}, expr...)
		if out, err := podmanOutput(ctx, s.UserContext, 15*time.Second, args...); err != nil {
			return fmt.Errorf("%v: %s", err, strings.TrimSpace(out))
		}
	}
	return nil
}

// ---------------------- STATUS ---------------------- //

type varnishCacheStatus struct {
	Allowed        bool    `json:"allowed"`
	ServiceRunning bool    `json:"service_running"`
	DomainStatus   string  `json:"domain_status"`
	Enabled        bool    `json:"enabled"`
	TTL            int     `json:"ttl"`
	TTLCustom      bool    `json:"ttl_custom"`
	HitRatio       float64 `json:"hit_ratio"`
	Hits           float64 `json:"hits"`
	Misses         float64 `json:"misses"`
	Objects        float64 `json:"objects"`
	// hits/misses/ratio above are this site's pages only, the account ones are for the tooltip
	AccountHitRatio float64 `json:"account_hit_ratio"`
}

type redisCacheStatus struct {
	Allowed        bool    `json:"allowed"`
	ServiceRunning bool    `json:"service_running"`
	Enabled        bool    `json:"enabled"`
	HitRatio       float64 `json:"hit_ratio"`
	Hits           float64 `json:"hits"`
	Misses         float64 `json:"misses"`
	UsedMemory     int64   `json:"used_memory"`
	MemoryLimit    int64   `json:"memory_limit"`
	SiteKeys       int64   `json:"site_keys"`
	SiteBytes      int64   `json:"site_bytes"`
	// false when the plugin's metrics are off, then hits/misses are account-wide
	SiteMetrics bool `json:"site_metrics"`
}

type opcacheStatus struct {
	Available     bool    `json:"available"`
	GlobalEnabled bool    `json:"global_enabled"`
	Enabled       bool    `json:"enabled"`
	HitRate       float64 `json:"hit_rate"`
	UsedMemory    int64   `json:"used_memory"`
	TotalMemory   int64   `json:"total_memory"`
	CachedScripts int64   `json:"cached_scripts"`
	PHPVersion    string  `json:"php_version"`
	Error         string  `json:"error,omitempty"`
}

type cacheMeasurement struct {
	CachedMS   float64        `json:"cached_ms"`
	DirectMS   float64        `json:"direct_ms"`
	CachedCode int            `json:"cached_code"`
	DirectCode int            `json:"direct_code"`
	ViaVarnish bool           `json:"via_varnish"`
	MeasuredAt int64          `json:"measured_at"`
	History    []measurePoint `json:"history,omitempty"`
}

type measurePoint struct {
	CachedMS   float64 `json:"cached_ms"`
	DirectMS   float64 `json:"direct_ms"`
	MeasuredAt int64   `json:"measured_at"`
}

const measureHistoryLen = 10

type wpCacheStatusResponse struct {
	Varnish     varnishCacheStatus `json:"varnish"`
	Redis       redisCacheStatus   `json:"redis"`
	Opcache     opcacheStatus      `json:"opcache"`
	Measurement *cacheMeasurement  `json:"measurement"`
}

func varnishStatusFor(ctx context.Context, a *appctx.App, s *wpCacheSite) varnishCacheStatus {
	v := varnishCacheStatus{Allowed: userAllowed(ctx, a, s, "varnish"), TTL: vclDefaultTTL}
	v.ServiceRunning = docker.IsServiceRunning(ctx, s.UserContext, "varnish")
	v.DomainStatus = cache.DomainVarnishStatus(s.Domain)
	v.Enabled = v.ServiceRunning && v.DomainStatus == "On"
	if content, err := os.ReadFile(vclPath(s.UserContext)); err == nil {
		if ttl, ok := parseVCLTTLs(string(content))[s.Domain]; ok {
			v.TTL, v.TTLCustom = ttl, true
		}
	}
	if v.ServiceRunning {
		v.AccountHitRatio, _, _, v.Objects = cache.VarnishHitStats(ctx, a, s.UserContext)
		v.Hits, v.Misses = varnishSiteHits(ctx, a, s)
		if v.Hits+v.Misses > 0 {
			v.HitRatio = v.Hits / (v.Hits + v.Misses)
		}
	}
	return v
}

// staticFileRE matches what the VCL treats as static, those aren't pages
const staticFileRE = `\\.(css|js|mjs|map|png|jpe?g|gif|webp|avif|svg|ico|woff2?|ttf|eot|otf|mp4|webm|mp3|pdf|zip)(\\?|$)`

// varnishSiteQuery builds a VSL query for this site's page requests, leaving out other sites in subfolders of the same domain
func varnishSiteQuery(s *wpCacheSite, otherFolders []string) string {
	q := `ReqHeader:Host ~ "^(www\\.)?` + strings.ReplaceAll(regexp.QuoteMeta(s.Domain), `\`, `\\`) + `$"`
	if s.Folder != "" {
		q += ` and ReqURL ~ "^/` + strings.ReplaceAll(regexp.QuoteMeta(s.Folder), `\`, `\\`) + `(/|\\?|$)"`
	} else if len(otherFolders) > 0 {
		quoted := make([]string, len(otherFolders))
		for i, f := range otherFolders {
			quoted[i] = strings.ReplaceAll(regexp.QuoteMeta(f), `\`, `\\`)
		}
		q += ` and not ReqURL ~ "^/(` + strings.Join(quoted, "|") + `)(/|\\?|$)"`
	}
	return q + ` and not ReqURL ~ "` + staticFileRE + `"`
}

// varnishSiteHits counts hits vs misses/passes for this site in varnish's in-memory request log
func varnishSiteHits(ctx context.Context, a *appctx.App, s *wpCacheSite) (hits, misses float64) {
	var others []string
	if s.Folder == "" {
		if rows, err := a.DB.QueryContext(ctx, "SELECT site_name FROM sites WHERE site_name LIKE ?", s.Domain+"/%"); err == nil {
			for rows.Next() {
				var name string
				if rows.Scan(&name) == nil {
					if _, folder := splitDomainAndFolder(name); folder != "" && isSafeWebsiteSubpath(folder) {
						others = append(others, folder)
					}
				}
			}
			rows.Close()
		}
	}
	out, err := podmanOutput(ctx, s.UserContext, 15*time.Second, "exec", "varnish",
		"varnishncsa", "-d", "-q", varnishSiteQuery(s, others), "-F", "%{Varnish:handling}x")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(out, "\n") {
		switch strings.TrimSpace(line) {
		case "hit":
			hits++
		case "miss", "pass":
			misses++
		}
	}
	return hits, misses
}

// parseSizeToBytes reads compose-style sizes like "0.1G" or "512M"
func parseSizeToBytes(v string) int64 {
	v = strings.ToUpper(strings.Trim(strings.TrimSpace(v), `"'`))
	mult := float64(1)
	switch {
	case strings.HasSuffix(v, "G"):
		mult, v = 1<<30, strings.TrimSuffix(v, "G")
	case strings.HasSuffix(v, "M"):
		mult, v = 1<<20, strings.TrimSuffix(v, "M")
	case strings.HasSuffix(v, "K"):
		mult, v = 1<<10, strings.TrimSuffix(v, "K")
	}
	f, err := strconv.ParseFloat(strings.TrimSuffix(v, "B"), 64)
	if err != nil {
		return 0
	}
	return int64(f * mult)
}

func parseRedisInfo(info string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if k, val, ok := strings.Cut(line, ":"); ok && !strings.HasPrefix(line, "#") {
			fields[k] = val
		}
	}
	return fields
}

func redisStatusFor(ctx context.Context, a *appctx.App, s *wpCacheSite) redisCacheStatus {
	rs := redisCacheStatus{Allowed: userAllowed(ctx, a, s, "redis")}
	if dropin, err := os.ReadFile(filepath.Join(s.HostPath, "wp-content", "object-cache.php")); err == nil {
		rs.Enabled = strings.Contains(string(dropin), redisDropinMarker)
	}
	ram, _ := docker.GetEnvValue(s.UserContext, "REDIS_RAM")
	rs.MemoryLimit = parseSizeToBytes(ram)

	rs.ServiceRunning = docker.IsServiceRunning(ctx, s.UserContext, "redis")
	if !rs.ServiceRunning {
		return rs
	}
	since := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	out, err := podmanOutput(ctx, s.UserContext, 15*time.Second, "exec", "redis", "sh", "-c",
		`redis-cli INFO; redis-cli --raw EVAL "$2" 0 "$1*" "$3"`, "sh", s.redisPrefix(), redisSiteStatsLua, since)
	if err != nil {
		return rs
	}
	f := parseRedisInfo(out)
	hits, _ := strconv.ParseFloat(f["keyspace_hits"], 64)
	misses, _ := strconv.ParseFloat(f["keyspace_misses"], 64)
	rs.Hits, rs.Misses = hits, misses
	if hits+misses > 0 {
		rs.HitRatio = hits / (hits + misses)
	}
	rs.UsedMemory, _ = strconv.ParseInt(f["used_memory"], 10, 64)
	if maxmem, _ := strconv.ParseInt(f["maxmemory"], 10, 64); maxmem > 0 {
		rs.MemoryLimit = maxmem
	}
	rs.SiteKeys, _ = strconv.ParseInt(strings.TrimSpace(f["op_site_keys"]), 10, 64)
	rs.SiteBytes, _ = strconv.ParseInt(strings.TrimSpace(f["op_site_bytes"]), 10, 64)
	if metrics, _ := strconv.ParseInt(strings.TrimSpace(f["op_site_metrics"]), 10, 64); metrics > 0 {
		rs.SiteMetrics = true
		rs.Hits, _ = strconv.ParseFloat(strings.TrimSpace(f["op_site_hits"]), 64)
		rs.Misses, _ = strconv.ParseFloat(strings.TrimSpace(f["op_site_misses"]), 64)
		rs.HitRatio = 0
		if rs.Hits+rs.Misses > 0 {
			rs.HitRatio = rs.Hits / (rs.Hits + rs.Misses)
		}
	}
	return rs
}

// redisSiteStatsLua walks this site's keys once for count and memory, and sums the redis-cache plugin's own per-request metrics from the last hour
const redisSiteStatsLua = `
local cursor, keys, bytes, hits, misses, metrics = "0", 0, 0, 0, 0, 0
repeat
  local r = redis.call("SCAN", cursor, "MATCH", ARGV[1], "COUNT", 1000)
  cursor = r[1]
  for _, k in ipairs(r[2]) do
    keys = keys + 1
    bytes = bytes + (redis.call("MEMORY", "USAGE", k) or 0)
    if string.sub(k, -19) == "redis-cache:metrics" and redis.call("TYPE", k).ok == "zset" then
      for _, m in ipairs(redis.call("ZRANGEBYSCORE", k, ARGV[2], "+inf")) do
        local h = tonumber(string.match(m, 's:4:"hits";i:(%d+);')) or 0
        local mi = tonumber(string.match(m, 's:6:"misses";i:(%d+);')) or 0
        hits, misses, metrics = hits + h, misses + mi, metrics + 1
      end
    end
  end
until cursor == "0"
return "op_site_keys:" .. keys .. "\nop_site_bytes:" .. bytes .. "\nop_site_hits:" .. hits .. "\nop_site_misses:" .. misses .. "\nop_site_metrics:" .. metrics
`

// opcacheProbePHP talks FastCGI to php-fpm directly so we read the pool's real opcache, not the CLI's
const opcacheProbePHP = `
$probe = '/tmp/.openpanel_opcache_probe.php';
file_put_contents($probe, '<?php $s = function_exists("opcache_get_status") ? @opcache_get_status(false) : false; echo json_encode(["enable" => (bool) ini_get("opcache.enable"), "memory" => (int) ini_get("opcache.memory_consumption"), "status" => $s]);');
$fp = @fsockopen('127.0.0.1', 9000, $en, $es, 3);
if (!$fp) { @unlink($probe); fwrite(STDERR, "php-fpm is not reachable: $es"); exit(1); }
stream_set_timeout($fp, 5);
$rec = function ($t, $c) { $l = strlen($c); return chr(1) . chr($t) . "\x00\x01" . chr($l >> 8) . chr($l & 255) . "\x00\x00" . $c; };
$nv = function ($n, $v) { $a = strlen($n); $b = strlen($v); return ($a < 128 ? chr($a) : pack('N', $a | 0x80000000)) . ($b < 128 ? chr($b) : pack('N', $b | 0x80000000)) . $n . $v; };
$params = ['SCRIPT_FILENAME' => $probe, 'SCRIPT_NAME' => '/probe.php', 'REQUEST_METHOD' => 'GET', 'QUERY_STRING' => '', 'SERVER_PROTOCOL' => 'HTTP/1.1', 'GATEWAY_INTERFACE' => 'CGI/1.1', 'REMOTE_ADDR' => '127.0.0.1', 'SERVER_NAME' => 'localhost', 'CONTENT_LENGTH' => '0'];
$p = ''; foreach ($params as $k => $v) { $p .= $nv($k, $v); }
fwrite($fp, $rec(1, "\x00\x01\x00\x00\x00\x00\x00\x00") . $rec(4, $p) . $rec(4, '') . $rec(5, ''));
$read = function ($n) use ($fp) { $d = ''; while (strlen($d) < $n && !feof($fp)) { $c = fread($fp, $n - strlen($d)); if ($c === false || $c === '') break; $d .= $c; } return $d; };
$out = '';
while (!feof($fp)) {
    $h = $read(8); if (strlen($h) < 8) break;
    $u = unpack('Cver/Ctype/nid/nlen/Cpad/Cres', $h);
    $c = $u['len'] ? $read($u['len']) : ''; if ($u['pad']) $read($u['pad']);
    if ($u['type'] == 6) $out .= $c;
    if ($u['type'] == 3) break;
}
fclose($fp); @unlink($probe);
$pos = strpos($out, "\r\n\r\n");
echo $pos === false ? $out : substr($out, $pos + 4);
`

func readUserIniOpcacheOff(hostPath string) bool {
	content, err := os.ReadFile(filepath.Join(hostPath, ".user.ini"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(content), "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.TrimSpace(k) != "opcache.enable" {
			continue
		}
		switch strings.ToLower(strings.Trim(strings.TrimSpace(v), `"'`)) {
		case "0", "off", "false", "no":
			return true
		}
	}
	return false
}

func opcacheStatusFor(ctx context.Context, s *wpCacheSite) opcacheStatus {
	o := opcacheStatus{PHPVersion: s.PHPVersion}
	if s.Litespeed {
		o.Error = "Per-site OPcache toggle is not available on OpenLiteSpeed, it does not read .user.ini files."
		return o
	}
	out, err := podmanOutput(ctx, s.UserContext, 15*time.Second, "exec", s.PHPContainer,
		"php", "-d", "disable_functions=", "-d", "open_basedir=none", "-d", "display_errors=0", "-r", opcacheProbePHP)
	if err != nil {
		o.Error = "Could not read OPcache status from " + s.PHPContainer + "."
		return o
	}
	var probe struct {
		Enable bool  `json:"enable"`
		Memory int64 `json:"memory"`
		Status *struct {
			Memory struct {
				Used   int64 `json:"used_memory"`
				Free   int64 `json:"free_memory"`
				Wasted int64 `json:"wasted_memory"`
			} `json:"memory_usage"`
			Stats struct {
				Scripts int64   `json:"num_cached_scripts"`
				HitRate float64 `json:"opcache_hit_rate"`
			} `json:"opcache_statistics"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &probe); err != nil {
		o.Error = "Unexpected OPcache status output."
		return o
	}
	o.Available = true
	o.GlobalEnabled = probe.Enable
	o.TotalMemory = probe.Memory << 20
	if probe.Status != nil {
		o.UsedMemory = probe.Status.Memory.Used + probe.Status.Memory.Wasted
		o.TotalMemory = probe.Status.Memory.Used + probe.Status.Memory.Free + probe.Status.Memory.Wasted
		o.CachedScripts = probe.Status.Stats.Scripts
		o.HitRate = probe.Status.Stats.HitRate / 100
	}
	o.Enabled = o.GlobalEnabled && !readUserIniOpcacheOff(s.HostPath)
	return o
}

// measurementPath sits next to the PageSpeed results, same file naming
func measurementPath(s *wpCacheSite) string {
	return filepath.Join("/etc/openpanel/openpanel/websites/cache_speed", strings.ReplaceAll(s.Site, "/", "_")+".json")
}

func loadMeasurement(s *wpCacheSite) *cacheMeasurement {
	raw, err := os.ReadFile(measurementPath(s))
	if err != nil {
		return nil
	}
	var m cacheMeasurement
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	return &m
}

// saveMeasurement keeps the latest result plus a short history for the trend line
func saveMeasurement(s *wpCacheSite, m *cacheMeasurement) error {
	if prev := loadMeasurement(s); prev != nil {
		m.History = prev.History
		if len(m.History) == 0 {
			m.History = []measurePoint{{CachedMS: prev.CachedMS, DirectMS: prev.DirectMS, MeasuredAt: prev.MeasuredAt}}
		}
	}
	m.History = append(m.History, measurePoint{CachedMS: m.CachedMS, DirectMS: m.DirectMS, MeasuredAt: m.MeasuredAt})
	if len(m.History) > measureHistoryLen {
		m.History = m.History[len(m.History)-measureHistoryLen:]
	}
	path := measurementPath(s)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func wpCacheStatus(ctx context.Context, a *appctx.App, s *wpCacheSite) wpCacheStatusResponse {
	var resp wpCacheStatusResponse
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); resp.Varnish = varnishStatusFor(ctx, a, s) }()
	go func() { defer wg.Done(); resp.Redis = redisStatusFor(ctx, a, s) }()
	go func() { defer wg.Done(); resp.Opcache = opcacheStatusFor(ctx, s) }()
	wg.Wait()
	resp.Measurement = loadMeasurement(s)
	return resp
}

// ---------------------- ACTIONS ---------------------- //

func setVarnishForSite(ctx context.Context, a *appctx.App, s *wpCacheSite, on bool) (string, error) {
	if on {
		if !docker.IsServiceRunning(ctx, s.UserContext, "varnish") {
			// same as the Varnish page: keep going even if the browser gives up
			opCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			if _, errMsg := cache.StartVarnish(opCtx, s.UserContext); errMsg != "" {
				return "", errors.New("Failed to start Varnish: " + errMsg)
			}
		}
		if out, err := runOpencli(ctx, "domains-varnish", s.Domain, "on"); err != nil {
			return "", fmt.Errorf("failed to enable Varnish for %s: %s", s.Domain, out)
		}
		return "Page cache is now enabled for " + s.Domain + ".", nil
	}
	if out, err := runOpencli(ctx, "domains-varnish", s.Domain, "off"); err != nil {
		return "", fmt.Errorf("failed to disable Varnish for %s: %s", s.Domain, out)
	}
	return "Page cache is now disabled for " + s.Domain + ".", nil
}

func runOpencli(ctx context.Context, args ...string) (string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(runCtx, "opencli", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func setVarnishTTL(ctx context.Context, s *wpCacheSite, ttl int) (string, error) {
	if !wpCacheTTLChoices[ttl] {
		return "", errors.New("invalid cache lifetime")
	}
	path := vclPath(s.UserContext)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("Varnish configuration file not found")
	}
	// the template default is 1h, so picking that just drops the override
	if ttl == vclDefaultTTL {
		ttl = 0
	}
	updated, err := setVCLTTL(string(content), s.Domain, ttl)
	if err != nil {
		return "", err
	}
	// write in place, the file is bind-mounted into the varnish container
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", err
	}
	if docker.IsServiceRunning(ctx, s.UserContext, "varnish") {
		if err := reloadVarnishVCL(ctx, s.UserContext); err != nil {
			_ = os.WriteFile(path, content, 0o644)
			_ = reloadVarnishVCL(ctx, s.UserContext)
			return "", fmt.Errorf("Varnish rejected the new configuration: %v", err)
		}
		_ = banVarnishSite(ctx, s)
	}
	return "Page cache lifetime updated.", nil
}

func ensureRedisRunning(ctx context.Context, s *wpCacheSite) error {
	if docker.IsServiceRunning(ctx, s.UserContext, "redis") {
		return nil
	}
	opCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	result := docker.StartOrStopContainer(opCtx, s.UserContext, "redis", "activate", "")
	if !result.Success || !docker.WaitForServiceRunning(opCtx, s.UserContext, "redis") {
		return errors.New("failed to start the Redis service")
	}
	return nil
}

func setRedisForSite(ctx context.Context, s *wpCacheSite, on bool) (string, error) {
	if on {
		if err := ensureRedisRunning(ctx, s); err != nil {
			return "", err
		}
		for _, c := range [][]string{
			{"config", "set", "WP_REDIS_HOST", "redis"},
			{"config", "set", "WP_REDIS_PORT", "6379", "--raw"},
			{"config", "set", "WP_REDIS_PREFIX", s.redisPrefix()},
			// otherwise a flush from one site wipes every other site's keys in the shared redis
			{"config", "set", "WP_REDIS_SELECTIVE_FLUSH", "true", "--raw"},
		} {
			if out, err := s.wp(ctx, append(c, "--skip-plugins", "--skip-themes")...); err != nil {
				return "", fmt.Errorf("wp %s failed: %s", strings.Join(c[:3], " "), out)
			}
		}
		if _, err := s.wp(ctx, "plugin", "is-installed", "redis-cache", "--skip-plugins", "--skip-themes"); err != nil {
			if out, err := s.wp(ctx, "plugin", "install", "redis-cache", "--skip-plugins", "--skip-themes"); err != nil {
				return "", fmt.Errorf("could not install the Redis Object Cache plugin: %s", out)
			}
		}
		if out, err := s.wp(ctx, "plugin", "activate", "redis-cache", "--skip-themes"); err != nil {
			return "", fmt.Errorf("could not activate the Redis Object Cache plugin: %s", out)
		}
		if out, err := s.wp(ctx, "redis", "enable", "--force", "--skip-themes"); err != nil {
			return "", fmt.Errorf("could not enable the object cache: %s", out)
		}
		// opcache may still hold the old wp-config.php without the redis constants, the drop-in would then hit 127.0.0.1 and 500
		reloadPHPFPM(ctx, s)
		return "Redis object cache is now enabled.", nil
	}

	_, _ = s.wp(ctx, "redis", "disable", "--skip-themes")
	// drop-in can outlive the plugin, remove it ourselves if it's still ours
	dropin := filepath.Join(s.HostPath, "wp-content", "object-cache.php")
	if content, err := os.ReadFile(dropin); err == nil && strings.Contains(string(content), redisDropinMarker) {
		_ = os.Remove(dropin)
	}
	if out, err := s.wp(ctx, "plugin", "deactivate", "redis-cache", "--skip-themes"); err != nil && !strings.Contains(out, "already deactivated") && !strings.Contains(out, "not installed") {
		return "", fmt.Errorf("could not deactivate the Redis Object Cache plugin: %s", out)
	}
	return "Redis object cache is now disabled.", nil
}

func flushRedisForSite(ctx context.Context, s *wpCacheSite) {
	if !docker.IsServiceRunning(ctx, s.UserContext, "redis") {
		return
	}
	_, _ = podmanOutput(ctx, s.UserContext, 30*time.Second, "exec", "redis", "sh", "-c",
		`redis-cli --scan --pattern "$1*" | xargs -r -n 500 redis-cli del >/dev/null`, "sh", s.redisPrefix())
}

// setUserIniOpcache adds/removes our opcache.enable=0 lines in the site's .user.ini, keeping whatever else the user has there
func setUserIniOpcache(content string, on bool) string {
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines)+2)
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == userIniMarker {
			continue
		}
		if k, _, ok := strings.Cut(t, "="); ok && strings.TrimSpace(k) == "opcache.enable" {
			continue
		}
		kept = append(kept, line)
	}
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	if !on {
		kept = append(kept, userIniMarker, "opcache.enable=0")
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, "\n") + "\n"
}

func setOpcacheForSite(ctx context.Context, s *wpCacheSite, on bool) (string, error) {
	if s.Litespeed {
		return "", errors.New("per-site OPcache toggle is not available on OpenLiteSpeed, it does not read .user.ini files")
	}
	path := filepath.Join(s.HostPath, ".user.ini")
	content, err := os.ReadFile(path)
	existed := err == nil
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	updated := setUserIniOpcache(string(content), on)
	switch {
	case updated == "" && existed:
		_ = os.Remove(path)
	case updated != "":
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return "", err
		}
		if !existed {
			if info, statErr := os.Stat(s.HostPath); statErr == nil {
				if st, ok := info.Sys().(*syscall.Stat_t); ok {
					_ = os.Chown(path, int(st.Uid), int(st.Gid))
				}
			}
		}
	}
	// .user.ini is cached per fpm worker for up to 5 min, a graceful reload applies it now
	reloadPHPFPM(ctx, s)
	if on {
		return "OPcache is now enabled for this site.", nil
	}
	return "OPcache is now disabled for this site.", nil
}

// reloadPHPFPM gracefully reloads the site's fpm master, which also resets its opcache
func reloadPHPFPM(ctx context.Context, s *wpCacheSite) {
	if s.Litespeed {
		return
	}
	_, _ = podmanOutput(ctx, s.UserContext, 15*time.Second, "exec", s.PHPContainer, "sh", "-c",
		`for p in /proc/[0-9]*; do grep -q "php-fpm: master" "$p/cmdline" 2>/dev/null && kill -USR2 "${p#/proc/}"; done; true`)
}

// noCachePHP renders the home page in a throwaway php process with opcache off and the redis drop-in disabled, so nothing touches the fpm pool visitors use
const noCachePHP = `
[, $root, $host, $path] = $argv;
define('WP_REDIS_DISABLED', true);
$script = $path . 'index.php';
$_SERVER = array_merge($_SERVER, ['HTTP_HOST' => $host, 'SERVER_NAME' => $host, 'HTTPS' => 'on', 'SERVER_PORT' => '443', 'REQUEST_METHOD' => 'GET', 'REQUEST_URI' => $path, 'SCRIPT_NAME' => $script, 'PHP_SELF' => $script, 'SCRIPT_FILENAME' => $root . '/index.php', 'DOCUMENT_ROOT' => $root, 'REMOTE_ADDR' => '127.0.0.1', 'HTTP_USER_AGENT' => 'OpenPanel cache check', 'SERVER_PROTOCOL' => 'HTTP/1.1']);
chdir($root);
$t = hrtime(true);
ob_start(function ($b) { return ''; });
register_shutdown_function(function () use ($t) { while (ob_get_level()) { ob_end_flush(); } fwrite(STDERR, sprintf("%.4f %d\n", (hrtime(true) - $t) / 1e9, http_response_code() ?: 200)); });
require $root . '/index.php';
`

// measureScript runs inside the php container, it sits on the account's private network so it reaches varnish and the webserver by name, skipping caddy and the WAF
const measureScript = `D="$1"; P="$2"; WS="$3"; V="$4"; ROOT="$5"; NC="$6"
m() { i=0; while [ $i -lt 6 ]; do curl -sk -o /dev/null --max-time 20 -A "OpenPanel cache check" -w '%{time_starttransfer} %{http_code}\n' "$@"; i=$((i+1)); done; }
echo "#cached"
if [ "$V" = "1" ]; then m --connect-to "$D:80:varnish:80" --connect-to "www.$D:80:varnish:80" -H "X-Forwarded-Proto: https" "http://$D$P"
else m --connect-to "$D:443:$WS:443" --connect-to "www.$D:443:$WS:443" "https://$D$P"; fi
echo "#direct"
# the image's php alias goes through a wrapper that adds ~0.7s per start
PHPBIN=/usr/local/bin/php; [ -x "$PHPBIN" ] || PHPBIN=php
i=0; while [ $i -lt 4 ]; do "$PHPBIN" -d opcache.enable=0 -d opcache.enable_cli=0 -d disable_functions= -d open_basedir=none -d display_errors=0 -d error_log=/dev/null -d memory_limit=512M -d max_execution_time=30 -r "$NC" "$ROOT" "$D" "$P" 2>&1 >/dev/null | tail -n 1; i=$((i+1)); done
`

// parseMeasureOutput returns the median ms and last status code per section, skipping the first (warm-up) request
func parseMeasureOutput(out string) map[string][2]float64 {
	samples := map[string][]float64{}
	codes := map[string]float64{}
	section := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			section = strings.TrimPrefix(line, "#")
			continue
		}
		fields := strings.Fields(line)
		if section == "" || len(fields) != 2 {
			continue
		}
		t, err1 := strconv.ParseFloat(fields[0], 64)
		code, err2 := strconv.ParseFloat(fields[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		samples[section] = append(samples[section], t*1000)
		codes[section] = code
	}
	res := map[string][2]float64{}
	for k, v := range samples {
		if len(v) > 1 {
			v = v[1:]
		}
		sorted := append([]float64(nil), v...)
		sort.Float64s(sorted)
		mid := sorted[len(sorted)/2]
		if len(sorted)%2 == 0 {
			mid = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
		}
		res[k] = [2]float64{float64(int(mid*10+0.5)) / 10, codes[k]}
	}
	return res
}

func measureSite(ctx context.Context, s *wpCacheSite) (*cacheMeasurement, error) {
	viaVarnish := docker.IsServiceRunning(ctx, s.UserContext, "varnish") && cache.DomainVarnishStatus(s.Domain) == "On"
	path := "/"
	if s.Folder != "" {
		path = "/" + s.Folder + "/"
	}
	v := "0"
	if viaVarnish {
		v = "1"
	}
	out, err := podmanOutput(ctx, s.UserContext, 150*time.Second, "exec", s.PHPContainer, "sh", "-c", measureScript, "sh", s.Domain, path, s.Webserver, v, s.Docroot, noCachePHP)
	if err != nil {
		return nil, fmt.Errorf("measurement failed: %s", strings.TrimSpace(out))
	}
	res := parseMeasureOutput(out)
	direct, ok := res["direct"]
	if !ok {
		return nil, errors.New("measurement failed: could not render the page without cache")
	}
	m := &cacheMeasurement{DirectMS: direct[0], DirectCode: int(direct[1]), ViaVarnish: viaVarnish, MeasuredAt: time.Now().Unix()}
	if cached, ok := res["cached"]; ok {
		m.CachedMS, m.CachedCode = cached[0], int(cached[1])
	} else {
		m.CachedMS, m.CachedCode = m.DirectMS, m.DirectCode
	}
	if err := saveMeasurement(s, m); err != nil {
		log.Printf("WEBSITES - could not save cache measurement for %s: %v", s.Site, err)
	}
	return m, nil
}

// ---------------------- HANDLERS ---------------------- //

func handleWPCacheStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	s, status, err := resolveWPCacheSite(a, r, r.URL.Query().Get("domain"))
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, wpCacheStatus(r.Context(), a, s))
}

func handleWPCacheAction(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	s, status, err := resolveWPCacheSite(a, r, r.Form.Get("domain"))
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	action := r.PathValue("action")
	on := r.Form.Get("state") == "on"
	onOff := map[bool]string{true: "enabled", false: "disabled"}[on]
	ip := reqip.ClientIP(r)

	var msg string
	switch action {
	case "varnish":
		if !userAllowed(ctx, a, s, "varnish") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Varnish is not available on your plan."})
			return
		}
		msg, err = setVarnishForSite(ctx, a, s, on)
		_ = logger.RecordUserAction(a.Config, s.Username, onOff+" Varnish page cache for WordPress site "+s.Site, ip)
	case "varnish-ttl":
		if !userAllowed(ctx, a, s, "varnish") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Varnish is not available on your plan."})
			return
		}
		ttl, _ := strconv.Atoi(r.Form.Get("ttl"))
		msg, err = setVarnishTTL(ctx, s, ttl)
		_ = logger.RecordUserAction(a.Config, s.Username, "set Varnish cache lifetime to "+strconv.Itoa(ttl)+"s for domain "+s.Domain, ip)
	case "redis":
		if !userAllowed(ctx, a, s, "redis") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Redis is not available on your plan."})
			return
		}
		msg, err = setRedisForSite(ctx, s, on)
		_ = logger.RecordUserAction(a.Config, s.Username, onOff+" Redis object cache for WordPress site "+s.Site, ip)
	case "opcache":
		msg, err = setOpcacheForSite(ctx, s, on)
		_ = logger.RecordUserAction(a.Config, s.Username, onOff+" OPcache for WordPress site "+s.Site, ip)
	case "purge":
		if docker.IsServiceRunning(ctx, s.UserContext, "varnish") {
			_ = banVarnishSite(ctx, s)
		}
		flushRedisForSite(ctx, s)
		// another object cache plugin (memcached etc.) only knows how to flush itself
		if dropin, readErr := os.ReadFile(filepath.Join(s.HostPath, "wp-content", "object-cache.php")); readErr == nil && !strings.Contains(string(dropin), redisDropinMarker) {
			_, _ = s.wp(ctx, "cache", "flush", "--skip-themes")
		}
		msg = "All caches purged for " + s.Site + "."
		_ = logger.RecordUserAction(a.Config, s.Username, "purged caches for WordPress site "+s.Site, ip)
	case "measure":
		m, mErr := measureSite(ctx, s)
		if mErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mErr.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"measurement": m})
		return
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported action"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": msg, "status": wpCacheStatus(ctx, a, s)})
}
