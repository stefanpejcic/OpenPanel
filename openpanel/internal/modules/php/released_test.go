package php

import (
	"testing"
	"time"
)

func TestReleasedSkipsUnreleasedVersions(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		v    VersionInfo
		want bool
	}{
		{"future flag", VersionInfo{IsFutureVersion: true}, false},
		{"no date", VersionInfo{}, false},
		{"release date ahead", VersionInfo{ReleaseDate: "2026-11-19"}, false},
		{"released today", VersionInfo{ReleaseDate: "2026-11-09"}, true},
		{"released earlier", VersionInfo{ReleaseDate: "2025-11-20"}, true},
	}
	for _, c := range cases {
		if got := c.v.released(now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
