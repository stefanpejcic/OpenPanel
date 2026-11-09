package web

import (
	"testing"
	"time"
)

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
	if tmpl, err := pages[0].template(); err != nil || tmpl == nil {
		t.Error("dropped page didn't parse again")
	}
}

// a page that fails to parse has to keep answering, it used to leave its lock held and hang every later request
func TestBrokenPageDoesNotLock(t *testing.T) {
	p := MustLoadPage("does/not/exist.html")
	for i := 0; i < 2; i++ {
		done := make(chan error, 1)
		go func() { _, err := p.template(); done <- err }()
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected a parse error")
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("call %d hung on the page lock", i+1)
		}
	}
}
