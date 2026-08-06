package dashboard

import "testing"

func TestBuildDashboardSectionsEmpty(t *testing.T) {
	sections := buildDashboardSections(map[string]bool{})
	if len(sections) != 0 {
		t.Errorf("expected no sections for an empty allowed set, got %d: %+v", len(sections), sections)
	}
}

func TestBuildDashboardSectionsFilesOnly(t *testing.T) {
	sections := buildDashboardSections(map[string]bool{"filemanager": true})
	if len(sections) != 1 || sections[0].Key != "files" {
		t.Fatalf("expected exactly one 'files' section, got %+v", sections)
	}
	for _, item := range sections[0].Items {
		if item.Key != "filemanager" {
			t.Errorf("expected only filemanager items, got item with key %q", item.Key)
		}
	}
	// 3 filemanager items in the source list.
	if len(sections[0].Items) != 3 {
		t.Errorf("expected 3 filemanager items, got %d: %+v", len(sections[0].Items), sections[0].Items)
	}
}

func TestBuildDashboardSectionsPreservesOrder(t *testing.T) {
	allowed := map[string]bool{"docker": true, "filemanager": true, "account": true}
	sections := buildDashboardSections(allowed)

	var keys []string
	for _, s := range sections {
		keys = append(keys, s.Key)
	}
	want := []string{"files", "docker", "account"}
	if len(keys) != len(want) {
		t.Fatalf("got sections %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("section order[%d] = %q, want %q (full: %v)", i, keys[i], want[i], keys)
		}
	}
}

func TestBuildDashboardSectionsTargetBlank(t *testing.T) {
	sections := buildDashboardSections(map[string]bool{"phpmyadmin": true})
	if len(sections) != 1 {
		t.Fatalf("expected one section, got %+v", sections)
	}
	for _, item := range sections[0].Items {
		if item.Href == "/phpmyadmin" && item.Target != "_blank" {
			t.Errorf("expected phpMyAdmin link to open in a new tab, got target=%q", item.Target)
		}
	}
}
