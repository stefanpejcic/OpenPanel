package filemanager

import (
	"fmt"
	"strings"
	"testing"
)

func TestDeleteLogLine(t *testing.T) {
	cases := []struct {
		kind, itemType string
		names          []string
		want           string
	}{
		{"trash", "file", []string{"site/a.txt"}, "moved file site/a.txt to Trash using File Manager"},
		{"permanent", "directory", []string{"site/old"}, "deleted directory site/old using File Manager"},
		{"symlink", "file", []string{"site/link"}, "deleted symlink site/link using File Manager"},
		{"trash", "file", []string{"a", "b"}, "moved 2 items to Trash using File Manager: a, b"},
		{"permanent", "file", []string{"a", "b", "c"}, "deleted 3 items using File Manager: a, b, c"},
	}
	for _, c := range cases {
		if got := deleteLogLine(c.kind, c.itemType, c.names); got != c.want {
			t.Errorf("deleteLogLine(%q, %v) = %q, want %q", c.kind, c.names, got, c.want)
		}
	}
}

func TestDeleteLogLineCapsNames(t *testing.T) {
	var names []string
	for i := 0; i < 15; i++ {
		names = append(names, fmt.Sprintf("f%d", i))
	}
	got := deleteLogLine("trash", "file", names)
	if !strings.HasPrefix(got, "moved 15 items to Trash") || !strings.HasSuffix(got, "f9 and 5 more") || strings.Contains(got, "f10") {
		t.Errorf("unexpected line %q", got)
	}
}

func TestTransferLogLine(t *testing.T) {
	cases := []struct {
		action, itemType string
		names            []string
		want             string
	}{
		{"copy", "file", []string{"site/a.txt"}, "copied file site/a.txt to backup using File Manager"},
		{"move", "directory", []string{"site/old"}, "moved site/old to backup using File Manager"},
		{"copy", "file", []string{"a", "b"}, "copied 2 items to backup using File Manager: a, b"},
		{"move", "file", []string{"a", "b", "c"}, "moved 3 items to backup using File Manager: a, b, c"},
	}
	for _, c := range cases {
		if got := transferLogLine(c.action, c.itemType, c.names, "backup"); got != c.want {
			t.Errorf("transferLogLine(%q, %v) = %q, want %q", c.action, c.names, got, c.want)
		}
	}
}
