package php

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
)

const (
	kb = dbtuning.KB
	mb = dbtuning.MB
	gb = dbtuning.GB
)

// PHPTuningStats is everything the PHP rules look at, gathered from the user's php.ini, the running container and its log
type PHPTuningStats struct {
	Version        string
	Running        bool
	RAMLimit       int64
	MemUsage       int64
	CPUs           float64
	OOMKilled      bool
	Settings       map[string]string
	Domains        []string
	WordPressSites int
	MaxChildren    int
	LogLines       []string
	AvailableKeys  []string
}

var phpSizeRE = regexp.MustCompile(`^(-?\d+)\s*([KMG]?)$`)

// parsePHPSize reads php.ini shorthand like 256M, -1 means unlimited
func parsePHPSize(v string) (int64, bool) {
	m := phpSizeRE.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(v)))
	if m == nil {
		return 0, false
	}
	n, _ := strconv.ParseInt(m[1], 10, 64)
	switch m[2] {
	case "K":
		n *= kb
	case "M":
		n *= mb
	case "G":
		n *= gb
	}
	return n, true
}

// formatPHPSize writes bytes the way the options page shows them, as M or G
func formatPHPSize(b int64) string {
	if b >= gb && b%gb == 0 {
		return strconv.FormatInt(b/gb, 10) + "G"
	}
	return strconv.FormatInt(dbtuning.RoundUp(b, mb)/mb, 10) + "M"
}

func (s *PHPTuningStats) usable(key string) bool {
	return slices.Contains(s.AvailableKeys, key)
}

func (s *PHPTuningStats) size(key string) (int64, bool) {
	return parsePHPSize(s.Settings[key])
}

func (s *PHPTuningStats) number(key string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(s.Settings[key]), 10, 64)
	return n, err == nil
}

func isOn(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "on", "yes", "true":
		return true
	}
	return false
}

var (
	memExhaustedRE = regexp.MustCompile(`Allowed memory size of (\d+) bytes exhausted`)
	maxExecRE      = regexp.MustCompile(`Maximum execution time of (\d+) seconds? exceeded`)
	postTooBigRE   = regexp.MustCompile(`POST Content-Length of (\d+) bytes exceeds the limit`)
	inputVarsRE    = regexp.MustCompile(`Input variables exceeded (\d+)`)
)

// logMatches returns how many log lines match and the largest number captured by the pattern
func (s *PHPTuningStats) logMatches(re *regexp.Regexp) (count int, largest int64) {
	for _, l := range s.LogLines {
		if m := re.FindStringSubmatch(l); m != nil {
			count++
			if n, err := strconv.ParseInt(m[1], 10, 64); err == nil && n > largest {
				largest = n
			}
		}
	}
	return count, largest
}

func (s *PHPTuningStats) logCount(pattern string) int {
	n := 0
	for _, l := range s.LogLines {
		if strings.Contains(l, pattern) {
			n++
		}
	}
	return n
}

// BuildPHPRecommendations applies the PHP rules to the gathered stats
func BuildPHPRecommendations(s PHPTuningStats) dbtuning.Report {
	report := dbtuning.NewReport()
	ram := s.RAMLimit

	report.Facts["server"] = "PHP " + s.Version
	if ram > 0 {
		report.Facts["memory_limit"] = dbtuning.HumanSize(ram)
	}
	if s.MemUsage > 0 {
		report.Facts["memory_used"] = dbtuning.HumanSize(s.MemUsage)
	}
	report.Facts["domains"] = strconv.Itoa(len(s.Domains))
	if s.WordPressSites > 0 {
		report.Facts["wordpress"] = strconv.Itoa(s.WordPressSites)
	}
	if s.MaxChildren > 0 {
		report.Facts["workers"] = strconv.Itoa(s.MaxChildren)
	}
	switch {
	case len(s.Domains) == 0:
		report.Notes = append(report.Notes, "No domains use PHP "+s.Version+" yet, these settings only matter once a domain is switched to it.")
	case !s.Running:
		report.Notes = append(report.Notes, "The PHP "+s.Version+" service is not running, so suggestions are based on its settings and memory limit only.")
	}
	if ram <= 0 {
		report.Notes = append(report.Notes, dbtuning.NoteNoMemoryLimit)
	}

	add := func(key, recommended, reason string) {
		if !s.usable(key) {
			return
		}
		cur := s.Settings[key]
		if strings.EqualFold(cur, recommended) {
			return
		}
		for _, r := range report.Recommendations {
			if r.Key == key {
				return
			}
		}
		report.Recommendations = append(report.Recommendations, dbtuning.Recommendation{Key: key, Label: key, Current: cur, Recommended: recommended, Reason: reason})
	}

	// memory_limit
	if cur, ok := s.size("memory_limit"); ok {
		workers := int64(max(s.MaxChildren, 1))
		exhausted, largest := s.logMatches(memExhaustedRE)
		switch {
		case cur == -1:
			target := int64(512 * mb)
			if ram > 0 {
				target = min(target, dbtuning.RoundUp(ram/2, 64*mb))
			}
			add("memory_limit", formatPHPSize(target), "Memory limit is unlimited, so one broken script can use all the memory of the PHP service and take down every site on it. A limit stops just that script instead.")
		case ram > 0 && cur > ram:
			add("memory_limit", formatPHPSize(dbtuning.RoundUp(ram/2, 64*mb)), fmt.Sprintf("A single script is allowed to use %s, more than the whole PHP service has (%s). It would be killed before reaching the limit.", s.Settings["memory_limit"], dbtuning.HumanSize(ram)))
		case exhausted > 0:
			target := max(cur*2, 256*mb)
			if ram > 0 {
				target = min(target, dbtuning.RoundUp(ram/2, 64*mb))
			}
			if target > cur {
				add("memory_limit", formatPHPSize(target), fmt.Sprintf("The log shows %d scripts that stopped with \"Allowed memory size of %s exhausted\". Raising the limit lets them finish.", exhausted, dbtuning.HumanSize(largest)))
			}
		case s.WordPressSites > 0 && cur < 256*mb && (ram == 0 || ram >= 512*mb):
			add("memory_limit", "256M", fmt.Sprintf("%d of your sites on PHP %s run WordPress, which needs 256M for the admin area, updates and plugins like WooCommerce.", s.WordPressSites, s.Version))
		}
		if ram > 0 && cur > 0 && s.MaxChildren > 1 && cur*workers > ram*4 {
			report.Notes = append(report.Notes, fmt.Sprintf("Up to %d PHP workers can run at once and each may use %s, together far more than the %s the PHP service has. If pages get slow or the service restarts under load, give it more memory or fewer workers on the Containers page.", s.MaxChildren, s.Settings["memory_limit"], dbtuning.HumanSize(ram)))
		}
	}

	// max_execution_time
	if cur, ok := s.number("max_execution_time"); ok {
		timeouts, _ := s.logMatches(maxExecRE)
		switch {
		case cur == 0:
			add("max_execution_time", "300", "Scripts can run forever, so a stuck request keeps a PHP worker busy until it's restarted. When all workers are stuck your sites stop responding.")
		case timeouts > 0 && cur < 300:
			add("max_execution_time", strconv.FormatInt(min(max(cur*2, 120), 300), 10), fmt.Sprintf("The log shows %d scripts that were stopped after %d seconds. Imports, backups and updates often need longer.", timeouts, cur))
		case s.WordPressSites > 0 && cur < 60:
			add("max_execution_time", "120", "WordPress updates, imports and backup plugins often take longer than "+strconv.FormatInt(cur, 10)+" seconds.")
		}
	}

	// upload_max_filesize and post_max_size, an upload can't be bigger than the whole POST
	upload, okU := s.size("upload_max_filesize")
	post, okP := s.size("post_max_size")
	if okU && okP {
		uploadTarget, postTarget := upload, post
		reason := ""
		if rejected, largest := s.logMatches(postTooBigRE); rejected > 0 {
			need := dbtuning.RoundUp(largest*6/5, mb)
			if ram > 0 {
				need = min(need, gb)
			}
			uploadTarget, postTarget = max(upload, need), max(post, need)
			reason = fmt.Sprintf("The log shows %d uploads or forms rejected because they were bigger than post_max_size (the largest was %s).", rejected, dbtuning.HumanSize(largest))
		} else if s.WordPressSites > 0 && upload < 64*mb {
			uploadTarget, postTarget = 64*mb, max(post, 64*mb)
			reason = "WordPress sites often need to upload themes, plugins and media bigger than " + s.Settings["upload_max_filesize"] + "."
		}
		if post != 0 && postTarget < uploadTarget {
			postTarget = uploadTarget
		}
		if uploadTarget > upload {
			add("upload_max_filesize", formatPHPSize(uploadTarget), reason)
		}
		if postTarget > post && post != 0 {
			r := reason
			if r == "" {
				r = fmt.Sprintf("post_max_size (%s) is smaller than upload_max_filesize (%s), so uploads bigger than %s fail even though they are allowed.", s.Settings["post_max_size"], s.Settings["upload_max_filesize"], s.Settings["post_max_size"])
			} else {
				r += " post_max_size has to be at least as big as upload_max_filesize."
			}
			add("post_max_size", formatPHPSize(postTarget), r)
		}
	}

	// max_input_vars
	if cur, ok := s.number("max_input_vars"); ok {
		if n, _ := s.logMatches(inputVarsRE); n > 0 {
			add("max_input_vars", strconv.FormatInt(max(cur*2, 3000), 10), fmt.Sprintf("The log shows %d forms that were cut off at %d fields. Large menus and product pages lose their settings when this happens.", n, cur))
		} else if s.WordPressSites > 0 && cur < 3000 {
			add("max_input_vars", "3000", fmt.Sprintf("WordPress menus, page builders and WooCommerce products often send more than %d fields, and anything past the limit is silently dropped.", cur))
		}
	}

	// errors should be logged, not shown to visitors
	if v, ok := s.Settings["display_errors"]; ok && isOn(v) && len(s.Domains) > 0 {
		add("display_errors", "Off", "Errors are shown to visitors, which reveals file paths and code details to attackers. Keep them in the log instead.")
	}
	if v, ok := s.Settings["log_errors"]; ok && !isOn(v) {
		add("log_errors", "On", "Errors are not logged, so there's no way to find out why a site breaks.")
	}

	if v, ok := s.Settings["date.timezone"]; ok && strings.TrimSpace(v) == "" {
		add("date.timezone", "UTC", "No timezone is set, so PHP logs a warning on every date call and falls back to UTC anyway.")
	}

	if n := s.logCount("server reached pm.max_children setting"); n > 0 {
		report.Notes = append(report.Notes, fmt.Sprintf("The PHP service ran out of workers %d times (pm.max_children = %d), so some visitors had to wait. Raise the number of workers or the memory of the PHP service on the Containers page.", n, s.MaxChildren))
	}
	if s.OOMKilled {
		report.Notes = append(report.Notes, "The PHP service was stopped for running out of memory. Lower memory_limit or give the service more memory on the Containers page.")
	}

	return report
}
