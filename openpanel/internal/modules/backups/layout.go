package backups

import (
	"slices"
	"strings"
)

// earlySections sit in front of html/ as backup/01-<name> so restoring them never reads the website files, archives from before that have them under the plain name after html/
var earlySections = []string{"mail", "mysql", "postgres", "vhosts"}

// canonicalName turns backup/01-<section>/... into backup/<section>/..., the rest of the code only knows the plain names
func canonicalName(name string) (string, bool) {
	rest, ok := strings.CutPrefix(name, "backup/01-")
	if !ok {
		return name, false
	}
	section, sub, _ := strings.Cut(rest, "/")
	if !slices.Contains(earlySections, section) {
		return name, false
	}
	if sub == "" {
		return "backup/" + section, true
	}
	return "backup/" + section + "/" + sub, true
}

// archiveLayout works out, while streaming, which sections an archive has in front under 01- and which still come after html/
type archiveLayout struct {
	early map[string]bool // archive has backup/01-<section>
	past  map[string]bool // stream went past where backup/01-<section> would be
}

func newArchiveLayout() *archiveLayout {
	return &archiveLayout{early: map[string]bool{}, past: map[string]bool{}}
}

// observe notes a member as it comes off the stream and returns its canonical name
func (l *archiveLayout) observe(name string) string {
	canonical, isEarly := canonicalName(name)
	if isEarly {
		section, _, _ := strings.Cut(strings.TrimPrefix(canonical, "backup/"), "/")
		l.early[section] = true
	}
	for _, section := range earlySections {
		dir := "backup/01-" + section
		if !insidePath(name, dir) && walkCompare(name, dir) > 0 {
			l.past[section] = true
		}
	}
	return canonical
}

// physical is where a canonical path sits in this archive, its 01- dir until the stream shows the archive doesn't have one
func (l *archiveLayout) physical(target string) string {
	rest, ok := strings.CutPrefix(target, "backup/")
	if !ok {
		return target
	}
	section, _, _ := strings.Cut(rest, "/")
	if !slices.Contains(earlySections, section) || (l.past[section] && !l.early[section]) {
		return target
	}
	return "backup/01-" + rest
}

// pastAll is pastAll for canonical targets, given the member name as it is in the archive
func (l *archiveLayout) pastAll(name string, targets []string) bool {
	for _, t := range targets {
		p := l.physical(t)
		if insidePath(name, p) || walkCompare(name, p) < 0 {
			return false
		}
	}
	return true
}
