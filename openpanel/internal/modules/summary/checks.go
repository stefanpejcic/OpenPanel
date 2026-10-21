package summary

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/backups"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
)

// vars so tests can point them at a temp dir
var (
	usersDir        = "/etc/openpanel/openpanel/core/users/"
	vulnDir         = "/etc/openpanel/wordpress/vulnerability/"
	sslDir          = "/etc/openpanel/caddy/ssl/"
	quotaReportPath = "/etc/openpanel/openpanel/quota_report.json"
	homeDir         = "/home/"
)

// checkVulnerabilities reads the scan results websites-vulnerability saves per WordPress site
func (b *builder) checkVulnerabilities() {
	if !b.can("wordpress") {
		return
	}
	checked, vulnerable := 0, 0
	for _, site := range b.sites {
		if !strings.EqualFold(site.Type, "wordpress") {
			continue
		}
		checked++
		raw, err := tenantfs.ReadFile(vulnDir + strings.ReplaceAll(site.Name, "/", "_") + ".json")
		if err != nil {
			continue
		}
		var res map[string]json.RawMessage
		if json.Unmarshal(raw, &res) != nil {
			continue
		}
		var names []string
		for key, v := range res {
			if key == "timestamp" {
				continue
			}
			var vuln struct {
				Data struct {
					Vulnerability []struct {
						Name string `json:"name"`
					} `json:"vulnerability"`
				} `json:"data"`
			}
			if json.Unmarshal(v, &vuln) != nil {
				continue
			}
			for _, x := range vuln.Data.Vulnerability {
				names = append(names, x.Name)
			}
		}
		if len(names) == 0 {
			continue
		}
		vulnerable++
		b.attention(Item{
			Title:  fmt.Sprintf("%s has %s", site.Name, plural(len(names), "known security hole", "known security holes")),
			Detail: "They're in " + joinAnd(vulnerableNames(names)) + ". An update usually fixes them.",
			Action: "Update WordPress and its plugins", Link: "/website?domain=" + site.Name,
		})
	}
	if checked > 0 && vulnerable == 0 {
		b.good("Your WordPress sites have no known security holes.")
	}
}

// checkMalware looks at the account's latest malware scan
func (b *builder) checkMalware() {
	if !b.can("malware_scan") {
		return
	}
	var last struct {
		StartedAt string `json:"started_at"`
		Infected  int    `json:"infected"`
		Status    string `json:"status"`
	}
	found := false
	eachLine(usersDir+b.username+"/malware_scans.jsonl", func(raw []byte) {
		if json.Unmarshal(raw, &last) == nil {
			found = true
		}
	})
	if !found {
		return
	}
	started, err := time.ParseInLocation("2006-01-02 15:04:05", last.StartedAt, b.from.Location())
	if err != nil {
		return
	}
	if last.Infected > 0 && !started.Before(b.from) && started.Before(b.to) {
		b.attention(Item{
			Title:  "Malware was found on your account",
			Detail: fmt.Sprintf("The scan found %s and locked them away in quarantine, so they can't run.", plural(last.Infected, "infected file", "infected files")),
			Action: "Review the quarantined files", Link: "/malware-scanner/quarantine",
		})
		return
	}
	if last.Infected == 0 {
		b.good("The last malware scan, on " + started.Format("2 January") + ", found nothing.")
	}
}

// certNotAfter reads the expiry of the domain's Let's Encrypt or custom certificate
func certNotAfter(domain string) (time.Time, bool) {
	for _, p := range []string{
		sslDir + "acme-v02.api.letsencrypt.org-directory/" + domain + "/" + domain + ".crt",
		sslDir + "custom/" + domain + "/" + domain + "/fullchain.pem",
	} {
		raw, err := tenantfs.ReadFile(p)
		if err != nil {
			continue
		}
		block, _ := pem.Decode(raw)
		if block == nil {
			continue
		}
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			return cert.NotAfter, true
		}
	}
	return time.Time{}, false
}

// checkSSL warns about certificates that should have renewed by now, renewal starts 30 days before expiry
func (b *builder) checkSSL() {
	checked, problems := 0, 0
	for _, d := range b.domains {
		notAfter, ok := certNotAfter(d.DomainURL)
		if !ok {
			continue
		}
		checked++
		days := int(time.Until(notAfter).Hours() / 24)
		if days >= 14 {
			continue
		}
		problems++
		it := Item{Detail: "Certificates renew by themselves, so something is blocking the renewal. Usually the domain no longer points to this server."}
		if days < 0 {
			it.Title = "The security certificate for " + d.DomainURL + " has expired"
			it.Detail = "Visitors see a 'not secure' warning instead of your website. " + it.Detail
		} else {
			it.Title = fmt.Sprintf("The security certificate for %s expires in %s", d.DomainURL, plural(days, "day", "days"))
			it.Detail = "After that, visitors see a 'not secure' warning. " + it.Detail
		}
		if b.can("ssl") {
			it.Action, it.Link = "Check SSL", "/domains/ssl"
		} else {
			it.Action = "Contact your hosting provider to renew it."
		}
		b.attention(it)
	}
	if checked > 0 && problems == 0 {
		b.good("All your security certificates are valid.")
	}
}

// ofelia prints one line per finished job run
var ofeliaFinishedRE = regexp.MustCompile(`\[Job "([^"]+)" \([0-9a-f]+\)\] Finished in "[^"]*", failed: (true|false), skipped: (?:true|false), error: (.*)$`)

type cronRuns struct {
	runs, failed int
	lastError    string
}

// parseCronRuns counts runs and failures per job from the cron container's log
func parseCronRuns(out string) map[string]*cronRuns {
	jobs := map[string]*cronRuns{}
	for _, line := range strings.Split(out, "\n") {
		m := ofeliaFinishedRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		j := jobs[m[1]]
		if j == nil {
			j = &cronRuns{}
			jobs[m[1]] = j
		}
		j.runs++
		if m[2] == "true" {
			j.failed++
			j.lastError = strings.TrimSpace(m[3])
		}
	}
	return jobs
}

// window is --since/--until for podman, podman returns nothing when --until is in the future so a day still running has no end
func (b *builder) window() []string {
	args := []string{"--since", b.from.Format(time.RFC3339)}
	if b.to.Before(time.Now()) {
		args = append(args, "--until", b.to.Format(time.RFC3339))
	}
	return args
}

func (b *builder) podman(ctx context.Context, args ...string) string {
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, _ := podmanmanager.Command(cmdCtx, b.userContext, podmanmanager.PodmanArgv(b.userContext, args...)).CombinedOutput()
	return string(out)
}

// checkCrons reports jobs that failed, a job that never succeeded is worse than one that failed once
func (b *builder) checkCrons(ctx context.Context) {
	if !b.can("crons") {
		return
	}
	jobs := parseCronRuns(b.podman(ctx, append(append([]string{"logs"}, b.window()...), "cron")...))
	names := make([]string, 0, len(jobs))
	for name := range jobs {
		names = append(names, name)
	}
	sort.Strings(names)
	failing := 0
	for _, name := range names {
		j := jobs[name]
		if j.failed == 0 {
			continue
		}
		failing++
		it := Item{Action: "Check the task", Link: "/cronjobs"}
		if j.lastError != "" && j.lastError != "none" {
			it.Detail = "Last error: " + j.lastError + "."
		}
		if j.failed == j.runs && j.runs > 1 {
			it.Title = fmt.Sprintf("Your scheduled task \"%s\" failed every time it ran (%d times)", name, j.runs)
			b.attention(it)
		} else {
			it.Title = fmt.Sprintf("Your scheduled task \"%s\" failed %d of the %d times it ran", name, j.failed, j.runs)
			b.look(it)
		}
	}
	if len(jobs) > 0 && failing == 0 {
		b.good("Your scheduled tasks ran without errors.")
	}
}

// containerEvents counts out-of-memory kills and crashes, 0, 137 and 143 are normal stops
func containerEvents(out string) (oom, crashed map[string]int) {
	oom, crashed = map[string]int{}, map[string]int{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		switch f[0] {
		case "oom":
			oom[f[1]]++
		case "died":
			if len(f) >= 3 && f[2] != "0" && f[2] != "137" && f[2] != "143" {
				crashed[f[1]]++
			}
		}
	}
	return oom, crashed
}

// checkServices reports services that ran out of memory or crashed, they were restarted automatically
func (b *builder) checkServices(ctx context.Context) {
	if !b.can("services") {
		return
	}
	args := append([]string{"events"}, b.window()...)
	if len(args) == 3 {
		// events streams forever without an end, so stop it at now
		args = append(args, "--until", time.Now().Format(time.RFC3339))
	}
	out := b.podman(ctx, append(args, "--filter", "event=died", "--filter", "event=oom", "--format", "{{.Status}} {{.Name}} {{.ContainerExitCode}}")...)
	oom, crashed := containerEvents(out)
	own := b.ownServices()
	for _, m := range []map[string]int{oom, crashed} {
		for name := range m {
			if !own[name] {
				delete(m, name)
			}
		}
	}
	for _, name := range topKeys(oom, len(oom)) {
		it := Item{Title: fmt.Sprintf("%s ran out of memory and had to be restarted (%s)", name, plural(oom[name], "time", "times")),
			Detail: "Your websites may have been slow or down for a moment while it restarted."}
		if b.can("docker") {
			it.Action, it.Link = "Raise its memory limit", "/containers"
		} else {
			it.Action = "Ask your hosting provider for more memory, or lower its use, for example fewer PHP workers or a smaller database cache."
		}
		b.look(it)
	}
	for _, name := range topKeys(crashed, len(crashed)) {
		if oom[name] > 0 {
			continue
		}
		it := Item{
			Title:  fmt.Sprintf("%s stopped unexpectedly and was restarted automatically (%s)", name, plural(crashed[name], "time", "times")),
			Action: "Check the service", Link: "/services",
		}
		// a handful of restarts can be a hiccup, dozens means it can't stay up
		if crashed[name] >= 10 {
			it.Detail = "Restarting this often usually means it can't start properly."
			if name == "cron" {
				it.Detail += " Your scheduled tasks may not be running."
			}
		}
		b.look(it)
	}
}

// ownServices are the account's compose services by service and container name, temporary containers like backups aren't the user's to fix
func (b *builder) ownServices() map[string]bool {
	own := map[string]bool{}
	compose, err := docker.LoadCompose(b.userContext)
	if err != nil {
		return own
	}
	services, _ := compose["services"].(map[string]any)
	for name, raw := range services {
		own[name] = true
		if svc, ok := raw.(map[string]any); ok {
			if cname, ok := svc["container_name"].(string); ok && cname != "" {
				own[cname] = true
			}
		}
	}
	// without jobs the cron service exits right away and keeps restarting, that's expected
	if !b.hasCronJobs() {
		delete(own, "cron")
	}
	return own
}

func (b *builder) hasCronJobs() bool {
	raw, err := tenantfs.ReadFile(homeDir + b.userContext + "/crons.ini")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "[job-") {
			return true
		}
	}
	return false
}

// dirSize adds up file sizes under dir, used for the Trash
func dirSize(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}

// checkDisk warns before the account runs out of disk or inodes and points at what can free space for this account
func (b *builder) checkDisk(ctx context.Context) {
	raw, err := tenantfs.ReadFile(quotaReportPath)
	if err != nil {
		return
	}
	var report struct {
		Users []struct {
			Username   string `json:"username"`
			DiskUsed   int64  `json:"disk_used"`
			DiskHard   int64  `json:"disk_hard"`
			InodesUsed int64  `json:"inodes_used"`
			InodesHard int64  `json:"inodes_hard"`
		} `json:"users"`
	}
	if json.Unmarshal(raw, &report) != nil {
		return
	}
	for _, u := range report.Users {
		if u.Username != b.username {
			continue
		}
		// quota numbers are 1K blocks
		used, limit := u.DiskUsed*1024, u.DiskHard*1024
		if limit > 0 {
			pct := int(used * 100 / limit)
			b.s.Usage = append(b.s.Usage, UsageBar{Title: "Disk space", Used: humanBytes(used), Total: humanBytes(limit), Percent: pct, Limit: 85})
			if pct >= 85 {
				it := Item{Title: fmt.Sprintf("Your disk is %d%% full", pct), Detail: fmt.Sprintf("You're using %s of %s.", humanBytes(used), humanBytes(limit))}
				trash := int64(0)
				if b.can("trash") {
					trash = dirSize(homeDir + b.userContext + "/.local/share/Trash")
				}
				switch {
				case trash*10 >= used:
					it.Detail += " The Trash alone holds " + humanBytes(trash) + "."
					it.Action, it.Link = "Empty the Trash", "/files.trash"
				case b.can("disk_usage"):
					it.Action, it.Link = "See what uses the space", "/disk-usage/"
				default:
					it.Action = "Delete files you no longer need, or ask your hosting provider for more disk space."
				}
				if pct >= 95 {
					it.Detail += " Once it's full, your websites can't save uploads and new emails can't arrive."
					b.attention(it)
				} else {
					b.look(it)
				}
			}
		}
		if u.InodesHard > 0 {
			pct := int(u.InodesUsed * 100 / u.InodesHard)
			b.s.Usage = append(b.s.Usage, UsageBar{Title: "Files and folders (inodes)", Used: num(int(u.InodesUsed)), Total: num(int(u.InodesHard)) + " files", Percent: pct, Limit: 85})
			if pct >= 85 {
				it := Item{Title: fmt.Sprintf("You're close to your file limit (%d%%)", pct),
					Detail: fmt.Sprintf("You have %s of the %s files your plan allows. Cache and session folders often hold most of them.", num(int(u.InodesUsed)), num(int(u.InodesHard)))}
				if b.can("inodes") {
					it.Action, it.Link = "See which folders have the most files", "/inodes-explorer"
				} else {
					it.Action = "Delete files you no longer need, or ask your hosting provider for a higher limit."
				}
				if pct >= 95 {
					b.attention(it)
				} else {
					b.look(it)
				}
			}
		}
		return
	}
}

var backupDateRE = regexp.MustCompile(`(20\d\d)[-_.]?(\d\d)[-_.]?(\d\d)`)

// checkBackups looks at the newest backup, skipped when the admin manages backups for this account
func (b *builder) checkBackups() {
	if !b.can("backups") || backups.AdminManagedBackups(b.userContext) {
		return
	}
	raw, err := tenantfs.ReadFile(homeDir + b.userContext + "/available_backups.json")
	if err != nil {
		return
	}
	var list []struct {
		BackupFile string `json:"backup_file"`
		Error      string `json:"error"`
	}
	if json.Unmarshal(raw, &list) != nil {
		return
	}
	var newest time.Time
	for _, bk := range list {
		if bk.Error != "" {
			continue
		}
		m := backupDateRE.FindStringSubmatch(bk.BackupFile)
		if m == nil {
			continue
		}
		if t, err := time.ParseInLocation("20060102", m[1]+m[2]+m[3], b.from.Location()); err == nil && t.After(newest) {
			newest = t
		}
	}
	if newest.IsZero() {
		if len(list) == 0 {
			b.look(Item{Title: "You don't have any backups yet", Detail: "Without one, a deleted file or a broken update can't be undone.", Action: "Set up backups", Link: "/backups"})
		}
		return
	}
	age := int(b.to.Sub(newest).Hours() / 24)
	if age > 7 {
		b.look(Item{Title: fmt.Sprintf("Your newest backup is %d days old", age), Detail: "Anything you changed since then isn't backed up.", Action: "Check your backups", Link: "/backups"})
		return
	}
	b.good("Your account was backed up on " + newest.Format("2 January") + ".")
}

// sizeBytes parses mailbox sizes like 512M, 1G or 1024K
func sizeBytes(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(strings.TrimSuffix(strings.ToUpper(s), "B")))
	if s == "" {
		return 0
	}
	mult := int64(1)
	switch s[len(s)-1] {
	case 'K':
		mult = 1 << 10
	case 'M':
		mult = 1 << 20
	case 'G':
		mult = 1 << 30
	case 'T':
		mult = 1 << 40
	}
	if mult > 1 {
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(n * float64(mult))
}

// checkMailboxes warns about mailboxes close to their quota, a full mailbox bounces new mail
func (b *builder) checkMailboxes(ctx context.Context) {
	if !b.can("emails") {
		return
	}
	set := map[string]bool{}
	for _, d := range b.domains {
		set[d.DomainURL] = true
	}
	lines := emails.GetEmailList(ctx, b.a, b.userID, b.username, set)
	if len(lines) == 0 {
		return
	}
	full := 0
	for _, line := range lines {
		f := strings.Fields(strings.TrimLeft(line, "* "))
		if len(f) < 3 {
			continue
		}
		limit, used := sizeBytes(f[1]), sizeBytes(f[2])
		if limit <= 0 || used*100 < limit*90 {
			continue
		}
		full++
		b.look(Item{
			Title:  fmt.Sprintf("The mailbox %s is almost full (%d%%)", f[0], used*100/limit),
			Detail: "Once it's full, new emails bounce back to the sender.",
			Action: "Raise its quota or delete old emails", Link: "/emails",
		})
	}
	if full == 0 {
		b.good("Your mailboxes have plenty of space.")
	}
}

var activityLineRE = regexp.MustCompile(`^(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d)\s+(\S+)\s+User \S+ (.*)$`)

// collectActivity lists panel logins and account changes, a login from an IP not seen in 30 days gets flagged
func (b *builder) collectActivity() {
	f, err := tenantfs.Open(usersDir + b.username + "/activity.log")
	if err != nil {
		return
	}
	defer f.Close()
	countries := loginCountries(usersDir + b.username + "/.lastlogin")
	seen := map[string]bool{}
	since := b.from.AddDate(0, 0, -30)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := activityLineRE.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		ts, err := time.ParseInLocation("2006-01-02 15:04:05", m[1], b.from.Location())
		if err != nil || ts.Before(since) || !ts.Before(b.to) {
			continue
		}
		ip, action := m[2], m[3]
		login := strings.HasPrefix(action, "logged in")
		if ts.Before(b.from) {
			if login {
				seen[ip] = true
			}
			continue
		}
		switch {
		case login:
			// the admin logging in as the user isn't the user's own login
			if strings.Contains(action, "admin API") {
				continue
			}
			b.s.Logins = append(b.s.Logins, Login{Time: ts, IP: ip, Country: countries[ip], New: !seen[ip]})
			seen[ip] = true
		case strings.HasPrefix(action, "logged out"):
		default:
			b.s.Changes = append(b.s.Changes, ts.Format("15:04")+" "+strings.TrimSuffix(action, "."))
		}
	}

	for _, l := range b.s.Logins {
		if !l.New {
			continue
		}
		it := Item{
			Title:  "Someone logged in to your account from a new IP, " + l.IP,
			Detail: "At " + l.Time.Format("15:04") + ". If that was you, all good.",
		}
		if b.can("twofa") && !b.twofaEnabled() {
			it.Detail += " If not, change your password and turn on two-factor authentication."
			it.Action, it.Link = "Secure your account", "/account/2fa"
		} else {
			it.Detail += " If not, change your password right away."
			it.Action, it.Link = "Change your password", "/account"
		}
		b.look(it)
	}
}

// loginCountries maps IPs to the country code each login saved, lines look like "IP: 1.2.3.4 - Country: RS - Login Time: ..."
func loginCountries(path string) map[string]string {
	out := map[string]string{}
	eachLine(path, func(raw []byte) {
		parts := strings.Split(strings.TrimSpace(string(raw)), " - ")
		if len(parts) != 3 || !strings.HasPrefix(parts[0], "IP: ") || !strings.HasPrefix(parts[1], "Country: ") {
			return
		}
		cc := strings.TrimSpace(strings.TrimPrefix(parts[1], "Country: "))
		if len(cc) == 2 {
			out[strings.TrimPrefix(parts[0], "IP: ")] = strings.ToUpper(cc)
		}
	})
	return out
}

func (b *builder) twofaEnabled() bool {
	var on *bool
	_ = b.a.DB.QueryRow("SELECT twofa_enabled FROM users WHERE id = ?", b.userID).Scan(&on)
	return on != nil && *on
}

// usageSample is one hourly line of resource_usage.txt, written by opencli docker-collect_stats
type usageSample struct {
	Timestamp string `json:"timestamp"`
	Memory    struct {
		Total    struct{ Bytes int64 } `json:"total"`
		Used     struct{ Bytes int64 } `json:"used"`
		UsagePct int                   `json:"usage_pct"`
	} `json:"memory"`
	CPU struct {
		Usage struct {
			Pct   float64 `json:"pct"`
			Human string  `json:"human"`
		} `json:"usage"`
		Total struct {
			Pct   float64 `json:"pct"`
			Human string  `json:"human"`
		} `json:"total"`
	} `json:"cpu"`
	Tasks struct {
		Current  int `json:"current"`
		Limit    int `json:"limit"`
		UsagePct int `json:"usage_pct"`
	} `json:"tasks"`
}

// resourcePeak tracks one resource over the day's hourly checks
type resourcePeak struct {
	pct, hot  int
	at        time.Time
	used, cap string
}

func (p *resourcePeak) add(pct int, at time.Time, used, cap string) {
	if pct >= 90 {
		p.hot++
	}
	if pct > p.pct || p.at.IsZero() {
		p.pct, p.at, p.used, p.cap = pct, at, used, cap
	}
}

// checkResources reports when the account ran into its CPU, memory or process limit
func (b *builder) checkResources() {
	raw, err := tenantfs.ReadFile(homeDir + b.userContext + "/resource_usage.txt")
	if err != nil {
		return
	}
	var cpu, mem, tasks resourcePeak
	samples := 0
	for _, line := range strings.Split(string(raw), "\n") {
		var u usageSample
		if json.Unmarshal([]byte(line), &u) != nil {
			continue
		}
		at, err := time.Parse(time.RFC3339, u.Timestamp)
		if err != nil || at.Before(b.from) || !at.Before(b.to) {
			continue
		}
		at = at.In(b.from.Location())
		samples++
		if u.CPU.Total.Pct > 0 {
			cpu.add(int(u.CPU.Usage.Pct*100/u.CPU.Total.Pct+0.5), at, u.CPU.Usage.Human, u.CPU.Total.Human)
		}
		if u.Memory.Total.Bytes > 0 {
			mem.add(u.Memory.UsagePct, at, humanBytes(u.Memory.Used.Bytes), humanBytes(u.Memory.Total.Bytes))
		}
		if u.Tasks.Limit > 0 {
			tasks.add(u.Tasks.UsagePct, at, num(u.Tasks.Current), num(u.Tasks.Limit))
		}
	}
	if samples == 0 {
		return
	}
	link := func(it *Item, fallback string) {
		switch {
		case b.can("usage"):
			it.Action, it.Link = "See resource usage", "/server/usage"
		case b.can("process_manager"):
			it.Action, it.Link = "See what's running", "/process-manager"
		default:
			it.Action = fallback
		}
	}
	checks := fmt.Sprintf("%d of the %d hourly checks", cpu.hot, samples)
	if cpu.used != "" {
		b.s.Usage = append(b.s.Usage, UsageBar{Title: "CPU at the busiest hour", Used: cpu.used, Total: cpu.cap, Percent: cpu.pct, Limit: 90})
		if cpu.hot > 0 {
			it := Item{Title: "Your account hit its CPU limit", Detail: fmt.Sprintf("It used %s of %s at %s, and was at the limit in %s. Pages load slower when that happens.", cpu.used, cpu.cap, cpu.at.Format("15:04"), checks)}
			link(&it, "Turn on caching for your websites, or ask your hosting provider for more CPU.")
			if cpu.hot*4 >= samples {
				b.attention(it)
			} else {
				b.look(it)
			}
		}
	}
	if mem.used != "" {
		b.s.Usage = append(b.s.Usage, UsageBar{Title: "Memory at the busiest hour", Used: mem.used, Total: mem.cap, Percent: mem.pct, Limit: 90})
		if mem.hot > 0 {
			it := Item{Title: fmt.Sprintf("Your account almost ran out of memory (%d%%)", mem.pct), Detail: fmt.Sprintf("It used %s of %s at %s. When memory runs out, services are stopped and restarted.", mem.used, mem.cap, mem.at.Format("15:04"))}
			link(&it, "Lower memory use, for example fewer PHP workers or a smaller database cache, or ask your hosting provider for more memory.")
			if mem.hot*4 >= samples {
				b.attention(it)
			} else {
				b.look(it)
			}
		}
	}
	if tasks.hot > 0 {
		it := Item{Title: fmt.Sprintf("Your account is close to its process limit (%d%%)", tasks.pct), Detail: fmt.Sprintf("%s of %s processes were running at %s. Once it's reached, new processes can't start and websites show errors.", tasks.used, tasks.cap, tasks.at.Format("15:04"))}
		link(&it, "Ask your hosting provider for a higher process limit.")
		b.attention(it)
	}
	if cpu.hot == 0 && mem.hot == 0 && tasks.hot == 0 {
		b.good("Your account stayed within its CPU and memory limits.")
	}
}
