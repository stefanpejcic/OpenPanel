package web

import "testing"

func TestPageCacheDropsLeastRecentlyUsed(t *testing.T) {
	pages := make([]*Page, pageCacheSize+3)
	for i := range pages {
		pages[i] = MustLoadPage("partials/_footer.html")
		pages[i].template()
	}
	pagesMu.Lock()
	n := len(loadedPages)
	pagesMu.Unlock()
	if n > pageCacheSize {
		t.Fatalf("%d pages cached, want at most %d", n, pageCacheSize)
	}
	if pages[0].tmpl != nil {
		t.Error("oldest page still cached")
	}
	if pages[len(pages)-1].tmpl == nil {
		t.Error("newest page not cached")
	}
	if pages[0].template() == nil {
		t.Error("dropped page didn't parse again")
	}
}
