package crons

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const twoJobs = `[job-exec "backup"]
schedule = 0 0 3 * * *
container = php-fpm-8.3
command = php backup.php

#[job-exec "cleanup"]
#schedule = 0 */5 * * * *
#container = apache
#command = rm -rf /tmp/cache
#no-overlap
`

func TestParseCronFileDisabled(t *testing.T) {
	jobs := ParseCronFile(twoJobs)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Disabled || !jobs[1].Disabled || !jobs[1].NoOverlap || jobs[1].Command != "rm -rf /tmp/cache" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
	if msg := ValidateCronFileFormat(twoJobs); msg != "" {
		t.Fatalf("a disabled job should pass the format check: %s", msg)
	}
	// only enabled jobs keep the cron container running
	if !hasCronJobs(strings.Split(twoJobs, "\n")) || hasCronJobs(strings.Split(twoJobs[strings.Index(twoJobs, "#["):], "\n")) {
		t.Fatal("hasCronJobs should only count enabled jobs")
	}
}

func TestRewriteCronJobToggleEditDelete(t *testing.T) {
	jobs := ParseCronFile(twoJobs)

	off, found := rewriteCronJob(twoJobs, func(j *CronJob) (bool, bool) {
		if j.Key() != jobs[0].Key() {
			return false, true
		}
		j.Disabled = true
		return true, true
	})
	if !found || !strings.HasPrefix(off, `#[job-exec "backup"]`+"\n#schedule = 0 0 3 * * *") {
		t.Fatalf("disable didn't comment the block out:\n%s", off)
	}
	if hasCronJobs(strings.Split(off, "\n")) {
		t.Fatal("no job should be enabled after disabling both")
	}

	// editing a disabled job keeps it disabled
	edited, _ := rewriteCronJob(off, func(j *CronJob) (bool, bool) {
		if !sameCronJob(*j, "cleanup", "0 */5 * * * *", "apache", "rm -rf /tmp/cache") {
			return false, true
		}
		*j = CronJob{Comment: "cleanup", Schedule: "0 0 * * * *", Container: "apache", Command: "true", Disabled: j.Disabled}
		return true, true
	})
	if got := ParseCronFile(edited); len(got) != 2 || !got[1].Disabled || got[1].Schedule != "0 0 * * * *" {
		t.Fatalf("edit lost the disabled state: %+v", got)
	}

	deleted, _ := rewriteCronJob(edited, func(j *CronJob) (bool, bool) { return j.Comment == "backup", false })
	if got := ParseCronFile(deleted); len(got) != 1 || got[0].Comment != "cleanup" {
		t.Fatalf("delete removed the wrong job: %+v", got)
	}
}

func TestNextCronRuns(t *testing.T) {
	from := time.Date(2026, 9, 26, 10, 0, 0, 0, time.UTC)
	runs, err := nextCronRuns("0 30 3 * * *", from, 3, time.UTC)
	if err != nil || len(runs) != 3 || !runs[0].Equal(time.Date(2026, 9, 27, 3, 30, 0, 0, time.UTC)) {
		t.Fatalf("runs=%v err=%v", runs, err)
	}
	if runs, _ := nextCronRuns("@every 90s", from, 1, time.UTC); !runs[0].Equal(from.Add(90 * time.Second)) {
		t.Fatalf("@every: %v", runs)
	}
	for _, bad := range []string{"", "0 3 * * *", "0 99 3 * * *", "@sometimes"} {
		if _, err := nextCronRuns(bad, from, 1, time.UTC); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestCronSummary(t *testing.T) {
	now := time.Date(2026, 9, 26, 10, 0, 0, 0, time.UTC)
	jobs := ParseCronFile(twoJobs)
	sum := cronSummary(jobs, now, time.UTC)
	if sum.Total != 2 || sum.Active != 1 || sum.NextJob != "backup" || sum.NextRun != "2026-09-27T03:00:00Z" {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestMoveCronJobs(t *testing.T) {
	content := twoJobs + "\n[job-exec \"db\"]\nschedule = 0 0 1 * * *\ncontainer = mariadb\ncommand = mysqldump x\n"
	out, moved, restart := moveCronJobs(content, "apache", "nginx")
	if moved != 1 || restart {
		t.Fatalf("expected the disabled apache job moved without a restart, moved=%d restart=%v", moved, restart)
	}
	jobs := ParseCronFile(out)
	if jobs[1].Container != "nginx" || !jobs[1].Disabled || jobs[2].Container != "mariadb" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
	if _, moved, restart := moveCronJobs(out, "mariadb", "mysql"); moved != 1 || !restart {
		t.Fatalf("enabled job should need a restart, moved=%d restart=%v", moved, restart)
	}
}

func TestCronServiceZone(t *testing.T) {
	cases := []struct {
		compose string
		tz      string
		mounted bool
	}{
		{"services:\n  cron:\n    image: mcuadros/ofelia\n", "", false},
		{"services:\n  cron:\n    environment:\n      TZ: Europe/Belgrade\n", "Europe/Belgrade", false},
		{"services:\n  cron:\n    environment:\n      - TZ=${TZ:-UTC}\n", "${TZ:-UTC}", false},
		{"services:\n  cron:\n    volumes:\n      - /etc/localtime:/etc/localtime:ro\n", "", true},
		{"services:\n  cron:\n    volumes:\n      - type: bind\n        source: /etc/localtime\n        target: /etc/localtime\n", "", true},
		{"services:\n  apache:\n    environment:\n      TZ: Asia/Tokyo\n", "", false},
	}
	for _, c := range cases {
		tz, mounted := cronServiceZone([]byte(c.compose))
		if tz != c.tz || mounted != c.mounted {
			t.Errorf("%q: got %q %v, want %q %v", c.compose, tz, mounted, c.tz, c.mounted)
		}
	}
}

func TestHostLocationMatchesZoneFile(t *testing.T) {
	dir := t.TempDir()
	zone, err := os.ReadFile("/usr/share/zoneinfo/Europe/Belgrade")
	if err != nil {
		t.Skip("no zoneinfo on this machine")
	}
	os.MkdirAll(filepath.Join(dir, "zoneinfo", "Europe"), 0o755)
	os.WriteFile(filepath.Join(dir, "zoneinfo", "Europe", "Belgrade"), zone, 0o644)
	os.WriteFile(filepath.Join(dir, "localtime"), zone, 0o644)
	origLocal, origDir := hostLocaltimePath, zoneinfoDir
	hostLocaltimePath, zoneinfoDir = filepath.Join(dir, "localtime"), filepath.Join(dir, "zoneinfo")
	t.Cleanup(func() { hostLocaltimePath, zoneinfoDir = origLocal, origDir })

	if got := hostLocation().String(); got != "Europe/Belgrade" {
		t.Fatalf("got %q, want Europe/Belgrade", got)
	}
}

func TestSetComposeCronTZ(t *testing.T) {
	base := "services:\n    backup:\n        image: x\n    cron:\n        command: daemon --config=/crons.ini\n%s        hostname: cron\n    docker-proxy:\n        image: y\n"
	cases := map[string]string{
		"no environment":   "",
		"map environment":  "        environment:\n            DOCKER_HOST: tcp://docker-proxy:2375\n",
		"map with TZ":      "        environment:\n            DOCKER_HOST: tcp://docker-proxy:2375\n            TZ: UTC\n",
		"list environment": "        environment:\n            - DOCKER_HOST=tcp://docker-proxy:2375\n",
		"list with TZ":     "        environment:\n            - TZ=UTC\n",
	}
	for name, env := range cases {
		in := strings.Replace(base, "%s", env, 1)
		out, err := setComposeCronTZ(in, "Europe/Belgrade")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if tz, _ := cronServiceZone([]byte(out)); tz != "Europe/Belgrade" {
			t.Errorf("%s: TZ = %q\n%s", name, tz, out)
		}
		// the other services are untouched
		if !strings.Contains(out, "    backup:\n        image: x\n") || !strings.Contains(out, "    docker-proxy:\n        image: y\n") {
			t.Errorf("%s: other services changed\n%s", name, out)
		}
		if strings.Count(out, "TZ") != 1 {
			t.Errorf("%s: expected one TZ\n%s", name, out)
		}
	}
	if _, err := setComposeCronTZ("services:\n    apache:\n        image: x\n", "UTC"); err == nil {
		t.Error("expected an error without a cron service")
	}
}
