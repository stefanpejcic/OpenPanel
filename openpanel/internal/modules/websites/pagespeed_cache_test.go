package websites

import (
	"testing"
	"time"
)

func TestPageSpeedFailureCache(t *testing.T) {
	c := &pageSpeedFailureCache{m: map[string]pageSpeedFailure{}}
	if _, ok := c.get("a.com"); ok {
		t.Fatal("empty cache should miss")
	}
	c.set("a.com", pageSpeedFailure{status: 500, message: "API Error"})
	if f, ok := c.get("a.com"); !ok || f.status != 500 || f.message != "API Error" {
		t.Fatalf("expected cached failure, got %+v %v", f, ok)
	}
	c.forget("a.com")
	if _, ok := c.get("a.com"); ok {
		t.Fatal("forget should drop the entry")
	}
	c.m["b.com"] = pageSpeedFailure{status: 500, at: time.Now().Add(-2 * pageSpeedFailureTTL)}
	if _, ok := c.get("b.com"); ok {
		t.Fatal("expired entry should miss")
	}
}
