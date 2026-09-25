package php

import (
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
)

func basePHPStats() PHPTuningStats {
	return PHPTuningStats{
		Version: "8.4", Running: true, RAMLimit: gb, MaxChildren: 5,
		Settings: map[string]string{
			"memory_limit": "256M", "max_execution_time": "120", "upload_max_filesize": "64M", "post_max_size": "64M",
			"max_input_vars": "3000", "display_errors": "Off", "log_errors": "On", "date.timezone": "Europe/Belgrade",
		},
		Domains:       []string{"example.com"},
		AvailableKeys: loadKeysFromFile("nobody-" + t0),
	}
}

// t0 keeps loadKeysFromFile from finding a per-user override on the test machine
const t0 = "phptuningtest"

func phpRec(r dbtuning.Report, key string) *dbtuning.Recommendation {
	for i := range r.Recommendations {
		if r.Recommendations[i].Key == key {
			return &r.Recommendations[i]
		}
	}
	return nil
}

func TestPHPHealthyHasNoRecommendations(t *testing.T) {
	if r := BuildPHPRecommendations(basePHPStats()); len(r.Recommendations) != 0 {
		t.Errorf("expected none, got %+v", r.Recommendations)
	}
}

func TestPHPParseAndFormatSize(t *testing.T) {
	for in, want := range map[string]int64{"256M": 256 * mb, "1G": gb, "512k": 512 * kb, "-1": -1, "134217728": 128 * mb} {
		if got, ok := parsePHPSize(in); !ok || got != want {
			t.Errorf("parsePHPSize(%q) = %d, %v", in, got, ok)
		}
	}
	for in, want := range map[int64]string{256 * mb: "256M", gb: "1G", 1536 * mb: "1536M"} {
		if got := formatPHPSize(in); got != want {
			t.Errorf("formatPHPSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestPHPMemoryExhaustedInLog(t *testing.T) {
	s := basePHPStats()
	s.Settings["memory_limit"] = "128M"
	s.LogLines = []string{`NOTICE: PHP message: PHP Fatal error:  Allowed memory size of 134217728 bytes exhausted (tried to allocate 20480 bytes) in /var/www/html/wp-includes/class-wpdb.php`}
	rec := phpRec(BuildPHPRecommendations(s), "memory_limit")
	if rec == nil || rec.Recommended != "256M" || !strings.Contains(rec.Reason, "128 MB") {
		t.Errorf("got %+v", rec)
	}
}

func TestPHPUnlimitedMemory(t *testing.T) {
	s := basePHPStats()
	s.Settings["memory_limit"] = "-1"
	if rec := phpRec(BuildPHPRecommendations(s), "memory_limit"); rec == nil || rec.Recommended != "512M" {
		t.Errorf("got %+v", rec)
	}
}

func TestPHPMemoryLimitAboveContainer(t *testing.T) {
	s := basePHPStats()
	s.Settings["memory_limit"] = "2G"
	if rec := phpRec(BuildPHPRecommendations(s), "memory_limit"); rec == nil || rec.Recommended != "512M" {
		t.Errorf("got %+v", rec)
	}
}

func TestPHPWordPressDefaults(t *testing.T) {
	s := basePHPStats()
	s.WordPressSites = 2
	s.Settings["memory_limit"] = "128M"
	s.Settings["upload_max_filesize"] = "2M"
	s.Settings["post_max_size"] = "8M"
	s.Settings["max_input_vars"] = "1000"
	s.Settings["max_execution_time"] = "30"
	r := BuildPHPRecommendations(s)
	for key, want := range map[string]string{"memory_limit": "256M", "upload_max_filesize": "64M", "post_max_size": "64M", "max_input_vars": "3000", "max_execution_time": "120"} {
		if rec := phpRec(r, key); rec == nil || rec.Recommended != want {
			t.Errorf("%s: got %+v", key, rec)
		}
	}
}

func TestPHPPostSmallerThanUpload(t *testing.T) {
	s := basePHPStats()
	s.Settings["upload_max_filesize"] = "256M"
	s.Settings["post_max_size"] = "8M"
	r := BuildPHPRecommendations(s)
	rec := phpRec(r, "post_max_size")
	if rec == nil || rec.Recommended != "256M" || !strings.Contains(rec.Reason, "smaller than upload_max_filesize") {
		t.Errorf("got %+v", rec)
	}
	if phpRec(r, "upload_max_filesize") != nil {
		t.Error("upload_max_filesize is fine")
	}
}

func TestPHPPostTooBigInLog(t *testing.T) {
	s := basePHPStats()
	s.LogLines = []string{`PHP Warning:  PHP Request Startup: POST Content-Length of 157286400 bytes exceeds the limit of 67108864 bytes in Unknown on line 0`}
	r := BuildPHPRecommendations(s)
	for _, key := range []string{"upload_max_filesize", "post_max_size"} {
		if rec := phpRec(r, key); rec == nil || rec.Recommended != "180M" {
			t.Errorf("%s: got %+v", key, rec)
		}
	}
}

func TestPHPTimeoutsAndErrors(t *testing.T) {
	s := basePHPStats()
	s.Settings["max_execution_time"] = "0"
	s.Settings["display_errors"] = "On"
	s.Settings["log_errors"] = "0"
	s.Settings["date.timezone"] = ""
	r := BuildPHPRecommendations(s)
	for key, want := range map[string]string{"max_execution_time": "300", "display_errors": "Off", "log_errors": "On", "date.timezone": "UTC"} {
		if rec := phpRec(r, key); rec == nil || rec.Recommended != want {
			t.Errorf("%s: got %+v", key, rec)
		}
	}
}

func TestPHPExecTimeoutInLog(t *testing.T) {
	s := basePHPStats()
	s.Settings["max_execution_time"] = "30"
	s.LogLines = []string{"PHP Fatal error:  Maximum execution time of 30 seconds exceeded in /var/www/html/index.php"}
	if rec := phpRec(BuildPHPRecommendations(s), "max_execution_time"); rec == nil || rec.Recommended != "120" {
		t.Errorf("got %+v", rec)
	}
}

func TestPHPNotesForUnusedVersionAndWorkers(t *testing.T) {
	s := basePHPStats()
	s.Domains = nil
	s.Settings["display_errors"] = "On"
	r := BuildPHPRecommendations(s)
	if phpRec(r, "display_errors") != nil {
		t.Error("display_errors only matters when a domain uses this version")
	}
	if len(r.Notes) == 0 || !strings.Contains(r.Notes[0], "No domains use PHP 8.4") {
		t.Errorf("notes: %v", r.Notes)
	}
	s = basePHPStats()
	s.LogLines = []string{"WARNING: [pool www] server reached pm.max_children setting (5), consider raising it"}
	if r := BuildPHPRecommendations(s); len(r.Notes) == 0 || !strings.Contains(r.Notes[0], "ran out of workers") {
		t.Errorf("notes: %v", r.Notes)
	}
}

func TestPHPOnlyKeysOnThePage(t *testing.T) {
	s := basePHPStats()
	s.Settings["memory_limit"] = "-1"
	s.AvailableKeys = []string{"display_errors"}
	if phpRec(BuildPHPRecommendations(s), "memory_limit") != nil {
		t.Error("memory_limit isn't on the page")
	}
}

func TestPHPTooManyWorkersIsANoteNotALowerLimit(t *testing.T) {
	s := basePHPStats()
	s.Settings["memory_limit"] = "512M"
	s.MaxChildren = 20
	r := BuildPHPRecommendations(s)
	if phpRec(r, "memory_limit") != nil {
		t.Error("memory_limit shouldn't be lowered just because many workers exist")
	}
	if len(r.Notes) == 0 || !strings.Contains(r.Notes[0], "20 PHP workers") {
		t.Errorf("notes: %v", r.Notes)
	}
}
