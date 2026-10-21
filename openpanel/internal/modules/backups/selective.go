package backups

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// backupItem is one pickable thing in a backup: a mailbox, a domain or a dns zone
type backupItem struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Count     int64  `json:"count,omitempty"`
	Mtime     int64  `json:"mtime,omitempty"`
	OnServer  bool   `json:"on_server"`
	Suspended bool   `json:"suspended,omitempty"`
}

var domainNameRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// readItems turns a saved mail/vhosts/dns list into items, vhosts and dns files become domain names
func readItems(listFile, kind string, domains map[string]bool) ([]backupItem, error) {
	f, err := os.Open(listFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	items := []backupItem{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		fields := strings.SplitN(sc.Text(), "\t", 3)
		if len(fields) != 3 {
			continue
		}
		size, _ := strconv.ParseInt(fields[0], 10, 64)
		second, _ := strconv.ParseInt(fields[1], 10, 64)
		item := backupItem{Name: fields[2], Size: size}
		switch kind {
		case "mail":
			item.Count = second
			domain, _, _ := strings.Cut(item.Name, "/")
			item.OnServer = domains[domain]
		case "vhosts", "dns":
			ext := map[string]string{"vhosts": ".conf", "dns": ".zone"}[kind]
			domain, ok := strings.CutSuffix(item.Name, ext)
			if !ok || strings.Contains(domain, "/") || !domainNameRE.MatchString(domain) {
				continue
			}
			item.Name, item.Mtime, item.OnServer = domain, second, domains[domain]
			item.Suspended = kind == "vhosts" && domainSuspended(domain)
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, sc.Err()
}

// selectiveTargets turns the posted items of a selective restore_target into archive paths
func selectiveTargets(restoreTarget string, values []string, plan *restorePlan) ([]string, error) {
	var targets []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		switch restoreTarget {
		case "paths":
			clean, ok := cleanBackupPath(v)
			if !ok || clean == "" {
				return nil, fmt.Errorf("invalid path: %s", v)
			}
			targets = append(targets, "backup/html/"+clean)
		case "mailboxes":
			domain, user, ok := strings.Cut(strings.ToLower(v), "/")
			if !ok || !domainNameRE.MatchString(domain) || user == "" || strings.ContainsAny(user, "/\x00") || user == "." || user == ".." {
				return nil, fmt.Errorf("invalid mailbox: %s", v)
			}
			if !plan.domains[domain] {
				return nil, fmt.Errorf("%s is not on your account, restore it from the Domains tab first", domain)
			}
			targets = append(targets, "backup/emails/"+domain+"/"+user)
		case "domains":
			domain := strings.ToLower(v)
			if !domainNameRE.MatchString(domain) {
				return nil, fmt.Errorf("invalid domain: %s", v)
			}
			if plan.suspended != nil && plan.suspended(domain) {
				return nil, fmt.Errorf("%s is suspended, unsuspend it before restoring its VirtualHost", domain)
			}
			// the domain's dns zone comes along when the backup has it
			targets = append(targets, "backup/dns/"+domain+".zone", "backup/vhosts/"+domain+".conf")
		case "dns_zones":
			domain := strings.ToLower(v)
			if !domainNameRE.MatchString(domain) {
				return nil, fmt.Errorf("invalid domain: %s", v)
			}
			if !plan.domains[domain] {
				return nil, fmt.Errorf("%s is not on your account, restore it from the Domains tab first", domain)
			}
			targets = append(targets, "backup/dns/"+domain+".zone")
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("nothing selected")
	}
	return targets, nil
}

// collectEntries streams the archive and returns the content of the wanted single files, for vhost files that must be read before their domain exists
func collectEntries(ctx context.Context, store remoteStore, backup string, targets []string) (map[string][]byte, error) {
	rc, err := store.Open(ctx, backup)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	gz, err := gzip.NewReader(rc)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	found := map[string][]byte{}
	layout := newArchiveLayout()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return found, err
		}
		raw := strings.Trim(header.Name, "/")
		name := layout.observe(raw)
		if layout.pastAll(raw, targets) {
			break
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		for _, t := range targets {
			if name == t {
				content, err := io.ReadAll(io.LimitReader(tr, 4<<20))
				if err != nil {
					return found, err
				}
				found[name] = content
				break
			}
		}
	}
	return found, nil
}

var vhostPHPRE = regexp.MustCompile(`php-fpm-(\d+\.\d+)`)

// vhostPHPVersion is the php-fpm version a vhost hands .php files to, "" for webservers that run php themselves
func vhostPHPVersion(content []byte) string {
	if m := vhostPHPRE.FindSubmatch(content); m != nil {
		return string(m[1])
	}
	return ""
}

var vhostDocrootRE = regexp.MustCompile(`(?mi)^\s*(?:DocumentRoot|root|docRoot)\s+"?([^";\s]+)`)

// vhostDocroot reads the docroot out of an apache/nginx/openlitespeed vhost, falling back to the panel's default
func vhostDocroot(content []byte, domain string) string {
	if m := vhostDocrootRE.FindSubmatch(content); m != nil {
		clean := path.Clean(string(m[1]))
		if strings.HasPrefix(clean, "/var/www/html/") {
			return clean
		}
	}
	return "/var/www/html/" + domain
}

// ensureDomain adds a domain from the backup back to the account when it's gone, the plan's domain limit still applies
func ensureDomain(ctx context.Context, a *appctx.App, userID int, username, domain, docroot, phpVersion string, plan *restorePlan) error {
	if plan.domains[domain] {
		return nil
	}
	var owners int
	if err := a.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM domains WHERE domain_url = ?", domain).Scan(&owners); err != nil {
		return err
	}
	if owners > 0 {
		return fmt.Errorf("%s belongs to another account on this server", domain)
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	if hostingPlan, err := a.QueryPlanDetailsByID(ctx, planID); err == nil {
		if limit := web.AtoiDefault(hostingPlan.DomainsLimit, 0); limit != 0 {
			existing, _ := a.AllDomainsForUser(ctx, userID)
			urls := make([]appctx.Domain, len(existing)+1)
			for i, d := range existing {
				urls[i] = appctx.Domain{DomainURL: d.DomainURL}
			}
			urls[len(existing)] = appctx.Domain{DomainURL: domain}
			if mains, _ := appctx.Categorize(urls); len(mains) > limit {
				return fmt.Errorf("adding %s would go over the domain limit of your hosting plan", domain)
			}
		}
	}

	args := []string{"domains-add", domain, username, "--docroot", docroot}
	if phpVersion != "" {
		args = append(args, "--php_version", phpVersion)
	}
	out, err := exec.CommandContext(ctx, "opencli", args...).CombinedOutput()
	if err != nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		return fmt.Errorf("adding %s failed: %s", domain, lines[len(lines)-1])
	}
	plan.domains[domain] = true
	return nil
}

// restoreSelectedItems is the paths/mailboxes/domains/dns_zones branch of handleRestoreFromBackup, it streams the archive instead of downloading all of it first
func restoreSelectedItems(ctx context.Context, a *appctx.App, req restoreRequest, store remoteStore, backup, restoreTarget string, plan *restorePlan) (int, map[string]any) {
	userID, username, userContext := req.UserID, req.Username, req.UserContext
	targets, err := selectiveTargets(restoreTarget, req.Form["items"], plan)
	if err != nil {
		return http.StatusBadRequest, map[string]any{"error": err.Error()}
	}
	uid, _ := a.GetUID(ctx, userContext)

	var extracted []string
	if restoreTarget == "domains" {
		// read the vhosts first so a missing domain can be added with its own docroot
		contents, err := collectEntries(ctx, store, backup, targets)
		if err != nil {
			return http.StatusInternalServerError, map[string]any{"error": err.Error()}
		}
		var failed []string
		for _, t := range targets {
			domain, isVhost := strings.CutSuffix(strings.TrimPrefix(t, "backup/vhosts/"), ".conf")
			content, ok := contents[t]
			if !isVhost || !ok {
				continue
			}
			if err := ensureDomain(ctx, a, userID, username, domain, vhostDocroot(content, domain), vhostPHPVersion(content), plan); err != nil {
				failed = append(failed, err.Error())
				continue
			}
			if dest, ok := plan.extractEntry(bytes.NewReader(content), t, uid); ok {
				extracted = append(extracted, dest)
			}
			// written after domains-add so the backed up zone replaces the fresh default one
			zone := "backup/dns/" + domain + ".zone"
			if zoneContent, ok := contents[zone]; ok {
				if dest, ok := plan.extractEntry(bytes.NewReader(zoneContent), zone, 0); ok {
					extracted = append(extracted, dest)
				}
			}
		}
		if len(failed) > 0 && len(extracted) == 0 {
			return http.StatusInternalServerError, map[string]any{"error": strings.Join(failed, "; ")}
		}
	} else {
		extracted, err = restoreSelected(ctx, store, backup, plan, uid, targets)
		if err != nil {
			return http.StatusInternalServerError, map[string]any{"error": err.Error()}
		}
	}
	if len(extracted) == 0 {
		return http.StatusNotFound, map[string]any{"error": "The selected items were not found in this backup."}
	}

	applyRestored(ctx, userContext, extracted, plan)
	_ = logger.RecordUserAction(a.Config, username, "restored "+strings.Join(req.Form["items"], ", ")+" from backup "+backup, req.IP)
	return http.StatusOK, map[string]any{"success": true, "message": "Restored (" + strconv.Itoa(len(extracted)) + " items)."}
}

// applyRestored makes restored config take effect: cron jobs, dns zones and vhosts
func applyRestored(ctx context.Context, userContext string, extracted []string, plan *restorePlan) {
	cronsPath := plan.sections["backup/crons.ini"]
	vhostsDir := plan.sections["backup/vhosts"]
	cronsDone, vhostsDone := false, false
	var phpDomains []string
	for _, p := range extracted {
		switch {
		case p == cronsPath && !cronsDone:
			crons.Reload(ctx, userContext)
			cronsDone = true
		case strings.HasPrefix(p, plan.zonesDir+"/") && strings.HasSuffix(p, ".zone"):
			dns.RestartDNSService(strings.TrimSuffix(filepath.Base(p), ".zone"))
		case strings.HasPrefix(p, vhostsDir+"/"):
			vhostsDone = true
			phpDomains = append(phpDomains, p)
		}
	}
	// the restored vhost picks the php version, php-domain also starts that php-fpm and records it on the domain
	for _, p := range phpDomains {
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		domain := strings.TrimSuffix(filepath.Base(p), ".conf")
		if version := vhostPHPVersion(content); version != "" && plan.domains[domain] {
			_ = exec.CommandContext(ctx, "opencli", "php-domain", domain, "--update", version).Run()
		}
	}
	if vhostsDone {
		if ws := webserver.GetEnvFileValue(userContext, "WEB_SERVER"); ws != "" && docker.IsServiceRunning(ctx, userContext, ws) {
			docker.ComposeContainer(ctx, userContext, ws, "restart")
		}
	}
}
