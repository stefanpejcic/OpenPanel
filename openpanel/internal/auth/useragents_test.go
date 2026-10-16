package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withUseragentFiles(t *testing.T, enabled, list string) (cfgPath, listPath string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath, listPath = filepath.Join(dir, "openpanel.config"), filepath.Join(dir, "blacklist_useragents.txt")
	os.WriteFile(cfgPath, []byte("[PANEL]\nblacklist_useragents="+enabled+"\n"), 0644)
	os.WriteFile(listPath, []byte(list), 0644)
	origCfg, origList := UseragentsConfigPath, BlacklistUseragentsPath
	UseragentsConfigPath, BlacklistUseragentsPath = cfgPath, listPath
	t.Cleanup(func() { UseragentsConfigPath, BlacklistUseragentsPath = origCfg, origList })
	return cfgPath, listPath
}

func uaStatus(h http.Handler, ua string) int {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

func TestBlockUseragentsEnabled(t *testing.T) {
	withUseragentFiles(t, "yes", "; widows\n#BLEXBot\n\nsqlmap\nAtomic_Email_Hunter\n")
	h := BlockUseragents()(okHandler)
	for ua, want := range map[string]int{
		"sqlmap/1.7.2#stable (https://sqlmap.org)":                  http.StatusForbidden,
		"Mozilla/4.0 (compatible; Widows 98)":                       http.StatusForbidden,
		"ATOMIC_EMAIL_HUNTER/4.0":                                   http.StatusForbidden,
		"Mozilla/5.0 (compatible; BLEXBot/1.0)":                     http.StatusOK, // commented out
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome": http.StatusOK,
		"": http.StatusOK,
	} {
		if got := uaStatus(h, ua); got != want {
			t.Errorf("%q: %d, want %d", ua, got, want)
		}
	}
}

func TestBlockUseragentsDisabled(t *testing.T) {
	withUseragentFiles(t, "no", "sqlmap\n")
	if got := uaStatus(BlockUseragents()(okHandler), "sqlmap/1.7"); got != http.StatusOK {
		t.Fatalf("disabled blacklist blocked a request: %d", got)
	}
}

func TestBlockUseragentsPicksUpChanges(t *testing.T) {
	cfgPath, listPath := withUseragentFiles(t, "no", "sqlmap\n")
	b := &useragentBlocker{}
	if b.blocked("sqlmap/1.7") {
		t.Fatal("blocked while disabled")
	}
	// OpenAdmin turns it on and adds an entry, the next check after the recheck window sees both
	os.WriteFile(cfgPath, []byte("blacklist_useragents=yes\n"), 0644)
	os.WriteFile(listPath, []byte("sqlmap\nnikto\n"), 0644)
	future := time.Now().Add(time.Minute)
	os.Chtimes(cfgPath, future, future)
	os.Chtimes(listPath, future, future)
	b.checkedAt = time.Now().Add(-useragentsRecheck - time.Second)
	if !b.blocked("Mozilla/5.00 (Nikto/2.5.0)") {
		t.Fatal("change to the list wasn't picked up")
	}
}

func TestBlockUseragentsMissingFiles(t *testing.T) {
	origCfg, origList := UseragentsConfigPath, BlacklistUseragentsPath
	UseragentsConfigPath, BlacklistUseragentsPath = "/nonexistent/openpanel.config", "/nonexistent/list.txt"
	t.Cleanup(func() { UseragentsConfigPath, BlacklistUseragentsPath = origCfg, origList })
	if got := uaStatus(BlockUseragents()(okHandler), "sqlmap/1.7"); got != http.StatusOK {
		t.Fatalf("missing files should allow everything, got %d", got)
	}
}
