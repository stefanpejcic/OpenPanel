package web

// the classic (menu_style=classic) sidebar, carried over from before the modern one

import (
	"net/http/httptest"
	"testing"
)

func TestClassicNavPathAndUploadDownloadActiveState(t *testing.T) {
	allowed := map[string]bool{"filemanager": true}

	linkActive := func(groups []NavGroup, href string) bool {
		for _, g := range groups {
			for _, l := range g.Links {
				if l.Href == href {
					return l.Active
				}
			}
		}
		t.Fatalf("no link with href %q found", href)
		return false
	}

	t.Run("upload (no method param)", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/file-manager/upload", nil)
		groups := BuildClassicSidebarNav(allowed, nil, NavPath(r))
		if !linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be active")
		}
		if linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be inactive")
		}
	})

	t.Run("upload (method=upload)", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/file-manager/upload?method=upload", nil)
		groups := BuildClassicSidebarNav(allowed, nil, NavPath(r))
		if !linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be active")
		}
		if linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be inactive")
		}
	})

	t.Run("download (method=download)", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/file-manager/upload?method=download", nil)
		groups := BuildClassicSidebarNav(allowed, nil, NavPath(r))
		if linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be inactive")
		}
		if !linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be active")
		}
	})
}

func TestClassicSidebarNavEmpty(t *testing.T) {
	groups := BuildClassicSidebarNav(map[string]bool{}, nil, "/dashboard")
	if len(groups) != 0 {
		t.Errorf("expected no groups for an empty allowed set, got %d: %+v", len(groups), groups)
	}
}

func TestClassicSidebarNavFilesGroup(t *testing.T) {
	allowed := map[string]bool{"filemanager": true, "ftp": true}
	groups := BuildClassicSidebarNav(allowed, nil, "/files")

	if len(groups) != 1 || groups[0].Label != "Files" {
		t.Fatalf("expected exactly one Files group, got %+v", groups)
	}
	g := groups[0]
	if !g.Open || !g.Active {
		t.Errorf("expected Files group to be open+active on /files, got open=%v active=%v", g.Open, g.Active)
	}

	var fileManagerLink *NavLink
	for i := range g.Links {
		if g.Links[i].Href == "/files" {
			fileManagerLink = &g.Links[i]
		}
	}
	if fileManagerLink == nil {
		t.Fatal("expected a /files link in the Files group")
	}
	if !fileManagerLink.Active {
		t.Error("expected the /files link to be marked active when request path is /files")
	}

	// backup_wizard wasn't granted, so its link shouldn't appear.
	for _, l := range g.Links {
		if l.Href == "/backup-wizard" {
			t.Error("did not expect a Backup Wizard link without the backup_wizard feature")
		}
	}
}

func TestClassicSidebarNavMySQLGroupPhpMyAdminGated(t *testing.T) {
	withPMA := BuildClassicSidebarNav(map[string]bool{"mysql": true, "phpmyadmin": true}, nil, "/mysql")
	withoutPMA := BuildClassicSidebarNav(map[string]bool{"mysql": true}, nil, "/mysql")

	hasPMA := func(groups []NavGroup) bool {
		for _, l := range groups[0].Links {
			if l.Href == "/mysql/phpmyadmin" {
				return true
			}
		}
		return false
	}

	if !hasPMA(withPMA) {
		t.Error("expected phpMyAdmin link when 'phpmyadmin' is allowed")
	}
	if hasPMA(withoutPMA) {
		t.Error("did not expect phpMyAdmin link when 'phpmyadmin' is not allowed")
	}
}

func TestClassicSidebarNavDockerGroupSimplePath(t *testing.T) {
	groups := BuildClassicSidebarNav(map[string]bool{"docker": true}, nil, "/containers/terminal")
	if len(groups) != 1 || groups[0].Label != "Containers" {
		t.Fatalf("expected a Containers group, got %+v", groups)
	}
	for _, l := range groups[0].Links {
		if l.Href == "/containers/terminal" && !l.Active {
			t.Error("expected the Terminal link to be active on /containers/terminal")
		}
	}
}

func TestResolveMenuStyle(t *testing.T) {
	cases := []struct{ cookie, configured, want string }{
		{"", "", "classic"},
		{"", "modern", "modern"},
		{"", "something", "classic"},
		{"classic", "modern", "classic"},
		{"modern", "classic", "modern"},
		{"bogus", "modern", "modern"},
	}
	for _, c := range cases {
		if got := resolveMenuStyle(c.cookie, c.configured); got != c.want {
			t.Errorf("resolveMenuStyle(%q, %q) = %q, want %q", c.cookie, c.configured, got, c.want)
		}
	}
}
