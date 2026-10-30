package web

import (
	"strings"
	"testing"
)

func TestActivateTab(t *testing.T) {
	allowed := map[string]bool{"wordpress": true, "autoinstaller": true}
	tabs := BuildPageTabs(allowed, nil, "/website", TabContext{})
	l := LayoutData{PageTabs: tabs, NavTrail: BuildNavTrail(allowed, nil, "/website", "/website", "example.com", tabs)}

	l.ActivateTab("/wordpress")
	var active []string
	for _, tab := range l.PageTabs {
		if tab.Active {
			active = append(active, tab.Label)
		}
	}
	if strings.Join(active, ",") != "WordPress" {
		t.Errorf("active tabs = %v, want only WordPress", active)
	}
	var trail []string
	for _, c := range l.NavTrail {
		trail = append(trail, c.Label+"="+c.Href)
	}
	if got := strings.Join(trail, " > "); got != "Websites=/sites > WordPress=/wordpress > example.com=" {
		t.Errorf("trail = %q", got)
	}

	// a tab the page doesn't have, or the plan doesn't grant, changes nothing
	before := tabLabels(l.PageTabs)
	l.ActivateTab("/nowhere")
	if tabLabels(l.PageTabs) != before || !l.PageTabs[1].Active {
		t.Error("an unknown href shouldn't move the active tab")
	}
	(&LayoutData{}).ActivateTab("/wordpress")
}
