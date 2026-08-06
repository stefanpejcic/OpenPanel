package crons

import (
	"strings"
	"testing"
)

func TestHasCronJobs(t *testing.T) {
	if hasCronJobs([]string{"foo", "bar"}) {
		t.Error("expected false for lines with no job-exec")
	}
	if !hasCronJobs([]string{"foo", `[job-exec "x"]`, "bar"}) {
		t.Error("expected true when a job-exec line is present")
	}
}

func TestValidateCronSchedule(t *testing.T) {
	if got := validateCronSchedule(""); got != "Schedule is empty." {
		t.Errorf("empty: got %q", got)
	}
	if got := validateCronSchedule("  "); got != "Schedule is empty." {
		t.Errorf("whitespace: got %q", got)
	}
	if got := validateCronSchedule("@daily"); got != "" {
		t.Errorf("@daily: got %q, want empty", got)
	}
	if got := validateCronSchedule("@every 10s"); got != "" {
		t.Errorf("@every: got %q, want empty", got)
	}
	if got := validateCronSchedule("0 0 1 * * *"); got != "" {
		t.Errorf("6-field: got %q, want empty", got)
	}
	if got := validateCronSchedule("0 1 * * *"); got == "" {
		t.Error("5-field: expected an error message")
	} else if !strings.Contains(got, "5 fields") {
		t.Errorf("5-field message = %q, want mention of 5 fields", got)
	}
	if got := validateCronSchedule("garbage"); got == "" {
		t.Error("garbage: expected an error message")
	} else if !strings.Contains(got, "6 fields") {
		t.Errorf("garbage message = %q, want mention of 6 fields", got)
	}
}

func TestContainsAnyPattern(t *testing.T) {
	if !containsAnyPattern("touch /tmp/x; image=evil", forbiddenPatterns) {
		t.Error("expected image= to be detected")
	}
	if !containsAnyPattern("NETWORK = host", forbiddenPatterns) {
		t.Error("expected case-insensitive network= to be detected")
	}
	if containsAnyPattern("echo imagexyz=1", forbiddenPatterns) {
		t.Error("expected word-boundary to prevent matching 'imagexyz='")
	}
	if !containsAnyPattern("job-run something", execPatterns) {
		t.Error("expected job-run to be detected")
	}
}

func TestParseCronFile(t *testing.T) {
	content := `[job-exec "backup-db"]
schedule = 0 0 1 * * *
container = mysql
command = /usr/local/bin/backup.sh --full

[job-exec "cleanup"]
schedule = @every 30s
container = nginx-proxy
command = rm -rf /tmp/cache/*

`
	jobs := ParseCronFile(content)
	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2: %+v", len(jobs), jobs)
	}
	if jobs[0].Comment != "backup-db" || jobs[0].Schedule != "0 0 1 * * *" || jobs[0].Container != "mysql" || jobs[0].Command != "/usr/local/bin/backup.sh --full" {
		t.Errorf("jobs[0] = %+v", jobs[0])
	}
	if jobs[1].Comment != "cleanup" || jobs[1].Schedule != "@every 30s" || jobs[1].Container != "nginx-proxy" || jobs[1].Command != "rm -rf /tmp/cache/*" {
		t.Errorf("jobs[1] = %+v", jobs[1])
	}
}

func TestSplitJobExecSections(t *testing.T) {
	content := `[job-exec "a"]
x = 1

[job-exec "b"]
y = 2
`
	sections := splitJobExecSections(content)
	if len(sections) != 3 {
		t.Fatalf("got %d sections, want 3: %q", len(sections), sections)
	}
	if sections[0] != "" {
		t.Errorf("sections[0] = %q, want empty (nothing before first match)", sections[0])
	}
	if sections[1] != "[job-exec \"a\"]\nx = 1\n\n" {
		t.Errorf("sections[1] = %q", sections[1])
	}
	if sections[2] != "[job-exec \"b\"]\ny = 2\n" {
		t.Errorf("sections[2] = %q", sections[2])
	}
}

func TestSplitJobExecSectionsNoMatch(t *testing.T) {
	sections := splitJobExecSections("no jobs here")
	if len(sections) != 1 || sections[0] != "no jobs here" {
		t.Errorf("sections = %v", sections)
	}
}

func TestReadLinesKeepEnds(t *testing.T) {
	lines := readLinesKeepEnds("a\nb\nc")
	if len(lines) != 3 || lines[0] != "a\n" || lines[1] != "b\n" || lines[2] != "c" {
		t.Errorf("lines = %q", lines)
	}
	if lines := readLinesKeepEnds(""); lines != nil {
		t.Errorf("empty input: got %v, want nil", lines)
	}
	if lines := readLinesKeepEnds("a\nb\n"); len(lines) != 2 || lines[1] != "b\n" {
		t.Errorf("trailing newline: got %q", lines)
	}
}

func TestSectionHeaderComment(t *testing.T) {
	got, ok := sectionHeaderComment(`[job-exec "backup-db"]`)
	if !ok || got != "backup-db" {
		t.Errorf("got %q, %v", got, ok)
	}
	if _, ok := sectionHeaderComment("no quotes here"); ok {
		t.Error("expected ok=false for a line with no quotes")
	}
}

func TestServiceNamesFromCompose(t *testing.T) {
	compose := map[string]any{
		"services": map[string]any{
			"mysql": map[string]any{}, "nginx-proxy": map[string]any{}, "cron": map[string]any{}, "docker-proxy": map[string]any{},
		},
	}
	names, ok := serviceNamesFromCompose(compose)
	if !ok {
		t.Fatal("expected ok=true")
	}
	set := map[string]bool{}
	for _, n := range names {
		set[n] = true
	}
	if len(names) != 2 || !set["mysql"] || !set["nginx-proxy"] {
		t.Errorf("names = %v, want exactly [mysql nginx-proxy] (cron/docker-proxy excluded)", names)
	}
}

func TestServiceNamesFromComposeMissing(t *testing.T) {
	if _, ok := serviceNamesFromCompose(map[string]any{}); ok {
		t.Error("expected ok=false when services key is missing")
	}
}
