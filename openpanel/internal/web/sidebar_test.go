package web

import (
	"net/http/httptest"
	"testing"
)

func TestNavPathAndUploadDownloadActiveState(t *testing.T) {
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
		groups := BuildSidebarNav(allowed, NavPath(r))
		if !linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be active")
		}
		if linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be inactive")
		}
	})

	t.Run("upload (method=upload)", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/file-manager/upload?method=upload", nil)
		groups := BuildSidebarNav(allowed, NavPath(r))
		if !linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be active")
		}
		if linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be inactive")
		}
	})

	t.Run("download (method=download)", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/file-manager/upload?method=download", nil)
		groups := BuildSidebarNav(allowed, NavPath(r))
		if linkActive(groups, "/file-manager/upload?method=upload") {
			t.Error("expected Upload from device to be inactive")
		}
		if !linkActive(groups, "/file-manager/upload?method=download") {
			t.Error("expected Download from URL to be active")
		}
	})
}

func TestBuildSidebarNavEmpty(t *testing.T) {
	groups := BuildSidebarNav(map[string]bool{}, "/dashboard")
	if len(groups) != 0 {
		t.Errorf("expected no groups for an empty allowed set, got %d: %+v", len(groups), groups)
	}
}

func TestBuildSidebarNavFilesGroup(t *testing.T) {
	allowed := map[string]bool{"filemanager": true, "ftp": true}
	groups := BuildSidebarNav(allowed, "/files")

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

func TestBuildSidebarNavMySQLGroupPhpMyAdminGated(t *testing.T) {
	withPMA := BuildSidebarNav(map[string]bool{"mysql": true, "phpmyadmin": true}, "/mysql")
	withoutPMA := BuildSidebarNav(map[string]bool{"mysql": true}, "/mysql")

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

func TestBuildSidebarNavDockerGroupSimplePath(t *testing.T) {
	groups := BuildSidebarNav(map[string]bool{"docker": true}, "/containers/terminal")
	if len(groups) != 1 || groups[0].Label != "Containers" {
		t.Fatalf("expected a Containers group, got %+v", groups)
	}
	for _, l := range groups[0].Links {
		if l.Href == "/containers/terminal" && !l.Active {
			t.Error("expected the Terminal link to be active on /containers/terminal")
		}
	}
}

func TestHasAnyPrefix(t *testing.T) {
	if !hasAnyPrefix("/mysql/users", "/mysql", "/postgresql") {
		t.Error("expected /mysql/users to match /mysql prefix")
	}
	if hasAnyPrefix("/domains", "/mysql", "/postgresql") {
		t.Error("did not expect /domains to match")
	}
}
